-- name: SetEntityPosition :one
-- Upsert: an entity has at most one position. Use this both for "spawn
-- this entity at (x,y)" and "this entity just moved." updated_at_tick
-- is bumped on every write so the broadcast layer can debounce.
INSERT INTO entity_positions (
  entity_id, region_id, x, y, updated_at_tick
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (entity_id) DO UPDATE
SET region_id = EXCLUDED.region_id,
    x = EXCLUDED.x,
    y = EXCLUDED.y,
    updated_at_tick = EXCLUDED.updated_at_tick
RETURNING *;

-- name: GetEntityPosition :one
SELECT * FROM entity_positions
WHERE entity_id = $1;

-- name: DeleteEntityPosition :exec
-- Called when an entity leaves the world: destroyed, moved into a
-- container, etc. Position is liveness state; absence of a row means
-- "not in the world."
DELETE FROM entity_positions
WHERE entity_id = $1;

-- name: GetEntitiesAtPosition :many
-- Drives "what's at (x,y) in region R" — pickup checks, collision
-- detection, click-to-inspect. Uses the (region_id, x, y) composite
-- index.
SELECT * FROM entity_positions
WHERE region_id = $1
  AND x = $2
  AND y = $3;

-- name: GetEntitiesInRegion :many
-- Broad scan: all positioned entities in a region. The region
-- goroutine uses this when warming its in-memory mirror; gameplay
-- code should prefer GetEntitiesAtPosition when possible.
SELECT * FROM entity_positions
WHERE region_id = $1;
