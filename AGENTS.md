# AGENTS.md — Hackways Authoritative Engineering & Quality Enforcement Rules

This document is the **canonical, permanent engineering specification and rule set** for all AI agents, automated coding assistants, contributors, and developers working on the **Hackways** platform.

Every agent **must read, understand, and strictly enforce** these rules before authoring code, modifying files, altering infrastructure, or declaring any task complete.

---

## 1. Hackways Production Quality Enforcement Rules

### 1.1 Real Functionality, Real Data
- Never use mock data, dummy records, fabricated API responses, placeholder business logic, or simulated success states in application functionality.
- Never replace a missing backend integration with hardcoded data or a fake implementation.
- All features must use the actual approved APIs, PostgreSQL database, and required external services.
- Never report a feature as complete unless its real functionality has been implemented and verified.
- If a dependency, credential, or service is unavailable, report the blocker honestly.
- Do not create test suites or test fixtures unless explicitly authorized or required by the existing mandatory CI workflow.

### 1.2 Architecture and Technology Discipline
- Preserve the four independent Next.js applications and two independent Go backend engines.
- Go owns authoritative business logic, authorization, database operations, and API behavior.
- TypeScript and TSX are for frontend implementation. Do not introduce duplicate TypeScript backend routing or business logic inside Next.js.
- Python is permitted only for narrowly scoped utilities such as email delivery or PDF generation when appropriate. Do not turn it into a second general-purpose backend.
- PostgreSQL is the authoritative database. Follow the approved `pgx` and `sqlc` workflow.
- **Never introduce `lucide-react` anywhere in Hackways.**
- Do not introduce unnecessary frameworks, ORMs, SQL translators, JSON-to-SQL engines, database adapters, routing layers, state-management libraries, or abstraction frameworks. **SQLGlot, `pgglot`, Psycopg, and `psycosql` are strictly prohibited.**
- JSON is a data interchange format, not a reason to add another database or query framework.
- Inspect the existing stack before proposing dependencies. Obtain explicit approval before changing architectural boundaries or approved technologies.

### 1.3 Data Integrity and Security
- Enforce authentication, authorization, role checks, resource ownership, and input validation on the backend. Frontend visibility is never a security boundary.
- Use parameterized SQL, approved migrations, database constraints, and transactions.
- Preserve RSVP capacity invariants, concurrency safety, and idempotency. Client-side checks are never authoritative.
- Prevent cross-user and cross-organizer data access.
- Never expose secrets, session credentials, internal service ports, or privileged operations to unauthorized clients.
- Never silently delete, reset, or overwrite real data.
- Do not use local JSON files or in-memory maps as substitutes for PostgreSQL or Azure Blob Storage in production.

### 1.4 First-Render UX and Freshness
- Diagnose the actual rendering and data-loading path before fixing stale content, flicker, missing modules, or layout shifts.
- Avoid rendering fabricated defaults or outdated values as if they were current.
- Fetch essential data on the server before rendering when practical and consistent with the existing architecture.
- Do not delay essential content until unnecessary client-side initialization or sequential API requests finish.
- Fetch independent resources concurrently where appropriate.
- Apply deliberate caching and revalidation policies based on data freshness, authorization, and privacy requirements.
- Never serve stale personalized permissions, RSVP status, or organizer data as authoritative.
- Use clear loading skeletons only when data is genuinely pending. Preserve layout space to avoid unnecessary shifts.
- Prevent older network responses from overwriting newer state.
- Avoid unnecessary full-page refreshes, component remounts, and duplicate requests.
- Do not disable all caching or force every request to PostgreSQL as a blanket fix.

### 1.5 Complete User Journeys
- Implement the complete requested workflow, not just the visible screen or happy path.
- Consider creation, editing, validation, cancellation, deletion, permissions, empty states, errors, retries, and recovery where relevant.
- A success notification must follow confirmed success from the actual operation.
- Preserve user input when recoverable operations fail.
- Make retries safe for operations that may be repeated.
- Ensure related views reflect consistent persisted state after mutations.
- Do not silently omit requirements that are difficult to implement.

