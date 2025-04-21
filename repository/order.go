// repository/order_repository.go
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
	ErrOrderNotFound = errors.New("order not found")
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error)
	ListByStatus(ctx context.Context, status models.OrderStatus, limit, offset int) ([]*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Order, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *models.Order) error {
	// Generate a new UUID if not provided
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	// Marshal addresses to JSON
	var shippingAddressJSON, billingAddressJSON []byte
	var err error

	if order.ShippingAddress != nil {
		shippingAddressJSON, err = json.Marshal(order.ShippingAddress)
		if err != nil {
			return fmt.Errorf("error marshaling shipping address: %w", err)
		}
	}

	if order.BillingAddress != nil {
		billingAddressJSON, err = json.Marshal(order.BillingAddress)
		if err != nil {
			return fmt.Errorf("error marshaling billing address: %w", err)
		}
	}

	query := `
		INSERT INTO orders (
			id, user_id, status, total_amount, currency, 
			created_at, updated_at, completed_at, shipping_address, 
			billing_address, payment_intent_id, stripe_customer_id, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	var userID sql.NullString
	if order.UserID != nil {
		userID.String = order.UserID.String()
		userID.Valid = true
	}

	var completedAt sql.NullTime
	if order.CompletedAt != nil {
		completedAt.Time = *order.CompletedAt
		completedAt.Valid = true
	}

	var paymentIntentID, stripeCustomerID sql.NullString
	if order.PaymentIntentID != nil {
		paymentIntentID.String = *order.PaymentIntentID
		paymentIntentID.Valid = true
	}
	if order.StripeCustomerID != nil {
		stripeCustomerID.String = *order.StripeCustomerID
		stripeCustomerID.Valid = true
	}

	err = r.db.QueryRowContext(
		ctx,
		query,
		order.ID,
		userID,
		order.Status,
		order.TotalAmount,
		order.Currency,
		order.CreatedAt,
		order.UpdatedAt,
		completedAt,
		shippingAddressJSON,
		billingAddressJSON,
		paymentIntentID,
		stripeCustomerID,
		order.Metadata,
	).Scan(&order.ID)

	if err != nil {
		return fmt.Errorf("error creating order: %w", err)
	}

	return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	query := `
		SELECT 
			id, user_id, status, total_amount, currency, 
			created_at, updated_at, completed_at, shipping_address, 
			billing_address, payment_intent_id, stripe_customer_id, metadata
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	var userID sql.NullString
	var completedAt sql.NullTime
	var shippingAddress, billingAddress sql.NullString
	var paymentIntentID, stripeCustomerID sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&userID,
		&order.Status,
		&order.TotalAmount,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
		&completedAt,
		&shippingAddress,
		&billingAddress,
		&paymentIntentID,
		&stripeCustomerID,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("error getting order by ID: %w", err)
	}

	// Parse UserID if present
	if userID.Valid {
		parsedID, err := uuid.Parse(userID.String)
		if err == nil {
			order.UserID = &parsedID
		}
	}

	// Parse CompletedAt if present
	if completedAt.Valid {
		order.CompletedAt = &completedAt.Time
	}

	// Parse addresses if present
	if shippingAddress.Valid {
		var addr models.Address
		if err := json.Unmarshal([]byte(shippingAddress.String), &addr); err == nil {
			order.ShippingAddress = &addr
		}
	}

	if billingAddress.Valid {
		var addr models.Address
		if err := json.Unmarshal([]byte(billingAddress.String), &addr); err == nil {
			order.BillingAddress = &addr
		}
	}

	// Parse payment IDs if present
	if paymentIntentID.Valid {
		order.PaymentIntentID = &paymentIntentID.String
	}

	if stripeCustomerID.Valid {
		order.StripeCustomerID = &stripeCustomerID.String
	}

	// Parse metadata if present
	if metadata.Valid {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(metadata.String), &meta); err == nil {
			order.Metadata = &meta
		}
	}

	return &order, nil
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, status, total_amount, currency, 
			created_at, updated_at, completed_at, shipping_address, 
			billing_address, payment_intent_id, stripe_customer_id, metadata
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing orders by user ID: %w", err)
	}
	defer rows.Close()

	return r.scanOrderRows(rows)
}

func (r *PostgresOrderRepository) ListByStatus(ctx context.Context, status models.OrderStatus, limit, offset int) ([]*models.Order, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, status, total_amount, currency, 
			created_at, updated_at, completed_at, shipping_address, 
			billing_address, payment_intent_id, stripe_customer_id, metadata
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing orders by status: %w", err)
	}
	defer rows.Close()

	return r.scanOrderRows(rows)
}

