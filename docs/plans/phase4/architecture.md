# Phase 4 Technical Architecture: Authentication & Session Security

## 1. Authentication & Session Flow

```mermaid
sequenceDiagram
    autonumber
    actor Client as User / Browser
    participant API as Global Backend (:8082)
    participant PG as PostgreSQL (sessions)

    Client->>API: POST /api/v1/auth/login {email, password}
    Note over API: Lookup user by email<br/>Verify bcrypt hash (cost >= 12)
    API->>API: Generate 32-byte crypto random token<br/>Compute SHA-256(token)
    API->>PG: INSERT INTO sessions (user_id, token_hash, expires_at)<br/>VALUES ($1, $2, NOW() + INTERVAL '30 days')
    API-->>Client: 200 OK + Set-Cookie: __Host-hackways_session=<token>;<br/>HttpOnly; Secure; SameSite=Lax; Path=/

    Note over Client,API: Subsequent Request to Protected Route
    Client->>API: GET /api/v1/organizers/me/events (Cookie: __Host-hackways_session)
    API->>API: Compute SHA-256(token from cookie)
    API->>PG: SELECT user_id, expires_at FROM sessions<br/>WHERE token_hash = $1 AND expires_at > NOW()
    Note over API: Inject User & Roles into request context
    API-->>Client: 200 OK {data: [...]}
```

---

## 2. Cookie Scoping & Cross-Subdomain Hijacking Prevention
- Setting the `__Host-` cookie prefix enforces three browser-level invariants:
  1. The cookie can only be set from a secure (HTTPS) origin.
  2. The cookie cannot specify a `Domain` attribute (it is strictly bound to the specific host, e.g., `app.hackways.in`).
  3. The `Path` attribute must be `/`.
- This ensures an attacker compromising a marketing subdomain cannot inject or hijack session cookies for `app.hackways.in` or `admin.hackways.in`.

---

## 3. Server-Side IDOR Mitigations
- Insecure Direct Object References (IDOR) are prevented by embedding ownership constraints into all SQL queries.
- When an organizer requests an event update:
  ```sql
  UPDATE events
  SET title = $3, description = $4, updated_at = NOW()
  WHERE id = $1 AND organizer_id = $2 AND deleted_at IS NULL;
  ```
- If the authenticated user is not the owning organizer, zero rows are affected and the backend returns `404 Not Found` or `403 Forbidden`.
