// internal/repositories/postgres/subscription_repository.go
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

// SubscriptionRepository implements the interfaces.SubscriptionRepository interface
type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository
func NewSubscriptionRepository(db *sql.DB) interfaces.SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

// Create adds a new subscription to the database
func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	if subscription.ID == uuid.Nil {
		subscription.ID = uuid.New()
	}

	now := time.Now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	query := `
		INSERT INTO subscriptions (
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14
		)
	`

	var cancelledAt interface{}
	if subscription.CancelledAt != nil {
		cancelledAt = *subscription.CancelledAt
	} else {
		cancelledAt = nil
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		subscription.ID,
		subscription.CustomerID,
		subscription.ProductID,
		subscription.PriceID,
		subscription.AddressID,
		subscription.Quantity,
		subscription.Status,
		subscription.StripeID,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		cancelledAt,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by its ID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var subscription models.Subscription
	var cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.CustomerID,
		&subscription.ProductID,
		&subscription.PriceID,
		&subscription.AddressID,
		&subscription.Quantity,
		&subscription.Status,
		&subscription.StripeID,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&cancelledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("subscription with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if cancelledAt.Valid {
		subscription.CancelledAt = &cancelledAt.Time
	}

	return &subscription, nil
}

// GetByStripeID retrieves a subscription by its Stripe ID
func (r *SubscriptionRepository) GetByStripeID(ctx context.Context, stripeID string) (*models.Subscription, error) {
	query := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_id = $1
	`

	var subscription models.Subscription
	var cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, stripeID).Scan(
		&subscription.ID,
		&subscription.CustomerID,
		&subscription.ProductID,
		&subscription.PriceID,
		&subscription.AddressID,
		&subscription.Quantity,
		&subscription.Status,
		&subscription.StripeID,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&cancelledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("subscription with Stripe ID %s not found", stripeID)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if cancelledAt.Valid {
		subscription.CancelledAt = &cancelledAt.Time
	}

	return &subscription, nil
}

// List retrieves all subscriptions with optional pagination
func (r *SubscriptionRepository) List(ctx context.Context, offset, limit int) ([]*models.Subscription, int, error) {
	countQuery := "SELECT COUNT(*) FROM subscriptions"
	listQuery := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// If no subscriptions, return early
	if total == 0 {
		return []*models.Subscription{}, 0, nil
	}

	// Get subscriptions with pagination
	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		var subscription models.Subscription
		var cancelledAt sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.CustomerID,
			&subscription.ProductID,
			&subscription.PriceID,
			&subscription.AddressID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.StripeID,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&cancelledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if cancelledAt.Valid {
			subscription.CancelledAt = &cancelledAt.Time
		}

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during subscription rows iteration: %w", err)
	}

	return subscriptions, total, nil
}

// ListByCustomerID retrieves all subscriptions for a specific customer
func (r *SubscriptionRepository) ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error) {
	query := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer ID: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		var subscription models.Subscription
		var cancelledAt sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.CustomerID,
			&subscription.ProductID,
			&subscription.PriceID,
			&subscription.AddressID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.StripeID,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&cancelledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if cancelledAt.Valid {
			subscription.CancelledAt = &cancelledAt.Time
		}

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during subscription rows iteration: %w", err)
	}

	return subscriptions, nil
}

// ListActiveByCustomerID retrieves active subscriptions for a specific customer
func (r *SubscriptionRepository) ListActiveByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error) {
	query := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE customer_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID, models.SubscriptionStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list active subscriptions by customer ID: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		var subscription models.Subscription
		var cancelledAt sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.CustomerID,
			&subscription.ProductID,
			&subscription.PriceID,
			&subscription.AddressID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.StripeID,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&cancelledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if cancelledAt.Valid {
			subscription.CancelledAt = &cancelledAt.Time
		}

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during subscription rows iteration: %w", err)
	}

	return subscriptions, nil
}

// ListByStatus retrieves subscriptions filtered by status
func (r *SubscriptionRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.Subscription, int, error) {
	countQuery := "SELECT COUNT(*) FROM subscriptions WHERE status = $1"
	listQuery := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, status).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions by status: %w", err)
	}

	// If no subscriptions, return early
	if total == 0 {
		return []*models.Subscription{}, 0, nil
	}

	// Get subscriptions with pagination
	rows, err := r.db.QueryContext(ctx, listQuery, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list subscriptions by status: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		var subscription models.Subscription
		var cancelledAt sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.CustomerID,
			&subscription.ProductID,
			&subscription.PriceID,
			&subscription.AddressID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.StripeID,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&cancelledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if cancelledAt.Valid {
			subscription.CancelledAt = &cancelledAt.Time
		}

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during subscription rows iteration: %w", err)
	}

	return subscriptions, total, nil
}

// ListDueForRenewal retrieves subscriptions due for renewal in the given time period
func (r *SubscriptionRepository) ListDueForRenewal(ctx context.Context, before time.Time) ([]*models.Subscription, error) {
	query := `
		SELECT 
			id, customer_id, product_id, price_id, address_id,
			quantity, status, stripe_id, current_period_start, current_period_end,
			cancel_at_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE status = $1 AND current_period_end <= $2
		ORDER BY current_period_end
	`

	rows, err := r.db.QueryContext(ctx, query, models.SubscriptionStatusActive, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions due for renewal: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		var subscription models.Subscription
		var cancelledAt sql.NullTime

		err := rows.Scan(
			&subscription.ID,
			&subscription.CustomerID,
			&subscription.ProductID,
			&subscription.PriceID,
			&subscription.AddressID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.StripeID,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&subscription.CancelAtPeriodEnd,
			&cancelledAt,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if cancelledAt.Valid {
			subscription.CancelledAt = &cancelledAt.Time
		}

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during subscription rows iteration: %w", err)
	}

	return subscriptions, nil
}

// Update updates an existing subscription
func (r *SubscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	subscription.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions SET
			customer_id = $1,
			product_id = $2,
			price_id = $3,
			address_id = $4,
			quantity = $5,
			status = $6,
			stripe_id = $7,
			current_period_start = $8,
			current_period_end = $9,
			cancel_at_period_end = $10,
			cancelled_at = $11,
			updated_at = $12
		WHERE id = $13
	`

	var cancelledAt interface{}
	if subscription.CancelledAt != nil {
		cancelledAt = *subscription.CancelledAt
	} else {
		cancelledAt = nil
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		subscription.CustomerID,
		subscription.ProductID,
		subscription.PriceID,
		subscription.AddressID,
		subscription.Quantity,
		subscription.Status,
		subscription.StripeID,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		cancelledAt,
		subscription.UpdatedAt,
		subscription.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with ID %s not found", subscription.ID)
	}

	return nil
}

