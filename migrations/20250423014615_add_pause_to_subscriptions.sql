-- +goose Up
-- +goose StatementBegin
-- Add the new status value to the enum
ALTER TYPE subscription_status ADD VALUE IF NOT EXISTS 'paused';

-- Add the new column
ALTER TABLE subscriptions ADD COLUMN resume_at TIMESTAMP WITH TIME ZONE;

-- Create an index for the new column
CREATE INDEX idx_subscriptions_resume_at ON subscriptions(resume_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the index
DROP INDEX IF EXISTS idx_subscriptions_resume_at;

-- Remove the column
ALTER TABLE subscriptions DROP COLUMN IF EXISTS resume_at;
-- +goose StatementEnd
