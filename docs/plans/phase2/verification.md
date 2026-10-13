# Phase 2 Verification & Evidence Plan: Toolchains & Dependency Hygiene

## 1. Checks Required Before Implementation
- Confirm Phase 1 exit gate is verified.
- Confirm `apps/web/package.json` contains `"lucide-react"` (baseline check).

---

## 2. Mandatory Verification Protocol

### Check 1: Absence of `lucide-react`
- **Command:** `cd apps/web && npm ls lucide-react`
- **Expected Outcome:** `(empty)` or package not found. Exit code 1 or empty list.
- **Evidence Required:** Output confirming zero matches.

### Check 2: Standalone Output Verification
- **Inspection Command:**
  ```powershell
  Get-Content apps/web/next.config.js | Select-String "output: 'standalone'"
  ```
- **Expected Outcome:** Matches line containing `output: 'standalone'`.

### Check 3: Go Driver & Build Integrity
- **Command:** `cd services/api && go list -m github.com/jackc/pgx/v5`
- **Expected Outcome:** Returns `github.com/jackc/pgx/v5 v5.x.x`.
- **Build Command:** `cd services/api && go build -o /dev/null ./cmd/api`
- **Expected Outcome:** Exit code 0 with zero compiler warnings.

### Check 4: Frontend Type Safety
- **Command:** `cd apps/web && npm run type-check`
- **Expected Outcome:** TypeScript check completes with exit code 0.

---

## 3. Exit Criteria for Phase 2
- [ ] `lucide-react` completely removed.
- [ ] `output: 'standalone'` enabled.
- [ ] `pgx/v5` present and verified in Go backend.
- [ ] TypeScript check passing.
- [ ] Go backend builds cleanly.
