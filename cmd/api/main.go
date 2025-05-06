// cmd/api/main.go
package main

import (
	"context"
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
	"github.com/dukerupert/walking-drum/internal/handlers"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/internal/stripe"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
)

func init() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func run(ctx context.Context, args []string, w io.Writer) error {
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
	
	// Initialize Stripe client
	stripeClient := stripe.NewClient(cfg.Stripe.SecretKey)
	
	// Initialize repositories
	repos := postgres.NewRepositories(db)
	
	// Initialize services
	productService := services.NewProductService(repos.Product, stripeClient)
	priceService := services.NewPriceService(repos.Price, repos.Product, stripeClient)
	customerService := services.NewCustomerService(repos.Customer, stripeClient)
	subscriptionService := services.NewSubscriptionService(
		repos.Subscription,
		repos.Customer,
		repos.Product,
		repos.Price,
		stripeClient,
	)
	
	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService)
	priceHandler := handlers.NewPriceHandler(priceService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)
	
	// Initialize logger
 logger := zerolog.New(os.Stdout)
	// Initialize server with handlers
	server := api.NewServer(
		cfg,
		db,
		&logger,
		productHandler,
		priceHandler,
		customerHandler,
		subscriptionHandler,
	)
	
	// Start the server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info().Uint("port", cfg.App.Port).Msg("Server listening on: ")
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
