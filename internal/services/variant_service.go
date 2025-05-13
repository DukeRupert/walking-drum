// internal/services/variant_service.go
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

// VariantService defines the interface for variant business logic
type VariantService interface {
	// Generate variants for a product based on its options
	GenerateVariantsForProduct(ctx context.Context, product *models.Product) error
	
	// Update variants for a product after options change
	UpdateVariantsForProduct(ctx context.Context, product *models.Product, oldOptions map[string][]string) error
	
	// Get variants for a product
	GetVariantsByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Variant, error)
	
	// Basic CRUD operations
	GetByID(ctx context.Context, id uuid.UUID) (*models.Variant, error)
	List(ctx context.Context, page, pageSize int, activeOnly bool) ([]*models.Variant, int, error)
	Update(ctx context.Context, id uuid.UUID, variantDTO *dto.VariantUpdateDTO) (*models.Variant, error)
	UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error
}

type variantService struct {
	variantRepo  interfaces.VariantRepository
	productRepo  interfaces.ProductRepository
	priceRepo    interfaces.PriceRepository
	stripeClient stripe.StripeService
	logger       zerolog.Logger
}

// NewVariantService creates a new variant service
func NewVariantService(
	variantRepo interfaces.VariantRepository,
	productRepo interfaces.ProductRepository,
	priceRepo interfaces.PriceRepository,
	stripeClient stripe.StripeService,
	logger *zerolog.Logger,
) VariantService {
	return &variantService{
		variantRepo:  variantRepo,
		productRepo:  productRepo,
		priceRepo:    priceRepo,
		stripeClient: stripeClient,
		logger:       logger.With().Str("component", "variant_service").Logger(),
	}
}

// Create adds a new variant to the system
func (s *variantService) Create(ctx context.Context, variantDTO *dto.VariantCreateDTO) (*models.Variant, error) {
	s.logger.Debug().
		Str("function", "variantService.Create").
		Interface("variantDTO", variantDTO).
		Msg("Starting variant creation")

	// 1. Validate variantDTO
	if problems := variantDTO.Valid(ctx); len(problems) > 0 {
		s.logger.Error().
			Str("function", "variantService.Create").
			Interface("problems", problems).
			Msg("Variant validation failed")
		return nil, fmt.Errorf("invalid variant data: %v", problems)
	}

	s.logger.Debug().
		Str("function", "variantService.Create").
		Msg("Variant validation passed")

	// 2. Verify that the product exists
	s.logger.Debug().
		Str("function", "variantService.Create").
		Str("product_id", variantDTO.ProductID.String()).
		Msg("Verifying product exists")

	product, err := s.productRepo.GetByID(ctx, variantDTO.ProductID)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Err(err).
			Str("product_id", variantDTO.ProductID.String()).
			Msg("Failed to retrieve associated product")
		return nil, fmt.Errorf("failed to verify product exists: %w", err)
	}

	if product == nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Str("product_id", variantDTO.ProductID.String()).
			Msg("Product not found")
		return nil, fmt.Errorf("product with ID %s not found", variantDTO.ProductID)
	}

	// 3. Verify that the price exists
	s.logger.Debug().
		Str("function", "variantService.Create").
		Str("price_id", variantDTO.PriceID.String()).
		Msg("Verifying price exists")

	price, err := s.priceRepo.GetByID(ctx, variantDTO.PriceID)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Err(err).
			Str("price_id", variantDTO.PriceID.String()).
			Msg("Failed to retrieve associated price")
		return nil, fmt.Errorf("failed to verify price exists: %w", err)
	}

	if price == nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Str("price_id", variantDTO.PriceID.String()).
			Msg("Price not found")
		return nil, fmt.Errorf("price with ID %s not found", variantDTO.PriceID)
	}

	// 4. Verify that the price is for the product
	if price.ProductID != variantDTO.ProductID {
		s.logger.Error().
			Str("function", "variantService.Create").
			Str("price_id", variantDTO.PriceID.String()).
			Str("price_product_id", price.ProductID.String()).
			Str("variant_product_id", variantDTO.ProductID.String()).
			Msg("Price does not belong to the product")
		return nil, fmt.Errorf("price with ID %s does not belong to product with ID %s",
			variantDTO.PriceID, variantDTO.ProductID)
	}

	// 5. Check if variant with these attributes already exists
	existingVariant, err := s.variantRepo.GetByAttributes(ctx,
		variantDTO.ProductID, variantDTO.Weight, variantDTO.Grind)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Err(err).
			Str("product_id", variantDTO.ProductID.String()).
			Str("weight", variantDTO.Weight).
			Str("grind", variantDTO.Grind).
			Msg("Error checking for existing variant")
		return nil, fmt.Errorf("error checking for existing variant: %w", err)
	}

	if existingVariant != nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Str("product_id", variantDTO.ProductID.String()).
			Str("weight", variantDTO.Weight).
			Str("grind", variantDTO.Grind).
			Str("existing_variant_id", existingVariant.ID.String()).
			Msg("Variant with these attributes already exists")
		return nil, fmt.Errorf("variant with these attributes already exists")
	}

	// 6. Create variant in database
	now := time.Now()
	variant := &models.Variant{
		ID:            uuid.New(),
		ProductID:     variantDTO.ProductID,
		PriceID:       variantDTO.PriceID,
		StripePriceID: variantDTO.StripePriceID,
		Weight:        variantDTO.Weight,
		Grind:         variantDTO.Grind,
		Active:        variantDTO.Active,
		StockLevel:    variantDTO.StockLevel,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.logger.Debug().
		Str("function", "variantService.Create").
		Str("variant_id", variant.ID.String()).
		Str("product_id", variant.ProductID.String()).
		Str("price_id", variant.PriceID.String()).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Msg("Preparing to save variant to database")

	if err := s.variantRepo.Create(ctx, variant); err != nil {
		s.logger.Error().
			Str("function", "variantService.Create").
			Err(err).
			Str("variant_id", variant.ID.String()).
			Msg("Failed to create variant in database")
		return nil, fmt.Errorf("failed to create variant in database: %w", err)
	}

	s.logger.Info().
		Str("function", "variantService.Create").
		Str("variant_id", variant.ID.String()).
		Str("product_id", variant.ProductID.String()).
		Str("price_id", variant.PriceID.String()).
		Str("stripe_price_id", variant.StripePriceID).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Int("stock_level", variant.StockLevel).
		Msg("Variant successfully created")

	return variant, nil
}

