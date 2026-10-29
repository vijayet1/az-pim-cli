package automation

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const oldHomebrewFormula = `class AzPimCli < Formula
  desc "Preserve custom description"
  homepage "https://github.com/vijayet1/az-pim-cli"
  version "1.16.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/vijayet1/az-pim-cli/releases/download/v1.16.0/az-pim-cli-1.16.0-macos-arm64.tar.gz"
      sha256 "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    end
    if Hardware::CPU.intel?
      url "https://github.com/vijayet1/az-pim-cli/releases/download/v1.16.0/az-pim-cli-1.16.0-macos-x86_64.tar.gz"
      sha256 "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    end
  end

  def install
    bin.install "az-pim-cli"
  end

  test do
    system "#{bin}/az-pim-cli", "--help"
  end
end
`

func TestHomebrewFormulaUpdate(t *testing.T) {
	plan := HomebrewUpdate{Tag: "v1.17.0", Commit: testSHA, Arm64SHA256: strings.Repeat("c", 64), AMD64SHA256: strings.Repeat("d", 64)}
	got, err := UpdateHomebrewFormula([]byte(oldHomebrewFormula), plan)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.ReplaceAll(oldHomebrewFormula, "1.16.0", "1.17.0")
	want = strings.ReplaceAll(want, strings.Repeat("a", 64), strings.Repeat("c", 64))
	want = strings.ReplaceAll(want, strings.Repeat("b", 64), strings.Repeat("d", 64))
	if string(got) != want {
		t.Fatalf("unexpected formula:\n%s", got)
	}
	again, err := UpdateHomebrewFormula(got, plan)
	if err != nil || string(again) != want {
		t.Fatalf("not idempotent: %v", err)
	}
}
func TestHomebrewFormulaRefusesUnsafeUpdates(t *testing.T) {
	for _, scenario := range []string{"downgrade", "same-version change", "wrong repository", "missing arch", "duplicate version", "malformed digest", "invalid tag"} {
		t.Run(scenario, func(t *testing.T) {
			p := HomebrewUpdate{Tag: "v1.17.0", Commit: testSHA, Arm64SHA256: strings.Repeat("c", 64), AMD64SHA256: strings.Repeat("d", 64)}
			text := oldHomebrewFormula
			switch scenario {
			case "downgrade":
				p.Tag = "v1.15.0"
			case "same-version change":
				p.Tag = "v1.16.0"
			case "wrong repository":
				text = strings.Replace(text, "/releases/download/v1.16.0/", "/wrong/download/v1.16.0/", 1)
			case "missing arch":
				text = strings.Replace(text, "macos-x86_64.tar.gz", "macos-unknown.tar.gz", 1)
			case "duplicate version":
				text += "  version \"1.16.0\"\n"
			case "malformed digest":
				p.Arm64SHA256 = "bad"
			case "invalid tag":
				p.Tag = "v1.17.0\nputs :bad"
			}
			if _, err := UpdateHomebrewFormula([]byte(text), p); err == nil {
				t.Fatal("accepted unsafe update")
			}
		})
	}
}
func TestHomebrewPlanRequiresCompleteManifest(t *testing.T) {
	m := ReleaseManifest{SchemaVersion: 1, Repository: Fork, Tag: "v1.17.0", Commit: testSHA, Upstream: fixturePin()}
	for _, name := range ExpectedAssetNames(m.Tag) {
		m.Assets = append(m.Assets, AssetDigest{name, 1, strings.Repeat("a", 64)})
	}
	p, err := HomebrewPlanFromManifest(m)
	if err != nil || p.Tag != "v1.17.0" || p.Arm64SHA256 != strings.Repeat("a", 64) {
		t.Fatalf("%+v %v", p, err)
	}
	m.Assets = m.Assets[1:]
	if _, err = HomebrewPlanFromManifest(m); err == nil {
		t.Fatal("accepted incomplete release")
	}
}

type homebrewFixture struct {
	*publishFixture
	formula   string
	sha       string
	conflicts int
	writes    int
	readOnly  bool
}

