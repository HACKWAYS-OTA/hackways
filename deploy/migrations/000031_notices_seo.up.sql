CREATE TABLE site_notices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{0,119}$'),
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 180),
    summary TEXT NOT NULL DEFAULT '' CHECK (char_length(summary) <= 500),
    body TEXT NOT NULL DEFAULT '' CHECK (char_length(body) <= 100000),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    published_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    updated_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CHECK (status <> 'published' OR (published_at IS NOT NULL AND char_length(body) > 0 AND char_length(summary) > 0))
);
CREATE INDEX site_notices_public_idx ON site_notices(published_at DESC,id) WHERE status='published';

CREATE TABLE notice_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notice_id UUID NOT NULL REFERENCES site_notices(id),
    blob_key TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    width INTEGER NOT NULL CHECK (width BETWEEN 1 AND 20000),
    height INTEGER NOT NULL CHECK (height BETWEEN 1 AND 20000),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX notice_images_notice_idx ON notice_images(notice_id,created_at);

CREATE TABLE site_seo_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE CHECK (char_length(path) BETWEEN 1 AND 240 AND left(path,1)='/'),
    title TEXT NOT NULL DEFAULT '' CHECK (char_length(title) <= 90),
    description TEXT NOT NULL DEFAULT '' CHECK (char_length(description) <= 320),
    canonical_path TEXT NOT NULL DEFAULT '' CHECK (char_length(canonical_path) <= 240),
    noindex BOOLEAN NOT NULL DEFAULT FALSE,
    planning_keywords TEXT[] NOT NULL DEFAULT '{}' CHECK (cardinality(planning_keywords) <= 2000 AND char_length(array_to_string(planning_keywords, ' ')) <= 160000),
    notes TEXT NOT NULL DEFAULT '' CHECK (char_length(notes) <= 4000),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_by UUID NOT NULL REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Public metadata must not disclose editorial fields for unpublished entities.
-- Route shape is separately validated by Go; this projection only checks publication.
CREATE FUNCTION site_seo_target_public(target_path TEXT) RETURNS BOOLEAN
LANGUAGE SQL STABLE AS $$
 SELECT CASE split_part(target_path,'/',2)
 WHEN 'notice' THEN EXISTS(SELECT 1 FROM site_notices WHERE slug=split_part(target_path,'/',3) AND status='published')
 WHEN 'events' THEN EXISTS(SELECT 1 FROM events WHERE slug=split_part(target_path,'/',3) AND status='published' AND visibility='public' AND deleted_at IS NULL)
 WHEN 'spaces' THEN split_part(target_path,'/',3)='' OR EXISTS(SELECT 1 FROM spaces WHERE slug=split_part(target_path,'/',3) AND status='active')
 WHEN 'cities' THEN split_part(target_path,'/',3)='' OR EXISTS(SELECT 1 FROM city_hubs WHERE slug=split_part(target_path,'/',3) AND status='published')
 WHEN 'programs' THEN split_part(target_path,'/',3)='' OR EXISTS(SELECT 1 FROM programs WHERE slug=split_part(target_path,'/',3) AND status IN ('open','closed'))
 ELSE TRUE END;
$$;
