-- name: DiscoverPromotionSettings :one
SELECT revision, updated_at FROM discover_promotion_settings WHERE singleton;

-- name: LockDiscoverPromotionSettings :one
SELECT *,clock_timestamp()::timestamptz AS database_now FROM discover_promotion_settings WHERE singleton FOR UPDATE;

-- name: LockPromotionActor :one
SELECT id FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE;

-- name: AdminDiscoverPromotions :many
SELECT p.slot,e.admin_reference AS reference,e.title,e.slug,e.starts_at AS event_starts_at,e.timezone,
e.status::text AS status,e.visibility,p.starts_at,p.ends_at,
(e.status='published' AND e.visibility='public' AND e.deleted_at IS NULL AND e.starts_at>clock_timestamp() AND p.ends_at>clock_timestamp())::boolean AS eligible
FROM discover_promotions p JOIN events e ON e.id=p.event_id ORDER BY p.slot;

-- name: PromotionEventChoices :many
SELECT admin_reference AS reference,title,slug,starts_at AS event_starts_at,timezone
FROM events WHERE status='published' AND visibility='public' AND deleted_at IS NULL AND starts_at>clock_timestamp()
AND (sqlc.arg(q)::text='' OR title ILIKE '%'||sqlc.arg(q)||'%' OR slug ILIKE '%'||sqlc.arg(q)||'%')
ORDER BY starts_at,id LIMIT 21 OFFSET sqlc.arg(page_offset)::int;

-- name: LockPromotionEvent :one
SELECT * FROM events WHERE admin_reference=$1 AND deleted_at IS NULL FOR UPDATE;

-- name: ClearDiscoverPromotions :exec
DELETE FROM discover_promotions;

-- name: InsertDiscoverPromotion :exec
INSERT INTO discover_promotions(slot,event_id,starts_at,ends_at) VALUES ($1,$2,$3,$4);

-- name: UpdateDiscoverPromotionSettings :exec
UPDATE discover_promotion_settings SET revision=revision+1,updated_by=$1,updated_at=clock_timestamp() WHERE singleton;

-- name: ActiveDiscoverPromotions :many
SELECT sqlc.embed(e),p.slot,LEAST(p.ends_at,e.starts_at)::timestamptz AS visible_until
FROM discover_promotions p JOIN events e ON e.id=p.event_id
WHERE p.starts_at<=sqlc.arg(at)::timestamptz AND p.ends_at>sqlc.arg(at)::timestamptz
AND e.status='published' AND e.visibility='public' AND e.deleted_at IS NULL AND e.starts_at>sqlc.arg(at)::timestamptz
ORDER BY p.slot;

-- name: NextDiscoverPromotionChange :one
SELECT min(boundary)::timestamptz AS next_change_at FROM (
SELECT p.starts_at AS boundary FROM discover_promotions p JOIN events e ON e.id=p.event_id
WHERE e.status='published' AND e.visibility='public' AND e.deleted_at IS NULL AND p.starts_at<e.starts_at AND p.ends_at>sqlc.arg(at)::timestamptz
UNION ALL
SELECT LEAST(p.ends_at,e.starts_at) AS boundary FROM discover_promotions p JOIN events e ON e.id=p.event_id
WHERE e.status='published' AND e.visibility='public' AND e.deleted_at IS NULL
) boundaries WHERE boundary>sqlc.arg(at)::timestamptz;
