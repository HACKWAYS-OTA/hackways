-- Never discard live recovery challenges or recognized-device history.
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM password_reset_challenges) OR EXISTS(SELECT 1 FROM recognized_devices)
 OR EXISTS(SELECT 1 FROM users WHERE device_baseline_at IS NOT NULL)
 THEN RAISE EXCEPTION 'Auth recovery/device records exist; retain schema 14 and roll back runtime only'; END IF;
END $$;
DROP TABLE recognized_devices;
DROP TABLE password_reset_challenges;
-- Retain additive columns for backward-compatible runtime rollback.
