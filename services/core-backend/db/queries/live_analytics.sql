-- name: LiveAnalyticsWindow :one
SELECT transaction_timestamp()::timestamptz AS period_end,
 (transaction_timestamp()-make_interval(secs=>sqlc.arg(window_seconds)::double precision))::timestamptz AS period_start;

-- name: LiveRegistrationActivity :many
WITH buckets AS (
 SELECT generate_series(sqlc.arg(period_start)::timestamptz,sqlc.arg(period_end)::timestamptz-INTERVAL '1 millisecond',sqlc.arg(bucket_seconds)::int*INTERVAL '1 second') AS bucket
), registrations AS (
 SELECT date_bin(sqlc.arg(bucket_seconds)::int*INTERVAL '1 second',r.registered_at,sqlc.arg(period_start)::timestamptz) AS bucket,count(*)::bigint AS registrations
 FROM rsvps r JOIN events e ON e.id=r.event_id WHERE e.deleted_at IS NULL AND r.registered_at>=sqlc.arg(period_start) AND r.registered_at<sqlc.arg(period_end) GROUP BY bucket
), arrivals AS (
 SELECT date_bin(sqlc.arg(bucket_seconds)::int*INTERVAL '1 second',checked_in_at,sqlc.arg(period_start)::timestamptz) AS bucket,count(*)::bigint AS check_ins
 FROM tickets WHERE checked_in_at>=sqlc.arg(period_start) AND checked_in_at<sqlc.arg(period_end) GROUP BY bucket
)
SELECT b.bucket::timestamptz AS bucket,COALESCE(r.registrations,0)::bigint AS registrations,COALESCE(a.check_ins,0)::bigint AS check_ins
FROM buckets b LEFT JOIN registrations r USING(bucket) LEFT JOIN arrivals a USING(bucket) ORDER BY b.bucket;
