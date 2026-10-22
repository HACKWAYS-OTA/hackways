DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM email_campaigns) OR EXISTS(SELECT 1 FROM promotion_contacts)
 THEN RAISE EXCEPTION 'Campaign/consent history exists; roll back runtime while retaining schema 15'; END IF;
END $$;
DROP TABLE email_campaign_deliveries;
DROP TABLE email_campaigns;
DROP TABLE promotion_contacts;
