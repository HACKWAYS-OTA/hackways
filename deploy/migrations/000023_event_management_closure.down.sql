DO $$ BEGIN IF EXISTS(SELECT 1 FROM events WHERE management_closed) THEN RAISE EXCEPTION 'Administrator closures must be resolved before removing their protection'; END IF; END $$;
ALTER TABLE events DROP COLUMN management_closed;
