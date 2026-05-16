-- +goose Up

-- The universal "thing in the world" handle. See DESIGN.md §6.2.
-- Every game object — character, NPC, item, corpse, projectile, world
-- object — has a row here. Per-type tables (Layer 3+) and the generic
-- components table layer behavior and state on top of this frame.
CREATE TABLE entities (
  id                UUID PRIMARY KEY,            -- UUIDv7, generated in Go
  season_id         INT NOT NULL REFERENCES seasons(id),
  entity_type       TEXT NOT NULL,
  created_at_tick   BIGINT NOT NULL,
  -- Soft-delete marker. NULL means "live." A periodic sweep (Layer 6)
  -- hard-deletes rows older than a retention window.
  destroyed_at_tick BIGINT,
  -- TEXT + CHECK rather than an enum: the set is closed but enum
  -- migrations are painful. entity_type is immutable in application
  -- code; character→corpse is destroy+create, two IDs.
  CHECK (entity_type IN (
    'character', 'npc', 'item', 'corpse', 'projectile', 'world_object'
  ))
);

-- Drives "all live characters in season N," "all NPCs in season N,"
-- etc. — the bread-and-butter lookup for per-type queries.
CREATE INDEX entities_season_type_idx
  ON entities (season_id, entity_type);

-- Partial index for the sweep job: most rows are live (NULL), so
-- indexing only the soft-deleted ones keeps the index tiny.
CREATE INDEX entities_destroyed_idx
  ON entities (destroyed_at_tick)
  WHERE destroyed_at_tick IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS entities;
