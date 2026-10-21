-- name: SetWorkflowTimeout :exec
SET LOCAL statement_timeout = '2000ms';

-- name: LockWorkflowEvent :one
SELECT *, clock_timestamp()::timestamptz AS database_now FROM events WHERE id = $1 AND deleted_at IS NULL FOR UPDATE;

-- name: GetWorkflowRSVP :one
SELECT * FROM rsvps WHERE id = $1;

-- name: GetAccountEventRSVP :one
SELECT * FROM rsvps WHERE event_id = $1 AND user_id = $2;

-- name: GetGuestEventRSVP :one
SELECT * FROM rsvps WHERE event_id = $1 AND lower(guest_email) = lower($2) ORDER BY created_at LIMIT 1;

-- name: InsertWorkflowRSVP :one
INSERT INTO rsvps(event_id, user_id, guest_name, guest_email, notes, status, approval_expires_at)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: SetWorkflowRSVPState :one
UPDATE rsvps SET status = sqlc.arg(status)::registration_status, offer_expires_at = sqlc.narg(offer_expires_at)::timestamptz,
    cancelled_at = CASE WHEN sqlc.arg(status)::registration_status = 'cancelled' THEN clock_timestamp() ELSE cancelled_at END,
    updated_at = clock_timestamp()
WHERE id = sqlc.arg(id)
AND (sqlc.arg(status)::registration_status <> 'confirmed'
    OR status NOT IN ('offered','pending_approval')
    OR (status='offered' AND offer_expires_at>clock_timestamp())
    OR (status='pending_approval' AND approval_expires_at>clock_timestamp()))
RETURNING *;

-- name: AdjustWorkflowInventory :exec
UPDATE events SET current_rsvp_count = current_rsvp_count + sqlc.arg(confirmed_delta)::integer,
    held_rsvp_count = held_rsvp_count + sqlc.arg(held_delta)::integer,
    updated_at = clock_timestamp()
WHERE id = sqlc.arg(event_id);

-- name: FirstWaitlistedRSVP :one
SELECT * FROM rsvps WHERE event_id = $1 AND status = 'waitlisted' ORDER BY registered_at, id LIMIT 1 FOR UPDATE;

-- name: ExpiredEventOffers :many
SELECT * FROM rsvps WHERE event_id = $1 AND status = 'offered' AND offer_expires_at <= clock_timestamp()
ORDER BY offer_expires_at, id LIMIT 100 FOR UPDATE;

-- name: WorkflowMaintenanceEvents :many
SELECT DISTINCT r.event_id FROM rsvps r JOIN events e ON e.id=r.event_id
WHERE (r.status = 'offered' AND r.offer_expires_at <= clock_timestamp())
OR (r.status = 'pending_approval' AND r.approval_expires_at <= clock_timestamp())
OR (r.status='waitlisted' AND e.status='published' AND e.waitlist_enabled AND e.starts_at>clock_timestamp()
    AND (e.registration_opens_at IS NULL OR e.registration_opens_at<=clock_timestamp())
    AND (e.registration_closes_at IS NULL OR e.registration_closes_at>clock_timestamp())
    AND (e.capacity IS NULL OR e.current_rsvp_count+e.held_rsvp_count<e.capacity))
ORDER BY r.event_id LIMIT 50;

-- name: EnsureTicket :one
INSERT INTO tickets(rsvp_id) VALUES ($1) ON CONFLICT(rsvp_id) DO UPDATE
SET id = CASE WHEN tickets.status = 'revoked' THEN gen_random_uuid() ELSE tickets.id END,
public_code = CASE WHEN tickets.status = 'revoked' THEN 'T-' || encode(public.gen_random_bytes(12),'hex') ELSE tickets.public_code END,
status = 'issued', checked_in_at = NULL, checked_in_by = NULL, updated_at = clock_timestamp()
RETURNING *;

-- name: GetAccountTicket :one
SELECT t.*, r.event_id FROM tickets t JOIN rsvps r ON r.id = t.rsvp_id
WHERE r.id = $1 AND r.user_id = $2 AND r.status = 'confirmed' AND t.status IN ('issued','scanned');

-- name: RevokeWorkflowTicket :exec
UPDATE tickets SET status = 'revoked', updated_at = clock_timestamp() WHERE rsvp_id = $1;

