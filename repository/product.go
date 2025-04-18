// repository/product.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductExists   = errors.New("product already exists with this name")
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByStripeProductID(ctx context.Context, stripeProductID string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Product, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Product, error)
}

type PostgresProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product *models.Product) error {
	// Generate a new UUID if not provided
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (
			id, name, description, is_active,
			created_at, updated_at, stripe_product_id, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Description,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
		product.StripeProductID,
		product.Metadata,
	).Scan(&product.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrProductExists
		}
		return fmt.Errorf("error creating product: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, is_active, 
			created_at, updated_at, stripe_product_id, metadata
		FROM products
		WHERE id = $1
	`

	var product models.Product
	var stripeProductID sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
		&stripeProductID,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error getting product by ID: %w", err)
	}

	if stripeProductID.Valid {
		val := stripeProductID.String
		product.StripeProductID = &val
	}

	// Handle JSON metadata conversion if needed

	return &product, nil
}

func (r *PostgresProductRepository) GetByStripeProductID(ctx context.Context, stripeProductID string) (*models.Product, error) {
	query := `
		SELECT 
			id, name, description, is_active, 
			created_at, updated_at, stripe_product_id, metadata
		FROM products
		WHERE stripe_product_id = $1
	`

	var product models.Product
	var stripeProductIDDb sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, stripeProductID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
		&stripeProductIDDb,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("error getting product by Stripe product ID: %w", err)
	}

	if stripeProductIDDb.Valid {
		val := stripeProductIDDb.String
		product.StripeProductID = &val
	}

	// Handle JSON metadata conversion if needed

	return &product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product *models.Product) error {
	// Update the timestamp
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET 
			name = $1,
			description = $2,
			is_active = $3,
			updated_at = $4,
			stripe_product_id = $5,
			metadata = $6
		WHERE id = $7
		RETURNING id
	`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.IsActive,
		product.UpdatedAt,
		product.StripeProductID,
		product.Metadata,
		product.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrProductNotFound
		}
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrProductExists
		}
		return fmt.Errorf("error updating product: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrProductNotFound
		}
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) List(ctx context.Context, limit, offset int) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, name, description, is_active, 
			created_at, updated_at, stripe_product_id, metadata
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product

	for rows.Next() {
		var product models.Product
		var stripeProductID sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&stripeProductID,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}

		if stripeProductID.Valid {
			val := stripeProductID.String
			product.StripeProductID = &val
		}

		// Handle JSON metadata conversion if needed

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}

func (r *PostgresProductRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, name, description, is_active, 
			created_at, updated_at, stripe_product_id, metadata
		FROM products
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing active products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product

	for rows.Next() {
		var product models.Product
		var stripeProductID sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&stripeProductID,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning product row: %w", err)
		}

		if stripeProductID.Valid {
			val := stripeProductID.String
			product.StripeProductID = &val
		}

		// Handle JSON metadata conversion if needed

		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, nil
}
