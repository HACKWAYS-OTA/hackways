# Phase 4 Implementation Guide: Security Foundations

## 1. Prerequisites
- Phase 3 complete (PostgreSQL `sessions` table and `sqlc` models compiled).
- Go crypto library available: `golang.org/x/crypto/bcrypt`.

---

## 2. Step-by-Step Implementation

### Step 1: Password Hashing (`internal/auth/password.go`)
1. Create `services/api/internal/auth/password.go`:
   ```go
   package auth

   import "golang.org/x/crypto/bcrypt"

   const BcryptCost = 12

   func HashPassword(password string) (string, error) {
       bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
       return string(bytes), err
   }

   func CheckPasswordHash(password, hash string) bool {
       err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
       return err == nil
   }
   ```
2. Write unit tests in `password_test.go`.

### Step 2: Session Store & Token Generation (`internal/auth/session.go`)
1. Generate cryptographically secure 32-byte tokens via `crypto/rand`.
2. Compute `sha256.Sum256(token)` to store `token_hash` in the database.
3. Implement `CreateSession`, `GetSessionByToken`, and `RevokeSession`.

### Step 3: Cookie Transport Helpers (`internal/auth/cookie.go`)
1. Implement `SetSessionCookie(w http.ResponseWriter, token string, isProduction bool)`.
2. Name cookie `__Host-hackways_session` (or `hackways_session` in local development when HTTP).
3. Set `HttpOnly=true`, `SameSite=Lax`, `Path=/`.

### Step 4: Authentication Middleware (`internal/auth/middleware.go`)
1. Implement `RequireAuth(sessionStore *SessionStore) func(http.Handler) http.Handler`.
2. Read cookie, validate token against store, attach user context.
3. Implement `RequireOrganizer` and `RequireAdmin` wrappers.

### Step 5: Negative Security Tests (`internal/auth/auth_test.go`)
1. Test missing cookie -> 401 Unauthorized.
2. Test invalid token -> 401 Unauthorized.
3. Test expired session -> 401 Unauthorized.
4. Test regular user accessing admin route -> 403 Forbidden.

---

## 3. What Must NOT Be Changed in Phase 4
- Never log plain tokens, session hashes, or raw passwords in slog.
- Never use JWTs in place of stateful database sessions (AD-004).
- Never allow client-side role claims to bypass backend verification.
