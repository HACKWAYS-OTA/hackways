ALTER TABLE email_campaigns ADD COLUMN configuration JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(configuration)='object' AND octet_length(configuration::text)<=65536);

CREATE TABLE admin_custom_messages (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), reference BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
 actor_user_id UUID NOT NULL REFERENCES users(id), creation_key TEXT NOT NULL, creation_hash TEXT NOT NULL,
 subject TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(), UNIQUE(actor_user_id,creation_key)
);
CREATE TABLE admin_custom_message_deliveries (
 message_id UUID NOT NULL REFERENCES admin_custom_messages(id), user_id UUID NOT NULL REFERENCES users(id),
 outbox_id UUID NOT NULL UNIQUE REFERENCES email_outbox(id), PRIMARY KEY(message_id,user_id)
);
