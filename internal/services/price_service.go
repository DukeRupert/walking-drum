// internal/services/price_service.go
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

// PriceService defines the interface for price business logic
type PriceService interface {
	Create(ctx context.Context, priceDTO *dto.PriceCreateDTO) (*models.Price, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error)
	List(ctx context.Context, page, pageSize int, includeInactive bool) ([]*models.Price, int, error)
	ListByProductID(ctx context.Context, productID uuid.UUID, includeInactive bool) ([]*models.Price, error)
	Update(ctx context.Context, id uuid.UUID, priceDTO *dto.PriceUpdateDTO) (*models.Price, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// priceService implements the PriceService interface
type priceService struct {
	priceRepo    interfaces.PriceRepository
	productRepo  interfaces.ProductRepository
	stripeClient stripe.StripeService
	logger       zerolog.Logger
}

// NewPriceService creates a new price service
func NewPriceService(
	priceRepo interfaces.PriceRepository,
	productRepo interfaces.ProductRepository,
	stripeClient stripe.StripeService,
	logger *zerolog.Logger,
) PriceService {
	return &priceService{
		priceRepo:    priceRepo,
		productRepo:  productRepo,
		stripeClient: stripeClient,
		logger:       logger.With().Str("component", "price_service").Logger(),
	}
}

// Create adds a new price to the system
func (s *priceService) Create(ctx context.Context, priceDTO *dto.PriceCreateDTO) (*models.Price, error) {
    s.logger.Debug().
        Str("function", "priceService.Create").
        Interface("priceDTO", priceDTO).
        Msg("Starting price creation")

    // 1. Validate priceDTO
    if problems := priceDTO.Valid(ctx); len(problems) > 0 {
        s.logger.Error().
            Str("function", "priceService.Create").
            Interface("problems", problems).
            Msg("Price validation failed")
        return nil, fmt.Errorf("invalid price data: %v", problems)
    }
    
    s.logger.Debug().
        Str("function", "priceService.Create").
        Msg("Price validation passed")

    // 2. Verify that the product exists
    s.logger.Debug().
        Str("function", "priceService.Create").
        Str("product_id", priceDTO.ProductID.String()).
        Msg("Verifying product exists")
        
    product, err := s.productRepo.GetByID(ctx, priceDTO.ProductID)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.Create").
            Err(err).
            Str("product_id", priceDTO.ProductID.String()).
            Msg("Failed to retrieve associated product")
        return nil, fmt.Errorf("failed to verify product exists: %w", err)
    }

    // 3. Create price in Stripe first
    s.logger.Debug().
        Str("function", "priceService.Create").
        Str("product_id", priceDTO.ProductID.String()).
        Str("stripe_product_id", product.StripeID).
        Int64("amount", priceDTO.Amount).
        Str("currency", priceDTO.Currency).
        Str("price_type", priceDTO.Type).
        Msg("Creating price in Stripe")
    
    // Prepare Stripe price creation parameters
    priceParams := &stripe.PriceCreateParams{
        ProductID:  product.StripeID,
        Currency:   priceDTO.Currency,
        UnitAmount: priceDTO.Amount,
        Nickname:   priceDTO.Name,
        Active:     priceDTO.Active,
        Metadata:   map[string]string{},
    }
    
    // Configure recurring or one-time price based on type
    if priceDTO.Type == "recurring" {
        priceParams.Recurring = &stripe.RecurringParams{
            Interval:      priceDTO.Interval,
            IntervalCount: priceDTO.IntervalCount,
        }
        
        // Add interval metadata only for recurring prices
        priceParams.Metadata["interval"] = priceDTO.Interval
        priceParams.Metadata["interval_count"] = fmt.Sprintf("%d", priceDTO.IntervalCount)
    }
    // For "one_time", we don't need to set Recurring or interval metadata
    
    stripePrice, err := s.stripeClient.CreatePrice(ctx, priceParams)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.Create").
            Err(err).
            Str("product_id", priceDTO.ProductID.String()).
            Int64("amount", priceDTO.Amount).
            Msg("Failed to create price in Stripe")
        return nil, fmt.Errorf("failed to create price in Stripe: %w", err)
    }
    
    s.logger.Debug().
        Str("function", "priceService.Create").
        Str("stripe_price_id", stripePrice.ID).
        Str("stripe_product_id", product.StripeID).
        Msg("Successfully created price in Stripe")

    // 4. Create price in local database
    now := time.Now()
    price := &models.Price{
        ID:            uuid.New(),
        ProductID:     priceDTO.ProductID,
        Name:          priceDTO.Name,
        Amount:        priceDTO.Amount,
        Currency:      priceDTO.Currency,
        Type:          priceDTO.Type,
        Active:        priceDTO.Active,
        StripeID:      stripePrice.ID,
        CreatedAt:     now,
        UpdatedAt:     now,
    }
    
    // Only set interval fields for recurring prices
    if priceDTO.Type == "recurring" {
        price.Interval = priceDTO.Interval
        price.IntervalCount = priceDTO.IntervalCount
    } else {
        // Ensure interval fields are zero values for one_time prices
        price.Interval = ""
        price.IntervalCount = 0
    }
    
    s.logger.Debug().
        Str("function", "priceService.Create").
        Str("price_id", price.ID.String()).
        Str("stripe_id", price.StripeID).
        Str("type", price.Type).
        Msg("Preparing to save price to database")

    // 5. Save to database
    if err := s.priceRepo.Create(ctx, price); err != nil {
        s.logger.Error().
            Str("function", "priceService.Create").
            Err(err).
            Str("price_id", price.ID.String()).
            Msg("Failed to create price in database")
            
        // If database creation fails, attempt to archive/deactivate the Stripe price
        s.logger.Debug().
            Str("function", "priceService.Create").
            Str("stripe_price_id", stripePrice.ID).
            Msg("Attempting to archive Stripe price after database failure")
            
        if archiveErr := s.stripeClient.ArchivePrice(ctx, stripePrice.ID); archiveErr != nil {
            // Log this error, but continue with the main error
            s.logger.Error().
                Str("function", "priceService.Create").
                Err(archiveErr).
                Str("stripe_price_id", stripePrice.ID).
                Msg("Failed to archive Stripe price after database error")
        } else {
            s.logger.Debug().
                Str("function", "priceService.Create").
                Str("stripe_price_id", stripePrice.ID).
                Msg("Successfully archived Stripe price after database failure")
        }
        
        return nil, fmt.Errorf("failed to create price in database: %w", err)
    }

    s.logger.Info().
        Str("function", "priceService.Create").
        Str("price_id", price.ID.String()).
        Str("stripe_id", price.StripeID).
        Str("product_id", price.ProductID.String()).
        Str("name", price.Name).
        Str("type", price.Type).
        Int64("amount", price.Amount).
        Msg("Price successfully created")

    return price, nil
}

