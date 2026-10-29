package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	b, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("git %v: %s %v", args, b, e)
	}
	return strings.TrimSpace(string(b))
}

type syncFixture struct {
	t                         *testing.T
	root, up, fork, work, sha string
	prs                       []map[string]any
	apiError                  error
	pushes                    []string
	failCreate                bool
}

func newSyncFixture(t *testing.T) *syncFixture {
	t.Helper()
	root := t.TempDir()
	f := &syncFixture{t: t, root: root, up: filepath.Join(root, "up"), fork: filepath.Join(root, "fork.git"), work: root}
	if e := os.Mkdir(f.up, 0755); e != nil {
		t.Fatal(e)
	}
	testGit(t, f.up, "init", "-b", "main")
	os.WriteFile(filepath.Join(f.up, "file"), []byte("base\n"), 0644)
	testGit(t, f.up, "add", ".")
	testGit(t, f.up, "commit", "-m", "base")
	testGit(t, root, "clone", "--bare", f.up, f.fork)
	os.WriteFile(filepath.Join(f.up, "file"), []byte("release\n"), 0644)
	testGit(t, f.up, "add", ".")
	testGit(t, f.up, "commit", "-m", "released change")
	f.sha = testGit(t, f.up, "rev-parse", "HEAD")
	testGit(t, f.up, "tag", "-a", "v1.18.0", "-m", "upstream release")
	os.WriteFile(filepath.Join(f.up, "unreleased"), []byte("must not sync\n"), 0644)
	testGit(t, f.up, "add", ".")
	testGit(t, f.up, "commit", "-m", "unreleased")
	testGit(t, root, "--git-dir", f.fork, "tag", "v99.0.0-fork.1", "main")
	return f
}
func (f *syncFixture) Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	if name == "gh" {
		if f.apiError != nil {
			return nil, f.apiError
		}
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "/releases/latest"):
			return []byte(`{"tag_name":"v1.18.0","html_url":"https://github.com/netr0m/az-pim-cli/releases/tag/v1.18.0","draft":false,"prerelease":false}`), nil
		case len(args) > 0 && args[0] == "api" && strings.Contains(joined, "/pulls?"):
			return json.Marshal(f.prs)
		case len(args) > 1 && args[0] == "pr" && args[1] == "create":
			if f.failCreate {
				return nil, fmt.Errorf("PR creation unavailable")
			}
			branch := "sync/upstream-v1.18.0"
			sha := testGit(f.t, f.root, "--git-dir", f.fork, "rev-parse", "refs/heads/"+branch)
			f.prs = []map[string]any{{"state": "open", "html_url": "https://github.com/" + Fork + "/pull/1", "head": map[string]any{"sha": sha, "ref": branch, "repo": map[string]string{"full_name": Fork}}, "base": map[string]string{"ref": "main"}}}
			return []byte("https://github.com/" + Fork + "/pull/1\n"), nil
		default:
			return nil, fmt.Errorf("unexpected gh request: %v", args)
		}
	}
	copied := append([]string(nil), args...)
	for i, a := range copied {
		if a == "https://github.com/"+Upstream+".git" {
			copied[i] = f.up
		}
		if a == "https://github.com/"+Fork+".git" {
			copied[i] = f.fork
		}
	}
	if name == "git" && containsArg(copied, "push") {
		f.pushes = append(f.pushes, strings.Join(copied, " "))
	}
	env = append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	return (ExecRunner{}).Run(ctx, dir, env, name, copied...)
}
func containsArg(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}
func (f *syncFixture) sync() (SyncResult, error) {
	return RunSync(context.Background(), f, SyncOptions{Repository: Fork, Upstream: Upstream, Base: "main", WorkDir: f.work})
}
func TestRunSync(t *testing.T) {
	t.Run("release not main and repeat reuses", func(t *testing.T) {
		f := newSyncFixture(t)
		before := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "main")
		tagBefore := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/tags/v99.0.0-fork.1")
		got, e := f.sync()
		if e != nil {
			t.Fatal(e)
		}
		if got.State != "created" || got.Pin.Commit != f.sha {
			t.Fatalf("%+v", got)
		}
		branch := "refs/heads/sync/upstream-v1.18.0"
		sha := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", branch)
		tree := testGit(t, f.root, "--git-dir", f.fork, "ls-tree", "-r", "--name-only", branch)
		if strings.Contains(tree, "unreleased") {
			t.Fatal(tree)
		}
		if testGit(t, f.root, "--git-dir", f.fork, "show", branch+":file") != "release" {
			t.Fatal("wrong source")
		}
		again, e := f.sync()
		if e != nil || again.State != "reused" {
			t.Fatalf("%+v %v", again, e)
		}
		if testGit(t, f.root, "--git-dir", f.fork, "rev-parse", branch) != sha || len(f.pushes) != 1 {
			t.Fatal("rewrote candidate")
		}
		if testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "main") != before || testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/tags/v99.0.0-fork.1") != tagBefore {
			t.Fatal("modified main/tags")
		}
	})
	t.Run("API failure never pushes", func(t *testing.T) {
		for _, status := range []int{401, 403, 429, 500} {
			f := newSyncFixture(t)
			f.apiError = &CommandError{HTTPStatus: status}
			if _, e := f.sync(); e == nil {
				t.Fatal("accepted API failure")
			}
			if len(f.pushes) != 0 {
				t.Fatal("pushed on API failure")
			}
		}
	})
	t.Run("closed PR stops", func(t *testing.T) {
		f := newSyncFixture(t)
		if _, e := f.sync(); e != nil {
			t.Fatal(e)
		}
		f.prs[0]["state"] = "closed"
		f.pushes = nil
		if _, e := f.sync(); e == nil {
			t.Fatal("accepted closed PR")
		}
		if len(f.pushes) != 0 {
			t.Fatal("rewrote closed candidate")
		}
	})
	t.Run("conflict never pushes", func(t *testing.T) {
		f := newSyncFixture(t)
		c := filepath.Join(f.root, "custom")
		testGit(t, f.root, "clone", f.fork, c)
		os.WriteFile(filepath.Join(c, "file"), []byte("fork conflict\n"), 0644)
		testGit(t, c, "add", ".")
		testGit(t, c, "commit", "-m", "customization")
		testGit(t, c, "push", "origin", "main")
		if _, e := f.sync(); e == nil {
			t.Fatal("accepted conflict")
		}
		if len(f.pushes) != 0 {
			t.Fatal("pushed conflict")
		}
	})
	t.Run("recover branch after PR API failure", func(t *testing.T) {
		f := newSyncFixture(t)
		f.failCreate = true
		if _, e := f.sync(); e == nil {
			t.Fatal("expected PR failure")
		}
		f.failCreate = false
		r, e := f.sync()
		if e != nil || r.State != "created" || len(f.pushes) != 1 {
			t.Fatalf("%+v %v pushes=%v", r, e, f.pushes)
		}
	})
	t.Run("metadata-only adoption and no-op", func(t *testing.T) {
		f := newSyncFixture(t)
		testGit(t, f.root, "--git-dir", f.fork, "fetch", "--no-tags", f.up, f.sha+":refs/heads/main")
		r, e := f.sync()
		if e != nil {
			t.Fatal(e)
		}
		if r.State != "created" {
			t.Fatal(r)
		}
		branch := "refs/heads/sync/upstream-v1.18.0"
		sha := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", branch)
		testGit(t, f.root, "--git-dir", f.fork, "update-ref", "refs/heads/main", sha)
		r, e = f.sync()
		if e != nil || r.State != "noop" {
			t.Fatalf("%+v %v", r, e)
		}
	})
	t.Run("retargeted tag rejected", func(t *testing.T) {
		f := newSyncFixture(t)
		if _, e := f.sync(); e != nil {
			t.Fatal(e)
		}
		sha := testGit(t, f.root, "--git-dir", f.fork, "rev-parse", "refs/heads/sync/upstream-v1.18.0")
		testGit(t, f.root, "--git-dir", f.fork, "update-ref", "refs/heads/main", sha)
		testGit(t, f.up, "tag", "-f", "v1.18.0", "HEAD")
		if _, e := f.sync(); e == nil {
			t.Fatal("accepted retargeted tag")
		}
	})
}
func TestExecRunnerRedactsCredentials(t *testing.T) {
	t.Setenv("GH_TOKEN", "fixture-secret")
	_, e := (ExecRunner{}).Run(context.Background(), t.TempDir(), nil, "sh", "-c", "echo fixture-secret >&2; exit 1")
	if e == nil || strings.Contains(e.Error(), "fixture-secret") {
		t.Fatalf("credential leaked: %v", e)
	}
}