// GenerateVariantsForProduct creates all possible variants for a product based on its options
func (s *variantService) GenerateVariantsForProduct(ctx context.Context, product *models.Product) error {
	s.logger.Debug().
		Str("function", "variantService.GenerateVariantsForProduct").
		Str("product_id", product.ID.String()).
		Interface("options", product.Options).
		Msg("Starting variant generation for product")

	// Check if product has options
	if product.Options == nil || len(product.Options) == 0 {
		s.logger.Warn().
			Str("function", "variantService.GenerateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Msg("Product has no options, skipping variant generation")
		return nil
	}

	// Extract weight and grind options
	weightOptions, hasWeight := product.Options["weight"]
	grindOptions, hasGrind := product.Options["grind"]

	// Validate options
	if !hasWeight || len(weightOptions) == 0 {
		s.logger.Error().
			Str("function", "variantService.GenerateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Msg("Product has no weight options")
		return fmt.Errorf("product must have weight options")
	}

	if !hasGrind || len(grindOptions) == 0 {
		s.logger.Error().
			Str("function", "variantService.GenerateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Msg("Product has no grind options")
		return fmt.Errorf("product must have grind options")
	}

	// Create a variant for each combination of weight and grind
	for _, weight := range weightOptions {
		for _, grind := range grindOptions {
			// Check if variant already exists
			existingVariant, err := s.variantRepo.GetByAttributes(ctx, product.ID, weight, grind)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("weight", weight).
					Str("grind", grind).
					Msg("Error checking for existing variant")
				return fmt.Errorf("error checking for existing variant: %w", err)
			}

			if existingVariant != nil {
				s.logger.Debug().
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("variant_id", existingVariant.ID.String()).
					Str("weight", weight).
					Str("grind", grind).
					Msg("Variant already exists, skipping")
				continue
			}

			// Create a Stripe product for the variant
			variantName := fmt.Sprintf("%s - %s, %s", product.Name, weight, grind)
			variantDescription := fmt.Sprintf("%s. Weight: %s, Grind: %s", product.Description, weight, grind)

			stripeProduct, err := s.stripeClient.CreateProduct(ctx, &stripe.ProductCreateParams{
				Name:        variantName,
				Description: variantDescription,
				Images:      []string{product.ImageURL},
				Active:      product.Active,
				Metadata: map[string]string{
					"product_id":  product.ID.String(),
					"weight":      weight,
					"grind":       grind,
					"origin":      product.Origin,
					"roast_level": product.RoastLevel,
				},
			})
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("weight", weight).
					Str("grind", grind).
					Msg("Failed to create Stripe product for variant")
				return fmt.Errorf("failed to create Stripe product for variant: %w", err)
			}

			s.logger.Debug().
				Str("function", "variantService.GenerateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("stripe_product_id", stripeProduct.ID).
				Str("weight", weight).
				Str("grind", grind).
				Msg("Created Stripe product for variant")

			// Create a default price for the variant if it's a one-time purchase
			stripePrice, err := s.stripeClient.CreatePrice(ctx, &stripe.PriceCreateParams{
				ProductID:  stripeProduct.ID,
				Currency:   "usd",
				UnitAmount: 1500, // Default price, $15.00
				Nickname:   "Default one-time price",
				Active:     true,
				Metadata: map[string]string{
					"product_id": product.ID.String(),
					"weight":     weight,
					"grind":      grind,
				},
			})
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("stripe_product_id", stripeProduct.ID).
					Msg("Failed to create Stripe price for variant")

				// Try to clean up the Stripe product since price creation failed
				if archiveErr := s.stripeClient.ArchiveProduct(ctx, stripeProduct.ID); archiveErr != nil {
					s.logger.Error().
						Err(archiveErr).
						Str("stripe_product_id", stripeProduct.ID).
						Msg("Failed to archive Stripe product after price creation error")
				}

				return fmt.Errorf("failed to create Stripe price for variant: %w", err)
			}

			// Save the price to our database
			now := time.Now()
			price := &models.Price{
				ID:            uuid.New(),
				ProductID:     product.ID,
				Name:          "Default one-time price",
				Amount:        1500,
				Currency:      "usd",
				Type:          "one_time",
				Active:        true,
				StripeID:      stripePrice.ID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.priceRepo.Create(ctx, price); err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("stripe_price_id", stripePrice.ID).
					Msg("Failed to save price to database")
				return fmt.Errorf("failed to save price to database: %w", err)
			}

			// Create subscription price if the product allows subscriptions
			if product.AllowSubscription {
				stripeSubPrice, err := s.stripeClient.CreatePrice(ctx, &stripe.PriceCreateParams{
					ProductID:  stripeProduct.ID,
					Currency:   "usd",
					UnitAmount: 1200, // Default subscription price, $12.00 (slightly cheaper)
					Nickname:   "Monthly subscription",
					Recurring: &stripe.RecurringParams{
						Interval:      "month",
						IntervalCount: 1,
					},
					Active: true,
					Metadata: map[string]string{
						"product_id": product.ID.String(),
						"weight":     weight,
						"grind":      grind,
						"type":       "subscription",
					},
				})
				if err != nil {
					s.logger.Error().
						Err(err).
						Str("function", "variantService.GenerateVariantsForProduct").
						Str("product_id", product.ID.String()).
						Str("stripe_product_id", stripeProduct.ID).
						Msg("Failed to create Stripe subscription price for variant")
					return fmt.Errorf("failed to create Stripe subscription price for variant: %w", err)
				}

				// Save the subscription price to our database
				subscriptionPrice := &models.Price{
					ID:            uuid.New(),
					ProductID:     product.ID,
					Name:          "Monthly subscription",
					Amount:        1200,
					Currency:      "usd",
					Type:          "recurring",
					Interval:      "month",
					IntervalCount: 1,
					Active:        true,
					StripeID:      stripeSubPrice.ID,
					CreatedAt:     now,
					UpdatedAt:     now,
				}

				if err := s.priceRepo.Create(ctx, subscriptionPrice); err != nil {
					s.logger.Error().
						Err(err).
						Str("function", "variantService.GenerateVariantsForProduct").
						Str("product_id", product.ID.String()).
						Str("stripe_price_id", stripeSubPrice.ID).
						Msg("Failed to save subscription price to database")
					return fmt.Errorf("failed to save subscription price to database: %w", err)
				}
			}

			// Create the variant in our database
			variant := &models.Variant{
				ID:            uuid.New(),
				ProductID:     product.ID,
				PriceID:       price.ID,
				StripePriceID: stripePrice.ID,
				Weight:        weight,
				Grind:         grind,
				Active:        product.Active,
				StockLevel:    product.StockLevel,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.variantRepo.Create(ctx, variant); err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.GenerateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("variant_id", variant.ID.String()).
					Msg("Failed to save variant to database")
				return fmt.Errorf("failed to save variant to database: %w", err)
			}

			s.logger.Info().
				Str("function", "variantService.GenerateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("variant_id", variant.ID.String()).
				Str("weight", weight).
				Str("grind", grind).
				Str("stripe_product_id", stripeProduct.ID).
				Msg("Successfully created variant")
		}
	}

	s.logger.Info().
		Str("function", "variantService.GenerateVariantsForProduct").
		Str("product_id", product.ID.String()).
		Msg("Successfully generated all variants for product")

	return nil
}

// GetByID retrieves a variant by its ID
func (s *variantService) GetByID(ctx context.Context, id uuid.UUID) (*models.Variant, error) {
	s.logger.Debug().
		Str("function", "variantService.GetByID").
		Str("variant_id", id.String()).
		Msg("Starting variant retrieval by ID")

	// Validate ID
	if id == uuid.Nil {
		s.logger.Error().
			Str("function", "variantService.GetByID").
			Msg("Nil UUID provided")
		return nil, fmt.Errorf("invalid variant ID: nil UUID")
	}

	// Call repository to fetch variant
	s.logger.Debug().
		Str("function", "variantService.GetByID").
		Str("variant_id", id.String()).
		Msg("Calling repository to fetch variant")

	variant, err := s.variantRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.GetByID").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to retrieve variant from repository")
		return nil, fmt.Errorf("failed to retrieve variant: %w", err)
	}

	// Check if variant was found
	if variant == nil {
		s.logger.Error().
			Str("function", "variantService.GetByID").
			Str("variant_id", id.String()).
			Msg("Variant not found")
		return nil, fmt.Errorf("variant with ID %s not found", id)
	}

	// Check stock level and log warning if low
	if variant.StockLevel < 10 {
		s.logger.Warn().
			Str("function", "variantService.GetByID").
			Str("variant_id", id.String()).
			Str("product_id", variant.ProductID.String()).
			Int("stock_level", variant.StockLevel).
			Msg("Variant has low stock level")
	}

	s.logger.Info().
		Str("function", "variantService.GetByID").
		Str("variant_id", id.String()).
		Str("product_id", variant.ProductID.String()).
		Str("price_id", variant.PriceID.String()).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Int("stock_level", variant.StockLevel).
		Bool("active", variant.Active).
		Msg("Variant successfully retrieved")

	return variant, nil
}

