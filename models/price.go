// models/price.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type BillingInterval string

const (
	BillingIntervalDay   BillingInterval = "day"
	BillingIntervalWeek  BillingInterval = "week"
	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

type Price struct {
	ID             uuid.UUID       `json:"id"`
	ProductID      uuid.UUID       `json:"product_id"`
	Amount         int64           `json:"amount"` // Amount in cents
	Currency       string          `json:"currency"`
	IntervalType   BillingInterval `json:"interval_type"`
	IntervalCount  int             `json:"interval_count"`
	TrialPeriodDays *int            `json:"trial_period_days,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	StripePriceID  *string          `json:"stripe_price_id,omitempty"`
	IsActive       bool            `json:"is_active"`
	Nickname       *string          `json:"nickname,omitempty"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	Product        *Product        `json:"product,omitempty"`
}