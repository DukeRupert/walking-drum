// repository/invoice_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/dukerupert/walking-drum/models"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrInvoiceExists   = errors.New("invoice already exists with this stripe invoice ID")
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *models.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*models.Invoice, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Invoice, error)
	ListBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*models.Invoice, error)
	Update(ctx context.Context, invoice *models.Invoice) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Invoice, error)
	ListByStatus(ctx context.Context, status models.InvoiceStatus, limit, offset int) ([]*models.Invoice, error)
}

type PostgresInvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &PostgresInvoiceRepository{
		db: db,
	}
}

func (r *PostgresInvoiceRepository) Create(ctx context.Context, invoice *models.Invoice) error {
	// Generate a new UUID if not provided
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now

	query := `
		INSERT INTO invoices (
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		) 
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
			$11, $12, $13, $14, $15
		)
		RETURNING id
	`

	var subscriptionID sql.NullString
	if invoice.SubscriptionID != nil {
		subscriptionID.String = invoice.SubscriptionID.String()
		subscriptionID.Valid = true
	}

	var invoicePDF sql.NullString
	if invoice.InvoicePDF != nil {
		invoicePDF.String = *invoice.InvoicePDF
		invoicePDF.Valid = true
	}

	var paymentIntentID sql.NullString
	if invoice.PaymentIntentID != nil {
		paymentIntentID.String = *invoice.PaymentIntentID
		paymentIntentID.Valid = true
	}

	var periodStart, periodEnd sql.NullTime
	if invoice.PeriodStart != nil {
		periodStart.Time = *invoice.PeriodStart
		periodStart.Valid = true
	}

	if invoice.PeriodEnd != nil {
		periodEnd.Time = *invoice.PeriodEnd
		periodEnd.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		invoice.ID,
		invoice.UserID,
		subscriptionID,
		invoice.Status,
		invoice.AmountDue,
		invoice.AmountPaid,
		invoice.Currency,
		invoicePDF,
		invoice.CreatedAt,
		invoice.UpdatedAt,
		invoice.StripeInvoiceID,
		paymentIntentID,
		periodStart,
		periodEnd,
		invoice.Metadata,
	).Scan(&invoice.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrInvoiceExists
		}
		return fmt.Errorf("error creating invoice: %w", err)
	}

	return nil
}

func (r *PostgresInvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		WHERE id = $1
	`

	var invoice models.Invoice
	var subscriptionID sql.NullString
	var invoicePDF, paymentIntentID sql.NullString
	var periodStart, periodEnd sql.NullTime
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invoice.ID,
		&invoice.UserID,
		&subscriptionID,
		&invoice.Status,
		&invoice.AmountDue,
		&invoice.AmountPaid,
		&invoice.Currency,
		&invoicePDF,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
		&invoice.StripeInvoiceID,
		&paymentIntentID,
		&periodStart,
		&periodEnd,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("error getting invoice by ID: %w", err)
	}

	if subscriptionID.Valid {
		subID, err := uuid.Parse(subscriptionID.String)
		if err == nil {
			invoice.SubscriptionID = &subID
		}
	}

	if invoicePDF.Valid {
		invoice.InvoicePDF = &invoicePDF.String
	}

	if paymentIntentID.Valid {
		invoice.PaymentIntentID = &paymentIntentID.String
	}

	if periodStart.Valid {
		invoice.PeriodStart = &periodStart.Time
	}

	if periodEnd.Valid {
		invoice.PeriodEnd = &periodEnd.Time
	}

	// Handle JSON metadata conversion if needed

	return &invoice, nil
}

func (r *PostgresInvoiceRepository) GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*models.Invoice, error) {
	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		WHERE stripe_invoice_id = $1
	`

	var invoice models.Invoice
	var subscriptionID sql.NullString
	var invoicePDF, paymentIntentID sql.NullString
	var periodStart, periodEnd sql.NullTime
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, stripeInvoiceID).Scan(
		&invoice.ID,
		&invoice.UserID,
		&subscriptionID,
		&invoice.Status,
		&invoice.AmountDue,
		&invoice.AmountPaid,
		&invoice.Currency,
		&invoicePDF,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
		&invoice.StripeInvoiceID,
		&paymentIntentID,
		&periodStart,
		&periodEnd,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("error getting invoice by Stripe invoice ID: %w", err)
	}

	if subscriptionID.Valid {
		subID, err := uuid.Parse(subscriptionID.String)
		if err == nil {
			invoice.SubscriptionID = &subID
		}
	}

	if invoicePDF.Valid {
		invoice.InvoicePDF = &invoicePDF.String
	}

	if paymentIntentID.Valid {
		invoice.PaymentIntentID = &paymentIntentID.String
	}

	if periodStart.Valid {
		invoice.PeriodStart = &periodStart.Time
	}

	if periodEnd.Valid {
		invoice.PeriodEnd = &periodEnd.Time
	}

	// Handle JSON metadata conversion if needed

	return &invoice, nil
}

