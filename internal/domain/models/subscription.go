package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

// Subscription status constants - matching Stripe's status values
const (
	SubscriptionStatusActive            = "active"
	SubscriptionStatusPastDue           = "past_due"
	SubscriptionStatusIncomplete        = "incomplete"
	SubscriptionStatusIncompleteExpired = "incomplete_expired"
	SubscriptionStatusTrialing          = "trialing"
	SubscriptionStatusCanceled          = "canceled" // Note the American spelling used by Stripe
	SubscriptionStatusUnpaid            = "unpaid"
	SubscriptionStatusPaused            = "paused" // Our custom status
)

// Subscription represents a customer's subscription to a product
type Subscription struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	ProductID  uuid.UUID  `json:"product_id"`
	PriceID    uuid.UUID  `json:"price_id"`
	AddressID  *uuid.UUID `json:"address_id,omitempty"` // Optional for checkout process

	// Stripe subscription data
	StripeID     string `json:"stripe_id"`      // Main subscription ID
	StripeItemID string `json:"stripe_item_id"` // ID of the individual subscription item

	// Quantity and status
	Quantity int    `json:"quantity"`
	Status   string `json:"status"` // Using Stripe's status values

	// Billing period
	CurrentPeriodStart time.Time `json:"current_period_start"` // When the current billing period started
	CurrentPeriodEnd   time.Time `json:"current_period_end"`   // When the current billing period ends
	NextDeliveryDate   time.Time `json:"next_delivery_date"`   // When the coffee will be delivered

	// Cancellation details
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`  // Whether to cancel at period end
	CanceledAt        *time.Time `json:"canceled_at,omitempty"` // When the subscription was canceled

	// Metadata
	Metadata  map[string]string `json:"metadata,omitempty"` // Additional data
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// SubscriptionWithDetails includes related entity details for API responses
type SubscriptionWithDetails struct {
	Subscription
	ProductName   string `json:"product_name"`
	ProductImage  string `json:"product_image,omitempty"`
	PriceName     string `json:"price_name"`
	Interval      string `json:"interval"`       // week, month, year
	IntervalCount int    `json:"interval_count"` // e.g., 2 for bi-weekly
	Amount        int64  `json:"amount"`         // Price amount in cents
	Currency      string `json:"currency"`       // USD, EUR, etc.
}
