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
SELECT recipient_email, status, created_at, sent_at, retry_count
FROM email_outbox WHERE event_id=$1 AND template_type='event_invitation'
ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountEventInvitations :one
SELECT count(*) FROM email_outbox WHERE event_id=$1 AND template_type='event_invitation';

-- name: CountNewEventInvitationRecipients :one
SELECT count(*) FROM unnest(sqlc.arg(emails)::text[]) AS candidate(email)
WHERE NOT EXISTS (SELECT 1 FROM email_outbox o WHERE o.event_id=sqlc.arg(event_id) AND o.template_type='event_invitation' AND lower(o.recipient_email)=candidate.email)
AND NOT EXISTS (SELECT 1 FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.event_id=sqlc.arg(event_id)
AND lower(COALESCE(u.email,r.guest_email))=candidate.email AND r.status IN ('confirmed','pending_approval','waitlisted','offered'));
