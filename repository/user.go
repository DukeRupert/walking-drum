// repository/user.go
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
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	// Generate a new UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (
			id, email, password_hash, name, 
			created_at, updated_at, stripe_customer_id, 
			is_active, metadata
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
		user.StripeCustomerID,
		user.IsActive,
		user.Metadata,
	).Scan(&user.ID)

	if err != nil {
		// Check for unique violation on email
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserExists
		}
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, name, 
			created_at, updated_at, stripe_customer_id, 
			is_active, metadata
		FROM users
		WHERE id = $1
	`

	var user models.User
	var stripeCustomerID sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&stripeCustomerID,
		&user.IsActive,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}

	if stripeCustomerID.Valid {
		val := stripeCustomerID.String
		user.StripeCustomerID = &val
	}

	// Handle JSON metadata conversion if needed

	return &user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, name, 
			created_at, updated_at, stripe_customer_id, 
			is_active, metadata
		FROM users
		WHERE email = $1
	`

	var user models.User
	var stripeCustomerID sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&stripeCustomerID,
		&user.IsActive,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	if stripeCustomerID.Valid {
		val := stripeCustomerID.String
		user.StripeCustomerID = &val
	}

	// Handle JSON metadata conversion if needed

	return &user, nil
}

func (r *PostgresUserRepository) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, name, 
			created_at, updated_at, stripe_customer_id, 
			is_active, metadata
		FROM users
		WHERE stripe_customer_id = $1
	`

	var user models.User
	var stripeCustomerIDDb sql.NullString
	var metadata sql.NullString

	err := r.db.QueryRowContext(ctx, query, stripeCustomerID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&stripeCustomerIDDb,
		&user.IsActive,
		&metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user by Stripe customer ID: %w", err)
	}

	if stripeCustomerIDDb.Valid {
		val := stripeCustomerIDDb.String
		user.StripeCustomerID = &val
	}

	// Handle JSON metadata conversion if needed

	return &user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	// Update the timestamp
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET 
			email = $1,
			password_hash = $2,
			name = $3,
			updated_at = $4,
			stripe_customer_id = $5,
			is_active = $6,
			metadata = $7
		WHERE id = $8
		RETURNING id
	`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.UpdatedAt,
		user.StripeCustomerID,
		user.IsActive,
		user.Metadata,
		user.ID,
	).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		// Check for unique violation on email
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserExists
		}
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1 RETURNING id`

	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&returnedID)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			id, email, password_hash, name, 
			created_at, updated_at, stripe_customer_id, 
			is_active, metadata
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		var user models.User
		var stripeCustomerID sql.NullString
		var metadata sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Name,
			&user.CreatedAt,
			&user.UpdatedAt,
			&stripeCustomerID,
			&user.IsActive,
			&metadata,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		if stripeCustomerID.Valid {
			val := stripeCustomerID.String
			user.StripeCustomerID = &val
		}

		// Handle JSON metadata conversion if needed

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
