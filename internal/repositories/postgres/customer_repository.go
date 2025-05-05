// internal/repositories/postgres/customer_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	"github.com/google/uuid"
)

// CustomerRepository implements the interfaces.CustomerRepository interface
type CustomerRepository struct {
	db *DB
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(db *DB) interfaces.CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// Create adds a new customer to the database
func (r *CustomerRepository) Create(ctx context.Context, customer *models.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}

	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	query := `
		INSERT INTO customers (
			id, email, first_name, last_name, phone_number,
			stripe_id, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		customer.ID,
		customer.Email,
		customer.FirstName,
		customer.LastName,
		customer.PhoneNumber,
		customer.StripeID,
		customer.Active,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by its ID
func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT 
			id, email, first_name, last_name, phone_number,
			stripe_id, active, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	var customer models.Customer
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.Email,
		&customer.FirstName,
		&customer.LastName,
		&customer.PhoneNumber,
		&customer.StripeID,
		&customer.Active,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// GetByEmail retrieves a customer by email address
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	query := `
		SELECT 
			id, email, first_name, last_name, phone_number,
			stripe_id, active, created_at, updated_at
		FROM customers
		WHERE email = $1
	`

	var customer models.Customer
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&customer.ID,
		&customer.Email,
		&customer.FirstName,
		&customer.LastName,
		&customer.PhoneNumber,
		&customer.StripeID,
		&customer.Active,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return &customer, nil
}

// GetByStripeID retrieves a customer by its Stripe ID
func (r *CustomerRepository) GetByStripeID(ctx context.Context, stripeID string) (*models.Customer, error) {
	query := `
		SELECT 
			id, email, first_name, last_name, phone_number,
			stripe_id, active, created_at, updated_at
		FROM customers
		WHERE stripe_id = $1
	`

	var customer models.Customer
	err := r.db.QueryRowContext(ctx, query, stripeID).Scan(
		&customer.ID,
		&customer.Email,
		&customer.FirstName,
		&customer.LastName,
		&customer.PhoneNumber,
		&customer.StripeID,
		&customer.Active,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer with Stripe ID %s not found", stripeID)
		}
		return nil, fmt.Errorf("failed to get customer by Stripe ID: %w", err)
	}

	return &customer, nil
}

// List retrieves customers with optional pagination and filtering
func (r *CustomerRepository) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Customer, int, error) {
	whereClause := ""
	if !includeInactive {
		whereClause = "WHERE active = true"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers %s", whereClause)
	listQuery := fmt.Sprintf(`
		SELECT 
			id, email, first_name, last_name, phone_number,
			stripe_id, active, created_at, updated_at
		FROM customers
		%s
		ORDER BY last_name, first_name
		LIMIT $1 OFFSET $2
	`, whereClause)

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// If no customers, return early
	if total == 0 {
		return []*models.Customer{}, 0, nil
	}

	// Get customers with pagination
	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*models.Customer, 0)
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.Email,
			&customer.FirstName,
			&customer.LastName,
			&customer.PhoneNumber,
			&customer.StripeID,
			&customer.Active,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during customer rows iteration: %w", err)
	}

	return customers, total, nil
}

// Update updates an existing customer
func (r *CustomerRepository) Update(ctx context.Context, customer *models.Customer) error {
	customer.UpdatedAt = time.Now()

	query := `
		UPDATE customers SET
			email = $1,
			first_name = $2,
			last_name = $3,
			phone_number = $4,
			stripe_id = $5,
			active = $6,
			updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		customer.Email,
		customer.FirstName,
		customer.LastName,
		customer.PhoneNumber,
		customer.StripeID,
		customer.Active,
		customer.UpdatedAt,
		customer.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer with ID %s not found", customer.ID)
	}

	return nil
}

// Delete removes a customer from the database
func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM customers WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer with ID %s not found", id)
	}

	return nil
}

// GetAddresses retrieves all addresses for a customer
func (r *CustomerRepository) GetAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error) {
	query := `
		SELECT 
			id, customer_id, line1, line2, city,
			state, postal_code, country, is_default,
			created_at, updated_at
		FROM addresses
		WHERE customer_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]*models.Address, 0)
	for rows.Next() {
		var address models.Address
		err := rows.Scan(
			&address.ID,
			&address.CustomerID,
			&address.Line1,
			&address.Line2,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.IsDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		addresses = append(addresses, &address)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during address rows iteration: %w", err)
	}

	return addresses, nil
}

// AddAddress adds a new address for a customer
func (r *CustomerRepository) AddAddress(ctx context.Context, address *models.Address) error {
	if address.ID == uuid.Nil {
		address.ID = uuid.New()
	}

	now := time.Now()
	address.CreatedAt = now
	address.UpdatedAt = now

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// If this is the default address, clear any existing default
	if address.IsDefault {
		_, err := tx.ExecContext(
			ctx,
			"UPDATE addresses SET is_default = false WHERE customer_id = $1",
			address.CustomerID,
		)
		if err != nil {
			return fmt.Errorf("failed to clear existing default address: %w", err)
		}
	}

	// Insert the new address
	query := `
		INSERT INTO addresses (
			id, customer_id, line1, line2, city,
			state, postal_code, country, is_default,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11
		)
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		address.ID,
		address.CustomerID,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
		address.CreatedAt,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add address: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateAddress updates an existing address
func (r *CustomerRepository) UpdateAddress(ctx context.Context, address *models.Address) error {
	address.UpdatedAt = time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// If this is being set as the default address, clear any existing default
	if address.IsDefault {
		_, err := tx.ExecContext(
			ctx,
			"UPDATE addresses SET is_default = false WHERE customer_id = $1",
			address.CustomerID,
		)
		if err != nil {
			return fmt.Errorf("failed to clear existing default address: %w", err)
		}
	}

	// Update the address
	query := `
		UPDATE addresses SET
			line1 = $1,
			line2 = $2,
			city = $3,
			state = $4,
			postal_code = $5,
			country = $6,
			is_default = $7,
			updated_at = $8
		WHERE id = $9
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
		address.UpdatedAt,
		address.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address with ID %s not found", address.ID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteAddress removes an address
func (r *CustomerRepository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM addresses WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address with ID %s not found", id)
	}

	return nil
}

// GetDefaultAddress gets the default shipping address for a customer
func (r *CustomerRepository) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*models.Address, error) {
	query := `
		SELECT 
			id, customer_id, line1, line2, city,
			state, postal_code, country, is_default,
			created_at, updated_at
		FROM addresses
		WHERE customer_id = $1 AND is_default = true
	`

	var address models.Address
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&address.ID,
		&address.CustomerID,
		&address.Line1,
		&address.Line2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no default address found for customer %s", customerID)
		}
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}

	return &address, nil
}

// SetDefaultAddress sets an address as the default for a customer
func (r *CustomerRepository) SetDefaultAddress(ctx context.Context, customerID, addressID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear any existing default
	_, err = tx.ExecContext(
		ctx,
		"UPDATE addresses SET is_default = false WHERE customer_id = $1",
		customerID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear existing default address: %w", err)
	}

	// Set the new default
	query := `
		UPDATE addresses SET
			is_default = true,
			updated_at = $1
		WHERE id = $2 AND customer_id = $3
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		time.Now(),
		addressID,
		customerID,
	)

	if err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address with ID %s not found for customer %s", addressID, customerID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
