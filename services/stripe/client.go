// services/stripe/client.go
package stripe

import (
	"os"

	"github.com/stripe/stripe-go/v74/client"
)

// Client is a wrapper around the Stripe client
type Client struct {
	sc *client.API
}

// NewClient creates a new Stripe client
func NewClient() *Client {
	apiKey := os.Getenv("STRIPE_SECRET_KEY")
	if apiKey == "" {
		// For development, you could fall back to a hardcoded test key
		// but in production, this should always come from environment variables
		apiKey = "sk_test_your_test_key_here"
	}

	sc := &client.API{}
	sc.Init(apiKey, nil)

	return &Client{
		sc: sc,
	}
}