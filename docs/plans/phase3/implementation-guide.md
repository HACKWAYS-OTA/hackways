# Phase 3 Implementation Guide: Database & SQL Engineering

## 1. Prerequisites
- Phase 2 complete (`pgx/v5` in `go.mod`).
- PostgreSQL 16 operational locally on port 5432.
- `sqlc` v1.27+ CLI installed.
- `golang-migrate` CLI installed.

---

## 2. Step-by-Step Implementation

### Step 1: Update Migration DDL
1. Edit `services/api/migrations/000001_initial_schema.up.sql`:
   - Rename table `registrations` to `rsvps`.
   - Update references and unique constraints on `rsvps`.
   - In table `events`, rename `registration_count` to `current_rsvp_count`.
   - Rename table `notifications` to `email_outbox` with columns:
     ```sql
     CREATE TABLE IF NOT EXISTS email_outbox (
         id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
         recipient_email TEXT NOT NULL,
         recipient_name  TEXT,
         template_type   TEXT NOT NULL,
         payload_json    JSONB NOT NULL DEFAULT '{}',
         status          notification_status NOT NULL DEFAULT 'pending',
         retry_count     INTEGER NOT NULL DEFAULT 0,
         scheduled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
         sent_at         TIMESTAMPTZ,
         created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
         updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
     );
     ```
2. Update `000001_initial_schema.down.sql` to drop `email_outbox` and `rsvps`.

### Step 2: Harmonize SQL Queries
1. Edit `services/api/db/queries/rsvps.sql` to reference `rsvps` and `email_outbox`.
2. Implement `LockEventForRSVP`, `CreateRSVP`, `IncrementRSVPCount`, and `EnqueueEmailOutbox`.

### Step 3: Compile Queries with sqlc
1. Run in `services/api`:
   ```bash
   sqlc generate
   ```
2. Verify that `services/api/db/` contains generated Go files.

### Step 4: Initialize pgxpool in Backend
1. Edit `services/api/cmd/api/main.go`:
   - Parse `cfg.DatabaseURL` using `pgxpool.ParseConfig`.
   - Set `poolConfig.MaxConns = 28`.
   - Create pool: `pool, err := pgxpool.NewWithConfig(ctx, poolConfig)`.
   - Defer `pool.Close()`.
2. Pass pool to `platform.HealthHandler`.

---

## 3. What Must NOT Be Changed in Phase 3
- Do not introduce raw SQL string concatenations or runtime ORMs.
- Do not bypass `statement_timeout = '2000ms'` on lock queries.
- Do not exceed 28 connections for the core backend pool.
