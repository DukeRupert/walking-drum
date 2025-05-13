// internal/repositories/interfaces/subscription_repository.go
package interfaces

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
)

type VariantRepository interface {
	// Get a variant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Variant, error)

	// Get variants by product ID
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Variant, error)

	// Get variant by product, weight, and grind
	GetByAttributes(ctx context.Context, productID uuid.UUID, weight string, grind string) (*models.Variant, error)

	// Get all variants with pagination
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]*models.Variant, int, error)

	// Create a new variant
	Create(ctx context.Context, variant *models.Variant) error

	// Update an existing variant
	Update(ctx context.Context, variant *models.Variant) error

	// Delete a variant
	Delete(ctx context.Context, id uuid.UUID) error

	// Update stock level
	UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error
}
