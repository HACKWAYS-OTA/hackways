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

-- name: MCPGetClient :one
SELECT * FROM mcp_oauth_clients WHERE id=$1;

-- name: MCPRegisterClient :one
INSERT INTO mcp_oauth_clients(id,name,redirect_uris,metadata_url,verified)
SELECT $1,$2,$3,$4,$5 WHERE (SELECT count(*) FROM mcp_oauth_clients)<2000
ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,redirect_uris=EXCLUDED.redirect_uris,metadata_url=EXCLUDED.metadata_url,verified=EXCLUDED.verified,updated_at=clock_timestamp()
RETURNING *;

-- name: MCPCreateAuthorizationRequest :one
INSERT INTO mcp_oauth_requests(client_id,redirect_uri,state,resource,code_challenge,scopes) SELECT $1,$2,$3,$4,$5,$6 WHERE (SELECT count(*) FROM mcp_oauth_requests)<10000 RETURNING *;

-- name: MCPAuthorizationRequest :one
SELECT r.*,c.name AS client_name,c.verified AS client_verified FROM mcp_oauth_requests r JOIN mcp_oauth_clients c ON c.id=r.client_id
WHERE r.id=$1 AND r.expires_at>clock_timestamp();

-- name: MCPLockAuthorizationRequest :one
SELECT * FROM mcp_oauth_requests WHERE id=$1 AND expires_at>clock_timestamp() FOR UPDATE;

-- name: MCPApproveAuthorizationRequest :one
UPDATE mcp_oauth_requests SET status='approved',user_id=$2,session_id=$3,code_hash=$4,scopes=$5
WHERE id=$1 AND status='pending' AND expires_at>clock_timestamp() RETURNING *;

-- name: MCPDenyAuthorizationRequest :execrows
UPDATE mcp_oauth_requests SET status='denied' WHERE id=$1 AND status='pending' AND expires_at>clock_timestamp();

-- name: MCPLockAuthorizationCode :one
SELECT r.*,s.expires_at AS session_expires_at FROM mcp_oauth_requests r JOIN sessions s ON s.id=r.session_id AND s.user_id=r.user_id
JOIN users u ON u.id=r.user_id WHERE r.code_hash=$1 AND r.status='approved' AND r.expires_at>clock_timestamp()
AND s.expires_at>clock_timestamp() AND u.deleted_at IS NULL AND u.email_verified FOR UPDATE OF r;

-- name: MCPConsumeAuthorizationCode :exec
UPDATE mcp_oauth_requests SET code_used_at=clock_timestamp() WHERE id=$1;

-- name: MCPCreateGrant :one
INSERT INTO mcp_oauth_grants(user_id,session_id,client_id,resource,scopes,expires_at)
SELECT s.user_id,s.id,$2,$3,$4,LEAST(s.expires_at,clock_timestamp()+INTERVAL '90 days') FROM sessions s
JOIN users u ON u.id=s.user_id WHERE s.id=$1 AND s.expires_at>clock_timestamp() AND u.deleted_at IS NULL
AND (SELECT count(*) FROM mcp_oauth_grants g WHERE g.user_id=s.user_id AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp())<50
RETURNING *;

-- name: MCPInsertAccessToken :one
INSERT INTO mcp_access_tokens(token_hash,grant_id,expires_at) SELECT $1,g.id,LEAST(g.expires_at,clock_timestamp()+INTERVAL '15 minutes')
FROM mcp_oauth_grants g WHERE g.id=$2 AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp()
AND (SELECT count(*) FROM mcp_access_tokens a WHERE a.grant_id=g.id AND a.expires_at>clock_timestamp())<4 RETURNING *;

-- name: MCPInsertRefreshToken :exec
INSERT INTO mcp_refresh_tokens(token_hash,grant_id,expires_at) SELECT $1,g.id,g.expires_at FROM mcp_oauth_grants g
WHERE g.id=$2 AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp();

-- name: MCPLockRefreshToken :one
SELECT t.token_hash,t.grant_id,t.expires_at,t.used_at,g.user_id,g.session_id,g.client_id,g.resource,g.scopes,g.revoked_at
FROM mcp_refresh_tokens t JOIN mcp_oauth_grants g ON g.id=t.grant_id JOIN sessions s ON s.id=g.session_id AND s.user_id=g.user_id
JOIN users u ON u.id=g.user_id WHERE t.token_hash=$1 AND s.expires_at>clock_timestamp() AND g.expires_at>clock_timestamp() AND u.deleted_at IS NULL AND u.email_verified
FOR UPDATE OF t,g;

-- name: MCPConsumeRefreshToken :exec
UPDATE mcp_refresh_tokens SET used_at=clock_timestamp() WHERE token_hash=$1;

-- name: MCPRevokeGrant :execrows
UPDATE mcp_oauth_grants SET revoked_at=COALESCE(revoked_at,clock_timestamp()),revocation_reason=$2 WHERE id=$1;

-- name: MCPGrantForToken :one
SELECT g.id,g.client_id FROM mcp_oauth_grants g WHERE g.id=(SELECT r.grant_id FROM mcp_refresh_tokens r WHERE r.token_hash=$1 UNION SELECT t.grant_id FROM mcp_access_tokens t WHERE t.token_hash=$1 LIMIT 1);

