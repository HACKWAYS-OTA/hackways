-- name: LockAuthenticationUser :one
SELECT id,email,password_hash,display_name,email_verified,device_baseline_at
FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE;

-- name: ReadEmailVerification :one
SELECT token_hash,expires_at,consumed_at,attempts,created_at,email
FROM email_verifications WHERE user_id=$1 FOR UPDATE;

-- name: FailEmailVerificationAttempt :exec
UPDATE email_verifications SET attempts=LEAST(attempts+1,5) WHERE user_id=$1;

-- name: ReadPasswordResetChallenge :one
SELECT token_hash,expires_at,consumed_at,attempts,created_at
FROM password_reset_challenges WHERE user_id=$1 FOR UPDATE;

-- name: CreatePasswordResetChallenge :exec
INSERT INTO password_reset_challenges(user_id,token_hash,expires_at)
VALUES($1,$2,clock_timestamp()+INTERVAL '15 minutes')
ON CONFLICT(user_id) DO UPDATE SET token_hash=EXCLUDED.token_hash,expires_at=EXCLUDED.expires_at,
consumed_at=NULL,attempts=0,created_at=clock_timestamp();

-- name: FindPasswordResetUser :one
SELECT user_id FROM password_reset_challenges WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>clock_timestamp();

-- name: ConsumePasswordResetChallenge :exec
UPDATE password_reset_challenges SET consumed_at=clock_timestamp() WHERE user_id=$1;

-- name: InvalidateUserEmailCodes :exec
UPDATE email_verifications SET consumed_at=clock_timestamp() WHERE user_id=$1;

-- name: RevokeUserOAuthExchanges :exec
DELETE FROM oauth_exchange_tokens WHERE user_id=$1;

-- name: ResetAccountPassword :execrows
UPDATE users SET password_hash=$2,email_verified=TRUE,updated_at=clock_timestamp()
WHERE id=$1 AND deleted_at IS NULL;

-- name: ForgetUserDevices :exec
DELETE FROM recognized_devices WHERE user_id=$1;

-- name: RememberRecognizedDevice :exec
INSERT INTO recognized_devices(user_id,token_hash,user_agent,expires_at)
VALUES($1,$2,$3,clock_timestamp()+INTERVAL '180 days')
ON CONFLICT(user_id,token_hash) DO UPDATE SET last_seen_at=clock_timestamp(),
user_agent=EXCLUDED.user_agent,expires_at=EXCLUDED.expires_at;

-- name: IsRecognizedDevice :one
SELECT EXISTS(SELECT 1 FROM recognized_devices WHERE user_id=$1 AND token_hash=$2 AND expires_at>clock_timestamp());

-- name: MarkDeviceBaseline :exec
UPDATE users SET device_baseline_at=clock_timestamp() WHERE id=$1 AND device_baseline_at IS NULL;

-- name: BoundRecognizedDevices :exec
DELETE FROM recognized_devices d WHERE d.user_id=$1 AND d.token_hash IN
(SELECT r.token_hash FROM recognized_devices r WHERE r.user_id=$1 ORDER BY r.last_seen_at DESC,r.token_hash OFFSET 50);

-- name: CleanupExpiredRecognizedDevices :exec
DELETE FROM recognized_devices WHERE (user_id,token_hash) IN
(SELECT user_id,token_hash FROM recognized_devices WHERE expires_at<clock_timestamp() LIMIT 500);

-- name: PasswordResetTokenActive :one
SELECT EXISTS(SELECT 1 FROM password_reset_challenges WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>clock_timestamp());

-- name: ReadOAuthExchangeUser :one
SELECT user_id FROM oauth_exchange_tokens
WHERE token_hash=$1 AND nonce_hash=$2 AND expires_at>clock_timestamp() AND used_at IS NULL;
