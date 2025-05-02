package domain

import (
	"time"
)

// WebhookEvent represents a processed Stripe webhook event
type WebhookEvent struct {
	ID            int64     `json:"id" db:"id"`
	StripeEventID string    `json:"stripe_event_id" db:"stripe_event_id"`
	Type          string    `json:"type" db:"type"`
	Object        string    `json:"object" db:"object"`
	Processed     bool      `json:"processed" db:"processed"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}