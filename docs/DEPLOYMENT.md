# Deployment and operations

## Current state

This is a repository-level runbook, not evidence of a healthy production deployment. The repository does not record a successful end-to-end deploy of the current configuration. PostgreSQL placement remains unresolved; see [ARCHITECTURE.md](ARCHITECTURE.md). ACS REST sending is implemented, but external email delivery remains unverified and requires configured credentials.

The production Compose file defines six application containers plus Nginx. It also defines a PostgreSQL container under the `local-db` profile, which is not started by the default `docker compose up`. Configured container limits total 1,800 MB including that optional 700 MB PostgreSQL cap, or 1,100 MB without it. The larger 2,295 MB number in the canonical resource table includes host/system allowance and is not the Compose container total. These are configured limits, not measured usage or proof that the VM is safe under load.

## CI and deployment path

`.github/workflows/ci.yml` selects affected frontend/backend/operations checks for PRs and branch pushes. Push checks are skipped when an open PR already owns the same branch checks. Dependency caches are enabled and superseded runs are cancelled. Documentation-only changes do not build images.

An explicitly selected `rc-*` tag runs `.github/workflows/release-candidate.yml`: full PostgreSQL 16/race, frontend and operations checks, followed by all six Linux images built on GitHub runners. Only the full commit SHA is published as an image tag; no `latest` tag is updated. Image builds run at most two at a time. Digest/test artifacts expire after one day. No workflow deploys to Azure or deletes packages. Production frontend builds on the VM remain prohibited.

Production is `main`, integration is `develop`, and new work goes through PRs. Private GitHub Free does not enforce the requested branch protections/required reviews; they remain an explicitly documented team policy. Do not claim paid-plan approval gates exist.

The checked-in deployer defaults and systemd unit now use `BRANCH=main` and `AUTO_DEPLOY_ENABLED=false`. This changes source configuration only; it does not install, enable or modify the VM service. Keep it disabled and use a separately approved manual release until all operational gates are verified.

Any future webhook activation requires a separately approved infrastructure change. It still requires signed successful `Release Candidate` workflow evidence for the configured repository/production branch. A workflow event alone is not release approval. An operator-owned `APPROVED_RELEASE_FILE` outside the checkout must select repository, branch, full lowercase commit SHA and positive increasing `generation`. Keep the directory/file write access restricted to the infrastructure/release operator and deployer as appropriate. Neither approval nor state belongs in Git. The daemon defaults to loopback; an explicitly approved reverse-proxy binding is required before external deliveries can reach it.

The deployer accepts only the currently approved SHA and persists the generation before checkout/activation. Stale/replayed generations, including after process restart, cannot activate. A valid new approved delivery arriving while busy is queued/coalesced rather than returned as HTTP 409. If a queued approval is superseded, it is rechecked and cannot activate the old SHA. Failed/interrupted attempts require a new operator generation; do not delete the state to retry or roll back. A rollback is another explicitly approved compatible SHA with a newer generation. In-memory queued events are not a durable message queue; after a process failure the operator must reconcile state and redeliver the current approval. These controls are regression-tested without real deployment operations.

The deployer fetches the configured branch, resets its checkout to the CI-verified commit, and invokes `deploy/scripts/deploy.sh`. The proposed script derives its project root from its own location and reads the root `.env`, or a protected file selected with `HACKWAYS_ENV_FILE`. This permits a separately staged checkout without replacing the existing application tree. It requires Docker Compose, Python 3 for a narrow configuration/migration utility, `migrate`, and private GHCR pull access. `DEPLOY_IMAGE_TAG` must be a full reviewed commit SHA; the script exports it as `IMAGE_TAG`, validates Compose, pulls images and applies pending migrations before starting services. Compose parses dotenv syntax; the file is never sourced as shell code. This matters for connection strings containing semicolons or other shell separators.

