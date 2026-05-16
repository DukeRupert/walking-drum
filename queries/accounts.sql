-- name: CreateAccount :one
INSERT INTO accounts (
  id, email, display_name, password_hash
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateAccountStatus :one
UPDATE accounts
SET status = $2
WHERE id = $1
RETURNING *;

-- name: SoftDeleteAccount :exec
UPDATE accounts
SET status = 'deleted', deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
