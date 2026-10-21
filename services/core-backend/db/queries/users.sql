-- name: GetUserByID :one
SELECT id, email, display_name, avatar_url, email_verified, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, avatar_url, email_verified, created_at, updated_at
FROM users
WHERE lower(email) = lower($1) AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id, email, display_name, avatar_url, email_verified, created_at, updated_at;

-- name: GetOrganizerByUserID :one
SELECT id, user_id, display_name, slug, bio, website_url, avatar_url, verified, created_at, updated_at
FROM organizers
WHERE user_id = $1;

-- name: CreateOrganizer :one
INSERT INTO organizers (user_id, display_name, slug)
VALUES (
    $1,
    $2,
    trim(regexp_replace(lower($2), '[^a-z0-9]+', '-', 'g'), '-') || '-' || left(gen_random_uuid()::text, 8)
)
RETURNING id, user_id, display_name, slug, bio, website_url, avatar_url, verified, created_at, updated_at;
