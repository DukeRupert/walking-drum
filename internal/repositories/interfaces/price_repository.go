// internal/repositories/interfaces/price_repository.go
package interfaces

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
)

// PriceRepository defines the interface for price data access
type PriceRepository interface {
	// Create adds a new price to the database
	Create(ctx context.Context, price *models.Price) error

	// GetByID retrieves a price by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error)

	// GetByStripeID retrieves a price by its Stripe ID
	GetByStripeID(ctx context.Context, stripeID string) (*models.Price, error)

	// List retrieves all prices, with optional filtering
	List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Price, int, error)

	// ListByProductID retrieves all prices for a specific product
	ListByProductID(ctx context.Context, productID uuid.UUID, includeInactive bool) ([]*models.Price, error)

	// Update updates an existing price
	Update(ctx context.Context, price *models.Price) error

	// Delete removes a price from the database
	Delete(ctx context.Context, id uuid.UUID) error
}
