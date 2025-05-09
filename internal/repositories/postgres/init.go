// internal/repositories/postgres/init.go
package postgres

import (
	"database/sql"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog"
)

// DB represents a PostgreSQL database connection pool
type DB struct {
	*sql.DB
}

// Repositories holds all repository instances
type Repositories struct {
	Product      interfaces.ProductRepository
	Price        interfaces.PriceRepository
	Customer     interfaces.CustomerRepository
	Subscription interfaces.SubscriptionRepository
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	// Wrap in our custom DB type
	db := &DB{
		DB: sqlDB,
	}

	return db, nil
}

// NewRepositories initializes all repositories
func NewRepositories(db *DB, logger *zerolog.Logger) *Repositories {
	sublogger := logger.With().Str("component", "repository").Logger()
	return &Repositories{
		Product:      NewProductRepository(db, sublogger),
		Price:        NewPriceRepository(db),
		Customer:     NewCustomerRepository(db),
		Subscription: NewSubscriptionRepository(db),
	}
}

// Close closes the database connection
func Close(db *DB) error {
	return db.Close()
}
