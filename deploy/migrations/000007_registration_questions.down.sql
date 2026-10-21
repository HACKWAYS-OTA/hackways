DO $$ BEGIN RAISE EXCEPTION 'Preserve submitted answers. Rollback requires a reviewed compatible registration-form transition.'; END $$;
