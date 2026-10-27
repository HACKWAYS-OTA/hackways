-- name: PublicNotices :many
SELECT slug,title,summary,published_at,updated_at FROM site_notices
WHERE status='published' ORDER BY published_at DESC,id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: PublicNotice :one
SELECT slug,title,summary,body,document,body_format,published_at,updated_at FROM site_notices WHERE slug=$1 AND status='published';

-- name: AdminNotices :many
SELECT id,reference,slug,title,summary,status,revision,published_at,updated_at FROM site_notices WHERE (sqlc.arg(q)::text='' OR title ILIKE '%'||sqlc.arg(q)||'%' OR slug ILIKE '%'||sqlc.arg(q)||'%')
AND (sqlc.arg(state)::text='' OR status=sqlc.arg(state))
ORDER BY updated_at DESC,id LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminNotice :one
SELECT * FROM site_notices WHERE reference=$1;

-- name: AdminNoticeBySlug :one
SELECT * FROM site_notices WHERE slug=$1;

-- name: LockSiteUtility :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(lock_key)::text,0));

-- name: CreateNotice :one
INSERT INTO site_notices(slug,title,summary,body,status,published_at,created_by,updated_by,document,body_format)
VALUES($1,$2,$3,$4,$5,CASE WHEN $5='published' THEN clock_timestamp() END,$6,$6,sqlc.narg(document)::jsonb,sqlc.arg(body_format)) RETURNING *;

-- name: UpdateNotice :one
UPDATE site_notices SET slug=sqlc.arg(slug),title=sqlc.arg(title),summary=sqlc.arg(summary),body=sqlc.arg(body),status=sqlc.arg(status),document=sqlc.narg(document)::jsonb,body_format=sqlc.arg(body_format),
published_at=CASE WHEN sqlc.arg(status)='published' THEN coalesce(published_at,clock_timestamp()) ELSE published_at END,
revision=revision+1,updated_at=clock_timestamp(),updated_by=sqlc.arg(actor)
WHERE reference=sqlc.arg(reference) AND revision=sqlc.arg(revision) AND (slug=sqlc.arg(slug) OR published_at IS NULL) RETURNING *;

-- name: AdminSeoPages :many
SELECT path,title,description,noindex,revision,cardinality(planning_keywords)::int AS keyword_count FROM site_seo_pages WHERE sqlc.arg(q)::text='' OR path ILIKE '%'||sqlc.arg(q)||'%' OR title ILIKE '%'||sqlc.arg(q)||'%'
ORDER BY path LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: AdminSeoPage :one
SELECT * FROM site_seo_pages WHERE path=$1;

-- name: CreateSeoPage :one
INSERT INTO site_seo_pages(path,title,description,canonical_path,noindex,planning_keywords,notes,updated_by)
VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING *;

-- name: UpdateSeoPage :one
UPDATE site_seo_pages SET title=sqlc.arg(title),description=sqlc.arg(description),canonical_path=sqlc.arg(canonical_path),
noindex=sqlc.arg(noindex),planning_keywords=sqlc.arg(planning_keywords),notes=sqlc.arg(notes),
revision=revision+1,updated_by=sqlc.arg(actor),updated_at=clock_timestamp()
WHERE path=sqlc.arg(path) AND revision=sqlc.arg(revision) RETURNING *;

-- name: PublicSeoPage :one
SELECT path,title,description,
CASE WHEN canonical_path<>'' AND site_seo_target_public(canonical_path) THEN canonical_path ELSE '' END::text AS canonical_path,
noindex FROM site_seo_pages WHERE path=$1 AND site_seo_target_public(path);

-- name: PublicSeoRules :many
SELECT path,CASE WHEN canonical_path<>'' AND site_seo_target_public(canonical_path) THEN canonical_path ELSE '' END::text AS canonical_path,noindex
FROM site_seo_pages WHERE (noindex OR (canonical_path<>'' AND site_seo_target_public(canonical_path))) AND site_seo_target_public(path)
ORDER BY path LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: ReadNoticeUpload :one
SELECT id FROM site_notices WHERE id=$1;

-- name: RecordNoticeImage :one
INSERT INTO notice_images(notice_id,blob_key,url,created_by,width,height) VALUES($1,$2,$3,$4,$5,$6)
ON CONFLICT(blob_key) DO NOTHING RETURNING url;

-- name: ReadNoticeImage :one
SELECT url FROM notice_images WHERE notice_id=$1 AND blob_key=$2;

-- name: PublicNoticeImages :many
SELECT i.url,i.width,i.height FROM notice_images i JOIN site_notices n ON n.id=i.notice_id
WHERE n.slug=$1 AND n.status='published' AND position(i.url in n.body)>0 ORDER BY i.created_at,i.id LIMIT 100;

-- name: AdminSeoSuggestions :many
SELECT route,count(*)::bigint AS views FROM analytics_events
WHERE event_name='page_view' AND created_at >= transaction_timestamp()-INTERVAL '7 days'
GROUP BY route ORDER BY count(*) DESC,route LIMIT 30;

-- name: AdminSeoTopicLibrary :many
SELECT topic_key,theme,title,keywords FROM seo_topic_library ORDER BY theme,title LIMIT 151;
