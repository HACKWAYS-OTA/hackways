-- name: AdminPersonDetail :one
SELECT u.id,u.display_name,u.email,u.email_verified,COALESCE(split_part(split_part(u.avatar_url,'?',1),'#',1),'')::text AS avatar_url,u.created_at,u.updated_at,
CASE WHEN a.user_id IS NOT NULL AND a.revoked_at IS NULL THEN 'admin' ELSE 'member' END::text AS access,
(SELECT count(*) FROM sessions s WHERE s.user_id=u.id AND s.expires_at>clock_timestamp())::bigint AS active_sessions,
o.id AS host_profile_id,o.display_name AS host_name,o.slug AS host_slug,o.verified AS host_verified
FROM users u LEFT JOIN platform_administrators a ON a.user_id=u.id LEFT JOIN organizers o ON o.user_id=u.id WHERE u.id=$1 AND u.deleted_at IS NULL;

-- name: AdminMessageDetail :one
SELECT m.id,m.recipient_email,m.recipient_name,m.template_type,m.status::text AS status,m.retry_count,m.scheduled_at,m.sent_at,m.first_sent_at,
m.created_at,m.updated_at,m.event_id,e.title AS event_title,m.rsvp_id,r.public_code AS registration_code,
(m.status='failed' AND m.first_sent_at IS NULL)::boolean AS can_retry,
CASE WHEN m.status='sent' THEN 'Sent' WHEN m.first_sent_at IS NOT NULL THEN 'Provider attempt recorded' ELSE 'No provider attempt recorded' END::text AS delivery_state
FROM email_outbox m LEFT JOIN events e ON e.id=m.event_id LEFT JOIN rsvps r ON r.id=m.rsvp_id WHERE m.id=$1;

-- name: AdminAssetDetail :one
SELECT e.id,e.title AS name,'event_banner'::text AS kind,split_part(split_part(e.banner_url,'?',1),'#',1)::text AS url,e.updated_at,e.revision,'events'::text AS parent_resource
FROM events e WHERE e.id=sqlc.arg(id) AND e.deleted_at IS NULL AND sqlc.arg(kind)::text='event_banner' AND e.banner_url IS NOT NULL AND e.banner_url<>''
UNION ALL
SELECT u.id,u.display_name AS name,'avatar'::text AS kind,split_part(split_part(u.avatar_url,'?',1),'#',1)::text AS url,u.updated_at,0::bigint AS revision,'people'::text AS parent_resource
FROM users u WHERE u.id=sqlc.arg(id) AND u.deleted_at IS NULL AND sqlc.arg(kind)::text='avatar' AND u.avatar_url IS NOT NULL AND u.avatar_url<>'';

-- name: AdminPersonSessions :many
SELECT id::text AS id,'Session'::text AS label,COALESCE(left(user_agent,200),'')::text AS detail,'sessions'::text AS resource,
CASE WHEN id=sqlc.arg(current_session_id)::uuid THEN 'current' WHEN expires_at>clock_timestamp() THEN 'active' ELSE 'expired' END::text AS status,last_seen AS occurred_at,created_at,expires_at
FROM sessions WHERE user_id=sqlc.arg(user_id) AND expires_at>clock_timestamp() ORDER BY last_seen DESC,id LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminPersonEvents :many
SELECT e.id::text AS id,e.title AS label,e.slug AS detail,'events'::text AS resource,e.status::text AS status,e.starts_at AS occurred_at
FROM events e JOIN organizers o ON o.id=e.organizer_id WHERE o.user_id=$1 AND e.deleted_at IS NULL ORDER BY e.created_at DESC,e.id LIMIT $2 OFFSET $3;

-- name: AdminPersonRegistrations :many
SELECT r.id::text AS id,r.public_code AS label,e.title AS detail,'registrations'::text AS resource,r.status::text AS status,r.registered_at AS occurred_at
FROM rsvps r JOIN events e ON e.id=r.event_id WHERE r.user_id=$1 ORDER BY r.registered_at DESC,r.id LIMIT $2 OFFSET $3;

