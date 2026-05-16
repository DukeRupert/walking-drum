-- name: CreateSession :one
INSERT INTO sessions (
  id, account_id, token_hash, ip_address, user_agent, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetSessionByTokenHash :one
-- Caller is still responsible for checking expires_at / revoked_at and
-- enforcing whatever "valid session" means in the auth layer. This query
-- only finds the row.
SELECT * FROM sessions
WHERE token_hash = $1;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = NOW(), revoke_reason = $2
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: ListActiveSessionsForAccount :many
SELECT * FROM sessions
WHERE account_id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC;
