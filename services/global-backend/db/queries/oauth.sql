-- name: CreateOAuthExchangeToken :exec
INSERT INTO oauth_exchange_tokens (token_hash, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: ConsumeOAuthExchangeToken :one
UPDATE oauth_exchange_tokens
SET used_at = NOW()
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
RETURNING user_id;

-- name: CleanupOAuthExchangeTokens :exec
DELETE FROM oauth_exchange_tokens
WHERE expires_at <= NOW() OR used_at < NOW() - INTERVAL '1 hour';