func newHomebrewFixture(t *testing.T) *homebrewFixture {
	f := newPublishFixture(t)
	f.existing = true
	f.published = true
	testGit(t, f.root, "--git-dir", f.fork, "tag", f.tag, f.selected)
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
	return &homebrewFixture{publishFixture: f, formula: oldHomebrewFormula, sha: strings.Repeat("a", 40)}
}
func (f *homebrewFixture) Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	path := "repos/vijayet1/homebrew-tap/contents/Formula/az-pim-cli.rb"
	if name == "gh" && containsArg(args, path) {
		if !containsArg(env, "GH_TOKEN=tap-fixture-token") {
			return nil, fmt.Errorf("tap call did not use dedicated credential")
		}
		if strings.Contains(strings.Join(args, " "), "tap-fixture-token") {
			return nil, fmt.Errorf("credential exposed in command arguments")
		}
		if containsArg(args, "PUT") {
			if f.readOnly {
				return nil, &CommandError{HTTPStatus: 403}
			}
			fields := map[string]string{}
			for _, arg := range args {
				key, value, ok := strings.Cut(arg, "=")
				if ok {
					fields[key] = value
				}
			}
			if fields["branch"] != "main" || fields["sha"] != f.sha {
				return nil, fmt.Errorf("wrong target branch or stale content SHA")
			}
			if f.conflicts > 0 {
				f.conflicts--
				f.sha = strings.Repeat("b", 40)
				return nil, &CommandError{HTTPStatus: 409}
			}
			decoded, err := base64.StdEncoding.DecodeString(fields["content"])
			if err != nil {
				return nil, err
			}
			f.formula = string(decoded)
			f.writes++
			return []byte(`{"commit":{"sha":"cccccccccccccccccccccccccccccccccccccccc"}}`), nil
		}
		return json.Marshal(map[string]string{"type": "file", "encoding": "base64", "sha": f.sha, "content": base64.StdEncoding.EncodeToString([]byte(f.formula))})
	}
	return f.publishFixture.Run(ctx, dir, env, name, args...)
}
func TestHomebrewCredentialIsIsolatedFromChildProcesses(t *testing.T) {
	dir := t.TempDir()
	script := `#!/bin/sh
if [ -n "$HOMEBREW_TAP_TOKEN" ]; then echo unexpected-private-env; exit 1; fi
case "$GH_TOKEN" in
 source-fixture-token) printf source ;;
 tap-fixture-token) printf target ;;
 *) exit 2 ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOMEBREW_TAP_TOKEN", "private-fixture-token")
	t.Setenv("GH_TOKEN", "source-fixture-token")
	for _, tc := range []struct {
		env  []string
		want string
	}{{nil, "source"}, {[]string{"GH_TOKEN=tap-fixture-token"}, "target"}} {
		out, err := (ExecRunner{}).Run(context.Background(), dir, tc.env, "gh")
		if err != nil || string(out) != tc.want {
			t.Fatalf("credential boundary failed: output=%q error=%v", out, err)
		}
	}
}

func TestHomebrewTapUpdate(t *testing.T) {
	for _, scenario := range []string{"update", "conflict retry", "conflicts exhausted", "protected tap", "draft", "missing token", "bad metadata"} {
		t.Run(scenario, func(t *testing.T) {
			f := newHomebrewFixture(t)
			t.Setenv("HOMEBREW_TAP_TOKEN", "tap-fixture-token")
			switch scenario {
			case "conflict retry":
				f.conflicts = 1
			case "conflicts exhausted":
				f.conflicts = 4
			case "bad metadata":
				f.sha = "invalid"
			case "protected tap":
				f.readOnly = true
			case "draft":
				f.published = false
			case "missing token":
				t.Setenv("HOMEBREW_TAP_TOKEN", "")
			}
			options := ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"}
			result, err := UpdateHomebrewTap(context.Background(), f, f.checkout, options)
			success := scenario == "update" || scenario == "conflict retry"
			if (err == nil) != success {
				t.Fatalf("%+v %v", result, err)
			}
			if success {
				if result.State != "updated" || f.writes != 1 || !strings.Contains(f.formula, `version "1.17.0"`) {
					t.Fatalf("%+v writes=%d", result, f.writes)
				}
				result, err = UpdateHomebrewTap(context.Background(), f, f.checkout, options)
				if err != nil || result.State != "noop" || f.writes != 1 {
					t.Fatalf("retry wrote formula: %+v %v", result, err)
				}
			} else if f.writes != 0 {
				t.Fatal("wrote formula after failed precondition")
			}
			if len(f.pushes) != 1 {
				t.Fatal("tap repair pushed a source branch or tag")
			}
			if scenario == "conflicts exhausted" && f.conflicts != 1 {
				t.Fatal("concurrent update retry was not bounded to three attempts")
			}
		})
	}
}
