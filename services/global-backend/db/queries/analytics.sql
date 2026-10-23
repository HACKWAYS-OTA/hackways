-- name: RecordAnalyticsBatch :one
WITH budget AS (
 INSERT INTO analytics_ingest_days(day,reserved_events) VALUES ((clock_timestamp() AT TIME ZONE 'UTC')::date,sqlc.arg(batch_size)::int)
 ON CONFLICT(day) DO UPDATE SET reserved_events=analytics_ingest_days.reserved_events+EXCLUDED.reserved_events
 WHERE analytics_ingest_days.reserved_events+EXCLUDED.reserved_events<=25000 RETURNING day
), inserted AS (
 INSERT INTO analytics_events(id,visitor_hash,session_hash,event_name,route,device,referrer_host,step)
 SELECT item.id,sqlc.arg(visitor_hash)::text,sqlc.arg(session_hash)::text,item.event_name,item.route,item.device,item.referrer_host,item.step
 FROM jsonb_to_recordset(sqlc.arg(events)::jsonb) AS item(id uuid,event_name text,route text,device text,referrer_host text,step smallint)
 WHERE EXISTS(SELECT 1 FROM budget) ON CONFLICT(id) DO NOTHING RETURNING id
) SELECT count(*) AS recorded,EXISTS(SELECT 1 FROM budget) AS within_budget FROM inserted;

-- name: CleanupAnalyticsEvents :exec
DELETE FROM analytics_events WHERE id IN(SELECT id FROM analytics_events WHERE created_at<clock_timestamp()-INTERVAL '90 days' ORDER BY created_at LIMIT 500);

-- name: CleanupAnalyticsBudget :exec
DELETE FROM analytics_ingest_days WHERE day<(clock_timestamp() AT TIME ZONE 'UTC')::date-90;

-- name: AnalyticsTrafficSummary :one
SELECT count(*) FILTER(WHERE event_name='page_view')::bigint AS page_views,
 count(DISTINCT visitor_hash) FILTER(WHERE event_name='page_view')::bigint AS visitors,
 count(DISTINCT session_hash) FILTER(WHERE event_name='page_view')::bigint AS sessions,
 count(*) FILTER(WHERE event_name='registration_complete')::bigint AS tracked_registrations
FROM analytics_events WHERE created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end);

-- name: AnalyticsTrafficDaily :many
SELECT to_char(date_trunc('day',created_at AT TIME ZONE 'UTC'),'YYYY-MM-DD')::text AS day,
 count(*) FILTER(WHERE event_name='page_view')::bigint AS page_views,
 count(DISTINCT visitor_hash) FILTER(WHERE event_name='page_view')::bigint AS visitors,
 count(*) FILTER(WHERE event_name='registration_complete')::bigint AS tracked_registrations
FROM analytics_events WHERE created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end)
GROUP BY day ORDER BY day;

-- name: AnalyticsTopPages :many
SELECT route,count(*)::bigint AS views,count(DISTINCT visitor_hash)::bigint AS visitors
FROM analytics_events WHERE event_name='page_view' AND created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end)
GROUP BY route ORDER BY views DESC,route LIMIT 10;

-- name: AnalyticsAcquisition :many
SELECT CASE WHEN referrer_host='' THEN 'Direct / internal' ELSE referrer_host END::text AS source,count(*)::bigint AS views
FROM analytics_events WHERE event_name='page_view' AND created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end)
GROUP BY source ORDER BY views DESC,source LIMIT 10;

-- name: AnalyticsDevices :many
SELECT device,count(*)::bigint AS views FROM analytics_events WHERE event_name='page_view'
AND created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end) GROUP BY device ORDER BY views DESC,device;

-- name: AnalyticsAccountSummary :one
SELECT
 (SELECT count(*) FROM users u WHERE u.deleted_at IS NULL AND u.created_at>=sqlc.arg(period_start) AND u.created_at<sqlc.arg(period_end)) AS new_accounts,
 (SELECT count(*) FROM users u WHERE u.deleted_at IS NULL AND u.email_verified AND u.created_at>=sqlc.arg(period_start) AND u.created_at<sqlc.arg(period_end)) AS verified_new_accounts,
 (SELECT count(*) FROM email_outbox o WHERE o.sent_at>=sqlc.arg(period_start) AND o.sent_at<sqlc.arg(period_end)) AS accepted_emails,
 (SELECT count(*) FROM email_outbox o WHERE o.status='failed' AND o.created_at>=sqlc.arg(period_start) AND o.created_at<sqlc.arg(period_end)) AS failed_emails,
 (SELECT count(*) FROM promotion_contacts WHERE subscribed) AS subscribed_contacts;

-- name: AnalyticsAccountDaily :many
WITH days AS (SELECT generate_series(sqlc.arg(period_start)::timestamptz AT TIME ZONE 'UTC',(sqlc.arg(period_end)::timestamptz AT TIME ZONE 'UTC')-INTERVAL '1 day',INTERVAL '1 day') AS date),
 accounts AS (SELECT (created_at AT TIME ZONE 'UTC')::date AS day,count(*) AS new_accounts FROM users u WHERE u.deleted_at IS NULL AND u.created_at>=sqlc.arg(period_start) AND created_at<sqlc.arg(period_end) GROUP BY day),
 messages AS (SELECT (sent_at AT TIME ZONE 'UTC')::date AS day,count(*) AS accepted_emails FROM email_outbox o WHERE o.sent_at>=sqlc.arg(period_start) AND sent_at<sqlc.arg(period_end) GROUP BY day)
SELECT to_char(days.date,'YYYY-MM-DD')::text AS day,COALESCE(a.new_accounts,0)::bigint AS new_accounts,COALESCE(m.accepted_emails,0)::bigint AS accepted_emails
FROM days LEFT JOIN accounts a ON a.day=(days.date)::date LEFT JOIN messages m ON m.day=(days.date)::date ORDER BY days.date;

-- name: AnalyticsCollectionState :one
SELECT (SELECT COALESCE(to_char(min(created_at) AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'')::text FROM analytics_events) AS first_event,
 COALESCE((SELECT reserved_events FROM analytics_ingest_days WHERE day=(clock_timestamp() AT TIME ZONE 'UTC')::date),0)::int AS reserved_today;
