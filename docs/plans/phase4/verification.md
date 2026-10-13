# Phase 4 Verification & Evidence Plan: Security Foundations

## 1. Checks Required Before Implementation
- Confirm Phase 3 exit gate is verified.
- Confirm PostgreSQL `sessions` table is present and accessible.

---

## 2. Mandatory Verification Protocol

### Check 1: Bcrypt Password Hashing Test
- **Command:** `cd services/api && go test ./internal/auth -run TestPassword -v`
- **Expected Outcome:** Password hash verified with bcrypt cost >= 12.

### Check 2: Session Creation & Revocation Test
- **Command:** `cd services/api && go test ./internal/auth -run TestSession -v`
- **Expected Outcome:** Token hashed with SHA-256; database stores only `token_hash`; instant revocation on logout verified.

### Check 3: Host Cookie Attribute Verification
- **Command:** `cd services/api && go test ./internal/auth -run TestCookie -v`
- **Expected Outcome:** Cookie string contains `HttpOnly`, `SameSite=Lax`, `Path=/`, and `__Host-` prefix in production mode.

### Check 4: Negative Security Test Suite
- **Command:** `cd services/api && go test ./internal/auth -run TestNegativeSecurity -v -race`
- **Expected Outcome:**
  - Tampered cookie -> 401 Unauthorized.
  - Expired token -> 401 Unauthorized.
  - Non-organizer mutating event -> 403 Forbidden.
  - Unprivileged user hitting admin route -> 403 Forbidden.

---

## 3. Exit Criteria for Phase 4
- [ ] Bcrypt password hashing passes with cost factor >= 12.
- [ ] Stateful sessions operational via PostgreSQL.
- [ ] `__Host-` cookies emitted correctly.
- [ ] Negative security tests pass 100%.
- [ ] Zero secret leakage in log tests.
