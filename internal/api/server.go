// internal/api/server.go
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/handlers"
	custommiddleware "github.com/dukerupert/walking-drum/internal/middleware"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Server represents the HTTP server
type Server struct {
	echo    *echo.Echo
	config  *config.Config
	db      *postgres.DB
	logger  *zerolog.Logger
	handler *handlers.Handlers
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

	// Initialize services
	s := services.CreateServices(cfg, r, logger)

	// Initialize handlers
	h := handlers.CreateHandlers(cfg, s, logger)

	// Create server
	server := &Server{
		echo:    e,
		config:  cfg,
		db:      db,
		handler: h,
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
