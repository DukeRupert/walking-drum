// internal/repositories/postgres/price_repository.go
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

// PriceRepository implements the interfaces.PriceRepository interface
type PriceRepository struct {
	db *sql.DB
}

// NewPriceRepository creates a new PriceRepository
func NewPriceRepository(db *sql.DB) interfaces.PriceRepository {
	return &PriceRepository{
		db: db,
	}
}

// Create adds a new price to the database
func (r *PriceRepository) Create(ctx context.Context, price *models.Price) error {
	if price.ID == uuid.Nil {
		price.ID = uuid.New()
	}

	now := time.Now()
	price.CreatedAt = now
	price.UpdatedAt = now

	query := `
		INSERT INTO prices (
			id, product_id, name, amount, currency, 
			interval, interval_count, active, stripe_id, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, 
			$10, $11
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		price.ID,
		price.ProductID,
		price.Name,
		price.Amount,
		price.Currency,
		price.Interval,
		price.IntervalCount,
		price.Active,
		price.StripeID,
		price.CreatedAt,
		price.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create price: %w", err)
	}

	return nil
}

// GetByID retrieves a price by its ID
func (r *PriceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error) {
	query := `
		SELECT 
			id, product_id, name, amount, currency,
			interval, interval_count, active, stripe_id,
			created_at, updated_at
		FROM prices
		WHERE id = $1
	`

	var price models.Price
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("price with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return &price, nil
}

// GetByStripeID retrieves a price by its Stripe ID
func (r *PriceRepository) GetByStripeID(ctx context.Context, stripeID string) (*models.Price, error) {
	query := `
		SELECT 
			id, product_id, name, amount, currency,
			interval, interval_count, active, stripe_id,
			created_at, updated_at
		FROM prices
		WHERE stripe_id = $1
	`

	var price models.Price
	err := r.db.QueryRowContext(ctx, query, stripeID).Scan(
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("price with Stripe ID %s not found", stripeID)
		}
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return &price, nil
}

// List retrieves all prices, with optional filtering
func (r *PriceRepository) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Price, int, error) {
	whereClause := ""
	if !includeInactive {
		whereClause = "WHERE active = true"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM prices %s", whereClause)
	listQuery := fmt.Sprintf(`
		SELECT 
			id, product_id, name, amount, currency,
			interval, interval_count, active, stripe_id,
			created_at, updated_at
		FROM prices
		%s
		ORDER BY product_id, amount
		LIMIT $1 OFFSET $2
	`, whereClause)

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count prices: %w", err)
	}

	// If no prices, return early
	if total == 0 {
		return []*models.Price{}, 0, nil
	}

	// Get prices with pagination
	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list prices: %w", err)
	}
	defer rows.Close()

	prices := make([]*models.Price, 0)
	for rows.Next() {
		var price models.Price
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during price rows iteration: %w", err)
	}

	return prices, total, nil
}

// ListByProductID retrieves all prices for a specific product
func (r *PriceRepository) ListByProductID(ctx context.Context, productID uuid.UUID, includeInactive bool) ([]*models.Price, error) {
	whereClause := "WHERE product_id = $1"
	if !includeInactive {
		whereClause += " AND active = true"
	}

	query := fmt.Sprintf(`
		SELECT 
			id, product_id, name, amount, currency,
			interval, interval_count, active, stripe_id,
			created_at, updated_at
		FROM prices
		%s
		ORDER BY amount
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list prices by product ID: %w", err)
	}
	defer rows.Close()

	prices := make([]*models.Price, 0)
	for rows.Next() {
		var price models.Price
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during price rows iteration: %w", err)
	}

	return prices, nil
}

// Update updates an existing price
func (r *PriceRepository) Update(ctx context.Context, price *models.Price) error {
	price.UpdatedAt = time.Now()

	query := `
		UPDATE prices SET
			product_id = $1,
			name = $2,
			amount = $3,
			currency = $4,
			interval = $5,
			interval_count = $6,
			active = $7,
			stripe_id = $8,
			updated_at = $9
		WHERE id = $10
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		price.ProductID,
		price.Name,
		price.Amount,
		price.Currency,
		price.Interval,
		price.IntervalCount,
		price.Active,
		price.StripeID,
		price.UpdatedAt,
		price.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("price with ID %s not found", price.ID)
	}

	return nil
}

// Delete removes a price from the database
func (r *PriceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM prices WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete price: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("price with ID %s not found", id)
	}

	return nil
}
