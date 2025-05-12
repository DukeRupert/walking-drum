// internal/services/stripe/mocks/client.go
package mocks

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/stretchr/testify/mock"
	stripego "github.com/stripe/stripe-go/v82"
)

// Client is a mock implementation of the stripe.Client
type Client struct {
	mock.Mock
}

// ProcessWebhook mocks the ProcessWebhook method
func (m *Client) ProcessWebhook(ctx context.Context, payload []byte, signature, webhookSecret string) error {
	args := m.Called(ctx, payload, signature, webhookSecret)
	return args.Error(0)
}

func (m *Client) ListProducts(ctx context.Context, limit int, active *bool) ([]*stripego.Product, error) {
	return nil, nil
}

// CreateProduct mocks the CreateProduct method
func (m *Client) CreateProduct(ctx context.Context, params *stripe.ProductCreateParams) (*stripego.Product, error) {
	args := m.Called(ctx, params)
	
	// Handle the first return value (could be nil)
	var product *stripego.Product
	if args.Get(0) != nil {
		product = args.Get(0).(*stripego.Product)
	}
	
	return product, args.Error(1)
}

// ArchiveProduct mocks the ArchiveProduct method
func (m *Client) ArchiveProduct(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

// UpdateProduct mocks the UpdateProduct method
func (m *Client) UpdateProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// CreatePrice creates a new price in Stripe
func (m *Client) CreatePrice(ctx context.Context, params *stripe.PriceCreateParams) (*stripego.Price, error) {
	args := m.Called(ctx, params)

	// Handle the first return value (could be nil)
	var price *stripego.Price
	if args.Get(0) != nil {
		price = args.Get(0).(*stripego.Price)
	}

	return price, args.Error(1)
}

func (m *Client) UpdatePrice(ctx context.Context, stripeID string, params *stripe.PriceUpdateParams) (*stripego.Price, error) {
	return nil, nil
}

// ArchivePrice deactivates a price in Stripe
func (m *Client) ArchivePrice(ctx context.Context, stripeID string) error {
	return nil
}

func (m *Client) CreateCustomer(ctx context.Context, params *stripe.CustomerCreateParams) (*stripego.Customer, error) {
	return nil, nil
}

func (m *Client) CreateEmbeddedCheckoutSession(
	customerStripeID string,
	priceStripeID string,
	productName string,
	quantity int,
	returnURL string,
) (*stripego.CheckoutSession, error) {
	return nil, nil
}

func (m *Client) RetrieveCheckoutSession(sessionID string) (*stripego.CheckoutSession, error) {
	return nil, nil
}