// GetByProductID retrieves all variants for a product
func (s *variantService) GetVariantsByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Variant, error) {
	s.logger.Debug().
		Str("function", "variantService.GetByProductID").
		Str("product_id", productID.String()).
		Msg("Starting variant retrieval by product ID")

	// Validate product ID
	if productID == uuid.Nil {
		s.logger.Error().
			Str("function", "variantService.GetByProductID").
			Msg("Nil UUID provided")
		return nil, fmt.Errorf("invalid product ID: nil UUID")
	}

	// Verify that the product exists
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.GetByProductID").
			Err(err).
			Str("product_id", productID.String()).
			Msg("Failed to retrieve product")
		return nil, fmt.Errorf("failed to verify product exists: %w", err)
	}

	if product == nil {
		s.logger.Error().
			Str("function", "variantService.GetByProductID").
			Str("product_id", productID.String()).
			Msg("Product not found")
		return nil, fmt.Errorf("product with ID %s not found", productID)
	}

	// Call repository to fetch variants
	s.logger.Debug().
		Str("function", "variantService.GetByProductID").
		Str("product_id", productID.String()).
		Msg("Calling repository to fetch variants")

	variants, err := s.variantRepo.GetByProductID(ctx, productID)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.GetByProductID").
			Err(err).
			Str("product_id", productID.String()).
			Msg("Failed to retrieve variants from repository")
		return nil, fmt.Errorf("failed to retrieve variants: %w", err)
	}

	s.logger.Info().
		Str("function", "variantService.GetByProductID").
		Str("product_id", productID.String()).
		Int("variants_count", len(variants)).
		Msg("Variants successfully retrieved")

	return variants, nil
}

