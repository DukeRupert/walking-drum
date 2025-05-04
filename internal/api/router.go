// internal/api/router.go
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// setupRouter configures all the routes for the server
func setupRouter(s *Server) {
	// Register all routes
	registerRoutes(s)

	// Add health check endpoint
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"name":   s.config.App.Name,
			"env":    s.config.App.Env,
		})
	})

	// Add 404 handler
	s.echo.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Route not found",
		})
	})
}
