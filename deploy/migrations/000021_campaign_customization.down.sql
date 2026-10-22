-- Runtime rollback is compatible; never discard saved configuration or message receipts.
DO $$ BEGIN
 RAISE EXCEPTION 'Migration 21 retains campaign settings and delivery history. Roll back application images while keeping this additive schema.';
END $$;
