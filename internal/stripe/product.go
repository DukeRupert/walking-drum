// internal/stripe/product.go
package stripe

import (
	"context"
	"errors"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/product"
)

// ProductCreateParams defines the parameters for creating a Stripe product
type ProductCreateParams struct {
	Name        string
	Description string
	Images      []string
	Active      bool
	Metadata    map[string]string
}

// CreateProduct creates a new product in Stripe
func (c *Client) CreateProduct(ctx context.Context, params *ProductCreateParams) (*stripe.Product, error) {
	if params == nil {
		return nil, errors.New("params cannot be nil")
	}

	productParams := &stripe.ProductParams{
		Name:        stripe.String(params.Name),
		Description: stripe.String(params.Description),
		Active:      stripe.Bool(params.Active),
	}

	if len(params.Images) > 0 {
		productParams.Images = stripe.StringSlice(params.Images)
	}

	if len(params.Metadata) > 0 {
		productParams.Metadata = make(map[string]string)
		for k, v := range params.Metadata {
			productParams.Metadata[k] = v
		}
	}

	return product.New(productParams)
}

// GetProduct retrieves a product from Stripe by ID
func (c *Client) GetProduct(ctx context.Context, id string) (*stripe.Product, error) {
	return product.Get(id, nil)
}

// UpdateProduct updates an existing product in Stripe
func (c *Client) UpdateProduct(ctx context.Context, id string, params *ProductCreateParams) (*stripe.Product, error) {
	if params == nil {
		return nil, errors.New("params cannot be nil")
	}

	productParams := &stripe.ProductParams{
		Name:        stripe.String(params.Name),
		Description: stripe.String(params.Description),
		Active:      stripe.Bool(params.Active),
	}

	if len(params.Images) > 0 {
		productParams.Images = stripe.StringSlice(params.Images)
	}

	if len(params.Metadata) > 0 {
		productParams.Metadata = make(map[string]string)
		for k, v := range params.Metadata {
			productParams.Metadata[k] = v
		}
	}

	return product.Update(id, productParams)
}

// ArchiveProduct marks a product as inactive in Stripe
func (c *Client) ArchiveProduct(ctx context.Context, id string) error {
	_, err := product.Update(id, &stripe.ProductParams{
		Active: stripe.Bool(false),
	})
	return err
}

// ListProducts retrieves a list of products from Stripe
func (c *Client) ListProducts(ctx context.Context, limit int, active *bool) ([]*stripe.Product, error) {
	params := &stripe.ProductListParams{}

	// Set the limit using the embedded ListParams
	params.ListParams.Limit = stripe.Int64(int64(limit))

	if active != nil {
		params.Active = stripe.Bool(*active)
	}

	iter := product.List(params)
	products := make([]*stripe.Product, 0)

	for iter.Next() {
		products = append(products, iter.Product())
	}

	return products, iter.Err()
}
