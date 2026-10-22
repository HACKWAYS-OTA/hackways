ALTER TABLE users ADD COLUMN IF NOT EXISTS device_baseline_at TIMESTAMPTZ;
ALTER TABLE email_verifications ADD COLUMN IF NOT EXISTS attempts SMALLINT NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5);

CREATE TABLE password_reset_challenges (
 user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 token_hash TEXT NOT NULL UNIQUE,
 expires_at TIMESTAMPTZ NOT NULL,
 consumed_at TIMESTAMPTZ,
 attempts SMALLINT NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE TABLE recognized_devices (
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash TEXT NOT NULL CHECK(length(token_hash)=64),
 user_agent TEXT NOT NULL CHECK(length(user_agent)<=512),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 last_seen_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 expires_at TIMESTAMPTZ NOT NULL,
 PRIMARY KEY(user_id,token_hash)
);
CREATE INDEX recognized_devices_expiry_idx ON recognized_devices(expires_at);
COMMENT ON TABLE recognized_devices IS 'Browser recognition for notification deduplication only; never an authentication factor.';
