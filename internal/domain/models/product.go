// internal/domain/models/product.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a coffee product available for subscription
type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Active      bool      `json:"active"`
	StockLevel  int       `json:"stock_level"`
	Weight      int       `json:"weight"` // Weight in grams
	Origin      string    `json:"origin"`
	RoastLevel  string    `json:"roast_level"`
	FlavorNotes string    `json:"flavor_notes"`
	StripeID    string    `json:"stripe_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
