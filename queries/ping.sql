-- Smoke-test query so sqlc generate has something to chew on
-- before any real schema lands. Delete when real queries arrive.

-- name: Ping :one
SELECT 1::int;
