// internal/api/server.go
package api

import (
	"context"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/handlers"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the HTTP server
type Server struct {
	echo              *echo.Echo
	config            *config.Config
	db                *postgres.DB
	productHandler    *handlers.ProductHandler
	priceHandler      *handlers.PriceHandler
	customerHandler   *handlers.CustomerHandler
	subscriptionHandler *handlers.SubscriptionHandler
}

// NewServer creates a new server instance with all its dependencies
func NewServer(
	cfg *config.Config,
	db *postgres.DB,
	productHandler *handlers.ProductHandler,
	priceHandler *handlers.PriceHandler,
	customerHandler *handlers.CustomerHandler,
	subscriptionHandler *handlers.SubscriptionHandler,
) *Server {
	e := echo.New()
	
	// Set server properties
	e.HideBanner = true
	e.Debug = cfg.App.Debug
	
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	// Create server
	server := &Server{
		echo:              e,
		config:            cfg,
		db:                db,
		productHandler:    productHandler,
		priceHandler:      priceHandler,
		customerHandler:   customerHandler,
		subscriptionHandler: subscriptionHandler,
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
	return s.echo.Shutdown(ctx)
}