// internal/domain/models/customer.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a subscriber in the system
type Customer struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	StripeID    string    `json:"stripe_id"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