// Add this helper function
func buildPriceMetadata(priceDTO *dto.PriceCreateDTO) map[string]string {
    metadata := make(map[string]string)
    
    // Only add interval and interval_count if recurring type is "recurring"
    if priceDTO.Type == "recurring" {
        metadata["interval"] = priceDTO.Interval
        metadata["interval_count"] = fmt.Sprintf("%d", priceDTO.IntervalCount)
    }
    
    return metadata
}

// GetByID retrieves a price by its ID
func (s *priceService) GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error) {
    s.logger.Debug().
        Str("function", "priceService.GetByID").
        Str("price_id", id.String()).
        Msg("Starting price retrieval by ID")
    
    // Validate ID
    if id == uuid.Nil {
        s.logger.Error().
            Str("function", "priceService.GetByID").
            Msg("Nil UUID provided")
        return nil, fmt.Errorf("invalid price ID: nil UUID")
    }
    
    // Call repository to fetch price
    s.logger.Debug().
        Str("function", "priceService.GetByID").
        Str("price_id", id.String()).
        Msg("Calling repository to fetch price")
        
    price, err := s.priceRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.GetByID").
            Err(err).
            Str("price_id", id.String()).
            Msg("Failed to retrieve price from repository")
        return nil, fmt.Errorf("failed to retrieve price: %w", err)
    }
    
    // Check if price was found
    if price == nil {
        s.logger.Error().
            Str("function", "priceService.GetByID").
            Str("price_id", id.String()).
            Msg("Price not found")
        return nil, fmt.Errorf("price with ID %s not found", id)
    }
    
    // Get associated product information if needed
    // This is optional - you might want to include product details 
    // with the price for convenience
    product, err := s.productRepo.GetByID(ctx, price.ProductID)
    if err != nil {
        s.logger.Warn().
            Str("function", "priceService.GetByID").
            Err(err).
            Str("price_id", id.String()).
            Str("product_id", price.ProductID.String()).
            Msg("Could not retrieve associated product information")
        // We don't return an error here, as the price information was found
        // The missing product info is just a warning
    } else {
        s.logger.Debug().
            Str("function", "priceService.GetByID").
            Str("price_id", id.String()).
            Str("product_id", price.ProductID.String()).
            Str("product_name", product.Name).
            Msg("Retrieved associated product information")
    }
    
    s.logger.Info().
        Str("function", "priceService.GetByID").
        Str("price_id", id.String()).
        Str("product_id", price.ProductID.String()).
        Str("name", price.Name).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Str("interval", price.Interval).
        Int("interval_count", price.IntervalCount).
        Msg("Price successfully retrieved")
        
    return price, nil
}