func (r *PostgresInvoiceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing invoices by user ID: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		var invoice models.Invoice
		var subscriptionID sql.NullString
		var invoicePDF, paymentIntentID sql.NullString
		var periodStart, periodEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID,
			&invoice.UserID,
			&subscriptionID,
			&invoice.Status,
			&invoice.AmountDue,
			&invoice.AmountPaid,
			&invoice.Currency,
			&invoicePDF,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.StripeInvoiceID,
			&paymentIntentID,
			&periodStart,
			&periodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning invoice row: %w", err)
		}

		if subscriptionID.Valid {
			subID, err := uuid.Parse(subscriptionID.String)
			if err == nil {
				invoice.SubscriptionID = &subID
			}
		}

		if invoicePDF.Valid {
			invoice.InvoicePDF = &invoicePDF.String
		}

		if paymentIntentID.Valid {
			invoice.PaymentIntentID = &paymentIntentID.String
		}

		if periodStart.Valid {
			invoice.PeriodStart = &periodStart.Time
		}

		if periodEnd.Valid {
			invoice.PeriodEnd = &periodEnd.Time
		}

		// Handle JSON metadata conversion if needed

		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, nil
}

func (r *PostgresInvoiceRepository) ListBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*models.Invoice, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, subscriptionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing invoices by subscription ID: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		var invoice models.Invoice
		var subID sql.NullString
		var invoicePDF, paymentIntentID sql.NullString
		var periodStart, periodEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID,
			&invoice.UserID,
			&subID,
			&invoice.Status,
			&invoice.AmountDue,
			&invoice.AmountPaid,
			&invoice.Currency,
			&invoicePDF,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.StripeInvoiceID,
			&paymentIntentID,
			&periodStart,
			&periodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning invoice row: %w", err)
		}

		if subID.Valid {
			id, err := uuid.Parse(subID.String)
			if err == nil {
				invoice.SubscriptionID = &id
			}
		}

		if invoicePDF.Valid {
			invoice.InvoicePDF = &invoicePDF.String
		}

		if paymentIntentID.Valid {
			invoice.PaymentIntentID = &paymentIntentID.String
		}

		if periodStart.Valid {
			invoice.PeriodStart = &periodStart.Time
		}

		if periodEnd.Valid {
			invoice.PeriodEnd = &periodEnd.Time
		}

		// Handle JSON metadata conversion if needed

		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, nil
}

