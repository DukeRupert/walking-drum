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
	variantService variantService
	stripeService stripe.StripeService
	logger       zerolog.Logger
}

// NewProductService creates a new instance of ProductService
func NewProductService(repo interfaces.ProductRepository, stripe stripe.StripeService, logger *zerolog.Logger) ProductService {
	return &productService{
		productRepo:  repo,
		stripeService: stripe,
		logger:       logger.With().Str("component", "product_service").Logger(),
	}
}

// Create adds a new product to the system (both in DB and Stripe)
func (s *productService) Create(ctx context.Context, productDTO *dto.ProductCreateDTO) (*models.Product, error) {
	s.logger.Debug().
		Str("function", "productService.Create").
		Interface("productDTO", productDTO).
		Msg("Starting product creation")

	// 1. Validate productDTO
	if problems := productDTO.Valid(ctx); len(problems) > 0 {
		s.logger.Error().
			Str("function", "productService.Create").
			Interface("problems", problems).
			Msg("Product validation failed")
		return nil, fmt.Errorf("invalid product data: %v", problems)
	}

	s.logger.Debug().
		Str("function", "productService.Create").
		Msg("Product validation passed")

	// 2. Create product in Stripe first
	s.logger.Debug().
		Str("function", "productService.Create").
		Str("name", productDTO.Name).
		Str("description", productDTO.Description).
		Str("imageURL", productDTO.ImageURL).
		Bool("active", productDTO.Active).
		Msg("Creating product in Stripe")

	stripeProduct, err := s.stripeService.CreateProduct(ctx, &stripe.ProductCreateParams{
		Name:        productDTO.Name,
		Description: productDTO.Description,
		Images:      []string{productDTO.ImageURL},
		Active:      productDTO.Active,
		Metadata: map[string]string{
			"origin":              productDTO.Origin,
			"roast_level":         productDTO.RoastLevel,
			"flavor_notes":        productDTO.FlavorNotes,
			"weight":              fmt.Sprintf("%d", productDTO.Weight),
			"allow_subscription":  fmt.Sprintf("%t", productDTO.AllowSubscription),
		},
	})
	if err != nil {
		s.logger.Error().
			Str("function", "productService.Create").
			Err(err).
			Msg("Failed to create product in Stripe")
		return nil, fmt.Errorf("failed to create product in Stripe: %w", err)
	}

	s.logger.Debug().
		Str("function", "productService.Create").
		Str("stripeProductID", stripeProduct.ID).
		Msg("Successfully created product in Stripe")

	// 3. Create product in local database
	now := time.Now()
	product := &models.Product{
		ID:                uuid.New(),
		Name:              productDTO.Name,
		Description:       productDTO.Description,
		ImageURL:          productDTO.ImageURL,
		Active:            productDTO.Active,
		StockLevel:        productDTO.StockLevel,
		Weight:            productDTO.Weight,
		Origin:            productDTO.Origin,
		RoastLevel:        productDTO.RoastLevel,
		FlavorNotes:       productDTO.FlavorNotes,
		Options:           productDTO.Options,
		AllowSubscription: productDTO.AllowSubscription,
		StripeID:          stripeProduct.ID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.logger.Debug().
		Str("function", "productService.Create").
		Str("productID", product.ID.String()).
		Str("stripeID", product.StripeID).
		Msg("Preparing to save product to database")

	// 4. Save to database
	if err := s.productRepo.Create(ctx, product); err != nil {
		s.logger.Error().
			Str("function", "productService.Create").
			Err(err).
			Str("productID", product.ID.String()).
			Msg("Failed to create product in database")

		// If database creation fails, archive the Stripe product
		s.logger.Debug().
			Str("function", "productService.Create").
			Str("stripeProductID", stripeProduct.ID).
			Msg("Attempting to archive Stripe product after database failure")

		if archiveErr := s.stripeService.ArchiveProduct(ctx, stripeProduct.ID); archiveErr != nil {
			// Log this error, but continue with the main error
			s.logger.Error().
				Str("function", "productService.Create").
				Err(archiveErr).
				Str("stripeProductID", stripeProduct.ID).
				Msg("Failed to archive Stripe product after database error")
		} else {
			s.logger.Debug().
				Str("function", "productService.Create").
				Str("stripeProductID", stripeProduct.ID).
				Msg("Successfully archived Stripe product after database failure")
		}

		return nil, fmt.Errorf("failed to create product in database: %w", err)
	}

	// 5. Create variants for the product based on options
	if product.Options != nil && len(product.Options) > 0 {
		s.logger.Debug().
			Str("function", "productService.Create").
			Str("productID", product.ID.String()).
			Interface("options", product.Options).
			Msg("Generating variants for product")

		err = s.variantService.GenerateVariantsForProduct(ctx, product)
		if err != nil {
			s.logger.Error().
				Str("function", "productService.Create").
				Err(err).
				Str("productID", product.ID.String()).
				Msg("Failed to generate variants for product")
				
			// We won't fail the product creation if variant generation fails,
			// but we'll log the error
			s.logger.Warn().
				Str("function", "productService.Create").
				Str("productID", product.ID.String()).
				Msg("Product was created but variants generation failed")
		}
	}

	s.logger.Info().
		Str("function", "productService.Create").
		Str("productID", product.ID.String()).
		Str("stripeID", product.StripeID).
		Str("name", product.Name).
		Int("stockLevel", product.StockLevel).
		Msg("Product successfully created")

	return product, nil
}

// GetByID retrieves a product by its ID
func (s *productService) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
    s.logger.Debug().
        Str("function", "productService.GetByID").
        Str("product_id", id.String()).
        Msg("Starting product retrieval by ID")
    
    // Validate ID
    if id == uuid.Nil {
        s.logger.Error().
            Str("function", "productService.GetByID").
            Msg("Nil UUID provided")
        return nil, fmt.Errorf("invalid product ID: nil UUID")
    }
    
    // 1. Call repository to fetch product
    s.logger.Debug().
        Str("function", "productService.GetByID").
        Str("product_id", id.String()).
        Msg("Calling repository to fetch product")
        
    product, err := s.productRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "productService.GetByID").
            Err(err).
            Str("product_id", id.String()).
            Msg("Failed to retrieve product from repository")
        return nil, fmt.Errorf("failed to retrieve product: %w", err)
    }
    
    // Check if product was found
    if product == nil {
        s.logger.Error().
            Str("function", "productService.GetByID").
            Str("product_id", id.String()).
            Msg("Product not found")
        return nil, fmt.Errorf("product with ID %s not found", id)
    }
    
    // Additional business logic can be added here
    // For example, check if the product is active for non-admin users
    // This would be an appropriate place to add that logic
    
    // Check stock level and log warning if low
    if product.StockLevel < 10 {
        s.logger.Warn().
            Str("function", "productService.GetByID").
            Str("product_id", id.String()).
            Str("product_name", product.Name).
            Int("stock_level", product.StockLevel).
            Msg("Product has low stock level")
    }
    
    s.logger.Info().
        Str("function", "productService.GetByID").
        Str("product_id", id.String()).
        Str("product_name", product.Name).
        Str("stripe_id", product.StripeID).
        Bool("active", product.Active).
        Int("stock_level", product.StockLevel).
        Msg("Product successfully retrieved")
        
    return product, nil
}

