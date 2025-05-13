package models

import (
	"time"

	"github.com/google/uuid"
)

// WeightOption represents available weight options for coffee products
type WeightOption string

// GrindOption represents available grind options for coffee products
type GrindOption string

// Weight options constants
const (
	WeightTwelveOunce WeightOption = "12oz"
	WeightThreePound  WeightOption = "3lb"
	WeightFivePound   WeightOption = "5lb"
)

// Grind options constants
const (
	GrindWholeBeanOption GrindOption = "Whole Bean"
	GrindDripOption      GrindOption = "Drip Ground"
)

// Variant represents a specific product variant (combination of product, weight, and grind)
type Variant struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	PriceID       uuid.UUID `json:"price_id"`
	StripePriceID string    `json:"stripe_price_id"`
	Weight        string    `json:"weight"`
	Grind         string    `json:"grind"`
	Active        bool      `json:"active"`
	StockLevel    int       `json:"stock_level"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// VariantWithDetails includes product and price details for API responses
type VariantWithDetails struct {
	Variant
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image,omitempty"`
	Origin       string `json:"origin,omitempty"`
	RoastLevel   string `json:"roast_level,omitempty"`
	FlavorNotes  string `json:"flavor_notes,omitempty"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	PriceName    string `json:"price_name,omitempty"`
}

// GetWeightOptions returns all available weight options
func GetWeightOptions() []WeightOption {
	return []WeightOption{
		WeightTwelveOunce,
		WeightThreePound,
		WeightFivePound,
	}
}

// GetGrindOptions returns all available grind options
func GetGrindOptions() []GrindOption {
	return []GrindOption{
		GrindWholeBeanOption,
		GrindDripOption,
	}
}