func (r *PostgresInvoiceRepository) Update(ctx context.Context, invoice *models.Invoice) error {
	// Update the timestamp
	invoice.UpdatedAt = time.Now()

	query := `
		UPDATE invoices
		SET 
			user_id = $1,
			subscription_id = $2,
			status = $3,
			amount_due = $4,
			amount_paid = $5,
			currency = $6,
			invoice_pdf = $7,
			updated_at = $8,
			stripe_invoice_id = $9,
			payment_intent_id = $10,
			period_start = $11,
			period_end = $12,
			metadata = $13
		WHERE id = $14
		RETURNING id
	`

	var subscriptionID sql.NullString
	if invoice.SubscriptionID != nil {
		subscriptionID.String = invoice.SubscriptionID.String()
		subscriptionID.Valid = true
	}

	var invoicePDF sql.NullString
	if invoice.InvoicePDF != nil {
		invoicePDF.String = *invoice.InvoicePDF
		invoicePDF.Valid = true
	}

	var paymentIntentID sql.NullString
	if invoice.PaymentIntentID != nil {
		paymentIntentID.String = *invoice.PaymentIntentID
		paymentIntentID.Valid = true
	}

	var periodStart, periodEnd sql.NullTime
	if invoice.PeriodStart != nil {
		periodStart.Time = *invoice.PeriodStart
		periodStart.Valid = true
	}

	if invoice.PeriodEnd != nil {
		periodEnd.Time = *invoice.PeriodEnd
		periodEnd.Valid = true
	}

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		invoice.UserID,
		subscriptionID,
		invoice.Status,
		invoice.AmountDue,
		invoice.AmountPaid,
		invoice.Currency,
		invoicePDF,
		invoice.UpdatedAt,
		invoice.StripeInvoiceID,
		paymentIntentID,
		periodStart,
		periodEnd,
		invoice.Metadata,
		invoice.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvoiceNotFound
		}
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrInvoiceExists
		}
		return fmt.Errorf("error updating invoice: %w", err)
	}

	return nil
}

func (r *PostgresInvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invoices WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvoiceNotFound
		}
		return fmt.Errorf("error deleting invoice: %w", err)
	}

	return nil
}

func (r *PostgresInvoiceRepository) List(ctx context.Context, limit, offset int) ([]*models.Invoice, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		var invoice models.Invoice
		var subscriptionID sql.NullString
		var invoicePDF, paymentIntentID sql.NullString
		var periodStart, periodEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID,
			&invoice.UserID,
			&subscriptionID,
			&invoice.Status,
			&invoice.AmountDue,
			&invoice.AmountPaid,
			&invoice.Currency,
			&invoicePDF,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.StripeInvoiceID,
			&paymentIntentID,
			&periodStart,
			&periodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning invoice row: %w", err)
		}

		if subscriptionID.Valid {
			subID, err := uuid.Parse(subscriptionID.String)
			if err == nil {
				invoice.SubscriptionID = &subID
			}
		}

		if invoicePDF.Valid {
			invoice.InvoicePDF = &invoicePDF.String
		}

		if paymentIntentID.Valid {
			invoice.PaymentIntentID = &paymentIntentID.String
		}

		if periodStart.Valid {
			invoice.PeriodStart = &periodStart.Time
		}

		if periodEnd.Valid {
			invoice.PeriodEnd = &periodEnd.Time
		}

		// Handle JSON metadata conversion if needed

		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, nil
}

func (r *PostgresInvoiceRepository) ListByStatus(ctx context.Context, status models.InvoiceStatus, limit, offset int) ([]*models.Invoice, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, subscription_id, status, 
			amount_due, amount_paid, currency, invoice_pdf, 
			created_at, updated_at, stripe_invoice_id, 
			payment_intent_id, period_start, period_end, metadata
		FROM invoices
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing invoices by status: %w", err)
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		var invoice models.Invoice
		var subscriptionID sql.NullString
		var invoicePDF, paymentIntentID sql.NullString
		var periodStart, periodEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID,
			&invoice.UserID,
			&subscriptionID,
			&invoice.Status,
			&invoice.AmountDue,
			&invoice.AmountPaid,
			&invoice.Currency,
			&invoicePDF,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&invoice.StripeInvoiceID,
			&paymentIntentID,
			&periodStart,
			&periodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning invoice row: %w", err)
		}

		if subscriptionID.Valid {
			subID, err := uuid.Parse(subscriptionID.String)
			if err == nil {
				invoice.SubscriptionID = &subID
			}
		}

		if invoicePDF.Valid {
			invoice.InvoicePDF = &invoicePDF.String
		}

		if paymentIntentID.Valid {
			invoice.PaymentIntentID = &paymentIntentID.String
		}

		if periodStart.Valid {
			invoice.PeriodStart = &periodStart.Time
		}

		if periodEnd.Valid {
			invoice.PeriodEnd = &periodEnd.Time
		}

		// Handle JSON metadata conversion if needed

		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, nil
}