-- name: GetEventByID :one
SELECT * FROM events
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPublishedEventByID :one
SELECT * FROM events
WHERE id = $1 AND (status = 'published' OR (status = 'cancelled' AND published_at IS NOT NULL)) AND deleted_at IS NULL;

-- name: GetPublishedEventBySlug :one
SELECT e.*, o.display_name AS organizer_name, o.bio AS organizer_bio,
       o.website_url AS organizer_website, o.verified AS organizer_verified,
       s.name AS space_name,s.slug AS space_slug
FROM events e
JOIN organizers o ON o.id = e.organizer_id
LEFT JOIN spaces s ON s.id=e.space_id
WHERE e.slug = $1 AND (e.status = 'published' OR (e.status = 'cancelled' AND e.published_at IS NOT NULL)) AND e.deleted_at IS NULL;

-- name: ListPublishedEvents :many
SELECT * FROM events
WHERE status = 'published'
  AND visibility = 'public'
  AND deleted_at IS NULL
  AND starts_at > NOW()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;

-- name: SearchPublishedEvents :many
SELECT * FROM events
WHERE status = 'published'
  AND visibility = 'public'
  AND deleted_at IS NULL
  AND to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $1)
ORDER BY ts_rank(to_tsvector('english', title || ' ' || description), plainto_tsquery('english', $1)) DESC
LIMIT $2 OFFSET $3;

-- name: UpdateEventStatus :exec
UPDATE events
SET status = sqlc.arg(status)::event_status,
    published_at = CASE WHEN sqlc.arg(status)::event_status = 'published' THEN NOW() ELSE published_at END,
    cancelled_at = CASE WHEN sqlc.arg(status)::event_status = 'cancelled' THEN NOW() ELSE cancelled_at END,
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: GetEventByIDAndOrganizer :one
SELECT * FROM events
WHERE id = $1 AND organizer_id = $2 AND deleted_at IS NULL;

-- name: GetEventBySlugAndOrganizer :one
SELECT * FROM events WHERE slug=$1 AND organizer_id=$2 AND deleted_at IS NULL;

-- name: ListOrganizerEvents :many
SELECT * FROM events
WHERE organizer_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateEventRecord :one
INSERT INTO events (
    organizer_id, status, format, title, slug, description,
    starts_at, ends_at, timezone, capacity, location_name, online_url,
    published_at, registration_mode, allow_guest,space_id,theme,theme_color,theme_font,theme_style,theme_display,visibility
)
VALUES (
    $1, sqlc.arg(status)::event_status, $2, $3,
    trim(regexp_replace(lower($3), '[^a-z0-9]+', '-', 'g'), '-') || '-' || left(gen_random_uuid()::text, 8),
    $4, $5, $6, $7, $8, $9, $10, CASE WHEN sqlc.arg(status)::event_status = 'published' THEN NOW() ELSE NULL END,
    COALESCE(NULLIF(sqlc.arg(registration_mode)::text,''),'instant'),
    COALESCE(NULLIF(sqlc.arg(registration_mode)::text,''),'instant') <> 'approval',sqlc.narg(space_id)::uuid,
    COALESCE(NULLIF(sqlc.arg(theme)::text,''),'minimal'),COALESCE(NULLIF(sqlc.arg(theme_color)::text,''),'#ddd4e9'),
    COALESCE(NULLIF(sqlc.arg(theme_font)::text,''),'futurist'),COALESCE(NULLIF(sqlc.arg(theme_style)::text,''),'soft'),
    COALESCE(NULLIF(sqlc.arg(theme_display)::text,''),'light'),COALESCE(NULLIF(sqlc.arg(visibility)::text,''),'public')
)
RETURNING *;
