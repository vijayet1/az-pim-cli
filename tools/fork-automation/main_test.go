package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

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
