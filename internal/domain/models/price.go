// internal/domain/models/price.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Price represents the pricing options for subscriptions
type Price struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	Name          string    `json:"name"`
	Amount        int64     `json:"amount"` // Price in cents
	Currency      string    `json:"currency"`
	Interval      string    `json:"interval"`       // week|month|year
	IntervalCount int       `json:"interval_count"` // Number of intervals between charges
	Active        bool      `json:"active"`
	StripeID      string    `json:"stripe_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
