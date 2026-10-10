# Security Architecture & Threat Model — Hackways

**Status:** Approved  
**Classification:** Core Security Reference  

---

## 1. Threat Model & Mitigations

### 1.1 Insecure Direct Object Reference (IDOR)
- **Risk:** Malicious organizers editing events they do not own or unauthorized attendees reading attendee rosters.
- **Mitigation:** Server-validated ownership queries (`WHERE id = $1 AND organizer_id = $2`). The frontend NEVER dictates authorization state.

### 1.2 RSVP Overbooking & Concurrency Race Conditions
- **Risk:** High-demand events having multiple attendees register for the last ticket simultaneously, exceeding capacity.
- **Mitigation:** Row-level locking via PostgreSQL `SELECT ... FOR UPDATE` inside a database transaction, combined with database unique constraints on `(event_id, email)`.

### 1.3 Credential Theft & Brute Force
- **Risk:** Password cracking, credential stuffing, session fixation.
- **Mitigation:**
  - Password hashing with bcrypt (cost factor >= 12).
  - Sessions stored in PostgreSQL; tokens hashed with SHA-256 (`token_hash`).
  - Session cookies marked `HttpOnly`, `SameSite=Lax`, and `Secure` (production).
  - Rate limiting on `/api/v1/auth/login` (5/min per IP) and `/api/v1/auth/signup` (3/min per IP).

### 1.4 Arbitrary File Uploads (Banner Images)
- **Risk:** Malicious file uploads (e.g. PHP/executables disguised as images).
- **Mitigation:**
  - File size capped at 5 MB.
  - Server-side magic-byte inspection (detecting true MIME types: image/jpeg, image/png, image/webp).
  - Storage paths randomized with UUIDs.
  - Files served with `Content-Disposition: inline` and strict `Content-Type`.

### 1.5 SQL Injection
- **Risk:** Arbitrary SQL execution via search or form inputs.
- **Mitigation:** Compile-time parameterized SQL queries generated via `sqlc`. Zero string interpolation or `fmt.Sprintf` in SQL queries.

---

## 2. Secrets Management
- No secrets in git repositories.
- `.env` strictly gitignored; `.env.example` provides safe placeholders.
- Production secrets injected via environment variables in runtime container configurations.
