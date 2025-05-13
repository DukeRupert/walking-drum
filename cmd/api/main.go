// cmd/api/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dukerupert/walking-drum/internal/api"
	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
)

func init() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func runMigrations(cfg *config.Config, logger *zerolog.Logger) error {
	// Run migrations
	m, err := migrate.New(
		"file://migrations",
		cfg.DB.MigrateURL)
	if err != nil {
		logger.Error().Err(err).Str("migrateURL", cfg.DB.MigrateURL).Msg("Failed to create migration instance")
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Log migration source and database URL
	logger.Debug().
		Str("source", "file://migrations").
		Str("migrateURL", cfg.DB.MigrateURL).
		Msg("Migration configuration")

	// Get migration version before running
	version, dirty, vErr := m.Version()
	if vErr != nil && vErr != migrate.ErrNilVersion {
		logger.Warn().Err(vErr).Msg("Failed to get current migration version")
	} else if vErr == migrate.ErrNilVersion {
		logger.Info().Msg("No migrations have been applied yet")
	} else {
		logger.Info().Uint("version", version).Bool("dirty", dirty).Msg("Current migration version")
	}

	// Run migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info().Msg("No migration changes detected")
		} else {
			logger.Error().Err(err).Msg("Migration failed")
			return fmt.Errorf("migration failed: %w", err)
		}
	} else {
		// Get the new version after successful migration
		newVersion, _, _ := m.Version()
		logger.Info().Uint("new_version", newVersion).Msg("Database migrations completed successfully")
	}

	// Close the migration
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		logger.Warn().Err(srcErr).Msg("Error closing migration source")
	}
	if dbErr != nil {
		logger.Warn().Err(dbErr).Msg("Error closing migration database connection")
	}

	// If both closing errors occurred, return a combined error
	if srcErr != nil && dbErr != nil {
		return fmt.Errorf("failed to close migration resources: %v, %v", srcErr, dbErr)
	} else if srcErr != nil {
		return fmt.Errorf("failed to close migration source: %w", srcErr)
	} else if dbErr != nil {
		return fmt.Errorf("failed to close migration database connection: %w", dbErr)
	}

	return nil
}

func run(ctx context.Context, args []string, w io.Writer) error {
	// Initialize logger
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := postgres.Connect(cfg.DB, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(cfg, &logger); err != nil {
        logger.Fatal().Err(err).Msg("Fatal migration error")
    }

	// Initialize server with handlers
	server := api.NewServer(
		cfg,
		db,
		&logger,
	)

	// Start the server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info().Uint("port", uint(cfg.App.Port)).Msg("Server listening on: ")
		serverErrors <- server.Start()
	}()

	// Wait for shutdown signal or server error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case err := <-serverErrors:
			if err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			}
		case <-ctx.Done():
			// Create shutdown context with timeout
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Shutdown the server gracefully
			if err := server.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "server shutdown error: %v\n", err)
			}
		}
		fmt.Println("Server gracefully stopped")
	}()

	wg.Wait()
	return nil
}

func main() {
	// Setup context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Run the application
	if err := run(ctx, os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
