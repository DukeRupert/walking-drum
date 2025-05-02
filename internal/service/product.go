// File: internal/service/product_service.go
package service

import (
	"context"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/models"
	"github.com/dukerupert/walking-drum/internal/repository"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
)

// ProductService implementation
type ProductService struct {
	productRepo      repository.ProductRepository
	productPriceRepo repository.ProductPriceRepository
}

// NewProductService creates a new product service
func NewProductService(
	productRepo repository.ProductRepository,
	productPriceRepo repository.ProductPriceRepository,
) *ProductService {
	return &ProductService{
		productRepo:      productRepo,
		productPriceRepo: productPriceRepo,
	}
}

// CreateProduct creates a new product and syncs it to Stripe
func (s *ProductService) CreateProduct(ctx context.Context, p *models.Product) error {
	// Set defaults
	p.Active = true
	
	// Create product in Stripe first
	stripeProduct, err := s.createStripeProduct(p)
	if err != nil {
		return fmt.Errorf("failed to create product in Stripe: %w", err)
	}
	
	// Store Stripe ID
	p.StripeProductID = stripeProduct.ID
	
	// Create in database
	if err := s.productRepo.Create(ctx, p); err != nil {
		// If database creation fails, we should try to delete from Stripe to keep consistency
		_, delErr := product.Del(stripeProduct.ID, nil)
		if delErr != nil {
			// Log this error but return the original error
			fmt.Printf("WARNING: Could not delete Stripe product after DB creation failed: %v\n", delErr)
		}
		return fmt.Errorf("failed to create product in database: %w", err)
	}
	
	return nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	// Get product with prices
	product, err := s.productRepo.GetWithPrices(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	
	return product, nil
}

// ListActiveProducts retrieves all active products with their prices
func (s *ProductService) ListActiveProducts(ctx context.Context) ([]*models.Product, error) {
	products, err := s.productRepo.ListActiveWithPrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active products: %w", err)
	}
	
	return products, nil
}

// UpdateProduct updates a product and syncs changes to Stripe
func (s *ProductService) UpdateProduct(ctx context.Context, p *models.Product) error {
	// Get existing product to ensure it exists
	_, err := s.productRepo.GetByID(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing product: %w", err)
	}
	
	// Update in Stripe
	if p.StripeProductID != "" {
		params := &stripe.ProductParams{
			Name:        stripe.String(p.Name),
			Description: stripe.String(p.Description),
			Active:      stripe.Bool(p.Active),
		}
		
		// Add metadata separately
		params.AddMetadata("origin", p.Origin)
		params.AddMetadata("roast_level", p.RoastLevel)
		
		_, err := product.Update(p.StripeProductID, params)
		if err != nil {
			return fmt.Errorf("failed to update product in Stripe: %w", err)
		}
	} else {
		// If product doesn't have a Stripe ID yet, create it
		stripeProduct, err := s.createStripeProduct(p)
		if err != nil {
			return fmt.Errorf("failed to create product in Stripe: %w", err)
		}
		
		// Store Stripe ID
		p.StripeProductID = stripeProduct.ID
	}
	
	// Update in database
	if err := s.productRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("failed to update product in database: %w", err)
	}
	
	return nil
}

// DeleteProduct deletes a product and associated resources
func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	// Get the product to ensure it exists and to get Stripe ID
	productObj, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	
	// Delete in Stripe if ID exists
	if productObj.StripeProductID != "" {
		_, err := product.Del(productObj.StripeProductID, nil)
		if err != nil {
			return fmt.Errorf("failed to delete product in Stripe: %w", err)
		}
	}
	
	// Delete in database
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product from database: %w", err)
	}
	
	return nil
}

// AddPrice adds a new price to a product
func (s *ProductService) AddPrice(ctx context.Context, productID int64, priceObj *models.ProductPrice) error {
	// First get the product to ensure it exists
	productObj, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	
	// Set product ID in price
	priceObj.ProductID = productID
	priceObj.Active = true
	
	// Create in Stripe
	stripePrice, err := s.createStripePrice(productObj, priceObj)
	if err != nil {
		return fmt.Errorf("failed to create price in Stripe: %w", err)
	}
	
	// Store Stripe ID
	priceObj.StripePriceID = stripePrice.ID
	
	// Create in database
	if err := s.productPriceRepo.Create(ctx, priceObj); err != nil {
		// If database creation fails, try to delete from Stripe
		_, delErr := price.Del(stripePrice.ID, nil)
		if delErr != nil {
			// Log this error but return the original error
			fmt.Printf("WARNING: Could not delete Stripe price after DB creation failed: %v\n", delErr)
		}
		return fmt.Errorf("failed to create price in database: %w", err)
	}
	
	return nil
}

