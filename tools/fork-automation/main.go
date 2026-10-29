package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/netr0m/az-pim-cli/internal/automation"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
func run(ctx context.Context, args []string, out, errOut io.Writer) int {
	fail := func(e error) int { fmt.Fprintln(errOut, e); return 1 }
	if len(args) == 0 {
		return fail(fmt.Errorf("expected sync or validate-pin command"))
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(errOut)
	repo := fs.String("repository", automation.Fork, "fork repository")
	up := fs.String("upstream", automation.Upstream, "upstream repository")
	base := fs.String("base", "main", "base branch")
	baseRef := fs.String("base-ref", "refs/remotes/origin/main", "freshly fetched validation base")
	tag := fs.String("tag", "", "independent stable fork tag")
	commit := fs.String("commit", "", "exact reviewed fork commit")
	artifacts := fs.String("artifacts", "dist/artifacts.json", "GoReleaser artifact metadata")
	if e := fs.Parse(args[1:]); e != nil {
		return 1
	}
	if fs.NArg() != 0 {
		return fail(fmt.Errorf("unexpected positional arguments"))
	}
	if *repo != automation.Fork || *up != automation.Upstream || *base != "main" {
		return fail(fmt.Errorf("repository/upstream/base do not match approved configuration"))
	}
	cwd, e := os.Getwd()
	if e != nil {
		return fail(e)
	}
	switch args[0] {
	case "sync":
		if os.Getenv("GITHUB_REF") != "refs/heads/main" || os.Getenv("GITHUB_REPOSITORY") != automation.Fork {
			return fail(fmt.Errorf("sync must run on fork main"))
		}
		if os.Getenv("GH_TOKEN") == "" {
			return fail(fmt.Errorf("GH_TOKEN is required for sync"))
		}
		result, e := automation.RunSync(ctx, automation.ExecRunner{}, automation.SyncOptions{Repository: *repo, Upstream: *up, Base: *base, WorkDir: os.TempDir()})
		if e != nil {
			return fail(e)
		}
		if e = json.NewEncoder(out).Encode(result); e != nil {
			return fail(e)
		}
	case "release-check", "release-tag", "release-verify":
		if os.Getenv("GITHUB_REF") != "refs/heads/main" || os.Getenv("GITHUB_REPOSITORY") != automation.Fork {
			return fail(fmt.Errorf("release must run on fork main"))
		}
		if os.Getenv("GH_TOKEN") == "" {
			return fail(fmt.Errorf("GH_TOKEN is required for release operations"))
		}
		o := automation.ReleaseOptions{Repository: *repo, Tag: *tag, Commit: *commit, Ref: os.Getenv("GITHUB_REF")}
		if args[0] == "release-check" {
			result, err := automation.CheckRelease(ctx, automation.ExecRunner{}, cwd, o)
			if err != nil {
				return fail(err)
			}
			if err = json.NewEncoder(out).Encode(result); err != nil {
				return fail(err)
			}
		} else {
			if err := automation.CheckReleaseEnvironment(ctx, automation.ExecRunner{}, cwd); err != nil {
				return fail(err)
			}
			var err error
			if args[0] == "release-tag" {
				err = automation.EnsureReleaseTag(ctx, automation.ExecRunner{}, cwd, o)
			} else {
				err = automation.VerifyRelease(ctx, automation.ExecRunner{}, cwd, o, *artifacts)
			}
			if err != nil {
				return fail(err)
			}
			fmt.Fprintln(out, `{"state":"verified"}`)
		}
	case "validate-pin":
		if e = automation.ValidateCandidate(ctx, automation.ExecRunner{}, cwd, *baseRef); e != nil {
			return fail(e)
		}
		fmt.Fprintln(out, `{"state":"validated"}`)
	default:
		return fail(fmt.Errorf("unknown command %q", args[0]))
	}
	return 0
}
