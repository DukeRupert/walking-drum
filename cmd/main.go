package main

import (
	"database/sql"
	"embed"
	"fmt"

	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dukerupert/walking-drum/handlers"
	"github.com/dukerupert/walking-drum/middleware"
	"github.com/dukerupert/walking-drum/repository"
	"github.com/dukerupert/walking-drum/services"
	"github.com/dukerupert/walking-drum/services/payment"
)

func init() {
	// Set up pretty logging for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set log level based on environment (DEBUG, INFO, etc.)
	// For troubleshooting, set to Debug level
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Debug().Msg("Logger initialized at debug level")
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOSTNAME")

	log.Debug().
		Str("dbUser", dbUser).
		Str("dbPassword", "***").
		Str("dbName", dbName).
		Str("dbHost", dbHost).
		Msg("Database connection parameters")

	// Create PostgreSQL connection string
	dbConnectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName,
	)

	// Connect to the database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}
	log.Info().Msg("Successfully connected to the database")

	// Configure Goose with embedded migrations
	goose.SetBaseFS(embedMigrations)

	// Set Goose's database dialect
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal().Err(err).Msg("Failed to set dialect")
	}

	// Run migrations
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Migrations completed successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	cartRepo := repository.NewCartRepository(db)
	cartItemRepo := repository.NewCartItemRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)

	// Initialize services
	paymentProcessor := payment.NewStripeProcessor()
	paymentMethodService := payment.NewPaymentMethodService(paymentProcessor)
	subscriptionService := services.NewSubscriptionService(
		paymentProcessor,
		subscriptionRepo,
		userRepo,
		priceRepo,
		productRepo,
		invoiceRepo,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	priceHandler := handlers.NewPriceHandler(priceRepo, productRepo)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo, userRepo, subscriptionRepo)
	cartHandler := handlers.NewCartHandler(cartRepo, cartItemRepo, productRepo, priceRepo)
	orderHandler := handlers.NewOrderHandler(
		orderRepo,
		orderItemRepo,
		cartRepo,
		cartItemRepo,
		userRepo,
		productRepo,
		priceRepo,
		subscriptionRepo,
	)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)
	paymentMethodHandler := handlers.NewPaymentMethodHandler(paymentMethodService)
	webhookHandler := payment.NewWebhookHandler(
		paymentProcessor,
		subscriptionRepo,
		invoiceRepo,
		userRepo,
		productRepo,
		priceRepo,
	)

	// Set up router
	router := mux.NewRouter()

	// Use the router
	http.Handle("/", router)

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.EnableCORS)

	// Register user routes
	apiRouter.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	apiRouter.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	apiRouter.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Register product routes
	apiRouter.HandleFunc("/products", productHandler.CreateProduct).Methods("POST")
	apiRouter.HandleFunc("/products", productHandler.ListProducts).Methods("GET")
	apiRouter.HandleFunc("/products/{id}", productHandler.GetProduct).Methods("GET")
	apiRouter.HandleFunc("/products/{id}", productHandler.UpdateProduct).Methods("PUT")
	apiRouter.HandleFunc("/products/{id}", productHandler.DeleteProduct).Methods("DELETE")

	// Register price routes
	apiRouter.HandleFunc("/prices", priceHandler.CreatePrice).Methods("POST")
	apiRouter.HandleFunc("/prices", priceHandler.ListPrices).Methods("GET")
	apiRouter.HandleFunc("/prices/{id}", priceHandler.GetPrice).Methods("GET")
	apiRouter.HandleFunc("/prices/{id}", priceHandler.UpdatePrice).Methods("PUT")
	apiRouter.HandleFunc("/prices/{id}", priceHandler.DeletePrice).Methods("DELETE")

	// Register invoice routes
	apiRouter.HandleFunc("/invoices", invoiceHandler.CreateInvoice).Methods("POST")
	apiRouter.HandleFunc("/invoices", invoiceHandler.ListInvoices).Methods("GET")
	apiRouter.HandleFunc("/invoices/{id}", invoiceHandler.GetInvoice).Methods("GET")
	apiRouter.HandleFunc("/invoices/{id}", invoiceHandler.UpdateInvoice).Methods("PUT")
	apiRouter.HandleFunc("/invoices/{id}/mark-as-paid", invoiceHandler.MarkAsPaid).Methods("POST")
	apiRouter.HandleFunc("/invoices/{id}", invoiceHandler.DeleteInvoice).Methods("DELETE")

	// Register cart routes
	apiRouter.HandleFunc("/carts", cartHandler.CreateCart).Methods("POST")
	apiRouter.HandleFunc("/carts/{id}", cartHandler.GetCart).Methods("GET")
	apiRouter.HandleFunc("/users/{user_id}/cart", cartHandler.GetUserCart).Methods("GET")
	apiRouter.HandleFunc("/sessions/{session_id}/cart", cartHandler.GetSessionCart).Methods("GET")
	apiRouter.HandleFunc("/carts/{id}/items", cartHandler.AddToCart).Methods("POST")
	apiRouter.HandleFunc("/carts/{id}/items/{item_id}", cartHandler.UpdateCartItem).Methods("PUT")
	apiRouter.HandleFunc("/carts/{id}/items/{item_id}", cartHandler.RemoveCartItem).Methods("DELETE")
	apiRouter.HandleFunc("/carts/{id}/clear", cartHandler.ClearCart).Methods("POST")
	apiRouter.HandleFunc("/carts/{id}", cartHandler.DeleteCart).Methods("DELETE")
	apiRouter.HandleFunc("/carts/clean-expired", cartHandler.CleanExpiredCarts).Methods("POST")

	// Register order routes
	apiRouter.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	apiRouter.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")
	apiRouter.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods("GET")
	apiRouter.HandleFunc("/orders/{id}", orderHandler.UpdateOrder).Methods("PUT")
	apiRouter.HandleFunc("/orders/{id}/cancel", orderHandler.CancelOrder).Methods("POST")
	apiRouter.HandleFunc("/orders/{id}", orderHandler.DeleteOrder).Methods("DELETE")
	apiRouter.HandleFunc("/users/{user_id}/orders", orderHandler.ListUserOrders).Methods("GET")

	// Register webhook route
	apiRouter.HandleFunc("/webhooks/stripe", webhookHandler.HandleWebhook).Methods("POST")

	// Register subscription routes
	apiRouter.HandleFunc("/subscriptions", subscriptionHandler.CreateSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions", subscriptionHandler.ListSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/{id}", subscriptionHandler.GetSubscription).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/user/{userID}", subscriptionHandler.GetUserSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/customer/{customerID}", subscriptionHandler.GetCustomerSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/{id}", subscriptionHandler.UpdateSubscription).Methods("PUT")
	apiRouter.HandleFunc("/subscriptions/{id}/cancel", subscriptionHandler.CancelSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions/{id}/reactivate", subscriptionHandler.ReactivateSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions/{id}/pause", subscriptionHandler.PauseSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions/{id}/resume", subscriptionHandler.ResumeSubscription).Methods("POST")

	// Register payment method routes
	apiRouter.HandleFunc("/payment-methods", paymentMethodHandler.CreatePaymentMethod).Methods("POST")
	// Create payment method
	apiRouter.HandleFunc("/payment-methods", paymentMethodHandler.CreatePaymentMethod).Methods("POST")
	// List payment methods for a customer
	apiRouter.HandleFunc("/payment-methods/customer/{customerID}", paymentMethodHandler.ListPaymentMethods).Methods("GET")
	// Get a specific payment method
	apiRouter.HandleFunc("/payment-methods/{id}", paymentMethodHandler.GetPaymentMethod).Methods("GET")
	// Update a payment method
	apiRouter.HandleFunc("/payment-methods/{id}", paymentMethodHandler.UpdatePaymentMethod).Methods("PUT")
	// Set payment method as default
	apiRouter.HandleFunc("/payment-methods/{id}/default/{customerID}", paymentMethodHandler.SetDefaultPaymentMethod).Methods("PUT")
	// Detach a payment method
	apiRouter.HandleFunc("/payment-methods/{id}", paymentMethodHandler.DetachPaymentMethod).Methods("DELETE")

	// Start the HTTP server
	port := "8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
