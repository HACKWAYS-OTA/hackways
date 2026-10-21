-- name: SetEventVisibility :one
UPDATE events SET visibility=$2 WHERE id=$1 RETURNING *;
