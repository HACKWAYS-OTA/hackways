# Phase 3 Agent Pre-Context: Database & SQL Engineering

## 1. Product Objective & Technical Target
Phase 3 establishes the authoritative data layer for Hackways in PostgreSQL 16. It ensures capacity invariants are strictly enforced via compile-time verified queries (`sqlc`) and pessimistic row-level locking.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 2.4](../../../AGENTS.md#L182-L200):**
   - Capacity reservations MUST execute inside an isolated transaction using `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'`.
   - Strict idempotency via `UNIQUE(event_id, user_id)`.
   - Client-side checks are cosmetic only.
2. **[`AGENTS.md` Rule 2.5](../../../AGENTS.md#L202-L215):**
   - Maximum 60 connections on PostgreSQL (`core-backend` max 28, `global-backend` max 22, reserved 10).
   - Compile-time verified SQL via `sqlc` only. **Raw SQL strings and ORMs are strictly prohibited.**
3. **[`AGENTS.md` Rule 1.1](../../../AGENTS.md#L9-L16):**
   - No mock data or in-memory stores as substitutes for PostgreSQL.

---

## 3. Verified Repository Facts Before Phase 3
- Existing migration [`000001_initial_schema.up.sql`](../../../services/api/migrations/000001_initial_schema.up.sql) names tables `registrations` and `notifications`.
- `sqlc.yaml` is configured for package `db`, sql_package `pgx/v5`.
- Query files exist in `services/api/db/queries/` but have not been compiled into Go.
- `services/api/cmd/api/main.go` has zero database pool initialization.

---

## 4. Expected Outputs
- Updated migration files using `rsvps` and `email_outbox`.
- Generated Go models in `services/api/db/`.
- `pgxpool.Pool` initialized in backend startup.
- Real database readiness check in `/health`.