// List retrieves all products with optional filtering
func (s *productService) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Product, int, error) {
	s.logger.Debug().
		Str("function", "productService.List").
		Int("offset", offset).
		Int("limit", limit).
		Bool("includeInactive", includeInactive).
		Msg("Starting product listing")

	// Call repository to list products with the provided parameters
	s.logger.Debug().
		Str("function", "productService.List").
		Msg("Calling repository to fetch products")

	products, total, err := s.productRepo.List(ctx, offset, limit, includeInactive)
	if err != nil {
		s.logger.Error().
			Str("function", "productService.List").
			Err(err).
			Int("offset", offset).
			Int("limit", limit).
			Bool("includeInactive", includeInactive).
			Msg("Failed to retrieve products from repository")
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	// Log the result count
	s.logger.Debug().
		Str("function", "productService.List").
		Int("products_count", len(products)).
		Int("total_count", total).
		Msg("Successfully retrieved products from repository")

	// Additional processing if needed (e.g., filtering, sorting)
	// For example, you might want to enhance the products with additional data
	// or apply business rules that shouldn't be in the repository layer

	// For example, checking stock levels and logging warning for low stock
	for _, product := range products {
		if product.StockLevel < 10 {
			s.logger.Warn().
				Str("function", "productService.List").
				Str("product_id", product.ID.String()).
				Str("product_name", product.Name).
				Int("stock_level", product.StockLevel).
				Msg("Product has low stock level")
		}
	}

	// Log sample of products being returned (limited to first 5)
	if len(products) > 0 {
		logCount := len(products)
		if logCount > 5 {
			logCount = 5
		}

		for i := 0; i < logCount; i++ {
			s.logger.Debug().
				Str("function", "productService.List").
				Str("product_id", products[i].ID.String()).
				Str("product_name", products[i].Name).
				Str("stripe_id", products[i].StripeID).
				Bool("active", products[i].Active).
				Int("stock_level", products[i].StockLevel).
				Msgf("Product %d/%d in results", i+1, logCount)
		}
	}

	s.logger.Info().
		Str("function", "productService.List").
		Int("total_products", total).
		Int("returned_products", len(products)).
		Int("offset", offset).
		Int("limit", limit).
		Bool("includeInactive", includeInactive).
		Msg("Product listing completed successfully")

	return products, total, nil
}

// Update updates an existing product
// Update updates an existing product
func (s *productService) Update(ctx context.Context, id uuid.UUID, productDTO *dto.ProductUpdateDTO) (*models.Product, error) {
	// Get request ID from context if available
	var requestID string
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	// 1. Get existing product
	existingProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("service", "ProductService.Update").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Error retrieving product for update")
		return nil, fmt.Errorf("error retrieving product: %w", err)
	}

	if existingProduct == nil {
		s.logger.Error().
			Str("service", "ProductService.Update").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Product not found")
		return nil, fmt.Errorf("product not found with id: %s", id)
	}

	// Store the old options for variant management
	oldOptions := make(map[string][]string)
	for k, v := range existingProduct.Options {
		oldValues := make([]string, len(v))
		copy(oldValues, v)
		oldOptions[k] = oldValues
	}

	// 2. Update fields from DTO
	// Only update fields that are provided in the DTO
	if productDTO.Name != nil {
		existingProduct.Name = *productDTO.Name
	}
	if productDTO.Description != nil {
		existingProduct.Description = *productDTO.Description
	}
	if productDTO.ImageURL != nil {
		existingProduct.ImageURL = *productDTO.ImageURL
	}
	if productDTO.Active != nil {
		existingProduct.Active = *productDTO.Active
	}
	if productDTO.StockLevel != nil {
		existingProduct.StockLevel = *productDTO.StockLevel
	}
	if productDTO.Weight != nil {
		existingProduct.Weight = *productDTO.Weight
	}
	if productDTO.Origin != nil {
		existingProduct.Origin = *productDTO.Origin
	}
	if productDTO.RoastLevel != nil {
		existingProduct.RoastLevel = *productDTO.RoastLevel
	}
	if productDTO.FlavorNotes != nil {
		existingProduct.FlavorNotes = *productDTO.FlavorNotes
	}
	if productDTO.Options != nil {
		existingProduct.Options = *productDTO.Options
	}
	if productDTO.AllowSubscription != nil {
		existingProduct.AllowSubscription = *productDTO.AllowSubscription
	}

	existingProduct.UpdatedAt = time.Now()

	// 3. Update in Stripe if product has a Stripe ID
	if existingProduct.StripeID != "" {
		err = s.stripeService.UpdateProduct(ctx, existingProduct)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("service", "ProductService.Update").
				Str("request_id", requestID).
				Str("product_id", id.String()).
				Str("stripe_id", existingProduct.StripeID).
				Msg("Error updating product in Stripe")
			return nil, fmt.Errorf("error updating product in Stripe: %w", err)
		}
		s.logger.Debug().
			Str("service", "ProductService.Update").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Str("stripe_id", existingProduct.StripeID).
			Msg("Successfully updated product in Stripe")
	} else {
		s.logger.Warn().
			Str("service", "ProductService.Update").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Product has no Stripe ID, skipping Stripe update")
	}

	// 4. Update in database
	err = s.productRepo.Update(ctx, existingProduct)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("service", "ProductService.Update").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Error updating product in database")
		return nil, fmt.Errorf("error updating product in database: %w", err)
	}

	// 5. Check if variants need to be updated
	if productDTO.Options != nil {
		optionsChanged := false
		
		// Check if options have changed
		if len(oldOptions) != len(existingProduct.Options) {
			optionsChanged = true
		} else {
			// Check each option
			for key, oldValues := range oldOptions {
				newValues, exists := existingProduct.Options[key]
				if !exists {
					optionsChanged = true
					break
				}
				
				// Check if values have changed
				if len(oldValues) != len(newValues) {
					optionsChanged = true
					break
				}
				
				// Check each value
				for i, oldValue := range oldValues {
					if i >= len(newValues) || oldValue != newValues[i] {
						optionsChanged = true
						break
					}
				}
				
				if optionsChanged {
					break
				}
			}
		}
		
		if optionsChanged {
			s.logger.Debug().
				Str("service", "ProductService.Update").
				Str("request_id", requestID).
				Str("product_id", id.String()).
				Interface("old_options", oldOptions).
				Interface("new_options", existingProduct.Options).
				Msg("Product options have changed, updating variants")
				
			err = s.variantService.UpdateVariantsForProduct(ctx, existingProduct, oldOptions)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("service", "ProductService.Update").
					Str("request_id", requestID).
					Str("product_id", id.String()).
					Msg("Error updating variants for product")
					
				// We won't fail the product update if variant update fails,
				// but we'll log the error
				s.logger.Warn().
					Str("service", "ProductService.Update").
					Str("request_id", requestID).
					Str("product_id", id.String()).
					Msg("Product was updated but variants update failed")
			}
		}
	}

	s.logger.Info().
		Str("service", "ProductService.Update").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Msg("Successfully updated product")

	return existingProduct, nil
}

