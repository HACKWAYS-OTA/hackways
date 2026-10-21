-- name: CreateOAuthState :exec
INSERT INTO oauth_states (state_hash, verifier, return_to, nonce_hash, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: ConsumeOAuthState :one
UPDATE oauth_states
SET consumed_at = NOW()
WHERE state_hash = $1 AND consumed_at IS NULL AND expires_at > NOW()
RETURNING verifier, return_to, nonce_hash;

-- name: CreateOAuthExchangeToken :exec
INSERT INTO oauth_exchange_tokens (token_hash, user_id, nonce_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeOAuthExchangeToken :one
UPDATE oauth_exchange_tokens
SET used_at = NOW()
WHERE token_hash = $1 AND nonce_hash = $2 AND used_at IS NULL AND expires_at > NOW()
RETURNING user_id;

-- name: CleanupOAuthExchangeTokens :exec
DELETE FROM oauth_exchange_tokens
WHERE expires_at <= NOW() OR used_at < NOW() - INTERVAL '1 hour';

-- name: CleanupOAuthStates :exec
DELETE FROM oauth_states
WHERE expires_at <= NOW() OR consumed_at < NOW() - INTERVAL '1 hour';
