// internal/stripe/client.go
package stripe

import (
	"github.com/stripe/stripe-go/v74/client"
)

// Client is a wrapper around the Stripe API client
type Client struct {
	api *client.API
}

// NewClient creates a new Stripe client
func NewClient(key string) *Client {
	api := client.New(key, nil)
	return &Client{api: api}
}