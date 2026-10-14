-- name: GetEventByID :one
SELECT * FROM events
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPublishedEvents :many
SELECT * FROM events
WHERE status = 'published'
  AND deleted_at IS NULL
  AND starts_at > NOW()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;

-- name: SearchPublishedEvents :many
SELECT * FROM events
WHERE status = 'published'
  AND deleted_at IS NULL
  AND to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $1)
ORDER BY ts_rank(to_tsvector('english', title || ' ' || description), plainto_tsquery('english', $1)) DESC
LIMIT $2 OFFSET $3;

-- name: UpdateEventStatus :exec
UPDATE events
SET status = $2,
    published_at = CASE WHEN $2 = 'published' THEN NOW() ELSE published_at END,
    cancelled_at = CASE WHEN $2 = 'cancelled' THEN NOW() ELSE cancelled_at END,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetEventByIDAndOrganizer :one
SELECT * FROM events
WHERE id = $1 AND organizer_id = $2 AND deleted_at IS NULL;
