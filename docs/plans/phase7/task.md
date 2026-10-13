# Phase 7 Task Register: Testing & CI Enforcement

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P7-TEST-001` | Implement Go backend concurrency race test (`concurrency_test.go`) | Phase 6 Exit Gate | `COMPLETED` | QA Lead | AD-005, -race |
| `P7-TEST-002` | Implement database transaction rollback integration tests | `P7-TEST-001` | `COMPLETED` | Database Lead | Real PostgreSQL |
| `P7-TEST-003` | Implement negative authorization security test runner | `P7-TEST-001` | `COMPLETED` | Security Lead | Phase 4 tests |
| `P7-TEST-004` | Configure frontend test runner & write component tests | Phase 6 Exit Gate | `COMPLETED` | Frontend Lead | Vitest/Jest |
| `P7-TEST-005` | Author k6 high-concurrency ticket drop load test script | `P7-TEST-001` | `COMPLETED` | QA Lead | M4-001 |
| `P7-TEST-006` | Harden `.github/workflows/ci.yml` across all 6 services | `P7-TEST-001`, `004` | `COMPLETED` | DevOps Lead | CI matrix |
| `P7-TEST-007` | Update `scripts/verify.sh` to execute the full verification suite | `P7-TEST-006` | `COMPLETED` | DevOps Lead | Pre-push hook |
| `P7-TEST-008` | Execute Phase 7 exit gate verification | All P7 Tasks | `COMPLETED` | Tech Lead | Phase Gate 7 |

---

## Detailed Task Specifications

### `P7-TEST-001`: Concurrency race test
- **Description:** Spawn 100 concurrent goroutines competing for 10 tickets on an event. Verify that exactly 10 transactions commit, 90 return 409 Conflict, and `current_rsvp_count == 10`. Execute with `-race`.
- **Acceptance Criteria:** 0 races detected; exactly 10 confirmed attendees in database.
- **Status:** `NOT_STARTED`.

### `P7-TEST-002`: Database integration tests
- **Description:** Verify foreign key constraints, partial unique constraints, and statement timeout cancellation.
- **Acceptance Criteria:** Tests pass cleanly against isolated test database.
- **Status:** `NOT_STARTED`.

### `P7-TEST-003`: Negative authorization test runner
- **Description:** Verify cross-organizer mutation rejection and unauthenticated request blocks.
- **Acceptance Criteria:** All negative assertions pass.
- **Status:** `NOT_STARTED`.

### `P7-TEST-004`: Frontend component tests
- **Description:** Test `EventCard` capacity progress bar rendering (80% amber, 100% rose) and button state.
- **Acceptance Criteria:** 100% component tests pass.
- **Status:** `NOT_STARTED`.

### `P7-TEST-005`: k6 ticket drop script
- **Description:** Author `deploy/tests/k6-rsvp-drop.js` simulating 1,000 VUs competing for 200 tickets.
- **Acceptance Criteria:** Verification metric `tickets_oversold == 0`.
- **Status:** `NOT_STARTED`.

### `P7-TEST-006`: CI workflow hardening
- **Description:** Update `.github/workflows/ci.yml` to run Go race tests for both backends and type-checks for all 4 frontends.
- **Acceptance Criteria:** CI matrix passes on push to `develop`.
- **Status:** `NOT_STARTED`.

### `P7-TEST-007`: `scripts/verify.sh` update
- **Description:** Update root script to run format, vet, race tests, and type checks.
- **Acceptance Criteria:** Exit code 0 on clean tree.
- **Status:** `NOT_STARTED`.

### `P7-TEST-008`: Phase 7 exit gate verification
- **Description:** Confirm all criteria in `docs/plans/phase7/verification.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 7 marked `VERIFIED`.
- **Status:** `NOT_STARTED`.
