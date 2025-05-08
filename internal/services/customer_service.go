// internal/services/customer_service.go
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CustomerService defines the interface for customer business logic
type CustomerService interface {
	Create(ctx context.Context, customerDTO *dto.CustomerCreateDTO) (*models.Customer, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	GetByEmail(ctx context.Context, email string) (*models.Customer, error)
	List(ctx context.Context, page, pageSize int, includeInactive bool) ([]*models.Customer, int, error)
	Update(ctx context.Context, id uuid.UUID, customerDTO *dto.CustomerUpdateDTO) (*models.Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error)
	AddAddress(ctx context.Context, customerID uuid.UUID, addressDTO *dto.AddressCreateDTO) (*models.Address, error)
	UpdateAddress(ctx context.Context, addressID uuid.UUID, addressDTO *dto.AddressUpdateDTO) (*models.Address, error)
	DeleteAddress(ctx context.Context, addressID uuid.UUID) error
	GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*models.Address, error)
	SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error
}

// customerService implements the CustomerService interface
type customerService struct {
	customerRepo interfaces.CustomerRepository
	stripeClient *stripe.Client
	logger 	zerolog.Logger
}

// NewCustomerService creates a new customer service
func NewCustomerService(customerRepo interfaces.CustomerRepository, stripeClient *stripe.Client, logger *zerolog.Logger) CustomerService {
	return &customerService{
		customerRepo: customerRepo,
		stripeClient: stripeClient,
		logger: logger.With().Str("component", "customer_service").Logger(),
	}
}

// Create adds a new customer to the system (both in DB and Stripe)
func (s *customerService) Create(ctx context.Context, customerDTO *dto.CustomerCreateDTO) (*models.Customer, error) {
    s.logger.Debug().
        Str("function", "customerService.Create").
        Interface("customerDTO", customerDTO).
        Msg("Starting customer creation")

    // 1. Validate customerDTO
    if problems := customerDTO.Valid(ctx); len(problems) > 0 {
        s.logger.Error().
            Str("function", "customerService.Create").
            Interface("problems", problems).
            Msg("Customer validation failed")
        return nil, fmt.Errorf("invalid customer data: %v", problems)
    }
    
    s.logger.Debug().
        Str("function", "customerService.Create").
        Msg("Customer validation passed")

    // 2. Check if customer already exists by email
    s.logger.Debug().
        Str("function", "customerService.Create").
        Str("email", customerDTO.Email).
        Msg("Checking if customer already exists")
        
    existingCustomer, err := s.customerRepo.GetByEmail(ctx, customerDTO.Email)
    
	if err != nil {
        // Check if it's a "not found" error - this is expected and not an actual error
        if strings.Contains(err.Error(), "not found") {
            s.logger.Debug().
                Str("function", "customerService.Create").
                Str("email", customerDTO.Email).
                Msg("No existing customer found with this email")
        } else {
            // This is a real database error
            s.logger.Error().
                Str("function", "customerService.Create").
                Err(err).
                Str("email", customerDTO.Email).
                Msg("Error checking for existing customer")
            return nil, fmt.Errorf("error checking for existing customer: %w", err)
        }
    } else if existingCustomer != nil {
        // Customer exists
        s.logger.Error().
            Str("function", "customerService.Create").
            Str("email", customerDTO.Email).
            Str("existing_id", existingCustomer.ID.String()).
            Msg("Customer with email already exists")
        return nil, fmt.Errorf("customer with email %s already exists", customerDTO.Email)
    }

    // 3. Create customer in Stripe first
    s.logger.Debug().
        Str("function", "customerService.Create").
        Str("email", customerDTO.Email).
        Str("name", fmt.Sprintf("%s %s", customerDTO.FirstName, customerDTO.LastName)).
        Msg("Creating customer in Stripe")
        
    fullName := fmt.Sprintf("%s %s", customerDTO.FirstName, customerDTO.LastName)
    
    stripeCustomer, err := s.stripeClient.CreateCustomer(ctx, &stripe.CustomerCreateParams{
        Email: customerDTO.Email,
        Name:  fullName,
        Phone: customerDTO.PhoneNumber,
        Metadata: map[string]string{
            "first_name": customerDTO.FirstName,
            "last_name":  customerDTO.LastName,
        },
    })
    if err != nil {
        s.logger.Error().
            Str("function", "customerService.Create").
            Err(err).
            Str("email", customerDTO.Email).
            Msg("Failed to create customer in Stripe")
        return nil, fmt.Errorf("failed to create customer in Stripe: %w", err)
    }
    
    s.logger.Debug().
        Str("function", "customerService.Create").
        Str("stripe_id", stripeCustomer.ID).
        Str("email", customerDTO.Email).
        Msg("Successfully created customer in Stripe")

    // 4. Create customer in local database
    now := time.Now()
    customer := &models.Customer{
        ID:          uuid.New(),
        Email:       customerDTO.Email,
        FirstName:   customerDTO.FirstName,
        LastName:    customerDTO.LastName,
        PhoneNumber: customerDTO.PhoneNumber,
        StripeID:    stripeCustomer.ID,
        Active:      true,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
    
    s.logger.Debug().
        Str("function", "customerService.Create").
        Str("customer_id", customer.ID.String()).
        Str("stripe_id", customer.StripeID).
        Msg("Preparing to save customer to database")

    // 5. Save to database
    if err := s.customerRepo.Create(ctx, customer); err != nil {
        s.logger.Error().
            Str("function", "customerService.Create").
            Err(err).
            Str("customer_id", customer.ID.String()).
            Msg("Failed to create customer in database")
            
        // If database creation fails, we should consider cleaning up the Stripe customer
        // This is optional, as you might want to keep the Stripe customer for reconciliation purposes
        s.logger.Debug().
            Str("function", "customerService.Create").
            Str("stripe_id", stripeCustomer.ID).
            Msg("Considering cleanup of Stripe customer after database failure")
            
        // Decision: Log the inconsistency but don't delete from Stripe
        // In a real production system, you might want to handle this differently
        s.logger.Warn().
            Str("function", "customerService.Create").
            Str("stripe_id", stripeCustomer.ID).
            Str("email", customerDTO.Email).
            Msg("Customer exists in Stripe but failed to save to database")
            
        return nil, fmt.Errorf("failed to create customer in database: %w", err)
    }

    s.logger.Info().
        Str("function", "customerService.Create").
        Str("customer_id", customer.ID.String()).
        Str("stripe_id", customer.StripeID).
        Str("email", customer.Email).
        Str("name", fmt.Sprintf("%s %s", customer.FirstName, customer.LastName)).
        Msg("Customer successfully created")

    return customer, nil
}

// GetByID retrieves a customer by its ID
func (s *customerService) GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	// TODO: Implement get customer by ID
	// 1. Call repository to fetch customer
	return nil, nil
}

