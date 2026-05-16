// Package db provides database connection setup for walking-drum.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect parses the given connection string, applies pool defaults, and
// returns a ready-to-use *pgxpool.Pool. It pings the database before
// returning so callers can fail fast on a misconfigured or unreachable DB.
//
// The caller is responsible for calling pool.Close() at shutdown.
func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %v", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	// Optional: a per-connection statement timeout so a single bad query
	// can't tie up a connection forever.
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "30000" // ms

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v\n", err)
	}

	// Eager connectivity check — fail at startup, not on first request.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %v", err)
	}

	return pool,  nil
}
