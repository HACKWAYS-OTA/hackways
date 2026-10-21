-- Forward-only production rollback: deploy the preceding application image.
-- This migration adds persisted tickets, audit and idempotency data. Do not
-- drop those records or remove enum values to roll back an application release.
-- A data-preserving schema removal requires a separately reviewed migration.
DO $$ BEGIN RAISE EXCEPTION 'Forward-only migration: roll back the application image; do not delete persisted workflow data'; END $$;
