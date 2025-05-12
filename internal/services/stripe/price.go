package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// PriceCreateParams defines parameters for creating a price in Stripe
type PriceCreateParams struct {
    ProductID     string            `json:"product_id"`
    Currency      string            `json:"currency"`
    UnitAmount    int64             `json:"unit_amount"`
    Nickname      string            `json:"nickname"`
    Recurring     *RecurringParams  `json:"recurring"`
    Active        bool              `json:"active"`
    Metadata      map[string]string `json:"metadata"`
}

// RecurringParams defines recurring parameters for a subscription price
type RecurringParams struct {
    Interval        string `json:"interval"`
    IntervalCount   int    `json:"interval_count"`
}

// CreatePrice creates a new price in Stripe
func (c *client) CreatePrice(ctx context.Context, params *PriceCreateParams) (*stripe.Price, error) {
    // Log the function entry
    c.logger.Info().
        Str("product_id", params.ProductID).
        Str("currency", params.Currency).
        Int64("unit_amount", params.UnitAmount).
        Msg("Executing CreatePrice()")

    // Check if client is initialized
    if c.api == nil {
        c.logger.Error().Msg("Stripe client not initialized")
        return nil, errors.New("stripe client not initialized")
    }

    // Create the Stripe params
    priceParams := &stripe.PriceCreateParams{
        Product:  stripe.String(params.ProductID),
        Currency: stripe.String(params.Currency),
        UnitAmount: stripe.Int64(params.UnitAmount),
        Active:   stripe.Bool(params.Active),
    }

    // Log the base parameters
    c.logger.Info().
        Str("product_id", *priceParams.Product).
        Str("currency", *priceParams.Currency).
        Int64("unit_amount", *priceParams.UnitAmount).
        Bool("active", *priceParams.Active).
        Msg("Base price parameters")

    // Add nickname if provided
    if params.Nickname != "" {
        priceParams.Nickname = stripe.String(params.Nickname)
        c.logger.Info().
            Str("nickname", params.Nickname).
            Msg("Adding nickname to price parameters")
    }

    // Add recurring parameters if provided
    if params.Recurring != nil {
        priceParams.Recurring = &stripe.PriceCreateRecurringParams{
            Interval:      stripe.String(params.Recurring.Interval),
            IntervalCount: stripe.Int64(int64(params.Recurring.IntervalCount)),
        }
        c.logger.Info().
            Str("interval", params.Recurring.Interval).
            Int("interval_count", params.Recurring.IntervalCount).
            Msg("Adding recurring parameters to price")
    }

    // Add metadata if provided
    if len(params.Metadata) > 0 {
        priceParams.Metadata = make(map[string]string)
        for k, v := range params.Metadata {
            priceParams.Metadata[k] = v
        }

        // Log each metadata key-value pair
        for k, v := range params.Metadata {
            c.logger.Info().
                Str("key", k).
                Str("value", v).
                Msg("Adding metadata to price parameters")
        }
    }

    // Log that we're about to call the Stripe API
    c.logger.Info().Msg("About to call Stripe API to create price")

    // Make the API call
    price, err := c.api.V1Prices.Create(ctx, priceParams)

    // Log the result
    if err != nil {
        c.logger.Error().
            Err(err).
            Msg("Failed to create price in Stripe")
        return nil, fmt.Errorf("failed to create price in Stripe: %w", err)
    }

    // Log successful creation with Stripe ID
    c.logger.Info().
        Str("stripe_id", price.ID).
        Str("product_id", price.Product.ID).
        Int64("unit_amount", price.UnitAmount).
        Str("currency", string(price.Currency)).
        Msg("Successfully created price in Stripe")

    return price, nil
}

