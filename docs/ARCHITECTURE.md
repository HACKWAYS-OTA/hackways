# System Architecture — Hackways

**Version:** 1.0  
**Status:** Approved  
**Classification:** Modular Monolith  

---

## 1. Executive Summary

Hackways is an event discovery, event management, and RSVP registration platform designed for high concurrency, strict consistency, and low operational overhead. It couples a Go modular monolith (`services/api`) with a Next.js 14 App Router frontend (`apps/web`) backed by a single PostgreSQL 16 database.

---

## 2. System Context Diagram

```text
┌────────────────────────────────────────────────────────────────────────┐
│                              Internet / CDN                            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ HTTPS
                        ┌───────────▼───────────┐
                        │    Next.js Web App    │  apps/web/
                        │   (App Router, SSR)   │  Port 3000
                        └───────────┬───────────┘
                                    │ REST /api/v1 (JSON)
                        ┌───────────▼───────────┐
                        │     Go API Server     │  services/api/
                        │  (Chi router, slog)   │  Port 8080
                        └─────┬───────────┬─────┘
                              │           │
                     pgx pool │           │ storage.Store
                              ▼           ▼
                      ┌──────────────┐  ┌───────────────────┐
                      │PostgreSQL 16 │  │  Object Storage   │
                      │(Single schema│  │ (Local FS / S3)   │
                      └──────────────┘  └───────────────────┘
```

---

## 3. Technology Rationale

| Concern | Selection | Rationale |
|---|---|---|
| **Backend Language** | Go 1.23 | Fast startup, small memory footprint, strong stdlib, native goroutines for concurrency |
| **HTTP Routing** | Chi v5 | 100% compatible with `net/http`, minimal allocations, composable middleware |
| **Database** | PostgreSQL 16 | ACID transactions, row-level locking (`SELECT FOR UPDATE`), GIN full-text search |
| **Driver & Pooling** | `pgx/v5` | High performance binary protocol, connection pooling, prepared statements |
| **Query Layer** | `sqlc` v1.27 | Compile-time SQL checking, zero runtime reflection, clean Go model generation |
| **Migrations** | `golang-migrate` | Reversible, numbered migrations (`.up.sql` / `.down.sql`) |
| **Frontend Framework** | Next.js 14 | React Server Components for SEO and low client bundle size |
| **Styling** | Tailwind CSS | Utility tokens, strict consistency, purge-clean production builds |
| **Sessions** | PostgreSQL Sessions | Revocable, stateful, opaque cookie tokens; no JWT expiration or revocation traps |

---

## 4. Domain Decomposition

All domain packages live in `services/api/internal/`:

```
services/api/internal/
├── platform/       # Cross-cutting: config, health, middleware, error envelopes
├── auth/           # Sessions, password hashing (bcrypt >= 12), auth middleware
├── users/          # User accounts, profiles, display identity
├── organizers/     # Organizer profiles, slugs, organization management
├── events/         # Event aggregate, state machine, categorization, full-text search
├── rsvps/          # Attendee registrations, transactional capacity checking
├── notifications/  # Transactional notification queue and email dispatch
└── storage/        # File upload validation (magic bytes) and storage adapters
```

---

## 5. Concurrency Guarantees (RSVP Overbooking Prevention)

The most critical reliability requirement is preventing duplicate registrations and capacity overbooking under high concurrency.

```sql
BEGIN;

-- 1. Acquire pessimistic row lock on event to serialize concurrent reservations
SELECT id, capacity, registration_count, status
FROM events
WHERE id = $1 AND status = 'published' AND deleted_at IS NULL
FOR UPDATE;

-- 2. Verify registration window is open and capacity is not exceeded
-- (Application check: registration_count < capacity)

-- 3. Insert registration (Database UNIQUE constraints prevent duplicate registrations)
INSERT INTO registrations (event_id, user_id, guest_email, guest_name)
VALUES ($1, $2, $3, $4);

-- 4. Atomically increment registration count
UPDATE events
SET registration_count = registration_count + 1
WHERE id = $1;

COMMIT;
```

---

## 6. Security Architecture

1. **Authentication:** Server-side sessions stored in PostgreSQL `sessions` table. Client receives an opaque session token stored in an `HttpOnly`, `SameSite=Lax`, `Secure` (production) cookie.
2. **Password Security:** Hashes stored using `bcrypt` with cost factor >= 12.
3. **Authorization:** Server-side ownership validation on every mutation. Frontend checks are cosmetic only.
4. **Data Isolation:** Soft deletion (`deleted_at`) prevents accidental data loss; active unique indexes ignore soft-deleted records.