// GetByEmail retrieves a customer by email address
func (s *customerService) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	// TODO: Implement get customer by email
	// 1. Call repository to fetch customer by email
	return nil, nil
}

// List retrieves all customers with pagination and filtering
func (s *customerService) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Customer, int, error) {
    s.logger.Debug().
        Str("function", "customerService.List").
        Int("offset", offset).
        Int("limit", limit).
        Bool("includeInactive", includeInactive).
        Msg("Starting customer listing")

    // Call repository to list customers with the provided parameters
    s.logger.Debug().
        Str("function", "customerService.List").
        Msg("Calling repository to fetch customers")
        
    customers, total, err := s.customerRepo.List(ctx, offset, limit, includeInactive)
    if err != nil {
        s.logger.Error().
            Str("function", "customerService.List").
            Err(err).
            Int("offset", offset).
            Int("limit", limit).
            Bool("includeInactive", includeInactive).
            Msg("Failed to retrieve customers from repository")
        return nil, 0, fmt.Errorf("failed to list customers: %w", err)
    }
    
    // Log the result count
    s.logger.Debug().
        Str("function", "customerService.List").
        Int("customers_count", len(customers)).
        Int("total_count", total).
        Msg("Successfully retrieved customers from repository")
    
    // Additional processing if needed (e.g., filtering, enrichment)
    
    s.logger.Info().
        Str("function", "customerService.List").
        Int("total_customers", total).
        Int("returned_customers", len(customers)).
        Int("offset", offset).
        Int("limit", limit).
        Bool("includeInactive", includeInactive).
        Msg("Customer listing completed successfully")
        
    return customers, total, nil
}

// Update updates an existing customer
func (s *customerService) Update(ctx context.Context, id uuid.UUID, customerDTO *dto.CustomerUpdateDTO) (*models.Customer, error) {
	// TODO: Implement customer update
	// 1. Get existing customer
	// 2. Update fields from DTO
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil, nil
}

// Delete removes a customer from the system
func (s *customerService) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement customer deletion
	// 1. Get existing customer
	// 2. Archive in Stripe
	// 3. Delete from database
	// 4. Handle errors
	return nil
}

// GetAddresses retrieves all addresses for a customer
func (s *customerService) GetAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error) {
	// TODO: Implement get addresses for customer
	// 1. Call repository to get addresses
	return nil, nil
}

// AddAddress adds a new address for a customer
func (s *customerService) AddAddress(ctx context.Context, customerID uuid.UUID, addressDTO *dto.AddressCreateDTO) (*models.Address, error) {
	// TODO: Implement add address
	// 1. Validate addressDTO
	// 2. Call repository to add address
	return nil, nil
}

// UpdateAddress updates an existing address
func (s *customerService) UpdateAddress(ctx context.Context, addressID uuid.UUID, addressDTO *dto.AddressUpdateDTO) (*models.Address, error) {
	// TODO: Implement update address
	// 1. Validate addressDTO
	// 2. Call repository to update address
	return nil, nil
}

// DeleteAddress removes an address
func (s *customerService) DeleteAddress(ctx context.Context, addressID uuid.UUID) error {
	// TODO: Implement delete address
	// 1. Call repository to delete address
	return nil
}

// GetDefaultAddress gets the default shipping address for a customer
func (s *customerService) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*models.Address, error) {
	// TODO: Implement get default address
	// 1. Call repository to get default address
	return nil, nil
}

// SetDefaultAddress sets an address as the default for a customer
func (s *customerService) SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error {
	// TODO: Implement set default address
	// 1. Call repository to set default address
	return nil
}