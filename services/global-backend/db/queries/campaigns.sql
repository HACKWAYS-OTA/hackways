-- name: CampaignList :many
SELECT reference,title,subject,status,revision,created_at,updated_at,queued_at FROM email_campaigns
WHERE (sqlc.arg(q)::text='' OR title ILIKE '%'||sqlc.arg(q)||'%' OR subject ILIKE '%'||sqlc.arg(q)||'%')
ORDER BY reference DESC LIMIT 21 OFFSET sqlc.arg(page_offset);

-- name: CampaignRead :one
SELECT * FROM email_campaigns WHERE reference=$1;

-- name: CampaignLock :one
SELECT * FROM email_campaigns WHERE reference=$1 FOR UPDATE;

-- name: CampaignCreate :one
INSERT INTO email_campaigns(title,subject,body,action_label,action_url,created_by,creation_key,creation_hash,configuration)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT(created_by,creation_key) DO UPDATE SET creation_key=EXCLUDED.creation_key
RETURNING *;

-- name: CampaignEdit :execrows
UPDATE email_campaigns SET title=$2,subject=$3,body=$4,action_label=$5,action_url=$6,configuration=$8,revision=revision+1,updated_at=clock_timestamp()
WHERE id=$1 AND status='draft' AND revision=$7;

-- name: CampaignSetQueued :exec
UPDATE email_campaigns SET status='queued',queued_at=clock_timestamp(),updated_at=clock_timestamp() WHERE id=$1;

-- name: CampaignCancel :exec
UPDATE email_campaigns SET status='cancelled',cancelled_at=clock_timestamp(),updated_at=clock_timestamp() WHERE id=$1;

-- name: CampaignAudienceCount :one
SELECT count(*) FROM promotion_contacts c WHERE c.subscribed
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email));

-- name: CampaignAudiencePreview :many
SELECT c.reference,c.email,c.display_name,c.source,c.consented_at FROM promotion_contacts c WHERE c.subscribed
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email))
ORDER BY c.reference LIMIT 20;

-- name: CampaignStats :one
SELECT count(*) FILTER(WHERE NOT d.is_test)::bigint AS total,
count(*) FILTER(WHERE NOT d.is_test AND o.status='pending')::bigint AS pending,
count(*) FILTER(WHERE NOT d.is_test AND o.status='sent')::bigint AS accepted,
count(*) FILTER(WHERE NOT d.is_test AND o.status='failed')::bigint AS failed,
count(*) FILTER(WHERE d.is_test)::bigint AS tests
FROM email_campaign_deliveries d JOIN email_outbox o ON o.id=d.outbox_id WHERE d.campaign_id=$1;

-- name: CampaignDeliveryList :many
SELECT o.admin_reference AS reference,o.recipient_email,o.status::text AS status,o.retry_count,o.sent_at,d.is_test
FROM email_campaign_deliveries d JOIN email_outbox o ON o.id=d.outbox_id
WHERE d.campaign_id=$1 ORDER BY o.created_at DESC,o.id LIMIT 21 OFFSET sqlc.arg(page_offset);

-- name: PromotionContactsList :many
SELECT reference,email,display_name,source,consent_source,consented_at,subscribed,unsubscribed_at FROM promotion_contacts
ORDER BY reference DESC LIMIT 21 OFFSET sqlc.arg(page_offset);

-- name: ImportPromotionContact :execrows
INSERT INTO promotion_contacts(email,display_name,source,consent_source,consented_at,subscribed)
VALUES($1,$2,'import',$3,$4,TRUE) ON CONFLICT(email) DO NOTHING;

-- name: ReadPromotionPreference :one
SELECT EXISTS(SELECT 1 FROM promotion_contacts WHERE email=lower(sqlc.arg(email)::text) AND subscribed);

-- name: SetPromotionPreference :exec
INSERT INTO promotion_contacts(email,display_name,user_id,source,consent_source,consented_at,subscribed,unsubscribed_at)
VALUES($1,$2,$3,'account','Account email preference',clock_timestamp(),$4,CASE WHEN $4 THEN NULL ELSE clock_timestamp() END)
ON CONFLICT(email) DO UPDATE SET user_id=EXCLUDED.user_id,source='account',consent_source=EXCLUDED.consent_source,
consented_at=EXCLUDED.consented_at,subscribed=EXCLUDED.subscribed,unsubscribed_at=EXCLUDED.unsubscribed_at,updated_at=clock_timestamp();

-- name: PromotionTestContact :one
INSERT INTO promotion_contacts(email,display_name,user_id,source,consent_source,consented_at,subscribed)
VALUES($1,$2,$3,'test','Admin requested preview email',clock_timestamp(),FALSE)
ON CONFLICT(email) DO UPDATE SET email=EXCLUDED.email RETURNING id;

