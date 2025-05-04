// internal/repositories/interfaces/customer_repository.go
package interfaces

import (
	"context"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	// Create adds a new customer to the database
	Create(ctx context.Context, customer *models.Customer) error

	// GetByID retrieves a customer by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)

	// GetByEmail retrieves a customer by email address
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)

	// GetByStripeID retrieves a customer by its Stripe ID
	GetByStripeID(ctx context.Context, stripeID string) (*models.Customer, error)

	// List retrieves customers with optional pagination and filtering
	List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Customer, int, error)

	// Update updates an existing customer
	Update(ctx context.Context, customer *models.Customer) error

	// Delete removes a customer from the database
	Delete(ctx context.Context, id uuid.UUID) error

	// GetAddresses retrieves all addresses for a customer
	GetAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error)

	// AddAddress adds a new address for a customer
	AddAddress(ctx context.Context, address *models.Address) error

	// UpdateAddress updates an existing address
	UpdateAddress(ctx context.Context, address *models.Address) error

	// DeleteAddress removes an address
	DeleteAddress(ctx context.Context, id uuid.UUID) error

	// GetDefaultAddress gets the default shipping address for a customer
	GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*models.Address, error)

	// SetDefaultAddress sets an address as the default for a customer
	SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error
}
