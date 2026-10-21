# Proposed release review — 20 October 2026

**Status: source proposal for review; release BLOCKED. No deployment, cleanup, password change, live configuration update, container removal or restart was performed.**

## Protected-source authentication result

At **14:21:56 UTC**, a final TLS-required, read-only connection attempt to `hackways_prod` used `/opt/hackways/.env` directly, without printing credentials or putting a password-bearing URI in process arguments. PostgreSQL **rejected the password**, as it had at 13:59:30 UTC. The approved alternative local credential file was still absent. This result establishes that the currently available protected source fails; it does not establish that the separately reported updated password was actually loaded into that source or is invalid.

Prior diagnostics established working DNS/TCP, a correctly parsed host/database/user, and equal stored credentials in the host environment and both running containers. Their latest inspected readiness responses were HTTP 503/database unhealthy. None of those settings was changed to try to repair authentication. A working protected current credential is still needed for DB backup, schema inspection and rehearsal.

## Source identity and review scope

- Base commit: `95388304079f32a3c5d205e6ac7699e7747a9e42`, branch `develop`.
- The proposal includes tracked modifications **and individually enumerated untracked files**, not only the output of a normal tracked-file diff. Its companion manifest identifies the complete proposed path set and tree object.
- No new Git commit, branch move, push or approval was created. The tree checksum in the manifest is **not a commit SHA** and must not be reported as a deployed/reviewed release commit.
- Review covered source/dependency/deployment changes, core authorization and transaction paths, owner status changes, upload ownership/completion, session/CORS handling, outbox state handling, migrations, shared controls, event editing and public pages. Compiler/linter/test results supplement this review; they do not establish all browser journeys, production security or concurrency behavior.
- Scope is broad because recent local product work was never committed: all four apps, two Go engines, shared API/UI, migrations 000002–000007, CI/container/ingress/deployer configuration and documentation. The file manifest and full binary-capable patch are authoritative for the exact proposal.

## Changes in the proposal

| Area | Proposed changes |
| --- | --- |
| Core attendee app | Discovery, event pages/theme integration, account/session flows, real registration forms, My events/tickets, Spaces and event creation |
| Organizer app | Owner-scoped dashboard, event editor/save/error dialogs, guest list/detail/status actions, registration questions and check-in |
| Public site | Landing/information/support/policy pages, regular reading typography, clean help topic routes, confirmed support address, no documentation sidebars |
| Admin app | Independent backend readiness display; no claim that full moderation APIs exist |
| Shared packages | Typed API client; root workspace controls for checkbox/radio/select/toggle, motion/provider defaults and no-fragment section jumps |
| Go engines | Ownership/authentication, database-backed sessions, capacity-locked registration/waitlist/approval/cancellation/transfer/tickets, revision-safe event/guest operations, real outbox/ACS and Azure upload verification |
| Schema | Additive workflow foundation, Spaces, public event information, display/theme, visibility and registration-question/answer history |
| Delivery | Corrected Compose nesting, production APP_ENV/API wiring, CI origin build arguments, locked workspace frontend Docker installs, explicit workspace standalone tracing, signed CI deployer and fail-fast deployment |
| Documentation | Current behavior/limitations, GitHub manager guide, sanitized inventory, backup/preflight evidence and corrected stale setup statements |

### Issues found and corrected during this review

1. **Shell execution of dotenv content.** The VM's ACS connection string is unquoted and contains shell separators. Sourcing its Docker dotenv file can split/truncate that value. The proposed deploy script now lets Compose parse the environment file instead of executing it. No live environment file was edited.
2. **Deployment checkout path mismatch.** The proposed script derives its root from its own location and supports `HACKWAYS_ENV_FILE` for a protected external environment file. A clean sibling checkout can be staged without assuming the non-Git live tree is a checkout.
3. **Unpinned release selection and unintended cleanup.** The proposed script requires a full reviewed SHA through `DEPLOY_IMAGE_TAG`, uses it for image tags, and no longer requests removal of orphan containers. No image prune is present.
4. **Docker/CI dependency drift.** All four frontend Dockerfiles now copy the root workspace manifests/lockfile and use `npm ci`, then build the API client and their own app. They retain independent images. No framework/dependency version was upgraded. Legacy app locks are not deleted, but no longer control these Docker installs.
5. **Hoisted dependency packaging.** Each Next.js config explicitly traces from the workspace root, matching the installed Next.js-14 configuration schema and the hoisted/shared workspace layout.
6. **Frontend readiness coverage.** The deployment script now explicitly probes all four frontend roots after Docker health waiting and checks Nginx before reload. These checks still do not exercise signed-in end-to-end journeys.
7. **Migration credential exposure.** A narrow Python-stdlib operations helper consumes Compose JSON privately, verifies equal backend DB URLs and passes the password to the existing migration CLI through its environment, not its URI arguments. Raw CLI output is withheld to prevent credentials/data leaking into deployment logs. Failure remains nonzero. The helper is not an application backend or DB abstraction; Go/pgx/sqlc behavior is unchanged.
8. **Verification/documentation drift.** Corrected stale claims about missing workflows/ACS, migration coverage and CI build arguments, and fixed the Makefile's invalid gofmt path. Existing backend/operations CI now includes operations syntax/compile and deployer checks.

