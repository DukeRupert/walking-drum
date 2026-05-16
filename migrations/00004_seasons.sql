-- +goose Up

-- The wipe boundary, made explicit. See DESIGN.md §5.3.
CREATE TABLE seasons (
  id              INT PRIMARY KEY,
  name            TEXT,
  status          TEXT NOT NULL
                    CHECK (status IN ('upcoming', 'active', 'ended')),
  world_seed      BIGINT NOT NULL,
  modifiers       JSONB NOT NULL DEFAULT '{}'::jsonb,
  starts_at       TIMESTAMPTZ NOT NULL,
  ends_at         TIMESTAMPTZ NOT NULL,
  wiped_at        TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Only one season can be 'active' at any time; enforced at the DB level
-- rather than in application code.
CREATE UNIQUE INDEX seasons_one_active ON seasons (status) WHERE status = 'active';

-- Per-account, per-season state. Bridges forever-accounts and
-- season-ephemeral characters. See DESIGN.md §5.4.
CREATE TABLE season_participation (
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  season_id       INT NOT NULL REFERENCES seasons(id),
  characters_made INT NOT NULL DEFAULT 0,
  deaths          INT NOT NULL DEFAULT 0,
  deepest_region  INT,
  final_summary   JSONB,
  PRIMARY KEY (account_id, season_id)
);

-- Seed season 1 with placeholder dates and seed. Tightened later via
-- normal UPDATE before the season actually opens.
INSERT INTO seasons (id, name, status, world_seed, starts_at, ends_at)
VALUES (
  1,
  'Season 1',
  'upcoming',
  0,
  '2026-06-01 00:00:00+00',
  '2026-09-01 00:00:00+00'
);

-- +goose Down
DROP TABLE IF EXISTS season_participation;
DROP TABLE IF EXISTS seasons;
