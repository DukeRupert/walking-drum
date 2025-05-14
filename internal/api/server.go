// internal/api/server.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/handlers"
	"github.com/dukerupert/walking-drum/internal/messaging/consumers"
	"github.com/dukerupert/walking-drum/internal/messaging/publishers"
	"github.com/dukerupert/walking-drum/internal/messaging/rabbitmq"
	custommiddleware "github.com/dukerupert/walking-drum/internal/middleware"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Server represents the HTTP server
type Server struct {
	echo                        *echo.Echo
	config                      *config.Config
	db                          *postgres.DB
	logger                      *zerolog.Logger
	handler                     *handlers.Handlers
	subscriptionRenewalConsumer *consumers.SubscriptionRenewalConsumer
	emailConsumer               *consumers.EmailConsumer
}

// NewServer creates a new server instance with all its dependencies
func NewServer(
	cfg *config.Config,
	db *postgres.DB,
	logger *zerolog.Logger,
) *Server {
	e := echo.New()

	// Set server properties
	e.HideBanner = true
	e.Debug = cfg.App.Debug

	// Add middleware
	e.Use(custommiddleware.RequestLogger(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // For development. In production, specify your frontend URL
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Initialize repositories
	r := postgres.CreateRepositories(db, logger)

	// Initialize RabbitMQ client
	rabbitClient, err := rabbitmq.NewClient(rabbitmq.Config{
		URL:               cfg.RabbitMQ.URL,
		ReconnectInterval: time.Second * 5,
		ExchangeName:      "coffee_subscription",
	}, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbitClient.Close()

	// Initialize publishers
	subscriptionPublisher := publishers.NewSubscriptionPublisher(rabbitClient, logger)

	// Initialize services
	s := services.CreateServices(cfg, r, subscriptionPublisher, logger)

	// Initialize consumers
	subscriptionRenewalConsumer := consumers.NewSubscriptionRenewalConsumer(
		rabbitClient,
		logger,
		s.Subscription,
		s.Product,
	)

	emailConsumer := consumers.NewEmailConsumer(
		rabbitClient,
		logger,
		s.Email,
	)

	// Start consumers
	if err := subscriptionRenewalConsumer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start subscription renewal consumer")
	}

	if err := emailConsumer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start email consumer")
	}

	// Initialize handlers
	h := handlers.CreateHandlers(cfg, s, logger)

	// Create server
	server := &Server{
		echo:                        e,
		config:                      cfg,
		db:                          db,
		handler:                     h,
		subscriptionRenewalConsumer: subscriptionRenewalConsumer,
		emailConsumer:               emailConsumer,
	}

	// Setup router
	server.setupRoutes()

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.config.App.Port))
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop consumers
	if err := s.subscriptionRenewalConsumer.Stop(); err != nil {
		s.logger.Error().Err(err).Msg("Error stopping subscription renewal consumer")
	}

	if err := s.emailConsumer.Stop(); err != nil {
		s.logger.Error().Err(err).Msg("Error stopping email consumer")
	}

	return s.echo.Shutdown(ctx)
}
