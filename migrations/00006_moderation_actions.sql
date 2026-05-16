-- +goose Up

-- Append-only audit trail for moderation. See DESIGN.md §5.6.
-- This table is the source of truth; accounts.status is a denormalized cache.
-- Unbanning is a new row with action_type='unban', NOT an UPDATE to a prior row.
CREATE TABLE moderation_actions (
  id              UUID PRIMARY KEY,
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  action_type     TEXT NOT NULL
                    CHECK (action_type IN ('ban', 'suspend', 'mute', 'warn', 'unban')),
  reason          TEXT NOT NULL,
  details         JSONB NOT NULL DEFAULT '{}'::jsonb,
  applied_by      UUID REFERENCES accounts(id),
  -- clock_timestamp() (not NOW()) so two actions appended in the same
  -- transaction still get distinct timestamps. Ordering matters here
  -- because 'unban' is recognized as overturning a prior ban via
  -- applied_at > the ban's applied_at.
  applied_at      TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
  expires_at      TIMESTAMPTZ
);

-- Per-account history, newest first — drives "show this user's mod log."
CREATE INDEX moderation_actions_account_idx
  ON moderation_actions (account_id, applied_at DESC);

-- "Find currently-active timed actions" — only rows with an expiry are
-- candidates, and most accounts have none, so a partial index is the
-- right tool.
CREATE INDEX moderation_actions_active_idx
  ON moderation_actions (account_id, expires_at)
  WHERE expires_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS moderation_actions;
