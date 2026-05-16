-- name: CreateEntity :one
-- Inserts just the entities row. Per-type tables (Layer 3) and the
-- entity_positions / components rows are added by the transactional
-- helper in Go, not here.
INSERT INTO entities (
  id, season_id, entity_type, created_at_tick
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetEntityByID :one
SELECT * FROM entities
WHERE id = $1;

-- name: SoftDeleteEntity :one
-- Sets destroyed_at_tick. Application code is responsible for also
-- deleting the entity_positions row in the same transaction (the
-- death/destroy helper does this); we don't cascade on soft-delete
-- so the audit log (Layer 6) can still inspect the entity's history.
UPDATE entities
SET destroyed_at_tick = $2
WHERE id = $1 AND destroyed_at_tick IS NULL
RETURNING *;

-- name: ListEntitiesByTypeInSeason :many
-- "All living characters in season N," "all NPCs in season N," etc.
-- Filters out soft-deleted rows so callers don't have to.
SELECT * FROM entities
WHERE season_id = $1
  AND entity_type = $2
  AND destroyed_at_tick IS NULL;
