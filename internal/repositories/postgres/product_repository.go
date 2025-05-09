// internal/repositories/postgres/product_repository.go
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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ProductRepository implements the interfaces.ProductRepository interface
type ProductRepository struct {
	db *DB
	logger zerolog.Logger
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *DB, logger zerolog.Logger) interfaces.ProductRepository {
	return &ProductRepository{
		db: db,
		logger: logger,
	}
}

// Create adds a new product to the database
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (
			id, name, description, image_url, active, stock_level, 
			weight, origin, roast_level, flavor_notes, stripe_id, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11, 
			$12, $13
		)
	`

	// Access the underlying *sql.DB through our custom DB type
	_, err := r.db.ExecContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.ImageURL,
		product.Active,
		product.StockLevel,
		product.Weight,
		product.Origin,
		product.RoastLevel,
		product.FlavorNotes,
		product.StripeID,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, image_url, active, stock_level,
			weight, origin, roast_level, flavor_notes, stripe_id,
			created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetByStripeID retrieves a product by its Stripe ID
func (r *ProductRepository) GetByStripeID(ctx context.Context, stripeID string) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, image_url, active, stock_level,
			weight, origin, roast_level, flavor_notes, stripe_id,
			created_at, updated_at
		FROM products
		WHERE stripe_id = $1
	`

	var product models.Product
	err := r.db.QueryRowContext(ctx, query, stripeID).Scan(
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product with Stripe ID %s not found", stripeID)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves all products, with optional filtering
func (r *ProductRepository) List(ctx context.Context, offset, limit int, includeInactive bool) ([]*models.Product, int, error) {
	r.logger.Info().Msg("Executing List()")

	whereClause := ""
	if !includeInactive {
		whereClause = "WHERE active = true"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	listQuery := fmt.Sprintf(`
		SELECT 
			id, name, description, image_url, active, stock_level,
			weight, origin, roast_level, flavor_notes, stripe_id,
			created_at, updated_at
		FROM products
		%s
		ORDER BY name
		LIMIT $1 OFFSET $2
	`, whereClause)

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	log.Debug().Int("total_count", total).Msg("Total record count")

	// If no products, return early
	if total == 0 {
		log.Debug().Msg("Total equals 0. Returning early")
		return []*models.Product{}, 0, nil
	}

	// Get products with pagination
	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := make([]*models.Product, 0)
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during product rows iteration: %w", err)
	}

	return products, total, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products SET
			name = $1,
			description = $2,
			image_url = $3,
			active = $4,
			stock_level = $5,
			weight = $6,
			origin = $7,
			roast_level = $8,
			flavor_notes = $9,
			stripe_id = $10,
			updated_at = $11
		WHERE id = $12
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.ImageURL,
		product.Active,
		product.StockLevel,
		product.Weight,
		product.Origin,
		product.RoastLevel,
		product.FlavorNotes,
		product.StripeID,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID %s not found", product.ID)
	}

	return nil
}

// Delete removes a product from the database
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM products WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID %s not found", id)
	}

	return nil
}

// UpdateStockLevel updates the stock level of a product
func (r *ProductRepository) UpdateStockLevel(ctx context.Context, id uuid.UUID, quantity int) error {
	query := `
		UPDATE products SET
			stock_level = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		quantity,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update product stock level: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID %s not found", id)
	}

	return nil
}
