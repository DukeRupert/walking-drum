package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"

	"github.com/dukerupert/walking-drum/handlers"
	"github.com/dukerupert/walking-drum/repository"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOSTNAME")

	// Create PostgreSQL connection string
	dbConnectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName,
	)

	// Connect to the database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to the database")

	// Configure Goose with embedded migrations
	goose.SetBaseFS(embedMigrations)

	// Set Goose's database dialect
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Run migrations
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	cartRepo := repository.NewCartRepository(db)
	cartItemRepo := repository.NewCartItemRepository(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	priceHandler := handlers.NewPriceHandler(priceRepo, productRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionRepo, userRepo, priceRepo)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo, userRepo, subscriptionRepo)
	cartHandler := handlers.NewCartHandler(cartRepo, cartItemRepo, productRepo, priceRepo)

	// Set up router
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

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

	// Register subscription routes
	apiRouter.HandleFunc("/subscriptions", subscriptionHandler.CreateSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions", subscriptionHandler.ListSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/{id}", subscriptionHandler.GetSubscription).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/{id}", subscriptionHandler.UpdateSubscription).Methods("PUT")
	apiRouter.HandleFunc("/subscriptions/{id}/cancel", subscriptionHandler.CancelSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions/{id}", subscriptionHandler.DeleteSubscription).Methods("DELETE")

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

	// Use the router
	http.Handle("/", router)

	// Start the HTTP server
	port := "8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
