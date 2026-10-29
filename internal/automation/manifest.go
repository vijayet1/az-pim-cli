package automation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ManifestName = "fork-release-manifest.json"

type AssetDigest struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type ReleaseManifest struct {
	SchemaVersion int           `json:"schema_version"`
	Repository    string        `json:"repository"`
	Tag           string        `json:"tag"`
	Commit        string        `json:"commit"`
	Upstream      UpstreamPin   `json:"upstream"`
	Assets        []AssetDigest `json:"assets"`
}

// ExpectedAssetNames is the publishing contract for .goreleaser.fork.yaml.
// A non-publishing GoReleaser rehearsal must verify it whenever the config changes.
func ExpectedAssetNames(tag string) []string {
	prefix := "az-pim-cli-" + strings.TrimPrefix(tag, "v") + "-"
	names := []string{}
	for _, target := range []string{"linux-x86_64", "linux-arm64", "windows-x86_64", "windows-arm64", "macos-x86_64", "macos-arm64", "macos-all"} {
		raw, archive := prefix+target, prefix+target+".tar.gz"
		if strings.HasPrefix(target, "windows-") {
			raw += ".exe"
			archive = prefix + target + ".zip"
		}
		names = append(names, raw, archive)
	}
	names = append(names, prefix+"source.tar.gz", prefix+"checksums.txt")
	sort.Strings(names)
	return names
}
func ValidateManifest(m ReleaseManifest, tag, commit string, pin UpstreamPin) error {
	if e := ValidatePin(pin); e != nil {
		return e
	}
	if _, e := ParseForkVersion(tag); e != nil {
		return e
	}
	if m.SchemaVersion != 1 || m.Repository != Fork || m.Tag != tag || m.Commit != commit || !shaPattern.MatchString(commit) || m.Upstream != pin {
		return fmt.Errorf("release manifest source identity mismatch")
	}
	expected := map[string]bool{}
	for _, name := range ExpectedAssetNames(tag) {
		expected[name] = true
	}
	if len(m.Assets) != len(expected) {
		return fmt.Errorf("release manifest does not cover all configured outputs")
	}
	for _, a := range m.Assets {
		if !expected[a.Name] || a.Size <= 0 || len(a.SHA256) != 64 {
			return fmt.Errorf("invalid or duplicate asset %q", a.Name)
		}
		b, e := hex.DecodeString(a.SHA256)
		if e != nil || len(b) != 32 || strings.ToLower(a.SHA256) != a.SHA256 {
			return fmt.Errorf("invalid asset digest")
		}
		delete(expected, a.Name)
	}
	return nil
}
func VerifyAssetBytes(a AssetDigest, b []byte) error {
	sum := sha256.Sum256(b)
	if a.Size != int64(len(b)) || a.SHA256 != hex.EncodeToString(sum[:]) {
		return fmt.Errorf("asset size/digest mismatch: %s", a.Name)
	}
	return nil
}
func BuildReleaseManifest(artifactFile, tag, commit string, pin UpstreamPin) (ReleaseManifest, error) {
	m := ReleaseManifest{SchemaVersion: 1, Repository: Fork, Tag: tag, Commit: commit, Upstream: pin}
	b, e := os.ReadFile(artifactFile)
	if e != nil {
		return m, e
	}
	var entries []struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if e = json.Unmarshal(b, &entries); e != nil {
		return m, e
	}
	root, e := filepath.Abs(filepath.Dir(artifactFile))
	if e != nil {
		return m, e
	}
	root, e = filepath.EvalSymlinks(root)
	if e != nil {
		return m, e
	}
	expected := map[string]bool{}
	for _, name := range ExpectedAssetNames(tag) {
		expected[name] = true
	}
	seen := map[string]bool{}
	for _, a := range entries {
		if !expected[a.Name] {
			continue
		}
		if seen[a.Name] {
			return m, fmt.Errorf("duplicate published artifact %s", a.Name)
		}
		p, e := filepath.Abs(a.Path)
		if e != nil {
			return m, e
		}
		p, e = filepath.EvalSymlinks(p)
		if e != nil {
			return m, e
		}
		rel, e := filepath.Rel(root, p)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return m, fmt.Errorf("artifact escapes dist directory")
		}
		info, e := os.Stat(p)
		if e != nil || !info.Mode().IsRegular() {
			return m, fmt.Errorf("artifact must be a regular file")
		}
		contents, e := os.ReadFile(p)
		if e != nil {
			return m, e
		}
		sum := sha256.Sum256(contents)
		m.Assets = append(m.Assets, AssetDigest{a.Name, int64(len(contents)), hex.EncodeToString(sum[:])})
		seen[a.Name] = true
	}
	sort.Slice(m.Assets, func(i, j int) bool { return m.Assets[i].Name < m.Assets[j].Name })
	return m, ValidateManifest(m, tag, commit, pin)
}
func verifyChecksums(data []byte, files map[string][]byte) error {
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != 64 {
			return fmt.Errorf("invalid checksum line")
		}
		name := strings.TrimPrefix(fields[1], "*")
		b, ok := files[name]
		if !ok || seen[name] {
			return fmt.Errorf("missing/duplicate checksum asset %q", name)
		}
		sum := sha256.Sum256(b)
		if fields[0] != hex.EncodeToString(sum[:]) {
			return fmt.Errorf("checksum mismatch: %s", name)
		}
		seen[name] = true
	}
	for name := range files {
		if name != ManifestName && !strings.HasSuffix(name, "-checksums.txt") && !seen[name] {
			return fmt.Errorf("asset missing checksum: %s", name)
		}
	}
	return nil
}
