ALTER TABLE events ADD COLUMN visibility TEXT NOT NULL DEFAULT 'public' CHECK(visibility IN ('public','private'));
CREATE INDEX idx_events_public_discovery ON events(starts_at,id) WHERE status='published' AND deleted_at IS NULL AND visibility='public';
