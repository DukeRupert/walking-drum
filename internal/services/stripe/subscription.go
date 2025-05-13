package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
)

// RetrieveSubscription retrieves a Stripe subscription by ID
func (c *client) RetrieveSubscription(subscriptionID string) (*stripe.Subscription, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	c.logger.Debug().Str("subscriptionID", subscriptionID).Msg("Retrieving subscription from Stripe")

	params := &stripe.SubscriptionParams{}
	subscription, err := subscription.Get(subscriptionID, params)
	if err != nil {
		c.logger.Error().Err(err).Str("subscriptionID", subscriptionID).Msg("Failed to retrieve subscription from Stripe")
		return nil, err
	}

	return subscription, nil
}
