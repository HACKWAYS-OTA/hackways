# Phase 9 Agent Pre-Context: Dual Go Backend Engines Decoupling & API Completion

## 1. Product Objective & Technical Target
Phase 9 transitions the Go backend layer from the initial consolidated `services/api` structure to the canonical **dual independent Go backend engines** strictly mandated by [`AGENTS.md` Section 2.1](../../../AGENTS.md#L94-L117):
1. **`services/core-backend`**: Go 1.23+ Engine listening on port `8081` (`core-backend:8081`) managing event lifecycle, ticket inventory, and transactional RSVP pessimistic capacity locks. Dedicated `pgxpool` cap of **28 connections**.
2. **`services/global-backend`**: Go 1.23+ Engine listening on port `8082` (`global-backend:8082`) managing authentication, sessions, RBAC, Azure Blob SAS write-only signing, and running the bounded 2-goroutine ACS email transactional outbox worker. Dedicated `pgxpool` cap of **22 connections**.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Section 2.1 & Section 3](../../../AGENTS.md#L94-L140):**
   - Merging or combining `core-backend` and `global-backend` into a single binary is strictly prohibited.
   - Core Backend owns authoritative event lifecycle and capacity locking.
   - Global Backend owns sessions, RBAC, Azure Blob SAS generation, and ACS email delivery.
2. **[`AGENTS.md` Section 2.2](../../../AGENTS.md#L71-L85):**
   - Core Backend: Resident ~35 MB, Docker limit **120 MB**, CPU quota 0.70, `GOMEMLIMIT=100MiB`, `GOGC=100`.
   - Global Backend: Resident ~35 MB, Docker limit **110 MB**, CPU quota 0.40, `GOMEMLIMIT=90MiB`, `GOGC=100`.
3. **[`AGENTS.md` Section 2.4](../../../AGENTS.md#L173-L200):**
   - Pessimistic Locking Mandatory: Capacity checks and reservations execute inside PostgreSQL transaction using `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
4. **[`AGENTS.md` Section 2.5](../../../AGENTS.md#L201-L216):**
   - Max 60 DB connections total: `core-backend` max 28, `global-backend` max 22, 10 reserved.
   - Compile-time verified SQL via `sqlc` only.

---

## 3. Verified Repository Facts Before Phase 9
- Foundational logic exists in `services/api` (`internal/auth`, `internal/events`, `internal/notifications`, `internal/platform`, `internal/rsvps`, `internal/storage`).
- All 27 Go unit/integration tests pass with `-race` locally and on GitHub Actions CI.
- `deploy/docker/docker-compose.prod.yml` and `deploy/nginx/conf.d/` already specify `core-backend:8081` and `global-backend:8082`.
- Endpoints for user registration (`POST /api/v1/auth/register`), login (`POST /api/v1/auth/login`), logout (`POST /api/v1/auth/logout`), current user (`GET /api/v1/auth/me`), and upload intent (`POST /api/v1/storage/upload-intent`) need complete HTTP route wiring in `global-backend`.

---

## 4. Expected Outputs
- `services/core-backend/` independent module (`go.mod`, port 8081, pool 28, events & rsvps).
- `services/global-backend/` independent module (`go.mod`, port 8082, pool 22, auth, storage, outbox worker).
- `deploy/migrations/` established as canonical schema migrations directory.
- `scripts/verify.ps1`, `scripts/verify.sh`, and `.github/workflows/ci.yml` updated to verify both backend engines with race detection.
