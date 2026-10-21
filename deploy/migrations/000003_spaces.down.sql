-- Community memberships and event associations are persisted user data.
-- Roll back application images while retaining this additive schema.
DO $$ BEGIN RAISE EXCEPTION 'Forward-only data-preserving rollback: retain Spaces and deploy the preceding application image'; END $$;
