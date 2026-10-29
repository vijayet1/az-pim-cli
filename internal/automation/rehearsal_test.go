package automation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// This opt-in integration test consumes real output from the pinned OSS builder.
// Normal unit tests do not require GoReleaser or generate release artifacts.
func TestGoReleaserArtifactContract(t *testing.T) {
	artifacts := os.Getenv("FORK_REHEARSAL_ARTIFACTS")
	if artifacts == "" {
		t.Skip("set FORK_REHEARSAL_ARTIFACTS to a temporary stable-tag rehearsal's dist/artifacts.json")
	}
	artifacts, e := filepath.Abs(artifacts)
	if e != nil {
		t.Fatal(e)
	}
	root := filepath.Dir(filepath.Dir(artifacts))
	t.Chdir(root)
	sha := testGit(t, root, "rev-parse", "HEAD")
	m, e := BuildReleaseManifest(artifacts, "v1.17.0", sha, fixturePin())
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Assets) != 16 {
		t.Fatalf("want 16 builder outputs, got %d", len(m.Assets))
	}
	b, e := os.ReadFile(artifacts)
	if e != nil {
		t.Fatal(e)
	}
	var entries []struct{ Name, Path string }
	if e = json.Unmarshal(b, &entries); e != nil {
		t.Fatal(e)
	}
	files := map[string][]byte{}
	for _, a := range m.Assets {
		for _, entry := range entries {
			if entry.Name == a.Name {
				b, e := os.ReadFile(entry.Path)
				if e != nil {
					t.Fatal(e)
				}
				if e = VerifyAssetBytes(a, b); e != nil {
					t.Fatal(e)
				}
				files[a.Name] = b
			}
		}
	}
	if e = verifyChecksums(files["az-pim-cli-1.17.0-checksums.txt"], files); e != nil {
		t.Fatal(e)
	}
}
