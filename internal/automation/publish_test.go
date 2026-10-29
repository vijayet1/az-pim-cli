package automation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestReleaseManifest(t *testing.T) {
	// Removing any target or trusting a mismatching digest must break verification.
	tag := "v1.17.0"
	pin := fixturePin()
	m := ReleaseManifest{SchemaVersion: 1, Repository: Fork, Tag: tag, Commit: testSHA, Upstream: pin}
	for _, name := range ExpectedAssetNames(tag) {
		m.Assets = append(m.Assets, AssetDigest{Name: name, Size: 1, SHA256: strings.Repeat("a", 64)})
	}
	if e := ValidateManifest(m, tag, testSHA, pin); e != nil {
		t.Fatal(e)
	}
	for _, scenario := range []string{"missing platform", "duplicate", "path traversal", "wrong source", "wrong repo", "bad digest", "zero size"} {
		t.Run(scenario, func(t *testing.T) {
			x := m
			x.Assets = append([]AssetDigest(nil), m.Assets...)
			switch scenario {
			case "missing platform":
				x.Assets = x.Assets[1:]
			case "duplicate":
				x.Assets = append(x.Assets, x.Assets[0])
			case "path traversal":
				x.Assets[0].Name = "../escape"
			case "wrong source":
				x.Commit = strings.Repeat("b", 40)
			case "wrong repo":
				x.Repository = Upstream
			case "bad digest":
				x.Assets[0].SHA256 = "bad"
			case "zero size":
				x.Assets[0].Size = 0
			}
			if e := ValidateManifest(x, tag, testSHA, pin); e == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}
func TestVerifyAssetBytes(t *testing.T) {
	bytes := []byte("verified asset")
	sum := sha256.Sum256(bytes)
	a := AssetDigest{Name: "safe", Size: int64(len(bytes)), SHA256: hex.EncodeToString(sum[:])}
	if e := VerifyAssetBytes(a, bytes); e != nil {
		t.Fatal(e)
	}
	if e := VerifyAssetBytes(a, []byte("tampered asset")); e == nil {
		t.Fatal("accepted tampered bytes")
	}
}
func TestBuildReleaseManifest(t *testing.T) {
	dir := t.TempDir()
	var entries []map[string]string
	for _, name := range ExpectedAssetNames("v1.17.0") {
		p := filepath.Join(dir, name)
		fixtureWrite(t, p, []byte("fixture"))
		entries = append(entries, map[string]string{"name": name, "path": p, "type": "Archive"})
	}
	b, _ := json.Marshal(entries)
	artifacts := filepath.Join(dir, "artifacts.json")
	fixtureWrite(t, artifacts, b)
	m, e := BuildReleaseManifest(artifacts, "v1.17.0", testSHA, fixturePin())
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Assets) != 16 {
		t.Fatalf("want 16 published assets; got %d", len(m.Assets))
	}
	entries[0]["path"] = "/etc/passwd"
	b, _ = json.Marshal(entries)
	fixtureWrite(t, artifacts, b)
	if _, e = BuildReleaseManifest(artifacts, "v1.17.0", testSHA, fixturePin()); e == nil {
		t.Fatal("accepted external artifact")
	}
}

type publishFixture struct {
	*syncFixture
	selected, checkout, tag string
	existing                bool
	published               bool
	assetBytes              map[string][]byte
	apiStatus               int
}

func newPublishFixture(t *testing.T) *publishFixture {
	f := newSyncFixture(t)
	r, e := f.sync()
	if e != nil {
		t.Fatal(e)
	}
	_ = r
	sha := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/heads/sync/upstream-v1.18.0")
	testGit(t, f.root, "--git-dir", f.fork, "update-ref", "refs/heads/main", sha)
	c := filepath.Join(f.root, "publisher")
	testGit(t, f.root, "clone", f.fork, c)
	return &publishFixture{syncFixture: f, selected: sha, checkout: c, tag: "v1.17.0", assetBytes: map[string][]byte{}}
}
func (f *publishFixture) Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	if name == "gh" {
		joined := strings.Join(args, " ")
		if len(args) > 3 && args[0] == "release" && args[1] == "upload" {
			if f.published {
				return nil, fmt.Errorf("cannot upload to published release")
			}
			b, e := os.ReadFile(args[3])
			if e != nil {
				return nil, e
			}
			f.assetBytes[filepath.Base(args[3])] = b
			return nil, nil
		}
		if strings.Contains(joined, "/releases/tags/") {
			if f.apiStatus != 0 {
				return nil, &CommandError{HTTPStatus: f.apiStatus}
			}
			if !f.existing {
				return nil, &CommandError{HTTPStatus: 404}
			}
			assets := []map[string]any{}
			for i, name := range f.assetNames() {
				assets = append(assets, map[string]any{"name": name, "id": i + 1, "size": len(f.assetBytes[name])})
			}
			return json.Marshal(map[string]any{"id": 42, "tag_name": f.tag, "draft": !f.published, "prerelease": false, "assets": assets})
		}
		if strings.Contains(joined, "/releases/assets/") {
			for _, arg := range args {
				if _, id, ok := strings.Cut(arg, "/releases/assets/"); ok {
					n, e := strconv.Atoi(id)
					if e != nil || n < 1 || n > len(f.assetBytes) {
						return nil, fmt.Errorf("invalid fixture asset")
					}
					return f.assetBytes[f.assetNames()[n-1]], nil
				}
			}
		}
		if strings.Contains(joined, "/releases?") {
			return []byte(`[{"tag_name":"v1.16.0","draft":false,"prerelease":false}]`), nil
		}
	}
	return f.syncFixture.Run(ctx, dir, env, name, args...)
}
func (f *publishFixture) check() (ReleaseResult, error) {
	return CheckRelease(context.Background(), f, f.checkout, ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"})
}
func TestReleaseCheck(t *testing.T) {
	for _, scenario := range []string{"independent stable", "unreviewed SHA", "wrong tag commit", "lower new version", "same tag retry", "fork-only patch", "higher remote version", "API error", "unreachable commit", "wrong ref"} {
		t.Run(scenario, func(t *testing.T) {
			f := newPublishFixture(t)
			opts := ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"}
			switch scenario {
			case "unreviewed SHA":
				opts.Commit = testSHA
			case "wrong tag commit":
				testGit(t, f.root, "--git-dir", f.fork, "tag", f.tag, f.sha)
			case "lower new version":
				opts.Tag = "v1.15.0"
			case "same tag retry":
				testGit(t, f.root, "--git-dir", f.fork, "tag", f.tag, f.selected)
			case "fork-only patch":
				fixtureWrite(t, filepath.Join(f.checkout, "fork-fix"), []byte("patch"))
				testGit(t, f.checkout, "add", ".")
				testGit(t, f.checkout, "commit", "-m", "fix: fork-only patch")
				testGit(t, f.checkout, "push", "origin", "main")
				opts.Commit = testGit(t, f.checkout, "rev-parse", "HEAD")
				opts.Tag = "v1.17.1"
			case "higher remote version":
				testGit(t, f.root, "--git-dir", f.fork, "tag", "v9.0.0", f.selected)
			case "API error":
				f.apiStatus = 403
			case "unreachable commit":
				fixtureWrite(t, filepath.Join(f.checkout, "unreviewed"), []byte("no"))
				testGit(t, f.checkout, "add", ".")
				testGit(t, f.checkout, "commit", "-m", "unreviewed")
				opts.Commit = testGit(t, f.checkout, "rev-parse", "HEAD")
			case "wrong ref":
				opts.Ref = "refs/heads/feature"
			}
			r, e := CheckRelease(context.Background(), f, f.checkout, opts)
			good := scenario == "independent stable" || scenario == "same tag retry" || scenario == "fork-only patch"
			if (e == nil) != good {
				t.Fatalf("%+v %v", r, e)
			}
			if good && r.Pin.Tag != "v1.18.0" {
				t.Fatal("fork version coupled to upstream", r)
			}
		})
	}
}
func TestReleaseTagImmutable(t *testing.T) {
	f := newPublishFixture(t)
	opts := ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"}
	if e := EnsureReleaseTag(context.Background(), f, f.checkout, opts); e != nil {
		t.Fatal(e)
	}
	before := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/tags/"+f.tag)
	if e := EnsureReleaseTag(context.Background(), f, f.checkout, opts); e != nil {
		t.Fatal(e)
	}
	if after := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/tags/"+f.tag); before != after {
		t.Fatal("tag moved")
	}
	opts.Commit = f.sha
	if e := EnsureReleaseTag(context.Background(), f, f.checkout, opts); e == nil {
		t.Fatal("moved immutable tag")
	}
	for _, push := range f.pushes {
		if strings.Contains(push, "--force") || strings.Contains(push, " -f ") {
			t.Fatal(push)
		}
	}
}
func (f *publishFixture) assetNames() []string {
	names := []string{}
	for name := range f.assetBytes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
func TestPublishedReleaseRetry(t *testing.T) {
	for _, scenario := range []string{"complete", "missing manifest", "missing platform", "tampered asset"} {
		t.Run(scenario, func(t *testing.T) {
			f := newPublishFixture(t)
			f.existing = true
			f.published = true
			testGit(t, f.root, "--git-dir", f.fork, "tag", f.tag, f.selected)
			testGit(t, f.root, "--git-dir", f.fork, "tag", "v9.0.0", f.selected)
			checksumName := "az-pim-cli-1.17.0-checksums.txt"
			var checks strings.Builder
			for _, name := range ExpectedAssetNames(f.tag) {
				if name == checksumName {
					continue
				}
				b := []byte(name)
				f.assetBytes[name] = b
				sum := sha256.Sum256(b)
				fmt.Fprintf(&checks, "%x  %s\n", sum, name)
			}
			f.assetBytes[checksumName] = []byte(checks.String())
			m := ReleaseManifest{SchemaVersion: 1, Repository: Fork, Tag: f.tag, Commit: f.selected, Upstream: UpstreamPin{1, Upstream, "v1.18.0", f.sha, releaseURL("v1.18.0")}}
			for _, name := range f.assetNames() {
				b := f.assetBytes[name]
				sum := sha256.Sum256(b)
				m.Assets = append(m.Assets, AssetDigest{name, int64(len(b)), hex.EncodeToString(sum[:])})
			}
			b, _ := json.Marshal(m)
			f.assetBytes[ManifestName] = b
			switch scenario {
			case "missing manifest":
				delete(f.assetBytes, ManifestName)
			case "missing platform":
				delete(f.assetBytes, "az-pim-cli-1.17.0-linux-arm64")
			case "tampered asset":
				f.assetBytes["az-pim-cli-1.17.0-linux-arm64"] = []byte("tamper")
			}
			result, e := f.check()
			if (e == nil) != (scenario == "complete") {
				t.Fatalf("%+v %v", result, e)
			}
			if e == nil && result.State != "already-published" {
				t.Fatal(result)
			}
		})
	}
}

func TestVerifyReleaseDraft(t *testing.T) {
	f := newPublishFixture(t)
	f.existing = true
	testGit(t, f.root, "--git-dir", f.fork, "tag", f.tag, f.selected)
	dist := filepath.Join(f.checkout, "dist")
	if e := os.Mkdir(dist, 0755); e != nil {
		t.Fatal(e)
	}
	checksumName := "az-pim-cli-1.17.0-checksums.txt"
	var checks strings.Builder
	for _, name := range ExpectedAssetNames(f.tag) {
		if name == checksumName {
			continue
		}
		b := []byte(name)
		f.assetBytes[name] = b
		sum := sha256.Sum256(b)
		fmt.Fprintf(&checks, "%x  %s\n", sum, name)
	}
	f.assetBytes[checksumName] = []byte(checks.String())
	entries := []map[string]string{}
	for _, name := range f.assetNames() {
		p := filepath.Join(dist, name)
		fixtureWrite(t, p, f.assetBytes[name])
		entries = append(entries, map[string]string{"name": name, "path": p, "type": "Archive"})
	}
	b, _ := json.Marshal(entries)
	artifacts := filepath.Join(dist, "artifacts.json")
	fixtureWrite(t, artifacts, b)
	opts := ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"}
	if e := VerifyRelease(context.Background(), f, f.checkout, opts, artifacts); e != nil {
		t.Fatal(e)
	}
	if len(f.assetBytes) != 17 || len(f.assetBytes[ManifestName]) == 0 {
		t.Fatal("manifest not uploaded")
	}
	f.published = true
	result, e := f.check()
	if e != nil || result.State != "already-published" {
		t.Fatalf("%+v %v", result, e)
	}
}

func TestChecksums(t *testing.T) {
	b := []byte("content")
	sum := sha256.Sum256(b)
	name := "binary"
	data := []byte(fmt.Sprintf("%x  %s\n", sum, name))
	if e := verifyChecksums(data, map[string][]byte{name: b}); e != nil {
		t.Fatal(e)
	}
	if e := verifyChecksums(data, map[string][]byte{name: []byte("corrupt")}); e == nil {
		t.Fatal("accepted wrong checksum")
	}
	if e := verifyChecksums(data, map[string][]byte{}); e == nil {
		t.Fatal("accepted absent asset")
	}
}
