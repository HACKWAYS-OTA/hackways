-- name: LockRestrictionDevice :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(token_hash)::text,6844530));

-- name: IsRestrictedDevice :one
SELECT EXISTS(SELECT 1 FROM device_restrictions WHERE token_hash=$1 AND blocked);

-- name: IsRestrictedAccount :one
SELECT EXISTS(SELECT 1 FROM account_restrictions WHERE user_id=$1 AND blocked);

-- name: RestrictedDeviceByHash :one
SELECT id,reference FROM device_restrictions WHERE token_hash=$1 AND blocked;

-- name: BindSessionDevice :exec
UPDATE sessions SET device_token_hash=$2 WHERE id=$1;

-- name: RetireRestrictedAccountSessions :exec
UPDATE sessions SET revoked_at=COALESCE(revoked_at,clock_timestamp()) WHERE user_id=ANY(sqlc.arg(user_ids)::uuid[]);

-- name: RetireRestrictedAccountGrants :exec
UPDATE mcp_oauth_grants SET revoked_at=COALESCE(revoked_at,clock_timestamp()) WHERE user_id=ANY(sqlc.arg(user_ids)::uuid[]);

-- name: RecordDeviceReviewAssociation :exec
INSERT INTO restriction_device_accounts(device_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING;

-- name: RequireDeviceAccountReview :execrows
INSERT INTO account_restrictions(user_id,reason,source_device_id)
VALUES($1,'Sign-in from a blocked browser requires administrator review',$2)
ON CONFLICT(user_id) DO UPDATE SET blocked=true,reason=EXCLUDED.reason,
source_device_id=EXCLUDED.source_device_id,revision=account_restrictions.revision+1,
updated_by=NULL,updated_at=clock_timestamp() WHERE NOT account_restrictions.blocked;

-- name: HasActiveAdminGrant :one
SELECT EXISTS(SELECT 1 FROM platform_administrators WHERE user_id=$1 AND revoked_at IS NULL);

-- name: AdminSetAccountRestriction :exec
INSERT INTO account_restrictions(user_id,blocked,reason,updated_by)
VALUES($1,$2,$3,$4) ON CONFLICT(user_id) DO UPDATE SET blocked=EXCLUDED.blocked,
reason=EXCLUDED.reason,source_device_id=NULL,updated_by=EXCLUDED.updated_by,
revision=account_restrictions.revision+1,updated_at=clock_timestamp();

-- name: AdminRecognizedDevice :one
SELECT token_hash,user_agent FROM recognized_devices WHERE user_id=$1 AND admin_reference=$2 AND expires_at>clock_timestamp();

-- name: AdminDeviceHasAdministrator :one
SELECT EXISTS(SELECT 1 FROM platform_administrators a WHERE a.revoked_at IS NULL AND a.user_id IN(
 SELECT rd.user_id FROM recognized_devices rd WHERE rd.token_hash=sqlc.arg(token_hash)
 UNION SELECT s.user_id FROM sessions s WHERE s.device_token_hash=sqlc.arg(token_hash) AND s.revoked_at IS NULL AND s.expires_at>clock_timestamp()
 UNION SELECT r.user_id FROM restriction_device_accounts r JOIN device_restrictions d ON d.id=r.device_id WHERE d.token_hash=sqlc.arg(token_hash)));

-- name: AdminLockDeviceAccounts :many
SELECT u.id FROM users u WHERE u.deleted_at IS NULL AND u.id IN (
 SELECT rd.user_id FROM recognized_devices rd WHERE rd.token_hash=sqlc.arg(token_hash)
 UNION SELECT s.user_id FROM sessions s WHERE s.device_token_hash=sqlc.arg(token_hash) AND s.revoked_at IS NULL AND s.expires_at>clock_timestamp()
 UNION SELECT r.user_id FROM restriction_device_accounts r JOIN device_restrictions d ON d.id=r.device_id WHERE d.token_hash=sqlc.arg(token_hash)
 UNION SELECT sqlc.arg(target_user_id)::uuid)
ORDER BY u.id LIMIT 101 FOR UPDATE OF u;

-- name: AdminBlockDevice :one
INSERT INTO device_restrictions(token_hash,user_agent,reason,updated_by)
VALUES($1,$2,$3,$4) ON CONFLICT(token_hash) DO UPDATE SET blocked=true,
user_agent=EXCLUDED.user_agent,reason=EXCLUDED.reason,updated_by=EXCLUDED.updated_by,
revision=device_restrictions.revision+1,updated_at=clock_timestamp()
RETURNING id,reference,updated_at;

-- name: AdminRestrictDeviceAccounts :exec
INSERT INTO account_restrictions(user_id,reason,source_device_id,updated_by)
SELECT unnest(sqlc.arg(user_ids)::uuid[]),sqlc.arg(reason)::text,sqlc.arg(device_id)::uuid,sqlc.arg(actor_id)::uuid
ON CONFLICT(user_id) DO UPDATE SET blocked=true,
reason=CASE WHEN account_restrictions.blocked THEN account_restrictions.reason ELSE EXCLUDED.reason END,
source_device_id=CASE WHEN account_restrictions.blocked THEN account_restrictions.source_device_id ELSE EXCLUDED.source_device_id END,
updated_by=EXCLUDED.updated_by,revision=account_restrictions.revision+1,updated_at=clock_timestamp();

-- name: AdminRecordDeviceAccounts :exec
INSERT INTO restriction_device_accounts(device_id,user_id)
SELECT sqlc.arg(device_id)::uuid,unnest(sqlc.arg(user_ids)::uuid[]) ON CONFLICT DO NOTHING;

-- name: AdminTouchRestrictedAccounts :exec
UPDATE users SET updated_at=clock_timestamp() WHERE id=ANY(sqlc.arg(user_ids)::uuid[]);

-- name: AdminLockRestrictedDevice :one
SELECT id,token_hash,updated_at,blocked FROM device_restrictions WHERE id=$1 FOR UPDATE;

-- name: AdminUnblockDevice :exec
UPDATE device_restrictions SET blocked=false,reason=$2,updated_by=$3,revision=revision+1,updated_at=clock_timestamp() WHERE id=$1;

-- name: AdminListRestrictedDevices :many
SELECT id,reference,user_agent,reason,revision,updated_at,true::boolean AS blocked
FROM device_restrictions WHERE blocked AND (sqlc.arg(q)::text='' OR user_agent ILIKE '%'||sqlc.arg(q)||'%' OR reason ILIKE '%'||sqlc.arg(q)||'%')
ORDER BY updated_at DESC,id LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListRestrictedAccounts :many
SELECT u.id,u.admin_reference AS reference,u.display_name,u.email,u.updated_at,a.reason,true::boolean AS blocked
FROM account_restrictions a JOIN users u ON u.id=a.user_id WHERE a.blocked AND u.deleted_at IS NULL
AND (sqlc.arg(q)::text='' OR u.display_name ILIKE '%'||sqlc.arg(q)||'%' OR u.email ILIKE '%'||sqlc.arg(q)||'%')
ORDER BY a.updated_at DESC,u.id LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminPersonDevices :many
SELECT d.admin_reference::text AS id,d.admin_reference::text AS reference,'Browser'::text AS label,left(d.user_agent,200)::text AS detail,
'devices'::text AS resource,CASE WHEN EXISTS(SELECT 1 FROM device_restrictions r WHERE r.token_hash=d.token_hash AND r.blocked) THEN 'blocked' ELSE 'recognized' END::text AS status,
d.last_seen_at AS occurred_at
FROM recognized_devices d WHERE d.user_id=$1 AND d.expires_at>clock_timestamp()
ORDER BY d.last_seen_at DESC,d.admin_reference LIMIT $2 OFFSET $3;

-- name: AdminReadRestrictedDevice :one
SELECT token_hash FROM device_restrictions WHERE id=$1;
