-- Bounded admin detail timelines and recipient delivery lookup.
CREATE INDEX audit_log_entity_idx ON audit_log(entity_id, id DESC);
CREATE INDEX audit_log_actor_idx ON audit_log(actor_user_id, id DESC);
CREATE INDEX email_outbox_recipient_timeline_idx ON email_outbox(lower(recipient_email), created_at DESC, id);
