// repository/price.go
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
	ErrPriceNotFound = errors.New("price not found")
	ErrPriceExists   = errors.New("price already exists with this stripe price ID")
)

type PriceRepository interface {
	Create(ctx context.Context, price *models.Price) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error)
	GetByStripePriceID(ctx context.Context, stripePriceID string) (*models.Price, error)
	ListByProductID(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]*models.Price, error)
	Update(ctx context.Context, price *models.Price) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Price, error)
	ListActive(ctx context.Context, limit, offset int) ([]*models.Price, error)
}

type PostgresPriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) PriceRepository {
	return &PostgresPriceRepository{
		db: db,
	}
}

func (r *PostgresPriceRepository) Create(ctx context.Context, price *models.Price) error {
	// Generate a new UUID if not provided
	if price.ID == uuid.Nil {
		price.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	price.CreatedAt = now
	price.UpdatedAt = now

	query := `
		INSERT INTO prices (
			id, product_id, amount, currency, interval_type, 
			interval_count, trial_period_days, created_at, updated_at, 
			stripe_price_id, is_active, nickname, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	var trialPeriodDays sql.NullInt32
	if price.TrialPeriodDays != nil {
		trialPeriodDays.Int32 = int32(*price.TrialPeriodDays)
		trialPeriodDays.Valid = true
	}

	var stripePriceID sql.NullString
	if price.StripePriceID != nil {
		stripePriceID.String = *price.StripePriceID
		stripePriceID.Valid = true
	}

	var nickname sql.NullString
	if price.Nickname != nil {
		nickname.String = *price.Nickname
		nickname.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		price.ID,
		price.ProductID,
		price.Amount,
		price.Currency,
		price.IntervalType,
		price.IntervalCount,
		trialPeriodDays,
		price.CreatedAt,
		price.UpdatedAt,
		stripePriceID,
		price.IsActive,
		nickname,
		price.Metadata,
	).Scan(&price.ID)

	if err != nil {
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrPriceExists
		}
		return fmt.Errorf("error creating price: %w", err)
	}

	return nil
}

func (r *PostgresPriceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Price, error) {
	query := `
		SELECT 
			id, product_id, amount, currency, interval_type, 
			interval_count, trial_period_days, created_at, updated_at, 
			stripe_price_id, is_active, nickname, metadata
		FROM prices
		WHERE id = $1
	`

	var price models.Price
	var trialPeriodDays sql.NullInt32
	var stripePriceID sql.NullString
	var nickname sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&price.ID,
		&price.ProductID,
		&price.Amount,
		&price.Currency,
		&price.IntervalType,
		&price.IntervalCount,
		&trialPeriodDays,
		&price.CreatedAt,
		&price.UpdatedAt,
		&stripePriceID,
		&price.IsActive,
		&nickname,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPriceNotFound
		}
		return nil, fmt.Errorf("error getting price by ID: %w", err)
	}

	if trialPeriodDays.Valid {
		val := int(trialPeriodDays.Int32)
		price.TrialPeriodDays = &val
	}

	if stripePriceID.Valid {
		val := stripePriceID.String
		price.StripePriceID = &val
	}

	if nickname.Valid {
		val := nickname.String
		price.Nickname = &val
	}

	// Handle JSON metadata conversion if needed

	return &price, nil
}

func (r *PostgresPriceRepository) GetByStripePriceID(ctx context.Context, stripePriceID string) (*models.Price, error) {
	query := `
		SELECT 
			id, product_id, amount, currency, interval_type, 
			interval_count, trial_period_days, created_at, updated_at, 
			stripe_price_id, is_active, nickname, metadata
		FROM prices
		WHERE stripe_price_id = $1
	`

	var price models.Price
	var trialPeriodDays sql.NullInt32
	var stripePriceIDDb sql.NullString
	var nickname sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, stripePriceID).Scan(
		&price.ID,
		&price.ProductID,
		&price.Amount,
		&price.Currency,
		&price.IntervalType,
		&price.IntervalCount,
		&trialPeriodDays,
		&price.CreatedAt,
		&price.UpdatedAt,
		&stripePriceIDDb,
		&price.IsActive,
		&nickname,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPriceNotFound
		}
		return nil, fmt.Errorf("error getting price by Stripe price ID: %w", err)
	}

	if trialPeriodDays.Valid {
		val := int(trialPeriodDays.Int32)
		price.TrialPeriodDays = &val
	}

	if stripePriceIDDb.Valid {
		val := stripePriceIDDb.String
		price.StripePriceID = &val
	}

	if nickname.Valid {
		val := nickname.String
		price.Nickname = &val
	}

	// Handle JSON metadata conversion if needed

	return &price, nil
}

func (r *PostgresPriceRepository) ListByProductID(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]*models.Price, error) {
	var query string
	var args []interface{}

	if activeOnly {
		query = `
			SELECT 
				id, product_id, amount, currency, interval_type, 
				interval_count, trial_period_days, created_at, updated_at, 
				stripe_price_id, is_active, nickname, metadata
			FROM prices
			WHERE product_id = $1 AND is_active = true
			ORDER BY amount ASC
		`
		args = []interface{}{productID}
	} else {
		query = `
			SELECT 
				id, product_id, amount, currency, interval_type, 
				interval_count, trial_period_days, created_at, updated_at, 
				stripe_price_id, is_active, nickname, metadata
			FROM prices
			WHERE product_id = $1
			ORDER BY amount ASC
		`
		args = []interface{}{productID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing prices by product ID: %w", err)
	}
	defer rows.Close()

	var prices []*models.Price

	for rows.Next() {
		var price models.Price
		var trialPeriodDays sql.NullInt32
		var stripePriceID sql.NullString
		var nickname sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&price.ID,
			&price.ProductID,
			&price.Amount,
			&price.Currency,
			&price.IntervalType,
			&price.IntervalCount,
			&trialPeriodDays,
			&price.CreatedAt,
			&price.UpdatedAt,
			&stripePriceID,
			&price.IsActive,
			&nickname,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning price row: %w", err)
		}

		if trialPeriodDays.Valid {
			val := int(trialPeriodDays.Int32)
			price.TrialPeriodDays = &val
		}

		if stripePriceID.Valid {
			val := stripePriceID.String
			price.StripePriceID = &val
		}

		if nickname.Valid {
			val := nickname.String
			price.Nickname = &val
		}

		// Handle JSON metadata conversion if needed

		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return prices, nil
}

func (r *PostgresPriceRepository) Update(ctx context.Context, price *models.Price) error {
	// Update the timestamp
	price.UpdatedAt = time.Now()

	query := `
		UPDATE prices
		SET 
			product_id = $1,
			amount = $2,
			currency = $3,
			interval_type = $4,
			interval_count = $5,
			trial_period_days = $6,
			updated_at = $7,
			stripe_price_id = $8,
			is_active = $9,
			nickname = $10,
			metadata = $11
		WHERE id = $12
		RETURNING id
	`

	var trialPeriodDays sql.NullInt32
	if price.TrialPeriodDays != nil {
		trialPeriodDays.Int32 = int32(*price.TrialPeriodDays)
		trialPeriodDays.Valid = true
	}

	var stripePriceID sql.NullString
	if price.StripePriceID != nil {
		stripePriceID.String = *price.StripePriceID
		stripePriceID.Valid = true
	}

	var nickname sql.NullString
	if price.Nickname != nil {
		nickname.String = *price.Nickname
		nickname.Valid = true
	}

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		price.ProductID,
		price.Amount,
		price.Currency,
		price.IntervalType,
		price.IntervalCount,
		trialPeriodDays,
		price.UpdatedAt,
		stripePriceID,
		price.IsActive,
		nickname,
		price.Metadata,
		price.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPriceNotFound
		}
		// Check for unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrPriceExists
		}
		return fmt.Errorf("error updating price: %w", err)
	}

	return nil
}

func (r *PostgresPriceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prices WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPriceNotFound
		}
		return fmt.Errorf("error deleting price: %w", err)
	}

	return nil
}

func (r *PostgresPriceRepository) List(ctx context.Context, limit, offset int) ([]*models.Price, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, product_id, amount, currency, interval_type, 
			interval_count, trial_period_days, created_at, updated_at, 
			stripe_price_id, is_active, nickname, metadata
		FROM prices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing prices: %w", err)
	}
	defer rows.Close()

	var prices []*models.Price

	for rows.Next() {
		var price models.Price
		var trialPeriodDays sql.NullInt32
		var stripePriceID sql.NullString
		var nickname sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&price.ID,
			&price.ProductID,
			&price.Amount,
			&price.Currency,
			&price.IntervalType,
			&price.IntervalCount,
			&trialPeriodDays,
			&price.CreatedAt,
			&price.UpdatedAt,
			&stripePriceID,
			&price.IsActive,
			&nickname,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning price row: %w", err)
		}

		if trialPeriodDays.Valid {
			val := int(trialPeriodDays.Int32)
			price.TrialPeriodDays = &val
		}

		if stripePriceID.Valid {
			val := stripePriceID.String
			price.StripePriceID = &val
		}

		if nickname.Valid {
			val := nickname.String
			price.Nickname = &val
		}

		// Handle JSON metadata conversion if needed

		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return prices, nil
}

func (r *PostgresPriceRepository) ListActive(ctx context.Context, limit, offset int) ([]*models.Price, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, product_id, amount, currency, interval_type, 
			interval_count, trial_period_days, created_at, updated_at, 
			stripe_price_id, is_active, nickname, metadata
		FROM prices
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing active prices: %w", err)
	}
	defer rows.Close()

	var prices []*models.Price

	for rows.Next() {
		var price models.Price
		var trialPeriodDays sql.NullInt32
		var stripePriceID sql.NullString
		var nickname sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&price.ID,
			&price.ProductID,
			&price.Amount,
			&price.Currency,
			&price.IntervalType,
			&price.IntervalCount,
			&trialPeriodDays,
			&price.CreatedAt,
			&price.UpdatedAt,
			&stripePriceID,
			&price.IsActive,
			&nickname,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning price row: %w", err)
		}

		if trialPeriodDays.Valid {
			val := int(trialPeriodDays.Int32)
			price.TrialPeriodDays = &val
		}

		if stripePriceID.Valid {
			val := stripePriceID.String
			price.StripePriceID = &val
		}

		if nickname.Valid {
			val := nickname.String
			price.Nickname = &val
		}

		// Handle JSON metadata conversion if needed

		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price rows: %w", err)
	}

	return prices, nil
}
