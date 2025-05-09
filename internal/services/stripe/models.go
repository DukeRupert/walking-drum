// internal/services/stripe/models.go
package stripe

// ProductUpdateParams represents the parameters for updating a Stripe product
type ProductUpdateParams struct {
	Name        string
	Description string
	Images      []string
	Active      bool
	Metadata    map[string]string
}

// Product represents a Stripe product
type Product struct {
	ID          string
	Name        string
	Description string
	Images      []string
	Active      bool
	Metadata    map[string]string
}