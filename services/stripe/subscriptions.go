// services/stripe/subscriptions.go
package stripe

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

// SubscriptionService handles Stripe subscription operations
type SubscriptionService struct {
	client *Client
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(client *Client) *SubscriptionService {
	return &SubscriptionService{
		client: client,
	}
}

// CreateSubscriptionInput defines the input for creating a subscription
type CreateSubscriptionInput struct {
	CustomerID      string
	PriceID         string
	Quantity        int64
	PaymentMethodID string
	IdempotencyKey  string
	MetadataOrderID string
	Description     string
}

// CreateSubscription creates a new subscription in Stripe
func (s *SubscriptionService) CreateSubscription(input CreateSubscriptionInput) (*stripe.Subscription, error) {
	// Create subscription parameters
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(input.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(input.PriceID),
				Quantity: stripe.Int64(input.Quantity),
			},
		},
	}

	// Set payment method if provided
	if input.PaymentMethodID != "" {
		params.DefaultPaymentMethod = stripe.String(input.PaymentMethodID)
	}

	// Set description if provided
	if input.Description != "" {
		params.Description = stripe.String(input.Description)
	}

	// Set metadata if provided
	if input.MetadataOrderID != "" {
		params.AddMetadata("order_id", input.MetadataOrderID)
	}

	// Set idempotency key if provided
	if input.IdempotencyKey != "" {
		params.SetIdempotencyKey(input.IdempotencyKey)
	} else {
		// Generate a UUID as idempotency key if not provided
		params.SetIdempotencyKey(uuid.New().String())
	}

	// Create the subscription
	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// CancelSubscription cancels a subscription immediately
func (s *SubscriptionService) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{
		Prorate: stripe.Bool(true),
	}

	sub, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

// UpdateSubscriptionQuantity updates the quantity of a subscription
func (s *SubscriptionService) UpdateSubscriptionQuantity(subscriptionID string, priceID string, quantity int64) (*stripe.Subscription, error) {
	// First get the subscription to find the subscription item ID
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	// Find the subscription item ID for the given price
	var subscriptionItemID string
	for _, item := range sub.Items.Data {
		if item.Price.ID == priceID {
			subscriptionItemID = item.ID
			break
		}
	}

	if subscriptionItemID == "" {
		return nil, fmt.Errorf("subscription item with price ID %s not found", priceID)
	}

	// Update the subscription
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(subscriptionItemID),
				Quantity: stripe.Int64(quantity),
			},
		},
	}

	updatedSub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, err
	}

	return updatedSub, nil
}