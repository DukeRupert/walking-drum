package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// CheckoutItem represents a line item for checkout
type CheckoutItem struct {
    PriceID  string
    Quantity int
}

// CreateEmbeddedCheckoutSession creates a Stripe checkout session for embedded checkout
func (c *client) CreateEmbeddedCheckoutSession(
	customerStripeID string,
	priceStripeID string,
	productName string,
	quantity int,
	returnURL string,
) (*stripe.CheckoutSession, error) {
	if quantity < 1 {
		quantity = 1
	}
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerStripeID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceStripeID),
				Quantity: stripe.Int64(int64(quantity)),
			},
		},
		Mode:      stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ReturnURL: stripe.String(returnURL),
		UIMode:    stripe.String("embedded"),
	}

	// Add metadata for better reporting and integration
	if productName != "" {
		if params.Metadata == nil {
			params.Metadata = make(map[string]string)
		}
		params.Metadata["product_name"] = productName
	}

	session, err := session.New(params)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create checkout session")
		return nil, err
	}

	return session, nil
}

// CreateMultiItemCheckoutSession creates a Stripe checkout session with multiple items
func (c *client) CreateMultiItemCheckoutSession(
	customerStripeID string,
	items []CheckoutItem,
	metadata map[string]string,
	returnURL string,
) (*stripe.CheckoutSession, error) {
	// Create the line items for the checkout session
	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	
	for _, item := range items {
		if item.PriceID == "" {
			c.logger.Error().Interface("item", item).Msg("Invalid price ID in line item")
			return nil, fmt.Errorf("invalid price ID in line item")
		}
		
		// Ensure quantity is valid
		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}
		
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(quantity)),
		})
	}
	
	// Create checkout session params
	params := &stripe.CheckoutSessionParams{
		Customer:  stripe.String(customerStripeID),
		LineItems: lineItems,
		Mode:      stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ReturnURL: stripe.String(returnURL),
		UIMode:    stripe.String("embedded"),
	}
	
	// Add metadata if provided
	if len(metadata) > 0 {
		params.Metadata = metadata
	}
	
	// Create the checkout session
	session, err := session.New(params)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create multi-item checkout session")
		return nil, err
	}
	
	return session, nil
}

// RetrieveCheckoutSession retrieves a Stripe checkout session
func (c *client) RetrieveCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	session, err := session.Get(sessionID, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to retrieve checkout session")
		return nil, err
	}

	return session, nil
}


