# Phase 4 Agent Pre-Context: Security Foundations

## 1. Product Objective & Technical Target
Phase 4 implements authentication, stateful sessions, role-based authorization, and tenant isolation in the Go backend. It enforces the principle that **the frontend is never a security boundary**.

---

## 2. Canonical Specifications to Obey
1. **[`AGENTS.md` Rule 1.3](../../../AGENTS.md#L41-L50):**
   - Enforce authentication, authorization, role checks, resource ownership, and input validation on the backend.
   - Prevent cross-user and cross-organizer data access (IDOR).
   - Never expose secrets, session credentials, internal service ports, or privileged operations to unauthorized clients.
2. **[`AGENTS.md` Rule 2.6](../../../AGENTS.md#L224-L232):**
   - Cryptographic cookie scoping:
     - Public & Attendee: `__Host-hackways_session` (`Secure; HttpOnly; SameSite=Lax; Path=/`).
     - Admin: `__Host-hackways_admin_session` (`Secure; HttpOnly; SameSite=Strict; Path=/`).
3. **[`docs/SECURITY.md`](../../SECURITY.md):**
   - Password hashing with bcrypt cost >= 12.
   - Sessions stored in PostgreSQL; tokens hashed with SHA-256 (`token_hash`).

---

## 3. Verified Repository Facts Before Phase 4
- `services/api/internal/auth/` currently contains only an empty `.gitkeep`.
- PostgreSQL `sessions` table is defined in the initial schema migration.
- `services/api/internal/platform/middleware.go` contains CORS, RequestID, and Recoverer, but lacks auth middleware.

---

## 4. Expected Outputs
- `internal/auth` package with password hashing, session management, and RBAC guards.
- Negative security test suite passing with 100% assertions met.
- Zero credential or token leakage in logs.
