// File: internal/repository/postgres/subscription_repository.go
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

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

// NewSubscriptionRepository creates a new PostgreSQL subscription repository
func NewSubscriptionRepository(db *pgxpool.Pool) repository.SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

// Create adds a new subscription to the database
func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			stripe_subscription_id, customer_id, price_id, status,
			current_period_start, current_period_end, cancel_at_period_end, canceled_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		subscription.StripeSubscriptionID,
		subscription.CustomerID,
		subscription.PriceID,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
	).Scan(&subscription.ID, &subscription.CreatedAt, &subscription.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by its ID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id int64) (*models.Subscription, error) {
	query := `
		SELECT id, stripe_subscription_id, customer_id, price_id, status,
		       current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	subscription := &models.Subscription{}
	var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("subscription not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Define midnight once
	midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	// Handle null values
	if currentPeriodStartNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodStart = &timeValue
	}
	if currentPeriodEndNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodEnd = &timeValue
	}
	if canceledAtNull.Valid {
		timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
		subscription.CanceledAt = &timeValue
	}

	return subscription, nil
}

// GetByStripeID retrieves a subscription by its Stripe ID
func (r *SubscriptionRepository) GetByStripeID(ctx context.Context, stripeSubscriptionID string) (*models.Subscription, error) {
	query := `
		SELECT id, stripe_subscription_id, customer_id, price_id, status,
		       current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	subscription := &models.Subscription{}
	var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time

	err := r.db.QueryRow(ctx, query, stripeSubscriptionID).Scan(
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("subscription not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get subscription by Stripe ID: %w", err)
	}

	// Define midnight once
	midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	// Handle null values
	if currentPeriodStartNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodStart = &timeValue
	}
	if currentPeriodEndNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodEnd = &timeValue
	}
	if canceledAtNull.Valid {
		timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
		subscription.CanceledAt = &timeValue
	}

	return subscription, nil
}

// ListByCustomerID retrieves all subscriptions for a customer
func (r *SubscriptionRepository) ListByCustomerID(ctx context.Context, customerID int64) ([]*models.Subscription, error) {
	query := `
		SELECT id, stripe_subscription_id, customer_id, price_id, status,
		       current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions for customer: %w", err)
	}
	defer rows.Close()

	subscriptions := []*models.Subscription{}
	for rows.Next() {
		subscription := &models.Subscription{}
		var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time

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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		// Define midnight once
		midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		// Handle null values
		if currentPeriodStartNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodStart = &timeValue
		}
		if currentPeriodEndNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodEnd = &timeValue
		}
		if canceledAtNull.Valid {
			timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
			subscription.CanceledAt = &timeValue
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subscriptions, nil
}

// ListByStatus retrieves all subscriptions with a given status
func (r *SubscriptionRepository) ListByStatus(ctx context.Context, status string) ([]*models.Subscription, error) {
	query := `
		SELECT id, stripe_subscription_id, customer_id, price_id, status,
		       current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by status: %w", err)
	}
	defer rows.Close()

	subscriptions := []*models.Subscription{}
	for rows.Next() {
		subscription := &models.Subscription{}
		var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time

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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		// Define midnight once
		midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		// Handle null values
		if currentPeriodStartNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodStart = &timeValue
		}
		if currentPeriodEndNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodEnd = &timeValue
		}
		if canceledAtNull.Valid {
			timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
			subscription.CanceledAt = &timeValue
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subscriptions, nil
}

