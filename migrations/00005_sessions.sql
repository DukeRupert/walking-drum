-- +goose Up

-- Active and recent auth sessions. See DESIGN.md §5.5.
-- token_hash is stored (never the raw token); same pattern as password storage.
-- Sessions are not WebSocket connections; a WebSocket may reconnect multiple
-- times within one session (grace period mechanic).
CREATE TABLE sessions (
  id              UUID PRIMARY KEY,
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  token_hash      TEXT NOT NULL UNIQUE,
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at      TIMESTAMPTZ NOT NULL,
  revoked_at      TIMESTAMPTZ,
  revoke_reason   TEXT
                    CHECK (revoke_reason IS NULL OR revoke_reason IN (
                      'user_logout', 'admin_revoke', 'expired', 'replaced'
                    ))
);

-- "Live sessions for this account" — narrow index on the active subset.
CREATE INDEX sessions_account_active_idx
  ON sessions (account_id) WHERE revoked_at IS NULL;

-- Expiry sweep — only scans not-already-revoked rows.
CREATE INDEX sessions_expires_idx
  ON sessions (expires_at) WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS sessions;
