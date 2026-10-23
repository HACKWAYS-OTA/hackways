-- name: AnalyticsRegistrationSummary :one
SELECT count(*)::bigint AS registrations,
 count(*) FILTER(WHERE r.status='confirmed')::bigint AS confirmed,
 count(*) FILTER(WHERE r.status='pending_approval')::bigint AS pending,
 count(*) FILTER(WHERE r.status IN ('waitlisted','offered'))::bigint AS waiting,
 count(*) FILTER(WHERE r.status IN ('cancelled','rejected','expired','transferred'))::bigint AS other
FROM rsvps r JOIN events e ON e.id=r.event_id WHERE e.deleted_at IS NULL AND r.registered_at>=sqlc.arg(period_start) AND r.registered_at<sqlc.arg(period_end);

-- name: AnalyticsEventSummary :one
SELECT
 (SELECT count(*) FROM events e WHERE e.deleted_at IS NULL AND e.created_at>=sqlc.arg(period_start) AND e.created_at<sqlc.arg(period_end)) AS events_created,
 (SELECT count(*) FROM tickets t WHERE t.checked_in_at>=sqlc.arg(period_start) AND t.checked_in_at<sqlc.arg(period_end)) AS check_ins;

-- name: AnalyticsRegistrationDaily :many
WITH days AS (SELECT generate_series(sqlc.arg(period_start)::timestamptz AT TIME ZONE 'UTC',(sqlc.arg(period_end)::timestamptz AT TIME ZONE 'UTC')-INTERVAL '1 day',INTERVAL '1 day') AS date),
 registrations AS (SELECT (r.registered_at AT TIME ZONE 'UTC')::date AS day,count(*) AS registrations,count(*) FILTER(WHERE r.status='confirmed') AS confirmed
 FROM rsvps r JOIN events e ON e.id=r.event_id WHERE e.deleted_at IS NULL AND r.registered_at>=sqlc.arg(period_start) AND r.registered_at<sqlc.arg(period_end) GROUP BY day),
 arrivals AS (SELECT (checked_in_at AT TIME ZONE 'UTC')::date AS day,count(*) AS check_ins FROM tickets t WHERE t.checked_in_at>=sqlc.arg(period_start) AND t.checked_in_at<sqlc.arg(period_end) GROUP BY day)
SELECT to_char(days.date,'YYYY-MM-DD')::text AS day,COALESCE(r.registrations,0)::bigint AS registrations,COALESCE(r.confirmed,0)::bigint AS confirmed,COALESCE(a.check_ins,0)::bigint AS check_ins
FROM days LEFT JOIN registrations r ON r.day=(days.date)::date LEFT JOIN arrivals a ON a.day=(days.date)::date ORDER BY days.date;

-- name: AnalyticsTopEvents :many
SELECT e.admin_reference AS reference,e.title,e.slug,count(*)::bigint AS registrations,
 count(*) FILTER(WHERE r.status='confirmed')::bigint AS confirmed
FROM rsvps r JOIN events e ON e.id=r.event_id WHERE e.deleted_at IS NULL AND r.registered_at>=sqlc.arg(period_start) AND r.registered_at<sqlc.arg(period_end)
GROUP BY e.id,e.admin_reference,e.title,e.slug ORDER BY registrations DESC,e.id LIMIT 10;
