// internal/domain/models/subscription.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPaused    = "paused"
	SubscriptionStatusCancelled = "cancelled"
)

// Subscription represents a customer's subscription to a product
type Subscription struct {
	ID                 uuid.UUID  `json:"id"`
	CustomerID         uuid.UUID  `json:"customer_id"`
	ProductID          uuid.UUID  `json:"product_id"`
	PriceID            uuid.UUID  `json:"price_id"`
	AddressID          uuid.UUID  `json:"address_id"`
	Quantity           int        `json:"quantity"`
	Status             string     `json:"status"` // active, paused, cancelled
	StripeID           string     `json:"stripe_id"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
