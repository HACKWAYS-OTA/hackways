# Phase 7 Implementation Plan: Testing & CI Enforcement

## 1. Metadata
- **Phase Identifier:** Phase 7 (`phase7`)
- **Title:** Testing and CI Enforcement
- **Owner Role:** QA Lead / DevOps Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 7 establishes end-to-end automated quality controls across the repository. It builds comprehensive test suites covering Go backend concurrency with the race detector (`-race`), real PostgreSQL integration, negative authorization gates, frontend type checking, and automated load testing via k6 to guarantee the zero-overbooking invariant under extreme surges.

**Expected Outcome:**
- Go backend unit and integration test suites passing with race detector (`go test ./... -v -race -count=1`).
- Concurrency test verifying that concurrent reservations never exceed event capacity $C$.
- Frontend test runners configured across all 4 applications.
- GitHub Actions CI workflow updated to enforce hygiene, backend tests, frontend type checks, and standalone builds.
- k6 load test script verifying sub-350ms latency and zero oversold tickets.

---

## 3. Scope & Non-Goals

### In Scope:
- Go backend unit, integration, and race detection tests.
- Transactional concurrency tests simulating race conditions against PostgreSQL.
- Frontend test configuration (Vitest/Jest for components, Playwright for critical attendee flows).
- Automated CI pipeline enforcement in `.github/workflows/ci.yml`.
- k6 load testing scripts (`deploy/tests/k6-rsvp-drop.js`).

### Explicit Non-Goals:
- Fabricating test results or bypassing failing tests to obtain green CI builds ([`AGENTS.md` Rule 1.10](../../AGENTS.md#L67-L77)).
- Running heavy k6 load tests against live production environments.

---

## 4. Architectural Requirements & Constraints
- Mandatory Go race detector: `go test ./... -v -race -count=1` ([`AGENTS.md` Section 4](../../AGENTS.md#L254-L265)).
- Concurrency invariant verification: A drop of 1,000 requests against a 200-capacity event must confirm exactly 200 attendees.
- Test database isolation: Automated tests must run against isolated test databases and rollback cleanly.

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 6 (Frontend Foundations complete with real APIs).
- **Prerequisite Services:** PostgreSQL 16 test database active.

---

## 6. Ordered Implementation Steps
1. **Step 1 (Backend Race & Concurrency Tests):** Author `services/core-backend/concurrency_test.go` spawning 100 concurrent goroutines against a 10-seat event.
2. **Step 2 (Database Integration Tests):** Author tests for rollbacks, foreign keys, and unique constraint conflicts.
3. **Step 3 (Frontend Component Tests):** Configure component tests verifying event card rendering and capacity percentage display.
4. **Step 4 (k6 Load Test Suite):** Author `k6-rsvp-drop.js` simulating high-concurrency ticket drops.
5. **Step 5 (CI Workflow Hardening):** Update `.github/workflows/ci.yml` to run all checks across all 6 services.

---

## 7. Expected File Changes
- `services/core-backend/concurrency_test.go`
- `apps/*/package.json` (test scripts)
- `.github/workflows/ci.yml`
- `deploy/tests/k6-rsvp-drop.js`
- `scripts/verify.sh`

---

## 8. Verification Commands & Evidence
- `go test ./... -v -race -count=1` -> PASS with 0 race warnings.
- `scripts/verify.sh` -> 100% checks succeed.
- k6 test output showing: `tickets_oversold: 0`.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] Go backend tests pass with `-race`.
- [ ] Concurrency test proves zero capacity breach under high load.
- [ ] All 4 frontends pass type-check and component tests.
- [ ] CI pipeline runs cleanly on GitHub Actions.
