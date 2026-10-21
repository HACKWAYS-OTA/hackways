# History-preserving import and candidate review

## Scope and branch policy

- Destination: private `HACKWAYS-OTA/hackways`.
- Original `varshith-dev/HACKWAYS-IN` remains unchanged; do not delete, archive, force-push or change its settings.
- Import the existing source `main` and `develop` refs and tags with their original object IDs. No squash, rebase, rewritten authors or replacement initial commit.
- Source refs inspected: main `f7cff5510cc8c3ab5e3471b0a319aa133565b494`; develop `95388304079f32a3c5d205e6ac7699e7747a9e42`.
- Importing the pre-existing main ref is distinct from authoring a new main commit. All recent product changes and launch fixes go on `release/2026-10-21-launch`, with a PR to `develop`. A separate reviewed PR promotes integration to `main`. No direct main development commits or automatic PR merging.
- Preserve four Next.js apps, two Go backends, shared packages, SQLc contracts, numbered migrations and deployment files in one repository.
- Actual credentials, environment files other than sanitized examples, keys, backups, binaries, caches and generated build output are excluded. Generated SQLc source remains versioned as required by the architecture.

## Import procedure

1. Fetch source refs read-only; inspect history and the full tracked/untracked proposal. Scan reachable blobs and proposal files for private keys, token patterns and actual approved protected credential values without emitting matches.
2. Verify destination refs remain empty. Temporarily suspend destination Actions only for importing historical refs, so obsolete workflows cannot build historical images. Restore Actions immediately afterward. Do not change source Actions, billing, packages, webhooks or repository visibility.
3. Push the unchanged historical main/develop refs and source tags without force. Verify destination object IDs against source. Use a separate organization remote; retain the existing origin.
4. Commit reviewed local source on the release branch and open a PR into destination develop. Preserve the original source ancestry. No auto-deployment webhook is configured.
5. Run affected PR checks. Select an immutable candidate via an explicit `rc-*` tag only after reviewing the source/security checks and current quota. The tag triggers full release integration tests and six CI image builds. It does not deploy.
6. Record the final candidate SHA, PR/run links, complete results and six image digests in the external handoff. Do not call the imported historical SHA or an uncommitted tree the candidate.

## DEVARA findings

### RSVP ownership

The email conflict lookup intentionally sees both guest and account-linked rows. Added an explicit account-owner check before either status evaluation or reset. An unauthenticated same-email request cannot revive cancelled/expired account registrations. Regression tests in the existing mandatory PostgreSQL suite compare the complete registration, ticket, answer snapshot, event/inventory, submission count, outbox count and audit count before/after rejection. The authenticated owner can still re-register. SQL query authority, event locking and architecture are unchanged.

### Deployment ordering

Automatic deployment stays disabled. The optional legacy daemon now requires an operator-owned, increasing approved-release selection and durable attempt state. It rejects unapproved/stale/replayed events, serializes/coalesces valid deliveries while busy, rechecks approval before activation and verifies production-branch ancestry. Five regression tests cover out-of-order/replayed deliveries across restart, 24 concurrent busy deliveries, superseded queued approvals, disabled/missing approval and failed-attempt retry policy. No daemon is installed or enabled by this source change.

## CI and cost controls

- PRs/branch pushes run affected-component checks, never image publication. Push runs with open PR coverage do not duplicate the PR's component checks. New source is not pushed directly to main.
- Explicit candidate tags run full integration coverage and six builds. GitHub runners only; Linux/amd64, SHA tags only, maximum two simultaneous image builds, bounded job timeouts, npm/Go/build-layer caches and pinned external Actions.
- Artifacts contain small test/digest records, not images/backups. Retention is one day. Repository log retention must be confirmed with DEVARA in repository Actions settings; no unsupported REST setting is assumed.
- No packages/current/rollback images are removed. Package retention and approved rollback selection remain an operator decision.
- Pre-build GitHub billing API: Free organization, 20 Linux minutes reported for October 2026, fully discounted, zero net amount. The new repository reports zero cache bytes and zero artifacts. Against the published 2,000-minute allowance, approximately 1,980 minutes remain before delayed/unreported usage. Organization-wide storage/budget configuration is not inferred from repository storage alone.
- Estimated cold candidate pipeline: 40-80 Linux runner minutes; actual usage will be measured from job timestamps. This is an estimate, not a capacity or cost guarantee. No paid overage, plan upgrade, larger runner or paid service is enabled.
- Public build origins are fixed explicitly in release workflow: `https://hackways.in`, `https://app.hackways.in`, `https://organizer.hackways.in`. Destination variables were also confirmed set by DEVARA.
- Required reviews and protected branches are not technically enforced by the current private Free plan. Use PR review and operator release approval without claiming enforced GitHub gates.

## Production boundary

This task does not deploy, clean up or modify production configuration, application data, credentials, SSH, networking or running containers. Prior backup/restore and migration-preflight evidence remains documented separately. Working diagnostic DB authentication does not mean the old application runtime credential has been changed. Production migration/rollback, complete ingress/readiness/journey checks and the approved manual release remain separate gates.
