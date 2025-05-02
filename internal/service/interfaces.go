package service

import (
	"context"
	
	"github.com/dukerupert/walking-drum/internal/models"
)

// ProductService defines business logic for products
type iProductService interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProduct(ctx context.Context, id int64) (*models.Product, error)
	ListActiveProducts(ctx context.Context) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id int64) error
	
	// Price management
	AddPrice(ctx context.Context, productID int64, price *models.ProductPrice) error
	UpdatePrice(ctx context.Context, price *models.ProductPrice) error
	RemovePrice(ctx context.Context, priceID int64) error
	
	// Stripe synchronization
	SyncFromStripe(ctx context.Context, stripeProductID string) error
	SyncToStripe(ctx context.Context, productID int64) error
}

// CustomerService defines business logic for customers
type CustomerService interface {
	RegisterCustomer(ctx context.Context, customer *models.Customer) error
	GetCustomer(ctx context.Context, id int64) (*models.Customer, error)
	GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer) error
	
	// Address management
	AddAddress(ctx context.Context, customerID int64, address *models.CustomerAddress) error
	UpdateAddress(ctx context.Context, address *models.CustomerAddress) error
	RemoveAddress(ctx context.Context, addressID int64) error
	SetDefaultAddress(ctx context.Context, customerID int64, addressID int64) error
	
	// Stripe synchronization
	SyncFromStripe(ctx context.Context, stripeCustomerID string) error
	SyncToStripe(ctx context.Context, customerID int64) error
}

// SubscriptionService defines business logic for subscriptions
type SubscriptionService interface {
	CreateSubscription(ctx context.Context, customerID int64, priceID int64) (*models.Subscription, error)
	GetSubscription(ctx context.Context, id int64) (*models.Subscription, error)
	ListCustomerSubscriptions(ctx context.Context, customerID int64) ([]*models.Subscription, error)
	
	// Subscription management
	PauseSubscription(ctx context.Context, id int64) error
	ResumeSubscription(ctx context.Context, id int64) error
	CancelSubscription(ctx context.Context, id int64, cancelAtPeriodEnd bool) error
	UpdateSubscriptionPrice(ctx context.Context, id int64, newPriceID int64) error
	
	// Stripe synchronization
	SyncFromStripe(ctx context.Context, stripeSubscriptionID string) error
	SyncToStripe(ctx context.Context, subscriptionID int64) error
}

// WebhookService defines business logic for webhook event handling
type WebhookService interface {
	ProcessEvent(ctx context.Context, stripeEventID string, eventType string, eventData []byte) error
	HandleSubscriptionUpdated(ctx context.Context, eventData []byte) error
	HandleSubscriptionDeleted(ctx context.Context, eventData []byte) error
	HandlePaymentSucceeded(ctx context.Context, eventData []byte) error
	HandlePaymentFailed(ctx context.Context, eventData []byte) error
}