// Delete removes a product from the system
func (s *productService) Delete(ctx context.Context, id uuid.UUID) error {
    s.logger.Debug().
        Str("function", "productService.Delete").
        Str("product_id", id.String()).
        Msg("Starting product deletion")

    // 1. Get existing product from repository
    s.logger.Debug().
        Str("function", "productService.Delete").
        Str("product_id", id.String()).
        Msg("Retrieving product from repository")
        
    product, err := s.productRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "productService.Delete").
            Err(err).
            Str("product_id", id.String()).
            Msg("Failed to retrieve product")
        return fmt.Errorf("failed to retrieve product for deletion: %w", err)
    }
    
    if product == nil {
        s.logger.Error().
            Str("function", "productService.Delete").
            Str("product_id", id.String()).
            Msg("Product not found")
        return fmt.Errorf("product with ID %s not found", id)
    }
    
    s.logger.Debug().
        Str("function", "productService.Delete").
        Str("product_id", id.String()).
        Str("product_name", product.Name).
        Str("stripe_id", product.StripeID).
        Msg("Product found, proceeding with deletion")

    // 2. Archive in Stripe first if the product has a Stripe ID
    if product.StripeID != "" {
        s.logger.Debug().
            Str("function", "productService.Delete").
            Str("product_id", id.String()).
            Str("stripe_id", product.StripeID).
            Msg("Archiving product in Stripe")
            
        err = s.stripeService.ArchiveProduct(ctx, product.StripeID)
        if err != nil {
            s.logger.Error().
                Str("function", "productService.Delete").
                Err(err).
                Str("product_id", id.String()).
                Str("stripe_id", product.StripeID).
                Msg("Failed to archive product in Stripe")
            return fmt.Errorf("failed to archive product in Stripe: %w", err)
        }
        
        s.logger.Debug().
            Str("function", "productService.Delete").
            Str("product_id", id.String()).
            Str("stripe_id", product.StripeID).
            Msg("Successfully archived product in Stripe")
    } else {
        s.logger.Warn().
            Str("function", "productService.Delete").
            Str("product_id", id.String()).
            Msg("Product has no Stripe ID, skipping Stripe archiving")
    }

    // 3. Delete from database
    s.logger.Debug().
        Str("function", "productService.Delete").
        Str("product_id", id.String()).
        Msg("Deleting product from database")
        
    err = s.productRepo.Delete(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "productService.Delete").
            Err(err).
            Str("product_id", id.String()).
            Msg("Failed to delete product from database")
            
        // If database deletion fails but we already archived in Stripe,
        // we should log this inconsistency
        if product.StripeID != "" {
            s.logger.Warn().
                Str("function", "productService.Delete").
                Str("product_id", id.String()).
                Str("stripe_id", product.StripeID).
                Msg("Inconsistent state: Product archived in Stripe but not deleted from database")
        }
        
        return fmt.Errorf("failed to delete product from database: %w", err)
    }
    
    s.logger.Info().
        Str("function", "productService.Delete").
        Str("product_id", id.String()).
        Str("product_name", product.Name).
        Str("stripe_id", product.StripeID).
        Msg("Product successfully deleted")
        
    return nil
}

