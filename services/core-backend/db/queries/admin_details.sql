-- name: AdminEventDetail :one
SELECT e.id,e.admin_reference AS reference,e.organizer_id,o.user_id AS host_user_id,(SELECT admin_reference FROM users WHERE id=o.user_id) AS host_user_reference,o.display_name AS host_name,e.title,e.slug,e.description,
e.status::text AS status,e.format::text AS format,e.starts_at,e.ends_at,e.timezone,e.capacity,e.current_rsvp_count,e.held_rsvp_count,
e.location_name,e.location_address,e.online_url,e.visibility,e.registration_mode,e.allow_guest,e.waitlist_enabled,e.offer_minutes,
e.registration_opens_at,e.registration_closes_at,COALESCE(split_part(split_part(e.banner_url,'?',1),'#',1),'')::text AS banner_url,e.space_id,linked_space.admin_reference AS space_reference,e.revision,e.created_at,e.updated_at,
(e.status!='cancelled' AND e.starts_at>clock_timestamp())::boolean AS editable
FROM events e JOIN organizers o ON o.id=e.organizer_id LEFT JOIN spaces linked_space ON linked_space.id=e.space_id WHERE e.id=$1 AND e.deleted_at IS NULL;

-- name: AdminSpaceDetail :one
SELECT s.id,s.admin_reference AS reference,s.organizer_id,o.user_id AS owner_user_id,(SELECT admin_reference FROM users WHERE id=o.user_id) AS owner_user_reference,o.display_name AS owner_name,s.name,s.slug,s.description,s.guidelines,s.website_url,
s.status,s.revision,s.created_at,s.updated_at,(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id)::bigint AS member_count
FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.id=$1;

-- name: AdminRegistrationDetail :one
SELECT r.id,r.admin_reference AS reference,r.public_code,r.event_id,e.admin_reference AS event_reference,e.title AS event_title,r.user_id,u.admin_reference AS user_reference,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS name,
COALESCE(u.email,r.guest_email,'')::text AS email,r.status::text AS status,r.notes,r.registered_at,r.updated_at,r.offer_expires_at,
r.approval_expires_at,r.transfer_count,t.id AS ticket_id,t.admin_reference AS ticket_reference,t.status AS ticket_status,
(e.status='published' AND e.starts_at>clock_timestamp())::boolean AS editable,
(e.status='published' AND e.starts_at>clock_timestamp() AND (e.registration_opens_at IS NULL OR e.registration_opens_at<=clock_timestamp())
 AND (e.registration_closes_at IS NULL OR e.registration_closes_at>clock_timestamp()))::boolean AS registration_open
FROM rsvps r JOIN events e ON e.id=r.event_id LEFT JOIN users u ON u.id=r.user_id LEFT JOIN tickets t ON t.rsvp_id=r.id WHERE r.id=$1;

-- name: AdminTicketDetail :one
SELECT t.id,t.admin_reference AS reference,t.public_code,t.rsvp_id,r.admin_reference AS rsvp_reference,r.event_id,e.admin_reference AS event_reference,e.title AS event_title,r.user_id,u.admin_reference AS user_reference,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS name,
COALESCE(u.email,r.guest_email,'')::text AS email,t.status,t.checked_in_at,t.checked_in_by,checked_actor.admin_reference AS checked_in_by_reference,t.created_at,t.updated_at,
(e.status='published' AND r.status='confirmed' AND t.status='issued' AND clock_timestamp()>=e.starts_at-INTERVAL '24 hours' AND clock_timestamp()<=e.ends_at)::boolean AS can_check_in
FROM tickets t JOIN rsvps r ON r.id=t.rsvp_id JOIN events e ON e.id=r.event_id LEFT JOIN users u ON u.id=r.user_id LEFT JOIN users checked_actor ON checked_actor.id=t.checked_in_by WHERE t.id=$1;

