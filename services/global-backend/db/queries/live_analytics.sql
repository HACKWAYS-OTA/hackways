-- name: LiveAnalyticsWindow :one
SELECT transaction_timestamp()::timestamptz AS period_end,
 (transaction_timestamp()-make_interval(secs=>sqlc.arg(window_seconds)::double precision))::timestamptz AS period_start;

-- name: LiveAnalyticsLimited :one
SELECT EXISTS(SELECT 1 FROM analytics_ingest_days
 WHERE day >= (sqlc.arg(period_start)::timestamptz AT TIME ZONE 'UTC')::date
 AND day <= (sqlc.arg(period_end)::timestamptz AT TIME ZONE 'UTC')::date
 AND reserved_events>=25000)::boolean AS limited;

-- name: LiveAccountSummary :one
SELECT
 (SELECT count(*) FROM users u WHERE u.deleted_at IS NULL AND u.created_at>=sqlc.arg(period_start) AND u.created_at<sqlc.arg(period_end)) AS new_accounts,
 (SELECT count(*) FROM email_outbox o WHERE o.sent_at>=sqlc.arg(period_start) AND o.sent_at<sqlc.arg(period_end)) AS accepted_emails,
 (SELECT count(*) FROM email_outbox o WHERE o.status='failed' AND o.created_at>=sqlc.arg(period_start) AND o.created_at<sqlc.arg(period_end)) AS failed_emails;