// UpdateStatus updates only the status of a subscription
func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE subscriptions SET
			status = $1,
			updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with ID %s not found", id)
	}

	// If cancelling, set the cancelled_at timestamp
	if status == models.SubscriptionStatusCancelled {
		cancelQuery := `
			UPDATE subscriptions SET
				cancelled_at = $1
			WHERE id = $2
		`

		_, err = r.db.ExecContext(ctx, cancelQuery, now, id)
		if err != nil {
			return fmt.Errorf("failed to update subscription cancelled_at: %w", err)
		}
	}

	return nil
}

// UpdatePeriod updates the current period information for a subscription
func (r *SubscriptionRepository) UpdatePeriod(ctx context.Context, id uuid.UUID, startDate, endDate time.Time) error {
	query := `
		UPDATE subscriptions SET
			current_period_start = $1,
			current_period_end = $2,
			updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, startDate, endDate, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update subscription period: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with ID %s not found", id)
	}

	return nil
}

// SetCancelAtPeriodEnd marks a subscription to be cancelled at the end of the period
func (r *SubscriptionRepository) SetCancelAtPeriodEnd(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error {
	query := `
		UPDATE subscriptions SET
			cancel_at_period_end = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, cancelAtPeriodEnd, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update subscription cancel_at_period_end: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with ID %s not found", id)
	}

	return nil
}

// Delete removes a subscription from the database
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM subscriptions WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with ID %s not found", id)
	}

	return nil
}

// GetWithRelatedData retrieves a subscription with its related product, price, customer and address data
func (r *SubscriptionRepository) GetWithRelatedData(ctx context.Context, id uuid.UUID) (*models.Subscription, *models.Product, *models.Price, *models.Customer, *models.Address, error) {
	query := `
		SELECT 
			s.id, s.customer_id, s.product_id, s.price_id, s.address_id,
			s.quantity, s.status, s.stripe_id, s.current_period_start, s.current_period_end,
			s.cancel_at_period_end, s.cancelled_at, s.created_at, s.updated_at,
			
			p.id, p.name, p.description, p.image_url, p.active,
			p.stock_level, p.weight, p.origin, p.roast_level, p.flavor_notes,
			p.stripe_id, p.created_at, p.updated_at,
			
			pr.id, pr.product_id, pr.name, pr.amount, pr.currency,
			pr.interval, pr.interval_count, pr.active, pr.stripe_id,
			pr.created_at, pr.updated_at,
			
			c.id, c.email, c.first_name, c.last_name, c.phone_number,
			c.stripe_id, c.active, c.created_at, c.updated_at,
			
			a.id, a.customer_id, a.line1, a.line2, a.city,
			a.state, a.postal_code, a.country, a.is_default,
			a.created_at, a.updated_at
		FROM subscriptions s
		JOIN products p ON s.product_id = p.id
		JOIN prices pr ON s.price_id = pr.id
		JOIN customers c ON s.customer_id = c.id
		JOIN addresses a ON s.address_id = a.id
		WHERE s.id = $1
	`

	var subscription models.Subscription
	var product models.Product
	var price models.Price
	var customer models.Customer
	var address models.Address
	var cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		// Subscription fields
		&subscription.ID,
		&subscription.CustomerID,
		&subscription.ProductID,
		&subscription.PriceID,
		&subscription.AddressID,
		&subscription.Quantity,
		&subscription.Status,
		&subscription.StripeID,
		&subscription.CurrentPeriodStart,
		&subscription.CurrentPeriodEnd,
		&subscription.CancelAtPeriodEnd,
		&cancelledAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,

		// Product fields
		&product.ID,
		&product.Name,
		&product.Description,
		&product.ImageURL,
		&product.Active,
		&product.StockLevel,
		&product.Weight,
		&product.Origin,
		&product.RoastLevel,
		&product.FlavorNotes,
		&product.StripeID,
		&product.CreatedAt,
		&product.UpdatedAt,

		// Price fields
		&price.ID,
		&price.ProductID,
		&price.Name,
		&price.Amount,
		&price.Currency,
		&price.Interval,
		&price.IntervalCount,
		&price.Active,
		&price.StripeID,
		&price.CreatedAt,
		&price.UpdatedAt,

		// Customer fields
		&customer.ID,
		&customer.Email,
		&customer.FirstName,
		&customer.LastName,
		&customer.PhoneNumber,
		&customer.StripeID,
		&customer.Active,
		&customer.CreatedAt,
		&customer.UpdatedAt,

		// Address fields
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
			return nil, nil, nil, nil, nil, fmt.Errorf("subscription with ID %s not found", id)
		}
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get subscription with related data: %w", err)
	}

	if cancelledAt.Valid {
		subscription.CancelledAt = &cancelledAt.Time
	}

	return &subscription, &product, &price, &customer, &address, nil
}
