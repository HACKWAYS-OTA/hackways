# Hackways Architecture

**Status:** Target architecture with implementation gaps called out below.
**Canonical constraints:** [AGENTS.md](../AGENTS.md)

## Service boundaries

The production topology has four independently deployed Next.js applications and two Go services. PostgreSQL is the authoritative store; Nginx is the only public ingress.

| Service | Directory | Responsibility | Current implementation status |
|---|---|---|---|
| Public web | `apps/public-web` | Marketing and links to product apps | Next.js app exists |
| Main app | `apps/core-web` (3001) | Discovery, registration, My events, account and event creation | Frontend pages call the two Go services |
| Event management | `apps/organizer-web` (3002) | Existing-event details, visibility, guest review and check-in | Management pages call owner-scoped Go APIs |
| Admin web | `apps/admin-web` | Restricted platform administration | Next.js app exists |
| Core backend | `services/core-backend` | Events and transactional RSVP capacity | Go service exists; current HTTP routes include event reads and RSVP create/read |
| Global backend | `services/global-backend` | Sessions, auth and upload SAS issuance | Go service exists; current HTTP routes include auth and storage |

**Topology decision is unresolved:** [AGENTS.md](../AGENTS.md) budgets a PostgreSQL container on the VM, while [ARCHITECTURE_PLAN.md](ARCHITECTURE_PLAN.md) and `deploy/docker/docker-compose.prod.yml` describe Azure Database for PostgreSQL as the production database. Do not infer that both deployments are supported until the team selects one and reconciles the config and resource budgets.

Global Go owns account/browser moderation and sign-in quarantine; both engines enforce database restrictions and retired access. See [the restriction contract](backend/ACCOUNT_DEVICE_RESTRICTIONS.md). Administrator event deletion/recovery retains existing records; see [Bin](backend/ADMIN_BIN.md).

Managed Space/event operation uses persisted role assignments enforced by core Go; global Go independently checks delegated media rights. Ownership remains unchanged and public membership is separate. See [the workspace contract](backend/WORKSPACE.md).

`packages/api-client` and `packages/shared-types` are intended shared contracts. They do not own business logic. TypeScript applications call Go APIs; authorization, persistence, validation, and RSVP invariants remain in Go and PostgreSQL.

`packages/ui` is the shared frontend control package, linked through root npm workspaces. Each independent Next.js root installs its provider/styles and transpiles the source; checkbox, radio group, select and toggle behavior is configured once. Base UI and Motion retain React 18 compatibility; this does not upgrade Next.js or introduce frontend business endpoints. See [the shared controls guide](../packages/ui/README.md).

The four frontend Dockerfiles install the root locked workspace with `npm ci`, build the shared API client, and then independently compile/package their target app. CI image construction is off the production VM; registry availability and standalone runtime contents must be verified for the proposed release.

## Canonical browser URLs

The main application is served on `hackways.in` alongside public-web, with event management at `/manage`. Nginx routes reserved app paths to core-web and `/manage` to organizer-web. Core assets use `/_core/_next/`; public assets use `/_next/`; management retains `/manage/_next/`. The independent services remain intact. Legacy app and organizer hosts only redirect. Auth and admin retain their dedicated hosts.

## Runtime request paths

```text
Browser -> Nginx -> matching Next.js application
Browser/app -> Nginx -> core-backend (events, RSVPs)
Browser/app -> Nginx -> global-backend (auth, storage SAS)
core-backend and global-backend -> PostgreSQL via pgxpool and sqlc queries
global-backend outbox worker -> configured email provider
Browser -> Azure Blob Storage using a short-lived write-only SAS
```

Public Blob media has a prepared custom read origin, `https://cdn.hackways.in`.
See [the CDN configuration and activation checks](../deploy/cloudflare/README.md).
Cloudflare Cloud Connector routes public reads directly to Azure; signed uploads
and server-side upload verification retain the Azure host. Compose forwards
`AZURE_CDN_DOMAIN` with an empty default. This is prepared configuration, not a live
cutover: Cloudflare access and HTTPS/content checks are still pending.

