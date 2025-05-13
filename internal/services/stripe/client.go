// internal/stripe/client.go
package stripe

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/rs/zerolog"
	"github.com/stripe/stripe-go/v82"
)

// StripeService defines operations for interacting with the Stripe API
type StripeService interface {
	// ProcessWebhook handles incoming webhook events from Stripe
	ProcessWebhook(ctx context.Context, payload []byte, signature, webhookSecret string) error

	// Product
	CreateProduct(ctx context.Context, params *ProductCreateParams) (*stripe.Product, error)
	ListProducts(ctx context.Context, limit int, active *bool) ([]*stripe.Product, error)
	UpdateProduct(ctx context.Context, p *models.Product) error
	ArchiveProduct(ctx context.Context, productID string) error

	// Customer
	CreateCustomer(ctx context.Context, params *CustomerCreateParams) (*stripe.Customer, error)

	// Price
	CreatePrice(ctx context.Context, params *PriceCreateParams) (*stripe.Price, error)
	UpdatePrice(ctx context.Context, stripeID string, params *PriceUpdateParams) (*stripe.Price, error)
	ArchivePrice(ctx context.Context, stripeID string) error

	// Subscription
	RetrieveSubscription(subscriptionID string) (*stripe.Subscription, error)

	// Checkout
	CreateEmbeddedCheckoutSession(customerStripeID, priceStripeID, productName string, quantity int, returnURL string) (*stripe.CheckoutSession, error)
	CreateMultiItemCheckoutSession(customerStripeID string, items []CheckoutItem, metadata map[string]string, returnURL string) (*stripe.CheckoutSession, error)
	RetrieveCheckoutSession(sessionID string) (*stripe.CheckoutSession, error)
}

// Client is a wrapper around the Stripe API client
type client struct {
	api    *stripe.Client
	logger zerolog.Logger
}

// NewClient creates a new Stripe client
func NewClient(secretKey string, logger zerolog.Logger) StripeService {
	// Ensure the key is not empty
	if secretKey == "" {
		// Log this error
		logger.Error().Msg("Missing stripe secretKey")
		return nil
	}

	// Set the API key for the stripe package as a whole
	stripe.Key = secretKey

	// Create a new Stripe client with the secret key
	api := stripe.NewClient(secretKey)

	return &client{
		api:    api,
		logger: logger.With().Str("service", "stripe").Logger(),
	}
}
