# Phase 4 Task Register: Security Foundations

## Tasks Checklist

| Task ID | Title | Prerequisites | Status | Owner | Blockers / References |
|:---:|:---|:---:|:---:|:---:|:---|
| `P4-SEC-001` | Implement bcrypt password hashing & verification (cost >= 12) | Phase 3 Exit Gate | `COMPLETED` | Security Lead | `golang.org/x/crypto` |
| `P4-SEC-002` | Implement PostgreSQL session store with SHA-256 token hashing | `P4-SEC-001` | `COMPLETED` | Backend Lead | `sessions` table |
| `P4-SEC-003` | Implement `__Host-` scoped cookie helpers (Lax & Strict) | `P4-SEC-002` | `COMPLETED` | Backend Lead | AD-004 |
| `P4-SEC-004` | Implement `RequireAuth` context injection middleware | `P4-SEC-003` | `COMPLETED` | Backend Lead | Chi middleware |
| `P4-SEC-005` | Implement `RequireOrganizer` and `RequireAdmin` RBAC guards | `P4-SEC-004` | `COMPLETED` | Backend Lead | Role checks |
| `P4-SEC-006` | Implement server-side ownership validator for organizer events | `P4-SEC-005` | `COMPLETED` | Backend Lead | IDOR mitigation |
| `P4-SEC-007` | Implement negative security tests for unauthorized / cross-tenant access | `P4-SEC-006` | `COMPLETED` | QA / Security | Automated tests |
| `P4-SEC-008` | Execute Phase 4 exit gate verification | All P4 Tasks | `COMPLETED` | Tech Lead | Phase Gate 4 |

---

## Detailed Task Specifications & Verified Evidence

### `P4-SEC-001`: Bcrypt password hashing
- **Description:** Implement `HashPassword(password string) (string, error)` and `CheckPasswordHash(password, hash string) bool` using bcrypt with cost 12.
- **Acceptance Criteria:** Unit tests pass verifying valid password match and invalid rejection.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/auth/password.go` implemented and verified by `password_test.go` confirming cost >= 12 and length validations (8 to 72 bytes).

### `P4-SEC-002`: PostgreSQL session store
- **Description:** Create session records in PostgreSQL with 32-byte crypto-random tokens; hash tokens with SHA-256 prior to database storage (`token_hash`).
- **Acceptance Criteria:** Lookup by plaintext token finds matching hashed row; expired sessions reject cleanly.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/auth/session.go` implemented and verified against live PostgreSQL instance via `session_test.go` (create, lookup, update last seen, revoke).

### `P4-SEC-003`: `__Host-` cookie helpers
- **Description:** Implement cookie emission with names `__Host-hackways_session` (Lax) and `__Host-hackways_admin_session` (Strict), setting `Path=/`, `HttpOnly=true`, and `Secure=true`.
- **Acceptance Criteria:** Set-Cookie headers verified via unit test.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/auth/cookie.go` implemented and verified via `TestCookieHelpers` in `session_test.go`.

### `P4-SEC-004`: `RequireAuth` middleware
- **Description:** Read cookie from incoming request, validate session against PostgreSQL, and attach authenticated `User` to `r.Context()`.
- **Acceptance Criteria:** Missing or invalid cookie yields HTTP 401 with `UNAUTHORIZED` envelope.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/auth/middleware.go` implements `RequireAuth` and verified via `TestRequireAuth_NegativeTests` in `middleware_test.go`.

### `P4-SEC-005`: RBAC middleware
- **Description:** Inspect authenticated user's organizer and admin roles.
- **Acceptance Criteria:** Non-organizer accessing organizer routes receives HTTP 403 `FORBIDDEN`.
- **Status:** `COMPLETED`.
- **Evidence:** `internal/auth/middleware.go` implements `RequireOrganizer` and verified via `TestRequireOrganizer_NegativeTests` in `middleware_test.go`.

### `P4-SEC-006`: Server-side ownership validation
- **Description:** Validate that mutating an event requires `organizer_id == authenticated_user_organizer_id` inside the SQL query (`WHERE id = $1 AND organizer_id = $2`).
- **Acceptance Criteria:** Attempting to update another organizer's event returns HTTP 404 or 403 without data mutation.
- **Status:** `COMPLETED`.
- **Evidence:** `ValidateEventOwnership` query generated and implemented in `internal/auth/middleware.go`.

### `P4-SEC-007`: Negative security tests
- **Description:** Automated tests executing tampered cookies, expired sessions, and cross-organizer updates.
- **Acceptance Criteria:** 100% of negative tests pass.
- **Status:** `COMPLETED`.
- **Evidence:** 7 test suites in `internal/auth` pass 100% verifying unauthenticated rejections, tampered tokens, and cross-role prohibitions.

### `P4-SEC-008`: Phase 4 exit gate verification
- **Description:** Confirm all criteria in `docs/plans/phase4/verification.md`.
- **Acceptance Criteria:** Exit gate approved; Phase 4 marked `VERIFIED`.
- **Status:** `COMPLETED`.
- **Evidence:** All exit gate conditions met and verified with real database assertions.
