# Database Architecture & Schema — Hackways

**Engine:** PostgreSQL 16  
**Default Schema:** `public`  
**Migration Tool:** `golang-migrate`  
**Query Compiler:** `sqlc` v1.27  

---

## 1. Schema Conventions

- **Primary Keys:** UUID (`gen_random_uuid()`) for all user-facing entities.
- **Timestamps:** UTC `TIMESTAMPTZ NOT NULL DEFAULT NOW()` for `created_at`; `TIMESTAMPTZ NOT NULL` for `updated_at`.
- **Timezone Awareness:** All scheduled times are stored as `TIMESTAMPTZ` alongside `timezone TEXT NOT NULL` for display.
- **Soft Deletion:** `deleted_at TIMESTAMPTZ` on core entities (`users`, `events`).
- **Naming:** `snake_case` column and table names.
- **Constraint Enforcement:** CHECK constraints and partial UNIQUE indexes on active rows.

---

## 2. Migration Sequence

```
services/api/migrations/
├── 000001_extensions.up.sql       # pgcrypto, uuid-ossp
├── 000002_users_sessions.up.sql   # users and sessions tables
├── 000003_organizers.up.sql       # organizers table with user 1:1 FK
├── 000004_categories.up.sql       # event_categories + initial seeds
├── 000005_events.up.sql           # events, event_assets, GIN search index
├── 000006_registrations.up.sql    # registrations table with capacity constraints
├── 000007_notifications.up.sql    # notifications queue table
├── 000008_indexes.up.sql          # Performance & filtering indexes
└── 000009_triggers.up.sql         # set_updated_at trigger functions
```

---

## 3. Core Entity Relations

```
User (1:1) ──── Session (1:N)
  │
  └── Organizer (1:1) ──── Event (1:N) ──── EventCategory (N:1)
                             │
                             ├── EventAsset (1:N)
                             │
                             └── Registration (1:N) ──── Notification (1:N)
```

### Table Definitions Summary
- `users`: id, email (unique active), password_hash, display_name, avatar_url, email_verified, timestamps, deleted_at.
- `sessions`: id, user_id (FK), token_hash (unique), user_agent, ip_address, expires_at, last_seen.
- `organizers`: id, user_id (FK unique), display_name, slug (unique), bio, website_url, verified.
- `event_categories`: id (BIGSERIAL), name, slug (unique), icon.
- `events`: id, organizer_id (FK), category_id (FK), status (draft, published, unpublished, cancelled), format, title, slug (unique), description, location, starts_at, ends_at, timezone, capacity, registration_count, tags.
- `registrations`: id, event_id (FK), user_id (FK nullable), guest_name, guest_email, status, notes, registered_at.
  - Partial unique: `(event_id, user_id)`
  - Partial unique: `(event_id, guest_email)`
- `notifications`: id, registration_id (FK), event_id (FK), recipient_email, type, status, subject, body_html, sent_at.
