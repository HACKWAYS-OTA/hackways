-- name: ListManagedEvents :many
SELECT * FROM events
WHERE organizer_id=sqlc.arg(organizer_id) AND deleted_at IS NULL
AND (sqlc.arg(search)::text='' OR strpos(lower(title),lower(sqlc.arg(search)::text))>0)
AND (sqlc.arg(view)::text='all'
 OR (sqlc.arg(view)::text='upcoming' AND status IN ('published','unpublished') AND ends_at>clock_timestamp())
 OR (sqlc.arg(view)::text='drafts' AND status='draft')
 OR (sqlc.arg(view)::text='past' AND status<>'draft' AND status<>'cancelled' AND ends_at<=clock_timestamp())
 OR (sqlc.arg(view)::text='cancelled' AND status='cancelled'))
ORDER BY CASE WHEN sqlc.arg(view)::text='upcoming' THEN starts_at END ASC,created_at DESC,id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: SetEventAppearance :one
UPDATE events SET theme=$2,theme_color=$3,theme_font=$4,theme_style=$5,theme_display=$6,
 banner_url=CASE WHEN sqlc.arg(remove_banner)::boolean THEN NULL ELSE banner_url END,
 revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 RETURNING *;

-- name: EventGuestSummary :one
SELECT count(*)::bigint AS total,
 count(*) FILTER(WHERE r.status='confirmed')::bigint AS confirmed,
 count(*) FILTER(WHERE r.status='pending_approval')::bigint AS pending,
 count(*) FILTER(WHERE r.status='waitlisted')::bigint AS waitlisted,
 count(*) FILTER(WHERE r.status='offered')::bigint AS offered,
 count(*) FILTER(WHERE r.status='cancelled')::bigint AS cancelled,
 count(*) FILTER(WHERE r.status='rejected')::bigint AS rejected,
 count(*) FILTER(WHERE r.status='expired')::bigint AS expired,
 count(*) FILTER(WHERE r.status='transferred')::bigint AS transferred,
 count(*) FILTER(WHERE t.status='scanned' AND r.status='confirmed')::bigint AS checked_in
FROM rsvps r LEFT JOIN tickets t ON t.rsvp_id=r.id WHERE r.event_id=$1;

-- name: EventRosterFiltered :many
SELECT r.id,r.public_code,r.user_id,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS guest_name,
 COALESCE(NULLIF(r.guest_email,''),u.email,'')::text AS guest_email,
 r.status,r.notes,r.registered_at,r.offer_expires_at,t.id AS ticket_id,t.status AS ticket_status,t.checked_in_at
FROM rsvps r LEFT JOIN tickets t ON t.rsvp_id=r.id LEFT JOIN users u ON u.id=r.user_id
WHERE r.event_id=sqlc.arg(event_id)
AND (sqlc.arg(status_filter)::text='all'
 OR (sqlc.arg(status_filter)::text='checked_in' AND t.status='scanned' AND r.status='confirmed')
 OR (sqlc.arg(status_filter)::text='not_arrived' AND r.status='confirmed' AND t.status='issued')
 OR r.status::text=sqlc.arg(status_filter)::text)
AND (sqlc.arg(search)::text='' OR strpos(lower(COALESCE(r.guest_name,'') || ' ' || COALESCE(r.guest_email,'') || ' ' || COALESCE(u.display_name,'') || ' ' || COALESCE(u.email,'') || ' ' || r.public_code),lower(sqlc.arg(search)::text))>0)
ORDER BY r.registered_at DESC,r.id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);
