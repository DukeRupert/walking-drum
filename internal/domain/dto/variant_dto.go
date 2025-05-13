// internal/domain/dto/variant_dto.go
package dto

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// VariantCreateDTO represents the data needed to create a new variant
type VariantCreateDTO struct {
	ProductID     uuid.UUID `json:"product_id"`
	PriceID       uuid.UUID `json:"price_id"`
	StripePriceID string    `json:"stripe_price_id"`
	Weight        string    `json:"weight"`
	Grind         string    `json:"grind"`
	Active        bool      `json:"active"`
	StockLevel    int       `json:"stock_level"`
}

// Valid validates the VariantCreateDTO
func (v *VariantCreateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if v.ProductID == uuid.Nil {
		problems["product_id"] = "product ID is required"
	}

	if v.PriceID == uuid.Nil {
		problems["price_id"] = "price ID is required"
	}

	if v.StripePriceID == "" {
		problems["stripe_price_id"] = "Stripe price ID is required"
	}

	if v.Weight == "" {
		problems["weight"] = "weight is required"
	} else if !isValidWeight(v.Weight) {
		problems["weight"] = "invalid weight option"
	}

	if v.Grind == "" {
		problems["grind"] = "grind is required"
	} else if !isValidGrind(v.Grind) {
		problems["grind"] = "invalid grind option"
	}

	if v.StockLevel < 0 {
		problems["stock_level"] = "stock level cannot be negative"
	}

	return problems
}

// VariantUpdateDTO represents the data needed to update an existing variant
type VariantUpdateDTO struct {
	ProductID     *uuid.UUID `json:"product_id,omitempty"`
	PriceID       *uuid.UUID `json:"price_id,omitempty"`
	StripePriceID *string    `json:"stripe_price_id,omitempty"`
	Weight        *string    `json:"weight,omitempty"`
	Grind         *string    `json:"grind,omitempty"`
	Active        *bool      `json:"active,omitempty"`
	StockLevel    *int       `json:"stock_level,omitempty"`
}

// Valid validates the VariantUpdateDTO
func (v *VariantUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if v.ProductID != nil && *v.ProductID == uuid.Nil {
		problems["product_id"] = "product ID cannot be empty when provided"
	}

	if v.PriceID != nil && *v.PriceID == uuid.Nil {
		problems["price_id"] = "price ID cannot be empty when provided"
	}

	if v.StripePriceID != nil && *v.StripePriceID == "" {
		problems["stripe_price_id"] = "Stripe price ID cannot be empty when provided"
	}

	if v.Weight != nil {
		if *v.Weight == "" {
			problems["weight"] = "weight cannot be empty when provided"
		} else if !isValidWeight(*v.Weight) {
			problems["weight"] = "invalid weight option"
		}
	}

	if v.Grind != nil {
		if *v.Grind == "" {
			problems["grind"] = "grind cannot be empty when provided"
		} else if !isValidGrind(*v.Grind) {
			problems["grind"] = "invalid grind option"
		}
	}

	if v.StockLevel != nil && *v.StockLevel < 0 {
		problems["stock_level"] = "stock level cannot be negative"
	}

	return problems
}

// Helper functions to validate variant options
func isValidWeight(weight string) bool {
	validWeights := []string{"12oz", "3lb", "5lb"}

	for _, validWeight := range validWeights {
		if weight == validWeight {
			return true
		}
	}

	return false
}

func isValidGrind(grind string) bool {
	validGrinds := []string{"Whole Bean", "Drip Ground"}

	for _, validGrind := range validGrinds {
		if grind == validGrind {
			return true
		}
	}

	return false
}

// VariantListResponse represents a paginated list of variants
type VariantListResponse struct {
	Data       []*VariantResponse `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

// VariantResponse represents a variant with product details for API responses
type VariantResponse struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image,omitempty"`
	Price        int64     `json:"price"`
	Currency     string    `json:"currency"`
	Weight       string    `json:"weight"`
	Grind        string    `json:"grind"`
	Active       bool      `json:"active"`
	StockLevel   int       `json:"stock_level"`
	Origin       string    `json:"origin,omitempty"`
	RoastLevel   string    `json:"roast_level,omitempty"`
	FlavorNotes  string    `json:"flavor_notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VariantOptionsResponse represents the available options for variants
type VariantOptionsResponse struct {
	Weights []string `json:"weights"`
	Grinds  []string `json:"grinds"`
}
