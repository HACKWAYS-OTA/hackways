# Phase 1 Verification & Evidence Plan: Discovery & Architecture Alignment

## 1. Checks Required Before Implementation
- Confirm working directory is repository root: `C:\Users\ADMIN\Desktop\RSVP-P`.
- Verify Git branch is `develop`.

---

## 2. Mandatory Verification Protocol

### Check 1: Go Test Suite Integrity
- **Command:** `cd services/api && go test ./... -v -count=1`
- **Expected Outcome:** 7 passed tests in `internal/platform/`. Exit code 0.
- **Evidence Required:** Full test runner stdout showing all tests marked `PASS`.

### Check 2: Go Static Analysis
- **Command:** `cd services/api && go vet ./...`
- **Expected Outcome:** Zero warnings or errors. Exit code 0.

### Check 3: Go Formatting Check
- **Command:** `cd services/api && gofmt -l .`
- **Expected Outcome:** Empty stdout once `.gitattributes` LF normalization is applied.

### Check 4: Planning Package Completeness
- **Verification Script / Check:**
  Verify that all 6 master documents and 64 phase documents exist:
  ```powershell
  Get-ChildItem -Path docs/plans -Recurse -Filter *.md | Measure-Object
  ```
- **Expected Outcome:** Exactly 70 markdown files in `docs/plans/`.

### Check 5: Working Tree Hygiene
- **Command:** `git status`
- **Expected Outcome:** No untracked `LICENSE` file containing unrelated third-party text.

---

## 3. Exit Criteria for Phase 1
- [x] All 7 Go unit tests verified passing.
- [x] `go vet` passes cleanly.
- [ ] All 70 plan documents created with complete repository-specific content.
- [ ] Untracked `LICENSE` anomaly resolved.
- [ ] Phase 1 `status.md` and `handoff.md` updated with evidence.
