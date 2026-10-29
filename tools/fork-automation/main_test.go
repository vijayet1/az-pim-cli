package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("fixture write failed") }
func TestCLIPropagatesOutputFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	for _, args := range [][]string{{"init", "-b", "main"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "--allow-empty", "-m", "base"}, {"update-ref", "refs/remotes/origin/main", "HEAD"}} {
		if b, e := exec.Command("git", args...).CombinedOutput(); e != nil {
			t.Fatalf("%s %v", b, e)
		}
	}
	var errOut bytes.Buffer
	if code := run(context.Background(), []string{"validate-pin"}, failedWriter{}, &errOut); code == 0 {
		t.Fatal("ignored output failure")
	}
}

func TestReleaseCLIRequiresMain(t *testing.T) {
	t.Setenv("GITHUB_REF", "refs/heads/feature")
	t.Setenv("GITHUB_REPOSITORY", "vijayet1/az-pim-cli")
	for _, command := range []string{"release-check", "release-tag", "release-verify", "homebrew-update"} {
		var out, errOut bytes.Buffer
		code := run(context.Background(), []string{command, "--tag", "v1.17.0", "--commit", strings.Repeat("a", 40)}, &out, &errOut)
		if code == 0 || !strings.Contains(errOut.String(), "release must run on fork main") {
			t.Fatalf("%s: %d %s", command, code, &errOut)
		}
	}
}

func TestHomebrewCLIRequiresDedicatedTokenBeforeAPI(t *testing.T) {
	t.Setenv("GITHUB_REF", "refs/heads/main")
	t.Setenv("GITHUB_REPOSITORY", "vijayet1/az-pim-cli")
	t.Setenv("GH_TOKEN", "fixture-source-token")
	t.Setenv("HOMEBREW_TAP_TOKEN", "")
	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"homebrew-update", "--tag", "v1.17.0", "--commit", strings.Repeat("a", 40)}, &out, &errOut)
	if code == 0 || !strings.Contains(errOut.String(), "HOMEBREW_TAP_TOKEN") {
		t.Fatalf("missing token was not rejected early: %s", &errOut)
	}
}

func TestCLIRejectsUnsafeInputBeforeCommands(t *testing.T) {
	for _, args := range [][]string{{}, {"unknown"}, {"sync", "--repository", "attacker/repo"}, {"sync", "--base", "feature"}, {"validate-pin", "--base-ref", "HEAD"}, {"release-check", "--tag", "v1.0.0;echo bad", "--commit", "abc"}, {"sync", "extra"}} {
		var out, errOut bytes.Buffer
		if code := run(context.Background(), args, &out, &errOut); code == 0 {
			t.Fatalf("accepted %v", args)
		}
		if out.Len() != 0 || strings.Contains(errOut.String(), "fixture-secret") {
			t.Fatalf("unexpected output: %s %s", &out, &errOut)
		}
	}
}
