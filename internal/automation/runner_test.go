package automation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecRunnerRejectsUnsupportedPrograms(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "executed")
	program := filepath.Join(dir, "unexpected-program")
	if err := os.WriteFile(program, []byte("#!/bin/sh\nprintf executed > '"+marker+"'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "sh", "bash", "/bin/sh", "/usr/bin/git", "git;echo bad", program} {
		t.Run(name, func(t *testing.T) {
			_, err := (ExecRunner{}).Run(context.Background(), dir, nil, name)
			if err == nil || !strings.Contains(err.Error(), "unsupported executable") {
				t.Fatalf("expected rejection before execution, got %v", err)
			}
		})
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("unsupported program executed: %v", err)
	}
}

func TestExecRunnerRunsAllowedPrograms(t *testing.T) {
	for _, tc := range []struct{ name, prefix string }{{"git", "git version "}, {"gh", "gh version "}} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := (ExecRunner{}).Run(context.Background(), t.TempDir(), nil, tc.name, "--version")
			if err != nil || !strings.HasPrefix(string(out), tc.prefix) {
				t.Fatalf("wrong executable dispatched: output=%q error=%v", out, err)
			}
		})
	}
}
