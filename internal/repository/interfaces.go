// File: internal/repository/interfaces.go
package repository

import (
	"context"
	
	"github.com/dukerupert/walking-drum/internal/models"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	ListActive(ctx context.Context) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetWithPrices(ctx context.Context, id int64) (*domain.Product, error)
	ListActiveWithPrices(ctx context.Context) ([]*domain.Product, error)
}

// ProductPriceRepository defines the interface for product price data access
type ProductPriceRepository interface {
	Create(ctx context.Context, price *domain.ProductPrice) error
	GetByID(ctx context.Context, id int64) (*domain.ProductPrice, error)
	ListByProductID(ctx context.Context, productID int64) ([]*domain.ProductPrice, error)
	Update(ctx context.Context, price *domain.ProductPrice) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetDefaultForProduct(ctx context.Context, productID int64) (*domain.ProductPrice, error)
}

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id int64) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	GetByStripeID(ctx context.Context, stripeCustomerID string) (*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetWithAddresses(ctx context.Context, id int64) (*domain.Customer, error)
	GetWithSubscriptions(ctx context.Context, id int64) (*domain.Customer, error)
}

// CustomerAddressRepository defines the interface for customer address data access
type CustomerAddressRepository interface {
	Create(ctx context.Context, address *domain.CustomerAddress) error
	GetByID(ctx context.Context, id int64) (*domain.CustomerAddress, error)
	ListByCustomerID(ctx context.Context, customerID int64) ([]*domain.CustomerAddress, error)
	Update(ctx context.Context, address *domain.CustomerAddress) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetDefaultForCustomer(ctx context.Context, customerID int64) (*domain.CustomerAddress, error)
	SetAsDefault(ctx context.Context, id int64, customerID int64) error
}

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *domain.Subscription) error
	GetByID(ctx context.Context, id int64) (*domain.Subscription, error)
	GetByStripeID(ctx context.Context, stripeSubscriptionID string) (*domain.Subscription, error)
	ListByCustomerID(ctx context.Context, customerID int64) ([]*domain.Subscription, error)
	ListByStatus(ctx context.Context, status string) ([]*domain.Subscription, error)
	Update(ctx context.Context, subscription *domain.Subscription) error
	
	// Additional methods as needed
	GetWithDetails(ctx context.Context, id int64) (*domain.Subscription, error)
	ListActiveWithDetails(ctx context.Context) ([]*domain.Subscription, error)
}

// WebhookEventRepository defines the interface for webhook event data access
type WebhookEventRepository interface {
	Create(ctx context.Context, event *domain.WebhookEvent) error
	GetByStripeID(ctx context.Context, stripeEventID string) (*domain.WebhookEvent, error)
	MarkAsProcessed(ctx context.Context, id int64) error
	ListUnprocessed(ctx context.Context) ([]*domain.WebhookEvent, error)
}