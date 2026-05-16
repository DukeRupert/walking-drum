-- +goose Up

-- Sparse table for entities that have a position in the world. See
-- DESIGN.md §6.3. Most entities have one row here; abstract /
-- container-held entities don't (nullable columns on `entities`
-- would lie about that).
CREATE TABLE entity_positions (
  entity_id        UUID PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
  -- FK to regions(id) is deliberately deferred to Layer 4; this
  -- column is a plain INT in the meantime to keep the layer seam
  -- honest. Layer 4 will ALTER TABLE ... ADD FOREIGN KEY.
  region_id        INT NOT NULL,
  x                INT NOT NULL,
  y                INT NOT NULL,
  -- Enables broadcast debouncing and grace-period reconciliation —
  -- ties a position update to the tick that produced it.
  updated_at_tick  BIGINT NOT NULL
);

-- "What's at (x,y) in region R" — the spatial hot path. Btree composite
-- is enough until measurement says otherwise (GiST/BRIN deferred).
CREATE INDEX entity_positions_region_xy_idx
  ON entity_positions (region_id, x, y);

-- "What's in this region" — broader scans used by the region goroutine.
CREATE INDEX entity_positions_region_idx
  ON entity_positions (region_id);

-- +goose Down
DROP TABLE IF EXISTS entity_positions;