// GetByAttributes retrieves a variant by product ID, weight, and grind
func (s *variantService) GetByAttributes(ctx context.Context, productID uuid.UUID, weight, grind string) (*models.Variant, error) {
	s.logger.Debug().
		Str("function", "variantService.GetByAttributes").
		Str("product_id", productID.String()).
		Str("weight", weight).
		Str("grind", grind).
		Msg("Starting variant retrieval by attributes")

	// Validate inputs
	if productID == uuid.Nil {
		s.logger.Error().
			Str("function", "variantService.GetByAttributes").
			Msg("Nil UUID provided")
		return nil, fmt.Errorf("invalid product ID: nil UUID")
	}

	if weight == "" {
		s.logger.Error().
			Str("function", "variantService.GetByAttributes").
			Msg("Empty weight provided")
		return nil, fmt.Errorf("weight cannot be empty")
	}

	if grind == "" {
		s.logger.Error().
			Str("function", "variantService.GetByAttributes").
			Msg("Empty grind provided")
		return nil, fmt.Errorf("grind cannot be empty")
	}

	// Call repository to fetch variant
	s.logger.Debug().
		Str("function", "variantService.GetByAttributes").
		Str("product_id", productID.String()).
		Str("weight", weight).
		Str("grind", grind).
		Msg("Calling repository to fetch variant")

	variant, err := s.variantRepo.GetByAttributes(ctx, productID, weight, grind)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.GetByAttributes").
			Err(err).
			Str("product_id", productID.String()).
			Str("weight", weight).
			Str("grind", grind).
			Msg("Failed to retrieve variant from repository")
		return nil, fmt.Errorf("failed to retrieve variant: %w", err)
	}

	// Check if variant was found
	if variant == nil {
		s.logger.Debug().
			Str("function", "variantService.GetByAttributes").
			Str("product_id", productID.String()).
			Str("weight", weight).
			Str("grind", grind).
			Msg("Variant not found")
		return nil, nil
	}

	s.logger.Info().
		Str("function", "variantService.GetByAttributes").
		Str("variant_id", variant.ID.String()).
		Str("product_id", variant.ProductID.String()).
		Str("price_id", variant.PriceID.String()).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Msg("Variant successfully retrieved")

	return variant, nil
}

// List retrieves all variants with pagination and enriched details
func (s *variantService) List(ctx context.Context, offset, limit int, activeOnly bool) ([]*models.Variant, int, error) {
	s.logger.Debug().
		Str("function", "variantService.List").
		Int("offset", offset).
		Int("limit", limit).
		Bool("activeOnly", activeOnly).
		Msg("Starting variant listing")

	// Call repository to list variants
	s.logger.Debug().
		Str("function", "variantService.List").
		Msg("Calling repository to fetch variants")

	variants, total, err := s.variantRepo.List(ctx, limit, offset, activeOnly)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.List").
			Err(err).
			Int("offset", offset).
			Int("limit", limit).
			Bool("activeOnly", activeOnly).
			Msg("Failed to retrieve variants from repository")
		return nil, 0, fmt.Errorf("failed to list variants: %w", err)
	}

	// Log the result count
	s.logger.Debug().
		Str("function", "variantService.List").
		Int("variants_count", len(variants)).
		Int("total_count", total).
		Msg("Successfully retrieved variants from repository")

	// For logging/debugging purposes, we can still fetch the details
	// but will return just the variants as required by the interface
	for _, variant := range variants {
		// Get product details for logging
		product, err := s.productRepo.GetByID(ctx, variant.ProductID)
		if err != nil {
			s.logger.Warn().
				Str("function", "variantService.List").
				Err(err).
				Str("variant_id", variant.ID.String()).
				Str("product_id", variant.ProductID.String()).
				Msg("Failed to retrieve product details for variant")
			continue
		}

		// Get price details for logging
		price, err := s.priceRepo.GetByID(ctx, variant.PriceID)
		if err != nil {
			s.logger.Warn().
				Str("function", "variantService.List").
				Err(err).
				Str("variant_id", variant.ID.String()).
				Str("price_id", variant.PriceID.String()).
				Msg("Failed to retrieve price details for variant")
			continue
		}

		// Just log the details rather than building a new structure
		s.logger.Debug().
			Str("variant_id", variant.ID.String()).
			Str("product_name", product.Name).
			Str("price_name", price.Name).
			Int64("amount", price.Amount).
			Str("weight", variant.Weight).
			Str("grind", variant.Grind).
			Msg("Variant details")
	}

	s.logger.Info().
		Str("function", "variantService.List").
		Int("total_variants", total).
		Int("returned_variants", len(variants)).
		Int("offset", offset).
		Int("limit", limit).
		Bool("activeOnly", activeOnly).
		Msg("Variant listing completed successfully")

	return variants, total, nil
}

