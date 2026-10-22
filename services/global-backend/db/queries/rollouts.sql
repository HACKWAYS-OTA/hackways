-- name: ReadRuntimeRollout :one
SELECT c.revision,c.cohort_salt,c.updated_at,v.reference,v.label,v.flags
FROM runtime_rollout_control c JOIN runtime_rollout_versions v ON v.reference=c.active_version WHERE c.singleton;

-- name: LockRuntimeRollout :one
SELECT * FROM runtime_rollout_control WHERE singleton FOR UPDATE;

-- name: ListRuntimeRolloutVersions :many
SELECT v.reference,v.label,v.flags,v.restored_from,v.created_at,COALESCE(u.display_name,'System')::text AS created_by
FROM runtime_rollout_versions v LEFT JOIN users u ON u.id=v.created_by
ORDER BY v.reference DESC LIMIT 21 OFFSET $1;

-- name: GetRuntimeRolloutVersion :one
SELECT reference,label,flags FROM runtime_rollout_versions WHERE reference=$1;

-- name: CreateRuntimeRolloutVersion :one
INSERT INTO runtime_rollout_versions(label,flags,restored_from,created_by) VALUES($1,$2,$3,$4) RETURNING reference;

-- name: ActivateRuntimeRolloutVersion :exec
UPDATE runtime_rollout_control SET active_version=$1,revision=revision+1,updated_at=clock_timestamp() WHERE singleton;
