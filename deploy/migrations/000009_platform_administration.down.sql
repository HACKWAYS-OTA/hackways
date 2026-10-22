-- Preserve access history. Roll back the application without reversing this additive schema.
DO $$ BEGIN
  IF EXISTS (SELECT 1 FROM platform_administrators) OR EXISTS (SELECT 1 FROM platform_admin_sessions) THEN
    RAISE EXCEPTION 'Administration records exist; retain this additive schema during application rollback';
  END IF;
END $$;
DROP TABLE platform_admin_sessions;
DROP TABLE platform_administrators;