// UpdatePrice updates an existing price
func (s *ProductService) UpdatePrice(ctx context.Context, priceObj *models.ProductPrice) error {
	// Get existing price to ensure it exists
	existing, err := s.productPriceRepo.GetByID(ctx, priceObj.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing price: %w", err)
	}
	
	// Update in Stripe - Note: Stripe doesn't allow price updates directly
	// We need to archive the old price and create a new one
	if existing.StripePriceID != "" {
		// Deactivate old price in Stripe
		params := &stripe.PriceParams{
			Active: stripe.Bool(false),
		}
		_, err := price.Update(existing.StripePriceID, params)
		if err != nil {
			return fmt.Errorf("failed to deactivate old price in Stripe: %w", err)
		}
		
		// Get the product
		productObj, err := s.productRepo.GetByID(ctx, priceObj.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get product for price: %w", err)
		}
		
		// Create new price in Stripe
		stripePrice, err := s.createStripePrice(productObj, priceObj)
		if err != nil {
			return fmt.Errorf("failed to create new price in Stripe: %w", err)
		}
		
		// Update Stripe ID
		priceObj.StripePriceID = stripePrice.ID
	}
	
	// Update in database
	if err := s.productPriceRepo.Update(ctx, priceObj); err != nil {
		return fmt.Errorf("failed to update price in database: %w", err)
	}
	
	return nil
}

// RemovePrice deactivates a price
func (s *ProductService) RemovePrice(ctx context.Context, priceID int64) error {
	// Get the price
	priceObj, err := s.productPriceRepo.GetByID(ctx, priceID)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}
	
	// Deactivate in Stripe
	if priceObj.StripePriceID != "" {
		params := &stripe.PriceParams{
			Active: stripe.Bool(false),
		}
		_, err := price.Update(priceObj.StripePriceID, params)
		if err != nil {
			return fmt.Errorf("failed to deactivate price in Stripe: %w", err)
		}
	}
	
	// We don't delete prices, just set them as inactive
	priceObj.Active = false
	
	// Update in database
	if err := s.productPriceRepo.Update(ctx, priceObj); err != nil {
		return fmt.Errorf("failed to deactivate price in database: %w", err)
	}
	
	return nil
}

// SyncFromStripe syncs a product from Stripe to the database
func (s *ProductService) SyncFromStripe(ctx context.Context, stripeProductID string) error {
	// Get product from Stripe
	stripeProduct, err := product.Get(stripeProductID, nil)
	if err != nil {
		return fmt.Errorf("failed to get product from Stripe: %w", err)
	}
	
	// Check if product already exists in our database
	var dbProduct *models.Product
	dbProduct, err = s.productRepo.GetByStripeID(ctx, stripeProductID)
	isNew := err != nil // If error, product doesn't exist yet
	
	if isNew {
		// Create new product
		dbProduct = &models.Product{
			StripeProductID: stripeProduct.ID,
			Name:            stripeProduct.Name,
			Description:     stripeProduct.Description,
			Origin:          stripeProduct.Metadata["origin"],
			RoastLevel:      stripeProduct.Metadata["roast_level"],
			Active:          stripeProduct.Active,
		}
		
		if err := s.productRepo.Create(ctx, dbProduct); err != nil {
			return fmt.Errorf("failed to create product from Stripe: %w", err)
		}
	} else {
		// Update existing product
		dbProduct.Name = stripeProduct.Name
		dbProduct.Description = stripeProduct.Description
		dbProduct.Origin = stripeProduct.Metadata["origin"]
		dbProduct.RoastLevel = stripeProduct.Metadata["roast_level"]
		dbProduct.Active = stripeProduct.Active
		
		if err := s.productRepo.Update(ctx, dbProduct); err != nil {
			return fmt.Errorf("failed to update product from Stripe: %w", err)
		}
	}
	
	// Now sync prices
	// List all prices for this product from Stripe
	params := &stripe.PriceListParams{
		Product: stripe.String(stripeProductID),
	}
	
	iter := price.List(params)
	for iter.Next() {
		stripePrice := iter.Price()
		
		// Try to find existing price in our database
		var dbPrice *models.ProductPrice
		dbPrice, err = s.productPriceRepo.GetByStripeID(ctx, stripePrice.ID)
		isPriceNew := err != nil // If error, price doesn't exist yet
		
		// Extract weight and grind from metadata
		weight := stripePrice.Metadata["weight"]
		grind := stripePrice.Metadata["grind"]
		
		// Convert Stripe price (in cents) to dollars
		priceAmount := float64(stripePrice.UnitAmount) / 100.0
		
		if isPriceNew {
			// Create new price
			dbPrice = &models.ProductPrice{
				ProductID:     dbProduct.ID,
				StripePriceID: stripePrice.ID,
				Weight:        weight,
				Grind:         grind,
				Price:         priceAmount,
				IsDefault:     len(dbProduct.Prices) == 0, // First price is default
				Active:        stripePrice.Active,
			}
			
			if err := s.productPriceRepo.Create(ctx, dbPrice); err != nil {
				return fmt.Errorf("failed to create price from Stripe: %w", err)
			}
		} else {
			// Update existing price
			dbPrice.Weight = weight
			dbPrice.Grind = grind
			dbPrice.Price = priceAmount
			dbPrice.Active = stripePrice.Active
			
			if err := s.productPriceRepo.Update(ctx, dbPrice); err != nil {
				return fmt.Errorf("failed to update price from Stripe: %w", err)
			}
		}
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error iterating Stripe prices: %w", err)
	}
	
	return nil
}