func (r *PostgresOrderRepository) Update(ctx context.Context, order *models.Order) error {
	// Update the timestamp
	order.UpdatedAt = time.Now()

	// Marshal addresses to JSON
	var shippingAddressJSON, billingAddressJSON []byte
	var err error

	if order.ShippingAddress != nil {
		shippingAddressJSON, err = json.Marshal(order.ShippingAddress)
		if err != nil {
			return fmt.Errorf("error marshaling shipping address: %w", err)
		}
	}

	if order.BillingAddress != nil {
		billingAddressJSON, err = json.Marshal(order.BillingAddress)
		if err != nil {
			return fmt.Errorf("error marshaling billing address: %w", err)
		}
	}

	query := `
		UPDATE orders
		SET 
			user_id = $1,
			status = $2,
			total_amount = $3,
			currency = $4,
			updated_at = $5,
			completed_at = $6,
			shipping_address = $7,
			billing_address = $8,
			payment_intent_id = $9,
			stripe_customer_id = $10,
			metadata = $11
		WHERE id = $12
		RETURNING id
	`

	var userID sql.NullString
	if order.UserID != nil {
		userID.String = order.UserID.String()
		userID.Valid = true
	}

	var completedAt sql.NullTime
	if order.CompletedAt != nil {
		completedAt.Time = *order.CompletedAt
		completedAt.Valid = true
	}

	var paymentIntentID, stripeCustomerID sql.NullString
	if order.PaymentIntentID != nil {
		paymentIntentID.String = *order.PaymentIntentID
		paymentIntentID.Valid = true
	}
	if order.StripeCustomerID != nil {
		stripeCustomerID.String = *order.StripeCustomerID
		stripeCustomerID.Valid = true
	}

	var returnedID uuid.UUID
	err = r.db.QueryRowContext(
		ctx,
		query,
		userID,
		order.Status,
		order.TotalAmount,
		order.Currency,
		order.UpdatedAt,
		completedAt,
		shippingAddressJSON,
		billingAddressJSON,
		paymentIntentID,
		stripeCustomerID,
		order.Metadata,
		order.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return fmt.Errorf("error updating order: %w", err)
	}

	return nil
}

func (r *PostgresOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM orders WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return fmt.Errorf("error deleting order: %w", err)
	}

	return nil
}

func (r *PostgresOrderRepository) List(ctx context.Context, limit, offset int) ([]*models.Order, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, status, total_amount, currency, 
			created_at, updated_at, completed_at, shipping_address, 
			billing_address, payment_intent_id, stripe_customer_id, metadata
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing orders: %w", err)
	}
	defer rows.Close()

	return r.scanOrderRows(rows)
}

// Helper method to scan order rows
func (r *PostgresOrderRepository) scanOrderRows(rows *sql.Rows) ([]*models.Order, error) {
	var orders []*models.Order

	for rows.Next() {
		var order models.Order
		var userID sql.NullString
		var completedAt sql.NullTime
		var shippingAddress, billingAddress sql.NullString
		var paymentIntentID, stripeCustomerID sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&order.ID,
			&userID,
			&order.Status,
			&order.TotalAmount,
			&order.Currency,
			&order.CreatedAt,
			&order.UpdatedAt,
			&completedAt,
			&shippingAddress,
			&billingAddress,
			&paymentIntentID,
			&stripeCustomerID,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning order row: %w", err)
		}

		// Parse UserID if present
		if userID.Valid {
			parsedID, err := uuid.Parse(userID.String)
			if err == nil {
				order.UserID = &parsedID
			}
		}

		// Parse CompletedAt if present
		if completedAt.Valid {
			order.CompletedAt = &completedAt.Time
		}

		// Parse addresses if present
		if shippingAddress.Valid {
			var addr models.Address
			if err := json.Unmarshal([]byte(shippingAddress.String), &addr); err == nil {
				order.ShippingAddress = &addr
			}
		}

		if billingAddress.Valid {
			var addr models.Address
			if err := json.Unmarshal([]byte(billingAddress.String), &addr); err == nil {
				order.BillingAddress = &addr
			}
		}

		// Parse payment IDs if present
		if paymentIntentID.Valid {
			order.PaymentIntentID = &paymentIntentID.String
		}

		if stripeCustomerID.Valid {
			order.StripeCustomerID = &stripeCustomerID.String
		}

		// Parse metadata if present
		if metadata.Valid {
			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metadata.String), &meta); err == nil {
				order.Metadata = &meta
			}
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	return orders, nil
}