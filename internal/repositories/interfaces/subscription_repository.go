// internal/repositories/interfaces/subscription_repository.go
package interfaces

import (
	"context"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
)

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	// Create adds a new subscription to the database
	Create(ctx context.Context, subscription *models.Subscription) error

	// GetByID retrieves a subscription by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)

	// GetByStripeID retrieves a subscription by its Stripe ID
	GetByStripeID(ctx context.Context, stripeID string) (*models.Subscription, error)

	// List retrieves all subscriptions with optional pagination
	List(ctx context.Context, offset, limit int) ([]*models.Subscription, int, error)

	// ListByCustomerID retrieves all subscriptions for a specific customer
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error)

	// ListActiveByCustomerID retrieves active subscriptions for a specific customer
	ListActiveByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error)

	// ListByStatus retrieves subscriptions filtered by status
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.Subscription, int, error)

	// ListDueForRenewal retrieves subscriptions due for renewal in the given time period
	ListDueForRenewal(ctx context.Context, before time.Time) ([]*models.Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *models.Subscription) error

	// UpdateStatus updates only the status of a subscription
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error

	// UpdatePeriod updates the current period information for a subscription
	UpdatePeriod(ctx context.Context, id uuid.UUID, startDate, endDate time.Time) error

	// SetCancelAtPeriodEnd marks a subscription to be cancelled at the end of the period
	SetCancelAtPeriodEnd(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error

	// Delete removes a subscription from the database
	Delete(ctx context.Context, id uuid.UUID) error

	// GetWithRelatedData retrieves a subscription with its related product, price, customer and address data
	GetWithRelatedData(ctx context.Context, id uuid.UUID) (*models.Subscription, *models.Product, *models.Price, *models.Customer, *models.Address, error)
}

