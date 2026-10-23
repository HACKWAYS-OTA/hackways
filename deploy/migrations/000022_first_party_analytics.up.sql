CREATE TABLE analytics_ingest_days (
 day DATE PRIMARY KEY,
 reserved_events INTEGER NOT NULL CHECK(reserved_events BETWEEN 0 AND 25000)
);
CREATE TABLE analytics_events (
 id UUID PRIMARY KEY,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 visitor_hash TEXT NOT NULL CHECK(length(visitor_hash)=64),
 session_hash TEXT NOT NULL CHECK(length(session_hash)=64),
 event_name TEXT NOT NULL CHECK(event_name IN ('page_view','registration_step','registration_complete')),
 route TEXT NOT NULL CHECK(length(route) BETWEEN 1 AND 240),
 device TEXT NOT NULL CHECK(device IN ('desktop','tablet','mobile')),
 referrer_host TEXT NOT NULL DEFAULT '' CHECK(length(referrer_host)<=253),
 step SMALLINT CHECK(step BETWEEN 0 AND 50)
);
CREATE INDEX analytics_events_created_idx ON analytics_events(created_at);
CREATE INDEX analytics_events_route_created_idx ON analytics_events(route,created_at);
CREATE INDEX users_analytics_created_idx ON users(created_at) WHERE deleted_at IS NULL;
CREATE INDEX events_analytics_created_idx ON events(created_at) WHERE deleted_at IS NULL;
CREATE INDEX rsvps_analytics_registered_idx ON rsvps(registered_at);
CREATE INDEX tickets_analytics_checkin_idx ON tickets(checked_in_at) WHERE checked_in_at IS NOT NULL;
CREATE INDEX email_outbox_analytics_sent_idx ON email_outbox(sent_at) WHERE sent_at IS NOT NULL;
CREATE INDEX email_outbox_analytics_failed_idx ON email_outbox(created_at) WHERE status='failed';
