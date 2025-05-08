package stripe

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// CreateEmbeddedCheckoutSession creates a Stripe checkout session for embedded checkout
func (c *Client) CreateEmbeddedCheckoutSession(
	customerStripeID string,
	priceStripeID string,
	productName string,
	returnURL string,
) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerStripeID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceStripeID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ReturnURL:  stripe.String(returnURL),
		UIMode:     stripe.String("embedded"),
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

// RetrieveCheckoutSession retrieves a Stripe checkout session
func (c *Client) RetrieveCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	session, err := session.Get(sessionID, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to retrieve checkout session")
		return nil, err
	}

	return session, nil
}
