-- name: ProgramListPublic :many
SELECT * FROM programs WHERE status IN ('open','closed') ORDER BY created_at DESC,id LIMIT 51 OFFSET @page_offset;
-- name: ProgramRead :one
SELECT * FROM programs WHERE slug= @slug;
-- name: ProgramLock :one
SELECT * FROM programs WHERE slug= @slug FOR UPDATE;
-- name: ProgramAdminList :many
SELECT p.*, (SELECT count(*) FROM program_applications a WHERE a.program_id=p.id AND a.status='submitted')::bigint AS application_count,
 (SELECT count(*) FROM program_applications a WHERE a.program_id=p.id AND a.status='submitted' AND a.decision='pending')::bigint AS pending_count
FROM programs p WHERE @search::text='' OR p.title ILIKE '%'|| @search||'%' ORDER BY p.created_at DESC,p.id LIMIT 21 OFFSET @page_offset;
-- name: ProgramCreate :one
INSERT INTO programs(slug,title,summary,overview,responsibilities,recognition,status,minimum_age,questions,opens_at,closes_at,created_by)
VALUES (@slug,@title,@summary,@overview,@responsibilities,@recognition,@status,@minimum_age,@questions,sqlc.narg(opens_at),sqlc.narg(closes_at),@created_by) RETURNING *;
-- name: ProgramUpdate :one
UPDATE programs SET title= @title,summary= @summary,overview= @overview,responsibilities= @responsibilities,recognition= @recognition,
 status= @status,minimum_age= @minimum_age,questions= @questions,opens_at=sqlc.narg(opens_at),closes_at=sqlc.narg(closes_at),revision=revision+1,updated_at=clock_timestamp()
WHERE id= @id AND revision= @revision RETURNING *;
-- name: ProgramTouchReview :exec
UPDATE programs SET review_revision=review_revision+1 WHERE id= @id;
-- name: ProgramApplicationLock :one
SELECT * FROM program_applications WHERE program_id= @program_id AND user_id= @user_id FOR UPDATE;
-- name: ProgramApplicationLockID :one
SELECT * FROM program_applications WHERE program_id= @program_id AND id= @id FOR UPDATE;
-- name: ProgramApplicationCreate :one
INSERT INTO program_applications(program_id,user_id) VALUES(@program_id,@user_id) ON CONFLICT(program_id,user_id) DO UPDATE SET user_id=EXCLUDED.user_id RETURNING *;
-- name: ProgramApplicationSave :one
UPDATE program_applications SET full_name= @full_name,city= @city,phone= @phone,biography= @biography,profile_url= @profile_url,proposal= @proposal,
 age_confirmed= @age_confirmed,age= @age,answers= @answers,revision=revision+1,updated_at=clock_timestamp()
WHERE id= @id AND revision= @revision AND status='draft' RETURNING *;
-- name: ProgramApplicationSubmit :one
UPDATE program_applications SET status='submitted',question_snapshot= @question_snapshot,submitted_at=clock_timestamp(),updated_at=clock_timestamp(),revision=revision+1
WHERE id= @id AND status='draft' RETURNING *;
-- name: ProgramApplicationWithdraw :one
UPDATE program_applications SET status='withdrawn',updated_at=clock_timestamp(),revision=revision+1 WHERE id= @id AND status IN ('draft','submitted') RETURNING *;
-- name: ProgramApplicationMine :one
SELECT id,program_id,full_name,city,phone,biography,profile_url,proposal,age_confirmed,age,answers,question_snapshot,status,revision,submitted_at,updated_at,published_decision,published_at
FROM program_applications WHERE program_id= @program_id AND user_id= @user_id;
-- name: ProgramApplicationsAdmin :many
SELECT a.*,u.email,u.avatar_url FROM program_applications a JOIN users u ON u.id=a.user_id
WHERE a.program_id= @program_id AND a.status<>'draft' AND (@state::text='' OR a.status= @state) AND (@decision_filter::text='' OR a.decision= @decision_filter)
 AND (@search::text='' OR a.full_name ILIKE '%'|| @search||'%' OR u.email ILIKE '%'|| @search||'%' OR a.city ILIKE '%'|| @search||'%')
