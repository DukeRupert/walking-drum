package models

import (
	"time"
)

// Subscription represents a recurring coffee subscription
type Subscription struct {
	ID                 int64      `json:"id" db:"id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	CustomerID         int64      `json:"customer_id" db:"customer_id"`
	PriceID            int64      `json:"price_id" db:"price_id"`
	Status             string     `json:"status" db:"status"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	
	// Relations (not stored in DB)
	Customer      *Customer     `json:"customer,omitempty" db:"-"`
	ProductPrice  *ProductPrice `json:"product_price,omitempty" db:"-"`
}