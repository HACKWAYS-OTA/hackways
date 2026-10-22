-- name: IsPlatformAdministrator :one
SELECT EXISTS(SELECT 1 FROM platform_administrators a JOIN users u ON u.id=a.user_id
WHERE a.user_id=$1 AND a.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified);

-- name: IsPlatformAdminSession :one
SELECT EXISTS(SELECT 1 FROM platform_admin_sessions g JOIN sessions s ON s.id=g.session_id
JOIN platform_administrators a ON a.user_id=s.user_id JOIN users u ON u.id=s.user_id
WHERE g.session_id=$1 AND a.user_id=$2 AND a.revoked_at IS NULL
AND s.expires_at>clock_timestamp() AND u.deleted_at IS NULL AND u.email_verified);

-- name: RecordPlatformAdminSession :exec
INSERT INTO platform_admin_sessions(session_id) VALUES ($1);

-- name: AdminListPeople :many
SELECT u.id,u.admin_reference AS reference,u.display_name,u.email,u.email_verified,
CASE WHEN a.user_id IS NOT NULL AND a.revoked_at IS NULL THEN 'admin' ELSE 'member' END::text AS access,u.created_at
FROM users u LEFT JOIN platform_administrators a ON a.user_id=u.id
WHERE u.deleted_at IS NULL AND (sqlc.arg(q)::text='' OR u.email ILIKE '%'||sqlc.arg(q)||'%' OR u.display_name ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR (CASE WHEN a.user_id IS NOT NULL AND a.revoked_at IS NULL THEN 'admin' ELSE 'member' END)=sqlc.arg(state)) ORDER BY u.created_at DESC,u.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListMessages :many
SELECT id,admin_reference AS reference,recipient_email,template_type,status::text AS status,retry_count,scheduled_at,sent_at,created_at
FROM email_outbox WHERE (sqlc.arg(q)::text='' OR recipient_email ILIKE '%'||sqlc.arg(q)||'%' OR template_type ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR status::text=sqlc.arg(state)) ORDER BY created_at DESC,id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListAssets :many
WITH assets AS (
SELECT e.id,e.admin_reference AS reference,e.title AS name,'event_banner'::text AS kind,split_part(split_part(e.banner_url,'?',1),'#',1)::text AS url,e.updated_at
FROM events e WHERE e.deleted_at IS NULL AND e.banner_url IS NOT NULL AND e.banner_url<>''
UNION ALL
SELECT u.id,u.admin_reference AS reference,u.display_name AS name,'avatar'::text AS kind,split_part(split_part(u.avatar_url,'?',1),'#',1)::text AS url,u.updated_at
FROM users u WHERE u.deleted_at IS NULL AND u.avatar_url IS NOT NULL AND u.avatar_url<>''
)
SELECT id,reference,name,kind,url,updated_at FROM assets
WHERE (sqlc.arg(q)::text='' OR name ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR kind=sqlc.arg(state)) ORDER BY updated_at DESC,id,kind
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;


-- name: GrantLocalPreviewAdmin :one
INSERT INTO platform_administrators(user_id)
SELECT id FROM users WHERE lower(email)=lower(sqlc.arg(email)) AND email_verified AND deleted_at IS NULL
ON CONFLICT(user_id) DO UPDATE SET revoked_at=NULL
RETURNING user_id;
