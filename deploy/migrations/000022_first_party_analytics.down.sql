DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM analytics_events) OR EXISTS(SELECT 1 FROM analytics_ingest_days)
 THEN RAISE EXCEPTION 'Analytics records must be explicitly archived before removing analytics tables'; END IF;
END $$;
DROP TABLE analytics_events;
DROP TABLE analytics_ingest_days;
DROP INDEX users_analytics_created_idx;
DROP INDEX events_analytics_created_idx;
DROP INDEX rsvps_analytics_registered_idx;
DROP INDEX tickets_analytics_checkin_idx;
DROP INDEX email_outbox_analytics_sent_idx;
DROP INDEX email_outbox_analytics_failed_idx;
