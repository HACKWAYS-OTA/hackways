-- name: IsRestrictedDevice :one
SELECT EXISTS(SELECT 1 FROM device_restrictions WHERE token_hash=$1 AND blocked);

-- name: IsRestrictedRegistrationIdentity :one
SELECT EXISTS(SELECT 1 FROM users u JOIN account_restrictions a ON a.user_id=u.id AND a.blocked
WHERE u.id=sqlc.narg(user_id)::uuid OR lower(u.email)=lower(sqlc.arg(email)::text));
