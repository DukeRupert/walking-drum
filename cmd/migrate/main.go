package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/dukerupert/walking-drum/pkg/config"
)

func main() {
	var migrationsPath string
	var direction string

	flag.StringVar(&migrationsPath, "path", "migrations", "Path to migration files")
	flag.StringVar(&direction, "direction", "up", "Migration direction (up or down)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Database connection string in the format required by golang-migrate
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	log.Default().Println(dbURL)

	// Initialize migrate
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		log.Fatalf("Migration initialization failed: %v", err)
	}

	// Execute migration based on direction
	var migrationErr error
	switch direction {
	case "up":
		migrationErr = m.Up()
	case "down":
		migrationErr = m.Down()
	default:
		log.Fatalf("Invalid migration direction: %s (must be 'up' or 'down')", direction)
	}

	if migrationErr != nil && migrationErr != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", migrationErr)
	}

	if migrationErr == migrate.ErrNoChange {
		log.Println("No migration executed (no change required)")
	} else {
		log.Printf("Migration '%s' successful", direction)
	}

	os.Exit(0)
}