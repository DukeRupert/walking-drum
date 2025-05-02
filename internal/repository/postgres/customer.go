// File: internal/repository/postgres/customer_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/dukerupert/walking-drum/internal/models"
	"github.com/dukerupert/walking-drum/internal/repository"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

// NewCustomerRepository creates a new PostgreSQL customer repository
func NewCustomerRepository(db *pgxpool.Pool) repository.CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// Create adds a new customer to the database
func (r *CustomerRepository) Create(ctx context.Context, customer *models.Customer) error {
	query := `
		INSERT INTO customers (
			stripe_customer_id, email, name, phone
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		customer.StripeCustomerID,
		customer.Email,
		customer.Name,
		customer.Phone,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by their ID
func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT id, stripe_customer_id, email, name, phone, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	customer := &models.Customer{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&customer.ID,
		&customer.StripeCustomerID,
		&customer.Email,
		&customer.Name,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("customer not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetByEmail retrieves a customer by their email
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*models.Customer, error) {
	query := `
		SELECT id, stripe_customer_id, email, name, phone, created_at, updated_at
		FROM customers
		WHERE email = $1
	`

	customer := &models.Customer{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&customer.ID,
		&customer.StripeCustomerID,
		&customer.Email,
		&customer.Name,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("customer not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// GetByStripeID retrieves a customer by their Stripe ID
func (r *CustomerRepository) GetByStripeID(ctx context.Context, stripeCustomerID string) (*models.Customer, error) {
	query := `
		SELECT id, stripe_customer_id, email, name, phone, created_at, updated_at
		FROM customers
		WHERE stripe_customer_id = $1
	`

	customer := &models.Customer{}
	err := r.db.QueryRow(ctx, query, stripeCustomerID).Scan(
		&customer.ID,
		&customer.StripeCustomerID,
		&customer.Email,
		&customer.Name,
		&customer.Phone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("customer not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get customer by Stripe ID: %w", err)
	}

	return customer, nil
}

// Update updates an existing customer
func (r *CustomerRepository) Update(ctx context.Context, customer *models.Customer) error {
	query := `
		UPDATE customers
		SET stripe_customer_id = $1, email = $2, name = $3, phone = $4, updated_at = $5
		WHERE id = $6
		RETURNING updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(
		ctx,
		query,
		customer.StripeCustomerID,
		customer.Email,
		customer.Name,
		customer.Phone,
		now,
		customer.ID,
	).Scan(&customer.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("customer not found: %w", err)
		}
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// Delete deletes a customer by their ID
func (r *CustomerRepository) Delete(ctx context.Context, id int64) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First delete dependent records
	// Subscriptions will be deleted automatically by ON DELETE CASCADE

	// Delete the customer
	query := `
		DELETE FROM customers
		WHERE id = $1
	`

	commandTag, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWithAddresses retrieves a customer with all their addresses
func (r *CustomerRepository) GetWithAddresses(ctx context.Context, id int64) (*models.Customer, error) {
	// First get the customer
	customer, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then get their addresses
	addressesQuery := `
		SELECT id, customer_id, line1, line2, city, state, postal_code, country, 
		       is_default, created_at, updated_at
		FROM customer_addresses
		WHERE customer_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, addressesQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		address := models.CustomerAddress{}
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
		
		customer.Addresses = append(customer.Addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating address rows: %w", err)
	}

	return customer, nil
}

// GetWithSubscriptions retrieves a customer with all their subscriptions
func (r *CustomerRepository) GetWithSubscriptions(ctx context.Context, id int64) (*models.Customer, error) {
	// First get the customer
	customer, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then get their subscriptions
	subscriptionsQuery := `
		SELECT s.id, s.stripe_subscription_id, s.customer_id, s.price_id, s.status,
		       s.current_period_start, s.current_period_end, s.cancel_at_period_end,
		       s.canceled_at, s.created_at, s.updated_at,
		       p.id, p.product_id, p.stripe_price_id, p.weight, p.grind, p.price,
		       p.is_default, p.active, p.created_at, p.updated_at,
		       pr.id, pr.stripe_product_id, pr.name, pr.description, pr.origin,
		       pr.roast_level, pr.active, pr.created_at, pr.updated_at
		FROM subscriptions s
		JOIN product_prices p ON s.price_id = p.id
		JOIN products pr ON p.product_id = pr.id
		WHERE s.customer_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.Query(ctx, subscriptionsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer subscriptions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		subscription := models.Subscription{}
		price := models.ProductPrice{}
		product := models.Product{}
		
		var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time
		var descriptionNull, originNull, roastLevelNull pgtype.Text
		
		err := rows.Scan(
			&subscription.ID,
			&subscription.StripeSubscriptionID,
			&subscription.CustomerID,
			&subscription.PriceID,
			&subscription.Status,
			&currentPeriodStartNull,
			&currentPeriodEndNull,
			&subscription.CancelAtPeriodEnd,
			&canceledAtNull,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			
			&price.ID,
			&price.ProductID,
			&price.StripePriceID,
			&price.Weight,
			&price.Grind,
			&price.Price,
			&price.IsDefault,
			&price.Active,
			&price.CreatedAt,
			&price.UpdatedAt,
			
			&product.ID,
			&product.StripeProductID,
			&product.Name,
			&descriptionNull,
			&originNull,
			&roastLevelNull,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}
		
		// Handle null values
		if currentPeriodStartNull.Valid {
			// Create a base time at midnight
			midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			
			// Add the microseconds
			timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
			
			// Assign to your pointer
			subscription.CurrentPeriodStart = &timeValue
		}
		if currentPeriodEndNull.Valid {
			// Create a base time at midnight
			midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			
			// Add the microseconds
			timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
			
			// Assign to your pointer
			subscription.CurrentPeriodEnd = &timeValue
		}
		if canceledAtNull.Valid {
			// Create a base time at midnight
			midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			
			// Add the microseconds
			timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
			
			// Assign to your pointer
			subscription.CanceledAt = &timeValue
		}
		if descriptionNull.Valid {
			product.Description = descriptionNull.String
		}
		if originNull.Valid {
			product.Origin = originNull.String
		}
		if roastLevelNull.Valid {
			product.RoastLevel = roastLevelNull.String
		}
		
		// Set up relationships
		price.Product = &product
		subscription.ProductPrice = &price
		
		customer.Subscriptions = append(customer.Subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return customer, nil
}