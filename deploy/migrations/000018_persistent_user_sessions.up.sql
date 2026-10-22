-- Preserve every still-active ordinary sign-in. Dedicated administrator
-- sessions retain their separate policy. Deleted/expired sessions are not revived.
UPDATE sessions s
SET expires_at = GREATEST(s.expires_at,s.created_at + INTERVAL '90 days')
WHERE s.expires_at > clock_timestamp()
AND NOT EXISTS (SELECT 1 FROM platform_admin_sessions a WHERE a.session_id=s.id);