// SyncToStripe syncs a product and all its prices to Stripe
func (s *ProductService) SyncToStripe(ctx context.Context, productID int64) error {
	// Get product with prices
	productObj, err := s.productRepo.GetWithPrices(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	
	// Sync product to Stripe
	var stripeProduct *stripe.Product
	
	if productObj.StripeProductID != "" {
		// Update existing product
		params := &stripe.ProductParams{
			Name:        stripe.String(productObj.Name),
			Description: stripe.String(productObj.Description),
			Active:      stripe.Bool(productObj.Active),
		}
		
		// Add metadata separately
		params.AddMetadata("origin", productObj.Origin)
		params.AddMetadata("roast_level", productObj.RoastLevel)
		
		stripeProduct, err = product.Update(productObj.StripeProductID, params)
		if err != nil {
			return fmt.Errorf("failed to update product in Stripe: %w", err)
		}
	} else {
		// Create new product
		stripeProduct, err = s.createStripeProduct(productObj)
		if err != nil {
			return fmt.Errorf("failed to create product in Stripe: %w", err)
		}
		
		// Update Stripe ID in database
		productObj.StripeProductID = stripeProduct.ID
		if err := s.productRepo.Update(ctx, productObj); err != nil {
			return fmt.Errorf("failed to update product with Stripe ID: %w", err)
		}
	}
	
	// Sync prices
	for i := range productObj.Prices {
		p := &productObj.Prices[i]
		if p.StripePriceID != "" {
			// Stripe doesn't allow updating prices, only archiving and creating new ones
			// Check if the price is inactive in our DB but active in Stripe
			if !p.Active {
				// Deactivate in Stripe
				params := &stripe.PriceParams{
					Active: stripe.Bool(false),
				}
				_, err := price.Update(p.StripePriceID, params)
				if err != nil {
					return fmt.Errorf("failed to deactivate price in Stripe: %w", err)
				}
			}
			// If price is active in our DB, we assume it's already properly set up in Stripe
		} else {
			// Create new price in Stripe
			stripePrice, err := s.createStripePrice(productObj, p)
			if err != nil {
				return fmt.Errorf("failed to create price in Stripe: %w", err)
			}
			
			// Update Stripe ID in database
			p.StripePriceID = stripePrice.ID
			if err := s.productPriceRepo.Update(ctx, p); err != nil {
				return fmt.Errorf("failed to update price with Stripe ID: %w", err)
			}
		}
	}
	
	return nil
}

// Helper to create a product in Stripe
func (s *ProductService) createStripeProduct(p *models.Product) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(p.Name),
		Description: stripe.String(p.Description),
		Active:      stripe.Bool(p.Active),
	}
	
	// Add metadata separately
	params.AddMetadata("origin", p.Origin)
	params.AddMetadata("roast_level", p.RoastLevel)
	
	return product.New(params)
}

// Helper to create a price in Stripe
func (s *ProductService) createStripePrice(product *models.Product, p *models.ProductPrice) (*stripe.Price, error) {
	// Convert price to cents for Stripe
	unitAmount := int64(p.Price * 100)
	
	params := &stripe.PriceParams{
		Currency:   stripe.String("usd"),
		Product:    stripe.String(product.StripeProductID),
		UnitAmount: stripe.Int64(unitAmount),
		Active:     stripe.Bool(p.Active),
	}
	
	// Add metadata separately
	params.AddMetadata("weight", p.Weight)
	params.AddMetadata("grind", p.Grind)
	
	// Add recurring details
	params.Recurring = &stripe.PriceRecurringParams{
		Interval:      stripe.String("month"),
		IntervalCount: stripe.Int64(1),
	}
	
	return price.New(params)
}