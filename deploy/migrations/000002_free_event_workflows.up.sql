-- Additive free-event workflow foundation. Apply before deploying either Go engine.
ALTER TYPE registration_status ADD VALUE IF NOT EXISTS 'pending_approval';
ALTER TYPE registration_status ADD VALUE IF NOT EXISTS 'offered';
ALTER TYPE registration_status ADD VALUE IF NOT EXISTS 'rejected';
ALTER TYPE registration_status ADD VALUE IF NOT EXISTS 'expired';
ALTER TYPE registration_status ADD VALUE IF NOT EXISTS 'transferred';

ALTER TABLE events
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1,
    ADD COLUMN held_rsvp_count INTEGER NOT NULL DEFAULT 0 CHECK (held_rsvp_count >= 0),
    ADD COLUMN registration_mode TEXT NOT NULL DEFAULT 'instant' CHECK (registration_mode IN ('instant', 'approval')),
    ADD COLUMN allow_guest BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN waitlist_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN offer_minutes INTEGER NOT NULL DEFAULT 10 CHECK (offer_minutes BETWEEN 5 AND 60);
ALTER TABLE events ADD CONSTRAINT events_inventory_within_capacity
    CHECK (capacity IS NULL OR current_rsvp_count + held_rsvp_count <= capacity) NOT VALID;

ALTER TABLE rsvps ADD COLUMN offer_expires_at TIMESTAMPTZ,
    ADD COLUMN approval_expires_at TIMESTAMPTZ,
    ADD COLUMN transfer_count INTEGER NOT NULL DEFAULT 0 CHECK (transfer_count BETWEEN 0 AND 1),
    ADD COLUMN transferred_from UUID REFERENCES rsvps(id) ON DELETE RESTRICT;
-- Fail for operator reconciliation if legacy case-variant identities collide.
-- Never delete or silently merge existing registrations during migration.
CREATE UNIQUE INDEX rsvps_guest_email_normalized_unique ON rsvps(event_id, lower(guest_email)) WHERE guest_email IS NOT NULL;
CREATE INDEX rsvps_waitlist_order_idx ON rsvps(event_id, registered_at, id) WHERE status = 'waitlisted';
CREATE INDEX rsvps_offer_expiry_idx ON rsvps(offer_expires_at, event_id) WHERE offer_expires_at IS NOT NULL;
CREATE INDEX rsvps_approval_expiry_idx ON rsvps(approval_expires_at, event_id) WHERE approval_expires_at IS NOT NULL;

CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rsvp_id UUID NOT NULL UNIQUE REFERENCES rsvps(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'scanned', 'revoked')),
    checked_in_at TIMESTAMPTZ,
    checked_in_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    entity_id UUID NOT NULL,
    action TEXT NOT NULL,
    request_id TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX audit_log_event_idx ON audit_log(event_id, id DESC);

-- Existing confirmed registrations also receive real ticket records.
INSERT INTO tickets(rsvp_id) SELECT id FROM rsvps WHERE status='confirmed' ON CONFLICT(rsvp_id) DO NOTHING;

CREATE TABLE idempotency_records (
    scope_hash TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    response_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    PRIMARY KEY (scope_hash, key_hash)
);
CREATE INDEX idempotency_expiry_idx ON idempotency_records(expires_at);

CREATE TABLE auth_rate_limits (
    scope_hash TEXT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL,
    PRIMARY KEY (scope_hash, window_start)
);
CREATE INDEX auth_rate_limits_expiry_idx ON auth_rate_limits(window_start);

ALTER TABLE email_outbox ADD COLUMN lease_token UUID,
    ADD COLUMN first_sent_at TIMESTAMPTZ;

CREATE TABLE email_verifications (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE audit_log IS 'Append-only through application queries; no secrets or form answers in metadata.';
