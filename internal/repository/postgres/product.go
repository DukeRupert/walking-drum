// File: internal/repository/postgres/product_repository.go
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

type ProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// Create adds a new product to the database
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (
			stripe_product_id, name, description, origin, roast_level, active
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.StripeProductID,
		product.Name,
		product.Description,
		product.Origin,
		product.RoastLevel,
		product.Active,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `
		SELECT id, stripe_product_id, name, description, origin, roast_level, 
		       active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &domain.Product{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.StripeProductID,
		&product.Name,
		&product.Description,
		&product.Origin,
		&product.RoastLevel,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// ListActive retrieves all active products
func (r *ProductRepository) ListActive(ctx context.Context) ([]*domain.Product, error) {
	query := `
		SELECT id, stripe_product_id, name, description, origin, roast_level, 
		       active, created_at, updated_at
		FROM products
		WHERE active = true
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active products: %w", err)
	}
	defer rows.Close()

	products := []*domain.Product{}
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID,
			&product.StripeProductID,
			&product.Name,
			&product.Description,
			&product.Origin,
			&product.RoastLevel,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET stripe_product_id = $1, name = $2, description = $3, 
		    origin = $4, roast_level = $5, active = $6, updated_at = $7
		WHERE id = $8
		RETURNING updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(
		ctx,
		query,
		product.StripeProductID,
		product.Name,
		product.Description,
		product.Origin,
		product.RoastLevel,
		product.Active,
		now,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("product not found: %w", err)
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// Delete deletes a product by its ID
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetWithPrices retrieves a product with all its prices
func (r *ProductRepository) GetWithPrices(ctx context.Context, id int64) (*domain.Product, error) {
	// First get the product
	product, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then get its prices
	pricesQuery := `
		SELECT id, product_id, stripe_price_id, weight, grind, price, 
		       is_default, active, created_at, updated_at
		FROM product_prices
		WHERE product_id = $1
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, pricesQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product prices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		price := domain.ProductPrice{}
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
		product.Prices = append(product.Prices, price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return product, nil
}

// ListActiveWithPrices retrieves all active products with their prices
func (r *ProductRepository) ListActiveWithPrices(ctx context.Context) ([]*domain.Product, error) {
	// First get all active products
	products, err := r.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	// If no products found, return empty slice
	if len(products) == 0 {
		return products, nil
	}

	// Create a map of product IDs for easier lookup
	productMap := make(map[int64]*domain.Product)
	productIDs := make([]interface{}, len(products))
	
	for i, product := range products {
		productMap[product.ID] = product
		productIDs[i] = product.ID
	}

	// Build the query with a variable number of placeholders for the IN clause
	placeholders := ""
	for i := range productIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
	}

	pricesQuery := fmt.Sprintf(`
		SELECT id, product_id, stripe_price_id, weight, grind, price, 
		       is_default, active, created_at, updated_at
		FROM product_prices
		WHERE product_id IN (%s) AND active = true
		ORDER BY product_id, price ASC
	`, placeholders)

	// Execute the query
	rows, err := r.db.Query(ctx, pricesQuery, productIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get product prices: %w", err)
	}
	defer rows.Close()

	// Process the results
	for rows.Next() {
		price := domain.ProductPrice{}
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
		
		if product, ok := productMap[price.ProductID]; ok {
			product.Prices = append(product.Prices, price)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return products, nil
}