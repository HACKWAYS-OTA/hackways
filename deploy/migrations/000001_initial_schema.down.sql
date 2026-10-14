-- 000001_initial_schema.down.sql

DROP TABLE IF EXISTS email_outbox;
DROP TABLE IF EXISTS rsvps;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS event_categories;
DROP TABLE IF EXISTS organizers;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS registration_status;
DROP TYPE IF EXISTS event_format;
DROP TYPE IF EXISTS event_status;