## Validation results

| Verification | Actual outcome |
| --- | --- |
| Root `npm run type-check` | PASS: four apps, API client and UI |
| Root `npm run lint` | PASS: four apps and UI |
| API client build | PASS |
| Final production application builds | PASS: admin 34s, core 53s, organizer 49s, public 41s; final workspace-tracing configs included |
| Core/global `go build ./...`, `go vet ./...` | PASS |
| Core/global `gofmt -l .` | No unformatted files reported |
| Deployer `go test ./...`, `go vet ./...` | PASS compilation/vet; module has no test files |
| Core/global `sqlc compile` | PASS |
| Standard Go test run, no application DB supplied | 23 core + 28 global tests passed; zero failed; 17 core + 5 global DB-backed tests skipped |
| Both Go `-race` commands | NOT RUN successfully: each failed before tests because CGO is disabled |
| New isolated local `_test` DB bootstrap | BLOCKED: local role lacks CREATE DATABASE. No existing application/test schema was used as fallback; no DB-backed fixture cleanup was run |
| Full corrected Compose model on VM | PASS, Docker Compose 5.6.0, protected host env, no printed resolved secrets and no containers launched |
| Proposed deploy/backup/restore scripts | `bash -n` PASS |
| Proposed migration helper | Python compile PASS; validate-only parsing of actual protected Compose model PASS; no connection/migration attempted by this check |
| Installed Next config schema | PASS for all four configs |
| CI YAML / CSS | Parse PASS; operations check present; 23 CSS files parsed |
| Docker COPY input/installer review | All referenced source inputs exist; four target Dockerfiles use root-lock `npm ci` |
| Root lockfile versus workspace manifests | No dependency/devDependency/peerDependency mismatches found |
| Diff whitespace | Tracked check passed; seven additional untracked-file trailing-whitespace/EOF findings were normalized without changing license wording. Final complete patch check is recorded in its manifest |
| Targeted sensitive-file/token/key patterns | No candidate findings; this is not a historical secret audit or guarantee that every possible credential format was detected |

Application builds ran on the local Windows machine from an isolated source copy, using the installed locked dependencies through a directory junction, Node 24.13.0 and a 1024 MB Node heap limit. API origins were public nonsecret build settings and loopback server API origins; no production credential files were copied into the build tree. Active `.next` directories were not used for builds. CI/Docker instead use Node 22/Linux; local success is not container-image or standalone runtime proof. Go here is 1.26.4; CI/modules target 1.23. The temporary source/build outputs remain under ignored `logs/release-review`; no review output was pruned.

## Findings that keep the release blocked

1. **Current DB authentication fails.** Obtain the current credential through an approved protected source and repeat the read-only check. No server password reset is authorized by this review.
2. **Current-source Linux CI and images are absent.** The historical CI run does not validate this broad proposal or the changed Docker/layout/deployment tooling. Run PostgreSQL-16 integration/race checks and build all six images for the accepted commit, then verify digests and host pull access.
3. **DB recovery and migration/rollback rehearsal are absent.** The private VM configuration backup is verified separately, but there is no managed DB dump/isolated restore or migration/compatible-image rollback proof. Existing backup/restore scripts still target `docker exec hackways-postgres` and must not be used against the managed target as though they provide recovery coverage.
4. **Ticket signing configuration is missing on the VM.** The protected source has no TICKET_SIGNING_SEED. A designated operator must arrange stable approved signing material before advertising signed-ticket readiness. No key was generated or existing credential rotated.
5. **Source/release governance remains unsettled.** Product repo migration, private protection support, accepted commit, named release/infrastructure owners and the actual installation/approval gate for the deployer remain pending. Devara is identified only as GitHub manager. A signed successful CI webhook is not, by itself, a human release approval gate.
6. **Ingress and application journeys need release-environment checks.** Current Nginx syntax passes, but updated route/TLS/admin restrictions and all six services must be verified together. Browser keyboard/mobile, signed-in flows, actual ACS/Blob completion, workload memory and complete standalone-image contents remain unverified.

## Demo data and cleanup

The up-migration scan found category reference rows in 000001 and ticket backfill from existing confirmed registrations in 000002; neither is fabricated attendee/event seed data. No proposed up migration inserts demo users, events or RSVPs. Existing integration fixtures remain test-only, guarded by an explicit isolated test DB/schema. They must never run against hackways_prod.

No existing production records have been classified for deletion. Cleanup candidates must still be identified and approved individually after working DB access and verified recovery. Preserve the three running containers, postgres/unrelated databases, networking/SSH, credentials and backups.

## Review handoff

The companion `proposed-release.patch`, diffstat, file-status list and manifest provide the concrete source proposal against the base SHA, including untracked files and binary assets. An isolated temporary Git index is used to prepare and verify the patch; the user's real index/branch is preserved. This package is a review artifact, not a deployed commit, approval, backup or production restore.

After the owners accept the source scope, create/review the real candidate commit and run its complete CI. Do not substitute a tree object, the old base SHA or these local checks for the accepted release commit and deployment evidence.
