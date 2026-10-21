# Verification

The CI workflow is the executable verification contract. It starts a disposable PostgreSQL 16 service, applies all numbered `deploy/migrations/*.up.sql` files in order, then runs `go vet` and race-enabled Go tests for the two backend modules. It type-checks and lints each existing Next.js app and the shared controls.

## Go services

Run from each service directory:

```powershell
go vet ./...
go test ./... -v -race -count=1
```

Database-backed tests require an isolated disposable database initialized with the current migration. Do not point tests at production or a developer database with valuable data. Some tests skip when PostgreSQL is unavailable; a skip is not evidence that database behavior passed.

## Frontend applications

Install once from the repository root with `npm ci` (or `npm install` while intentionally updating dependencies). The root workspace lockfile is authoritative for local development, frontend CI and the four independent frontend Docker builds. Each Docker dependency stage copies the workspace manifests and shared packages, uses `npm ci`, and builds the API client before building its own app. `packages/ui` is linked into all four apps; their normal type checks include its imported source. Shared control usage and defaults are documented in [packages/ui/README.md](../packages/ui/README.md). Container image construction still needs actual CI verification; application checks alone do not verify the Docker COPY and runtime layout.

Run from each `apps/<app>` directory:

```powershell
npm run type-check
npm run lint
```

The `test` script currently aliases those two checks; there is no separate component or browser test suite configured. Do not report end-to-end coverage that the repository does not have.

For release review, production builds can run from an isolated local source copy so `.next` output does not disturb active development servers. Record the Node/dependency environment and any differences from CI. Such builds do not replace Linux Docker image construction, standalone runtime checks or browser verification. The four configs explicitly trace from the workspace root so hoisted dependencies/shared packages can be packaged independently.

## SQL generation

After changing query sources, run `sqlc generate` from each affected backend directory. The configs use the canonical schema in `deploy/migrations/`. Inspect generated diffs and compile the affected Go module.

## Reporting

The management extensions are covered by the existing lifecycle suite: owner/anonymous isolation for summary and roster, literal guest search, status/event-view validation, actual summary counts, revision-checked appearance and removal, and removal of a scanned guest from the not-arrived queue. Standard Go tests pass with the isolated schema; the local race command still fails before execution because CGO is disabled. HTTP redirects and compilation do not establish a signed-in browser walkthrough or real external Azure completion.

The existing owner lifecycle suite now also exercises Space creation replay/conflicts, member-versus-owner access, archive/join behavior, event association, theme/display persistence, optional public content, and exclusion of link-only events from Discover search and Space feeds. It reuses the existing isolated fixture helper. Migrations 000003–000006 must be present before running it. Passing backend checks do not establish browser calendar/wheel interaction or signed-in creation, and no new browser suite has been introduced.

Report exact commands, exit status, skipped integration checks, and unavailable services. A clean type-check or Go compilation does not prove authorization, capacity invariants, Azure delivery, or complete user workflows.

## Free-event workflow verification (2026-10-19)

The backend workflow suites require an explicitly isolated DATABASE_URL. Tests no longer fall back to the application database. The shared test guard accepts a database ending in `_test` or a dedicated `workflow_*_ci_*` schema without a public-schema fallback. Existing tests can delete outbox rows; never point them at application data. The configured Linux CI uses its own PostgreSQL-16 test database and applies every numbered up migration in order.

Local execution used separate core/global schemas on PostgreSQL 18.6 with search_path excluding public. `go test ./... -v -count=1` passed in both engines. Coverage includes actual SQL-backed capacity contention, replay/approval/waitlist transitions, revision/ownership/cancellation, transfer and ticket/check-in checks. `go test ./... -v -race -count=1` could not execute because CGO is disabled and no C compiler is configured. A normal passing suite is not a race-detector result. Source generation, builds and vet also passed. External ACS/Blob calls, production deployment, browser interactions and multi-process stress are separate unverified dependencies.

## Registration form verification (2026-10-20)

Migration 000007 is required for both engines before generation/deployment. The existing lifecycle suite now covers owner-only form editing, anonymous/cross-owner guest-detail denial, registration/event matching, required answer enforcement, stale form conflicts, short question codes, persisted snapshot labels after edits/removal, typed answer rejection and exact replay after a form revision changes. It reuses the existing isolated owner fixture. No production/application fixtures were introduced.

Local sqlc generation, Go builds/vet, both standard isolated-schema suites and both frontend type-check/lint scripts passed. The race command failed before execution because CGO is disabled. Both changed CSS files parse and scoped design detection reports no findings. New management routes compile and redirect unauthenticated requests to sign-in; actual public registration/form reads respond. This is not evidence of signed-in drawer keyboard/mobile interaction or external Azure delivery. Local PostgreSQL is 18.6, while mandatory Linux CI targets PostgreSQL 16.
## Owner guest-status verification (2026-10-20)

Extended the existing mandatory lifecycle and outbox tests using their existing isolated fixtures. The lifecycle case verifies anonymous/cross-owner denial, silent cancellation, exact replay and payload conflict, stale guest rejection, real inventory/ticket changes, restoration with a persisted custom email message, and rejection after check-in. The outbox case also checks plain text retention and HTML escaping of the custom message. Both full standard Go suites passed with `search_path` restricted to their existing separate workflow CI schemas; no application records were used. Logs: `logs/guest-status-core-verification.log` and `logs/guest-status-global-verification.log`.

Both Go builds and vet passed. Both race commands failed before execution because CGO is disabled. Organizer type checks/lint passed; the status stylesheet parses and scoped design detection returned no findings. Updated local backends report ready health checks, and the new endpoint rejects an anonymous request with 401. Authenticated browser focus/mobile behavior and actual ACS delivery remain unverified.