// Update updates an existing variant
func (s *variantService) Update(ctx context.Context, id uuid.UUID, variantDTO *dto.VariantUpdateDTO) (*models.Variant, error) {
	s.logger.Debug().
		Str("function", "variantService.Update").
		Str("variant_id", id.String()).
		Interface("variantDTO", variantDTO).
		Msg("Starting variant update")

	// 1. Validate input parameters
	if id == uuid.Nil {
		s.logger.Error().
			Str("function", "variantService.Update").
			Msg("Nil UUID provided for variant ID")
		return nil, fmt.Errorf("invalid variant ID: nil UUID")
	}

	if problems := variantDTO.Valid(ctx); len(problems) > 0 {
		s.logger.Error().
			Str("function", "variantService.Update").
			Interface("problems", problems).
			Msg("Variant update validation failed")
		return nil, fmt.Errorf("invalid variant data: %v", problems)
	}

	// 2. Get existing variant
	s.logger.Debug().
		Str("function", "variantService.Update").
		Str("variant_id", id.String()).
		Msg("Retrieving existing variant")

	existingVariant, err := s.variantRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Update").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to retrieve existing variant")
		return nil, fmt.Errorf("failed to retrieve existing variant: %w", err)
	}

	if existingVariant == nil {
		s.logger.Error().
			Str("function", "variantService.Update").
			Str("variant_id", id.String()).
			Msg("Variant not found")
		return nil, fmt.Errorf("variant with ID %s not found", id)
	}

	// 3. Check if related entities need to be verified
	if variantDTO.ProductID != nil && *variantDTO.ProductID != existingVariant.ProductID {
		// Verify new product exists
		product, err := s.productRepo.GetByID(ctx, *variantDTO.ProductID)
		if err != nil {
			s.logger.Error().
				Str("function", "variantService.Update").
				Err(err).
				Str("product_id", variantDTO.ProductID.String()).
				Msg("Failed to retrieve product")
			return nil, fmt.Errorf("failed to verify product exists: %w", err)
		}

		if product == nil {
			s.logger.Error().
				Str("function", "variantService.Update").
				Str("product_id", variantDTO.ProductID.String()).
				Msg("Product not found")
			return nil, fmt.Errorf("product with ID %s not found", *variantDTO.ProductID)
		}

		// Update product ID
		existingVariant.ProductID = *variantDTO.ProductID
	}

	if variantDTO.PriceID != nil && *variantDTO.PriceID != existingVariant.PriceID {
		// Verify new price exists
		price, err := s.priceRepo.GetByID(ctx, *variantDTO.PriceID)
		if err != nil {
			s.logger.Error().
				Str("function", "variantService.Update").
				Err(err).
				Str("price_id", variantDTO.PriceID.String()).
				Msg("Failed to retrieve price")
			return nil, fmt.Errorf("failed to verify price exists: %w", err)
		}

		if price == nil {
			s.logger.Error().
				Str("function", "variantService.Update").
				Str("price_id", variantDTO.PriceID.String()).
				Msg("Price not found")
			return nil, fmt.Errorf("price with ID %s not found", *variantDTO.PriceID)
		}

		// Verify price belongs to product
		if price.ProductID != existingVariant.ProductID {
			s.logger.Error().
				Str("function", "variantService.Update").
				Str("price_id", variantDTO.PriceID.String()).
				Str("price_product_id", price.ProductID.String()).
				Str("variant_product_id", existingVariant.ProductID.String()).
				Msg("Price does not belong to the product")
			return nil, fmt.Errorf("price with ID %s does not belong to product with ID %s",
				*variantDTO.PriceID, existingVariant.ProductID)
		}

		// Update price ID
		existingVariant.PriceID = *variantDTO.PriceID
	}

	// 4. Update other fields from DTO
	if variantDTO.StripePriceID != nil {
		existingVariant.StripePriceID = *variantDTO.StripePriceID
	}

	if variantDTO.Weight != nil {
		existingVariant.Weight = *variantDTO.Weight
	}

	if variantDTO.Grind != nil {
		existingVariant.Grind = *variantDTO.Grind
	}

	if variantDTO.Active != nil {
		existingVariant.Active = *variantDTO.Active
	}

	if variantDTO.StockLevel != nil {
		existingVariant.StockLevel = *variantDTO.StockLevel
	}

	// 5. Check if a variant with these updated attributes already exists
	if (variantDTO.ProductID != nil || variantDTO.Weight != nil || variantDTO.Grind != nil) &&
		(variantDTO.ProductID != nil && *variantDTO.ProductID != existingVariant.ProductID ||
			variantDTO.Weight != nil && *variantDTO.Weight != existingVariant.Weight ||
			variantDTO.Grind != nil && *variantDTO.Grind != existingVariant.Grind) {

		productID := existingVariant.ProductID
		if variantDTO.ProductID != nil {
			productID = *variantDTO.ProductID
		}

		weight := existingVariant.Weight
		if variantDTO.Weight != nil {
			weight = *variantDTO.Weight
		}

		grind := existingVariant.Grind
		if variantDTO.Grind != nil {
			grind = *variantDTO.Grind
		}

		// Check if another variant with these attributes exists
		otherVariant, err := s.variantRepo.GetByAttributes(ctx, productID, weight, grind)
		if err != nil {
			s.logger.Error().
				Str("function", "variantService.Update").
				Err(err).
				Str("product_id", productID.String()).
				Str("weight", weight).
				Str("grind", grind).
				Msg("Error checking for existing variant")
			return nil, fmt.Errorf("error checking for existing variant: %w", err)
		}

		if otherVariant != nil && otherVariant.ID != id {
			s.logger.Error().
				Str("function", "variantService.Update").
				Str("product_id", productID.String()).
				Str("weight", weight).
				Str("grind", grind).
				Str("existing_variant_id", otherVariant.ID.String()).
				Msg("Another variant with these attributes already exists")
			return nil, fmt.Errorf("another variant with these attributes already exists")
		}
	}

	// 6. Update in database
	existingVariant.UpdatedAt = time.Now()

	s.logger.Debug().
		Str("function", "variantService.Update").
		Str("variant_id", id.String()).
		Msg("Updating variant in database")

	if err := s.variantRepo.Update(ctx, existingVariant); err != nil {
		s.logger.Error().
			Str("function", "variantService.Update").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to update variant in database")
		return nil, fmt.Errorf("failed to update variant in database: %w", err)
	}

	s.logger.Info().
		Str("function", "variantService.Update").
		Str("variant_id", existingVariant.ID.String()).
		Str("product_id", existingVariant.ProductID.String()).
		Str("price_id", existingVariant.PriceID.String()).
		Str("weight", existingVariant.Weight).
		Str("grind", existingVariant.Grind).
		Int("stock_level", existingVariant.StockLevel).
		Msg("Variant successfully updated")

	return existingVariant, nil
}