ORDER BY a.submitted_at DESC NULLS LAST,a.created_at DESC,a.id LIMIT 21 OFFSET @page_offset;
-- name: ProgramApplicationAdminRead :one
SELECT a.*,u.email,u.avatar_url FROM program_applications a JOIN users u ON u.id=a.user_id WHERE a.program_id= @program_id AND a.id= @id AND a.status<>'draft';
-- name: ProgramDecisionSave :one
UPDATE program_applications SET decision= @decision,reviewer_notes= @reviewer_notes,reviewed_by= @reviewed_by,reviewed_at=clock_timestamp(),revision=revision+1,updated_at=clock_timestamp()
WHERE id= @id AND revision= @revision AND status='submitted' AND published_at IS NULL RETURNING *;
-- name: ProgramReviewStats :one
SELECT count(*) FILTER(WHERE status='submitted')::bigint AS submitted,
count(*) FILTER(WHERE status='submitted' AND decision='pending')::bigint AS pending,
count(*) FILTER(WHERE status='submitted' AND decision<>'pending' AND published_at IS NULL)::bigint AS ready,
count(*) FILTER(WHERE status='submitted' AND published_at IS NOT NULL)::bigint AS released,
count(*) FILTER(WHERE status='submitted' AND decision='accepted')::bigint AS accepted,
count(*) FILTER(WHERE status='submitted' AND decision='rejected')::bigint AS rejected,
count(*) FILTER(WHERE status='submitted' AND decision='waitlisted')::bigint AS waitlisted
FROM program_applications WHERE program_id= @program_id;
-- name: ProgramReleaseLookup :one
SELECT * FROM program_releases WHERE program_id= @program_id AND request_key= @request_key;
-- name: ProgramReleaseCreate :one
INSERT INTO program_releases(program_id,request_key,payload_hash,application_count,released_by)
VALUES(@program_id,@request_key,@payload_hash,@application_count,@released_by) RETURNING *;
-- name: ProgramReleaseDecisions :one
WITH published AS (
 UPDATE program_applications SET published_decision=decision,published_at=clock_timestamp(),revision=revision+1,updated_at=clock_timestamp()
 WHERE program_id= @program_id AND status='submitted' AND decision<>'pending' AND published_at IS NULL
 RETURNING id,program_id,user_id,full_name
), messages AS (
 INSERT INTO email_outbox(recipient_email,recipient_name,template_type,payload_json)
 SELECT u.email,a.full_name,'program_results_released',jsonb_build_object('application_id',a.id,'program_title',p.title,'program_slug',p.slug)
 FROM published a JOIN programs p ON p.id=a.program_id JOIN users u ON u.id=a.user_id
 WHERE u.deleted_at IS NULL AND u.email_verified
 RETURNING id
) SELECT count(*)::bigint FROM published;

-- name: ProgramEnqueueReceipt :execrows
INSERT INTO email_outbox(recipient_email,recipient_name,template_type,payload_json)
SELECT u.email,a.full_name,'program_application_received',jsonb_build_object('application_id',a.id,'program_title',p.title,'program_slug',p.slug)
FROM program_applications a JOIN programs p ON p.id=a.program_id JOIN users u ON u.id=a.user_id
WHERE a.id= @id AND a.status='submitted' AND u.deleted_at IS NULL AND u.email_verified;

-- name: ProgramNotificationActive :one
SELECT EXISTS(SELECT 1 FROM email_outbox o
 JOIN program_applications a ON a.id=(o.payload_json->>'application_id')::uuid
 JOIN users u ON u.id=a.user_id
 WHERE o.id= @id AND a.status='submitted' AND u.deleted_at IS NULL AND u.email_verified
 AND lower(u.email)=lower(o.recipient_email)
 AND (o.template_type='program_application_received' OR (o.template_type='program_results_released' AND a.published_at IS NOT NULL)));
-- name: ProgramReleaseList :many
SELECT id,application_count,created_at FROM program_releases WHERE program_id= @program_id ORDER BY created_at DESC,id LIMIT 20;
-- name: ProgramActivityAppend :exec
INSERT INTO program_activity(program_id,application_id,actor_id,action) VALUES(@program_id,sqlc.narg(application_id),@actor_id,@action);