Internal backend and database ports must not be published publicly. By explicit operator decision on 2026-10-22, admin sign-in is reachable from any network; dedicated host-scoped admin sessions, explicit database grants, daily OTP and backend authorization protect the console and APIs. Nginx retains TLS/rate limits and restricts readiness diagnostics to internal networks. Production credentials belong in Azure Key Vault and runtime environment, never source control.

## Data and concurrency

First-party analytics remains in global-backend/PostgreSQL, with optional browser
consent, bounded public ingestion and existing admin authorization on reports.
Core-backend owns event/registration aggregates. See [analytics and audience
contracts](backend/ANALYTICS.md) for retention, quotas, report semantics and deployment.

The only checked-in migration sequence is `deploy/migrations/`. Both backend sqlc configurations must compile against that schema. Generated query code is committed with its SQL source under each service's `db/` directory.

RSVP creation locks the event row inside a PostgreSQL transaction, sets a local 2-second statement timeout, checks publication and capacity, inserts idempotently, updates the count, writes the email outbox row, and commits. Do not move these checks to the clients or make capacity decisions from cached counts.

RSVP detail reads must require an authenticated session and scope the query to the requesting user. Organizer roster reads must additionally check event ownership in the backend.

## Resource and deployment boundaries

- Four web apps and two Go services remain separate deployable units.
- `deploy/deployer` is an auxiliary host-level systemd HTTP service outside those six application services. Treat it as an extra deployment control-plane process, not part of the six-service product topology; consolidating deployment control into the CI runner would remove this extra listener and execution path.
- Frontend builds run in CI; production hosts only pull prebuilt standalone Next.js images.
- Backend pgx pools remain bounded by the 28-connection core and 22-connection global allocations, with ten connections reserved for administration and migrations.
- Upload bytes go directly from browser to Azure Blob Storage. Do not persist media on VM disk or proxy file bodies through Go.
- Transactional email starts from an outbox row written with the RSVP. A provider failure must leave the row retryable and must never be reported as successful delivery.

## Known implementation gaps

These are verified from the current source tree and must not be confused with completed architecture:

1. ACS REST sending and operation reconciliation are implemented, but actual external delivery is unverified and local credentials are missing. Sending is paused without consuming retries when unconfigured. No mounted runtime uses logged email as delivery success.
2. Outbox claims use atomic `UPDATE ... FOR UPDATE SKIP LOCKED`, a five-minute lease, and a lease token on outcome writes. Two entries are claimed for two sends. Stable ACS operation IDs are reconciled before resend; unknown outcomes outside the replay window require operator review. Exactly-once inbox delivery is not claimed.
3. The free-event APIs include revision-checked editing/status changes, scoped roster/audit/check-in, registration management and custom registration questions with answer snapshots. Payments, admin moderation, scheduled publishing and offline synchronization remain. See [the implementation matrix](backend/README.md).
4. Public guest RSVP creation remains supported, and verified-email account claiming is now implemented; direct guest RSVP retrieval is not: the detail endpoint requires an account session and only returns a row owned by that account. RSVP creation now reads registration windows under the same event-row lock and optional-authenticates attendees so account-backed RSVP ownership and notification email come from the session. Attendees may edit the registration name without modifying their account profile.
5. Upload intent verifies event ownership or binds avatars to the session user. Event banner completion verifies Azure Blob properties before an owner/revision-scoped future-event metadata attachment in global-backend. Event publication remains in core-backend. Actual Azure calls remain unverified locally because credentials are missing; no media bytes are buffered through either engine.
6. Event creation and host profile setup live in core-web on port 3001. The organizer web on 3002 handles existing events only. Both consume the same owner-scoped Go APIs; moving a page does not move its business logic. Authenticated browser create/edit journeys still require end-to-end verification.
7. Admin has independent Home, Events, Spaces, People, Registrations, Tickets, Moderation, Messaging, Storage, Audit and Sign-in pages plus record details. Core/global endpoints use generated SQL, explicit database admin grants and dedicated sessions. Reviewed Go actions cover event/space lifecycle, permitted registration changes, check-in, person name/access/session controls, eligible outbox retry and image metadata detachment; existing locks, ownership, versions, audits and idempotency remain authoritative. Independent moderation reports/decisions/appeals remain unimplemented. Migrations 000009/000010/000011 and production activation are pending; deployment of the accumulated approved changes is authorized. See [the current admin contract](planning/ADMIN_DASHBOARD.md).
8. `services/api` and `apps/web` in older planning notes refer to the superseded monolithic layout. Current code lives in the six directories listed above; historical phase documents are archival and are not current runbooks.
9. `deploy/scripts/backup-db.sh` and `restore-db.sh` target the optional local PostgreSQL container, not the Azure database topology described by the deployment plan. No end-to-end backup or restore path is verified for the production database.
10. Core-web has registration/sign-in, account verification/guest claim and My events pages. Organizer-web retains sign-in for its own host-scoped management session. Production subdomains intentionally do not share `__Host-` cookies, so a host may need to sign in separately when opening management. Tokens are never forwarded in URLs or browser storage.

