-- name: LockEventForRSVP :one
SELECT id, capacity, current_rsvp_count, status
FROM events
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: CreateRSVP :one
INSERT INTO rsvps (
    id,
    event_id,
    user_id,
    guest_name,
    guest_email,
    status,
    notes,
    registered_at,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    'confirmed',
    $5,
    NOW(),
    NOW(),
    NOW()
)
ON CONFLICT (event_id, user_id) DO NOTHING
RETURNING *;

-- name: IncrementRSVPCount :exec
UPDATE events
SET current_rsvp_count = current_rsvp_count + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: GetRSVPByID :one
SELECT * FROM rsvps
WHERE id = $1;

-- name: ListRSVPsByEvent :many
SELECT * FROM rsvps
WHERE event_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: EnqueueEmailOutbox :one
INSERT INTO email_outbox (
    id,
    rsvp_id,
    event_id,
    recipient_email,
    recipient_name,
    template_type,
    payload_json,
    status,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    'pending',
    NOW(),
    NOW()
)
RETURNING *;

-- name: FetchPendingOutbox :many
SELECT id, rsvp_id, event_id, recipient_email, recipient_name, template_type, payload_json, retry_count
FROM email_outbox
WHERE status = 'pending' AND scheduled_at <= NOW()
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxSent :exec
UPDATE email_outbox
SET status = 'sent',
    sent_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: MarkOutboxFailed :exec
UPDATE email_outbox
SET retry_count = retry_count + 1,
    scheduled_at = $2,
    status = $3,
    error_message = $4,
    updated_at = NOW()
WHERE id = $1;
