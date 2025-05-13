// internal/repositories/postgres/db.go
package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog"
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	*sql.DB
	*zerolog.Logger
}

// Repositories holds all repository instances
type Repositories struct {
	Product      interfaces.ProductRepository
	Variant      interfaces.VariantRepository
	Price        interfaces.PriceRepository
	Customer     interfaces.CustomerRepository
	Subscription interfaces.SubscriptionRepository
}

// NewRepositories initializes all repositories
func CreateRepositories(db *DB, logger *zerolog.Logger) *Repositories {
	return &Repositories{
		Product:      NewProductRepository(db, logger),
		Variant:      NewVariantRepository(db, logger),
		Price:        NewPriceRepository(db),
		Customer:     NewCustomerRepository(db),
		Subscription: NewSubscriptionRepository(db),
	}
}

// Close closes the database connection
func Close(db *DB) error {
	return db.Close()
}

// Connect establishes a connection to the PostgreSQL database
func  Connect(dbConfig config.DBConfig, logger *zerolog.Logger) (*DB, error) {
	// Initialize connection to PostgreSQL server (not a specific DB)
	baseConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.SSLMode)

	// Connect to 'postgres' default database first
	baseDB, err := sql.Open("postgres", baseConnStr+" dbname=postgres")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to postgres database")
	}
	defer baseDB.Close()

	// Check if our database exists
	var exists bool
	err = baseDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", "coffee_subscriptions").Scan(&exists)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to check if database exists")
	}

	// Create the database if it doesn't exist
	if !exists {
		_, err = baseDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbConfig.Name))
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create database")
		}
		logger.Info().Msg(fmt.Sprintf("Created database %s", dbConfig.Name))
	}

	// Create the database connection
	db, err := sql.Open("postgres", dbConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is working
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Debug().Str("name", dbConfig.Name).Msg("Connected to PostgreSQL database")

	sublogger := logger.With().Str("component", "database").Logger()

	return &DB{db, &sublogger}, nil
}

// Close closes the database connection pool
func (db *DB) Close() error {
	return db.DB.Close()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(txFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, rollback the transaction
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			// An error occurred, rollback the transaction
			tx.Rollback()
		} else {
			// All good, commit the transaction
			err = tx.Commit()
		}
	}()

	err = txFunc(tx)
	return err
}
