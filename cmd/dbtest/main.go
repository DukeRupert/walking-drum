package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dukerupert/walking-drum/pkg/config"
	"github.com/dukerupert/walking-drum/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	log.Println("Connecting to database...")
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	err = db.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database version: %v", err)
	}

	fmt.Printf("Successfully connected to PostgreSQL!\nVersion: %s\n", version)

	// Simple table check
	var tableCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&tableCount)
	if err != nil {
		log.Fatalf("Failed to check tables: %v", err)
	}

	fmt.Printf("Number of tables in public schema: %d\n", tableCount)

	// Check which expected tables exist
	tables := []string{"products", "product_prices", "customers", "customer_addresses", "subscriptions", "webhook_events"}
	fmt.Println("\nTable existence check:")
	
	for _, table := range tables {
		var exists bool
		err = db.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`, table).Scan(&exists)
		
		if err != nil {
			log.Fatalf("Failed to check if %s table exists: %v", table, err)
		}
		
		status := "✅ exists"
		if !exists {
			status = "❌ does not exist"
		}
		fmt.Printf("  - %s: %s\n", table, status)
	}
}