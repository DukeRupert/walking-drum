package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dukerupert/walking-drum/internal/db"
	"github.com/dukerupert/walking-drum/internal/envfile"
)

func main() {
	if err := envfile.Load(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := runMigrations(pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	fmt.Println("ok")
}

// runMigrations applies all pending goose migrations. Goose needs a
// database/sql handle, which we obtain by wrapping the pgx pool via
// stdlib.OpenDBFromPool. Closing that handle does not close the pool.
func runMigrations(pool *pgxpool.Pool) error {
	stdDB := stdlib.OpenDBFromPool(pool)
	defer stdDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(stdDB, "migrations"); err != nil {
		return fmt.Errorf("up: %w", err)
	}
	return nil
}