`run-migrations.py` consumes the resolved Compose JSON privately, verifies matching backend database targets and removes the password from the URI passed to the existing migration CLI. The password is supplied through the PostgreSQL process environment. Resolved JSON, the password-bearing URI and raw migration diagnostics are not printed. This operations utility does not introduce a Python backend or change Go/pgx/sqlc application behavior. The helper and complete image/deployment path still require rehearsal with the approved migration CLI; syntax checks are not a production migration result. The [migration PostgreSQL driver](https://github.com/golang-migrate/migrate/blob/master/database/postgres/README.md) and [its PostgreSQL environment support](https://pkg.go.dev/github.com/lib/pq) document the CLI connection behavior; no lib/pq application dependency was added.

After Docker health waiting, the proposed script probes all four frontend roots explicitly, validates Nginx and reloads it. It does not remove orphan containers or prune images. Any command failure stops the script; a partial rollout still requires operator recovery. No production builds run on the VM.

The deployment checkout must be clean of tracked changes before the deployer resets it, and the approved commit must be on the fetched production branch. Approval is rechecked before checkout and activation. Minimal deployment state is atomically recorded under `DEPLOY_STATE_FILE`; deployment logs go to the systemd journal. The approval/state files are operational selection metadata, not application persistence or database substitutes.

## Configuration hazards to resolve before production

### ACS first-send replacement review

The approved replacement distinguishes a new persisted outbox attempt from a
reclaimed attempt using `first_sent_at`, not `retry_count`. New operations POST
before polling; retries reconcile the original operation and never resend after
an ambiguous lookup. Existing repeatability identities and the conservative
replay cutoff remain intact. No database migration is required for this fix.

For this isolated fix, the explicitly applied `build-global-backend` PR label
selects the single-image `Global Backend Candidate` workflow after exact-head
mandatory CI passes. It publishes only global-backend and never deploys. Do not
trigger the six-image `rc-*` workflow. Keep the other five approved digests and
rollback images. See the [replacement proposal](handover/ACS_GLOBAL_BACKEND_REPLACEMENT.md).

### Free-event workflow rollout

Apply migration `000002_free_event_workflows` before deploying the updated engines. It preserves existing registrations and derives ticket rows for confirmed registrations; normalized guest-email collisions cause a rollback for operator reconciliation. Its down script intentionally refuses destructive enum/data removal. Review forward repair or a compatible application rollback. The local legacy database was migrated atomically with psql and has no migration-runner version table; reconcile its known schema version before using golang-migrate there. Do not force a dirty production migration.

Configure the core-only `TICKET_SIGNING_SEED` with a stable base64-encoded 32-byte Ed25519 seed. Configure global-backend `ACS_ENDPOINT`, `ACS_ACCESS_KEY`, `ACS_SENDER_ADDRESS` or the existing ACS connection-string/sender variables. No credentials means email sending stays paused and Blob intents return unavailable. Explicit production `CORS_ALLOWED_ORIGINS` must be HTTPS origins; the local HTTP example is rejected in production. Compose forwards the new variables. Go development listeners default to loopback; production containers bind internally. See [backend operation and remaining workflows](backend/README.md).

### Public website destinations

The public website uses `NEXT_PUBLIC_SITE_URL` for canonical metadata, sitemap URLs, and robots, `NEXT_PUBLIC_APP_URL` for attendee links, and `NEXT_PUBLIC_ORGANIZER_URL` for organizer links. Each value must be an absolute HTTP(S) origin without a path, query, fragment, or credentials. There are no hardcoded production-domain fallbacks. Unset values resolve to local ports 3000, 3001, and 3002 respectively.

For local work, copy `apps/public-web/.env.example` to `.env.local` if overrides are needed. For a preview or deployment, set all three to that environment's public origins before building. These values are compiled into Next.js output; changing only a running container's environment does not update already-built links. The public-web Dockerfile accepts the three matching build arguments, and the image workflow forwards the matching GitHub repository variables. Set those repository variables to the intended deployment origins before publishing an image, or pass the build arguments explicitly in a separate environment-specific build. This repository change does not configure GitHub variables or publish images.

### Attendee application (port 3001)

The attendee application now uses locally hosted FG Futurist and Google Sans Flex fonts. The core-web image copies `public/` into both supported standalone layouts so these assets remain available in the runtime container. The frontend apps remain independently built and deployed.

`apps/core-web/.env.example` documents `NEXT_PUBLIC_API_URL` (local Go engine: `http://localhost:8081/api/v1`), `NEXT_PUBLIC_APP_URL` (local attendee origin: port 3001, used for canonical metadata), `NEXT_PUBLIC_PUBLIC_URL` (local public site: port 3000), and `NEXT_PUBLIC_ORGANIZER_URL` (local organizer app: port 3002). Set these for the target environment before building; the core-web Dockerfile accepts matching build arguments. Runtime-only environment changes do not reliably change values already compiled into client JavaScript. The working-copy image workflow now forwards both browser API paths as `/api/v1`, the public origin from `NEXT_PUBLIC_SITE_URL`, and the attendee/organizer origins from their matching repository variables. Confirm those variables and the resulting bundle before publishing a release; this configuration is not proof that a current image was built or deployed.

For local browser calls, configure the Go engine's `CORS_ALLOWED_ORIGINS` to include `http://localhost:3001` (and `http://127.0.0.1:3001` if used). The current local development process was started with these origins. It uses the existing local PostgreSQL database; no records or migrations were created by the frontend redesign. Catalog reads use the existing paginated Go endpoint, and reservations still go directly to the Go RSVP endpoint. Full-text search and pagination run against the backend; the format/date filter controls were removed at the user's request. Initial catalog content is fetched on the Next.js server; API errors are rendered explicitly rather than displayed as an empty catalog.

Deploy a core engine image containing `GET /api/v1/events/by-slug/{slug}` before publishing the attendee image with the new event routes. The query uses the existing unique slug constraint; no migration is required. The original UUID endpoint remains available. Catalog loading UI is scoped to `/discover` so unresolved event slugs can return actual HTTP 404 responses before streaming the event layout.

### Other operational configuration

- Set `DOCKER_REGISTRY` to the actual lowercased GitHub owner/repository namespace. `.env.example` currently assumes `ghcr.io/varshith-dev`; CI derives the owner from the actual GitHub repository.
- Set `DATABASE_URL` explicitly. The Compose fallback targets a local `postgres` hostname even though the database service is disabled by default. Do not rely on that fallback for Azure Database for PostgreSQL.
- Configure strong production `SESSION_SECRET`, Azure storage credentials, and other secrets from the approved secret store. `.env.example` contains development placeholders and is not a production environment file.
- The admin Nginx allowlist currently uses private network ranges and loopback. It does not identify approved office public IPs and may not behave as intended behind Cloudflare.
- The public Nginx file exposes only the signed deployment webhook for the deployer. Direct access to the deployer port must also be blocked by host/network firewall rules; this repository does not prove those rules exist.
- The deployment checks container health but does not roll back a failed release. A failed migration or unhealthy new containers can leave a partial deployment requiring operator recovery.
- `backup-db.sh` and `restore-db.sh` call `docker exec hackways-postgres`; they do not use `DATABASE_URL` and cannot back up or restore the planned Azure PostgreSQL service. The backup script's “streamed to Azure” claim also depends on the Azure CLI accepting its stdin invocation, which has not been verified. There is no verified recovery path for the unresolved production database topology.
- The Compose memory/CPU limits differ from the canonical resource table in [AGENTS.md](../AGENTS.md); reconcile them before claiming the VM budget is enforced.

## Verification boundaries

Repository checks can validate source and configuration syntax, but they cannot prove Azure credentials, GitHub webhook delivery, GHCR permissions, firewall rules, database reachability, or production health. Record those results only after running the corresponding checks against the target environment. See [TESTING.md](TESTING.md) for local verification commands and known limitations.

The aspirational Azure topology and migration plan are documented separately in [ARCHITECTURE_PLAN.md](ARCHITECTURE_PLAN.md); that document is not the as-built runbook.


## Optional local Linux Actions runner

CI and release image jobs accept the repository variable HACKWAYS_CI_RUNNER_LABELS
as a JSON array of runner labels. Its absence preserves ubuntu-latest. A dedicated
self-hosted Linux runner can build the private repository without hosted-minute
usage. Restrict it to trusted repository work, provide Docker, PostgreSQL client,
C/C++ compiler and required runner libraries, and keep production credentials out
of its workspace. This does not authorize builds on the production VM. Restore
the variable and remove a temporary runner after the release so future checks do
not wait for an offline machine. No automatic Azure deployment is introduced.

## Europe email transport

The Europe email/communication resources are hackways-email-eu-prod and
hackways-comm-eu-prod, using mail.hackways.in with noreply, notify and updates.
Domain ownership, SPF, DKIM and DKIM2 are verified. India resources remain intact.
Store credentials only through the existing protected runtime configuration.
Before changing ACS_ENDPOINT/ACS_ACCESS_KEY, quiesce the worker and verify there
are no pending/processing operations with persisted first_sent_at. Such operations
are bound to the original ACS resource and must be reconciled there; blindly
replaying them in another region can duplicate delivery. Resume and verify actual
provider acceptance after configuration. This is a regional standby/cutover
procedure, not automatic failover or a whole-platform availability guarantee.


### Scoped candidate image publication

`HACKWAYS_CANDIDATE_SERVICES` optionally selects a JSON service array (default all
six) for image publication; full mandatory integration still runs. This patch selects
`global-backend` and `admin-web`, retaining the other four immutable runtime refs.
`HACKWAYS_SKIP_CACHE_EXPORT=true` skips optional cache writes after cache exporter
stalls; cache reads remain enabled. Neither variable triggers automatic deployment.
