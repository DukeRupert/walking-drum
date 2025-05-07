// internal/services/product_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ProductService defines the interface for product business logic
type ProductService interface {
    Create(ctx context.Context, productDTO *dto.ProductCreateDTO) (*models.Product, error)
    GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
    List(ctx context.Context, page, pageSize int, includeInactive bool) ([]*models.Product, int, error)
    Update(ctx context.Context, id uuid.UUID, productDTO *dto.ProductUpdateDTO) (*models.Product, error)
    Delete(ctx context.Context, id uuid.UUID) error
    UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error
}

// productService is the private implementation of ProductService
type productService struct {
    productRepo  interfaces.ProductRepository
    stripeClient *stripe.Client
	logger		 zerolog.Logger
}

// NewProductService creates a new instance of ProductService
func NewProductService(repo interfaces.ProductRepository, stripe *stripe.Client, logger zerolog.Logger) ProductService {
    return &productService{
        productRepo:  repo,
        stripeClient: stripe,
		logger:	logger.With().Str("component", "product_service").Logger(),
    }
}

// Create adds a new product to the system (both in DB and Stripe)
func (s *productService) Create(ctx context.Context, productDTO *dto.ProductCreateDTO) (*models.Product, error) {
    // 1. Validate productDTO
    if problems := productDTO.Valid(ctx); len(problems) > 0 {
        return nil, fmt.Errorf("invalid product data: %v", problems)
    }

    // 2. Create product in Stripe first
    stripeProduct, err := s.stripeClient.CreateProduct(ctx, &stripe.ProductCreateParams{
        Name:        productDTO.Name,
        Description: productDTO.Description,
        Images:      []string{productDTO.ImageURL},
        Active:      productDTO.Active,
        Metadata: map[string]string{
            "origin":       productDTO.Origin,
            "roast_level":  productDTO.RoastLevel,
            "flavor_notes": productDTO.FlavorNotes,
            "weight":       fmt.Sprintf("%d", productDTO.Weight),
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create product in Stripe: %w", err)
    }

    // 3. Create product in local database
    now := time.Now()
    product := &models.Product{
        ID:          uuid.New(),
        Name:        productDTO.Name,
        Description: productDTO.Description,
        ImageURL:    productDTO.ImageURL,
        Active:      productDTO.Active,
        StockLevel:  productDTO.StockLevel,
        Weight:      productDTO.Weight,
        Origin:      productDTO.Origin,
        RoastLevel:  productDTO.RoastLevel,
        FlavorNotes: productDTO.FlavorNotes,
        StripeID:    stripeProduct.ID,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // 4. Save to database
    if err := s.productRepo.Create(ctx, product); err != nil {
        // If database creation fails, archive the Stripe product
        if archiveErr := s.stripeClient.ArchiveProduct(ctx, stripeProduct.ID); archiveErr != nil {
            // Log this error, but continue with the main error
            s.logger.Error().Err(archiveErr).Msg("Failed to archive Stripe product after database error")
        }
        return nil, fmt.Errorf("failed to create product in database: %w", err)
    }

    return product, nil
}

// GetByID retrieves a product by its ID
func (s *productService) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	// TODO: Implement get product by ID
	// 1. Call repository to fetch product
	return nil, nil
}

// List retrieves all products with optional filtering
func (s *productService) List(ctx context.Context, page, pageSize int, includeInactive bool) ([]*models.Product, int, error) {
	// TODO: Implement product listing
	// 1. Calculate offset from page and pageSize
	// 2. Call repository to list products
	return nil, 0, nil
}

// Update updates an existing product
func (s *productService) Update(ctx context.Context, id uuid.UUID, productDTO *dto.ProductUpdateDTO) (*models.Product, error) {
	// TODO: Implement product update
	// 1. Get existing product
	// 2. Update fields from DTO
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil, nil
}

// Delete removes a product from the system
func (s *productService) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement product deletion
	// 1. Get existing product
	// 2. Archive in Stripe
	// 3. Delete from database
	// 4. Handle errors
	return nil
}

// UpdateStockLevel updates the stock level of a product
func (s *productService) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	// TODO: Implement stock level update
	// 1. Call repository to update stock level
	return nil
}