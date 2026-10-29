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
8. If Homebrew distribution is enabled, update only the tap formula metadata using the verified published manifest.

Expected downloadable outputs: six platform binaries and their archives, a universal macOS binary and archive, source archive, checksum file, and the verification manifest. The release notes identify the upstream tag/SHA and independent fork tag/SHA.

The tag-triggered publisher and release-please caller have been removed to keep a single release owner. Reintroducing those paths requires a separate design review.

## Retry and recovery

- Missing/expired App configuration: correct the environment/App setup, then rerun the producer. Do not broaden an unrelated PAT.
- Upstream API failure or missing stable release: stop; errors are not treated as an absent fork release or a guessed version.
- Deleted/retargeted upstream tag or invalid source pin: inspect the upstream change; no automatic rewrite is permitted.
- Main moved during proposal preparation: rerun from fresh main. No force-push recovery is used.
- Tag/build/upload failure: rerun the manual workflow with the identical tag and SHA. A tag at another SHA is rejected. An existing draft can be replaced only after source/version checks; a newer already-published version can make an older draft ineligible to publish latest.
- Identical complete published release: verify and report a release no-op without editing assets or the latest pointer. If Homebrew is enabled, the protected job still runs to repair or confirm the tap formula.
- Published release missing its manifest, targets, or valid checksums: stop for operator investigation. Historical releases are not automatically repaired. Prefer a new reviewed fork version rather than moving an existing tag.
- Environment API inaccessible or protection missing: publishing stops before writes. Confirm plan support, reviewer rules, admin-bypass setting, main-only branch policy, and token actions: read permission.

## Homebrew distribution

The optional post-publication step updates `Formula/az-pim-cli.rb` on `vijayet1/homebrew-tap` main. It changes only the stable version, two macOS archive URLs, and their SHA-256 values, preserving the existing install/test code. It refuses draft releases, incomplete published manifests, unknown formula layouts, downgrades, and changed checksums at the same version. Concurrent content changes use a bounded reread/retry with the GitHub contents API SHA; no force push or branch-rule bypass is used.

### One-time credential setup

1. Create an expiring fine-grained PAT restricted to **only `vijayet1/homebrew-tap`**, with repository **Contents: read and write**. No Workflows, Administration, or Pull requests permission is needed for the single-file contents update.
2. Add it as environment secret **`HOMEBREW_TAP_TOKEN`** in this repository's protected **`fork-release`** environment, not as a repository-wide secret. For an interactive terminal upload that does not put the value in command arguments/history:

   ```bash
   gh secret set HOMEBREW_TAP_TOKEN --repo vijayet1/az-pim-cli --env fork-release
   ```

3. After the workflow change is merged and the secret is configured, enable the **repository** variable:

   ```bash
   gh variable set HOMEBREW_TAP_ENABLED --repo vijayet1/az-pim-cli --body true
   ```

The existing repository-wide `TAP_REPO_PAT` is not used or copied automatically: its scope has not been verified. This integration does not broaden the sync App installation or the source repository's GITHUB_TOKEN. The dedicated tap credential is passed only to fixed-path tap API calls, never in arguments or logs.

The protected publisher checks that the optional secret is present before release writes. A permission failure during the later tap update can still leave a successfully published release with an outdated formula; do not delete or republish that release to repair Homebrew. Fix the credential/configuration and dispatch the same tag and original source SHA again. With `HOMEBREW_TAP_ENABLED=true`, the already-published path skips all tag/build/release writes but still requires environment approval before checking/updating the tap. An already-current formula is a no-op. A formula newer than the requested release is refused rather than downgraded. If tap branch protection later requires PRs, adjust the integration through review; do not grant bypass solely to make the direct update succeed.

The updater does not execute formula code. Review and test formula behavior in the tap. A read-only `homebrew-plan` command can preview the verified metadata when the trusted tool is run from a checkout of the exact selected release commit; `homebrew-update` is restricted to the protected release context.

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

For macOS, install the fork through its tap:

```bash
brew install vijayet1/tap/az-pim-cli
brew update
brew upgrade vijayet1/tap/az-pim-cli
```

If Homebrew 7 reports an untrusted formula, review `vijayet1/homebrew-tap/Formula/az-pim-cli.rb` before explicitly trusting that formula with `brew trust --formula vijayet1/tap/az-pim-cli`. The updater does not alter local Homebrew trust settings.

Homebrew uses the version committed to the tap; publishing assets alone does not update an installation. The optional distribution setup above keeps that formula synchronized after verified publication.

Other platforms can download the selected stable assets from `https://github.com/vijayet1/az-pim-cli/releases` and verify checksums before use. The Go module remains `github.com/netr0m/az-pim-cli`: upstream `go install ...@latest` does not install this fork's release.

## Rollback

Disable the enable variables/workflows and revoke the sync App installation/key. Revert automation through a reviewed PR if needed. Keep published history intact; do not restore the old direct-main/force-tag workflow as rollback.
