// internal/api/server.go
package api

import (
	"context"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the HTTP server
type Server struct {
	echo   *echo.Echo
	config *config.Config
	db     *postgres.DB
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, db *postgres.DB) *Server {
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
		echo:   e,
		config: cfg,
		db:     db,
	}

	// Setup router
	setupRouter(server)

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.config.App.Port))
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.echo.Shutdown(ctx)
}
