// File: internal/repository/postgres/customer_address_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/models"
	"github.com/dukerupert/walking-drum/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerAddressRepository struct {
	db *pgxpool.Pool
}

// NewCustomerAddressRepository creates a new PostgreSQL customer address repository
func NewCustomerAddressRepository(db *pgxpool.Pool) repository.CustomerAddressRepository {
	return &CustomerAddressRepository{
		db: db,
	}
}

// Create adds a new customer address to the database
func (r *CustomerAddressRepository) Create(ctx context.Context, address *domain.CustomerAddress) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If this is the default address, unset any existing default for this customer
	if address.IsDefault {
		_, err = tx.Exec(ctx, `
			UPDATE customer_addresses
			SET is_default = false
			WHERE customer_id = $1 AND is_default = true
		`, address.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to unset existing default addresses: %w", err)
		}
	}

	// Insert the new address
	query := `
		INSERT INTO customer_addresses (
			customer_id, line1, line2, city, state, postal_code, country, is_default
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		address.CustomerID,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
	).Scan(&address.ID, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer address: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a customer address by its ID
func (r *CustomerAddressRepository) GetByID(ctx context.Context, id int64) (*domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, line1, line2, city, state, postal_code, country, 
		       is_default, created_at, updated_at
		FROM customer_addresses
		WHERE id = $1
	`

	address := &domain.CustomerAddress{}
	var line2Null pgtype.Text
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&address.ID,
		&address.CustomerID,
		&address.Line1,
		&line2Null,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("address not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	
	if line2Null.Valid {
		address.Line2 = line2Null.String
	}

	return address, nil
}

// ListByCustomerID retrieves all addresses for a customer
func (r *CustomerAddressRepository) ListByCustomerID(ctx context.Context, customerID int64) ([]*domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, line1, line2, city, state, postal_code, country, 
		       is_default, created_at, updated_at
		FROM customer_addresses
		WHERE customer_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses for customer: %w", err)
	}
	defer rows.Close()

	addresses := []*domain.CustomerAddress{}
	for rows.Next() {
		address := &domain.CustomerAddress{}
		var line2Null pgtype.Text
		
		err := rows.Scan(
			&address.ID,
			&address.CustomerID,
			&address.Line1,
			&line2Null,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.IsDefault,
			&address.CreatedAt,
			&address.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address row: %w", err)
		}
		
		if line2Null.Valid {
			address.Line2 = line2Null.String
		}
		
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating address rows: %w", err)
	}

	return addresses, nil
}

// Update updates an existing customer address
func (r *CustomerAddressRepository) Update(ctx context.Context, address *domain.CustomerAddress) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If this is the default address, unset any existing default for this customer
	if address.IsDefault {
		_, err = tx.Exec(ctx, `
			UPDATE customer_addresses
			SET is_default = false
			WHERE customer_id = $1 AND is_default = true AND id != $2
		`, address.CustomerID, address.ID)
		if err != nil {
			return fmt.Errorf("failed to unset existing default addresses: %w", err)
		}
	}

	// Update the address
	query := `
		UPDATE customer_addresses
		SET line1 = $1, line2 = $2, city = $3, state = $4, postal_code = $5,
		    country = $6, is_default = $7, updated_at = $8
		WHERE id = $9 AND customer_id = $10
		RETURNING updated_at
	`

	now := time.Now()
	err = tx.QueryRow(
		ctx,
		query,
		address.Line1,
		address.Line2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
		now,
		address.ID,
		address.CustomerID,
	).Scan(&address.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("address not found or does not belong to customer: %w", err)
		}
		return fmt.Errorf("failed to update address: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// File: internal/repository/postgres/customer_address_repository.go (continued)

// Delete deletes a customer address by its ID
func (r *CustomerAddressRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM customer_addresses
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("address not found")
	}

	return nil
}

// GetDefaultForCustomer retrieves the default address for a customer
func (r *CustomerAddressRepository) GetDefaultForCustomer(ctx context.Context, customerID int64) (*domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, line1, line2, city, state, postal_code, country, 
		       is_default, created_at, updated_at
		FROM customer_addresses
		WHERE customer_id = $1 AND is_default = true
		LIMIT 1
	`

	address := &domain.CustomerAddress{}
	var line2Null pgtype.Text
	
	err := r.db.QueryRow(ctx, query, customerID).Scan(
		&address.ID,
		&address.CustomerID,
		&address.Line1,
		&line2Null,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no default is set, try to get the most recent address
			query = `
				SELECT id, customer_id, line1, line2, city, state, postal_code, country, 
				       is_default, created_at, updated_at
				FROM customer_addresses
				WHERE customer_id = $1
				ORDER BY created_at DESC
				LIMIT 1
			`
			
			err = r.db.QueryRow(ctx, query, customerID).Scan(
				&address.ID,
				&address.CustomerID,
				&address.Line1,
				&line2Null,
				&address.City,
				&address.State,
				&address.PostalCode,
				&address.Country,
				&address.IsDefault,
				&address.CreatedAt,
				&address.UpdatedAt,
			)
			
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, fmt.Errorf("no addresses found for customer")
				}
				return nil, fmt.Errorf("failed to get most recent address: %w", err)
			}
			
			return address, nil
		}
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}
	
	if line2Null.Valid {
		address.Line2 = line2Null.String
	}

	return address, nil
}

// SetAsDefault sets an address as the default for a customer
func (r *CustomerAddressRepository) SetAsDefault(ctx context.Context, id int64, customerID int64) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First, verify that the address exists and belongs to the customer
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM customer_addresses 
			WHERE id = $1 AND customer_id = $2
		)
	`, id, customerID).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("failed to verify address: %w", err)
	}
	
	if !exists {
		return fmt.Errorf("address not found or does not belong to customer")
	}

	// Unset any existing default addresses for this customer
	_, err = tx.Exec(ctx, `
		UPDATE customer_addresses
		SET is_default = false, updated_at = $1
		WHERE customer_id = $2 AND is_default = true
	`, time.Now(), customerID)
	if err != nil {
		return fmt.Errorf("failed to unset existing default addresses: %w", err)
	}

	// Set the new default address
	_, err = tx.Exec(ctx, `
		UPDATE customer_addresses
		SET is_default = true, updated_at = $1
		WHERE id = $2
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set address as default: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}