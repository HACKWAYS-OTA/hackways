-- name: AdminGlobalOverview :one
SELECT
(SELECT count(*) FROM users WHERE deleted_at IS NULL) AS accounts,
(SELECT count(*) FROM platform_administrators a JOIN users u ON u.id=a.user_id WHERE a.revoked_at IS NULL AND u.deleted_at IS NULL AND u.email_verified) AS administrators,
(SELECT count(*) FROM email_outbox WHERE status='pending') AS pending_emails,
(SELECT count(*) FROM email_outbox WHERE status='failed') AS failed_emails,
(SELECT count(*) FROM sessions WHERE expires_at>clock_timestamp()) AS active_sessions;
