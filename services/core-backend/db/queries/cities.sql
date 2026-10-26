-- name: ListPublicCities :many
SELECT c.id,c.slug,c.name,c.region,c.country,c.timezone,c.description,c.banner_url,c.updated_at,
 (SELECT count(*) FROM city_hub_events ce JOIN events e ON e.id=ce.event_id WHERE ce.city_id=c.id AND e.visibility='public' AND e.status='published' AND e.deleted_at IS NULL AND e.ends_at>now())::bigint AS event_count
FROM city_hubs c WHERE c.status='published' ORDER BY c.name,c.id LIMIT $1 OFFSET $2;

-- name: ReadPublicCity :one
SELECT id,slug,name,region,country,timezone,description,banner_url,updated_at FROM city_hubs WHERE slug=$1 AND status='published';

-- name: ListCityEvents :many
SELECT e.* FROM events e JOIN city_hub_events ce ON ce.event_id=e.id JOIN city_hubs c ON c.id=ce.city_id
WHERE c.id=$1 AND c.status='published' AND e.visibility='public' AND e.deleted_at IS NULL
AND (e.status='published' OR (e.status='cancelled' AND e.published_at IS NOT NULL))
AND ((sqlc.arg(past)::boolean AND e.ends_at<=now()) OR (NOT sqlc.arg(past)::boolean AND e.status='published' AND e.ends_at>now()))
ORDER BY CASE WHEN sqlc.arg(past)::boolean THEN e.ends_at END DESC,e.starts_at,e.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: AdminListCities :many
SELECT id,reference,slug,name,region,country,timezone,description,banner_url,status,revision,created_at,updated_at FROM city_hubs
WHERE (sqlc.arg(q)::text='' OR name ILIKE '%' || sqlc.arg(q)::text || '%') ORDER BY name,id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ReadCityAdmin :one
SELECT * FROM city_hubs WHERE reference=$1;

-- name: LockCity :one
SELECT * FROM city_hubs WHERE reference=$1 FOR UPDATE;

-- name: CreateCity :one
INSERT INTO city_hubs(slug,name,region,country,timezone,description,status,created_by,creation_key,creation_hash)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(created_by,creation_key) DO UPDATE SET creation_key=EXCLUDED.creation_key RETURNING *;

-- name: UpdateCity :one
UPDATE city_hubs SET name=$2,region=$3,country=$4,timezone=$5,description=$6,status=$7,revision=revision+1,updated_at=clock_timestamp()
WHERE reference=$1 AND revision=$8 RETURNING *;

-- name: BumpCityRevision :exec
UPDATE city_hubs SET revision=revision+1,updated_at=clock_timestamp() WHERE id=$1;

-- name: AddCityEvent :exec
INSERT INTO city_hub_events(city_id,event_id,added_by) SELECT $1,e.id,$3 FROM events e WHERE e.id=$2 AND e.visibility='public' AND e.status='published' AND e.deleted_at IS NULL ON CONFLICT DO NOTHING;

-- name: RemoveCityEvent :exec
DELETE FROM city_hub_events WHERE city_id=$1 AND event_id=$2;

-- name: CityEventEligible :one
SELECT EXISTS(SELECT 1 FROM events WHERE id=$1 AND visibility='public' AND status='published' AND deleted_at IS NULL);

-- name: AdminCityEvents :many
SELECT e.id,e.slug,e.title,e.starts_at,e.status,e.visibility FROM events e JOIN city_hub_events ce ON ce.event_id=e.id WHERE ce.city_id=$1 ORDER BY e.starts_at DESC,e.id LIMIT $2 OFFSET $3;

-- name: CityEventChoices :many
SELECT id,title,slug,starts_at,location_name FROM events WHERE visibility='public' AND status='published' AND deleted_at IS NULL AND (sqlc.arg(q)::text='' OR title ILIKE '%' || sqlc.arg(q)::text || '%') ORDER BY starts_at DESC,id LIMIT 30;
