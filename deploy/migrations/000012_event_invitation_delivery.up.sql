-- An invitation is a durable outbox message, separate from a registration.
-- One invitation per event/email prevents duplicate bulk sends and safe retries.
CREATE UNIQUE INDEX email_outbox_event_invitation_recipient_idx
ON email_outbox(event_id, lower(recipient_email)) WHERE template_type = 'event_invitation';
CREATE INDEX email_outbox_event_invitation_timeline_idx
ON email_outbox(event_id, created_at DESC, id) WHERE template_type = 'event_invitation';
