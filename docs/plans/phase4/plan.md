# Phase 4 Implementation Plan: Security Foundations

## 1. Metadata
- **Phase Identifier:** Phase 4 (`phase4`)
- **Title:** Security Foundations
- **Owner Role:** Security Lead / Backend Engineer
- **Current Status:** `PLANNED`
- **Revision Date:** 2026-10-13

---

## 2. Purpose & Expected Outcome
Phase 4 establishes the authoritative authentication, authorization, session management, and tenant isolation security foundations in the Go backend engines. It eliminates IDOR vulnerabilities, enforces cryptographic host-scoped cookies, implements password hashing (bcrypt >= 12), and establishes negative security test suites.

**Expected Outcome:**
- Server-side session management active in PostgreSQL (`sessions` table).
- Host-scoped cookies configured: `__Host-hackways_session` (Lax) and `__Host-hackways_admin_session` (Strict).
- Role-based authorization middleware enforcing attendee, organizer, and admin boundaries.
- Server-side resource ownership validation preventing IDOR on all mutation endpoints.
- Negative security tests verifying that unauthorized and cross-tenant requests are rejected with 401/403.

---

## 3. Scope & Non-Goals

### In Scope:
- Password hashing with `golang.org/x/crypto/bcrypt` (cost >= 12).
- Stateful session creation, lookup, and revocation in PostgreSQL.
- Auth middleware (`RequireAuth`, `RequireOrganizer`, `RequireAdmin`).
- Cryptographic session cookie transport headers.
- Negative security test suite covering unauthorized access and cross-organizer data access.

### Explicit Non-Goals:
- Building frontend authentication forms (Phase 6).
- Integrating third-party OAuth/OIDC providers (deferred to future milestone).
- Azure Blob SAS upload signing (Phase 5).

---

## 4. Architectural Requirements & Constraints
- Frontend visibility is never a security boundary ([`AGENTS.md` Rule 1.3](../../AGENTS.md#L41-L50)).
- Public & Attendee Sessions: `__Host-hackways_session` (`Secure; HttpOnly; SameSite=Lax; Path=/`).
- Admin Sessions: `__Host-hackways_admin_session` (`Secure; HttpOnly; SameSite=Strict; Path=/`).
- Sensitive secrets and private tokens must never be logged in `slog` output.

---

## 5. Prerequisites & Dependencies
- **Prerequisite Phase:** Phase 3 (Database & SQL Engineering).
- **Prerequisite Decisions:** AD-004 (Host-scoped cookie sessions approved).

---

## 6. Ordered Implementation Steps
1. **Step 1 (Crypto & Password Hashing):** Implement `auth/password.go` using `bcrypt` (cost 12).
2. **Step 2 (Session Management):** Implement `auth/session.go` storing SHA-256 token hashes in PostgreSQL.
3. **Step 3 (Cookie Transport):** Implement cookie writing helpers with `__Host-` prefixes.
4. **Step 4 (Auth Middleware):** Implement `RequireAuth`, `RequireOrganizer`, and `RequireAdmin` middleware.
5. **Step 5 (Negative Security Tests):** Implement test suite in `internal/auth/auth_test.go` verifying 401/403 responses on tampering.

---

## 7. Expected File Changes
- `services/api/internal/auth/password.go` & `password_test.go`
- `services/api/internal/auth/session.go` & `session_test.go`
- `services/api/internal/auth/middleware.go` & `middleware_test.go`
- `services/api/internal/auth/cookie.go`

---

## 8. Verification Commands & Evidence
- `cd services/api && go test ./internal/auth/... -v -race -count=1` -> PASS.
- Negative test output confirming cross-tenant access attempts receive HTTP 403 Forbidden.

---

## 9. Acceptance Criteria & Phase Exit Conditions
- [ ] Bcrypt password hashing verified with cost factor >= 12.
- [ ] Stateful sessions operational with PostgreSQL storage.
- [ ] `__Host-` scoped cookies emitted with Secure, HttpOnly, and SameSite attributes.
- [ ] Ownership checks verified via negative security tests.
