// models/subscription.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive          SubscriptionStatus = "active"
	SubscriptionStatusPastDue         SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled        SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid          SubscriptionStatus = "unpaid"
	SubscriptionStatusTrialing        SubscriptionStatus = "trialing"
	SubscriptionStatusIncomplete      SubscriptionStatus = "incomplete"
	SubscriptionStatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
)

type Subscription struct {
	ID                 uuid.UUID          `json:"id"`
	UserID             uuid.UUID          `json:"user_id"`
	PriceID            uuid.UUID          `json:"price_id"`
	Quantity           int                `json:"quantity"`
	Status             SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelAt           *time.Time         `json:"cancel_at,omitempty"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	EndedAt            *time.Time         `json:"ended_at,omitempty"`
	TrialStart         *time.Time         `json:"trial_start,omitempty"`
	TrialEnd           *time.Time         `json:"trial_end,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	StripeSubscriptionID string           `json:"stripe_subscription_id"`
	StripeCustomerID    string           `json:"stripe_customer_id"`
	CollectionMethod   string            `json:"collection_method"`
	CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
	Metadata           *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	User               *User              `json:"user,omitempty"`
	Price              *Price             `json:"price,omitempty"`
}