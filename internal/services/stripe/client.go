// internal/stripe/client.go
package stripe

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

// StripeService defines operations for interacting with the Stripe API
type StripeService interface {
	// ProcessWebhook handles incoming webhook events from Stripe
	ProcessWebhook(ctx context.Context, payload []byte, signature, webhookSecret string) error

	// Other Stripe operations would be defined here
}

// Client is a wrapper around the Stripe API client
type Client struct {
	api    *client.API
	logger zerolog.Logger
}

// NewClient creates a new Stripe client
func NewClient(secretKey string, logger zerolog.Logger) *Client {
	// Ensure the key is not empty
	if secretKey == "" {
		// Log this error
		logger.Error().Msg("Missing stripe secretKey")
		return nil
	}

	// Set the API key for the stripe package as a whole
	stripe.Key = secretKey

	// Create a new Stripe client with the secret key
	api := client.New(secretKey, nil)

	return &Client{
		api:    api,
		logger: logger.With().Str("component", "stripe_service").Logger(),
	}
}