-- name: CheckInWorkflowTicket :one
UPDATE tickets t SET status = 'scanned', checked_in_at = clock_timestamp(), checked_in_by = sqlc.arg(actor_id), updated_at = clock_timestamp()
FROM rsvps r WHERE t.id = sqlc.arg(ticket_id) AND t.rsvp_id = r.id AND r.event_id = sqlc.arg(event_id)
AND r.status = 'confirmed' AND t.status = 'issued' RETURNING t.*;

-- name: GetWorkflowTicket :one
SELECT t.*, r.event_id FROM tickets t JOIN rsvps r ON r.id = t.rsvp_id WHERE t.id = $1;

-- name: AppendWorkflowAudit :exec
INSERT INTO audit_log(actor_user_id,event_id,entity_id,action,request_id,metadata) VALUES ($1,$2,$3,$4,$5,$6);

-- name: GetIdempotentResponse :one
SELECT * FROM idempotency_records WHERE scope_hash = $1 AND key_hash = $2 AND expires_at > clock_timestamp();

-- name: StoreIdempotentResponse :exec
INSERT INTO idempotency_records(scope_hash,key_hash,request_hash,response_json) VALUES ($1,$2,$3,$4)
ON CONFLICT(scope_hash,key_hash) DO UPDATE SET request_hash = EXCLUDED.request_hash,
response_json = EXCLUDED.response_json, created_at = NOW(), expires_at = EXCLUDED.expires_at
WHERE idempotency_records.expires_at <= NOW();

-- name: ListAccountRegistrations :many
SELECT r.id, r.public_code, r.event_id, r.status, r.offer_expires_at, r.registered_at, e.title, e.slug, e.starts_at, e.ends_at, e.timezone, e.status AS event_status
FROM rsvps r JOIN events e ON e.id = r.event_id
WHERE r.user_id = $1 ORDER BY e.starts_at DESC, r.id LIMIT $2 OFFSET $3;

-- name: EventRoster :many
SELECT r.id,r.public_code,r.user_id,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS guest_name,
    COALESCE(NULLIF(r.guest_email,''),u.email,'')::text AS guest_email,
    r.status,r.notes,r.registered_at,r.offer_expires_at,t.id AS ticket_id,t.status AS ticket_status,t.checked_in_at
FROM rsvps r LEFT JOIN tickets t ON t.rsvp_id = r.id LEFT JOIN users u ON u.id = r.user_id
WHERE r.event_id = $1 ORDER BY r.registered_at,r.id LIMIT $2 OFFSET $3;

-- name: ListWorkflowAudit :many
SELECT id, actor_user_id, entity_id, action, metadata, created_at FROM audit_log WHERE event_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3;

-- name: UpdateWorkflowEvent :one
UPDATE events SET title=$2, description=$3, format=$4, starts_at=$5, ends_at=$6, timezone=$7,
capacity=$8, location_name=$9, location_address=$10, online_url=$11, registration_opens_at=$12,
registration_closes_at=$13, registration_mode=$14, allow_guest=$15, waitlist_enabled=$16,
offer_minutes=$17, revision=revision+1, updated_at=clock_timestamp()
WHERE id=$1 RETURNING *;

-- name: TransitionWorkflowEvent :one
UPDATE events SET status=sqlc.arg(status)::event_status, revision=revision+1, updated_at=clock_timestamp(),
published_at=CASE WHEN sqlc.arg(status)::event_status='published' THEN clock_timestamp() ELSE published_at END,
cancelled_at=CASE WHEN sqlc.arg(status)::event_status='cancelled' THEN clock_timestamp() ELSE cancelled_at END
WHERE id=sqlc.arg(id) RETURNING *;

-- name: CancelEventRegistrations :exec
UPDATE rsvps SET status='cancelled', cancelled_at=clock_timestamp(), offer_expires_at=NULL, updated_at=clock_timestamp()
WHERE event_id=$1 AND status IN ('confirmed','waitlisted','pending_approval','offered');

-- name: RevokeEventTickets :exec
UPDATE tickets SET status='revoked',updated_at=clock_timestamp() WHERE rsvp_id IN (SELECT id FROM rsvps WHERE event_id=$1);

-- name: EnqueueEventNotification :exec
INSERT INTO email_outbox(rsvp_id,event_id,recipient_email,recipient_name,template_type,payload_json)
SELECT r.id,r.event_id,COALESCE(u.email,r.guest_email),COALESCE(u.display_name,r.guest_name),sqlc.arg(template_type)::text,sqlc.arg(payload_json)::jsonb
FROM rsvps r LEFT JOIN users u ON u.id=r.user_id
WHERE r.event_id=sqlc.arg(event_id) AND r.status IN ('confirmed','offered','pending_approval','waitlisted')
AND COALESCE(u.email,r.guest_email) IS NOT NULL;

