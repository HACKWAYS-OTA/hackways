-- name: EnqueueEventInvitation :execrows
INSERT INTO email_outbox(event_id, recipient_email, template_type, payload_json)
SELECT sqlc.arg(event_id), sqlc.arg(email)::text, 'event_invitation', sqlc.arg(payload)::jsonb
WHERE NOT EXISTS (
 SELECT 1 FROM rsvps r LEFT JOIN users u ON u.id=r.user_id
 WHERE r.event_id=sqlc.arg(event_id) AND lower(COALESCE(u.email,r.guest_email))=lower(sqlc.arg(email)::text)
 AND r.status IN ('confirmed','pending_approval','waitlisted','offered')
)
ON CONFLICT (event_id, lower(recipient_email)) WHERE template_type='event_invitation' DO NOTHING;

-- name: ListEventInvitations :many
SELECT o.recipient_email,o.status,o.created_at,o.sent_at,o.retry_count,
 COALESCE(r.public_code,'')::text AS registration_code,COALESCE(r.status::text,'')::text AS registration_status,t.checked_in_at,
 COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS guest_name
FROM email_outbox o
LEFT JOIN users u ON lower(u.email)=lower(o.recipient_email) AND u.deleted_at IS NULL
LEFT JOIN LATERAL (SELECT rr.public_code,rr.status,rr.id,rr.guest_name FROM rsvps rr WHERE rr.event_id=o.event_id AND (rr.user_id=u.id OR rr.user_id IS NULL AND lower(rr.guest_email)=lower(o.recipient_email)) ORDER BY (rr.status IN ('confirmed','pending_approval','waitlisted','offered')) DESC,rr.registered_at DESC,rr.id LIMIT 1) r ON true
LEFT JOIN tickets t ON t.rsvp_id=r.id
WHERE o.event_id=sqlc.arg(event_id) AND o.template_type='event_invitation'
AND (sqlc.arg(search)::text='' OR lower(o.recipient_email) LIKE '%' || lower(sqlc.arg(search)::text) || '%' OR lower(COALESCE(NULLIF(r.guest_name,''),u.display_name,'')) LIKE '%' || lower(sqlc.arg(search)::text) || '%')
AND (sqlc.arg(state_filter)::text='all' OR sqlc.arg(state_filter)::text='awaiting' AND r.id IS NULL OR sqlc.arg(state_filter)::text='responded' AND r.status IN ('confirmed','pending_approval','waitlisted','offered') OR sqlc.arg(state_filter)::text IN ('sent','failed') AND o.status::text=sqlc.arg(state_filter)::text OR sqlc.arg(state_filter)::text='queued' AND o.status='pending')
ORDER BY o.created_at DESC,o.id LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: AccountEventInvitation :one
SELECT EXISTS(SELECT 1 FROM email_outbox WHERE event_id=$1 AND template_type='event_invitation' AND lower(recipient_email)=lower($2::text))::boolean AS invited;

-- name: EventInvitationSummary :one
SELECT count(*)::integer AS total,count(*) FILTER(WHERE status='sent')::integer AS sent,
 count(*) FILTER(WHERE status='pending')::integer AS queued,
 count(*) FILTER(WHERE status='failed')::integer AS failed,
 count(*) FILTER(WHERE EXISTS(SELECT 1 FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=o.event_id AND lower(COALESCE(u.email,r.guest_email))=lower(o.recipient_email) AND r.status IN ('confirmed','pending_approval','waitlisted','offered')))::integer AS responded,
 count(*) FILTER(WHERE NOT EXISTS(SELECT 1 FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=o.event_id AND lower(COALESCE(u.email,r.guest_email))=lower(o.recipient_email)))::integer AS awaiting
FROM email_outbox o WHERE o.event_id=$1 AND template_type='event_invitation';

-- name: CountEventInvitations :one
SELECT count(*) FROM email_outbox WHERE event_id=$1 AND template_type='event_invitation';

-- name: CountNewEventInvitationRecipients :one
SELECT count(*) FROM unnest(sqlc.arg(emails)::text[]) AS candidate(email)
WHERE NOT EXISTS (SELECT 1 FROM email_outbox o WHERE o.event_id=sqlc.arg(event_id) AND o.template_type='event_invitation' AND lower(o.recipient_email)=candidate.email)
AND NOT EXISTS (SELECT 1 FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=sqlc.arg(event_id)
AND lower(COALESCE(u.email,r.guest_email))=candidate.email AND r.status IN ('confirmed','pending_approval','waitlisted','offered'));
