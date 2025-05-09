// internal/repositories/postgres/init.go
package postgres

import (
	"database/sql"

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
	Price        interfaces.PriceRepository
	Customer     interfaces.CustomerRepository
	Subscription interfaces.SubscriptionRepository
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
