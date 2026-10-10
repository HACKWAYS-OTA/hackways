-- name: LockEventForRSVP :one
SELECT id, capacity, registration_count, status
FROM events
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: CreateRegistration :one
INSERT INTO registrations (event_id, user_id, guest_name, guest_email, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: IncrementRegistrationCount :exec
UPDATE events
SET registration_count = registration_count + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: GetRegistrationByID :one
SELECT * FROM registrations
WHERE id = $1;
