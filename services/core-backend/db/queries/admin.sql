-- name: IsPlatformAdministrator :one
SELECT EXISTS(SELECT 1 FROM platform_administrators a JOIN users u ON u.id=a.user_id
WHERE a.user_id=$1 AND a.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified);

-- name: IsPlatformAdminSession :one
SELECT EXISTS(SELECT 1 FROM platform_admin_sessions g JOIN sessions s ON s.id=g.session_id
JOIN platform_administrators a ON a.user_id=s.user_id JOIN users u ON u.id=s.user_id
WHERE g.session_id=$1 AND a.user_id=$2 AND a.revoked_at IS NULL
AND g.otp_verified_at>clock_timestamp()-INTERVAL '24 hours' AND s.expires_at>clock_timestamp() AND s.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified);

-- name: IsPlatformAdminIdentitySession :one
SELECT EXISTS(SELECT 1 FROM platform_admin_sessions g JOIN sessions s ON s.id=g.session_id
JOIN platform_administrators a ON a.user_id=s.user_id JOIN users u ON u.id=s.user_id
WHERE g.session_id=$1 AND a.user_id=$2 AND a.revoked_at IS NULL
AND s.expires_at>clock_timestamp() AND s.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified);

-- name: RecordPlatformAdminSession :exec
-- Trusted, already verified server-side issuance; password sign-in uses the
-- separate unverified-session query before the email challenge.
INSERT INTO platform_admin_sessions(session_id,otp_verified_at) VALUES ($1,clock_timestamp());

-- name: AdminListEvents :many
SELECT e.id,e.admin_reference AS reference,e.title,e.slug,e.status::text AS status,o.display_name AS host_name,e.starts_at,e.current_rsvp_count,e.capacity
FROM events e JOIN organizers o ON o.id=e.organizer_id
WHERE e.deleted_at IS NULL AND (sqlc.arg(q)::text='' OR e.title ILIKE '%'||sqlc.arg(q)||'%' OR e.slug ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR e.status::text=sqlc.arg(state)) ORDER BY e.created_at DESC,e.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListSpaces :many
SELECT s.id,s.admin_reference AS reference,s.name,s.slug,s.status,o.display_name AS owner_name,
(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id) AS member_count,s.created_at
FROM spaces s JOIN organizers o ON o.id=s.organizer_id
WHERE (sqlc.arg(q)::text='' OR s.name ILIKE '%'||sqlc.arg(q)||'%' OR s.slug ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR s.status=sqlc.arg(state)) ORDER BY s.created_at DESC,s.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListRegistrations :many
SELECT r.id,r.admin_reference AS reference,r.public_code,e.title AS event_title,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS name,
COALESCE(u.email,r.guest_email,'')::text AS email,r.status::text AS status,r.registered_at
FROM rsvps r JOIN events e ON e.id=r.event_id LEFT JOIN users u ON u.id=r.user_id
WHERE (sqlc.arg(q)::text='' OR r.public_code ILIKE '%'||sqlc.arg(q)||'%' OR e.title ILIKE '%'||sqlc.arg(q)||'%' OR COALESCE(u.email,r.guest_email,'') ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR r.status::text=sqlc.arg(state)) ORDER BY r.registered_at DESC,r.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListTickets :many
SELECT t.id,t.admin_reference AS reference,t.public_code,e.title AS event_title,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS name,t.status,t.checked_in_at,t.created_at
FROM tickets t JOIN rsvps r ON r.id=t.rsvp_id JOIN events e ON e.id=r.event_id LEFT JOIN users u ON u.id=r.user_id
WHERE (sqlc.arg(q)::text='' OR t.public_code ILIKE '%'||sqlc.arg(q)||'%' OR e.title ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR t.status=sqlc.arg(state)) ORDER BY t.created_at DESC,t.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListModeration :many
SELECT e.id,e.admin_reference AS reference,e.title,e.slug,e.status::text AS status,o.display_name AS host_name,o.verified AS host_verified,e.updated_at
FROM events e JOIN organizers o ON o.id=e.organizer_id
WHERE e.deleted_at IS NULL AND e.status IN ('draft','unpublished')
AND (sqlc.arg(q)::text='' OR e.title ILIKE '%'||sqlc.arg(q)||'%' OR e.slug ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR e.status::text=sqlc.arg(state)) ORDER BY e.updated_at DESC,e.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminListAudit :many
SELECT a.id,a.id AS reference,a.action,COALESCE(u.display_name,'System')::text AS actor_name,COALESCE(e.title,'')::text AS event_title,a.entity_id,a.created_at
FROM audit_log a LEFT JOIN users u ON u.id=a.actor_user_id LEFT JOIN events e ON e.id=a.event_id
WHERE (sqlc.arg(q)::text='' OR a.action ILIKE '%'||sqlc.arg(q)||'%' OR COALESCE(e.title,'') ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR a.action=sqlc.arg(state)) ORDER BY a.id DESC
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;
