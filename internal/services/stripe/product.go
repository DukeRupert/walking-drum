// internal/stripe/product.go
package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/product"
)

// ProductCreateParams defines the parameters for creating a Stripe product
type ProductCreateParams struct {
	Name        string
	Description string
	Images      []string
	Active      bool
	Metadata    map[string]string
}

// CreateProduct creates a new product in Stripe
func (c *Client) CreateProduct(ctx context.Context, params *ProductCreateParams) (*stripe.Product, error) {
	// Log the function entry
	c.logger.Info().Msg("Executing CreateProduct()")

	// Check if client is initialized
	if c.api == nil {
		c.logger.Error().Msg("Stripe client not initialized")
		return nil, errors.New("stripe client not initialized")
	}

	// Create the Stripe params
	productParams := &stripe.ProductCreateParams{
		Name:        stripe.String(params.Name),
		Description: stripe.String(params.Description),
		Active:      stripe.Bool(params.Active),
	}

	// Log the base parameters
	c.logger.Info().
		Str("name", *productParams.Name).
		Str("description", *productParams.Description).
		Bool("active", *productParams.Active).
		Msg("Base product parameters")

	// Add images if provided
	if len(params.Images) > 0 {
		productParams.Images = stripe.StringSlice(params.Images)
		c.logger.Info().
			Strs("images", params.Images).
			Msg("Adding images to product parameters")
	} else {
		c.logger.Info().Msg("No images provided for product")
	}

	// Add metadata if provided
	if len(params.Metadata) > 0 {
		productParams.Metadata = make(map[string]string)
		for k, v := range params.Metadata {
			productParams.Metadata[k] = v
		}

		// Log each metadata key-value pair
		for k, v := range params.Metadata {
			c.logger.Info().
				Str("key", k).
				Str("value", v).
				Msg("Adding metadata to product parameters")
		}
	} else {
		c.logger.Info().Msg("No metadata provided for product")
	}

	// Log the final productParams structure
	// We can't directly log the stripe.ProductParams struct as it contains pointers
	// So we'll log a summary
	c.logger.Info().
		Str("params_summary", fmt.Sprintf(
			"Final product params - Name: %s, Description: %s, Active: %t, Images count: %d, Metadata count: %d",
			*productParams.Name,
			*productParams.Description,
			*productParams.Active,
			len(productParams.Images),
			len(productParams.Metadata),
		)).
		Msg("About to call Stripe API")

	// Make the API call
	product, err := c.api.V1Products.Create(ctx, productParams)

	// Log the result
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create product in Stripe")
		return nil, err
	}

	// Log successful creation with Stripe ID
	c.logger.Info().
		Str("stripe_id", product.ID).
		Str("product_name", product.Name).
		Msg("Successfully created product in Stripe")

	return product, nil
}

// GetProduct retrieves a product from Stripe by ID
func (c *Client) GetProduct(ctx context.Context, id string) (*stripe.Product, error) {
	return product.Get(id, nil)
}

// UpdateProduct updates an existing product in Stripe
func (c *Client) UpdateProduct(ctx context.Context, id string, params *ProductCreateParams) (*stripe.Product, error) {
	if params == nil {
		return nil, errors.New("params cannot be nil")
	}

	productParams := &stripe.ProductParams{
		Name:        stripe.String(params.Name),
		Description: stripe.String(params.Description),
		Active:      stripe.Bool(params.Active),
	}

	if len(params.Images) > 0 {
		productParams.Images = stripe.StringSlice(params.Images)
	}

	if len(params.Metadata) > 0 {
		productParams.Metadata = make(map[string]string)
		for k, v := range params.Metadata {
			productParams.Metadata[k] = v
		}
	}

	return product.Update(id, productParams)
}

// ArchiveProduct archives (deactivates) a product in Stripe
func (c *Client) ArchiveProduct(ctx context.Context, stripeID string) error {
    // Log the function entry
    c.logger.Info().
        Str("stripe_id", stripeID).
        Msg("Executing ArchiveProduct()")

    // Check if client is initialized
    if c.api == nil {
        c.logger.Error().Msg("Stripe client not initialized")
        return errors.New("stripe client not initialized")
    }

    // Validate input
    if stripeID == "" {
        c.logger.Error().Msg("Empty Stripe ID provided")
        return errors.New("stripe product ID cannot be empty")
    }

    c.logger.Debug().
        Str("stripe_id", stripeID).
        Msg("Creating update parameters to archive product")

    // Create the update parameters to archive the product
    // In Stripe, archiving a product is done by setting "active" to false
    params := &stripe.ProductUpdateParams{
        Active: stripe.Bool(false),
    }

    c.logger.Info().
        Str("stripe_id", stripeID).
        Bool("active", false).
        Msg("About to call Stripe API to archive product")

    // Make the API call
    product, err := c.api.V1Products.Update(ctx, stripeID, params)

    // Log the result
    if err != nil {
        c.logger.Error().
            Err(err).
            Str("stripe_id", stripeID).
            Msg("Failed to archive product in Stripe")
        return fmt.Errorf("failed to archive product in Stripe: %w", err)
    }

    // Verify the product was actually archived
    if product.Active {
        c.logger.Warn().
            Str("stripe_id", product.ID).
            Str("product_name", product.Name).
            Msg("Product archiving may have failed: product still marked as active")
        return fmt.Errorf("product archiving may have failed: product still marked as active")
    }

    // Log successful archiving with Stripe ID and product name
    c.logger.Info().
        Str("stripe_id", product.ID).
        Str("product_name", product.Name).
        Bool("active", product.Active).
        Msg("Successfully archived product in Stripe")

    return nil
}

// ListProducts retrieves a list of products from Stripe
func (c *Client) ListProducts(ctx context.Context, limit int, active *bool) ([]*stripe.Product, error) {
	params := &stripe.ProductListParams{}

	// Set the limit using the embedded ListParams
	params.ListParams.Limit = stripe.Int64(int64(limit))

	if active != nil {
		params.Active = stripe.Bool(*active)
	}

	iter := product.List(params)
	products := make([]*stripe.Product, 0)

	for iter.Next() {
		products = append(products, iter.Product())
	}

	return products, iter.Err()
}
