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
