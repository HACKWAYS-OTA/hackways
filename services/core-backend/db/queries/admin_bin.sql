-- name: AdminListEventBin :many
SELECT id,admin_reference AS reference,title,slug,status::text AS status,revision,deleted_at
FROM events
WHERE deleted_at IS NOT NULL AND (sqlc.arg(q)::text='' OR title ILIKE '%'||sqlc.arg(q)||'%' OR slug ILIKE '%'||sqlc.arg(q)||'%')
ORDER BY deleted_at DESC,id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: LockAdminBinEvent :one
SELECT id,admin_reference AS reference,title,status,revision,deleted_at,ends_at,clock_timestamp()::timestamptz AS database_now
FROM events WHERE id=$1 FOR UPDATE;

-- name: MoveAdminEventToBin :one
UPDATE events SET deleted_at=clock_timestamp(),revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND deleted_at IS NULL
RETURNING id,admin_reference AS reference,title,status::text AS status,revision,deleted_at;

-- name: RestoreAdminBinEvent :one
UPDATE events SET deleted_at=NULL,
status=CASE WHEN status='published' THEN 'unpublished'::event_status ELSE status END,
revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND deleted_at IS NOT NULL
RETURNING id,admin_reference AS reference,title,status::text AS status,revision,deleted_at;