### 1.6 Maintainable Code and Controlled Changes
- Inspect existing code, callers, dependencies, data contracts, and configuration before modifying files.
- Make the smallest safe change that solves the actual problem.
- Reuse existing approved patterns instead of creating duplicate implementations.
- Avoid giant components, duplicated business rules, circular dependencies, dead code, unnecessary wrappers, and speculative abstractions.
- Do not rewrite working routing, authentication, database access, sessions, or API contracts without explicit approval.
- Do not combine unrelated features, dependency upgrades, and broad refactors in one change.
- Preserve backward compatibility unless a breaking change has been explicitly approved.

### 1.7 Performance and Resource Limits
- Respect the approved Azure VM limits of 2 vCPU, 4 GB RAM, and 50 GB NVMe disk.
- Never run production frontend builds on the production VM when the approved architecture requires external CI builds.
- Avoid unbounded database queries, excessive API requests, oversized frontend bundles, unnecessary polling, and unbounded connection pools.
- Paginate large datasets and load expensive content only when appropriate.
- Inspect database query plans and indexes when performance evidence indicates a problem.
- Do not claim throughput, concurrency capacity, or resource safety without measurements.
- Treat the single VM as a shared failure domain. One service must not consume all resources needed by the others.

### 1.8 Failure Handling and Observability
- Handle expected network, database, authentication, and external-service failures explicitly.
- Do not swallow errors or report success after a failed operation.
- Use bounded retries, timeouts, and backoff where appropriate. Avoid retry storms and duplicate side effects.
- Record actionable server-side errors with useful request context without logging secrets or sensitive personal data.
- Ensure health checks reflect actual service readiness rather than merely an open port.
- Document unresolved failure modes and operational dependencies.

### 1.9 Accessibility and Responsive Behavior
- Preserve semantic HTML, keyboard accessibility, visible focus states, appropriate contrast, and accessible form labels.
- Ensure layouts work across supported mobile, tablet, and desktop viewports.
- Avoid interface patterns that depend exclusively on hover, color, or animation.
- Respect reduced-motion preferences and prevent avoidable content shifts.
- Reuse the existing design system and component patterns instead of introducing unnecessary UI dependencies.

### 1.10 Verification and Honest Reporting
- Do not fabricate test results, build output, deployment status, database responses, or successful integrations.
- Execute the relevant available verification commands and report their actual outcomes.
- Inspect errors and regressions instead of suppressing them to obtain a passing build.
- Where tests are explicitly authorized or required by mandatory CI, keep test fixtures isolated from production data.
- If verification cannot run, identify the exact missing service, credential, permission, or environment.
- A successful compilation alone does not prove correct authorization, data integrity, concurrency safety, or complete user workflows.

### 1.11 Mandatory Workflow Before Every Change
1. Read the relevant `AGENTS.md` rules and architecture documentation.
2. Inspect the actual implementation and reproduce or trace the reported problem where possible.
3. Identify the root cause and the smallest compliant solution.
4. Check security, data integrity, first-render behavior, failure handling, resource impact, and dependency changes.
5. Implement only the authorized change.
6. Run relevant verification commands and inspect their actual output.
7. Review the final diff for unrelated changes, fake functionality, exposed secrets, and architectural violations.
8. Update documentation when contracts, architecture, schemas, configuration, or deployment behavior change.
9. Report what changed, what was verified, and what remains unresolved.

*If the request conflicts with these rules, stop before making the conflicting change. Explain the conflict and obtain explicit approval.*

---

## 2. Non-Negotiable Architectural Constraints

### 2.1 The Six Independent Production Services
Hackways is a decoupled monorepo composed of **four independent Next.js web applications** and **two Go backend engines**. Merging, combining, or blurring the boundaries of these services is **strictly prohibited**.

```
hackways/
├── apps/
│   ├── public-web/           # Next.js 14 SSG/ISR (hackways.in - landing, marketing & SEO)
│   ├── core-web/             # Next.js 14 App Router (app.hackways.in - attendee discovery & RSVP)
│   ├── organizer-web/        # Next.js 14 App Router (organizer.hackways.in - event studio & roster)
│   └── admin-web/            # Next.js 14 App Router (admin.hackways.in - network-isolated admin)
├── services/
│   ├── core-backend/         # Go 1.23+ Engine (core-backend:8081 - events & RSVP capacity locking)
│   └── global-backend/       # Go 1.23+ Engine (global-backend:8082 - auth, storage SAS & ACS email)
├── packages/
│   ├── api-client/           # Shared auto-generated TypeScript API client
│   └── shared-types/         # Shared TypeScript domain contracts
├── deploy/
│   ├── docker/               # Production Docker Compose topologies
│   ├── nginx/                # Reverse proxy, rate limiting, and subdomain routing
│   └── scripts/              # Automated database backup, restore, and deploy scripts
└── docs/                     # Canonical architecture, database, security, and planning docs
```