// UpdateVariantsForProduct updates variants after product options change
func (s *variantService) UpdateVariantsForProduct(ctx context.Context, product *models.Product, oldOptions map[string][]string) error {
	s.logger.Debug().
		Str("function", "variantService.UpdateVariantsForProduct").
		Str("product_id", product.ID.String()).
		Interface("new_options", product.Options).
		Interface("old_options", oldOptions).
		Msg("Starting variant update for product")

	// Get existing variants
	existingVariants, err := s.variantRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "variantService.UpdateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Msg("Failed to get existing variants")
		return fmt.Errorf("failed to get existing variants: %w", err)
	}

	// Build a map of existing variant combinations for quick lookup
	existingCombinations := make(map[string]*models.Variant)
	for _, v := range existingVariants {
		key := v.Weight + ":" + v.Grind
		existingCombinations[key] = v
	}

	// Determine which combinations are new and need to be created
	newVariants := make([]struct {
		weight string
		grind  string
	}, 0)

	// Extract current weight and grind options
	weightOptions, hasWeight := product.Options["weight"]
	grindOptions, hasGrind := product.Options["grind"]

	if hasWeight && hasGrind {
		for _, weight := range weightOptions {
			for _, grind := range grindOptions {
				key := weight + ":" + grind
				if _, exists := existingCombinations[key]; !exists {
					newVariants = append(newVariants, struct {
						weight string
						grind  string
					}{weight, grind})
				}
			}
		}
	}

	// Create new variants for the added options
	for _, newVariant := range newVariants {
		// Create a Stripe product for the variant
		variantName := fmt.Sprintf("%s - %s, %s", product.Name, newVariant.weight, newVariant.grind)
		variantDescription := fmt.Sprintf("%s. Weight: %s, Grind: %s", product.Description, newVariant.weight, newVariant.grind)

		stripeProduct, err := s.stripeClient.CreateProduct(ctx, &stripe.ProductCreateParams{
			Name:        variantName,
			Description: variantDescription,
			Images:      []string{product.ImageURL},
			Active:      product.Active,
			Metadata: map[string]string{
				"product_id":  product.ID.String(),
				"weight":      newVariant.weight,
				"grind":       newVariant.grind,
				"origin":      product.Origin,
				"roast_level": product.RoastLevel,
			},
		})
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("weight", newVariant.weight).
				Str("grind", newVariant.grind).
				Msg("Failed to create Stripe product for new variant")
			return fmt.Errorf("failed to create Stripe product for new variant: %w", err)
		}

		// Create a default price for the variant
		stripePrice, err := s.stripeClient.CreatePrice(ctx, &stripe.PriceCreateParams{
			ProductID:  stripeProduct.ID,
			Currency:   "usd",
			UnitAmount: 1500, // Default price
			Nickname:   "Default one-time price",
			Active:     true,
			Metadata: map[string]string{
				"product_id": product.ID.String(),
				"weight":     newVariant.weight,
				"grind":      newVariant.grind,
			},
		})
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("stripe_product_id", stripeProduct.ID).
				Msg("Failed to create Stripe price for new variant")
			return fmt.Errorf("failed to create Stripe price for new variant: %w", err)
		}

		// Save the price to our database
		now := time.Now()
		price := &models.Price{
			ID:            uuid.New(),
			ProductID:     product.ID,
			Name:          "Default one-time price",
			Amount:        1500,
			Currency:      "usd",
			Type:          "one_time",
			Active:        true,
			StripeID:      stripePrice.ID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.priceRepo.Create(ctx, price); err != nil {
			s.logger.Error().
				Err(err).
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("stripe_price_id", stripePrice.ID).
				Msg("Failed to save price to database")
			return fmt.Errorf("failed to save price to database: %w", err)
		}

		// Create subscription price if the product allows subscriptions
		if product.AllowSubscription {
			stripeSubPrice, err := s.stripeClient.CreatePrice(ctx, &stripe.PriceCreateParams{
				ProductID:  stripeProduct.ID,
				Currency:   "usd",
				UnitAmount: 1200, // Default subscription price (slightly cheaper)
				Nickname:   "Monthly subscription",
				Recurring: &stripe.RecurringParams{
					Interval:      "month",
					IntervalCount: 1,
				},
				Active: true,
				Metadata: map[string]string{
					"product_id": product.ID.String(),
					"weight":     newVariant.weight,
					"grind":      newVariant.grind,
					"type":       "subscription",
				},
			})
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.UpdateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("stripe_product_id", stripeProduct.ID).
					Msg("Failed to create Stripe subscription price for new variant")
				return fmt.Errorf("failed to create Stripe subscription price for new variant: %w", err)
			}

			// Save the subscription price to our database
			subscriptionPrice := &models.Price{
				ID:            uuid.New(),
				ProductID:     product.ID,
				Name:          "Monthly subscription",
				Amount:        1200,
				Currency:      "usd",
				Type:          "recurring",
				Interval:      "month",
				IntervalCount: 1,
				Active:        true,
				StripeID:      stripeSubPrice.ID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.priceRepo.Create(ctx, subscriptionPrice); err != nil {
				s.logger.Error().
					Err(err).
					Str("function", "variantService.UpdateVariantsForProduct").
					Str("product_id", product.ID.String()).
					Str("stripe_price_id", stripeSubPrice.ID).
					Msg("Failed to save subscription price to database")
				return fmt.Errorf("failed to save subscription price to database: %w", err)
			}
		}

		// Create the variant in our database
		variant := &models.Variant{
			ID:            uuid.New(),
			ProductID:     product.ID,
			PriceID:       price.ID,
			StripePriceID: stripePrice.ID,
			Weight:        newVariant.weight,
			Grind:         newVariant.grind,
			Active:        product.Active,
			StockLevel:    product.StockLevel,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.variantRepo.Create(ctx, variant); err != nil {
			s.logger.Error().
				Err(err).
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("variant_id", variant.ID.String()).
				Msg("Failed to save variant to database")
			return fmt.Errorf("failed to save variant to database: %w", err)
		}

		s.logger.Info().
			Str("function", "variantService.UpdateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Str("variant_id", variant.ID.String()).
			Str("weight", newVariant.weight).
			Str("grind", newVariant.grind).
			Msg("Successfully created new variant")
	}

	// Determine which variants need to be removed
	removedVariants := make([]*models.Variant, 0)
	
	// Check if the combination still exists in the new options
	for _, variant := range existingVariants {
		weightExists := false
		grindExists := false
		
		// Check if the variant's weight is still in the weight options
		if weightOptions, ok := product.Options["weight"]; ok {
			for _, w := range weightOptions {
				if w == variant.Weight {
					weightExists = true
					break
				}
			}
		}
		
		// Check if the variant's grind is still in the grind options
		if grindOptions, ok := product.Options["grind"]; ok {
			for _, g := range grindOptions {
				if g == variant.Grind {
					grindExists = true
					break
				}
			}
		}
		
		// If either the weight or grind is no longer in the options, the variant should be removed
		if !weightExists || !grindExists {
			removedVariants = append(removedVariants, variant)
			s.logger.Debug().
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("variant_id", variant.ID.String()).
				Str("weight", variant.Weight).
				Str("grind", variant.Grind).
				Msg("Variant marked for removal")
		}
	}

	// Remove the variants that are no longer valid
	for _, variant := range removedVariants {
		// Check if this variant has any active subscriptions
		// Ideally, we would have a subscription repository method for this
		// But for now, we'll just log a warning and proceed with deletion
		s.logger.Warn().
			Str("function", "variantService.UpdateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Str("variant_id", variant.ID.String()).
			Msg("Checking for active subscriptions for variant before deletion")
		
		// Archive the Stripe product if we have a Stripe ID for it
		// For this implementation, we would need to have the Stripe product ID stored with the variant
		// For now, we'll just log a warning
		s.logger.Warn().
			Str("function", "variantService.UpdateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Str("variant_id", variant.ID.String()).
			Msg("Would archive Stripe product for variant here")
		
		// Delete the variant from the database
		if err := s.variantRepo.Delete(ctx, variant.ID); err != nil {
			s.logger.Error().
				Err(err).
				Str("function", "variantService.UpdateVariantsForProduct").
				Str("product_id", product.ID.String()).
				Str("variant_id", variant.ID.String()).
				Msg("Failed to delete variant from database")
			return fmt.Errorf("failed to delete variant from database: %w", err)
		}
		
		s.logger.Info().
			Str("function", "variantService.UpdateVariantsForProduct").
			Str("product_id", product.ID.String()).
			Str("variant_id", variant.ID.String()).
			Msg("Successfully deleted variant")
	}

	s.logger.Info().
		Str("function", "variantService.UpdateVariantsForProduct").
		Str("product_id", product.ID.String()).
		Int("new_variants", len(newVariants)).
		Int("removed_variants", len(removedVariants)).
		Msg("Successfully updated variants for product")

	return nil
}

// Delete removes a variant from the system
func (s *variantService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug().
		Str("function", "variantService.Delete").
		Str("variant_id", id.String()).
		Msg("Starting variant deletion")

	// 1. Get existing variant from repository
	s.logger.Debug().
		Str("function", "variantService.Delete").
		Str("variant_id", id.String()).
		Msg("Retrieving variant from repository")

	variant, err := s.variantRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Delete").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to retrieve variant")
		return fmt.Errorf("failed to retrieve variant for deletion: %w", err)
	}

	if variant == nil {
		s.logger.Error().
			Str("function", "variantService.Delete").
			Str("variant_id", id.String()).
			Msg("Variant not found")
		return fmt.Errorf("variant with ID %s not found", id)
	}

	s.logger.Debug().
		Str("function", "variantService.Delete").
		Str("variant_id", id.String()).
		Str("product_id", variant.ProductID.String()).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Msg("Variant found, proceeding with deletion")

	// 2. TODO: Check if there are any active subscriptions using this variant
	// If implementing this check, add appropriate repository method and code here

	// 3. Delete from database
	s.logger.Debug().
		Str("function", "variantService.Delete").
		Str("variant_id", id.String()).
		Msg("Deleting variant from database")

	err = s.variantRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.Delete").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to delete variant from database")
		return fmt.Errorf("failed to delete variant from database: %w", err)
	}

	s.logger.Info().
		Str("function", "variantService.Delete").
		Str("variant_id", id.String()).
		Str("product_id", variant.ProductID.String()).
		Str("weight", variant.Weight).
		Str("grind", variant.Grind).
		Msg("Variant successfully deleted")

	return nil
}