-- name: QueueCampaignMessages :one
WITH audience AS MATERIALIZED (
 SELECT c.id,c.email,c.display_name,gen_random_uuid() AS outbox_id,encode(gen_random_bytes(32),'hex')::text AS unsubscribe_key
 FROM promotion_contacts c WHERE
 ((sqlc.arg(test_only)::boolean AND c.id=sqlc.narg(test_contact_id)::uuid)
 OR (NOT sqlc.arg(test_only)::boolean AND c.subscribed AND
 (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email))))
 AND (sqlc.arg(test_only)::boolean OR ((sqlc.arg(audience_source)::text='all' OR c.source=sqlc.arg(audience_source)) AND (coalesce(cardinality(sqlc.arg(recipient_emails)::text[]),0)=0 OR c.email=ANY(sqlc.arg(recipient_emails)::text[]))))
 AND NOT EXISTS(SELECT 1 FROM email_campaign_deliveries d WHERE d.campaign_id=sqlc.arg(campaign_id) AND d.contact_id=c.id AND d.revision=sqlc.arg(revision) AND d.is_test=sqlc.arg(test_only))
 ORDER BY c.reference LIMIT sqlc.arg(max_recipients) FOR SHARE OF c
), outboxes AS (
 INSERT INTO email_outbox(id,recipient_email,recipient_name,template_type,payload_json,scheduled_at)
 SELECT a.outbox_id,a.email,a.display_name,'admin_promotion',sqlc.arg(base_payload)::jsonb || jsonb_build_object('unsubscribe_url','https://hackways.in/auth/unsubscribe?key='||a.unsubscribe_key,'is_test',sqlc.arg(test_only)::boolean), CASE WHEN sqlc.arg(test_only)::boolean THEN clock_timestamp() ELSE coalesce(sqlc.narg(send_at)::timestamptz,clock_timestamp()) END
 FROM audience a RETURNING id
), deliveries AS (
 INSERT INTO email_campaign_deliveries(outbox_id,campaign_id,contact_id,revision,is_test,unsubscribe_hash)
 SELECT a.outbox_id,sqlc.arg(campaign_id),a.id,sqlc.arg(revision),sqlc.arg(test_only),encode(digest(a.unsubscribe_key,'sha256'),'hex') FROM audience a JOIN outboxes o ON o.id=a.outbox_id
 RETURNING outbox_id
) SELECT count(*) FROM deliveries;

-- name: CampaignMessageActive :one
SELECT EXISTS(SELECT 1 FROM email_campaign_deliveries d JOIN email_campaigns m ON m.id=d.campaign_id
JOIN promotion_contacts c ON c.id=d.contact_id JOIN email_outbox o ON o.id=d.outbox_id
WHERE d.outbox_id=$1 AND m.status<>'cancelled' AND lower(o.recipient_email)=c.email
AND ((d.is_test AND m.revision=d.revision) OR (NOT d.is_test AND m.status='queued' AND c.subscribed
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email)))));

-- name: UnsubscribePromotion :execrows
UPDATE promotion_contacts c SET subscribed=FALSE,unsubscribed_at=clock_timestamp(),updated_at=clock_timestamp()
WHERE c.id=(SELECT d.contact_id FROM email_campaign_deliveries d WHERE d.unsubscribe_hash=$1);


-- name: CampaignAudienceCountSelected :one
SELECT count(*) FROM promotion_contacts c WHERE c.subscribed
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email))
AND (sqlc.arg(audience_source)::text='all' OR c.source=sqlc.arg(audience_source)) AND (coalesce(cardinality(sqlc.arg(recipient_emails)::text[]),0)=0 OR c.email=ANY(sqlc.arg(recipient_emails)::text[]));


-- name: CampaignAudiencePreviewSelected :many
SELECT c.reference,c.email,c.display_name,c.source,c.consented_at FROM promotion_contacts c WHERE c.subscribed
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email))
AND (sqlc.arg(audience_source)::text='all' OR c.source=sqlc.arg(audience_source)) AND (coalesce(cardinality(sqlc.arg(recipient_emails)::text[]),0)=0 OR c.email=ANY(sqlc.arg(recipient_emails)::text[]))
ORDER BY c.reference LIMIT 20;


-- name: SetNewAccountPromotionPreference :exec
INSERT INTO promotion_contacts AS contact(email,display_name,user_id,source,consent_source,consented_at,subscribed,unsubscribed_at)
SELECT lower(u.email),u.display_name,u.id,'account','New account email preference',clock_timestamp(),sqlc.arg(subscribed)::boolean,
 CASE WHEN sqlc.arg(subscribed)::boolean THEN NULL ELSE clock_timestamp() END
FROM users u WHERE u.id=sqlc.arg(user_id)
ON CONFLICT(email) DO UPDATE SET user_id=EXCLUDED.user_id,display_name=EXCLUDED.display_name,source='account',
 subscribed=contact.subscribed AND EXCLUDED.subscribed,
 unsubscribed_at=CASE WHEN NOT EXCLUDED.subscribed THEN clock_timestamp() ELSE contact.unsubscribed_at END,
 updated_at=clock_timestamp();


-- name: CampaignRecipientSearch :many
SELECT c.reference,c.email,c.display_name,c.source FROM promotion_contacts c
WHERE c.subscribed AND (sqlc.arg(source)::text='all' OR c.source=sqlc.arg(source))
AND (c.source='import' OR EXISTS(SELECT 1 FROM users u WHERE u.id=c.user_id AND u.email_verified AND u.deleted_at IS NULL AND lower(u.email)=c.email))
AND (sqlc.arg(q)::text='' OR c.email ILIKE sqlc.arg(q)||'%' OR c.display_name ILIKE sqlc.arg(q)||'%')
ORDER BY c.email,c.reference LIMIT 21;
