# Phase 6 Verification & Evidence Plan: Frontend Foundations

## 1. Checks Required Before Implementation
- Confirm Phase 5 exit gate is verified.
- Confirm backend API is responding to `GET /api/v1/events`.

---

## 2. Mandatory Verification Protocol

### Check 1: Absence of Prohibited Dependencies
- **Command:**
  ```powershell
  Get-ChildItem -Path apps -Recurse -Filter package.json | ForEach-Object { Get-Content $_.FullName | Select-String "lucide" }
  ```
- **Expected Outcome:** Empty stdout across all apps.

### Check 2: Absence of Hardcoded Mock Events
- **Command:**
  ```powershell
  Get-ChildItem -Path apps -Recurse -Include *.tsx,*.ts | ForEach-Object { Get-Content $_.FullName | Select-String "SAMPLE_EVENTS" }
  ```
- **Expected Outcome:** Empty stdout across all apps.

### Check 3: TypeScript Type Checking Across All 4 Frontends
- **Commands:**
  - `cd apps/public-web && npm run type-check`
  - `cd apps/core-web && npm run type-check`
  - `cd apps/organizer-web && npm run type-check`
  - `cd apps/admin-web && npm run type-check`
- **Expected Outcome:** Exit code 0 for all 4 commands.

### Check 4: Standalone Output Configuration Verification
- **Commands:** Verify `output: 'standalone'` exists in every `next.config.js`.

---

## 3. Exit Criteria for Phase 6
- [ ] All 4 independent Next.js apps exist and type-check cleanly.
- [ ] Mock data and browser alerts completely removed.
- [ ] Real events rendered via Server Components.
- [ ] Zero instances of `lucide-react`.
- [ ] `output: 'standalone'` configured across all 4 frontends.
