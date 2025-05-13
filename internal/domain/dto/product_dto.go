// internal/domain/dto/product_dto.go
package dto

import (
	"context"
	"net/url"
	"strings"
)

// Validator is an object that can be validated
type Validator interface {
	// Valid checks the object and returns any problems
	// If len(problems) == 0 then the object is valid
	Valid(ctx context.Context) map[string]string
}

// Valid roast levels
var validRoastLevels = map[string]bool{
	"light":  true,
	"medium": true,
	"dark":   true,
}

// Valid options keys
var validOptionKeys = map[string]bool{
	"weight": true,
	"grind":  true,
}

// ProductCreateDTO represents the data needed to create a new product
type ProductCreateDTO struct {
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	ImageURL          string              `json:"image_url"`
	Active            bool                `json:"active"`
	StockLevel        int                 `json:"stock_level"`
	Weight            int                 `json:"weight"` // Weight in grams
	Origin            string              `json:"origin"`
	RoastLevel        string              `json:"roast_level"`
	FlavorNotes       string              `json:"flavor_notes"`
	Options           map[string][]string `json:"options"`            // Product options (e.g., weight, grind)
	AllowSubscription bool                `json:"allow_subscription"` // Flag to indicate if product can be subscribed to
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

	if p.RoastLevel != "" && !validRoastLevels[strings.ToLower(p.RoastLevel)] {
		problems["roast_level"] = "must be one of: light, medium, dark"
	}

	// Validate ImageURL if provided
	if p.ImageURL != "" {
		_, err := url.ParseRequestURI(p.ImageURL)
		if err != nil {
			problems["image_url"] = "must be a valid URL"
		}
	}

	// Validate options
	if p.Options != nil {
		for key, values := range p.Options {
			if !validOptionKeys[key] {
				problems["options."+key] = "is not a valid option key"
			}
			if len(values) == 0 {
				problems["options."+key] = "must have at least one value"
			}
		}

		// Check that weight and grind options are defined if product allows subscription
		if p.AllowSubscription {
			if _, hasWeight := p.Options["weight"]; !hasWeight {
				problems["options.weight"] = "weight options are required for subscription products"
			}
			if _, hasGrind := p.Options["grind"]; !hasGrind {
				problems["options.grind"] = "grind options are required for subscription products"
			}
		}
	}

	return problems
}

// ProductUpdateDTO represents the data needed to update a product
// Using pointers for all fields to differentiate between zero values and absence
type ProductUpdateDTO struct {
	Name              *string              `json:"name"`
	Description       *string              `json:"description"`
	ImageURL          *string              `json:"image_url"`
	Active            *bool                `json:"active"`
	StockLevel        *int                 `json:"stock_level"`
	Weight            *int                 `json:"weight"` // Weight in grams
	Origin            *string              `json:"origin"`
	RoastLevel        *string              `json:"roast_level"`
	FlavorNotes       *string              `json:"flavor_notes"`
	Options           *map[string][]string `json:"options"`            // Product options
	AllowSubscription *bool                `json:"allow_subscription"` // Flag to indicate if product can be subscribed to
}

// Valid performs validation on the ProductUpdateDTO fields
func (dto *ProductUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	// Name validation
	if dto.Name != nil {
		name := *dto.Name
		if len(name) < 2 {
			problems["name"] = "must be at least 2 characters long"
		} else if len(name) > 100 {
			problems["name"] = "must not exceed 100 characters"
		}
	}

	// Description validation
	if dto.Description != nil {
		if len(*dto.Description) > 1000 {
			problems["description"] = "must not exceed 1000 characters"
		}
	}

	// ImageURL validation
	if dto.ImageURL != nil && *dto.ImageURL != "" {
		_, err := url.ParseRequestURI(*dto.ImageURL)
		if err != nil {
			problems["image_url"] = "must be a valid URL"
		}
	}

	// StockLevel validation
	if dto.StockLevel != nil && *dto.StockLevel < 0 {
		problems["stock_level"] = "must not be negative"
	}

	// Weight validation
	if dto.Weight != nil && *dto.Weight < 1 {
		problems["weight"] = "must be at least 1 gram"
	}

	// Origin validation
	if dto.Origin != nil && len(*dto.Origin) > 100 {
		problems["origin"] = "must not exceed 100 characters"
	}

	// RoastLevel validation
	if dto.RoastLevel != nil {
		level := strings.ToLower(*dto.RoastLevel)
		if !validRoastLevels[level] {
			problems["roast_level"] = "must be one of: light, medium, dark"
		}
	}

	// FlavorNotes validation
	if dto.FlavorNotes != nil && len(*dto.FlavorNotes) > 255 {
		problems["flavor_notes"] = "must not exceed 255 characters"
	}

	// Validate options
	if dto.Options != nil {
		for key, values := range *dto.Options {
			if !validOptionKeys[key] {
				problems["options."+key] = "is not a valid option key"
			}
			if len(values) == 0 {
				problems["options."+key] = "must have at least one value"
			}
		}

		// Check that weight and grind options are defined if product allows subscription
		if dto.AllowSubscription != nil && *dto.AllowSubscription {
			options := *dto.Options
			if _, hasWeight := options["weight"]; !hasWeight {
				problems["options.weight"] = "weight options are required for subscription products"
			}
			if _, hasGrind := options["grind"]; !hasGrind {
				problems["options.grind"] = "grind options are required for subscription products"
			}
		}
	}

	return problems
}

// ProductResponseDTO represents the data returned to the client
type ProductResponseDTO struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	Description       string              `json:"description"`
	ImageURL          string              `json:"image_url"`
	Active            bool                `json:"active"`
	StockLevel        int                 `json:"stock_level"`
	Weight            int                 `json:"weight"`
	Origin            string              `json:"origin"`
	RoastLevel        string              `json:"roast_level"`
	FlavorNotes       string              `json:"flavor_notes"`
	Options           map[string][]string `json:"options"`
	AllowSubscription bool                `json:"allow_subscription"`
	CreatedAt         string              `json:"created_at"`
	UpdatedAt         string              `json:"updated_at"`
}