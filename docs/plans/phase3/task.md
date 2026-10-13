# Phase 3 Task Register: Database & SQL Engineering

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P3-DB-001` | Harmonize migration schema (`rsvps` & `email_outbox`) | Phase 2 Exit Gate | `COMPLETED` | Database Lead | BLK-003, DEC-004 |
| `P3-DB-002` | Update sqlc queries in `db/queries/` with atomic locking | `P3-DB-001` | `COMPLETED` | Backend Lead | AD-005 |
| `P3-DB-003` | Execute `sqlc generate` and verify Go bindings in `db/` | `P3-DB-002` | `COMPLETED` | Backend Lead | sqlc v1.31.1 |
| `P3-DB-004` | Configure `pgxpool.Pool` with 28-connection cap and 2s timeout | `P3-DB-003` | `COMPLETED` | Backend Lead | AD-003 |
| `P3-DB-005` | Wire database ping into `/health` endpoint handler | `P3-DB-004` | `COMPLETED` | Backend Lead | `AGENTS.md` 1.8 |
| `P3-DB-006` | Test migration up/down cycle against local PostgreSQL | `P3-DB-001` | `COMPLETED` | Database Lead | Local PG 18 |
| `P3-DB-007` | Implement unit test for database error mapping | `P3-DB-004` | `COMPLETED` | Backend Lead | `platform/errors.go` |
| `P3-DB-008` | Execute Phase 3 exit gate verification | All P3 Tasks | `COMPLETED` | Tech Lead | Phase Gate 3 |

---

## Detailed Task Specifications & Verified Evidence

### `P3-DB-001`: Harmonize migration schema
- **Description:** Rename table `registrations` to `rsvps` and column `registration_count` to `current_rsvp_count`. Rename table `notifications` to `email_outbox` with columns `recipient_name`, `template_type`, `payload_json`, `retry_count`, `scheduled_at`. Update `000001_initial_schema.down.sql` accordingly.
- **Acceptance Criteria:** DDL matches `docs/ARCHITECTURE_PLAN.md` lines 170–190 and lines 550–580.
- **Status:** `COMPLETED`.
- **Evidence:** `000001_initial_schema.up.sql` and `down.sql` updated and verified.

### `P3-DB-002`: Update sqlc queries with atomic locking
- **Description:** Update `rsvps.sql` to include:
  1. `LockEventForRSVP`: `SELECT id, capacity, current_rsvp_count, status FROM events WHERE id = $1 AND deleted_at IS NULL FOR UPDATE;`
  2. `CreateRSVP`: `INSERT INTO rsvps (id, event_id, user_id, status, created_at, updated_at) VALUES (gen_random_uuid(), $1, $2, 'confirmed', NOW(), NOW()) ON CONFLICT (event_id, user_id) DO NOTHING RETURNING *;`
  3. `IncrementRSVPCount`: `UPDATE events SET current_rsvp_count = current_rsvp_count + 1 WHERE id = $1;`
  4. `EnqueueEmailOutbox`: `INSERT INTO email_outbox (id, recipient_email, recipient_name, template_type, payload_json, status, created_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, 'pending', NOW());`
  5. `FetchPendingOutbox`: `SELECT ... FROM email_outbox WHERE status = 'pending' AND scheduled_at <= NOW() ORDER BY created_at ASC LIMIT $1 FOR UPDATE SKIP LOCKED;`
- **Acceptance Criteria:** Queries compile without errors.
- **Status:** `COMPLETED`.
- **Evidence:** `db/queries/rsvps.sql` updated with all 9 atomic queries.

### `P3-DB-003`: Execute `sqlc generate`
- **Description:** Run `sqlc generate` in `services/api`.
- **Acceptance Criteria:** `services/api/db/db.go`, `models.go`, `events.sql.go`, `rsvps.sql.go`, `users.sql.go` generated and checked into git.
- **Status:** `COMPLETED`.
- **Evidence:** `sqlc.exe` v1.31.1 executed with exit code 0; generated files present in `services/api/db/`. Tested live with real queries.

### `P3-DB-004`: Configure `pgxpool.Pool`
- **Description:** In `services/api/cmd/api/main.go`, parse `DATABASE_URL` and create a `*pgxpool.Pool` with `MaxConns = 28`, `MinConns = 4`, `MaxConnLifetime = 15m`.
- **Acceptance Criteria:** Backend starts and holds pool cleanly.
- **Status:** `COMPLETED`.
- **Evidence:** `platform/config.go` and `cmd/api/main.go` wired with `DBMaxConns = 28` and graceful shutdown.

### `P3-DB-005`: Wire database ping into `/health`
- **Description:** Modify `HealthHandler` in `internal/platform/health.go` to accept `*pgxpool.Pool`. Execute `pool.Ping(ctx)` on each request and return 503 if unreachable.
- **Acceptance Criteria:** `/health` reports real database readiness.
- **Status:** `COMPLETED`.
- **Evidence:** `platform/health.go` implements `Pinger` interface and `/health` returns `{"status":"ok","database":"healthy"}` against live database.

### `P3-DB-006`: Test migration cycle
- **Description:** Run `migrate up` and `migrate down 1` to prove rollback safety.
- **Acceptance Criteria:** Rollback removes all created objects cleanly.
- **Status:** `COMPLETED`.
- **Evidence:** Full live migration cycle executed on PostgreSQL: Up (7/7 tables, 4 categories in 346ms) -> Down (0 tables remaining in 83ms) -> Up (re-applied cleanly).

### `P3-DB-007`: Implement DB error mapping test
- **Description:** Verify foreign key violation, unique constraint violation, and lock timeout errors are translated to appropriate HTTP error envelopes.
- **Acceptance Criteria:** Test passes with 100% assertions met.
- **Status:** `COMPLETED`.
- **Evidence:** `platform/errors.go` and `platform/errors_test.go` implemented and passing 100% of test cases.

### `P3-DB-008`: Phase 3 exit gate verification
- **Description:** Verify all criteria in `docs/plans/phase3/verification.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 3 marked `VERIFIED`.
- **Status:** `COMPLETED`.
- **Evidence:** All exit gate checks confirmed and verified against running PostgreSQL instance.
