-- 000001_initial_schema.up.sql
-- Hackways Core Schema Definition

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enums
CREATE TYPE event_status AS ENUM ('draft', 'published', 'unpublished', 'cancelled');
CREATE TYPE event_format AS ENUM ('in_person', 'online', 'hybrid');
CREATE TYPE registration_status AS ENUM ('confirmed', 'cancelled', 'waitlisted');
CREATE TYPE notification_type AS ENUM ('rsvp_confirmation', 'rsvp_cancellation', 'event_updated', 'event_cancelled');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed');

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL,
    avatar_url      TEXT,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT users_display_name_len CHECK (char_length(display_name) BETWEEN 1 AND 100)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_active_idx
    ON users (lower(email))
    WHERE deleted_at IS NULL;

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    user_agent  TEXT,
    ip_address  INET,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);

-- Organizers table
CREATE TABLE IF NOT EXISTS organizers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    display_name    TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    bio             TEXT,
    website_url     TEXT,
    avatar_url      TEXT,
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT organizers_display_name_len CHECK (char_length(display_name) BETWEEN 1 AND 150),
    CONSTRAINT organizers_slug_format CHECK (slug ~* '^[a-z0-9-]+$')
);

-- Event Categories table
CREATE TABLE IF NOT EXISTS event_categories (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT,
    icon        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id            UUID NOT NULL REFERENCES organizers(id) ON DELETE RESTRICT,
    category_id             BIGINT REFERENCES event_categories(id) ON DELETE SET NULL,
    status                  event_status NOT NULL DEFAULT 'draft',
    format                  event_format NOT NULL DEFAULT 'in_person',
    title                   TEXT NOT NULL,
    slug                    TEXT NOT NULL UNIQUE,
    description             TEXT NOT NULL,
    short_description       TEXT,
    banner_url              TEXT,
    location_name           TEXT,
    location_address        TEXT,
    online_url              TEXT,
    starts_at               TIMESTAMPTZ NOT NULL,
    ends_at                 TIMESTAMPTZ NOT NULL,
    timezone                TEXT NOT NULL,
    capacity                INTEGER,
    current_rsvp_count      INTEGER NOT NULL DEFAULT 0,
    registration_opens_at   TIMESTAMPTZ,
    registration_closes_at  TIMESTAMPTZ,
    tags                    TEXT[] NOT NULL DEFAULT '{}',
    published_at            TIMESTAMPTZ,
    cancelled_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    CONSTRAINT events_title_len CHECK (char_length(title) BETWEEN 1 AND 200),
    CONSTRAINT events_ends_after_starts CHECK (ends_at > starts_at),
    CONSTRAINT events_capacity_positive CHECK (capacity IS NULL OR capacity > 0),
    CONSTRAINT events_current_rsvp_count_nonneg CHECK (current_rsvp_count >= 0)
);

CREATE INDEX IF NOT EXISTS events_status_idx ON events (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS events_starts_at_idx ON events (starts_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS events_organizer_idx ON events (organizer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS events_search_idx ON events
    USING gin(to_tsvector('english', title || ' ' || description))
    WHERE deleted_at IS NULL AND status = 'published';

-- RSVPs table
CREATE TABLE IF NOT EXISTS rsvps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    guest_name      TEXT,
    guest_email     TEXT,
    status          registration_status NOT NULL DEFAULT 'confirmed',
    notes           TEXT,
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rsvps_identity_req CHECK (user_id IS NOT NULL OR guest_email IS NOT NULL),
    CONSTRAINT rsvps_user_event_unique UNIQUE (event_id, user_id),
    CONSTRAINT rsvps_guest_event_unique UNIQUE (event_id, guest_email)
);

CREATE INDEX IF NOT EXISTS rsvps_event_idx ON rsvps (event_id);
CREATE INDEX IF NOT EXISTS rsvps_event_status_idx ON rsvps (event_id, status);
CREATE INDEX IF NOT EXISTS rsvps_user_idx ON rsvps (user_id);

-- Email Outbox table (Transactional Outbox Pattern)
CREATE TABLE IF NOT EXISTS email_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rsvp_id         UUID REFERENCES rsvps(id) ON DELETE SET NULL,
    event_id        UUID REFERENCES events(id) ON DELETE SET NULL,
    recipient_email TEXT NOT NULL,
    recipient_name  TEXT,
    template_type   TEXT NOT NULL,
    payload_json    JSONB NOT NULL DEFAULT '{}',
    status          notification_status NOT NULL DEFAULT 'pending',
    retry_count     INTEGER NOT NULL DEFAULT 0,
    scheduled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS email_outbox_pending_idx ON email_outbox (status, scheduled_at) WHERE status = 'pending';

-- Seed initial categories
INSERT INTO event_categories (name, slug, description, icon) VALUES
    ('Technology', 'technology', 'Hackathons, developer conferences, and coding meetups', '💻'),
    ('Business', 'business', 'Networking, pitch competitions, and startup events', '💼'),
    ('Design', 'design', 'UI/UX workshops, product design sessions', '🎨'),
    ('Community', 'community', 'Local tech gatherings and social meetups', '🌐')
ON CONFLICT (slug) DO NOTHING;
