# Phase 7 Agent Pre-Context: Testing & CI Enforcement

## 1. Product Objective & Technical Target
Phase 7 establishes automated verification across all 6 services. Its central mission is to **empirically verify that the system never breaches event capacity under race conditions**.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 1.10](../../../AGENTS.md#L67-L77):**
   - **Do not fabricate test results, build output, or successful integrations.**
   - Execute relevant verification commands and report actual outcomes.
   - Inspect errors and regressions instead of suppressing them.
2. **[`AGENTS.md` Section 4](../../../AGENTS.md#L254-L265):**
   - Go backend tests must run with the race detector: `go test ./... -v -race -count=1`.
   - Frontend type checks & tests: `npm test && npm run type-check`.
   - Code formatting & linting: `gofmt -w ./...`, `go vet ./...`, and `npm run lint`.
3. **[`docs/TESTING.md`](../../TESTING.md):**
   - Real integration tests against Docker PostgreSQL 16.
   - k6 load tests simulating ticket drop surges.

---

## 3. Verified Repository Facts Before Phase 7
- Go backend unit tests currently pass (7 tests in `services/api/internal/platform/`).
- `apps/web/package.json` currently lacks a test runner script (`"test"` script missing).
- `.github/workflows/ci.yml` currently tests Go backend and runs `npm run type-check`.

---

## 4. Expected Outputs
- Automated concurrency test proving zero double-booking under load.
- Passing Go backend tests with `-race`.
- Passing frontend component and type-check tests.
- Hardened CI workflow covering all 6 services.