-- name: MCPListConnections :many
SELECT g.id,c.name AS client_name,c.verified AS client_verified,g.scopes,g.created_at,g.expires_at,g.revoked_at,
 (g.scopes && ARRAY['admin:read','admin:write']::text[])::boolean AS admin_connection
FROM mcp_oauth_grants g JOIN mcp_oauth_clients c ON c.id=g.client_id WHERE g.user_id=$1 AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp()
ORDER BY g.created_at DESC,g.id LIMIT 50;

-- name: MCPRevokeOwnedGrant :execrows
UPDATE mcp_oauth_grants SET revoked_at=COALESCE(revoked_at,clock_timestamp()),revocation_reason='Account disconnect'
WHERE id=$1 AND user_id=$2;

-- name: MCPLockGrant :one
SELECT * FROM mcp_oauth_grants WHERE id=$1 AND revoked_at IS NULL AND expires_at>clock_timestamp() FOR UPDATE;

-- name: MCPCreateActionIntent :one
INSERT INTO mcp_action_intents(id,grant_id,tool_name,api_service,method,api_path,payload_json,request_hash)
SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE (SELECT count(*) FROM mcp_action_intents WHERE grant_id=$2 AND status IN ('pending','approved','executing') AND expires_at>clock_timestamp())<10
RETURNING *;

-- name: MCPReadActionIntent :one
SELECT a.*,g.user_id,g.scopes,c.name AS client_name FROM mcp_action_intents a JOIN mcp_oauth_grants g ON g.id=a.grant_id JOIN mcp_oauth_clients c ON c.id=g.client_id
WHERE a.id=$1 AND a.expires_at>clock_timestamp() AND g.revoked_at IS NULL AND g.expires_at>clock_timestamp();

-- name: MCPApproveActionIntent :execrows
UPDATE mcp_action_intents a SET status='approved',approved_at=clock_timestamp(),approved_session_id=$3
WHERE a.id=$1 AND a.grant_id IN(SELECT g.id FROM mcp_oauth_grants g WHERE g.user_id=$2 AND g.revoked_at IS NULL)
AND status='pending' AND expires_at>clock_timestamp();

-- name: MCPDenyActionIntent :execrows
UPDATE mcp_action_intents a SET status='denied' WHERE a.id=$1 AND a.grant_id IN(SELECT g.id FROM mcp_oauth_grants g WHERE g.user_id=$2)
AND status='pending' AND expires_at>clock_timestamp();

-- name: MCPClaimActionIntent :one
UPDATE mcp_action_intents SET status='executing',lease_hash=$3,lease_expires_at=clock_timestamp()+INTERVAL '30 seconds'
WHERE id=$1 AND grant_id=$2 AND expires_at>clock_timestamp() AND approved_at IS NOT NULL
AND status='approved' RETURNING *;

-- name: MCPFinishActionIntent :execrows
UPDATE mcp_action_intents SET status=$3,response_status=$4,result_json=$5,finished_at=clock_timestamp(),lease_expires_at=NULL
WHERE id=$1 AND lease_hash=$2 AND status='executing';

-- name: MCPCleanupRequests :exec
DELETE FROM mcp_oauth_requests WHERE id IN(SELECT id FROM mcp_oauth_requests WHERE expires_at<clock_timestamp()-INTERVAL '30 minutes' ORDER BY expires_at LIMIT 500);

-- name: MCPCleanupActions :exec
DELETE FROM mcp_action_intents WHERE id IN(SELECT id FROM mcp_action_intents WHERE expires_at<clock_timestamp()-INTERVAL '1 day' ORDER BY expires_at LIMIT 500);

-- name: MCPCleanupAccessTokens :exec
DELETE FROM mcp_access_tokens WHERE token_hash IN(SELECT token_hash FROM mcp_access_tokens WHERE expires_at<clock_timestamp()-INTERVAL '1 day' ORDER BY expires_at LIMIT 500);

-- name: MCPCleanupGrants :exec
DELETE FROM mcp_oauth_grants WHERE id IN(SELECT id FROM mcp_oauth_grants WHERE expires_at<clock_timestamp()-INTERVAL '1 day' OR revoked_at<clock_timestamp()-INTERVAL '1 day' ORDER BY expires_at LIMIT 100);

-- name: MCPLockClientRegistry :exec
SELECT pg_advisory_xact_lock(68445130191025::bigint);

-- name: MCPLockAccountGrants :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(user_id)::uuid::text,6844513));

-- name: MCPAppendAudit :exec
INSERT INTO audit_log(actor_user_id,entity_id,action,request_id,metadata) VALUES($1,$2,$3,$4,$5);

-- name: MCPCleanupClients :exec
DELETE FROM mcp_oauth_clients WHERE id IN(SELECT c.id FROM mcp_oauth_clients c WHERE c.updated_at<clock_timestamp()-INTERVAL '7 days'
AND NOT EXISTS(SELECT 1 FROM mcp_oauth_requests r WHERE r.client_id=c.id) AND NOT EXISTS(SELECT 1 FROM mcp_oauth_grants g WHERE g.client_id=c.id) ORDER BY c.updated_at LIMIT 100);
