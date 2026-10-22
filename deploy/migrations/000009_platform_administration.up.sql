-- Additive admin access records. No accounts or privileges are seeded.
CREATE TABLE platform_administrators (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
    granted_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);
CREATE TABLE platform_admin_sessions (
    session_id UUID PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE platform_administrators IS 'Explicit platform access; verified email and a live admin session are also required.';
