-- name: EventWorkspaceAccess :one
SELECT e.id AS resource_id,e.slug,sqlc.embed(o),
 CASE WHEN o.user_id=sqlc.arg(actor_id)::uuid THEN 'owner' ELSE COALESCE(a.role,s.role,'') END::text AS access_role
FROM events e JOIN organizers o ON o.id=e.organizer_id
LEFT JOIN event_access a ON a.event_id=e.id AND a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
LEFT JOIN space_access s ON s.space_id=e.space_id AND s.user_id=sqlc.arg(actor_id) AND s.revoked_at IS NULL
WHERE e.deleted_at IS NULL AND (e.id=sqlc.narg(resource_id)::uuid OR e.slug=sqlc.arg(slug)::text)
AND (o.user_id=sqlc.arg(actor_id) OR a.id IS NOT NULL OR s.id IS NOT NULL);

-- name: SpaceWorkspaceAccess :one
SELECT s.id AS resource_id,s.slug,sqlc.embed(o),
 CASE WHEN o.user_id=sqlc.arg(actor_id)::uuid THEN 'owner' ELSE COALESCE(a.role,'') END::text AS access_role
FROM spaces s JOIN organizers o ON o.id=s.organizer_id
LEFT JOIN space_access a ON a.space_id=s.id AND a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
WHERE (s.id=sqlc.narg(resource_id)::uuid OR s.slug=sqlc.arg(slug)::text)
AND (o.user_id=sqlc.arg(actor_id) OR a.id IS NOT NULL);

-- name: ListWorkspaceSpaces :many
WITH permitted AS (
 SELECT s.id FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE o.user_id=sqlc.arg(actor_id)::uuid
 UNION SELECT a.space_id FROM space_access a WHERE a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
)
SELECT s.id,s.slug,s.name,s.description,s.logo_url,s.banner_url,s.status,s.revision,s.updated_at,CASE WHEN o.user_id=sqlc.arg(actor_id)::uuid THEN 'owner' ELSE a.role END::text AS access_role
FROM permitted p JOIN spaces s ON s.id=p.id JOIN organizers o ON o.id=s.organizer_id
LEFT JOIN space_access a ON a.space_id=s.id AND a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
WHERE (o.user_id=sqlc.arg(actor_id) OR a.id IS NOT NULL)
AND (sqlc.arg(q)::text='' OR s.name ILIKE '%' || sqlc.arg(q)::text || '%')
ORDER BY s.updated_at DESC,s.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ListWorkspaceEvents :many
WITH permitted AS (
 SELECT e.id FROM events e JOIN organizers o ON o.id=e.organizer_id WHERE o.user_id=sqlc.arg(actor_id)::uuid AND e.deleted_at IS NULL
 UNION SELECT a.event_id FROM event_access a WHERE a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
 UNION SELECT e.id FROM space_access a JOIN events e ON e.space_id=a.space_id WHERE a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL AND e.deleted_at IS NULL
)
SELECT e.*,CASE WHEN o.user_id=sqlc.arg(actor_id)::uuid THEN 'owner' ELSE COALESCE(a.role,s.role) END::text AS access_role
FROM permitted allowed JOIN events e ON e.id=allowed.id JOIN organizers o ON o.id=e.organizer_id
LEFT JOIN spaces p ON p.id=e.space_id
LEFT JOIN event_access a ON a.event_id=e.id AND a.user_id=sqlc.arg(actor_id) AND a.revoked_at IS NULL
LEFT JOIN space_access s ON s.space_id=e.space_id AND s.user_id=sqlc.arg(actor_id) AND s.revoked_at IS NULL
WHERE e.deleted_at IS NULL AND (o.user_id=sqlc.arg(actor_id) OR a.id IS NOT NULL OR s.id IS NOT NULL)
AND (sqlc.arg(space_slug)::text='' OR p.slug=sqlc.arg(space_slug))
AND (sqlc.arg(q)::text='' OR e.title ILIKE '%' || sqlc.arg(q)::text || '%')
AND (sqlc.arg(view)::text='all' OR (sqlc.arg(view)::text='drafts' AND e.status='draft') OR (sqlc.arg(view)::text='upcoming' AND e.status='published' AND e.ends_at>now()) OR (sqlc.arg(view)::text='past' AND e.ends_at<=now()) OR (sqlc.arg(view)::text='cancelled' AND e.status='cancelled'))
ORDER BY e.starts_at DESC,e.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: WorkspaceAccountByEmail :one
SELECT id,display_name,email FROM users WHERE lower(email)=sqlc.arg(email)::text AND email_verified AND deleted_at IS NULL;