-- name: AdminPersonMessages :many
SELECT m.id::text AS id,m.template_type AS label,m.recipient_email AS detail,'messaging'::text AS resource,m.status::text AS status,m.created_at AS occurred_at
FROM email_outbox m JOIN users u ON lower(m.recipient_email)=lower(u.email) WHERE u.id=$1 ORDER BY m.created_at DESC,m.id LIMIT $2 OFFSET $3;

-- name: AdminGlobalEntityHistory :many
SELECT a.id::text AS id,a.action AS label,COALESCE(u.display_name,'System')::text AS detail,'audit'::text AS resource,''::text AS status,a.created_at AS occurred_at
FROM audit_log a LEFT JOIN users u ON u.id=a.actor_user_id WHERE a.entity_id=$1 OR a.actor_user_id=$1 ORDER BY a.id DESC LIMIT $2 OFFSET $3;

-- name: AdminUserLock :one
SELECT id,display_name,email_verified,updated_at FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE;

-- name: AdminUpdatePersonName :exec
UPDATE users SET display_name=$2,updated_at=clock_timestamp() WHERE id=$1 AND deleted_at IS NULL;

-- name: AdminTouchPerson :exec
UPDATE users SET updated_at=clock_timestamp() WHERE id=$1 AND deleted_at IS NULL;

-- name: AdminLockRoleChanges :exec
SELECT pg_advisory_xact_lock(7261042201);

-- name: AdminGrantAccess :exec
INSERT INTO platform_administrators(user_id,granted_by) VALUES($1,$2)
ON CONFLICT(user_id) DO UPDATE SET revoked_at=NULL,granted_by=EXCLUDED.granted_by,granted_at=clock_timestamp();

-- name: AdminRevokeAccess :execrows
UPDATE platform_administrators SET revoked_at=clock_timestamp() WHERE user_id=$1 AND revoked_at IS NULL;

-- name: AdminActiveCount :one
SELECT count(*)::bigint FROM platform_administrators a JOIN users u ON u.id=a.user_id WHERE a.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified;

-- name: AdminRevokePersonSession :execrows
DELETE FROM sessions WHERE user_id=$1 AND id=$2 AND id<>sqlc.arg(current_session_id)::uuid;

-- name: AdminRevokeAdminSessions :exec
DELETE FROM sessions WHERE user_id=$1 AND id IN (SELECT session_id FROM platform_admin_sessions);

-- name: AdminRetryMessage :one
UPDATE email_outbox SET status='pending',retry_count=0,scheduled_at=clock_timestamp(),updated_at=clock_timestamp(),error_message=NULL,lease_token=NULL
WHERE id=$1 AND status='failed' AND first_sent_at IS NULL AND updated_at=$2 RETURNING id;

-- name: AdminDetachEventBanner :one
UPDATE events SET banner_url=NULL,revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND deleted_at IS NULL AND updated_at=$2 AND banner_url IS NOT NULL RETURNING id;

-- name: AdminDetachAvatar :one
UPDATE users SET avatar_url=NULL,updated_at=clock_timestamp() WHERE id=$1 AND deleted_at IS NULL AND updated_at=$2 AND avatar_url IS NOT NULL RETURNING id;

-- name: AppendAdminAudit :exec
INSERT INTO audit_log(actor_user_id,event_id,entity_id,action,request_id,metadata) VALUES($1,$2,$3,$4,$5,$6);

-- name: AdminActionReplay :one
SELECT request_hash,response_json FROM idempotency_records WHERE scope_hash=$1 AND key_hash=$2 AND expires_at>clock_timestamp();

-- name: AdminActionRemember :exec
INSERT INTO idempotency_records(scope_hash,key_hash,request_hash,response_json) VALUES($1,$2,$3,$4)
ON CONFLICT(scope_hash,key_hash) DO UPDATE SET request_hash=EXCLUDED.request_hash,response_json=EXCLUDED.response_json,created_at=clock_timestamp(),expires_at=EXCLUDED.expires_at
WHERE idempotency_records.expires_at<=clock_timestamp();

-- name: AdminSerializeAction :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(scope)::text,0));
