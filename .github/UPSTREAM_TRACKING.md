# Upstream tracking and fork releases

## Policy

This fork tracks published stable releases from `netr0m/az-pim-cli`, not upstream main. Automation proposes a PR; a maintainer reviews and merges it. Publishing is a separate manual, approval-gated operation.

Fork releases use independent stable `vMAJOR.MINOR.PATCH` versions. A fork version does not identify its upstream base. `.github/upstream-release.json`, the PR, release notes, and `fork-release-manifest.json` record the upstream release and full commit SHA separately.

For example, a fork v1.17.0 may incorporate upstream v1.18.0. This is illustrative: query the actual fork tags/releases before selecting the next version. New versions must exceed all existing stable fork tags and published stable releases. Fork-only fixes can advance the fork version without changing the source pin.

Published tags and releases are immutable. Never force-move a tag to repair a failed build. Upstream tags are fetched into a separate Git namespace and must continue to resolve to their recorded commit.

Changing this tracking policy does not undo upstream commits already merged into fork main. Review the first sync diff; no history reset is part of migration.

## Workflows

| Workflow file | Responsibility |
| --- | --- |
| `sync-upstream.yaml` | Daily/manual stable-release proposal; never pushes main or publishes |
| `automation-checks.yaml` | Automation tests, source-pin validation, and workflow linting |
| `on-push-main.yaml` | Existing application tests/scans, without release-please publishing |
| `on-tag.yaml` | Manual fork release; no tag-push trigger |

`.goreleaser.fork.yaml` is the authoritative fork publishing configuration. It is complete and works with GoReleaser OSS. `.goreleaser.yaml` remains the upstream configuration and must not be used to publish fork releases.

## Operator setup

These actions require repository administration and explicit approval. The implementation does not provision them.

1. During cutover, disable the old scheduled workflow before merging the implementation. Do not run the old direct-main/force-tag path alongside the new path.
2. Create a dedicated GitHub App and install it only on `vijayet1/az-pim-cli`. Grant repository Contents: write, Workflows: write, and Pull requests: write. No organization-wide installation or broad PAT is needed.
3. Create environment `upstream-sync`, restricted to branch `main`. Set environment variable `UPSTREAM_SYNC_APP_ID` and environment secret `UPSTREAM_SYNC_APP_PRIVATE_KEY`. Do not store the private key as a repository-wide secret.
4. Create environment `fork-release` with at least one required reviewer, admin bypass disabled, and a custom deployment branch policy containing only branch `main` (not a tag rule). The publisher checks these rules through read-only environment APIs and refuses writes if they are missing, unavailable, or inconsistent. Its token needs contents: write and actions: read; no App key is used for publishing.
5. Review branch protection/rulesets for main. Require PR review and the application/automation checks, including validation against an up-to-date base. Configure stale-review dismissal and strict up-to-date checks as appropriate; do not assume administrators cannot bypass rules by default.
6. Audit existing repository-wide secrets and all workflows that consume them. Remove the obsolete release-please token only after confirming it has no other consumer. Do not reuse TAP_REPO_PAT or another existing token without checking its purpose and scope.
7. Test the complete flow in a disposable repository with separate App installation and environment settings. Adjust the explicit repository identities/config in that test copy and verify every write targets the disposable repository. Do not create test releases in the production fork.
8. Once setup and rehearsal pass, set repository variable `UPSTREAM_SYNC_ENABLED=true`. Set repository variable `FORK_RELEASE_ENABLED=true` only when publication is ready. These are repository variables because job-level enable checks run before environment variables are available.

Checkout does not persist Git credentials. The producer is built from trusted main before it handles candidate content. Its short-lived App token is revoked by the token action's post-job cleanup. The producer never executes candidate tests or scripts.

This is not a sandbox for malicious same-repository workflow changes. PR CI still includes the existing scanner integrations. Review upstream workflow changes and repository-secret exposure explicitly. Keep the App key environment-scoped and the fork-release environment protected.

## Propose and review a sync

After operator setup, a maintainer may dispatch the producer:

```bash
gh workflow run sync-upstream.yaml -R vijayet1/az-pim-cli --ref main
gh run list -R vijayet1/az-pim-cli --workflow sync-upstream.yaml --limit 3
gh pr list -R vijayet1/az-pim-cli --base main --state open
```

Expected behavior:

- GitHub's designated latest published stable upstream release is resolved to its exact commit, peeling annotated tags.
- A branch named `sync/upstream-vMAJOR.MINOR.PATCH` merges that commit and records the source pin.
- A same-release open PR is reused without rewriting its history. Older open proposals are reported, not automatically closed.
- If main already records and incorporates the same source, the producer reports a no-op.
- If legacy main contains the release without metadata, a metadata-only PR is proposed.
- Main and all published tags remain unchanged until a maintainer integrates the PR.

Inspect the source pin, full diff, fork-specific behavior, and especially `.github/workflows/*` changes. App-authenticated PR creation allows PR CI to run; confirm its actual status rather than assuming it ran:

```bash
gh pr checks <pr-number> -R vijayet1/az-pim-cli
gh pr diff <pr-number> -R vijayet1/az-pim-cli
```

Merge upstream integration PRs using **Create a merge commit**, not squash or rebase. The publisher and sync validator require the pinned upstream commit to remain an ancestor of fork main.

