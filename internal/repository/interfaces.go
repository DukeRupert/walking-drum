// File: internal/repository/interfaces.go
package repository

import (
	"context"
	
	"github.com/dukerupert/walking-drum/internal/models"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	ListActive(ctx context.Context) ([]*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetWithPrices(ctx context.Context, id int64) (*models.Product, error)
	ListActiveWithPrices(ctx context.Context) ([]*models.Product, error)
	GetByStripeID(ctx context.Context, stripeID string) (*models.Product, error)
}

// ProductPriceRepository defines the interface for product price data access
type ProductPriceRepository interface {
	Create(ctx context.Context, price *models.ProductPrice) error
	GetByID(ctx context.Context, id int64) (*models.ProductPrice, error)
	ListByProductID(ctx context.Context, productID int64) ([]*models.ProductPrice, error)
	Update(ctx context.Context, price *models.ProductPrice) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetDefaultForProduct(ctx context.Context, productID int64) (*models.ProductPrice, error)
	GetByStripeID(ctx context.Context, stripeID string) (*models.ProductPrice, error)
}

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	Create(ctx context.Context, customer *models.Customer) error
	GetByID(ctx context.Context, id int64) (*models.Customer, error)
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)
	GetByStripeID(ctx context.Context, stripeCustomerID string) (*models.Customer, error)
	Update(ctx context.Context, customer *models.Customer) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetWithAddresses(ctx context.Context, id int64) (*models.Customer, error)
	GetWithSubscriptions(ctx context.Context, id int64) (*models.Customer, error)
}

// CustomerAddressRepository defines the interface for customer address data access
type CustomerAddressRepository interface {
	Create(ctx context.Context, address *models.CustomerAddress) error
	GetByID(ctx context.Context, id int64) (*models.CustomerAddress, error)
	ListByCustomerID(ctx context.Context, customerID int64) ([]*models.CustomerAddress, error)
	Update(ctx context.Context, address *models.CustomerAddress) error
	Delete(ctx context.Context, id int64) error
	
	// Additional methods as needed
	GetDefaultForCustomer(ctx context.Context, customerID int64) (*models.CustomerAddress, error)
	SetAsDefault(ctx context.Context, id int64, customerID int64) error
}

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, id int64) (*models.Subscription, error)
	GetByStripeID(ctx context.Context, stripeSubscriptionID string) (*models.Subscription, error)
	ListByCustomerID(ctx context.Context, customerID int64) ([]*models.Subscription, error)
	ListByStatus(ctx context.Context, status string) ([]*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	
	// Additional methods as needed
	GetWithDetails(ctx context.Context, id int64) (*models.Subscription, error)
	ListActiveWithDetails(ctx context.Context) ([]*models.Subscription, error)
}

// WebhookEventRepository defines the interface for webhook event data access
type WebhookEventRepository interface {
	Create(ctx context.Context, event *models.WebhookEvent) error
	GetByStripeID(ctx context.Context, stripeEventID string) (*models.WebhookEvent, error)
	MarkAsProcessed(ctx context.Context, id int64) error
	ListUnprocessed(ctx context.Context) ([]*models.WebhookEvent, error)
}