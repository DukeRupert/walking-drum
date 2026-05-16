-- +goose Up
-- No-op: verifies the migration tool can apply and roll back a migration end-to-end.
SELECT 1;

-- +goose Down
SELECT 1;
