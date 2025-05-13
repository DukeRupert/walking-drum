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
	products.POST("", s.handler.Product.Create)
	products.GET("", s.handler.Product.List)
	products.GET("/:id", s.handler.Product.Get)
	products.PUT("/:id", s.handler.Product.Update)
	products.DELETE("/:id", s.handler.Product.Delete)
	products.PATCH("/:id/stock", s.handler.Product.UpdateStockLevel)

	// Add variant routes
	variants := v1.Group("/variants")
	variants.GET("", s.handler.Variant.List)
	variants.GET("/:id", s.handler.Variant.Get)
	variants.PUT("/:id", s.handler.Variant.Update)
	variants.PATCH("/:id/stock", s.handler.Variant.UpdateStockLevel)
	variants.GET("/product/:productId", s.handler.Variant.ListByProduct)

	// Price routes
	prices := v1.Group("/prices")
	prices.POST("", s.handler.Price.Create)
	prices.GET("", s.handler.Price.List)
	prices.GET("/:id", s.handler.Price.Get)
	prices.PUT("/:id", s.handler.Price.Update)
	prices.DELETE("/:id", s.handler.Price.Delete)
	products.GET("/product/:productId", s.handler.Price.ListByProduct)

	// Customer routes
	customers := v1.Group("/customers")
	customers.POST("", s.handler.Customer.Create)
	customers.GET("", s.handler.Customer.List)
	customers.GET("/:id", s.handler.Customer.Get)
	customers.GET("/email/:email", s.handler.Customer.GetByEmail)
	customers.PUT("/:id", s.handler.Customer.Update)
	customers.DELETE("/:id", s.handler.Customer.Delete)

	// Subscription routes
	subscriptions := v1.Group("/subscriptions")
	subscriptions.POST("", s.handler.Subscription.Create)
	subscriptions.GET("", s.handler.Subscription.List)
	subscriptions.GET("/:id", s.handler.Subscription.Get)
	subscriptions.PUT("/:id", s.handler.Subscription.Update)
	subscriptions.POST("/:id/cancel", s.handler.Subscription.Cancel)
	subscriptions.POST("/:id/pause", s.handler.Subscription.Pause)
	subscriptions.POST("/:id/resume", s.handler.Subscription.Resume)
	customers.GET("/:customerId/subscriptions", s.handler.Subscription.ListByCustomer)

	// Stripe Checkout routes
	checkout := v1.Group("/checkout")
	checkout.POST("/create-session", s.handler.Checkout.CreateSession)
	checkout.POST("/create-multi-item-session", s.handler.Checkout.CreateMultiItemSession)
	checkout.GET("/verify-session", s.handler.Checkout.VerifySession)

	// Webhook route - no authentication middleware for this route
	v1.POST("/webhooks/stripe", s.handler.Webhook.HandleWebhook)

	// Health check route
	s.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}
