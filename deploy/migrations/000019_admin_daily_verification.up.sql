ALTER TABLE platform_admin_sessions ADD COLUMN otp_verified_at timestamptz;
CREATE TABLE admin_session_challenges (
    session_id uuid PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
    code_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    expires_at timestamptz NOT NULL DEFAULT clock_timestamp()+INTERVAL '10 minutes',
    consumed_at timestamptz
);
UPDATE sessions s SET expires_at=GREATEST(s.expires_at,s.created_at+INTERVAL '90 days')
WHERE s.expires_at>clock_timestamp() AND EXISTS(SELECT 1 FROM platform_admin_sessions a WHERE a.session_id=s.id);