-- name: CleanupWorkflowKeys :exec
DELETE FROM idempotency_records WHERE (scope_hash,key_hash) IN
(SELECT scope_hash,key_hash FROM idempotency_records WHERE expires_at < NOW() ORDER BY expires_at LIMIT 500);

-- name: ResetWorkflowRSVP :one
UPDATE rsvps SET status=$2, notes=$3, cancelled_at=NULL, offer_expires_at=NULL,
approval_expires_at=$4, registered_at=clock_timestamp(), updated_at=clock_timestamp() WHERE id=$1 RETURNING *;

-- name: EnqueueRegistrationNotification :execrows
INSERT INTO email_outbox(rsvp_id,event_id,recipient_email,recipient_name,template_type,payload_json)
SELECT r.id,r.event_id,COALESCE(u.email,r.guest_email),COALESCE(u.display_name,r.guest_name),sqlc.arg(template_type)::text,sqlc.arg(payload_json)::jsonb
FROM rsvps r LEFT JOIN users u ON u.id=r.user_id WHERE r.id=sqlc.arg(rsvp_id)
AND COALESCE(u.email,r.guest_email) IS NOT NULL;

-- name: ExpiredEventApplications :many
SELECT * FROM rsvps WHERE event_id=$1 AND status='pending_approval' AND approval_expires_at<=clock_timestamp()
ORDER BY approval_expires_at,id LIMIT 100 FOR UPDATE;

-- name: GetTicketByRSVP :one
SELECT * FROM tickets WHERE rsvp_id=$1;

-- name: GuestClaimEvents :many
SELECT DISTINCT r.event_id FROM rsvps r JOIN events e ON e.id=r.event_id
WHERE e.deleted_at IS NULL AND r.user_id IS NULL AND lower(r.guest_email)=lower(sqlc.arg(email)::text)
AND NOT EXISTS (SELECT 1 FROM rsvps account_rsvp WHERE account_rsvp.event_id=r.event_id AND account_rsvp.user_id=sqlc.arg(user_id))
ORDER BY r.event_id LIMIT 50;

-- name: ClaimGuestEventRegistration :one
UPDATE rsvps target SET user_id=sqlc.arg(user_id),updated_at=clock_timestamp()
WHERE target.event_id=sqlc.arg(event_id) AND target.user_id IS NULL AND lower(target.guest_email)=lower(sqlc.arg(email)::text)
AND target.id=(SELECT r.id FROM rsvps r WHERE r.event_id=sqlc.arg(event_id) AND r.user_id IS NULL AND lower(r.guest_email)=lower(sqlc.arg(email)::text) ORDER BY r.created_at,r.id LIMIT 1)
RETURNING target.*;

-- name: EventWorkflowActivity :one
SELECT EXISTS(SELECT 1 FROM rsvps active WHERE active.event_id=sqlc.arg(event_id) AND active.status IN ('confirmed','offered','waitlisted','pending_approval')) AS registration_active,
EXISTS(SELECT 1 FROM rsvps queued WHERE queued.event_id=sqlc.arg(event_id) AND queued.status IN ('waitlisted','offered')) AS waitlist_active;

-- name: SoftDeleteWorkflowEvent :exec
UPDATE events SET deleted_at=clock_timestamp(),revision=revision+1,updated_at=clock_timestamp() WHERE id=$1;

-- name: EditRegistrationNotes :one
UPDATE rsvps SET notes=$2,updated_at=clock_timestamp() WHERE id=$1 RETURNING *;

-- name: FindVerifiedTransferRecipient :one
SELECT id,email,display_name FROM users WHERE lower(email)=lower($1) AND email_verified AND deleted_at IS NULL;

-- name: CreateTransferredRegistration :one
INSERT INTO rsvps(event_id,user_id,guest_name,guest_email,status,transfer_count,transferred_from)
VALUES ($1,$2,$3,$4,'confirmed',1,$5) RETURNING *;

-- name: RegistrationWaitlistPosition :one
SELECT count(*)::bigint FROM rsvps r JOIN rsvps mine ON mine.id=$1
WHERE mine.status='waitlisted' AND r.event_id=mine.event_id AND r.status='waitlisted'
AND (r.registered_at,r.id)<=(mine.registered_at,mine.id);
