# AGENTS.md — Hackways Engineering Standards

This file is the authoritative reference for all agents working on the Hackways platform.
Every agent must read this before writing code, making architectural decisions, or claiming a task complete.

---

## 1. Repository Structure

```
hackways/
├── apps/
│   └── web/                  # Next.js 14 App Router (TypeScript, Tailwind)
├── services/
│   └── api/                  # Go application (chi, pgx, sqlc)
│       ├── cmd/api/          # main.go entrypoint
│       ├── internal/
│       │   ├── auth/         # Session management, middleware
│       │   ├── users/        # User accounts and profiles
│       │   ├── organizers/   # Organizer profiles and permissions
│       │   ├── events/       # Event domain (create, edit, publish)
│       │   ├── rsvps/        # Registration and RSVP workflows
│       │   ├── notifications/# Notification dispatch
│       │   ├── platform/     # Health, config, shared middleware
│       │   └── storage/      # File/asset storage interface
│       ├── migrations/       # SQL migration files (up/down)
│       └── db/               # sqlc-generated query code
├── packages/
│   └── api-client/           # Generated TypeScript API client (if needed)
├── docs/
│   ├── adr/                  # Architecture Decision Records
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── DATABASE.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   └── TESTING.md
├── scripts/                  # Dev/ops scripts
├── docker-compose.yml
├── .env.example
├── AGENTS.md                 ← this file
├── MEMORY.md
├── CONTEXT.md
├── WORKLOGS.md
└── IMPLEMENTATION_PLAN.md
```

---

## 2. Agent Ownership Map

