# Phase 2 Agent Continuation & Phase Handoff

## 1. Current Phase & Status
- **Current Phase:** Phase 2 (`phase2`)
- **Status:** `VERIFIED`
- **Next Phase:** Phase 3 (`phase3`: Database & SQL Engineering)

---

## 2. Documents to Read First for Phase 3
1. [`docs/plans/phase3/plan.md`](../phase3/plan.md)
2. [`docs/plans/phase3/task.md`](../phase3/task.md)
3. [`docs/plans/phase3/implementation-guide.md`](../phase3/implementation-guide.md)
4. [`AGENTS.md`](../../../AGENTS.md) (Rule 2.4 Concurrency & RSVP Capacity Invariants, Rule 2.5 PostgreSQL 16 & Data Engineering Standards)

---

## 3. Current Verified Repository State
- Phase 1 verified.
- Phase 2 verified.
- `lucide-react` completely removed from all frontend manifests and lockfiles.
- `output: "standalone"` enabled in `apps/web/next.config.js`.
- `github.com/jackc/pgx/v5 v5.11.0` integrated into `services/api/go.mod` and compiling cleanly.
- `apps/web` and `packages/api-client` type-check passing with 0 diagnostic errors.
- `next lint` passing with 0 errors/warnings.
- Backend unit tests passing (7/7, 0.955s).

---

## 4. Phase 3 Immediate Next Actions
1. Reconcile database schema naming: update migrations from `registrations` -> `rsvps` and `notifications` -> `email_outbox` (`P3-DB-001`).
2. Update `services/api/db/queries/rsvps.sql` with compile-time verified queries including `SELECT ... FOR UPDATE` with `statement_timeout = '2000ms'` (`P3-DB-002`).
3. Run `sqlc generate` to regenerate typed Go query code (`P3-DB-003`).
4. Initialize `pgxpool.Pool` in `services/api/cmd/api/main.go` obeying the 28-connection cap (`P3-DB-004`).
5. Update health check endpoint in `services/api/internal/platform/health.go` to ping PostgreSQL (`P3-DB-005`).

---

## 5. Conditions Required Before Advancing to Phase 4
- All Phase 3 tasks `COMPLETED` and `VERIFIED`.
- Concurrency invariants strictly respected in all SQL queries.
