package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestReleaseCLIRequiresMain(t *testing.T) {
	t.Setenv("GITHUB_REF", "refs/heads/feature")
	t.Setenv("GITHUB_REPOSITORY", "vijayet1/az-pim-cli")
	for _, command := range []string{"release-check", "release-tag", "release-verify"} {
		var out, errOut bytes.Buffer
		code := run(context.Background(), []string{command, "--tag", "v1.17.0", "--commit", strings.Repeat("a", 40)}, &out, &errOut)
		if code == 0 || !strings.Contains(errOut.String(), "release must run on fork main") {
			t.Fatalf("%s: %d %s", command, code, &errOut)
		}
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
