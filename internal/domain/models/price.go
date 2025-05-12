// internal/domain/models/price.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Price represents the pricing options for subscriptions or one-time purchases
type Price struct {
    ID            uuid.UUID `json:"id"`
    ProductID     uuid.UUID `json:"product_id"`
    Name          string    `json:"name"`
    Amount        int64     `json:"amount"` // Price in cents
    Currency      string    `json:"currency"`
    Type          string    `json:"type"`   // one_time|recurring
    Interval      string    `json:"interval,omitempty"`       // week|month|year (used only for recurring)
    IntervalCount int       `json:"interval_count,omitempty"` // Number of intervals between charges (used only for recurring)
    Active        bool      `json:"active"`
    StripeID      string    `json:"stripe_id"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
