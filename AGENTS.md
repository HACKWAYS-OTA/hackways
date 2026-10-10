# AGENTS.md — Hackways Engineering Standards

This document is the authoritative engineering specification for all agents, contributors, and automated systems working on the **Hackways** platform.
Every agent must review this specification before authoring code, modifying infrastructure, or declaring a task complete.

---

## 1. Production Repository Structure

Hackways is structured as a decoupled monorepo supporting 6 core production services running on a constrained VM target (2 vCPU, 4 GB RAM):

```
hackways/
├── apps/
│   ├── public-web/           # Next.js 14 SSG/ISR (hackways.in - landing & SEO)
│   ├── core-web/             # Next.js 14 App Router (app.hackways.in - attendee & RSVP)
│   ├── organizer-web/        # Next.js 14 App Router (organizer.hackways.in - event studio)
│   └── admin-web/            # Next.js 14 App Router (admin.hackways.in - network-isolated)
├── services/
│   ├── core-backend/         # Go 1.23+ Engine (core-backend:8081 - events & RSVP locking)
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── events/       # Event aggregate & lifecycle
│   │   │   ├── rsvps/        # Concurrency-safe capacity & reservation engine
│   │   │   └── platform/     # Config, health, middleware
│   │   ├── queries/          # sqlc queries for core domain
│   │   └── Dockerfile
│   └── global-backend/       # Go 1.23+ Engine (global-backend:8082 - auth & platform)
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── auth/         # Sessions, RBAC, password security
│       │   ├── users/        # User accounts & organizer profiles
│       │   ├── notify/       # Asynchronous email & webhook workers
│       │   └── storage/      # Media upload validation & storage interface
│       ├── queries/          # sqlc queries for identity & platform
│       └── Dockerfile
├── packages/
│   ├── api-client/           # Auto-generated TypeScript client
│   └── shared-types/         # Shared domain contract types
├── deploy/
│   ├── docker/               # Production Docker Compose topologies & overrides
│   ├── nginx/                # Reverse proxy, rate limiting, and subdomain routing
│   └── scripts/              # Automated database backup, restore, and deploy scripts
├── docs/
│   ├── adr/                  # Architecture Decision Records
│   ├── planning/             # CONTEXT.md, MEMORY.md, WORKLOGS.md, IMPLEMENTATION_PLAN.md
│   ├── ARCHITECTURE_PLAN.md  # Definitive 6-service production blueprint
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── DATABASE.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   └── TESTING.md
├── .github/                  # CI workflows, templates, CODE_OF_CONDUCT, CONTRIBUTING
├── docker-compose.yml        # Local development topology (PostgreSQL + Mailpit)
├── Makefile                  # Root task automation
├── AGENTS.md                 ← this specification
└── README.md
```

---

## 2. Agent Ownership Map

| Agent Role | Owned Modules / Scopes |
| :--- | :--- |
| **Chief Architect / Supervisor** | `AGENTS.md`, `docs/planning/*`, `docs/ARCHITECTURE_PLAN.md`, API contracts |
| **Product & Domain Architect** | `docs/adr/`, product requirements, acceptance criteria |
| **Core Backend Engineer** | `services/core-backend/internal/*`, `services/core-backend/cmd/` |
| **Global Backend Engineer** | `services/global-backend/internal/*`, `services/global-backend/cmd/` |
| **PostgreSQL & Data Engineer** | Database migrations, queries, `docs/DATABASE.md` |
| **Frontend & Design Systems** | `apps/*/`, UI design tokens, client packages |
| **Security & Reliability Engineer** | `.github/SECURITY.md`, `docs/SECURITY.md`, auth/session reviews |
| **QA & Test Automation** | `**/*.test.*`, `**/*_test.go`, k6 load tests, `docs/TESTING.md` |
| **DevOps & Infrastructure** | `deploy/*`, `docker-compose.yml`, `.github/workflows/*` |

---

## 3. Language & Infrastructure Rules

### Go Backend Engines (`services/*`)
- Go 1.23+, modules enabled (`go.mod`).
- HTTP Router: Standard library `net/http` + `github.com/go-chi/chi/v5`.
- Database Driver: `jackc/pgx/v5` with connection pooling (`pgxpool`).
- Queries: Compile-time verified `sqlc` queries only. No raw SQL strings or ORMs.
- Logging: Standard library `log/slog` emitting structured JSON.
- **Resource Constraints (4 GB VM Budget):**
  - Run with `GOMEMLIMIT` environment variables (`100MiB` for core, `90MiB` for global).
  - Hard Docker memory caps: Core Backend (120M), Global Backend (110M).
  - Bounded connection pools: Core (max 28), Global (max 22). Total must not exceed 50.
- **RSVP Concurrency Invariant:**
  - Reservations MUST use pessimistic row-level locking: `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
  - Double-booking or overselling is strictly prohibited under any load condition.

### Next.js Frontends (`apps/*`)
- Next.js 14+ App Router with TypeScript strict mode.
- Tailwind CSS for all styling — no inline styles or CSS modules.
- Form state & validation: `react-hook-form` and `zod`.
- Server Components by default; Client Components only where direct DOM interaction is required.
- **Production Build Policy:**
  - `output: 'standalone'` in `next.config.js`.
  - Next.js applications MUST be compiled in GitHub Actions CI off-node. Never run `npm run build` on the 4 GB production VM.
  - Runtime memory caps: `--max-old-space-size=200` max.

### PostgreSQL 16
- Managed via numbered migrations (`services/api/migrations/` or consolidated `deploy/migrations/`).
- Primary keys: `UUID` (`gen_random_uuid()`) for entities; `BIGSERIAL` for internal link tables.
- All timestamps: `TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
- Maximum database connections: 60 (`shared_buffers = 256MB`, `work_mem = 4MB`).

---

## 4. Session & Network Security Standards

- **Host-Scoped Cookies:** Attendee/Organizer apps use `__Host-hackways_session`; Admin app uses `__Host-hackways_admin_session`.
- **Admin Isolation:** `admin.hackways.in` is restricted at the Nginx layer to VPN subnets / authorized IP ranges.
- **Rate Limiting:** Enforced via Nginx zones (`auth_limit: 5r/s`, `rsvp_limit: 10r/s`, `general_limit: 30r/s`).
- **No Secrets in Source Control:** `.env` is strictly gitignored; safe placeholders live in `.env.example`.

---

## 5. Prohibited Shortcuts

The following practices are **strictly prohibited**:
1. In-memory maps, JSON files, or mocks as substitutes for PostgreSQL on production paths.
2. Building frontend bundles directly on the production host (triggers OOM kills).
3. Optimistic client-side RSVP checks without database-level transactional locks (`SELECT ... FOR UPDATE`).
4. Storing or logging sensitive fields (passwords, tokens, cookies).
5. Using `interface{}` or `any` to bypass type safety.
6. Direct commits or unreviewed pushes to `main` or `develop`.

---

## 6. Definition of Done

A task moves to `verified` only when:
- [ ] End-to-end functionality is implemented and verified against PostgreSQL.
- [ ] Concurrency safety is maintained under load with zero race hazards.
- [ ] Server-side input validation and authorization checks are enforced.
- [ ] Automated tests pass with real commands (`go test ./... -race`, `npm test`).
- [ ] Code formatting passes (`gofmt -w ./...`, `npm run lint`).
- [ ] Documentation and work logs in `docs/planning/` are updated.
