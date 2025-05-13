// internal/repositories/postgres/subscription_repository.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// variantRepository implements the VariantRepository interface
type variantRepository struct {
	db     *sql.DB
	logger *zerolog.Logger
}

// NewVariantRepository creates a new variant repository
func NewVariantRepository(db *sql.DB, logger *zerolog.Logger) VariantRepository {
	return &variantRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID retrieves a variant by ID
func (r *variantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Variant, error) {
	query := `
		SELECT id, product_id, price_id, stripe_price_id, 
			   weight, grind, active, stock_level, 
			   created_at, updated_at
		FROM variants
		WHERE id = $1
	`

	var variant models.Variant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&variant.ID,
		&variant.ProductID,
		&variant.PriceID,
		&variant.StripePriceID,
		&variant.Weight,
		&variant.Grind,
		&variant.Active,
		&variant.StockLevel,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Variant not found
		}
		return nil, fmt.Errorf("error querying variant by id: %w", err)
	}

	return &variant, nil
}

// GetByProductID retrieves variants by product ID
func (r *variantRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Variant, error) {
	query := `
		SELECT id, product_id, price_id, stripe_price_id, 
			   weight, grind, active, stock_level, 
			   created_at, updated_at
		FROM variants
		WHERE product_id = $1
		ORDER BY weight, grind
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("error querying variants by product id: %w", err)
	}
	defer rows.Close()

	variants := []*models.Variant{}
	for rows.Next() {
		var variant models.Variant
		err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.PriceID,
			&variant.StripePriceID,
			&variant.Weight,
			&variant.Grind,
			&variant.Active,
			&variant.StockLevel,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning variant row: %w", err)
		}
		variants = append(variants, &variant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating variant rows: %w", err)
	}

	return variants, nil
}

// GetByAttributes retrieves a variant by product ID, weight, and grind
func (r *variantRepository) GetByAttributes(ctx context.Context, productID uuid.UUID, weight string, grind string) (*models.Variant, error) {
	query := `
		SELECT id, product_id, price_id, stripe_price_id, 
			   weight, grind, active, stock_level, 
			   created_at, updated_at
		FROM variants
		WHERE product_id = $1 AND weight = $2 AND grind = $3
	`

	var variant models.Variant
	err := r.db.QueryRowContext(ctx, query, productID, weight, grind).Scan(
		&variant.ID,
		&variant.ProductID,
		&variant.PriceID,
		&variant.StripePriceID,
		&variant.Weight,
		&variant.Grind,
		&variant.Active,
		&variant.StockLevel,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Variant not found
		}
		return nil, fmt.Errorf("error querying variant by attributes: %w", err)
	}

	return &variant, nil
}

// List retrieves all variants with pagination
func (r *variantRepository) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*models.Variant, int, error) {
	// First, get the total count
	countQuery := `SELECT COUNT(*) FROM variants`
	if activeOnly {
		countQuery += ` WHERE active = true`
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting variants: %w", err)
	}

	// Then, get the paginated results
	query := `
		SELECT id, product_id, price_id, stripe_price_id, 
			   weight, grind, active, stock_level, 
			   created_at, updated_at
		FROM variants
	`
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY product_id, weight, grind
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying variants: %w", err)
	}
	defer rows.Close()

	variants := []*models.Variant{}
	for rows.Next() {
		var variant models.Variant
		err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.PriceID,
			&variant.StripePriceID,
			&variant.Weight,
			&variant.Grind,
			&variant.Active,
			&variant.StockLevel,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning variant row: %w", err)
		}
		variants = append(variants, &variant)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating variant rows: %w", err)
	}

	return variants, totalCount, nil
}

// Create creates a new variant
func (r *variantRepository) Create(ctx context.Context, variant *models.Variant) error {
	query := `
		INSERT INTO variants (
			id, product_id, price_id, stripe_price_id,
			weight, grind, active, stock_level,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		variant.ID,
		variant.ProductID,
		variant.PriceID,
		variant.StripePriceID,
		variant.Weight,
		variant.Grind,
		variant.Active,
		variant.StockLevel,
		variant.CreatedAt,
		variant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating variant: %w", err)
	}

	return nil
}

// Update updates an existing variant
func (r *variantRepository) Update(ctx context.Context, variant *models.Variant) error {
	query := `
		UPDATE variants
		SET product_id = $2,
			price_id = $3,
			stripe_price_id = $4,
			weight = $5,
			grind = $6,
			active = $7,
			stock_level = $8,
			updated_at = $9
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		variant.ID,
		variant.ProductID,
		variant.PriceID,
		variant.StripePriceID,
		variant.Weight,
		variant.Grind,
		variant.Active,
		variant.StockLevel,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error updating variant: %w", err)
	}

	return nil
}

// Delete deletes a variant
func (r *variantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM variants WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting variant: %w", err)
	}

	return nil
}

// UpdateStockLevel updates the stock level of a variant
func (r *variantRepository) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `
		UPDATE variants
		SET stock_level = $2,
			updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, quantity, time.Now())
	if err != nil {
		return fmt.Errorf("error updating variant stock level: %w", err)
	}

	return nil
}