// PriceUpdateParams defines parameters for updating a price in Stripe
type PriceUpdateParams struct {
    Nickname      *string           `json:"nickname,omitempty"`
    Active        *bool             `json:"active,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}

// UpdatePrice updates an existing price in Stripe
func (c *client) UpdatePrice(ctx context.Context, stripeID string, params *PriceUpdateParams) (*stripe.Price, error) {
    // Log the function entry
    c.logger.Info().
        Str("stripe_id", stripeID).
        Msg("Executing UpdatePrice()")

    // Check if client is initialized
    if c.api == nil {
        c.logger.Error().Msg("Stripe client not initialized")
        return nil, errors.New("stripe client not initialized")
    }

    // Validate input
    if stripeID == "" {
        c.logger.Error().Msg("Empty Stripe price ID provided")
        return nil, errors.New("stripe price ID cannot be empty")
    }

    // Create the Stripe params
    priceParams := &stripe.PriceUpdateParams{}

    // Log what fields we're updating
    c.logger.Debug().
        Str("stripe_id", stripeID).
        Msg("Creating update parameters for price")

    // Add nickname if provided
    if params.Nickname != nil {
        priceParams.Nickname = params.Nickname
        c.logger.Info().
            Str("nickname", *params.Nickname).
            Msg("Updating price nickname")
    }

    // Add active status if provided
    if params.Active != nil {
        priceParams.Active = params.Active
        c.logger.Info().
            Bool("active", *params.Active).
            Msg("Updating price active status")
    }

    // Add metadata if provided
    if len(params.Metadata) > 0 {
        priceParams.Metadata = make(map[string]string)
        for k, v := range params.Metadata {
            priceParams.Metadata[k] = v
        }

        // Log each metadata key-value pair
        for k, v := range params.Metadata {
            c.logger.Info().
                Str("key", k).
                Str("value", v).
                Msg("Updating metadata for price")
        }
    }

    // Log that we're about to call the Stripe API
    c.logger.Info().
        Str("stripe_id", stripeID).
        Msg("About to call Stripe API to update price")

    // Make the API call
    price, err := c.api.V1Prices.Update(ctx, stripeID, priceParams)

    // Log the result
    if err != nil {
        c.logger.Error().
            Err(err).
            Str("stripe_id", stripeID).
            Msg("Failed to update price in Stripe")
        return nil, fmt.Errorf("failed to update price in Stripe: %w", err)
    }

    // Log successful update
    c.logger.Info().
        Str("stripe_id", price.ID).
        Str("product_id", price.Product.ID).
        Bool("active", price.Active).
        Msg("Successfully updated price in Stripe")

    return price, nil
}

// ArchivePrice deactivates a price in Stripe
func (c *client) ArchivePrice(ctx context.Context, stripeID string) error {
    // Log the function entry
    c.logger.Info().
        Str("stripe_id", stripeID).
        Msg("Executing ArchivePrice()")

    // Check if client is initialized
    if c.api == nil {
        c.logger.Error().Msg("Stripe client not initialized")
        return errors.New("stripe client not initialized")
    }

    // Validate input
    if stripeID == "" {
        c.logger.Error().Msg("Empty Stripe price ID provided")
        return errors.New("stripe price ID cannot be empty")
    }

    c.logger.Debug().
        Str("stripe_id", stripeID).
        Msg("Creating update parameters to archive price")

    // Create the update parameters to archive the price
    // In Stripe, archiving a price is done by setting "active" to false
    params := &stripe.PriceUpdateParams{
        Active: stripe.Bool(false),
    }

    c.logger.Info().
        Str("stripe_id", stripeID).
        Bool("active", false).
        Msg("About to call Stripe API to archive price")

    // Make the API call
    price, err := c.api.V1Prices.Update(ctx, stripeID, params)

    // Log the result
    if err != nil {
        c.logger.Error().
            Err(err).
            Str("stripe_id", stripeID).
            Msg("Failed to archive price in Stripe")
        return fmt.Errorf("failed to archive price in Stripe: %w", err)
    }

    // Verify the price was actually archived
    if price.Active {
        c.logger.Warn().
            Str("stripe_id", price.ID).
            Msg("Price archiving may have failed: price still marked as active")
        return fmt.Errorf("price archiving may have failed: price still marked as active")
    }

    // Log successful archiving with Stripe ID
    c.logger.Info().
        Str("stripe_id", price.ID).
        Bool("active", price.Active).
        Str("product_id", price.Product.ID).
        Msg("Successfully archived price in Stripe")

    return nil
}