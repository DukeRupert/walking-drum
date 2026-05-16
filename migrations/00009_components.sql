-- +goose Up

-- Generic sparse storage for optional, extensible per-entity state.
-- See DESIGN.md §6.4. Component types are a closed set defined by Go
-- code (typed structs serialized to JSONB); promote a component to its
-- own table only when §4.3's criteria are met.
CREATE TABLE components (
  entity_id        UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  component_type   TEXT NOT NULL,
  state            JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at_tick  BIGINT NOT NULL,
  updated_at_tick  BIGINT NOT NULL,
  -- One component of each type per entity. If multiple are ever
  -- needed, that's the signal to promote the component to its own
  -- table.
  PRIMARY KEY (entity_id, component_type)
);

-- The PK indexes (entity_id, component_type) for "what components does
-- this entity have." This reverse-direction index drives the other
-- access pattern: "find all entities with component X" (iterating
-- a system over its inputs). Joined queries filter on
-- entities.destroyed_at_tick IS NULL in application code.
CREATE INDEX components_type_entity_idx
  ON components (component_type, entity_id);

-- +goose Down
DROP TABLE IF EXISTS components;
