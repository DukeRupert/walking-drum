// internal/api/routes.go
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// registerRoutes registers all the API routes
func registerRoutes(s *Server) {
	// API group
	api := s.echo.Group("/api")

	// Version 1 group
	v1 := api.Group("/v1")

	// Products routes
	productsGroup := v1.Group("/products")
	productsGroup.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "List products endpoint - to be implemented",
		})
	})
	productsGroup.GET("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Get product endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})

	// Prices routes
	pricesGroup := v1.Group("/prices")
	pricesGroup.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "List prices endpoint - to be implemented",
		})
	})
	pricesGroup.GET("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Get price endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})

	// Customers routes
	customersGroup := v1.Group("/customers")
	customersGroup.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Create customer endpoint - to be implemented",
		})
	})
	customersGroup.GET("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Get customer endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})
	customersGroup.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Update customer endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})

	// Subscriptions routes
	subscriptionsGroup := v1.Group("/subscriptions")
	subscriptionsGroup.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Create subscription endpoint - to be implemented",
		})
	})
	subscriptionsGroup.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "List subscriptions endpoint - to be implemented",
		})
	})
	subscriptionsGroup.GET("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Get subscription endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})
	subscriptionsGroup.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Update subscription endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})
	subscriptionsGroup.DELETE("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Cancel subscription endpoint - to be implemented",
			"id":      c.Param("id"),
		})
	})

	// Webhook route for Stripe events
	v1.POST("/webhooks/stripe", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Stripe webhook endpoint - to be implemented",
		})
	})
}