-- name: ListSpaceAccess :many
SELECT a.id,a.user_id,u.display_name,u.email,a.role,a.revision,a.revoked_at,a.updated_at
FROM space_access a JOIN users u ON u.id=a.user_id WHERE a.space_id=$1 AND u.deleted_at IS NULL
ORDER BY a.created_at,a.id LIMIT $2 OFFSET $3;

-- name: ListEventAccess :many
SELECT a.id,a.user_id,u.display_name,u.email,a.role,a.revision,a.revoked_at,a.updated_at
FROM event_access a JOIN users u ON u.id=a.user_id WHERE a.event_id=$1 AND u.deleted_at IS NULL
ORDER BY a.created_at,a.id LIMIT $2 OFFSET $3;

-- name: ReadSpaceAccess :one
SELECT * FROM space_access WHERE space_id=$1 AND user_id=$2;

-- name: ReadEventAccess :one
SELECT * FROM event_access WHERE event_id=$1 AND user_id=$2;

-- name: ReadSpaceAccessByID :one
SELECT * FROM space_access WHERE space_id=$1 AND id=$2;

-- name: ReadEventAccessByID :one
SELECT * FROM event_access WHERE event_id=$1 AND id=$2;

-- name: CreateSpaceAccess :one
INSERT INTO space_access(space_id,user_id,role,granted_by,updated_by) VALUES($1,$2,$3,$4,$4) RETURNING *;

-- name: CreateEventAccess :one
INSERT INTO event_access(event_id,user_id,role,granted_by,updated_by) VALUES($1,$2,$3,$4,$4) RETURNING *;

-- name: UpdateSpaceAccess :one
UPDATE space_access SET role=$3,revoked_at=NULL,revision=revision+1,updated_by=$4,updated_at=clock_timestamp()
WHERE space_id=$1 AND user_id=$2 AND revision=$5 RETURNING *;

-- name: UpdateEventAccess :one
UPDATE event_access SET role=$3,revoked_at=NULL,revision=revision+1,updated_by=$4,updated_at=clock_timestamp()
WHERE event_id=$1 AND user_id=$2 AND revision=$5 RETURNING *;

-- name: RevokeSpaceAccess :one
UPDATE space_access SET revoked_at=clock_timestamp(),revision=revision+1,updated_by=$4,updated_at=clock_timestamp()
WHERE space_id=$1 AND id=$2 AND revision=$3 AND revoked_at IS NULL RETURNING *;

-- name: RevokeEventAccess :one
UPDATE event_access SET revoked_at=clock_timestamp(),revision=revision+1,updated_by=$4,updated_at=clock_timestamp()
WHERE event_id=$1 AND id=$2 AND revision=$3 AND revoked_at IS NULL RETURNING *;

-- name: LockSpaceAccessForActor :many
SELECT id FROM space_access WHERE space_id=$1 AND user_id=$2 FOR SHARE;

-- name: LockEventAccessForActor :many
SELECT id FROM event_access WHERE event_id=$1 AND user_id=$2 FOR SHARE;

-- name: EventSpaceForAccess :one
SELECT space_id FROM events WHERE id=$1 AND deleted_at IS NULL;

-- name: ListInheritedEventAccess :many
SELECT a.id,a.user_id,u.display_name,u.email,a.role,a.revision,a.revoked_at,a.updated_at
FROM events e JOIN space_access a ON a.space_id=e.space_id JOIN users u ON u.id=a.user_id
WHERE e.id=$1 AND a.revoked_at IS NULL AND u.deleted_at IS NULL
AND NOT EXISTS(SELECT 1 FROM event_access x WHERE x.event_id=e.id AND x.user_id=a.user_id AND x.revoked_at IS NULL)
ORDER BY a.created_at,a.id LIMIT $2 OFFSET $3;
