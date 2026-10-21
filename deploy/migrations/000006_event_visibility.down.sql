DO $$ BEGIN RAISE EXCEPTION 'Keep link-only events private. Rollback requires a reviewed compatible visibility transition.'; END $$;
