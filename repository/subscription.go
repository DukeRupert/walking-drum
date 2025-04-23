// repository/subscription_repository.go
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
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSubscriptionExists   = errors.New("subscription already exists with this stripe subscription ID")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*models.Subscription, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*models.Subscription, error)
    GetByCustomerID(ctx context.Context, customerID string, status string, limit, offset int) ([]*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Subscription, error)
	ListByStatus(ctx context.Context, status models.SubscriptionStatus, limit, offset int) ([]*models.Subscription, error)
}

type PostgresSubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &PostgresSubscriptionRepository{
		db: db,
	}
}

func (r *PostgresSubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
    // Generate a new UUID if not provided
    if subscription.ID == uuid.Nil {
        subscription.ID = uuid.New()
    }

    // Set timestamps
    now := time.Now()
    subscription.CreatedAt = now
    subscription.UpdatedAt = now

    query := `
        INSERT INTO subscriptions (
            id, user_id, price_id, quantity, status, 
            current_period_start, current_period_end, cancel_at, 
            canceled_at, ended_at, trial_start, trial_end, 
            created_at, updated_at, stripe_subscription_id, 
            stripe_customer_id, collection_method, 
            cancel_at_period_end, metadata, resume_at
        ) 
        VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 
            $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
        )
        RETURNING id
    `

    var cancelAt sql.NullTime
    if subscription.CancelAt != nil {
        cancelAt.Time = *subscription.CancelAt
        cancelAt.Valid = true
    }

    var canceledAt sql.NullTime
    if subscription.CanceledAt != nil {
        canceledAt.Time = *subscription.CanceledAt
        canceledAt.Valid = true
    }

    var endedAt sql.NullTime
    if subscription.EndedAt != nil {
        endedAt.Time = *subscription.EndedAt
        endedAt.Valid = true
    }

    var trialStart sql.NullTime
    if subscription.TrialStart != nil {
        trialStart.Time = *subscription.TrialStart
        trialStart.Valid = true
    }

    var trialEnd sql.NullTime
    if subscription.TrialEnd != nil {
        trialEnd.Time = *subscription.TrialEnd
        trialEnd.Valid = true
    }

    var resumeAt sql.NullTime
    if subscription.ResumeAt != nil {
        resumeAt.Time = *subscription.ResumeAt
        resumeAt.Valid = true
    }

    err := r.db.QueryRowContext(
        ctx,
        query,
        subscription.ID,
        subscription.UserID,
        subscription.PriceID,
        subscription.Quantity,
        subscription.Status,
        subscription.CurrentPeriodStart,
        subscription.CurrentPeriodEnd,
        cancelAt,
        canceledAt,
        endedAt,
        trialStart,
        trialEnd,
        subscription.CreatedAt,
        subscription.UpdatedAt,
        subscription.StripeSubscriptionID,
        subscription.StripeCustomerID,
        subscription.CollectionMethod,
        subscription.CancelAtPeriodEnd,
        subscription.Metadata,
        resumeAt,
    ).Scan(&subscription.ID)

    if err != nil {
        // Check for unique violation
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
            return ErrSubscriptionExists
        }
        return fmt.Errorf("error creating subscription: %w", err)
    }

    return nil
}

func (r *PostgresSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
    query := `
        SELECT 
            id, user_id, price_id, quantity, status, 
            current_period_start, current_period_end, cancel_at, 
            canceled_at, ended_at, trial_start, trial_end, 
            created_at, updated_at, stripe_subscription_id, 
            stripe_customer_id, collection_method, 
            cancel_at_period_end, metadata, resume_at
        FROM subscriptions
        WHERE id = $1
    `

    var subscription models.Subscription
    var cancelAt, canceledAt, endedAt, trialStart, trialEnd, resumeAt sql.NullTime
    var metadata sql.NullString

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &subscription.ID,
        &subscription.UserID,
        &subscription.PriceID,
        &subscription.Quantity,
        &subscription.Status,
        &subscription.CurrentPeriodStart,
        &subscription.CurrentPeriodEnd,
        &cancelAt,
        &canceledAt,
        &endedAt,
        &trialStart,
        &trialEnd,
        &subscription.CreatedAt,
        &subscription.UpdatedAt,
        &subscription.StripeSubscriptionID,
        &subscription.StripeCustomerID,
        &subscription.CollectionMethod,
        &subscription.CancelAtPeriodEnd,
        &metadata,
        &resumeAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrSubscriptionNotFound
        }
        return nil, fmt.Errorf("error getting subscription by ID: %w", err)
    }

    if cancelAt.Valid {
        t := cancelAt.Time
        subscription.CancelAt = &t
    }

    if canceledAt.Valid {
        t := canceledAt.Time
        subscription.CanceledAt = &t
    }

    if endedAt.Valid {
        t := endedAt.Time
        subscription.EndedAt = &t
    }

    if trialStart.Valid {
        t := trialStart.Time
        subscription.TrialStart = &t
    }

    if trialEnd.Valid {
        t := trialEnd.Time
        subscription.TrialEnd = &t
    }

    if resumeAt.Valid {
        t := resumeAt.Time
        subscription.ResumeAt = &t
    }

    // Handle JSON metadata conversion if needed

    return &subscription, nil
}

