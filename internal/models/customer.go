package models

import (
	"time"
)

// Customer represents a registered user who can create subscriptions
type Customer struct {
	ID              int64     `json:"id" db:"id"`
	StripeCustomerID string    `json:"stripe_customer_id" db:"stripe_customer_id"`
	Email           string    `json:"email" db:"email"`
	Name            string    `json:"name" db:"name"`
	Phone           string    `json:"phone" db:"phone"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	
	// Relations (not stored in DB)
	Addresses     []CustomerAddress `json:"addresses,omitempty" db:"-"`
	Subscriptions []Subscription    `json:"subscriptions,omitempty" db:"-"`
}