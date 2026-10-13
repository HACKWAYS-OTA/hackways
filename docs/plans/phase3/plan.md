# Phase 3 Implementation Plan: Database & SQL Engineering

## 1. Metadata
- **Phase Identifier:** Phase 3 (`phase3`)
- **Title:** Database and SQL Engineering
- **Owner Role:** Database Lead / Backend Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 3 establishes the authoritative PostgreSQL 16 persistence layer, compiles type-safe SQL queries via `sqlc`, resolves naming discrepancies between migrations and specifications, establishes connection pooling limits, and implements the pessimistic locking RSVP capacity engine (`SELECT ... FOR UPDATE` with 2000ms statement timeout).

**Expected Outcome:**
- PostgreSQL migrations harmonized (`rsvps` and `email_outbox` tables).
- Numbered migration files (`.up.sql` and `.down.sql`) reversible and verified.
- `sqlc generate` executed, emitting type-safe Go models in `services/api/db/`.
- `pgxpool.Pool` initialized with connection limits (core max 28, global max 22, total <= 60).
- Concurrency-safe capacity queries compiled and validated.

---

## 3. Scope & Non-Goals

### In Scope:
- Updating `000001_initial_schema.up.sql` to align with `rsvps` and `email_outbox` naming.
- Harmonizing query files in `services/api/db/queries/*.sql`.
- Running `sqlc generate` to create type-safe Go database bindings.
- Configuring `pgxpool` with strict timeouts (`statement_timeout = '2000ms'`, connection pool max conns).
- Implementing database health check pings in `/health`.

### Explicit Non-Goals:
- Writing HTTP route handlers or authentication controllers (Phase 4).
- Implementing Azure Blob SAS signing (Phase 5).
- Building frontend views (Phase 6).

---

## 4. Architectural Requirements & Constraints
- Maximum 60 connections across the PostgreSQL instance ([`AGENTS.md` Rule 2.5](../../AGENTS.md#L202-L215)).
- Compile-time parameterized SQL via `sqlc` only. **No raw SQL strings, ORMs, or SQLGlot.**
- Capacity checks and RSVP reservations must execute inside an isolated transaction using `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'` ([`AGENTS.md` Rule 2.4](../../AGENTS.md#L182-L200)).

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 2 (Toolchain Baseline with `pgx/v5`).
- **Prerequisite Decisions:** DEC-004 (Schema table naming reconciliation approved).
- **External Prerequisites:** PostgreSQL 16 running locally (via `docker compose up -d` or local service).

---

## 6. Ordered Implementation Steps
1. **Step 1 (Schema Harmonization):** Update `000001_initial_schema.up.sql` and `down.sql` to use table `rsvps` (with `current_rsvp_count`) and `email_outbox`.
2. **Step 2 (Query Updates):** Update `db/queries/rsvps.sql` to implement atomic reservation with `ON CONFLICT (event_id, user_id) DO NOTHING` and transactional row locking.
3. **Step 3 (sqlc Code Generation):** Run `sqlc generate` in `services/api`. Verify generated files in `services/api/db/`.
4. **Step 4 (Connection Pool Wiring):** Initialize `pgxpool.Pool` in `services/api/cmd/api/main.go` using configuration parameters from `platform/config.go`.
5. **Step 5 (Health Check Upgrade):** Update `platform/health.go` to execute `pool.Ping(ctx)`.

---

## 7. Expected File Changes
- `services/api/migrations/000001_initial_schema.up.sql` & `.down.sql`
- `services/api/db/queries/events.sql`, `rsvps.sql`, `users.sql`
- Generated files in `services/api/db/` (`db.go`, `models.go`, `*.sql.go`)
- `services/api/internal/platform/health.go`
- `services/api/cmd/api/main.go`

---

## 8. Verification Commands & Evidence
- `make migrate-up` -> Migrations apply cleanly.
- `make migrate-down` -> Rollback succeeds without leaving orphan types or tables.
- `cd services/api && sqlc generate` -> Generates Go files with 0 errors.
- `curl http://localhost:8080/health` -> Returns `{"status":"ok","database":"healthy"}`.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] Schema naming aligned with `AGENTS.md` and `ARCHITECTURE_PLAN.md`.
- [ ] Numbered migrations apply and rollback cleanly.
- [ ] `sqlc` generates type-safe Go structs.
- [ ] Database pool configured with statement timeouts and max 28 connections.
- [ ] Health check reflects live database connectivity.