Do not merge an older proposal after adopting a newer upstream release. Candidate validation rejects recorded-source downgrades; require current checks against the latest base.

A conflict stops before branch push. Resolve it in a separate reviewed integration PR, preserving fork changes and recording the exact upstream pin. Automation also stops on unexpectedly modified candidate branches rather than overwriting human work. Merge or inspect such work manually.

To decline a release, close its PR. The producer will not automatically reopen a closed-unmerged candidate for that same release. It reports the condition for operator inspection; a newer upstream release may receive a separate proposal. Disable the producer's enable variable if repeated failure notifications are unwanted while that release remains latest.

## Publish an approved fork release

First inspect current fork state (read-only):

```bash
gh api repos/vijayet1/az-pim-cli/tags --paginate
gh release list -R vijayet1/az-pim-cli --limit 20
gh api repos/vijayet1/az-pim-cli/commits/main --jq .sha
```

Select an unused stable fork version and the full SHA of the reviewed main commit. The selected commit must contain a valid upstream source pin and incorporate that upstream commit. A release can use an older reviewed main commit, but a newly published version must still increase relative to existing stable fork versions.

The following dispatch is a publishing operation, not a read-only check. Replace both placeholders before running:

```bash
gh workflow run on-tag.yaml -R vijayet1/az-pim-cli --ref main \
  -f tag='<approved-stable-fork-tag>' \
  -f commit='<full-reviewed-main-commit-sha>'
```

The preflight job builds its validator from trusted current main, validates the selected SHA before preparing its worktree, checks release state, runs tests, and performs a non-publishing build. The write-enabled job waits for fork-release environment approval, then repeats the source/release checks.

Publishing stages:

1. Validate fork/upstream identities, source ancestry, stable version ordering, and existing tag/release state.
2. Verify environment protection before any tag/asset write.
3. Create the tag only if absent, never with force.
4. Build/upload a draft with pinned GoReleaser OSS.
5. Verify the complete output set, downloaded asset hashes, and checksum contents.
6. Upload a source-and-assets manifest for future verification.
7. Recheck version ordering, publish as stable, and verify the latest pointer and published assets.

Expected downloadable outputs: six platform binaries and their archives, a universal macOS binary and archive, source archive, checksum file, and the verification manifest. The release notes identify the upstream tag/SHA and independent fork tag/SHA.

The tag-triggered publisher and release-please caller have been removed to keep a single release owner. Reintroducing those paths requires a separate design review.

## Retry and recovery

- Missing/expired App configuration: correct the environment/App setup, then rerun the producer. Do not broaden an unrelated PAT.
- Upstream API failure or missing stable release: stop; errors are not treated as an absent fork release or a guessed version.
- Deleted/retargeted upstream tag or invalid source pin: inspect the upstream change; no automatic rewrite is permitted.
- Main moved during proposal preparation: rerun from fresh main. No force-push recovery is used.
- Tag/build/upload failure: rerun the manual workflow with the identical tag and SHA. A tag at another SHA is rejected. An existing draft can be replaced only after source/version checks; a newer already-published version can make an older draft ineligible to publish latest.
- Identical complete published release: verify and report a no-op without editing assets or the latest pointer.
- Published release missing its manifest, targets, or valid checksums: stop for operator investigation. Historical releases are not automatically repaired. Prefer a new reviewed fork version rather than moving an existing tag.
- Environment API inaccessible or protection missing: publishing stops before writes. Confirm plan support, reviewer rules, admin-bypass setting, main-only branch policy, and token actions: read permission.

## Local verification

Use Go 1.27.3 from go.mod, Git, gh, golangci-lint 2.15.0 (the pre-commit pin), actionlint 1.7.12, and GoReleaser OSS 2.19.0. New/changed Actions are pinned to resolved commit SHAs with version comments.

The linter must support the Go toolchain being analyzed. Use the pinned pre-commit Go hook or a compatible build of golangci-lint 2.15.0; do not reuse an older linter binary built for Go 1.26 against this Go 1.27 module.

```bash
go test ./internal/automation ./tools/fork-automation -count=1
go test ./...
go vet ./internal/automation ./tools/fork-automation
golangci-lint run
pre-commit run --all-files
actionlint .github/workflows/sync-upstream.yaml .github/workflows/automation-checks.yaml .github/workflows/on-push-main.yaml .github/workflows/on-tag.yaml
goreleaser check --config .goreleaser.fork.yaml
git diff --check
```

A local non-publishing build may create ignored `dist/` output:

```bash
goreleaser release --config .goreleaser.fork.yaml --snapshot --skip=publish --clean
```

For a stable-tag rehearsal, use a separate temporary clone, create a local fixture tag there, and run GoReleaser with `--skip=publish`. Never create the fixture tag in the shared working repository. Local builds cannot prove GitHub App permissions, required checks, environment approvals, or draft publication/retry behavior; those need the disposable-repository acceptance run.

## Installation

Download the selected stable release assets from `https://github.com/vijayet1/az-pim-cli/releases`. Verify checksums before use. The Go module remains `github.com/netr0m/az-pim-cli`: upstream `go install ...@latest` does not install this fork's release. Automatic Homebrew/package-manager updates are not configured by these changes.

## Rollback

Disable the enable variables/workflows and revoke the sync App installation/key. Revert automation through a reviewed PR if needed. Keep published history intact; do not restore the old direct-main/force-tag workflow as rollback.
