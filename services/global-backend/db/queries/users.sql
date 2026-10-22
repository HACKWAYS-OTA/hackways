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

-- name: GetUserByGoogleSubject :one
SELECT u.id, u.email, u.display_name, u.avatar_url, u.email_verified, u.created_at, u.updated_at
FROM user_identities i
JOIN users u ON u.id = i.user_id
WHERE i.provider = 'google' AND i.subject = $1 AND u.deleted_at IS NULL;

-- name: GetOrCreateOAuthUser :one
WITH inserted AS (
    INSERT INTO users (email, password_hash, display_name, avatar_url, email_verified)
    VALUES ($1, $2, $3, $4, TRUE)
    ON CONFLICT (lower(email)) WHERE deleted_at IS NULL DO NOTHING
    RETURNING id, email, display_name, avatar_url, email_verified, created_at, updated_at
)
SELECT id, email, display_name, avatar_url, email_verified, created_at, updated_at
FROM inserted
UNION ALL
SELECT id, email, display_name, avatar_url, email_verified, created_at, updated_at
FROM users
WHERE lower(email) = lower($1) AND deleted_at IS NULL
LIMIT 1;

-- name: LinkUserIdentity :exec
INSERT INTO user_identities (user_id, provider, subject)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified = TRUE, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrganizerByUserID :one
SELECT id, user_id, display_name, slug, bio, website_url, avatar_url, verified, created_at, updated_at
FROM organizers
WHERE user_id = $1;

-- name: UpdateOwnDisplayName :one
UPDATE users SET display_name = $2, updated_at = clock_timestamp()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, display_name, avatar_url, email_verified;
