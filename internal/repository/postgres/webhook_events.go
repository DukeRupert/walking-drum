// File: internal/repository/postgres/webhook_event_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/dukerupert/walking-drum/internal/models"
	"github.com/dukerupert/walking-drum/internal/repository"
)

type WebhookEventRepository struct {
	db *pgxpool.Pool
}

// NewWebhookEventRepository creates a new PostgreSQL webhook event repository
func NewWebhookEventRepository(db *pgxpool.Pool) repository.WebhookEventRepository {
	return &WebhookEventRepository{
		db: db,
	}
}

// Create adds a new webhook event to the database
func (r *WebhookEventRepository) Create(ctx context.Context, event *models.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (
			stripe_event_id, type, object, processed
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		event.StripeEventID,
		event.Type,
		event.Object,
		event.Processed,
	).Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create webhook event: %w", err)
	}

	return nil
}

// GetByStripeID retrieves a webhook event by its Stripe ID
func (r *WebhookEventRepository) GetByStripeID(ctx context.Context, stripeEventID string) (*models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, object, processed, created_at
		FROM webhook_events
		WHERE stripe_event_id = $1
	`

	event := &models.WebhookEvent{}
	err := r.db.QueryRow(ctx, query, stripeEventID).Scan(
		&event.ID,
		&event.StripeEventID,
		&event.Type,
		&event.Object,
		&event.Processed,
		&event.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("webhook event not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get webhook event: %w", err)
	}

	return event, nil
}

// MarkAsProcessed marks a webhook event as processed
func (r *WebhookEventRepository) MarkAsProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE webhook_events
		SET processed = true
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark webhook event as processed: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("webhook event not found")
	}

	return nil
}

// ListUnprocessed retrieves all unprocessed webhook events
func (r *WebhookEventRepository) ListUnprocessed(ctx context.Context) ([]*models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, object, processed, created_at
		FROM webhook_events
		WHERE processed = false
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list unprocessed webhook events: %w", err)
	}
	defer rows.Close()

	events := []*models.WebhookEvent{}
	for rows.Next() {
		event := &models.WebhookEvent{}
		err := rows.Scan(
			&event.ID,
			&event.StripeEventID,
			&event.Type,
			&event.Object,
			&event.Processed,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event row: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook event rows: %w", err)
	}

	return events, nil
}
