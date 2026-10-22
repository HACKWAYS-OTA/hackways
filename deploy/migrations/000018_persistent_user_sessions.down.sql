-- Intentionally retain already-issued session expiry dates on application
-- rollback. Shortening them would sign users out contrary to the retention policy.
SELECT 1;
