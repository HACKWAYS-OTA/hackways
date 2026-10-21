DO $$ BEGIN RAISE EXCEPTION 'Retain persisted event information and roll back the application image'; END $$;