| Service | Technology | Ingress Route / Domain | Primary Responsibilities |
| :--- | :--- | :--- | :--- |
| **Public Web** | Next.js 14 SSG/ISR | `hackways.in` | Public landing pages, marketing, SEO metadata, pre-rendered catalogs. |
| **Core Web** | Next.js 14 App Router | `app.hackways.in` | Attendee event discovery, details, tickets, and interactive RSVP flows. |
| **Organizer Web** | Next.js 14 App Router | `organizer.hackways.in` | Event creation studio, ticket tier controls, attendee roster, CSV exports. |
| **Admin Web** | Next.js 14 App Router | `admin.hackways.in` *(VPN/IP only)* | Platform auditing, user moderation, event controls, health overview. |
| **Core Backend** | Go 1.23+ (`chi` + `pgx/v5`) | `core-backend:8081` | Event lifecycle, ticket inventory, transactional RSVP capacity engine. |
| **Global Backend**| Go 1.23+ (`chi` + `pgx/v5`) | `global-backend:8082`| Sessions, RBAC, Azure Blob SAS signing, ACS Email outbox worker. |

---

### 2.2 Azure VM Resource Budget (2 vCPU, 4 GB RAM, 50 GB NVMe)
The production environment runs on a single Azure VM (`Standard_B2s` or `Standard_D2s_v5`). All configurations must obey these hard resource caps:

| Component | Resident (RSS) | Docker Hard Limit | CPU Quota (Max 2.0) | Memory Governance Strategy |
| :--- | :--- | :--- | :--- | :--- |
| **PostgreSQL 16** | ~220 MB | **700 MB** | 0.60 | `shared_buffers = 256MB`, `work_mem = 4MB`, `max_connections = 60` |
| **Core Backend (Go)** | ~35 MB | **120 MB** | 0.70 | `GOMEMLIMIT=100MiB`, `GOGC=100` |
| **Global Backend (Go)** | ~35 MB | **110 MB** | 0.40 | `GOMEMLIMIT=90MiB`, `GOGC=100`, bounded 2-routine email worker |
| **Core Web (Next.js)** | ~110 MB | **250 MB** | 0.50 | `NODE_OPTIONS="--max-old-space-size=190"`, standalone output |
| **Organizer Web (Next.js)**| ~100 MB | **230 MB** | 0.40 | `NODE_OPTIONS="--max-old-space-size=180"`, standalone output |
| **Admin Web (Next.js)** | ~80 MB | **170 MB** | 0.25 | `NODE_OPTIONS="--max-old-space-size=130"`, standalone output |
| **Public Web (Next.js)** | ~60 MB | **140 MB** | 0.25 | `NODE_OPTIONS="--max-old-space-size=100"`, standalone output |
| **Nginx Ingress** | ~25 MB | **75 MB** | 0.20 | Event loop worker processes: 2, worker_connections: 4096 |
| **Linux Host & System** | ~350 MB | ~500 MB | 0.20 | Systemd, SSH, journald, Docker daemon |
| **Ceiling Total** | **~1,215 MB** | **~2,495 MB** | **~3.50 (Burstable)**| **Safe Headroom: ~1,601 MB for disk page cache & spikes** |

#### Strict Host Rules:
1. **Off-Node Next.js Builds:** **NEVER run `npm run build` or `next build` on the production VM.** Next.js compilation consumes up to 2.5 GB of RAM and triggers kernel OOM kills. All container images MUST be compiled in GitHub Actions CI with `output: 'standalone'` and pulled as pre-built runtimes.
2. **NVMe Swap Safety Net:** The host VM must have a **3.0 GB swapfile** configured on NVMe with `vm.swappiness = 10` as an emergency buffer.
3. **Docker Daemon Log Rotation (50 GB Disk Cap):** All containers must enforce log limits via `/etc/docker/daemon.json` (`max-size: "10m"`, `max-file: "3"`). Total container log footprint must never exceed 210 MB.

