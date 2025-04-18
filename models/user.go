// models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"` // Not exposed in JSON
	Name            string     `json:"name"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	StripeCustomerID *string    `json:"stripe_customer_id,omitempty"`
	IsActive        bool       `json:"is_active"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
}