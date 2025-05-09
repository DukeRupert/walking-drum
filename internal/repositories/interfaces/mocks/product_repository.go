// internal/repositories/interfaces/mocks/product_repository.go
package mocks

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ProductRepository is a mock implementation of interfaces.ProductRepository
type ProductRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	
	// Handle the first return value (could be nil)
	var product *models.Product
	if args.Get(0) != nil {
		product = args.Get(0).(*models.Product)
	}
	
	return product, args.Error(1)
}

// GetByStripeID mocks the GetByStripeID method
func (m *ProductRepository) GetByStripeID(ctx context.Context, stripeID string) (*models.Product, error) {
	args := m.Called(ctx, stripeID)
	
	// Handle the first return value (could be nil)
	var product *models.Product
	if args.Get(0) != nil {
		product = args.Get(0).(*models.Product)
	}
	
	return product, args.Error(1)
}

// List mocks the List method
func (m *ProductRepository) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Product, int, error) {
	args := m.Called(ctx, offset, limit, includeInactive)
	
	// Handle the first return value (could be nil)
	var products []*models.Product
	if args.Get(0) != nil {
		products = args.Get(0).([]*models.Product)
	}
	
	return products, args.Int(1), args.Error(2)
}

// Update mocks the Update method
func (m *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UpdateStockLevel mocks the UpdateStockLevel method
func (m *ProductRepository) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}