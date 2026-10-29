package automation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCandidateRejectsSymbolicPin(t *testing.T) {
	f := newSyncFixture(t)
	if _, e := f.sync(); e != nil {
		t.Fatal(e)
	}
	c := filepath.Join(f.root, "symbolic")
	testGit(t, f.root, "clone", f.fork, c)
	testGit(t, c, "checkout", "sync/upstream-v1.18.0")
	pin, e := os.ReadFile(filepath.Join(c, PinPath))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Remove(filepath.Join(c, PinPath)); e != nil {
		t.Fatal(e)
	}
	if e = os.Symlink(string(pin), filepath.Join(c, PinPath)); e != nil {
		t.Fatal(e)
	}
	testGit(t, c, "add", ".")
	testGit(t, c, "commit", "-m", "change pin to symlink")
	if e = ValidateCandidate(context.Background(), f, c, "refs/remotes/origin/main"); e == nil {
		t.Fatal("candidate accepted symbolic pin containing JSON")
	}
}
func TestCandidateRejectsRetargetedBasePin(t *testing.T) {
	f := newSyncFixture(t)
	if _, e := f.sync(); e != nil {
		t.Fatal(e)
	}
	base := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/heads/sync/upstream-v1.18.0")
	testGit(t, f.root, "--git-dir", f.fork, "update-ref", "refs/heads/main", base)
	c := filepath.Join(f.root, "newer")
	testGit(t, f.root, "clone", f.fork, c)
	newSHA := testGit(t, f.up, "rev-parse", "HEAD")
	testGit(t, f.up, "tag", "v1.19.0", newSHA)
	testGit(t, c, "fetch", "--no-tags", f.up, newSHA)
	testGit(t, c, "merge", "--no-ff", "FETCH_HEAD", "-m", "merge newer release")
	pin := UpstreamPin{1, Upstream, "v1.19.0", newSHA, releaseURL("v1.19.0")}
	b, _ := json.Marshal(pin)
	fixtureWrite(t, filepath.Join(c, PinPath), b)
	testGit(t, c, "add", ".")
	testGit(t, c, "commit", "-m", "record newer source")
	testGit(t, f.up, "tag", "-f", "v1.18.0", newSHA)
	if e := ValidateCandidate(context.Background(), f, c, "refs/remotes/origin/main"); e == nil {
		t.Fatal("candidate accepted retargeted base source")
	}
}
func TestReleaseRejectsUnorderableStableTag(t *testing.T) {
	f := newPublishFixture(t)
	testGit(t, f.root, "--git-dir", f.fork, "tag", "v18446744073709551616.0.0", f.selected)
	if _, e := f.check(); e == nil {
		t.Fatal("published without ordering all existing stable-shaped tags")
	}
}
