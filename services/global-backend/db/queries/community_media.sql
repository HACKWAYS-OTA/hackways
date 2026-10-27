-- name: ReadOwnedUploadSpace :one
SELECT s.id FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.id=$1 AND (o.user_id=$2 OR EXISTS(SELECT 1 FROM space_access a WHERE a.space_id=s.id AND a.user_id=$2 AND a.revoked_at IS NULL AND a.role IN ('manager','editor')));

-- name: AttachSpaceImage :one
UPDATE spaces s SET
 logo_url=CASE WHEN sqlc.arg(kind)::text='logo' THEN sqlc.narg(image_url)::text ELSE s.logo_url END,
 banner_url=CASE WHEN sqlc.arg(kind)::text='banner' THEN sqlc.narg(image_url)::text ELSE s.banner_url END,
 revision=s.revision+1,updated_at=clock_timestamp()
FROM organizers o WHERE s.organizer_id=o.id AND s.id=sqlc.arg(id) AND (o.user_id=sqlc.arg(user_id) OR EXISTS(SELECT 1 FROM space_access a WHERE a.space_id=s.id AND a.user_id=sqlc.arg(user_id) AND a.revoked_at IS NULL AND a.role IN ('manager','editor'))) AND s.revision=sqlc.arg(revision)
RETURNING s.id,s.slug,s.revision,s.logo_url,s.banner_url;

-- name: ReadSpaceImageState :one
SELECT s.id,s.slug,s.revision,s.logo_url,s.banner_url FROM spaces s JOIN organizers o ON o.id=s.organizer_id WHERE s.id=$1 AND (o.user_id=$2 OR EXISTS(SELECT 1 FROM space_access a WHERE a.space_id=s.id AND a.user_id=$2 AND a.revoked_at IS NULL AND a.role IN ('manager','editor')));

-- name: ReadCityUpload :one
SELECT id,reference,revision,banner_url FROM city_hubs WHERE id=$1;

-- name: AttachCityBanner :one
UPDATE city_hubs SET banner_url=sqlc.narg(image_url)::text,revision=revision+1,updated_at=clock_timestamp()
WHERE id=sqlc.arg(id) AND revision=sqlc.arg(revision) RETURNING id,reference,revision,banner_url;

-- name: AuditCommunityMedia :exec
INSERT INTO audit_log(actor_user_id,entity_id,action,request_id) VALUES($1,$2,$3,$4);

-- name: LockEventMediaResource :one
SELECT space_id,banner_url,revision FROM events WHERE id=$1 AND deleted_at IS NULL AND NOT management_closed FOR UPDATE;
-- name: LockSpaceMediaResource :one
SELECT id FROM spaces WHERE id=$1 FOR UPDATE;
-- name: LockMediaSpaceGrant :many
SELECT id FROM space_access WHERE space_id=$1 AND user_id=$2 FOR SHARE;
-- name: LockMediaEventGrant :many
SELECT id FROM event_access WHERE event_id=$1 AND user_id=$2 FOR SHARE;

-- name: AuditEventMedia :exec
INSERT INTO audit_log(actor_user_id,event_id,entity_id,action,request_id) VALUES($1,$2,$2,'event.banner_updated',$3);