func (r *PostgresSubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*models.Subscription, error) {
    query := `
        SELECT 
            id, user_id, price_id, quantity, status, 
            current_period_start, current_period_end, cancel_at, 
            canceled_at, ended_at, trial_start, trial_end, 
            created_at, updated_at, stripe_subscription_id, 
            stripe_customer_id, collection_method, 
            cancel_at_period_end, metadata, resume_at
        FROM subscriptions
        WHERE stripe_subscription_id = $1
    `

    var subscription models.Subscription
    var cancelAt, canceledAt, endedAt, trialStart, trialEnd, resumeAt sql.NullTime
    var metadata sql.NullString

    err := r.db.QueryRowContext(ctx, query, stripeSubscriptionID).Scan(
        &subscription.ID,
        &subscription.UserID,
        &subscription.PriceID,
        &subscription.Quantity,
        &subscription.Status,
        &subscription.CurrentPeriodStart,
        &subscription.CurrentPeriodEnd,
        &cancelAt,
        &canceledAt,
        &endedAt,
        &trialStart,
        &trialEnd,
        &subscription.CreatedAt,
        &subscription.UpdatedAt,
        &subscription.StripeSubscriptionID,
        &subscription.StripeCustomerID,
        &subscription.CollectionMethod,
        &subscription.CancelAtPeriodEnd,
        &metadata,
        &resumeAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrSubscriptionNotFound
        }
        return nil, fmt.Errorf("error getting subscription by Stripe subscription ID: %w", err)
    }

    if cancelAt.Valid {
        t := cancelAt.Time
        subscription.CancelAt = &t
    }

    if canceledAt.Valid {
        t := canceledAt.Time
        subscription.CanceledAt = &t
    }

    if endedAt.Valid {
        t := endedAt.Time
        subscription.EndedAt = &t
    }

    if trialStart.Valid {
        t := trialStart.Time
        subscription.TrialStart = &t
    }

    if trialEnd.Valid {
        t := trialEnd.Time
        subscription.TrialEnd = &t
    }

    if resumeAt.Valid {
        t := resumeAt.Time
        subscription.ResumeAt = &t
    }

    // Handle JSON metadata conversion if needed

    return &subscription, nil
}

