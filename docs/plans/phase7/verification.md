# Phase 7 Verification & Evidence Plan: Testing & CI Enforcement

## 1. Checks Required Before Implementation
- Confirm Phase 6 exit gate is verified.
- Confirm PostgreSQL test database is ready.

---

## 2. Mandatory Verification Protocol

### Check 1: Go Race Detector Execution
- **Command:** `cd services/api && go test ./... -v -race -count=1`
- **Expected Outcome:** 100% PASS with zero data race warnings.

### Check 2: Concurrency Lock Test
- **Command:** `cd services/api && go test -v -race -run TestRSVPConcurrencyCapacity`
- **Expected Outcome:** Exactly $C$ reservations confirmed, remaining rejected with 409 Conflict. Zero overbooking.

### Check 3: Frontend Component Tests
- **Command:** `cd apps/core-web && npm test`
- **Expected Outcome:** All component assertions pass.

### Check 4: CI Workflow Syntax & Simulation
- **Command:** `gh act` or push to test branch on GitHub.
- **Expected Outcome:** All CI jobs (`repo-hygiene`, `backend-ci`, `frontend-ci`) complete green.

### Check 5: Root Pre-Push Verification Suite
- **Command:** `./scripts/verify.sh`
- **Expected Outcome:** Outputs `ALL CHECKS PASSED SUCCESSFULLY.` with exit code 0.

---

## 3. Exit Criteria for Phase 7
- [ ] Go tests pass with `-race`.
- [ ] Concurrency test proves capacity safety under load.
- [ ] Frontend tests passing across all applications.
- [ ] CI pipeline fully operational on GitHub.
- [ ] `scripts/verify.sh` passes 100%.
