package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ReleaseOptions struct{ Repository, Tag, Commit, Ref string }
type ReleaseResult struct {
	State  string      `json:"state"`
	Tag    string      `json:"tag"`
	Commit string      `json:"commit"`
	Pin    UpstreamPin `json:"pin"`
}
type releaseAsset struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
	Size int64  `json:"size"`
}
type githubRelease struct {
	ID         int64          `json:"id"`
	Tag        string         `json:"tag_name"`
	Draft      *bool          `json:"draft"`
	Prerelease *bool          `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

func (c commands) release(tag string) (*githubRelease, error) {
	b, e := c.api("repos/" + Fork + "/releases/tags/" + tag)
	if is404(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var rel githubRelease
	if e = json.Unmarshal(b, &rel); e != nil {
		return nil, e
	}
	if rel.ID <= 0 || rel.Tag != tag || rel.Draft == nil || rel.Prerelease == nil || *rel.Prerelease {
		return nil, fmt.Errorf("invalid stable fork release response")
	}
	return &rel, nil
}
func (c commands) remoteTags() (map[string]string, error) {
	raw, e := c.git("ls-remote", "--tags", "https://github.com/"+Fork+".git")
	if e != nil {
		return nil, e
	}
	tags := map[string]string{}
	peeled := map[string]bool{}
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || !shaPattern.MatchString(fields[0]) || !strings.HasPrefix(fields[1], "refs/tags/") {
			return nil, fmt.Errorf("invalid remote tag response")
		}
		tag := strings.TrimPrefix(fields[1], "refs/tags/")
		p := strings.HasSuffix(tag, "^{}")
		tag = strings.TrimSuffix(tag, "^{}")
		if p || !peeled[tag] {
			tags[tag] = fields[0]
		}
		if p {
			peeled[tag] = true
		}
	}
	return tags, nil
}
func (c commands) checkVersion(tag string, tags map[string]string) error {
	selected, e := ParseForkVersion(tag)
	if e != nil {
		return e
	}
	check := func(existing string) error {
		v, e := ParseForkVersion(existing)
		if e != nil {
			if versionPattern.MatchString(existing) {
				return fmt.Errorf("cannot order existing stable version %s: %w", existing, e)
			}
			return nil
		}
		if existing == tag {
			return nil
		}
		if CompareVersions(selected, v) <= 0 {
			return fmt.Errorf("fork version %s does not exceed existing stable version %s", tag, existing)
		}
		return nil
	}
	for existing := range tags {
		if e = check(existing); e != nil {
			return e
		}
	}
	for page := 1; page <= 1000; page++ {
		b, e := c.api("repos/" + Fork + "/releases?per_page=100&page=" + strconv.Itoa(page))
		if e != nil {
			return e
		}
		var releases []struct {
			Tag   string `json:"tag_name"`
			Draft *bool  `json:"draft"`
			Pre   *bool  `json:"prerelease"`
		}
		if e = json.Unmarshal(b, &releases); e != nil {
			return e
		}
		for _, rel := range releases {
			if rel.Draft == nil || rel.Pre == nil {
				return fmt.Errorf("release list missing state")
			}
			if !*rel.Draft && !*rel.Pre {
				if e = check(rel.Tag); e != nil {
					return e
				}
			}
		}
		if len(releases) < 100 {
			return nil
		}
	}
	return fmt.Errorf("release pagination limit exceeded")
}
func (c commands) downloadAsset(asset releaseAsset) ([]byte, error) {
	if asset.ID <= 0 {
		return nil, fmt.Errorf("invalid release asset ID")
	}
	return c.gh("api", "--hostname", "github.com", "repos/"+Fork+"/releases/assets/"+strconv.FormatInt(asset.ID, 10), "-H", "Accept: application/octet-stream")
}
func (c commands) verifyRemote(rel *githubRelease, m ReleaseManifest, requireManifest bool) error {
	assets := map[string]releaseAsset{}
	for _, a := range rel.Assets {
		if _, exists := assets[a.Name]; exists {
			return fmt.Errorf("duplicate remote asset")
		}
		assets[a.Name] = a
	}
	wantCount := len(m.Assets)
	if requireManifest {
		wantCount++
	}
	if len(assets) != wantCount {
		return fmt.Errorf("release asset set is incomplete or unexpected")
	}
	files := map[string][]byte{}
	for _, a := range m.Assets {
		remote, ok := assets[a.Name]
		if !ok || remote.Size != a.Size {
			return fmt.Errorf("missing or wrong-size asset %s", a.Name)
		}
		b, e := c.downloadAsset(remote)
		if e != nil {
			return e
		}
		if e = VerifyAssetBytes(a, b); e != nil {
			return e
		}
		files[a.Name] = b
	}
	checksumName := "az-pim-cli-" + strings.TrimPrefix(m.Tag, "v") + "-checksums.txt"
	if e := verifyChecksums(files[checksumName], files); e != nil {
		return e
	}
	if requireManifest {
		a, ok := assets[ManifestName]
		if !ok {
			return fmt.Errorf("persisted release manifest missing")
		}
		b, e := c.downloadAsset(a)
		if e != nil {
			return e
		}
		var persisted ReleaseManifest
		if e = json.Unmarshal(b, &persisted); e != nil {
			return e
		}
		if e = ValidateManifest(persisted, m.Tag, m.Commit, m.Upstream); e != nil {
			return e
		}
		expected, _ := json.Marshal(m)
		actual, _ := json.Marshal(persisted)
		if string(expected) != string(actual) {
			return fmt.Errorf("persisted manifest differs from verified assets")
		}
	}
	return nil
}
func (c commands) verifyPublished(rel *githubRelease, o ReleaseOptions, pin UpstreamPin) error {
	var asset *releaseAsset
	for i := range rel.Assets {
		if rel.Assets[i].Name == ManifestName {
			asset = &rel.Assets[i]
			break
		}
	}
	if asset == nil {
		return fmt.Errorf("published release has no verification manifest; operator inspection required")
	}
	b, e := c.downloadAsset(*asset)
	if e != nil {
		return e
	}
	var m ReleaseManifest
	if e = json.Unmarshal(b, &m); e != nil {
		return e
	}
	if e = ValidateManifest(m, o.Tag, o.Commit, pin); e != nil {
		return e
	}
	return c.verifyRemote(rel, m, true)
}
func CheckRelease(ctx context.Context, r Runner, dir string, o ReleaseOptions) (ReleaseResult, error) {
	result := ReleaseResult{Tag: o.Tag, Commit: o.Commit}
	if o.Repository != Fork || o.Ref != "refs/heads/main" || !shaPattern.MatchString(o.Commit) {
		return result, fmt.Errorf("release requires fork main and a full lowercase commit SHA")
	}
	if _, e := ParseForkVersion(o.Tag); e != nil {
		return result, e
	}
	c := commands{ctx, r, dir}
	head, e := c.git("rev-parse", "HEAD")
	if e != nil || head != o.Commit {
		return result, fmt.Errorf("checkout is not the selected commit")
	}
	if _, e = c.git("fetch", "--no-tags", "https://github.com/"+Fork+".git", "refs/heads/main:refs/remotes/origin/main"); e != nil {
		return result, e
	}
	ok, e := c.ancestor(o.Commit, "refs/remotes/origin/main")
	if e != nil || !ok {
		return result, fmt.Errorf("selected commit is not reviewed fork main history")
	}
	pin, e := c.pinAt(o.Commit)
	if e != nil || pin == nil {
		return result, fmt.Errorf("selected commit lacks valid upstream pin: %v", e)
	}
	if e = c.verifyPin(*pin, o.Commit); e != nil {
		return result, e
	}
	result.Pin = *pin
	tags, e := c.remoteTags()
	if e != nil {
		return result, e
	}
	tagCommit := tags[o.Tag]
	if tagCommit != "" && tagCommit != o.Commit {
		return result, fmt.Errorf("immutable fork tag points to another commit")
	}
	rel, e := c.release(o.Tag)
	if e != nil {
		return result, e
	}
	published := rel != nil && !*rel.Draft
	if published {
		if e = c.verifyPublished(rel, o, *pin); e != nil {
			return result, e
		}
	} else {
		if e = c.checkVersion(o.Tag, tags); e != nil {
			return result, e
		}
	}
	result.State, e = ClassifyRelease(tagCommit, o.Commit, published, published)
	return result, e
}
func EnsureReleaseTag(ctx context.Context, r Runner, dir string, o ReleaseOptions) error {
	result, e := CheckRelease(ctx, r, dir, o)
	if e != nil {
		return e
	}
	if result.State == "already-published" {
		return nil
	}
	c := commands{ctx, r, dir}
	tags, e := c.remoteTags()
	if e != nil {
		return e
	}
	if tags[o.Tag] == o.Commit {
		return nil
	}
	if tags[o.Tag] != "" {
		return fmt.Errorf("immutable tag mismatch")
	}
	local, e := c.git("show-ref", "--verify", "--quiet", "refs/tags/"+o.Tag)
	_ = local
	if e == nil {
		sha, e := c.git("rev-parse", "refs/tags/"+o.Tag+"^{commit}")
		if e != nil || sha != o.Commit {
			return fmt.Errorf("local tag mismatch")
		}
	} else if exitCode(e, 1) {
		if _, e = c.git("tag", "-a", o.Tag, o.Commit, "-m", "Fork release "+o.Tag); e != nil {
			return e
		}
	} else {
		return e
	}
	if e = c.push("https://github.com/"+Fork+".git", "refs/tags/"+o.Tag+":refs/tags/"+o.Tag); e != nil {
		tags, readErr := c.remoteTags()
		if readErr != nil || tags[o.Tag] != o.Commit {
			return e
		}
	}
	return nil
}
func VerifyRelease(ctx context.Context, r Runner, dir string, o ReleaseOptions, artifacts string) error {
	result, e := CheckRelease(ctx, r, dir, o)
	if e != nil {
		return e
	}
	if result.State == "already-published" {
		return nil
	}
	c := commands{ctx, r, dir}
	rel, e := c.release(o.Tag)
	if e != nil || rel == nil || !*rel.Draft {
		return fmt.Errorf("expected existing draft before asset verification: %v", e)
	}
	m, e := BuildReleaseManifest(artifacts, o.Tag, o.Commit, result.Pin)
	if e != nil {
		return e
	}
	if e = c.verifyRemote(rel, m, false); e != nil {
		return e
	}
	temp, e := os.MkdirTemp("", "fork-release-manifest-")
	if e != nil {
		return e
	}
	defer func() { _ = os.RemoveAll(temp) }()
	b, _ := json.MarshalIndent(m, "", "  ")
	path := filepath.Join(temp, ManifestName)
	if e = os.WriteFile(path, append(b, '\n'), 0600); e != nil {
		return e
	}
	if _, e = c.gh("release", "upload", o.Tag, path, "--repo", Fork); e != nil {
		return e
	}
	rel, e = c.release(o.Tag)
	if e != nil || rel == nil || !*rel.Draft {
		return fmt.Errorf("draft state changed during verification")
	}
	return c.verifyRemote(rel, m, true)
}
