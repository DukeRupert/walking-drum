// File: internal/repository/postgres/product_price_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/dukerupert/walking-drum/internal/models"
	"github.com/dukerupert/walking-drum/internal/repository"
)

type ProductPriceRepository struct {
	db *pgxpool.Pool
}

// NewProductPriceRepository creates a new PostgreSQL product price repository
func NewProductPriceRepository(db *pgxpool.Pool) repository.ProductPriceRepository {
	return &ProductPriceRepository{
		db: db,
	}
}

// Create adds a new product price to the database
func (r *ProductPriceRepository) Create(ctx context.Context, price *domain.ProductPrice) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If this is the default price, unset any existing default for this product
	if price.IsDefault {
		_, err = tx.Exec(ctx, `
			UPDATE product_prices
			SET is_default = false
			WHERE product_id = $1 AND is_default = true
		`, price.ProductID)
		if err != nil {
			return fmt.Errorf("failed to unset existing default prices: %w", err)
		}
	}

	// Insert the new price
	query := `
		INSERT INTO product_prices (
			product_id, stripe_price_id, weight, grind, price, is_default, active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		price.ProductID,
		price.StripePriceID,
		price.Weight,
		price.Grind,
		price.Price,
		price.IsDefault,
		price.Active,
	).Scan(&price.ID, &price.CreatedAt, &price.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product price: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a product price by its ID
func (r *ProductPriceRepository) GetByID(ctx context.Context, id int64) (*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, stripe_price_id, weight, grind, price, 
		       is_default, active, created_at, updated_at
		FROM product_prices
		WHERE id = $1
	`

	price := &domain.ProductPrice{}
	err := r.db.QueryRow(ctx, query, id).Scan(
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("price not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return price, nil
}

// ListByProductID retrieves all prices for a product
func (r *ProductPriceRepository) ListByProductID(ctx context.Context, productID int64) ([]*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, stripe_price_id, weight, grind, price, 
		       is_default, active, created_at, updated_at
		FROM product_prices
		WHERE product_id = $1
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list prices for product: %w", err)
	}
	defer rows.Close()

	prices := []*domain.ProductPrice{}
	for rows.Next() {
		price := &domain.ProductPrice{}
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price row: %w", err)
		}
		prices = append(prices, price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return prices, nil
}

// Update updates an existing product price
func (r *ProductPriceRepository) Update(ctx context.Context, price *domain.ProductPrice) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// If this is the default price, unset any existing default for this product
	if price.IsDefault {
		_, err = tx.Exec(ctx, `
			UPDATE product_prices
			SET is_default = false
			WHERE product_id = $1 AND is_default = true AND id != $2
		`, price.ProductID, price.ID)
		if err != nil {
			return fmt.Errorf("failed to unset existing default prices: %w", err)
		}
	}

	// Update the price
	query := `
		UPDATE product_prices
		SET stripe_price_id = $1, weight = $2, grind = $3, price = $4,
		    is_default = $5, active = $6, updated_at = $7
		WHERE id = $8
		RETURNING updated_at
	`

	now := time.Now()
	err = tx.QueryRow(
		ctx,
		query,
		price.StripePriceID,
		price.Weight,
		price.Grind,
		price.Price,
		price.IsDefault,
		price.Active,
		now,
		price.ID,
	).Scan(&price.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("price not found: %w", err)
		}
		return fmt.Errorf("failed to update price: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a product price by its ID
func (r *ProductPriceRepository) Delete(ctx context.Context, id int64) error {
	// Check if the price is currently used by any subscriptions
	var subscriptionCount int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM subscriptions WHERE price_id = $1
	`, id).Scan(&subscriptionCount)
	
	if err != nil {
		return fmt.Errorf("failed to check if price is used by subscriptions: %w", err)
	}
	
	if subscriptionCount > 0 {
		return fmt.Errorf("cannot delete price: it is used by %d subscriptions", subscriptionCount)
	}

	// Delete the price if not in use
	query := `
		DELETE FROM product_prices
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete price: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("price not found")
	}

	return nil
}

// GetDefaultForProduct retrieves the default price for a product
func (r *ProductPriceRepository) GetDefaultForProduct(ctx context.Context, productID int64) (*domain.ProductPrice, error) {
	query := `
		SELECT id, product_id, stripe_price_id, weight, grind, price, 
		       is_default, active, created_at, updated_at
		FROM product_prices
		WHERE product_id = $1 AND is_default = true AND active = true
		LIMIT 1
	`

	price := &domain.ProductPrice{}
	err := r.db.QueryRow(ctx, query, productID).Scan(
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no default is set, try to get the first active price
			query = `
				SELECT id, product_id, stripe_price_id, weight, grind, price, 
				       is_default, active, created_at, updated_at
				FROM product_prices
				WHERE product_id = $1 AND active = true
				ORDER BY price ASC
				LIMIT 1
			`
			
			err = r.db.QueryRow(ctx, query, productID).Scan(
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
			)
			
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, fmt.Errorf("no active prices found for product")
				}
				return nil, fmt.Errorf("failed to get first active price: %w", err)
			}
			
			return price, nil
		}
		return nil, fmt.Errorf("failed to get default price: %w", err)
	}

	return price, nil
}