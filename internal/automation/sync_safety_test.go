package automation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncRejectsSourcePinSymlink(t *testing.T) {
	f := newSyncFixture(t)
	outside := filepath.Join(f.root, "outside")
	fixtureWrite(t, outside, []byte("unchanged"))
	fixtureDirectory(t, filepath.Join(f.up, ".github"))
	if e := os.Symlink(outside, filepath.Join(f.up, PinPath)); e != nil {
		t.Fatal(e)
	}
	testGit(t, f.up, "add", ".")
	testGit(t, f.up, "commit", "-m", "malformed upstream pin")
	f.sha = testGit(t, f.up, "rev-parse", "HEAD")
	testGit(t, f.up, "tag", "-f", "v1.18.0", "HEAD")
	if _, e := f.sync(); e == nil {
		t.Error("accepted symlink pin")
	}
	b, e := os.ReadFile(outside)
	if e != nil || string(b) != "unchanged" {
		t.Fatal("candidate write escaped checkout")
	}
	if len(f.pushes) != 0 {
		t.Fatal("pushed invalid candidate")
	}
}
func TestSyncReportsOlderProposals(t *testing.T) {
	f := newSyncFixture(t)
	f.prs = []map[string]any{{"state": "open", "html_url": "https://github.com/" + Fork + "/pull/77", "head": map[string]any{"sha": testSHA, "ref": "sync/upstream-v1.17.0", "repo": map[string]string{"full_name": Fork}}, "base": map[string]string{"ref": "main"}}}
	result, e := f.sync()
	if e != nil {
		t.Fatal(e)
	}
	if len(result.OlderPRs) != 1 || result.OlderPRs[0] != "https://github.com/"+Fork+"/pull/77" {
		t.Fatalf("older proposals missing: %+v", result)
	}
}
func TestSyncRejectsModifiedCandidate(t *testing.T) {
	f := newSyncFixture(t)
	if _, e := f.sync(); e != nil {
		t.Fatal(e)
	}
	c := filepath.Join(f.root, "edit")
	testGit(t, f.root, "clone", f.fork, c)
	testGit(t, c, "checkout", "sync/upstream-v1.18.0")
	fixtureWrite(t, filepath.Join(c, "unexpected"), []byte("extra change"))
	testGit(t, c, "add", ".")
	testGit(t, c, "commit", "-m", "unexpected candidate modification")
	testGit(t, c, "push", "origin", "HEAD")
	f.prs[0]["head"].(map[string]any)["sha"] = testGit(t, c, "rev-parse", "HEAD")
	f.pushes = nil
	if _, e := f.sync(); e == nil {
		t.Fatal("accepted modified candidate")
	}
	if len(f.pushes) != 0 {
		t.Fatal("rewrote modified candidate")
	}
}