// UpdateStockLevel updates the stock level of a variant
func (s *variantService) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	s.logger.Debug().
		Str("function", "variantService.UpdateStockLevel").
		Str("variant_id", id.String()).
		Int("new_quantity", quantity).
		Msg("Starting stock level update")

	// 1. Validate inputs
	if id == uuid.Nil {
		s.logger.Error().
			Str("function", "variantService.UpdateStockLevel").
			Msg("Nil UUID provided for variant ID")
		return fmt.Errorf("invalid variant ID: nil UUID")
	}

	if quantity < 0 {
		s.logger.Error().
			Str("function", "variantService.UpdateStockLevel").
			Int("quantity", quantity).
			Msg("Negative quantity provided")
		return fmt.Errorf("stock level cannot be negative")
	}

	// 2. Get existing variant to verify it exists
	s.logger.Debug().
		Str("function", "variantService.UpdateStockLevel").
		Str("variant_id", id.String()).
		Msg("Retrieving existing variant")

	existingVariant, err := s.variantRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.UpdateStockLevel").
			Err(err).
			Str("variant_id", id.String()).
			Msg("Failed to retrieve existing variant")
		return fmt.Errorf("failed to retrieve existing variant: %w", err)
	}

	if existingVariant == nil {
		s.logger.Error().
			Str("function", "variantService.UpdateStockLevel").
			Str("variant_id", id.String()).
			Msg("Variant not found")
		return fmt.Errorf("variant with ID %s not found", id)
	}

	// Record the old stock level for logging
	oldStockLevel := existingVariant.StockLevel

	// 3. Update stock level in repository
	s.logger.Debug().
		Str("function", "variantService.UpdateStockLevel").
		Str("variant_id", id.String()).
		Int("old_stock", oldStockLevel).
		Int("new_stock", quantity).
		Msg("Updating stock level in repository")

	err = s.variantRepo.UpdateStockLevel(ctx, id, quantity)
	if err != nil {
		s.logger.Error().
			Str("function", "variantService.UpdateStockLevel").
			Err(err).
			Str("variant_id", id.String()).
			Int("quantity", quantity).
			Msg("Failed to update stock level in repository")
		return fmt.Errorf("failed to update stock level: %w", err)
	}

	// 4. Log warning if stock level is low
	if quantity < 10 {
		s.logger.Warn().
			Str("function", "variantService.UpdateStockLevel").
			Str("variant_id", id.String()).
			Int("stock_level", quantity).
			Msg("Variant now has low stock level")
	}

	s.logger.Info().
		Str("function", "variantService.UpdateStockLevel").
		Str("variant_id", id.String()).
		Int("old_stock", oldStockLevel).
		Int("new_stock", quantity).
		Int("difference", quantity-oldStockLevel).
		Msg("Stock level updated successfully")

	return nil
}

// GetAvailableOptions returns the available weight and grind options for variants
func (s *variantService) GetAvailableOptions() (*dto.VariantOptionsResponse, error) {
	s.logger.Debug().
		Str("function", "variantService.GetAvailableOptions").
		Msg("Getting available variant options")

	// Convert model option types to string slices
	weightOptions := make([]string, 0, len(models.GetWeightOptions()))
	for _, option := range models.GetWeightOptions() {
		weightOptions = append(weightOptions, string(option))
	}

	grindOptions := make([]string, 0, len(models.GetGrindOptions()))
	for _, option := range models.GetGrindOptions() {
		grindOptions = append(grindOptions, string(option))
	}

	options := &dto.VariantOptionsResponse{
		Weights: weightOptions,
		Grinds:  grindOptions,
	}

	s.logger.Info().
		Str("function", "variantService.GetAvailableOptions").
		Int("weight_options_count", len(weightOptions)).
		Int("grind_options_count", len(grindOptions)).
		Strs("weights", weightOptions).
		Strs("grinds", grindOptions).
		Msg("Available variant options retrieved successfully")

	return options, nil
}
