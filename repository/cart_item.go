// repository/cart_item_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/dukerupert/walking-drum/models"
)

var (
	ErrCartItemNotFound = errors.New("cart item not found")
)

type CartItemRepository interface {
	Create(ctx context.Context, item *models.CartItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.CartItem, error)
	ListByCartID(ctx context.Context, cartID uuid.UUID) ([]*models.CartItem, error)
	Update(ctx context.Context, item *models.CartItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCartID(ctx context.Context, cartID uuid.UUID) error
}

type PostgresCartItemRepository struct {
	db *sql.DB
}

func NewCartItemRepository(db *sql.DB) CartItemRepository {
	return &PostgresCartItemRepository{
		db: db,
	}
}

func (r *PostgresCartItemRepository) Create(ctx context.Context, item *models.CartItem) error {
	// Generate a new UUID if not provided
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	query := `
		INSERT INTO cart_items (
			id, cart_id, product_id, price_id, quantity, 
			unit_price, is_subscription, created_at, 
			updated_at, options, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var priceID sql.NullString
	if item.PriceID != nil {
		priceID.String = item.PriceID.String()
		priceID.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ID,
		item.CartID,
		item.ProductID,
		priceID,
		item.Quantity,
		item.UnitPrice,
		item.IsSubscription,
		item.CreatedAt,
		item.UpdatedAt,
		item.Options,
		item.Metadata,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("error creating cart item: %w", err)
	}

	return nil
}

func (r *PostgresCartItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CartItem, error) {
	query := `
		SELECT 
			id, cart_id, product_id, price_id, quantity, 
			unit_price, is_subscription, created_at, 
			updated_at, options, metadata
		FROM cart_items
		WHERE id = $1
	`

	var item models.CartItem
	var priceID sql.NullString
	var options, metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.CartID,
		&item.ProductID,
		&priceID,
		&item.Quantity,
		&item.UnitPrice,
		&item.IsSubscription,
		&item.CreatedAt,
		&item.UpdatedAt,
		&options,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("error getting cart item by ID: %w", err)
	}

	if priceID.Valid {
		parsedID, err := uuid.Parse(priceID.String)
		if err == nil {
			item.PriceID = &parsedID
		}
	}

	// Handle JSON options and metadata conversion if needed

	return &item, nil
}

func (r *PostgresCartItemRepository) ListByCartID(ctx context.Context, cartID uuid.UUID) ([]*models.CartItem, error) {
	query := `
		SELECT 
			id, cart_id, product_id, price_id, quantity, 
			unit_price, is_subscription, created_at, 
			updated_at, options, metadata
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, fmt.Errorf("error listing cart items by cart ID: %w", err)
	}
	defer rows.Close()

	var items []*models.CartItem

	for rows.Next() {
		var item models.CartItem
		var priceID sql.NullString
		var options, metadata sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&priceID,
			&item.Quantity,
			&item.UnitPrice,
			&item.IsSubscription,
			&item.CreatedAt,
			&item.UpdatedAt,
			&options,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning cart item row: %w", err)
		}

		if priceID.Valid {
			parsedID, err := uuid.Parse(priceID.String)
			if err == nil {
				item.PriceID = &parsedID
			}
		}

		// Handle JSON options and metadata conversion if needed

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cart item rows: %w", err)
	}

	return items, nil
}

func (r *PostgresCartItemRepository) Update(ctx context.Context, item *models.CartItem) error {
	// Update the timestamp
	item.UpdatedAt = time.Now()

	query := `
		UPDATE cart_items
		SET 
			product_id = $1,
			price_id = $2,
			quantity = $3,
			unit_price = $4,
			is_subscription = $5,
			updated_at = $6,
			options = $7,
			metadata = $8
		WHERE id = $9
		RETURNING id
	`

	var priceID sql.NullString
	if item.PriceID != nil {
		priceID.String = item.PriceID.String()
		priceID.Valid = true
	}

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ProductID,
		priceID,
		item.Quantity,
		item.UnitPrice,
		item.IsSubscription,
		item.UpdatedAt,
		item.Options,
		item.Metadata,
		item.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartItemNotFound
		}
		return fmt.Errorf("error updating cart item: %w", err)
	}

	return nil
}

func (r *PostgresCartItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartItemNotFound
		}
		return fmt.Errorf("error deleting cart item: %w", err)
	}

	return nil
}

func (r *PostgresCartItemRepository) DeleteByCartID(ctx context.Context, cartID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`

	_, err := r.db.ExecContext(ctx, query, cartID)
	if err != nil {
		return fmt.Errorf("error deleting cart items by cart ID: %w", err)
	}

	return nil
}