// UpdateStockLevel updates the stock level of a product (simplified version)
func (s *productService) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	// Get request ID from context if available
	var requestID string
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	s.logger.Debug().
		Str("service", "ProductService.UpdateStockLevel").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Int("new_quantity", quantity).
		Msg("Starting stock level update")

	// 1. Get existing product
	existingProduct, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("service", "ProductService.UpdateStockLevel").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Error retrieving product for stock update")
		return fmt.Errorf("error retrieving product: %w", err)
	}

	if existingProduct == nil {
		s.logger.Error().
			Str("service", "ProductService.UpdateStockLevel").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Product not found")
		return fmt.Errorf("product not found with id: %s", id)
	}

	// Record the old stock level for logging
	oldStockLevel := existingProduct.StockLevel

	// 2. Update stock level
	existingProduct.StockLevel = quantity
	existingProduct.UpdatedAt = time.Now()

	// 3. Update in database
	err = s.productRepo.Update(ctx, existingProduct)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("service", "ProductService.UpdateStockLevel").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Msg("Error updating product stock in database")
		return fmt.Errorf("error updating product stock in database: %w", err)
	}

	s.logger.Info().
		Str("service", "ProductService.UpdateStockLevel").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Str("product_name", existingProduct.Name).
		Int("old_stock", oldStockLevel).
		Int("new_stock", quantity).
		Msg("Successfully updated product stock level")

	return nil
}