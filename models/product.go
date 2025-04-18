// models/product.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	StripeProductID *string    `json:"stripe_product_id,omitempty"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
}