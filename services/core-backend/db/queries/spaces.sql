-- name: ListPublicSpaces :many
SELECT s.id,s.organizer_id,s.name,s.slug,s.description,s.guidelines,s.website_url,s.logo_url,s.banner_url,s.status,s.revision,s.created_at,s.updated_at,
o.display_name AS owner_name,(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id)::bigint AS member_count
FROM spaces s JOIN organizers o ON o.id=s.organizer_id
WHERE s.status='active' AND (sqlc.arg(q)::text='' OR to_tsvector('simple',s.name || ' ' || s.description) @@ plainto_tsquery('simple',sqlc.arg(q)::text))
ORDER BY s.created_at DESC,s.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ReadPublicSpace :one
SELECT s.id,s.organizer_id,s.name,s.slug,s.description,s.guidelines,s.website_url,s.logo_url,s.banner_url,s.status,s.revision,s.created_at,s.updated_at,
o.display_name AS owner_name,(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id)::bigint AS member_count
FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.slug=$1;

-- name: ReadSpaceSummaryByID :one
SELECT s.id,s.organizer_id,s.name,s.slug,s.description,s.guidelines,s.website_url,s.logo_url,s.banner_url,s.status,s.revision,s.created_at,s.updated_at,
o.display_name AS owner_name,(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id)::bigint AS member_count
FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.id=$1;

-- name: ListMySpaces :many
SELECT s.id,s.organizer_id,s.name,s.slug,s.description,s.guidelines,s.website_url,s.logo_url,s.banner_url,s.status,s.revision,s.created_at,s.updated_at,
o.display_name AS owner_name,(o.user_id=sqlc.arg(user_id))::boolean AS owned,
(SELECT count(*) FROM space_memberships m WHERE m.space_id=s.id)::bigint AS member_count
FROM spaces s JOIN organizers o ON o.id=s.organizer_id
WHERE (sqlc.arg(owned_only)::boolean AND o.user_id=sqlc.arg(user_id))
OR (NOT sqlc.arg(owned_only)::boolean AND EXISTS(SELECT 1 FROM space_memberships m WHERE m.space_id=s.id AND m.user_id=sqlc.arg(user_id)))
ORDER BY s.created_at DESC,s.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ReadSpaceMembership :one
SELECT (o.user_id=$2)::boolean AS owned,
EXISTS(SELECT 1 FROM space_memberships m WHERE m.space_id=s.id AND m.user_id=$2) AS joined,
(o.user_id=$2 OR EXISTS(SELECT 1 FROM space_access a WHERE a.space_id=s.id AND a.user_id=$2 AND a.revoked_at IS NULL))::boolean AS managed
FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.id=$1;

-- name: LockSpace :one
SELECT * FROM spaces WHERE id=$1 FOR UPDATE;

-- name: ReadSpaceID :one
SELECT id FROM spaces WHERE id=$1;

-- name: LockOwnedActiveSpace :one
SELECT * FROM spaces WHERE id=$1 AND organizer_id=$2 AND status='active' FOR UPDATE;

-- name: LockOwnedActiveSpaceBySlug :one
SELECT * FROM spaces WHERE slug=$1 AND organizer_id=$2 AND status='active' FOR UPDATE;

-- name: ReadOwnedSpace :one
SELECT * FROM spaces WHERE id=$1 AND organizer_id=$2;

-- name: ReadSpaceCreation :one
SELECT * FROM spaces WHERE organizer_id=$1 AND creation_key=$2;

-- name: CreateSpace :one
INSERT INTO spaces(organizer_id,name,slug,description,guidelines,website_url,creation_key,creation_hash)
VALUES($1,$2,COALESCE(NULLIF(trim(regexp_replace(lower($2),'[^a-z0-9]+','-','g'),'-'),''),'space') || '-' || left(gen_random_uuid()::text,8),$3,$4,$5,$6,$7)
ON CONFLICT(organizer_id,creation_key) DO NOTHING RETURNING *;

-- name: UpdateSpace :one
UPDATE spaces SET name=$3,description=$4,guidelines=$5,website_url=$6,revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND organizer_id=$2 AND revision=$7 RETURNING *;

-- name: SetSpaceStatus :one
UPDATE spaces SET status=$3,revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND organizer_id=$2 AND revision=$4 RETURNING *;

-- name: JoinSpace :exec
INSERT INTO space_memberships(space_id,user_id) VALUES($1,$2) ON CONFLICT(space_id,user_id) DO NOTHING;

-- name: LeaveSpace :exec
DELETE FROM space_memberships WHERE space_id=$1 AND user_id=$2;

-- name: ListSpaceEvents :many
SELECT * FROM events WHERE space_id=$1 AND deleted_at IS NULL AND visibility='public'
AND (status='published' OR (status='cancelled' AND published_at IS NOT NULL))
AND ((sqlc.arg(past)::boolean AND ends_at<=clock_timestamp()) OR (NOT sqlc.arg(past)::boolean AND status='published' AND ends_at>clock_timestamp()))
ORDER BY CASE WHEN sqlc.arg(past)::boolean THEN ends_at END DESC, starts_at,id
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: SetEventSpace :one
UPDATE events SET space_id=$2,revision=revision+1,updated_at=clock_timestamp() WHERE id=$1 RETURNING *;
