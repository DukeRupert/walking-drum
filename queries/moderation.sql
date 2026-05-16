-- name: AppendModerationAction :one
-- Append-only. Reversing a prior action means inserting a new row
-- with action_type='unban'.
INSERT INTO moderation_actions (
  id, account_id, action_type, reason, details, applied_by, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListModerationActionsForAccount :many
SELECT * FROM moderation_actions
WHERE account_id = $1
ORDER BY applied_at DESC;

-- name: FindActiveBansAndSuspensions :many
-- A punitive action is currently in effect for $1 if:
--   - it is a 'ban' or 'suspend' (mute and warn are their own thing),
--   - it has not yet expired (or has no expiry), AND
--   - no later 'unban' action has overturned it.
SELECT m.*
FROM moderation_actions m
WHERE m.account_id = $1
  AND m.action_type IN ('ban', 'suspend')
  AND (m.expires_at IS NULL OR m.expires_at > NOW())
  AND NOT EXISTS (
    SELECT 1 FROM moderation_actions u
    WHERE u.account_id = m.account_id
      AND u.action_type = 'unban'
      AND u.applied_at > m.applied_at
  )
ORDER BY m.applied_at DESC;
