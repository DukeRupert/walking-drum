-- name: GetActiveSeason :one
SELECT * FROM seasons
WHERE status = 'active';

-- name: GetSeasonByID :one
SELECT * FROM seasons
WHERE id = $1;

-- name: UpdateSeasonStatus :one
-- Caller is responsible for transition validity. The unique partial index
-- on (status) WHERE status='active' ensures at most one active season.
UPDATE seasons
SET status = $2
WHERE id = $1
RETURNING *;
