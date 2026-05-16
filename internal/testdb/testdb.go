// Package testdb wires up a real Postgres connection for integration tests.
// Tests skip when no DATABASE_URL is available; otherwise WithTx hands out
// a transaction-bound *sqlc.Queries that is rolled back on test cleanup.
package testdb

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dukerupert/walking-drum/internal/db"
	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/envfile"
)

var (
	poolOnce sync.Once
	pool     *pgxpool.Pool
	poolErr  error
)

func setup() (*pgxpool.Pool, error) {
	poolOnce.Do(func() {
		root, err := findRepoRoot()
		if err != nil {
			poolErr = fmt.Errorf("find repo root: %w", err)
			return
		}
		if err := envfile.Load(filepath.Join(root, ".env")); err != nil && !errors.Is(err, fs.ErrNotExist) {
			poolErr = fmt.Errorf("load .env: %w", err)
			return
		}

		url := os.Getenv("TEST_DATABASE_URL")
		if url == "" {
			url = os.Getenv("DATABASE_URL")
		}
		if url == "" {
			poolErr = errors.New("DATABASE_URL not set")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, poolErr = db.Connect(ctx, url)
		if poolErr != nil {
			return
		}

		stdDB := stdlib.OpenDBFromPool(pool)
		defer stdDB.Close()
		if poolErr = goose.SetDialect("postgres"); poolErr != nil {
			return
		}
		if poolErr = goose.Up(stdDB, filepath.Join(root, "migrations")); poolErr != nil {
			return
		}
	})
	return pool, poolErr
}

// WithTx returns a fresh transaction and an sqlc.Queries bound to it.
// The transaction is rolled back automatically when the test ends, so
// tests are isolated and leave no rows behind.
func WithTx(t *testing.T) (*sqlc.Queries, pgx.Tx) {
	t.Helper()
	p, err := setup()
	if err != nil {
		t.Skipf("integration test skipped: %v", err)
	}
	ctx := context.Background()
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return sqlc.New(tx), tx
}

// findRepoRoot walks up from the test's cwd until it finds a go.mod,
// so callers don't need to know how deep their package lives.
func findRepoRoot() (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", errors.New("no go.mod ancestor of cwd")
		}
		d = parent
	}
}