-- name: AdminAuditDetail :one
SELECT a.id,a.id AS reference,a.action,a.actor_user_id,u.admin_reference AS actor_user_reference,COALESCE(u.display_name,'System')::text AS actor_name,a.event_id,e.admin_reference AS event_reference,COALESCE(e.title,'')::text AS event_title,
a.entity_id,a.request_id,a.metadata::text AS metadata,a.created_at
FROM audit_log a LEFT JOIN users u ON u.id=a.actor_user_id LEFT JOIN events e ON e.id=a.event_id WHERE a.id=$1;

-- name: AdminEntityHistory :many
SELECT a.id::text AS id,a.id::text AS reference,a.action AS label,COALESCE(u.display_name,'System')::text AS detail,'audit'::text AS resource,''::text AS status,a.created_at AS occurred_at
FROM audit_log a LEFT JOIN users u ON u.id=a.actor_user_id WHERE a.entity_id=$1 ORDER BY a.id DESC LIMIT $2 OFFSET $3;

-- name: AdminEventRegistrations :many
SELECT r.id::text AS id,r.admin_reference::text AS reference,r.public_code AS label,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS detail,'registrations'::text AS resource,r.status::text AS status,r.registered_at AS occurred_at
FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=$1 ORDER BY r.registered_at DESC,r.id LIMIT $2 OFFSET $3;

-- name: AdminEventTickets :many
SELECT t.id::text AS id,t.admin_reference::text AS reference,t.public_code AS label,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS detail,'tickets'::text AS resource,t.status,t.created_at AS occurred_at
FROM tickets t JOIN rsvps r ON r.id=t.rsvp_id LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=$1 ORDER BY t.created_at DESC,t.id LIMIT $2 OFFSET $3;

-- name: AdminEventHistory :many
SELECT a.id::text AS id,a.id::text AS reference,a.action AS label,COALESCE(u.display_name,'System')::text AS detail,'audit'::text AS resource,''::text AS status,a.created_at AS occurred_at
FROM audit_log a LEFT JOIN users u ON u.id=a.actor_user_id WHERE a.event_id=$1 ORDER BY a.id DESC LIMIT $2 OFFSET $3;

-- name: AdminSpaceEvents :many
SELECT e.id::text AS id,e.admin_reference::text AS reference,e.title AS label,e.slug AS detail,'events'::text AS resource,e.status::text AS status,e.starts_at AS occurred_at
FROM events e WHERE e.space_id=$1 AND e.deleted_at IS NULL ORDER BY e.created_at DESC,e.id LIMIT $2 OFFSET $3;

-- name: AdminSpaceMembers :many
SELECT u.id::text AS id,u.admin_reference::text AS reference,u.display_name AS label,u.email AS detail,'people'::text AS resource,''::text AS status,m.joined_at AS occurred_at
FROM space_memberships m JOIN users u ON u.id=m.user_id WHERE m.space_id=$1 AND u.deleted_at IS NULL ORDER BY m.joined_at DESC,u.id LIMIT $2 OFFSET $3;

-- name: AdminOrganizerByID :one
SELECT * FROM organizers WHERE id=$1;

-- name: AdminEditEvent :one
UPDATE events SET title=$2,description=$3,capacity=$4,visibility=$5,revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND deleted_at IS NULL AND revision=$6 RETURNING id;

-- name: AdminResolveReferenceID :one
SELECT id FROM events WHERE sqlc.arg(resource)::text IN ('events','moderation') AND admin_reference=sqlc.arg(reference)::bigint
UNION ALL
SELECT id FROM spaces WHERE sqlc.arg(resource)::text='spaces' AND admin_reference=sqlc.arg(reference)::bigint
UNION ALL
SELECT id FROM rsvps WHERE sqlc.arg(resource)::text='registrations' AND admin_reference=sqlc.arg(reference)::bigint
UNION ALL
SELECT id FROM tickets WHERE sqlc.arg(resource)::text='tickets' AND admin_reference=sqlc.arg(reference)::bigint;
