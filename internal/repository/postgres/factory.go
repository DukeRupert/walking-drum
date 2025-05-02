// File: internal/repository/postgres/factory.go
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/dukerupert/walking-drum/internal/repository"
)

// RepositoryFactory creates and returns all repository instances
type RepositoryFactory struct {
	db *pgxpool.Pool
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(db *pgxpool.Pool) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// ProductRepository returns a new product repository
func (f *RepositoryFactory) ProductRepository() repository.ProductRepository {
	return NewProductRepository(f.db)
}

// ProductPriceRepository returns a new product price repository
func (f *RepositoryFactory) ProductPriceRepository() repository.ProductPriceRepository {
	return NewProductPriceRepository(f.db)
}

// CustomerRepository returns a new customer repository
func (f *RepositoryFactory) CustomerRepository() repository.CustomerRepository {
	return NewCustomerRepository(f.db)
}

// CustomerAddressRepository returns a new customer address repository
func (f *RepositoryFactory) CustomerAddressRepository() repository.CustomerAddressRepository {
	return NewCustomerAddressRepository(f.db)
}

// SubscriptionRepository returns a new subscription repository
func (f *RepositoryFactory) SubscriptionRepository() repository.SubscriptionRepository {
	return NewSubscriptionRepository(f.db)
}

// WebhookEventRepository returns a new webhook event repository
func (f *RepositoryFactory) WebhookEventRepository() repository.WebhookEventRepository {
	return NewWebhookEventRepository(f.db)
}