// List retrieves all prices with optional filtering
func (s *priceService) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Price, int, error) {
    s.logger.Debug().
        Str("function", "priceService.List").
        Int("offset", offset).
        Int("limit", limit).
        Bool("includeInactive", includeInactive).
        Msg("Starting price listing")

    // Call repository to list prices with the provided parameters
    s.logger.Debug().
        Str("function", "priceService.List").
        Msg("Calling repository to fetch prices")
        
    prices, total, err := s.priceRepo.List(ctx, offset, limit, includeInactive)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.List").
            Err(err).
            Int("offset", offset).
            Int("limit", limit).
            Bool("includeInactive", includeInactive).
            Msg("Failed to retrieve prices from repository")
        return nil, 0, fmt.Errorf("failed to list prices: %w", err)
    }
    
    // Log the result count
    s.logger.Debug().
        Str("function", "priceService.List").
        Int("prices_count", len(prices)).
        Int("total_count", total).
        Msg("Successfully retrieved prices from repository")
    
    // Additional processing if needed (e.g., formatting, conversion)
    
    s.logger.Info().
        Str("function", "priceService.List").
        Int("total_prices", total).
        Int("returned_prices", len(prices)).
        Int("offset", offset).
        Int("limit", limit).
        Bool("includeInactive", includeInactive).
        Msg("Price listing completed successfully")
        
    return prices, total, nil
}

// ListByProductID retrieves all prices for a specific product
func (s *priceService) ListByProductID(ctx context.Context, productID uuid.UUID, includeInactive bool) ([]*models.Price, error) {
    s.logger.Debug().
        Str("function", "priceService.ListByProductID").
        Str("product_id", productID.String()).
        Bool("includeInactive", includeInactive).
        Msg("Starting price listing by product ID")

    // Validate product ID
    if productID == uuid.Nil {
        s.logger.Error().
            Str("function", "priceService.ListByProductID").
            Msg("Nil UUID provided for product ID")
        return nil, fmt.Errorf("invalid product ID: nil UUID")
    }
    
    // Call repository to list prices for the specified product
    s.logger.Debug().
        Str("function", "priceService.ListByProductID").
        Str("product_id", productID.String()).
        Msg("Calling repository to fetch prices by product ID")
        
    prices, err := s.priceRepo.ListByProductID(ctx, productID, includeInactive)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.ListByProductID").
            Err(err).
            Str("product_id", productID.String()).
            Bool("includeInactive", includeInactive).
            Msg("Failed to retrieve prices from repository")
        return nil, fmt.Errorf("failed to list prices for product %s: %w", productID, err)
    }
    
    // Log the result count
    s.logger.Debug().
        Str("function", "priceService.ListByProductID").
        Int("prices_count", len(prices)).
        Str("product_id", productID.String()).
        Msg("Successfully retrieved prices for product")
    
    // Additional processing if needed
    
    s.logger.Info().
        Str("function", "priceService.ListByProductID").
        Str("product_id", productID.String()).
        Int("prices_count", len(prices)).
        Bool("includeInactive", includeInactive).
        Msg("Price listing by product ID completed successfully")
        
    return prices, nil
}

