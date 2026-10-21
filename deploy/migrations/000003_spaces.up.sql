CREATE TABLE spaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES organizers(id) ON DELETE RESTRICT,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9-]+$'),
    description TEXT NOT NULL CHECK (char_length(description) BETWEEN 1 AND 6000),
    guidelines TEXT NOT NULL DEFAULT '' CHECK (char_length(guidelines) <= 10000),
    website_url TEXT NOT NULL DEFAULT '' CHECK (char_length(website_url) <= 2048),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
    creation_key TEXT NOT NULL CHECK (char_length(creation_key) BETWEEN 16 AND 128),
    creation_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, organizer_id),
    UNIQUE (organizer_id, creation_key)
);
CREATE INDEX spaces_active_created_idx ON spaces(created_at DESC,id) WHERE status='active';
CREATE INDEX spaces_organizer_idx ON spaces(organizer_id,created_at DESC,id);
CREATE INDEX spaces_search_idx ON spaces USING GIN(to_tsvector('simple',name || ' ' || description)) WHERE status='active';

CREATE TABLE space_memberships (
    space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(space_id,user_id)
);
CREATE INDEX space_memberships_user_idx ON space_memberships(user_id,space_id);

ALTER TABLE events ADD COLUMN space_id UUID;
ALTER TABLE events ADD CONSTRAINT events_space_owner_fk FOREIGN KEY(space_id,organizer_id) REFERENCES spaces(id,organizer_id) ON DELETE RESTRICT;
CREATE INDEX events_space_feed_idx ON events(space_id,starts_at,id) WHERE space_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON TABLE spaces IS 'Owner-led communities. Membership does not grant event authoring or management rights.';

ALTER TABLE rsvps ADD COLUMN public_code TEXT NOT NULL DEFAULT ('R-' || encode(public.gen_random_bytes(8),'hex'));
ALTER TABLE rsvps ADD CONSTRAINT rsvps_public_code_unique UNIQUE(public_code);
ALTER TABLE tickets ADD COLUMN public_code TEXT NOT NULL DEFAULT ('T-' || encode(public.gen_random_bytes(12),'hex'));
ALTER TABLE tickets ADD CONSTRAINT tickets_public_code_unique UNIQUE(public_code);

ALTER TABLE events ADD COLUMN theme TEXT NOT NULL DEFAULT 'minimal' CHECK(theme IN ('minimal','aurora','ambient','paper','lilac','confetti','pattern','playful','seasonal')),
 ADD COLUMN theme_color TEXT NOT NULL DEFAULT '#ddd4e9' CHECK(theme_color ~ '^#[0-9a-fA-F]{6}$'),
 ADD COLUMN theme_font TEXT NOT NULL DEFAULT 'futurist' CHECK(theme_font IN ('futurist','google')),
 ADD COLUMN theme_style TEXT NOT NULL DEFAULT 'soft' CHECK(theme_style IN ('soft','crisp'));
