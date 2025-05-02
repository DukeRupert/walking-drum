package models

import (
	"time"
)

// Product represents a coffee product available for subscription
type Product struct {
	ID              int64     `json:"id" db:"id"`
	StripeProductID string    `json:"stripe_product_id" db:"stripe_product_id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Origin          string    `json:"origin" db:"origin"`
	RoastLevel      string    `json:"roast_level" db:"roast_level"`
	Active          bool      `json:"active" db:"active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	
	// Relations (not stored in DB)
	Prices []ProductPrice `json:"prices,omitempty" db:"-"`
}