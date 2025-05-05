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

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// API group
	api := s.echo.Group("/api")
	
	// Product routes
	products := api.Group("/products")
	products.POST("", s.productHandler.Create)
	products.GET("", s.productHandler.List)
	products.GET("/:id", s.productHandler.Get)
	products.PUT("/:id", s.productHandler.Update)
	products.DELETE("/:id", s.productHandler.Delete)
	products.PATCH("/:id/stock", s.productHandler.UpdateStockLevel)
	
	// Price routes
	prices := api.Group("/prices")
	prices.POST("", s.priceHandler.Create)
	prices.GET("", s.priceHandler.List)
	prices.GET("/:id", s.priceHandler.Get)
	prices.PUT("/:id", s.priceHandler.Update)
	prices.DELETE("/:id", s.priceHandler.Delete)
	products.GET("/:productId/prices", s.priceHandler.ListByProduct)
	
	// Customer routes
	customers := api.Group("/customers")
	customers.POST("", s.customerHandler.Create)
	customers.GET("", s.customerHandler.List)
	customers.GET("/:id", s.customerHandler.Get)
	customers.GET("/email/:email", s.customerHandler.GetByEmail)
	customers.PUT("/:id", s.customerHandler.Update)
	customers.DELETE("/:id", s.customerHandler.Delete)
	
	// Subscription routes
	subscriptions := api.Group("/subscriptions")
	subscriptions.POST("", s.subscriptionHandler.Create)
	subscriptions.GET("", s.subscriptionHandler.List)
	subscriptions.GET("/:id", s.subscriptionHandler.Get)
	subscriptions.PUT("/:id", s.subscriptionHandler.Update)
	subscriptions.POST("/:id/cancel", s.subscriptionHandler.Cancel)
	subscriptions.POST("/:id/pause", s.subscriptionHandler.Pause)
	subscriptions.POST("/:id/resume", s.subscriptionHandler.Resume)
	subscriptions.POST("/:id/change-product", s.subscriptionHandler.ChangeProduct)
	subscriptions.POST("/:id/change-price", s.subscriptionHandler.ChangePrice)
	subscriptions.POST("/:id/change-quantity", s.subscriptionHandler.ChangeQuantity)
	subscriptions.POST("/:id/change-address", s.subscriptionHandler.ChangeAddress)
	customers.GET("/:customerId/subscriptions", s.subscriptionHandler.ListByCustomer)
	customers.GET("/:customerId/subscriptions/active", s.subscriptionHandler.ListActiveByCustomer)
	
	// Health check route
	s.echo.GET("/healthz", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf(":%d", s.config.App.Port))
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}