CREATE TABLE promotion_contacts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 reference BIGSERIAL NOT NULL UNIQUE,
 email TEXT NOT NULL UNIQUE CHECK(email=lower(email) AND length(email)<=320),
 display_name TEXT NOT NULL DEFAULT '' CHECK(length(display_name)<=100),
 user_id UUID REFERENCES users(id),
 source TEXT NOT NULL CHECK(source IN ('account','import','test')),
 consent_source TEXT NOT NULL CHECK(length(consent_source) BETWEEN 1 AND 1000),
 consented_at TIMESTAMPTZ NOT NULL,
 subscribed BOOLEAN NOT NULL DEFAULT FALSE,
 unsubscribed_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX promotion_contacts_opted_in_idx ON promotion_contacts(reference) WHERE subscribed;
CREATE TABLE email_campaigns (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 reference BIGSERIAL NOT NULL UNIQUE,
 title TEXT NOT NULL CHECK(length(title) BETWEEN 1 AND 150),
 subject TEXT NOT NULL CHECK(length(subject) BETWEEN 1 AND 150),
 body TEXT NOT NULL CHECK(length(body) BETWEEN 1 AND 20000),
 action_label TEXT NOT NULL DEFAULT '' CHECK(length(action_label)<=80),
 action_url TEXT NOT NULL DEFAULT '' CHECK(length(action_url)<=2000),
 status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','queued','cancelled')),
 revision BIGINT NOT NULL DEFAULT 1,
 created_by UUID NOT NULL REFERENCES users(id),
 creation_key TEXT NOT NULL CHECK(length(creation_key) BETWEEN 16 AND 128),
 creation_hash TEXT NOT NULL CHECK(length(creation_hash)=64),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 queued_at TIMESTAMPTZ,
 cancelled_at TIMESTAMPTZ,
 UNIQUE(created_by,creation_key)
);
CREATE TABLE email_campaign_deliveries (
 outbox_id UUID PRIMARY KEY REFERENCES email_outbox(id),
 campaign_id UUID NOT NULL REFERENCES email_campaigns(id),
 contact_id UUID NOT NULL REFERENCES promotion_contacts(id),
 revision BIGINT NOT NULL,
 is_test BOOLEAN NOT NULL DEFAULT FALSE,
 unsubscribe_hash TEXT NOT NULL UNIQUE CHECK(length(unsubscribe_hash)=64),
 UNIQUE(campaign_id,contact_id,revision,is_test)
);
CREATE INDEX email_campaign_deliveries_campaign_idx ON email_campaign_deliveries(campaign_id,is_test);