// Update updates an existing subscription
func (r *SubscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET stripe_subscription_id = $1, customer_id = $2, price_id = $3, status = $4,
		    current_period_start = $5, current_period_end = $6, 
		    cancel_at_period_end = $7, canceled_at = $8, updated_at = $9
		WHERE id = $10
		RETURNING updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(
		ctx,
		query,
		subscription.StripeSubscriptionID,
		subscription.CustomerID,
		subscription.PriceID,
		subscription.Status,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
		now,
		subscription.ID,
	).Scan(&subscription.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("subscription not found: %w", err)
		}
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// GetWithDetails retrieves a subscription with product and price details
func (r *SubscriptionRepository) GetWithDetails(ctx context.Context, id int64) (*models.Subscription, error) {
	query := `
		SELECT s.id, s.stripe_subscription_id, s.customer_id, s.price_id, s.status,
		       s.current_period_start, s.current_period_end, s.cancel_at_period_end,
		       s.canceled_at, s.created_at, s.updated_at,
		       p.id, p.product_id, p.stripe_price_id, p.weight, p.grind, p.price,
		       p.is_default, p.active, p.created_at, p.updated_at,
		       pr.id, pr.stripe_product_id, pr.name, pr.description, pr.origin,
		       pr.roast_level, pr.active, pr.created_at, pr.updated_at,
		       c.id, c.stripe_customer_id, c.email, c.name, c.phone,
		       c.created_at, c.updated_at
		FROM subscriptions s
		JOIN product_prices p ON s.price_id = p.id
		JOIN products pr ON p.product_id = pr.id
		JOIN customers c ON s.customer_id = c.id
		WHERE s.id = $1
	`

	subscription := &models.Subscription{}
	price := &models.ProductPrice{}
	product := &models.Product{}
	customer := &models.Customer{}

	var currentPeriodStartNull, currentPeriodEndNull, canceledAtNull pgtype.Time
	var descriptionNull, originNull, roastLevelNull pgtype.Text

	err := r.db.QueryRow(ctx, query, id).Scan(
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
			return nil, fmt.Errorf("subscription not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get subscription details: %w", err)
	}

	// Handle null values
	// Define midnight once
	midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	// Handle null values
	if currentPeriodStartNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodStart = &timeValue
	}
	if currentPeriodEndNull.Valid {
		timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
		subscription.CurrentPeriodEnd = &timeValue
	}
	if canceledAtNull.Valid {
		timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
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
	price.Product = product
	subscription.ProductPrice = price
	subscription.Customer = customer

	return subscription, nil
}

// ListActiveWithDetails retrieves all active subscriptions with details
func (r *SubscriptionRepository) ListActiveWithDetails(ctx context.Context) ([]*models.Subscription, error) {
	query := `
		SELECT s.id, s.stripe_subscription_id, s.customer_id, s.price_id, s.status,
		       s.current_period_start, s.current_period_end, s.cancel_at_period_end,
		       s.canceled_at, s.created_at, s.updated_at,
		       p.id, p.product_id, p.stripe_price_id, p.weight, p.grind, p.price,
		       p.is_default, p.active, p.created_at, p.updated_at,
		       pr.id, pr.stripe_product_id, pr.name, pr.description, pr.origin,
		       pr.roast_level, pr.active, pr.created_at, pr.updated_at,
		       c.id, c.stripe_customer_id, c.email, c.name, c.phone,
		       c.created_at, c.updated_at
		FROM subscriptions s
		JOIN product_prices p ON s.price_id = p.id
		JOIN products pr ON p.product_id = pr.id
		JOIN customers c ON s.customer_id = c.id
		WHERE s.status = 'active'
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := []*models.Subscription{}
	for rows.Next() {
		subscription := &models.Subscription{}
		price := &models.ProductPrice{}
		product := &models.Product{}
		customer := &models.Customer{}

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

			&customer.ID,
			&customer.StripeCustomerID,
			&customer.Email,
			&customer.Name,
			&customer.Phone,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		// Define midnight once
		midnight := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		// Handle null values
		if currentPeriodStartNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodStartNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodStart = &timeValue
		}
		if currentPeriodEndNull.Valid {
			timeValue := midnight.Add(time.Duration(currentPeriodEndNull.Microseconds) * time.Microsecond)
			subscription.CurrentPeriodEnd = &timeValue
		}
		if canceledAtNull.Valid {
			timeValue := midnight.Add(time.Duration(canceledAtNull.Microseconds) * time.Microsecond)
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
		price.Product = product
		subscription.ProductPrice = price
		subscription.Customer = customer

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subscriptions, nil
}