## Source of truth

Spaces, membership and owner-only event association now live in core-backend/PostgreSQL, with directory/profile/settings in core-web. This adds no service or framework. Event themes, light/dark mode and public/link-only visibility are persisted scalar fields. Discoverability is filtered in Go/sqlc queries, not by hiding frontend elements. Existing event management remains on port 3002; creation remains on 3001. Public browser routes use slugs and short registration references; private APIs retain UUID transport compatibility. Admin page URLs and visible identifiers use positive numeric record references, resolved by Go/sqlc to internal UUIDs. Migration 000011 adds identity references on seven existing tables without replacing their UUID keys or foreign keys and deliberately refuses rollback of reference columns to preserve links. See API.md and migrations 000003–000006/000011.

- Architecture constraints and verification rules: [AGENTS.md](../AGENTS.md)
- Database schema and migration sequence: [DATABASE.md](DATABASE.md), `deploy/migrations/`
- Actual HTTP routes and cookie behavior: [API.md](API.md), backend router setup
- Local and CI verification: [TESTING.md](TESTING.md)
- Architecture roadmap: [ARCHITECTURE_PLAN.md](ARCHITECTURE_PLAN.md), subject to reconciliation with these implementation gaps


## Local admin data review

The local admin console can run the existing two Go engines against an isolated, private PostgreSQL copy exported read-only from production. `ADMIN_PREVIEW_ONLY` is constrained to loopback and the `hackways_admin_preview` database. Normal mutation routes and background workers are excluded in that mode; admin sessions/grants are local to the copy. This adds no production service, JSON data store, ORM or direct TypeScript database access. See the [admin preview contract](planning/ADMIN_DASHBOARD.md) for the copy boundary, startup and evidence.


Confirmed registration emails now render an inline signed QR plus a calendar
attachment in global-backend. Its generated eligibility query reads the actual
issued ticket and current recipient; core-backend retains authoritative one-use
admission and revocation. Both engines use the same protected admission signing
seed and wire format; no browser credential storage or new service is added.
See [the email contract](engineering/email-templates.md#confirmed-event-passes-2026-10-23).

## Remote MCP

The existing global Go engine hosts `mcp.hackways.in/mcp` and OAuth; it calls the
existing authoritative Go HTTP APIs. No seventh service or TypeScript business
backend is introduced. Separate consent/action/revocation pages render in core-web
and admin-web, with a public-web landing page at `/mcp`. See [MCP](backend/MCP.md)
and the [tool catalog](backend/MCP_TOOLS.md). Enable only after migration 24 and
verified images; Go 1.25 is limited to global-backend's official MCP SDK requirement.


## City Hubs and community branding

City Hubs are admin-curated city pages, separate from owner-managed membership Spaces. Core Go owns metadata, publication and public event visibility. Global Go extends the existing Azure SAS/HEAD-completion flow for Space logos/covers and admin city banners. Additive migrations 27/28 preserve existing workflows. No new service or framework is introduced. See [the contract](backend/CITY_HUBS.md) and [surface guide](design/SPACES_CITY_HUBS.md).
