// repository/cart_repository.go
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
	ErrCartNotFound = errors.New("cart not found")
)

type CartRepository interface {
	Create(ctx context.Context, cart *models.Cart) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Cart, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error)
	GetBySessionID(ctx context.Context, sessionID string) (*models.Cart, error)
	Update(ctx context.Context, cart *models.Cart) error
	Delete(ctx context.Context, id uuid.UUID) error
	CleanExpiredCarts(ctx context.Context) (int, error)
}

type PostgresCartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &PostgresCartRepository{
		db: db,
	}
}

func (r *PostgresCartRepository) Create(ctx context.Context, cart *models.Cart) error {
	// Generate a new UUID if not provided
	if cart.ID == uuid.Nil {
		cart.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	cart.CreatedAt = now
	cart.UpdatedAt = now

	// Set expiration if not provided (default 7 days)
	if cart.ExpiresAt == nil {
		expires := now.Add(7 * 24 * time.Hour)
		cart.ExpiresAt = &expires
	}

	query := `
		INSERT INTO carts (
			id, user_id, session_id, created_at, 
			updated_at, expires_at, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var userID sql.NullString
	if cart.UserID != nil {
		userID.String = cart.UserID.String()
		userID.Valid = true
	}

	var sessionID sql.NullString
	if cart.SessionID != nil {
		sessionID.String = *cart.SessionID
		sessionID.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		cart.ID,
		userID,
		sessionID,
		cart.CreatedAt,
		cart.UpdatedAt,
		cart.ExpiresAt,
		cart.Metadata,
	).Scan(&cart.ID)

	if err != nil {
		return fmt.Errorf("error creating cart: %w", err)
	}

	return nil
}

func (r *PostgresCartRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Cart, error) {
	query := `
		SELECT 
			id, user_id, session_id, created_at, 
			updated_at, expires_at, metadata
		FROM carts
		WHERE id = $1
	`

	var cart models.Cart
	var userID, sessionID sql.NullString
	var expiresAt sql.NullTime
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cart.ID,
		&userID,
		&sessionID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
		&expiresAt,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("error getting cart by ID: %w", err)
	}

	if userID.Valid {
		parsedID, err := uuid.Parse(userID.String)
		if err == nil {
			cart.UserID = &parsedID
		}
	}

	if sessionID.Valid {
		cart.SessionID = &sessionID.String
	}

	if expiresAt.Valid {
		cart.ExpiresAt = &expiresAt.Time
	}

	// Handle JSON metadata conversion if needed

	return &cart, nil
}

func (r *PostgresCartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	query := `
		SELECT 
			id, user_id, session_id, created_at, 
			updated_at, expires_at, metadata
		FROM carts
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var cart models.Cart
	var uID, sessionID sql.NullString
	var expiresAt sql.NullTime
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&cart.ID,
		&uID,
		&sessionID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
		&expiresAt,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("error getting cart by user ID: %w", err)
	}

	if uID.Valid {
		parsedID, err := uuid.Parse(uID.String)
		if err == nil {
			cart.UserID = &parsedID
		}
	}

	if sessionID.Valid {
		cart.SessionID = &sessionID.String
	}

	if expiresAt.Valid {
		cart.ExpiresAt = &expiresAt.Time
	}

	// Handle JSON metadata conversion if needed

	return &cart, nil
}

func (r *PostgresCartRepository) GetBySessionID(ctx context.Context, sessionID string) (*models.Cart, error) {
	query := `
		SELECT 
			id, user_id, session_id, created_at, 
			updated_at, expires_at, metadata
		FROM carts
		WHERE session_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var cart models.Cart
	var userID, sID sql.NullString
	var expiresAt sql.NullTime
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&cart.ID,
		&userID,
		&sID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
		&expiresAt,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("error getting cart by session ID: %w", err)
	}

	if userID.Valid {
		parsedID, err := uuid.Parse(userID.String)
		if err == nil {
			cart.UserID = &parsedID
		}
	}

	if sID.Valid {
		cart.SessionID = &sID.String
	}

	if expiresAt.Valid {
		cart.ExpiresAt = &expiresAt.Time
	}

	// Handle JSON metadata conversion if needed

	return &cart, nil
}

func (r *PostgresCartRepository) Update(ctx context.Context, cart *models.Cart) error {
	// Update the timestamp
	cart.UpdatedAt = time.Now()

	query := `
		UPDATE carts
		SET 
			user_id = $1,
			session_id = $2,
			updated_at = $3,
			expires_at = $4,
			metadata = $5
		WHERE id = $6
		RETURNING id
	`

	var userID sql.NullString
	if cart.UserID != nil {
		userID.String = cart.UserID.String()
		userID.Valid = true
	}

	var sessionID sql.NullString
	if cart.SessionID != nil {
		sessionID.String = *cart.SessionID
		sessionID.Valid = true
	}

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		sessionID,
		cart.UpdatedAt,
		cart.ExpiresAt,
		cart.Metadata,
		cart.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartNotFound
		}
		return fmt.Errorf("error updating cart: %w", err)
	}

	return nil
}

func (r *PostgresCartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM carts WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrCartNotFound
		}
		return fmt.Errorf("error deleting cart: %w", err)
	}

	return nil
}

func (r *PostgresCartRepository) CleanExpiredCarts(ctx context.Context) (int, error) {
	query := `DELETE FROM carts WHERE expires_at < NOW() RETURNING id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("error cleaning expired carts: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating deleted cart rows: %w", err)
	}

	return count, nil
}