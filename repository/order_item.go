// repository/order_item_repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/dukerupert/walking-drum/models"
)

var (
	ErrOrderItemNotFound = errors.New("order item not found")
)

type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error)
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error)
	Update(ctx context.Context, item *models.OrderItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error
}

type PostgresOrderItemRepository struct {
	db *sql.DB
}

func NewOrderItemRepository(db *sql.DB) OrderItemRepository {
	return &PostgresOrderItemRepository{
		db: db,
	}
}

func (r *PostgresOrderItemRepository) Create(ctx context.Context, item *models.OrderItem) error {
	// Generate a new UUID if not provided
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	// Set total price if not already set
	if item.TotalPrice == 0 {
		item.TotalPrice = item.UnitPrice * int64(item.Quantity)
	}

	query := `
		INSERT INTO order_items (
			id, order_id, product_id, price_id, subscription_id, 
			quantity, unit_price, total_price, is_subscription, 
			created_at, updated_at, options, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	var priceID, subscriptionID sql.NullString
	if item.PriceID != nil {
		priceID.String = item.PriceID.String()
		priceID.Valid = true
	}
	if item.SubscriptionID != nil {
		subscriptionID.String = item.SubscriptionID.String()
		subscriptionID.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ID,
		item.OrderID,
		item.ProductID,
		priceID,
		subscriptionID,
		item.Quantity,
		item.UnitPrice,
		item.TotalPrice,
		item.IsSubscription,
		item.CreatedAt,
		item.UpdatedAt,
		item.Options,
		item.Metadata,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("error creating order item: %w", err)
	}

	return nil
}

func (r *PostgresOrderItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OrderItem, error) {
	query := `
		SELECT 
			id, order_id, product_id, price_id, subscription_id, 
			quantity, unit_price, total_price, is_subscription, 
			created_at, updated_at, options, metadata
		FROM order_items
		WHERE id = $1
	`

	var item models.OrderItem
	var priceID, subscriptionID sql.NullString
	var options, metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&priceID,
		&subscriptionID,
		&item.Quantity,
		&item.UnitPrice,
		&item.TotalPrice,
		&item.IsSubscription,
		&item.CreatedAt,
		&item.UpdatedAt,
		&options,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderItemNotFound
		}
		return nil, fmt.Errorf("error getting order item by ID: %w", err)
	}

	if priceID.Valid {
		parsedID, err := uuid.Parse(priceID.String)
		if err == nil {
			item.PriceID = &parsedID
		}
	}

	if subscriptionID.Valid {
		parsedID, err := uuid.Parse(subscriptionID.String)
		if err == nil {
			item.SubscriptionID = &parsedID
		}
	}

	// Parse options if present
	if options.Valid {
		var opts map[string]interface{}
		if err := json.Unmarshal([]byte(options.String), &opts); err == nil {
			item.Options = &opts
		}
	}

	// Parse metadata if present
	if metadata.Valid {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(metadata.String), &meta); err == nil {
			item.Metadata = &meta
		}
	}

	return &item, nil
}

func (r *PostgresOrderItemRepository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*models.OrderItem, error) {
	query := `
		SELECT 
			id, order_id, product_id, price_id, subscription_id, 
			quantity, unit_price, total_price, is_subscription, 
			created_at, updated_at, options, metadata
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error listing order items by order ID: %w", err)
	}
	defer rows.Close()

	var items []*models.OrderItem

	for rows.Next() {
		var item models.OrderItem
		var priceID, subscriptionID sql.NullString
		var options, metadata sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&priceID,
			&subscriptionID,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
			&item.IsSubscription,
			&item.CreatedAt,
			&item.UpdatedAt,
			&options,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning order item row: %w", err)
		}

		if priceID.Valid {
			parsedID, err := uuid.Parse(priceID.String)
			if err == nil {
				item.PriceID = &parsedID
			}
		}

		if subscriptionID.Valid {
			parsedID, err := uuid.Parse(subscriptionID.String)
			if err == nil {
				item.SubscriptionID = &parsedID
			}
		}

		// Parse options if present
		if options.Valid {
			var opts map[string]interface{}
			if err := json.Unmarshal([]byte(options.String), &opts); err == nil {
				item.Options = &opts
			}
		}

		// Parse metadata if present
		if metadata.Valid {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metadata.String), &meta); err == nil {
				item.Metadata = &meta
			}
		}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order item rows: %w", err)
	}

	return items, nil
}

func (r *PostgresOrderItemRepository) Update(ctx context.Context, item *models.OrderItem) error {
	// Update the timestamp
	item.UpdatedAt = time.Now()

	// Update total price based on unit price and quantity
	item.TotalPrice = item.UnitPrice * int64(item.Quantity)

	query := `
		UPDATE order_items
		SET 
			product_id = $1,
			price_id = $2,
			subscription_id = $3,
			quantity = $4,
			unit_price = $5,
			total_price = $6,
			is_subscription = $7,
			updated_at = $8,
			options = $9,
			metadata = $10
		WHERE id = $11
		RETURNING id
	`

	var priceID, subscriptionID sql.NullString
	if item.PriceID != nil {
		priceID.String = item.PriceID.String()
		priceID.Valid = true
	}
	if item.SubscriptionID != nil {
		subscriptionID.String = item.SubscriptionID.String()
		subscriptionID.Valid = true
	}

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.ProductID,
		priceID,
		subscriptionID,
		item.Quantity,
		item.UnitPrice,
		item.TotalPrice,
		item.IsSubscription,
		item.UpdatedAt,
		item.Options,
		item.Metadata,
		item.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderItemNotFound
		}
		return fmt.Errorf("error updating order item: %w", err)
	}

	return nil
}

func (r *PostgresOrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderItemNotFound
		}
		return fmt.Errorf("error deleting order item: %w", err)
	}

	return nil
}

func (r *PostgresOrderItemRepository) DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE order_id = $1`

	_, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("error deleting order items by order ID: %w", err)
	}

	return nil
}