package models

import (
	"time"
)

// ProductPrice represents pricing options for a product
type ProductPrice struct {
	ID             int64     `json:"id" db:"id"`
	ProductID      int64     `json:"product_id" db:"product_id"`
	StripePriceID  string    `json:"stripe_price_id" db:"stripe_price_id"`
	Weight         string    `json:"weight" db:"weight"`
	Grind          string    `json:"grind" db:"grind"`
	Price          float64   `json:"price" db:"price"`
	IsDefault      bool      `json:"is_default" db:"is_default"`
	Active         bool      `json:"active" db:"active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	
	// Relations (not stored in DB)
	Product *Product `json:"product,omitempty" db:"-"`
}