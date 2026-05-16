-- +goose Up

-- Durable human identity. Survives all seasons. See DESIGN.md §5.1.
CREATE TABLE accounts (
  id              UUID PRIMARY KEY,
  email           CITEXT NOT NULL UNIQUE,
  email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
  display_name    TEXT NOT NULL UNIQUE,
  password_hash   TEXT NOT NULL,
  totp_secret     TEXT,
  totp_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
  status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'suspended', 'banned', 'deleted')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_login_at   TIMESTAMPTZ,
  deleted_at      TIMESTAMPTZ
);

-- Most accounts are active; only index the interesting ones.
CREATE INDEX accounts_status_idx ON accounts (status) WHERE status != 'active';

-- Sparse, cross-season per-account state. ECS pattern applied to accounts.
-- See DESIGN.md §5.2.
CREATE TABLE account_flags (
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  flag_type       TEXT NOT NULL,
  flag_value      JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (account_id, flag_type)
);

-- +goose Down
DROP TABLE IF EXISTS account_flags;
DROP TABLE IF EXISTS accounts;