| Agent Role                        | Owned Modules / Files                                                    |
|-----------------------------------|--------------------------------------------------------------------------|
| Chief Architect / Supervisor      | AGENTS.md, CONTEXT.md, WORKLOGS.md, IMPLEMENTATION_PLAN.md, docs/ARCHITECTURE.md, contracts |
| Product & Domain Architect        | docs/adr/, product requirements, acceptance criteria                     |
| Go Backend Engineer               | services/api/internal/*, services/api/cmd/                               |
| PostgreSQL & Data Engineer        | services/api/migrations/, services/api/db/, docs/DATABASE.md             |
| Frontend & Design Systems         | apps/web/                                                                 |
| Security & Reliability Engineer   | docs/SECURITY.md, cross-cutting security review                          |
| QA & Test Automation Engineer     | **/*.test.*, **/*_test.go, docs/TESTING.md                               |
| DevOps & Release Engineer         | docker-compose.yml, Dockerfile*, .env.example, scripts/, CI config       |
| Independent Code Reviewer         | Review all cross-module or cross-agent work; no self-approval            |

**Agents must not edit files outside their agreed ownership without supervisor coordination.**

---

## 3. Language and Framework Rules

### Go (services/api/)
- Go 1.23+, modules enabled (`go.mod`)
- HTTP: `net/http` + `chi` router
- Database: `pgx/v5` + `sqlc`-generated queries
- Migrations: `golang-migrate/migrate`
- Structured logging: `log/slog` (stdlib)
- Configuration: environment variables via `os.Getenv` + a config struct; no config files committed with secrets
- No global mutable state
- No fat handlers — handlers call service functions; service functions call repository functions
- Use `context.Context` everywhere for cancellation and deadlines
- Interfaces only at genuine boundaries (storage, email, external providers)
- Graceful shutdown on SIGTERM/SIGINT

### Next.js (apps/web/)
- Next.js 14+ App Router
- TypeScript strict mode (`"strict": true`)
- Tailwind CSS for all styling — no CSS Modules, no inline styles
- `zod` for schema validation (forms and API response parsing)
- `react-hook-form` for form state
- Server Components by default; Client Components only where interaction requires them
- No client-side global state stores unless justified (avoid Redux/Zustand for server data)
- All protected API calls go through server actions or route handlers that validate sessions
- No hardcoded user IDs, tokens, or credentials in frontend code

### Database
- PostgreSQL 16+
- All schema changes via numbered migration files (`000001_*.up.sql` / `000001_*.down.sql`)
- `sqlc` generates all query code — no raw SQL strings in application code
- All queries in `services/api/db/queries/*.sql`

---

## 4. Formatting and Linting

### Go
```bash
gofmt -w ./...
go vet ./...
# golangci-lint run (when installed)
```

### TypeScript / Next.js
```bash
npm run lint        # ESLint
npm run type-check  # tsc --noEmit
npm run format      # Prettier
```

All formatting must pass before a task is marked `verified`.

---

## 5. Testing Commands and Acceptance Requirements

### Go backend
```bash
cd services/api
go test ./... -race -count=1
go test ./... -race -count=1 -run Integration  # integration tests (requires DB)
```

### Frontend
```bash
cd apps/web
npm test           # Vitest unit/component tests
npm run test:e2e   # Playwright end-to-end
```

### Migrations
```bash
migrate -path services/api/migrations -database "$DATABASE_URL" up
migrate -path services/api/migrations -database "$DATABASE_URL" down 1
```

**A task is not verified until its tests pass with real commands and real output.**

---

## 6. Database and API Conventions

### Database
- Primary keys: `UUID` (`gen_random_uuid()`) for user-facing entities; `BIGSERIAL` acceptable for internal junction tables
- Timestamps: `TIMESTAMPTZ NOT NULL DEFAULT NOW()` for `created_at`; `TIMESTAMPTZ NOT NULL` for `updated_at` (updated by trigger or app)
- Soft deletes: Use `deleted_at TIMESTAMPTZ` where records must not be hard-deleted (users, events)
- All event times stored as `TIMESTAMPTZ`; `timezone TEXT NOT NULL` stored alongside for display
- Snake_case column names
- Never store passwords in plaintext — use `bcrypt` hash in a `password_hash` column

### API
- Base path: `/api/v1`
- JSON request/response bodies
- Consistent error envelope:
  ```json
  { "error": { "code": "EVENT_NOT_FOUND", "message": "Event not found", "details": {} } }
  ```
- HTTP status codes: 200 OK, 201 Created, 204 No Content, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 409 Conflict, 422 Unprocessable Entity, 500 Internal Server Error
- Pagination: cursor-based for large sets; `?limit=20&cursor=<opaque>` on list endpoints
- All list endpoints have a default and maximum page size
- Versioned at `/api/v1` — breaking changes require a new version

---

## 7. Security Expectations

- **No secrets in source control.** Use `.env` (gitignored) locally; environment variables in production.
- **All protected endpoints require server-validated sessions.** Frontend visibility checks are cosmetic only.
- **CSRF protection** on all state-changing cookie-authenticated endpoints.
- **Rate limiting** on auth, RSVP, and file upload endpoints.
- **SQL injection prevention**: use `sqlc`-generated parameterized queries; no `fmt.Sprintf` into SQL.
- **File uploads**: validate MIME type and file size server-side; store in object storage, not local disk.
- **Sensitive data** (emails, personal info) must never appear in logs.
- **Session cookies**: `HttpOnly`, `Secure` (in production), `SameSite=Lax`.
- **bcrypt** for password hashing (cost factor ≥ 12).
- **Least-privilege DB credentials**: app user has SELECT/INSERT/UPDATE/DELETE only; migrations run as a separate user.

---

## 8. Git and Collaboration Rules

- Branch naming: `feat/<slug>`, `fix/<slug>`, `chore/<slug>`, `docs/<slug>`
- Commit messages: imperative mood, ≤72 chars subject, blank line, then body if needed
- Do not push directly to `main`; use PRs with at least one review
- Never commit: `.env`, `*.pem`, credentials, build artifacts, `node_modules/`, `vendor/` (Go modules cached elsewhere)
- Each PR must reference a task ID from IMPLEMENTATION_PLAN.md

---

## 9. Prohibited Shortcuts

The following are **strictly prohibited** in any implementation:

1. JSON files or in-memory maps as a substitute for PostgreSQL
2. Hardcoded user IDs, tokens, or fake event data on production paths
3. Mock API responses returned to real users
4. Placeholder buttons, non-functional forms, or fake success messages
5. `TODO: implement` stubs merged as "complete" without flagging as incomplete
6. `interface{}` / `any` used to skip type safety
7. Skipping database migrations — schema changes must have a migration file
8. Printing or logging sensitive fields (passwords, tokens, full session secrets)
9. Bypassing authorization with `// TODO: add auth check`
10. Claiming a feature is done without executing the verification commands

---

## 10. Definition of Done

A task moves to `verified` only when ALL applicable conditions are met:

- [ ] Intended behavior is implemented end-to-end
- [ ] Backend and frontend contracts agree (schema, status codes, field names)
- [ ] Real data persisted and retrieved from PostgreSQL
- [ ] Input validation and authorization enforced on the server
- [ ] Edge cases handled (duplicate, empty, capacity limit, unauthorized)
- [ ] Tests written and passing (`go test ./... -race` or `npm test`)
- [ ] Loading, error, empty, and success UI states implemented
- [ ] Database migrations included for schema changes
- [ ] No mock persistence on production path
- [ ] No placeholder controls in delivered workflow
- [ ] Implementation reviewed by at least one other agent or the supervisor
- [ ] WORKLOGS.md updated with outcome, commands run, and test results
- [ ] Remaining limitations explicitly documented

A successful compile is not done. A visually complete page is not done. An agent's assertion is not evidence.
