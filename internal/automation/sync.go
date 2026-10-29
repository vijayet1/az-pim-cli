package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func validPRURL(value string) bool {
	prefix := "https://github.com/" + Fork + "/pull/"
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	n, e := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 64)
	return e == nil && n > 0
}

type SyncOptions struct{ Repository, Upstream, Base, WorkDir string }
type SyncResult struct {
	State    string      `json:"state"`
	Pin      UpstreamPin `json:"pin"`
	Branch   string      `json:"branch"`
	PRURL    string      `json:"pr_url,omitempty"`
	OlderPRs []string    `json:"older_prs,omitempty"`
}
type pullRequest struct {
	State string `json:"state"`
	URL   string `json:"html_url"`
	Head  struct {
		SHA  string `json:"sha"`
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

func cloneFork(ctx context.Context, r Runner, work string) (commands, func(), error) {
	dir, e := os.MkdirTemp(work, "fork-automation-")
	if e != nil {
		return commands{}, nil, e
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	c := commands{ctx, r, dir}
	if _, e = c.git("clone", "--no-tags", "--no-checkout", "https://github.com/"+Fork+".git", "candidate"); e != nil {
		cleanup()
		return c, nil, e
	}
	c.dir = filepath.Join(dir, "candidate")
	for _, pair := range [][2]string{{"user.name", "upstream-sync[bot]"}, {"user.email", "upstream-sync@users.noreply.github.com"}} {
		if _, e = c.git("config", pair[0], pair[1]); e != nil {
			cleanup()
			return c, nil, e
		}
	}
	return c, cleanup, nil
}
func (c commands) fetchRelease(tag string) (string, error) {
	if _, e := ParseForkVersion(tag); e != nil {
		return "", e
	}
	ref := "refs/upstream/releases/" + tag
	if _, e := c.git("fetch", "--no-tags", "https://github.com/"+Upstream+".git", "refs/tags/"+tag+":"+ref); e != nil {
		return "", fmt.Errorf("fetch pinned upstream tag: %w", e)
	}
	sha, e := c.git("rev-parse", "--verify", ref+"^{commit}")
	if e != nil || !shaPattern.MatchString(sha) {
		return "", fmt.Errorf("upstream tag does not resolve to a full commit: %v", e)
	}
	return sha, nil
}
func (c commands) pinAt(ref string) (*UpstreamPin, error) {
	names, e := c.git("ls-tree", "-r", "--name-only", ref, "--", PinPath)
	if e != nil {
		return nil, e
	}
	if names == "" {
		return nil, nil
	}
	b, e := c.git("show", ref+":"+PinPath)
	if e != nil {
		return nil, e
	}
	p, e := DecodePin([]byte(b))
	if e != nil {
		return nil, e
	}
	return &p, nil
}
func (c commands) ancestor(a, b string) (bool, error) {
	_, e := c.git("merge-base", "--is-ancestor", a, b)
	if e == nil {
		return true, nil
	}
	if exitCode(e, 1) {
		return false, nil
	}
	return false, e
}
func (c commands) verifyPin(p UpstreamPin, ref string) error {
	if e := ValidatePin(p); e != nil {
		return e
	}
	sha, e := c.fetchRelease(p.Tag)
	if e != nil {
		return e
	}
	if sha != p.Commit {
		return fmt.Errorf("upstream tag %s was retargeted", p.Tag)
	}
	ok, e := c.ancestor(p.Commit, ref)
	if e != nil {
		return e
	}
	if !ok {
		return fmt.Errorf("pinned upstream commit is not incorporated in %s", ref)
	}
	return nil
}
func checkAdvance(old *UpstreamPin, next UpstreamPin) error {
	if old == nil {
		return nil
	}
	a, _ := ParseForkVersion(old.Tag)
	b, _ := ParseForkVersion(next.Tag)
	if CompareVersions(b, a) < 0 {
		return fmt.Errorf("refusing upstream downgrade from %s to %s", old.Tag, next.Tag)
	}
	if old.Tag == next.Tag && old.Commit != next.Commit {
		return fmt.Errorf("upstream tag commit changed")
	}
	return nil
}
func (c commands) prs(branch, state string) ([]pullRequest, error) {
	path := "repos/" + Fork + "/pulls?state=" + state + "&base=main"
	if branch != "" {
		path += "&head=" + url.QueryEscape("vijayet1:"+branch)
	}
	var all []pullRequest
	for page := 1; page <= 1000; page++ {
		b, e := c.api(path + "&per_page=100&page=" + strconv.Itoa(page))
		if e != nil {
			return nil, e
		}
		var prs []pullRequest
		if e = json.Unmarshal(b, &prs); e != nil {
			return nil, e
		}
		all = append(all, prs...)
		if len(prs) < 100 {
			return all, nil
		}
	}
	return nil, fmt.Errorf("pull request pagination limit exceeded")
}
func (c commands) prepare(base string, p UpstreamPin) (string, error) {
	if _, e := c.git("checkout", "--detach", base); e != nil {
		return "", e
	}
	incorporated, e := c.ancestor(p.Commit, "HEAD")
	if e != nil {
		return "", e
	}
	if !incorporated {
		if _, e = c.git("merge", "--no-ff", "--no-commit", p.Commit); e != nil {
			return "", fmt.Errorf("release merge needs human conflict resolution: %w", e)
		}
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	directory := filepath.Join(c.dir, ".github")
	if info, err := os.Lstat(directory); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf(".github must be a real directory")
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if e = os.MkdirAll(directory, 0755); e != nil {
		return "", e
	}
	if info, err := os.Lstat(filepath.Join(c.dir, PinPath)); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("source pin must be a regular file")
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if e = os.WriteFile(filepath.Join(c.dir, PinPath), append(b, '\n'), 0644); e != nil {
		return "", e
	}
	if _, e = c.git("add", "--", PinPath); e != nil {
		return "", e
	}
	return c.git("write-tree")
}
func RunSync(ctx context.Context, r Runner, o SyncOptions) (SyncResult, error) {
	var result SyncResult
	if o.Repository != Fork || o.Upstream != Upstream || o.Base != "main" {
		return result, fmt.Errorf("sync is restricted to %s main from %s", Fork, Upstream)
	}
	c := commands{ctx, r, o.WorkDir}
	b, e := c.api("repos/" + Upstream + "/releases/latest")
	if e != nil {
		return result, e
	}
	release, e := ParseStableRelease(b)
	if e != nil {
		return result, e
	}
	c, cleanup, e := cloneFork(ctx, r, o.WorkDir)
	if e != nil {
		return result, e
	}
	defer cleanup()
	sha, e := c.fetchRelease(release.Tag)
	if e != nil {
		return result, e
	}
	pin := UpstreamPin{1, Upstream, release.Tag, sha, release.URL}
	branch := "sync/upstream-" + pin.Tag
	result = SyncResult{State: "noop", Pin: pin, Branch: branch}
	baseSHA, e := c.git("rev-parse", "refs/remotes/origin/main")
	if e != nil {
		return result, e
	}
	old, e := c.pinAt(baseSHA)
	if e != nil {
		return result, e
	}
	if old != nil {
		if e = c.verifyPin(*old, baseSHA); e != nil {
			return result, e
		}
	}
	if e = checkAdvance(old, pin); e != nil {
		return result, e
	}
	if old != nil && *old == pin {
		return result, nil
	}
	open, e := c.prs("", "open")
	if e != nil {
		return result, e
	}
	for _, pr := range open {
		if pr.Head.Repo.FullName == Fork && strings.HasPrefix(pr.Head.Ref, "sync/upstream-") && pr.Head.Ref != branch {
			if !validPRURL(pr.URL) {
				return result, fmt.Errorf("invalid older proposal URL")
			}
			result.OlderPRs = append(result.OlderPRs, pr.URL)
		}
	}
	prs, e := c.prs(branch, "all")
	if e != nil {
		return result, e
	}
	if len(prs) > 1 {
		return result, fmt.Errorf("multiple candidate PRs require inspection")
	}
	if len(prs) == 1 && prs[0].State != "open" {
		return result, fmt.Errorf("candidate PR is closed; inspect it before retrying")
	}
	remote, e := c.git("ls-remote", "--heads", "origin", "refs/heads/"+branch)
	if e != nil {
		return result, e
	}
	if remote != "" {
		existing := strings.Fields(remote)[0]
		if !shaPattern.MatchString(existing) {
			return result, fmt.Errorf("invalid candidate ref")
		}
		if len(prs) == 1 && (prs[0].Head.SHA != existing || prs[0].Head.Ref != branch || prs[0].Head.Repo.FullName != Fork || prs[0].Base.Ref != "main") {
			return result, fmt.Errorf("PR identity differs from candidate branch")
		}
		ref := "refs/remotes/origin/" + branch
		if _, e = c.git("fetch", "--no-tags", "origin", "refs/heads/"+branch+":"+ref); e != nil {
			return result, e
		}
		previous, e := c.pinAt(existing)
		if e != nil || previous == nil || *previous != pin {
			return result, fmt.Errorf("unexpected candidate source pin: %v", e)
		}
		parent, e := c.git("rev-parse", existing+"^1")
		if e != nil {
			return result, e
		}
		ok, e := c.ancestor(parent, baseSHA)
		if e != nil || !ok {
			return result, fmt.Errorf("candidate base is not fork main history")
		}
		expected, e := c.prepare(parent, pin)
		if e != nil {
			return result, e
		}
		actual, e := c.git("rev-parse", existing+"^{tree}")
		if e != nil || expected != actual {
			return result, fmt.Errorf("candidate has unexpected content; refusing to rewrite it")
		}
		if len(prs) == 1 {
			if !validPRURL(prs[0].URL) {
				return result, fmt.Errorf("unexpected pull request URL")
			}
			result.State = "reused"
			result.PRURL = prs[0].URL
			return result, nil
		}
	} else {
		if len(prs) != 0 {
			return result, fmt.Errorf("open PR has no candidate branch")
		}
		if _, e = c.prepare(baseSHA, pin); e != nil {
			return result, e
		}
		if _, e = c.git("commit", "-m", "chore(sync): merge upstream "+pin.Tag); e != nil {
			return result, e
		}
		current, e := c.git("ls-remote", "--heads", "origin", "refs/heads/main")
		if e != nil || len(strings.Fields(current)) != 2 || strings.Fields(current)[0] != baseSHA {
			return result, fmt.Errorf("fork main moved during preparation; retry from fresh main")
		}
		if e = c.push("origin", "HEAD:refs/heads/"+branch); e != nil {
			return result, e
		}
	}
	body := fmt.Sprintf("Track stable upstream release %s\n\nUpstream commit: `%s`\n\nPreserves fork history. Review workflow changes and run all checks before merging. No fork release is published by this PR.", pin.ReleaseURL, pin.Commit)
	if len(result.OlderPRs) > 0 {
		body += "\n\nOlder open proposals (review separately; not automatically closed):\n" + strings.Join(result.OlderPRs, "\n")
	}
	opened, e := c.gh("pr", "create", "--repo", Fork, "--base", "main", "--head", branch, "--title", "chore(sync): merge upstream "+pin.Tag, "--body", body)
	if e != nil {
		again, lookupErr := c.prs(branch, "open")
		if lookupErr != nil || len(again) != 1 {
			return result, e
		}
		remote, readErr := c.git("ls-remote", "--heads", "origin", "refs/heads/"+branch)
		if readErr != nil || len(strings.Fields(remote)) != 2 || again[0].Head.SHA != strings.Fields(remote)[0] || again[0].Head.Ref != branch || again[0].Head.Repo.FullName != Fork || again[0].Base.Ref != "main" {
			return result, fmt.Errorf("concurrent PR does not match candidate")
		}
		result.State = "reused"
		result.PRURL = again[0].URL
	} else {
		result.State = "created"
		result.PRURL = strings.TrimSpace(string(opened))
	}
	if !validPRURL(result.PRURL) {
		return result, fmt.Errorf("unexpected pull request URL")
	}
	return result, nil
}
