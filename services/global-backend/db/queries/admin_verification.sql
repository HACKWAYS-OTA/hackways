-- name: CreateUnverifiedAdminSession :exec
INSERT INTO platform_admin_sessions(session_id) VALUES($1);

-- name: LockAdminVerificationSession :one
SELECT s.id,a.otp_verified_at,clock_timestamp()::timestamptz AS database_now
FROM sessions s JOIN platform_admin_sessions a ON a.session_id=s.id
WHERE s.id=sqlc.arg(session_id)::uuid AND s.user_id=sqlc.arg(user_id)::uuid AND s.expires_at>clock_timestamp()
FOR UPDATE OF s,a;

-- name: ReadAdminSessionChallenge :one
SELECT * FROM admin_session_challenges WHERE session_id=$1 FOR UPDATE;

-- name: CreateAdminSessionChallenge :exec
INSERT INTO admin_session_challenges(session_id,code_hash) VALUES($1,$2)
ON CONFLICT(session_id) DO UPDATE SET code_hash=EXCLUDED.code_hash,created_at=clock_timestamp(),expires_at=clock_timestamp()+INTERVAL '10 minutes',consumed_at=NULL;

-- name: MarkAdminSessionVerified :exec
UPDATE platform_admin_sessions SET otp_verified_at=clock_timestamp() WHERE session_id=$1;

-- name: ConsumeAdminSessionChallenge :exec
UPDATE admin_session_challenges SET consumed_at=clock_timestamp() WHERE session_id=$1;

-- name: AdminSessionCodeActive :one
SELECT EXISTS(SELECT 1 FROM admin_session_challenges c JOIN sessions s ON s.id=c.session_id
WHERE c.code_hash=$1 AND c.consumed_at IS NULL AND c.expires_at>clock_timestamp() AND s.expires_at>clock_timestamp());

-- name: AdminSessionVerificationDeadline :one
SELECT (otp_verified_at+INTERVAL '24 hours')::timestamptz FROM platform_admin_sessions WHERE session_id=$1;
