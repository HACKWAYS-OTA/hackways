-- Google identity links and one-time cross-origin session handoffs.
-- The existing users and host-only session cookies remain authoritative.
CREATE TABLE IF NOT EXISTS user_identities (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL,
    subject    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_identities_provider_check CHECK (provider IN ('google')),
    CONSTRAINT user_identities_provider_subject_unique UNIQUE (provider, subject),
    CONSTRAINT user_identities_provider_user_unique UNIQUE (provider, user_id)
);

CREATE INDEX IF NOT EXISTS user_identities_user_id_idx ON user_identities (user_id);

CREATE TABLE IF NOT EXISTS oauth_states (
    state_hash  TEXT PRIMARY KEY,
    verifier    TEXT NOT NULL,
    return_to   TEXT NOT NULL,
    nonce_hash  TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS oauth_states_expiry_idx
    ON oauth_states (expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE IF NOT EXISTS oauth_exchange_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nonce_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at   TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS oauth_exchange_tokens_expiry_idx
    ON oauth_exchange_tokens (expires_at)
    WHERE used_at IS NULL;
