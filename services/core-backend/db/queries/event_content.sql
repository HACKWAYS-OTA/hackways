-- name: SetEventContent :one
UPDATE events SET faq_text=COALESCE(sqlc.narg(faq_text)::text,faq_text),
contact_url=COALESCE(sqlc.narg(contact_url)::text,contact_url),rewards=COALESCE(sqlc.narg(rewards)::text,rewards)
WHERE id=$1 RETURNING *;
