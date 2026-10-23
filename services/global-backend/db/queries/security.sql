-- name: CountAuthAttempt :one
INSERT INTO auth_rate_limits(scope_hash,window_start,attempts)
VALUES ($1,date_trunc('minute',clock_timestamp()),1)
ON CONFLICT(scope_hash,window_start) DO UPDATE SET attempts=auth_rate_limits.attempts+1
RETURNING attempts;

-- name: CleanupAuthAttempts :exec
DELETE FROM auth_rate_limits WHERE (scope_hash,window_start) IN
(SELECT scope_hash,window_start FROM auth_rate_limits WHERE window_start<NOW()-INTERVAL '1 hour' LIMIT 500);

-- name: ClaimEmailOutbox :many
WITH picked AS (
 SELECT id FROM email_outbox WHERE status='pending' AND scheduled_at<=NOW()
 AND (COALESCE(cardinality(sqlc.arg(template_types)::text[]),0)=0 OR template_type=ANY(sqlc.arg(template_types)::text[]))
 ORDER BY scheduled_at,id LIMIT sqlc.arg(batch_limit)::int FOR UPDATE SKIP LOCKED
)
UPDATE email_outbox o SET scheduled_at=NOW()+INTERVAL '5 minutes',lease_token=gen_random_uuid(),updated_at=NOW()
FROM picked WHERE o.id=picked.id
RETURNING o.id,o.rsvp_id,o.event_id,o.recipient_email,o.recipient_name,o.template_type,o.payload_json,o.retry_count,o.lease_token,o.first_sent_at;

-- name: BeginOutboxProviderAttempt :one
UPDATE email_outbox SET first_sent_at=COALESCE(first_sent_at,clock_timestamp()) WHERE id=$1 AND lease_token=$2 AND status='pending'
RETURNING first_sent_at;

-- name: MarkClaimedOutboxSent :execrows
UPDATE email_outbox SET status='sent',sent_at=NOW(),updated_at=NOW(),error_message=NULL
WHERE id=$1 AND lease_token=$2 AND status='pending';

-- name: MarkClaimedOutboxFailed :execrows
UPDATE email_outbox SET retry_count=retry_count+1,scheduled_at=$3,status=$4,error_message=$5,updated_at=NOW()
WHERE id=$1 AND lease_token=$2 AND status='pending';

-- name: ReadOwnedUploadEvent :one
SELECT e.id FROM events e JOIN organizers o ON o.id=e.organizer_id
WHERE e.id=$1 AND o.user_id=$2 AND e.deleted_at IS NULL AND e.status!='cancelled' AND e.starts_at>clock_timestamp();

-- name: AttachOwnedEventBanner :one
UPDATE events e SET banner_url=$3,
revision=CASE WHEN e.banner_url=$3 THEN e.revision ELSE e.revision+1 END,
updated_at=CASE WHEN e.banner_url=$3 THEN e.updated_at ELSE clock_timestamp() END
FROM organizers o WHERE e.organizer_id=o.id AND e.id=$1 AND o.user_id=$2
AND e.deleted_at IS NULL AND e.status IN ('draft','unpublished','published') AND e.starts_at>clock_timestamp()
AND (e.revision=$4 OR (e.banner_url=$3 AND e.revision=$4+1))
RETURNING e.*;

-- name: SetStorageWriteTimeout :exec
SET LOCAL statement_timeout = '2000ms';

-- name: ListOwnedNotifications :many
SELECT o.id,o.event_id,o.template_type,o.status,o.retry_count,o.scheduled_at,o.sent_at,o.created_at
FROM email_outbox o JOIN rsvps r ON r.id=o.rsvp_id
WHERE r.user_id=$1 ORDER BY o.created_at DESC LIMIT $2 OFFSET $3;

-- name: RetryOwnedNotification :one
UPDATE email_outbox o SET status='pending',retry_count=0,scheduled_at=clock_timestamp(),updated_at=clock_timestamp()
FROM rsvps r WHERE o.rsvp_id=r.id AND r.user_id=$2 AND o.id=$1 AND o.status='failed'
AND o.updated_at<clock_timestamp()-INTERVAL '5 minutes' RETURNING o.id;

-- name: ListAccountSessions :many
SELECT id,user_agent,created_at,expires_at,last_seen FROM sessions WHERE user_id=$1 AND expires_at>NOW() ORDER BY last_seen DESC LIMIT 100;

-- name: RevokeOwnedSession :execrows
DELETE FROM sessions WHERE id=$1 AND user_id=$2;

-- name: CreateEmailVerification :exec
INSERT INTO email_verifications(user_id,email,token_hash,expires_at)
VALUES ($1,$2,$3,NOW()+INTERVAL '15 minutes') ON CONFLICT(user_id)
DO UPDATE SET email=EXCLUDED.email,token_hash=EXCLUDED.token_hash,expires_at=EXCLUDED.expires_at,consumed_at=NULL,attempts=0,created_at=NOW();

-- name: ConsumeEmailVerification :one
UPDATE email_verifications SET consumed_at=clock_timestamp()
WHERE token_hash=$1 AND consumed_at IS NULL AND attempts<5 AND expires_at>clock_timestamp() RETURNING user_id,email;

-- name: VerifyAccountEmail :execrows
UPDATE users SET email_verified=TRUE,updated_at=clock_timestamp() WHERE id=$1 AND lower(email)=lower($2) AND deleted_at IS NULL;

-- name: ReadOutboxCurrentState :one
SELECT r.status AS registration_status,r.offer_expires_at,e.status AS event_status
FROM email_outbox o LEFT JOIN rsvps r ON r.id=o.rsvp_id LEFT JOIN events e ON e.id=o.event_id WHERE o.id=$1;

-- name: VerificationCodeActive :one
SELECT EXISTS(SELECT 1 FROM email_verifications WHERE token_hash=$1 AND consumed_at IS NULL AND attempts<5 AND expires_at>clock_timestamp());

-- name: ReadConfirmationPass :one
SELECT t.id AS ticket_id, t.public_code AS ticket_code, r.event_id,
 r.public_code AS registration_code, COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS recipient_name,
 e.title,e.slug,e.description,e.starts_at,e.ends_at,e.timezone,e.banner_url,
 e.location_name,e.location_address,e.online_url,e.revision,o.created_at AS queued_at
FROM email_outbox o JOIN rsvps r ON r.id=o.rsvp_id JOIN tickets t ON t.rsvp_id=r.id
JOIN events e ON e.id=r.event_id LEFT JOIN users u ON u.id=r.user_id
WHERE o.id=$1 AND r.status='confirmed' AND t.status IN ('issued','scanned')
AND e.status='published' AND e.deleted_at IS NULL AND e.ends_at>clock_timestamp()-INTERVAL '24 hours'
AND lower(o.recipient_email)=lower(COALESCE(u.email,r.guest_email));