func (r *PostgresSubscriptionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Subscription, error) {
    query := `
        SELECT 
            id, user_id, price_id, quantity, status, 
            current_period_start, current_period_end, cancel_at, 
            canceled_at, ended_at, trial_start, trial_end, 
            created_at, updated_at, stripe_subscription_id, 
            stripe_customer_id, collection_method, 
            cancel_at_period_end, metadata, resume_at
        FROM subscriptions
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, fmt.Errorf("error listing subscriptions by user ID: %w", err)
    }
    defer rows.Close()

    var subscriptions []*models.Subscription

    for rows.Next() {
        var subscription models.Subscription
        var cancelAt, canceledAt, endedAt, trialStart, trialEnd, resumeAt sql.NullTime
        var metadata sql.NullString

        err := rows.Scan(
            &subscription.ID,
            &subscription.UserID,
            &subscription.PriceID,
            &subscription.Quantity,
            &subscription.Status,
            &subscription.CurrentPeriodStart,
            &subscription.CurrentPeriodEnd,
            &cancelAt,
            &canceledAt,
            &endedAt,
            &trialStart,
            &trialEnd,
            &subscription.CreatedAt,
            &subscription.UpdatedAt,
            &subscription.StripeSubscriptionID,
            &subscription.StripeCustomerID,
            &subscription.CollectionMethod,
            &subscription.CancelAtPeriodEnd,
            &metadata,
            &resumeAt,
        )

        if err != nil {
            return nil, fmt.Errorf("error scanning subscription row: %w", err)
        }

        if cancelAt.Valid {
            t := cancelAt.Time
            subscription.CancelAt = &t
        }

        if canceledAt.Valid {
            t := canceledAt.Time
            subscription.CanceledAt = &t
        }

        if endedAt.Valid {
            t := endedAt.Time
            subscription.EndedAt = &t
        }

        if trialStart.Valid {
            t := trialStart.Time
            subscription.TrialStart = &t
        }

        if trialEnd.Valid {
            t := trialEnd.Time
            subscription.TrialEnd = &t
        }

        if resumeAt.Valid {
            t := resumeAt.Time
            subscription.ResumeAt = &t
        }

        // Handle JSON metadata conversion if needed

        subscriptions = append(subscriptions, &subscription)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating subscription rows: %w", err)
    }

    return subscriptions, nil
}

func (r *PostgresSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*models.Subscription, error) {
    if limit <= 0 {
        limit = 10 // Default limit
    }

    var query string
    var args []interface{}

    if status != "" {
        // Query with status filter
        query = `
            SELECT 
                id, user_id, price_id, quantity, status, 
                current_period_start, current_period_end, cancel_at, 
                canceled_at, ended_at, trial_start, trial_end, 
                created_at, updated_at, stripe_subscription_id, 
                stripe_customer_id, collection_method, 
                cancel_at_period_end, metadata, resume_at
            FROM subscriptions
            WHERE user_id = $1 AND status = $2
            ORDER BY created_at DESC
            LIMIT $3 OFFSET $4
        `
        args = []interface{}{userID, status, limit, offset}
    } else {
        // Query without status filter
        query = `
            SELECT 
                id, user_id, price_id, quantity, status, 
                current_period_start, current_period_end, cancel_at, 
                canceled_at, ended_at, trial_start, trial_end, 
                created_at, updated_at, stripe_subscription_id, 
                stripe_customer_id, collection_method, 
                cancel_at_period_end, metadata, resume_at
            FROM subscriptions
            WHERE user_id = $1
            ORDER BY created_at DESC
            LIMIT $2 OFFSET $3
        `
        args = []interface{}{userID, limit, offset}
    }

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("error listing subscriptions by user ID: %w", err)
    }
    defer rows.Close()

    var subscriptions []*models.Subscription

    for rows.Next() {
        var subscription models.Subscription
        var cancelAt, canceledAt, endedAt, trialStart, trialEnd, resumeAt sql.NullTime
        var metadata sql.NullString

        err := rows.Scan(
            &subscription.ID,
            &subscription.UserID,
            &subscription.PriceID,
            &subscription.Quantity,
            &subscription.Status,
            &subscription.CurrentPeriodStart,
            &subscription.CurrentPeriodEnd,
            &cancelAt,
            &canceledAt,
            &endedAt,
            &trialStart,
            &trialEnd,
            &subscription.CreatedAt,
            &subscription.UpdatedAt,
            &subscription.StripeSubscriptionID,
            &subscription.StripeCustomerID,
            &subscription.CollectionMethod,
            &subscription.CancelAtPeriodEnd,
            &metadata,
            &resumeAt,
        )

        if err != nil {
            return nil, fmt.Errorf("error scanning subscription row: %w", err)
        }

        if cancelAt.Valid {
            t := cancelAt.Time
            subscription.CancelAt = &t
        }

        if canceledAt.Valid {
            t := canceledAt.Time
            subscription.CanceledAt = &t
        }

        if endedAt.Valid {
            t := endedAt.Time
            subscription.EndedAt = &t
        }

        if trialStart.Valid {
            t := trialStart.Time
            subscription.TrialStart = &t
        }

        if trialEnd.Valid {
            t := trialEnd.Time
            subscription.TrialEnd = &t
        }

        if resumeAt.Valid {
            t := resumeAt.Time
            subscription.ResumeAt = &t
        }

        // Handle JSON metadata conversion if needed

        subscriptions = append(subscriptions, &subscription)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating subscription rows: %w", err)
    }

    return subscriptions, nil
}

func (r *PostgresSubscriptionRepository) GetByCustomerID(ctx context.Context, customerID string, status string, limit, offset int) ([]*models.Subscription, error) {
    if limit <= 0 {
        limit = 10 // Default limit
    }

    var query string
    var args []interface{}

    if status != "" {
        // Query with status filter
        query = `
            SELECT 
                id, user_id, price_id, quantity, status, 
                current_period_start, current_period_end, cancel_at, 
                canceled_at, ended_at, trial_start, trial_end, 
                created_at, updated_at, stripe_subscription_id, 
                stripe_customer_id, collection_method, 
                cancel_at_period_end, metadata, resume_at
            FROM subscriptions
            WHERE stripe_customer_id = $1 AND status = $2
            ORDER BY created_at DESC
            LIMIT $3 OFFSET $4
        `
        args = []interface{}{customerID, status, limit, offset}
    } else {
        // Query without status filter
        query = `
            SELECT 
                id, user_id, price_id, quantity, status, 
                current_period_start, current_period_end, cancel_at, 
                canceled_at, ended_at, trial_start, trial_end, 
                created_at, updated_at, stripe_subscription_id, 
                stripe_customer_id, collection_method, 
                cancel_at_period_end, metadata, resume_at
            FROM subscriptions
            WHERE stripe_customer_id = $1
            ORDER BY created_at DESC
            LIMIT $2 OFFSET $3
        `
        args = []interface{}{customerID, limit, offset}
    }

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("error listing subscriptions by customer ID: %w", err)
    }
    defer rows.Close()

    var subscriptions []*models.Subscription

    for rows.Next() {
        var subscription models.Subscription
        var cancelAt, canceledAt, endedAt, trialStart, trialEnd, resumeAt sql.NullTime
        var metadata sql.NullString

        err := rows.Scan(
            &subscription.ID,
            &subscription.UserID,
            &subscription.PriceID,
            &subscription.Quantity,
            &subscription.Status,
            &subscription.CurrentPeriodStart,
            &subscription.CurrentPeriodEnd,
            &cancelAt,
            &canceledAt,
            &endedAt,
            &trialStart,
            &trialEnd,
            &subscription.CreatedAt,
            &subscription.UpdatedAt,
            &subscription.StripeSubscriptionID,
            &subscription.StripeCustomerID,
            &subscription.CollectionMethod,
            &subscription.CancelAtPeriodEnd,
            &metadata,
            &resumeAt,
        )

        if err != nil {
            return nil, fmt.Errorf("error scanning subscription row: %w", err)
        }

        if cancelAt.Valid {
            t := cancelAt.Time
            subscription.CancelAt = &t
        }

        if canceledAt.Valid {
            t := canceledAt.Time
            subscription.CanceledAt = &t
        }

        if endedAt.Valid {
            t := endedAt.Time
            subscription.EndedAt = &t
        }

        if trialStart.Valid {
            t := trialStart.Time
            subscription.TrialStart = &t
        }

        if trialEnd.Valid {
            t := trialEnd.Time
            subscription.TrialEnd = &t
        }

        if resumeAt.Valid {
            t := resumeAt.Time
            subscription.ResumeAt = &t
        }

        // Handle JSON metadata conversion if needed

        subscriptions = append(subscriptions, &subscription)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating subscription rows: %w", err)
    }

    return subscriptions, nil
}

func (r *PostgresSubscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
    // Update the timestamp
    subscription.UpdatedAt = time.Now()

    query := `
        UPDATE subscriptions
        SET 
            user_id = $1,
            price_id = $2,
            quantity = $3,
            status = $4,
            current_period_start = $5,
            current_period_end = $6,
            cancel_at = $7,
            canceled_at = $8,
            ended_at = $9,
            trial_start = $10,
            trial_end = $11,
            updated_at = $12,
            stripe_subscription_id = $13,
            stripe_customer_id = $14,
            collection_method = $15,
            cancel_at_period_end = $16,
            metadata = $17,
            resume_at = $18
        WHERE id = $19
        RETURNING id
    `

    var cancelAt sql.NullTime
    if subscription.CancelAt != nil {
        cancelAt.Time = *subscription.CancelAt
        cancelAt.Valid = true
    }

    var canceledAt sql.NullTime
    if subscription.CanceledAt != nil {
        canceledAt.Time = *subscription.CanceledAt
        canceledAt.Valid = true
    }

    var endedAt sql.NullTime
    if subscription.EndedAt != nil {
        endedAt.Time = *subscription.EndedAt
        endedAt.Valid = true
    }

    var trialStart sql.NullTime
    if subscription.TrialStart != nil {
        trialStart.Time = *subscription.TrialStart
        trialStart.Valid = true
    }

    var trialEnd sql.NullTime
    if subscription.TrialEnd != nil {
        trialEnd.Time = *subscription.TrialEnd
        trialEnd.Valid = true
    }

    var resumeAt sql.NullTime
    if subscription.ResumeAt != nil {
        resumeAt.Time = *subscription.ResumeAt
        resumeAt.Valid = true
    }

    var returnedID uuid.UUID
    err := r.db.QueryRowContext(
        ctx,
        query,
        subscription.UserID,
        subscription.PriceID,
        subscription.Quantity,
        subscription.Status,
        subscription.CurrentPeriodStart,
        subscription.CurrentPeriodEnd,
        cancelAt,
        canceledAt,
        endedAt,
        trialStart,
        trialEnd,
        subscription.UpdatedAt,
        subscription.StripeSubscriptionID,
        subscription.StripeCustomerID,
        subscription.CollectionMethod,
        subscription.CancelAtPeriodEnd,
        subscription.Metadata,
        resumeAt,
        subscription.ID,
    ).Scan(&returnedID)

    if err != nil {
        if err == sql.ErrNoRows {
            return ErrSubscriptionNotFound
        }
        // Check for unique violation
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
            return ErrSubscriptionExists
        }
        return fmt.Errorf("error updating subscription: %w", err)
    }

    return nil
}

func (r *PostgresSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("error deleting subscription: %w", err)
	}

	return nil
}

func (r *PostgresSubscriptionRepository) List(ctx context.Context, limit, offset int) ([]*models.Subscription, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, price_id, quantity, status, 
			current_period_start, current_period_end, cancel_at, 
			canceled_at, ended_at, trial_start, trial_end, 
			created_at, updated_at, stripe_subscription_id, 
			stripe_customer_id, collection_method, 
			cancel_at_period_end, metadata
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		var subscription models.Subscription
		var cancelAt, canceledAt, endedAt, trialStart, trialEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.PriceID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&cancelAt,
			&canceledAt,
			&endedAt,
			&trialStart,
			&trialEnd,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.StripeSubscriptionID,
			&subscription.StripeCustomerID,
			&subscription.CollectionMethod,
			&subscription.CancelAtPeriodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning subscription row: %w", err)
		}

		if cancelAt.Valid {
			t := cancelAt.Time
			subscription.CancelAt = &t
		}

		if canceledAt.Valid {
			t := canceledAt.Time
			subscription.CanceledAt = &t
		}

		if endedAt.Valid {
			t := endedAt.Time
			subscription.EndedAt = &t
		}

		if trialStart.Valid {
			t := trialStart.Time
			subscription.TrialStart = &t
		}

		if trialEnd.Valid {
			t := trialEnd.Time
			subscription.TrialEnd = &t
		}

		// Handle JSON metadata conversion if needed

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subscriptions, nil
}

func (r *PostgresSubscriptionRepository) ListByStatus(ctx context.Context, status models.SubscriptionStatus, limit, offset int) ([]*models.Subscription, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, user_id, price_id, quantity, status, 
			current_period_start, current_period_end, cancel_at, 
			canceled_at, ended_at, trial_start, trial_end, 
			created_at, updated_at, stripe_subscription_id, 
			stripe_customer_id, collection_method, 
			cancel_at_period_end, metadata
		FROM subscriptions
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing subscriptions by status: %w", err)
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		var subscription models.Subscription
		var cancelAt, canceledAt, endedAt, trialStart, trialEnd sql.NullTime
		var metadata sql.NullString

		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.PriceID,
			&subscription.Quantity,
			&subscription.Status,
			&subscription.CurrentPeriodStart,
			&subscription.CurrentPeriodEnd,
			&cancelAt,
			&canceledAt,
			&endedAt,
			&trialStart,
			&trialEnd,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.StripeSubscriptionID,
			&subscription.StripeCustomerID,
			&subscription.CollectionMethod,
			&subscription.CancelAtPeriodEnd,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning subscription row: %w", err)
		}

		if cancelAt.Valid {
			t := cancelAt.Time
			subscription.CancelAt = &t
		}

		if canceledAt.Valid {
			t := canceledAt.Time
			subscription.CanceledAt = &t
		}

		if endedAt.Valid {
			t := endedAt.Time
			subscription.EndedAt = &t
		}

		if trialStart.Valid {
			t := trialStart.Time
			subscription.TrialStart = &t
		}

		if trialEnd.Valid {
			t := trialEnd.Time
			subscription.TrialEnd = &t
		}

		// Handle JSON metadata conversion if needed

		subscriptions = append(subscriptions, &subscription)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscription rows: %w", err)
	}

	return subscriptions, nil
}