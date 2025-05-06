// internal/api/routes.go
package api

import (
	"github.com/labstack/echo/v4"
)

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// API group
	api := s.echo.Group("/api")

	// Version 1 group
	v1 := api.Group("/v1")

	// Product routes
	products := v1.Group("/products")
	products.POST("", s.productHandler.Create)
	products.GET("", s.productHandler.List)
	products.GET("/:id", s.productHandler.Get)
	products.PUT("/:id", s.productHandler.Update)
	products.DELETE("/:id", s.productHandler.Delete)
	products.PATCH("/:id/stock", s.productHandler.UpdateStockLevel)
	
	// Price routes
	prices := v1.Group("/prices")
	prices.POST("", s.priceHandler.Create)
	prices.GET("", s.priceHandler.List)
	prices.GET("/:id", s.priceHandler.Get)
	prices.PUT("/:id", s.priceHandler.Update)
	prices.DELETE("/:id", s.priceHandler.Delete)
	products.GET("/product/:productId", s.priceHandler.ListByProduct)
	
	// Customer routes
	customers := v1.Group("/customers")
	customers.POST("", s.customerHandler.Create)
	customers.GET("", s.customerHandler.List)
	customers.GET("/:id", s.customerHandler.Get)
	customers.GET("/email/:email", s.customerHandler.GetByEmail)
	customers.PUT("/:id", s.customerHandler.Update)
	customers.DELETE("/:id", s.customerHandler.Delete)
	
	// Subscription routes
	subscriptions := v1.Group("/subscriptions")
	subscriptions.POST("", s.subscriptionHandler.Create)
	subscriptions.GET("", s.subscriptionHandler.List)
	subscriptions.GET("/:id", s.subscriptionHandler.Get)
	subscriptions.PUT("/:id", s.subscriptionHandler.Update)
	subscriptions.POST("/:id/cancel", s.subscriptionHandler.Cancel)
	subscriptions.POST("/:id/pause", s.subscriptionHandler.Pause)
	subscriptions.POST("/:id/resume", s.subscriptionHandler.Resume)
	customers.GET("/:customerId/subscriptions", s.subscriptionHandler.ListByCustomer)
	
	// Health check route
	s.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}
