// cmd/api/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	log.Print("hello world")
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Print configuration (with secrets hidden)
	if cfg.App.Debug {
		config.PrintConfig()
	}

	// Initialize database
	db, err := postgres.Connect(cfg.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	m, err := migrate.New(
		"file://migrations",
		cfg.DB.MigrateURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to find migration configuration")
	}
	if err := m.Up(); err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("Failed up migration")
	}

	log.Info().Msg("Database migrations complete")

	// Initialize repositories
	_ = postgres.NewRepositories(db)

	// Initialize server
	server := api.NewServer(cfg, db)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	if err := server.Shutdown(); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}
	fmt.Println("Server gracefully stopped")
}