---

### 2.3 Cloud Persistence & Messaging (Azure Managed Services)

#### Azure Blob Storage (Zero-Disk Upload Architecture)
- **NO PERSISTENT MEDIA ON LOCAL VM DISK.** The 50 GB NVMe disk is strictly reserved for the operating system, Docker images, and PostgreSQL data.
- Uploads (event covers, attendee avatars, organizer exports) MUST use **direct client-to-blob upload via Write-Only Shared Access Signature (SAS) tokens** generated by `global-backend` with a 15-minute expiration.
- Uploaded media MUST NEVER be proxied or buffered through Go application RAM.
- Public read traffic is served directly from Azure Blob Storage or Cloudflare Edge CDN (`media-uploads` container).
- Database backups are streamed directly to the private `backups` container in Azure Blob Storage via `deploy/scripts/backup-db.sh`.

#### Azure Communication Services (ACS) Email
- **NO LOCAL SMTP CONTAINERS.** Do not run local mail transfer agents (Postfix, Sendmail, Exim) or heavy Celery/Redis queues on the VM.
- Transactional emails (RSVP confirmations, magic links, notifications) MUST use the **PostgreSQL Transactional Outbox Pattern**:
  - The confirmation endpoint writes a row to `email_outbox` inside the same database transaction that confirms the RSVP.
  - A bounded, 2-goroutine worker in `global-backend` polls pending outbox entries and dispatches them via the Azure Communication Services REST API / Go SDK.
  - Failures use exponential backoff up to 5 retries.

#### Secrets Management & Azure Key Vault
- **ZERO SECRETS IN SOURCE CODE.** No API keys, connection strings, database passwords, or session secrets in Git.
- `.env` is strictly gitignored; safe documentation templates live in `.env.example`.
- In production, credentials originate from **Azure Key Vault** and are injected via environment variables or Azure VM Managed Identity.

---

### 2.4 Concurrency & RSVP Capacity Invariants

The central domain invariant of Hackways: **An event with capacity $C$ must NEVER confirm more than $C$ attendees under any concurrent load.**

#### Invariant Rules:
1. **Pessimistic Locking Mandatory:** Capacity checks and RSVP reservations MUST execute inside an isolated PostgreSQL transaction using:
   ```sql
   SET LOCAL statement_timeout = '2000ms';
   SELECT id, capacity, current_rsvp_count, status
   FROM events
   WHERE id = $1 AND deleted_at IS NULL
   FOR UPDATE;
   ```
2. **Strict Idempotency:** Double-registrations by the same user are prevented by database-enforced unique constraints on `(event_id, user_id)`:
   ```sql
   INSERT INTO rsvps (id, event_id, user_id, status, created_at, updated_at)
   VALUES (gen_random_uuid(), $1, $2, 'confirmed', NOW(), NOW())
   ON CONFLICT (event_id, user_id) DO NOTHING;
   ```
3. **No Optimistic Client Bypasses:** Client-side checks or cache-based checks are strictly cosmetic. Database-level row locking is the sole authority for capacity.
4. **Statement Timeout Protection:** All lock transactions MUST specify `statement_timeout = '2000ms'` to fail fast during extreme surges, preventing connection pool exhaustion.

---

### 2.5 PostgreSQL 16 & Data Engineering Standards

- **Maximum Connections:** Hard cap of **60 connections** on the PostgreSQL instance.
- **Connection Pool Allocation:**
  - `core-backend` (`pgxpool`): Max 28 connections.
  - `global-backend` (`pgxpool`): Max 22 connections.
  - Reserved for admin, migrations, CLI: 10 connections.
- **Query Verification:** **Compile-time verified SQL via `sqlc` only.** Raw SQL strings, unparameterized queries, and runtime ORMs (GORM, Ent) are strictly prohibited.
- **Migrations:** All schema alterations must be managed via sequentially numbered migration scripts (`deploy/migrations/` or `services/api/migrations/`).
- **Destructive Changes Prohibited:** No dropping tables, deleting production columns, or altering types in a single step without backward-compatible transitional migrations.

---

### 2.6 Network, Session & Admin Isolation Standards

