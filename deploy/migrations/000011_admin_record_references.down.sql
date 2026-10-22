-- Numeric references may already appear in saved links and audit history.
-- Application rollback retains these additive columns and their values.
DO $$ BEGIN
  RAISE EXCEPTION 'Admin references are permanent. Roll back application images while retaining the additive schema.';
END $$;
