// internal/domain/dto/product_dto.go
package dto

import (
	"context"
)

// Validator is an object that can be validated
type Validator interface {
	// Valid checks the object and returns any problems
	// If len(problems) == 0 then the object is valid
	Valid(ctx context.Context) map[string]string
}

// ProductCreateDTO represents the data needed to create a new product
type ProductCreateDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Active      bool   `json:"active"`
	StockLevel  int    `json:"stock_level"`
	Weight      int    `json:"weight"` // Weight in grams
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	FlavorNotes string `json:"flavor_notes"`
}

// Valid validates the ProductCreateDTO
func (p *ProductCreateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if p.Name == "" {
		problems["name"] = "name is required"
	}

	if p.Description == "" {
		problems["description"] = "description is required"
	}

	if p.Weight <= 0 {
		problems["weight"] = "weight must be greater than 0"
	}

	if p.StockLevel < 0 {
		problems["stock_level"] = "stock level cannot be negative"
	}

	return problems
}

// ProductUpdateDTO represents the data that can be updated for a product
type ProductUpdateDTO struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Active      *bool  `json:"active,omitempty"`
	StockLevel  *int   `json:"stock_level,omitempty"`
	Weight      *int   `json:"weight,omitempty"` // Weight in grams
	Origin      string `json:"origin,omitempty"`
	RoastLevel  string `json:"roast_level,omitempty"`
	FlavorNotes string `json:"flavor_notes,omitempty"`
}

// Valid validates the ProductUpdateDTO
func (p *ProductUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if p.Weight != nil && *p.Weight <= 0 {
		problems["weight"] = "weight must be greater than 0"
	}

	if p.StockLevel != nil && *p.StockLevel < 0 {
		problems["stock_level"] = "stock level cannot be negative"
	}

	return problems
}

// ProductResponseDTO represents the data returned to the client
type ProductResponseDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Active      bool   `json:"active"`
	StockLevel  int    `json:"stock_level"`
	Weight      int    `json:"weight"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	FlavorNotes string `json:"flavor_notes"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}