- **Nginx Ingress Routing:** Nginx is the sole public-facing HTTP/HTTPS port on the VM. Internal backend ports (8081, 8082) and PostgreSQL (5432) must NEVER be exposed directly to the public internet.
- **Rate Limiting Zones:** Nginx must enforce rate limits:
  - Auth zone: `5r/s` (burst=5 nodelay) on `/api/v1/auth/*`
  - RSVP zone: `10r/s` (burst=5 nodelay) on `/api/v1/rsvps/*`
  - General zone: `30r/s` (burst=20 nodelay) on all other paths
- **Admin Isolation:** `admin.hackways.in` MUST be restricted at the Nginx layer to approved office static IPs and VPN subnets (`allow <subnet>; deny all;`).
- **Cryptographic Cookie Scoping:**
  - Public & Attendee Sessions: `__Host-hackways_session` (`Secure; HttpOnly; SameSite=Lax; Path=/`)
  - Platform Admin Sessions: `__Host-hackways_admin_session` (`Secure; HttpOnly; SameSite=Strict; Path=/`)
  - Binding to `__Host-` prevents cross-subdomain cookie hijacking.

---

## 3. Prohibited Shortcuts & Anti-Patterns

The following actions are **strictly prohibited** for all agents and contributors:
1. **Silently Changing Architecture:** Modifying the 6-service structure, introducing unapproved frameworks, or altering API routing without explicit user authorization.
2. **Merging Services:** Combining `core-backend` and `global-backend` into a single binary, or combining frontend apps into a single Next.js project.
3. **Introducing Heavy Dependencies:** Adding Redis, Kafka, RabbitMQ, MongoDB, Elasticsearch, or Kubernetes without documented justification and user approval.
4. **In-Memory Mocks in Production Paths:** Using in-memory maps, local JSON files, or mock drivers in place of PostgreSQL or Azure Blob Storage.
5. **Storing Uploads on VM Disk:** Creating local `/uploads` folders on the VM instead of using Azure Blob Storage SAS tokens.
6. **Bypassing Database Locks:** Using non-locking `SELECT COUNT(*)` or client-side checks for RSVP capacity enforcement.
7. **Exposing Secrets:** Committing `.env` files, credentials, connection strings, or private keys to source control.
8. **Destructive Database Scripts:** Executing `DROP TABLE`, `TRUNCATE`, or destructive unreviewed SQL scripts against production.
9. **Building on Production VM:** Invoking `npm run build` on the Azure VM host.
10. **Bypassing Type Safety:** Using `any` in TypeScript or `interface{}` in Go to bypass type checking.
11. **Claiming Unverified Success:** Stating that a test, migration, build, or deployment passed without executing real verification commands.
12. **Introducing Prohibited Libraries:** Never import or install `lucide-react`, `SQLGlot`, `pgglot`, `Psycopg`, or `psycosql`.

---

## 4. Definition of Done & Verification Protocol

A task or pull request moves to `verified` only when the following checklist is completed:

- [ ] **Real Functionality:** Real requested functionality is implemented with real PostgreSQL data and Go APIs. Zero mock data or placeholder logic.
- [ ] **Architecture Compliant:** Preserves 6 independent services, Azure VM limits, and domain boundaries.
- [ ] **Database Verified:** All queries generated via `sqlc`; migrations numbered and reversible.
- [ ] **Concurrency Safe:** Capacity logic uses transactional `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
- [ ] **Azure Integrations Adhered:** Media uses Azure Blob SAS tokens; email uses ACS outbox pattern.
- [ ] **Memory Bounds Respected:** Go engines obey `GOMEMLIMIT`; Next.js configured with `output: 'standalone'`.
- [ ] **Automated Tests Pass:** Real commands executed and verified:
  - Go backend tests with race detector: `go test ./... -v -race -count=1`
  - Frontend type checks & tests: `npm test && npm run type-check`
- [ ] **Code Formatting & Linting Pass:**
  - Go formatting: `gofmt -w ./...` and `go vet ./...`
  - TypeScript linting: `npm run lint`
- [ ] **First-Render UX:** Layout space preserved, server rendering used where appropriate, zero stale-content flashes.
- [ ] **Documentation Updated:** Changes logged in `docs/planning/WORKLOGS.md` and relevant `docs/` files.
