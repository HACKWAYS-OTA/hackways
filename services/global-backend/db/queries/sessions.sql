-- name: CreateSession :one
INSERT INTO sessions (
    id,
    user_id,
    token_hash,
    user_agent,
    ip_address,
    created_at,
    expires_at,
    last_seen
) VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    NOW(),
    $5,
    NOW()
)
RETURNING id, user_id, token_hash, user_agent, ip_address, created_at, expires_at, last_seen, admin_reference, device_token_hash, revoked_at;

-- name: GetSessionByTokenHash :one
SELECT s.id, s.user_id, s.token_hash, s.user_agent, s.ip_address, s.created_at, s.expires_at, s.last_seen,
       u.email, u.display_name, u.avatar_url, u.email_verified,s.revoked_at,
       (EXISTS(SELECT 1 FROM account_restrictions a WHERE a.user_id=u.id AND a.blocked)
        OR EXISTS(SELECT 1 FROM device_restrictions d WHERE d.token_hash=s.device_token_hash AND d.blocked))::boolean AS restricted
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token_hash = $1
  AND s.expires_at > NOW()
  AND u.deleted_at IS NULL;

-- name: UpdateSessionLastSeen :exec
UPDATE sessions
SET last_seen = NOW()
WHERE id = $1 AND last_seen < NOW() - INTERVAL '5 minutes';

-- name: DeleteSessionByTokenHash :exec
DELETE FROM sessions
WHERE token_hash = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: UpdateSessionActivityBatch :exec
UPDATE sessions SET last_seen=clock_timestamp()
WHERE id=ANY(sqlc.arg(ids)::uuid[]) AND last_seen<clock_timestamp()-INTERVAL '5 minutes'
AND expires_at>clock_timestamp() AND revoked_at IS NULL;
