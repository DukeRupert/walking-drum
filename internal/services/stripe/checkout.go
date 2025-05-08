package stripe

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// CreateSubscriptionCheckoutSession creates a checkout session for a subscription
func (c *Client) CreateSubscriptionCheckoutSession(
	customerStripeID string,
	priceStripeID string,
	successURL string,
	cancelURL string,
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
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		UIMode:     stripe.String("embedded"),
	}

	s, err := session.New(params)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create checkout session")
		return nil, err
	}

	return s, nil
}