-- name: AdminMessageCreate :one
INSERT INTO admin_custom_messages(actor_user_id,creation_key,creation_hash,subject) VALUES($1,$2,$3,$4)
ON CONFLICT(actor_user_id,creation_key) DO UPDATE SET creation_key=EXCLUDED.creation_key RETURNING *;

-- name: AdminMessageQueuedCount :one
SELECT count(*) FROM admin_custom_message_deliveries WHERE message_id=$1;

-- name: QueueAdminCustomMessages :many
WITH recipients AS MATERIALIZED (
 SELECT id, email, display_name,gen_random_uuid() AS outbox_id FROM users
 WHERE lower(email)=ANY(sqlc.arg(recipient_emails)::text[]) AND email_verified AND deleted_at IS NULL
 ORDER BY id FOR SHARE
), outboxes AS (
 INSERT INTO email_outbox(id,recipient_email,recipient_name,template_type,payload_json)
 SELECT outbox_id,email,display_name,'admin_service_message',sqlc.arg(payload)::jsonb FROM recipients
 RETURNING id,admin_reference AS reference
), deliveries AS (
 INSERT INTO admin_custom_message_deliveries(message_id,user_id,outbox_id)
 SELECT sqlc.arg(message_id),r.id,r.outbox_id FROM recipients r JOIN outboxes o ON o.id=r.outbox_id RETURNING outbox_id
) SELECT o.reference FROM outboxes o JOIN deliveries d ON d.outbox_id=o.id;

-- name: AdminMessageRecipientActive :one
SELECT EXISTS(SELECT 1 FROM admin_custom_message_deliveries d JOIN users u ON u.id=d.user_id
JOIN email_outbox o ON o.id=d.outbox_id
WHERE d.outbox_id=$1 AND u.deleted_at IS NULL AND u.email_verified AND lower(u.email)=lower(o.recipient_email));
