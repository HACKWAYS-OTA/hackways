-- name: AdminCoreOverview :one
SELECT
(SELECT count(*) FROM events WHERE deleted_at IS NULL AND status='published' AND starts_at>clock_timestamp()) AS upcoming_events,
(SELECT count(*) FROM events WHERE deleted_at IS NULL AND status IN ('draft','unpublished')) AS moderation_queue,
(SELECT count(*) FROM rsvps r JOIN events e ON e.id=r.event_id WHERE r.status='pending_approval' AND e.deleted_at IS NULL AND e.starts_at>clock_timestamp()) AS pending_approvals,
(SELECT count(*) FROM rsvps r JOIN events e ON e.id=r.event_id WHERE r.status='waitlisted' AND e.deleted_at IS NULL AND e.starts_at>clock_timestamp()) AS waiting_list,
(SELECT count(*) FROM discover_promotions p JOIN events e ON e.id=p.event_id WHERE p.starts_at<=clock_timestamp() AND p.ends_at>clock_timestamp() AND e.status='published' AND e.visibility='public' AND e.deleted_at IS NULL AND e.starts_at>clock_timestamp()) AS active_promotions;
