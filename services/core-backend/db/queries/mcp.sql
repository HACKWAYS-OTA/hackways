-- name: MCPAccessIdentity :one
SELECT g.id AS grant_id,g.client_id,g.resource,g.scopes,g.session_id,u.id AS user_id,u.email,u.display_name,u.avatar_url,u.email_verified,
 LEAST(t.expires_at,g.expires_at,s.expires_at)::timestamptz AS expires_at
FROM mcp_access_tokens t JOIN mcp_oauth_grants g ON g.id=t.grant_id JOIN sessions s ON s.id=g.session_id AND s.user_id=g.user_id
JOIN users u ON u.id=g.user_id
WHERE t.token_hash=$1 AND g.revoked_at IS NULL AND t.expires_at>clock_timestamp() AND g.expires_at>clock_timestamp()
AND s.expires_at>clock_timestamp() AND u.deleted_at IS NULL AND u.email_verified;

-- name: MCPExecutionAllowed :one
SELECT EXISTS(SELECT 1 FROM mcp_action_intents a JOIN mcp_oauth_grants g ON g.id=a.grant_id
WHERE a.id=sqlc.arg(intent_id) AND a.grant_id=sqlc.arg(grant_id) AND a.request_hash=sqlc.arg(request_hash)
AND a.api_service=sqlc.arg(api_service) AND a.method=sqlc.arg(method) AND a.api_path=sqlc.arg(api_path)
AND a.status='executing' AND a.approved_at IS NOT NULL AND a.approved_session_id IS NOT NULL
AND EXISTS(SELECT 1 FROM sessions approved WHERE approved.id=a.approved_session_id AND approved.user_id=g.user_id AND approved.expires_at>clock_timestamp())
AND a.lease_hash=sqlc.arg(lease_hash) AND a.lease_expires_at>clock_timestamp() AND a.expires_at>clock_timestamp()
AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp());
