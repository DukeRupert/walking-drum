-- name: SetComponent :one
-- Upsert keyed on (entity_id, component_type). The PK enforces "one of
-- each type per entity," so this is the only sensible write shape.
-- created_at_tick survives the ON CONFLICT path (keep the original);
-- updated_at_tick is bumped on every write.
INSERT INTO components (
  entity_id, component_type, state, created_at_tick, updated_at_tick
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (entity_id, component_type) DO UPDATE
SET state = EXCLUDED.state,
    updated_at_tick = EXCLUDED.updated_at_tick
RETURNING *;

-- name: GetComponent :one
SELECT * FROM components
WHERE entity_id = $1 AND component_type = $2;

-- name: DeleteComponent :exec
DELETE FROM components
WHERE entity_id = $1 AND component_type = $2;

-- name: ListEntitiesWithComponent :many
-- "Find all live entities with component X" — the system-iterates-over-
-- inputs pattern. JOINed to entities so we can filter out soft-deleted
-- rows; the (component_type, entity_id) index drives the components
-- side of the join.
SELECT c.entity_id, c.component_type, c.state, c.created_at_tick, c.updated_at_tick
FROM components c
JOIN entities e ON e.id = c.entity_id
WHERE c.component_type = $1
  AND e.destroyed_at_tick IS NULL;
