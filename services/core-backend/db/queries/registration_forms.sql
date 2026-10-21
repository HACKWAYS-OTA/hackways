-- name: ListRegistrationQuestions :many
SELECT public_code,label,description,kind,required,position,options FROM event_registration_questions
WHERE event_id=$1 AND active ORDER BY position,id;

-- name: ReadRegistrationForm :one
SELECT e.registration_form_revision,e.registration_name_required,
 COALESCE((SELECT jsonb_agg(jsonb_build_object('public_code',q.public_code,'label',q.label,'description',q.description,'kind',q.kind,'required',q.required,'options',q.options) ORDER BY q.position,q.id)
 FROM event_registration_questions q WHERE q.event_id=e.id AND q.active),'[]'::jsonb)::jsonb AS questions
FROM events e WHERE e.id=$1 AND e.deleted_at IS NULL;

-- name: DeactivateRegistrationQuestions :exec
UPDATE event_registration_questions SET active=false,updated_at=clock_timestamp() WHERE event_id=$1 AND active;

-- name: CreateRegistrationQuestion :one
INSERT INTO event_registration_questions(event_id,label,description,kind,required,position,options)
VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING public_code;

-- name: UpdateRegistrationQuestion :one
UPDATE event_registration_questions SET label=$3,description=$4,kind=$5,required=$6,position=$7,options=$8,active=true,updated_at=clock_timestamp()
WHERE event_id=$1 AND public_code=$2 RETURNING public_code;

-- name: UpdateRegistrationForm :exec
UPDATE events SET registration_form_revision=registration_form_revision+1,registration_name_required=$2,revision=revision+1,updated_at=clock_timestamp() WHERE id=$1;

-- name: InsertRegistrationSubmission :exec
INSERT INTO registration_submissions(rsvp_id,form_revision,answers) VALUES($1,$2,$3);

-- name: LatestRegistrationAnswers :one
SELECT answers,form_revision,created_at FROM registration_submissions WHERE rsvp_id=$1 ORDER BY id DESC LIMIT 1;

-- name: ManagedGuestDetail :one
SELECT r.id,r.public_code,r.user_id,COALESCE(NULLIF(r.guest_name,''),u.display_name,'')::text AS guest_name,
 COALESCE(NULLIF(r.guest_email,''),u.email,'')::text AS guest_email,
 r.status,r.notes,r.registered_at,r.updated_at,r.offer_expires_at,t.id AS ticket_id,t.status AS ticket_status,t.checked_in_at
FROM rsvps r LEFT JOIN tickets t ON t.rsvp_id=r.id LEFT JOIN users u ON u.id=r.user_id
WHERE r.event_id=$1 AND r.public_code=$2;