// Update updates an existing price
func (s *priceService) Update(ctx context.Context, id uuid.UUID, priceDTO *dto.PriceUpdateDTO) (*models.Price, error) {
	// TODO: Implement price update
	// 1. Get existing price
	// 2. Update fields from DTO
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil, nil
}

// Delete removes a price from the system
func (s *priceService) Delete(ctx context.Context, id uuid.UUID) error {
    s.logger.Debug().
        Str("function", "priceService.Delete").
        Str("price_id", id.String()).
        Msg("Starting price deletion")

    // 1. Get existing price from repository
    s.logger.Debug().
        Str("function", "priceService.Delete").
        Str("price_id", id.String()).
        Msg("Retrieving price from repository")
        
    price, err := s.priceRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.Delete").
            Err(err).
            Str("price_id", id.String()).
            Msg("Failed to retrieve price")
        return fmt.Errorf("failed to retrieve price for deletion: %w", err)
    }
    
    if price == nil {
        s.logger.Error().
            Str("function", "priceService.Delete").
            Str("price_id", id.String()).
            Msg("Price not found")
        return fmt.Errorf("price with ID %s not found", id)
    }
    
    s.logger.Debug().
        Str("function", "priceService.Delete").
        Str("price_id", id.String()).
        Str("price_name", price.Name).
        Str("stripe_id", price.StripeID).
        Msg("Price found, proceeding with deletion")

    // 2. Check if there are any active subscriptions using this price
    // If you have a way to check for this, you should implement it here
    // For now, I'll just add a placeholder comment
    // This would prevent deletion of prices that are in use

    // 3. Archive in Stripe first if the price has a Stripe ID
    if price.StripeID != "" {
        s.logger.Debug().
            Str("function", "priceService.Delete").
            Str("price_id", id.String()).
            Str("stripe_id", price.StripeID).
            Msg("Archiving price in Stripe")
            
        err = s.stripeClient.ArchivePrice(ctx, price.StripeID)
        if err != nil {
            s.logger.Error().
                Str("function", "priceService.Delete").
                Err(err).
                Str("price_id", id.String()).
                Str("stripe_id", price.StripeID).
                Msg("Failed to archive price in Stripe")
            return fmt.Errorf("failed to archive price in Stripe: %w", err)
        }
        
        s.logger.Debug().
            Str("function", "priceService.Delete").
            Str("price_id", id.String()).
            Str("stripe_id", price.StripeID).
            Msg("Successfully archived price in Stripe")
    } else {
        s.logger.Warn().
            Str("function", "priceService.Delete").
            Str("price_id", id.String()).
            Msg("Price has no Stripe ID, skipping Stripe archiving")
    }

    // 4. Delete from database
    s.logger.Debug().
        Str("function", "priceService.Delete").
        Str("price_id", id.String()).
        Msg("Deleting price from database")
        
    err = s.priceRepo.Delete(ctx, id)
    if err != nil {
        s.logger.Error().
            Str("function", "priceService.Delete").
            Err(err).
            Str("price_id", id.String()).
            Msg("Failed to delete price from database")
            
        // If database deletion fails but we already archived in Stripe,
        // we should log this inconsistency
        if price.StripeID != "" {
            s.logger.Warn().
                Str("function", "priceService.Delete").
                Str("price_id", id.String()).
                Str("stripe_id", price.StripeID).
                Msg("Inconsistent state: Price archived in Stripe but not deleted from database")
        }
        
        return fmt.Errorf("failed to delete price from database: %w", err)
    }
    
    s.logger.Info().
        Str("function", "priceService.Delete").
        Str("price_id", id.String()).
        Str("price_name", price.Name).
        Str("product_id", price.ProductID.String()).
        Str("stripe_id", price.StripeID).
        Msg("Price successfully deleted")
        
    return nil
}