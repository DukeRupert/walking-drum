// internal/domain/dto/price_dto.go
package dto

import (
	"context"

	"github.com/google/uuid"
)

// PriceCreateDTO represents the data needed to create a new price
type PriceCreateDTO struct {
    ProductID     uuid.UUID `json:"product_id"`
    Name          string    `json:"name"`
    Amount        int64     `json:"amount"` // Price in cents
    Currency      string    `json:"currency"`
    Type          string    `json:"type"`   // one_time|recurring
    Interval      string    `json:"interval,omitempty"`       // week|month|year (required if type is recurring)
    IntervalCount int       `json:"interval_count,omitempty"` // Number of intervals between charges (required if type is recurring)
    Active        bool      `json:"active"`
}

// Valid validates the PriceCreateDTO
func (p *PriceCreateDTO) Valid(ctx context.Context) map[string]string {
    problems := make(map[string]string)

    if p.ProductID == uuid.Nil {
        problems["product_id"] = "product ID is required"
    }

    if p.Name == "" {
        problems["name"] = "name is required"
    }

    if p.Amount <= 0 {
        problems["amount"] = "amount must be greater than 0"
    }

    if p.Currency == "" {
        problems["currency"] = "currency is required"
    }

    if p.Type == "" {
        problems["type"] = "type is required"
    } else if p.Type != "one_time" && p.Type != "recurring" {
        problems["type"] = "type must be 'one_time' or 'recurring'"
    }

    // Check interval and interval_count only if type is recurring
    if p.Type == "recurring" {
        if p.Interval == "" {
            problems["interval"] = "interval is required for recurring prices"
        } else if p.Interval != "week" && p.Interval != "month" && p.Interval != "year" {
            problems["interval"] = "interval must be 'week', 'month', or 'year'"
        }

        if p.IntervalCount <= 0 {
            problems["interval_count"] = "interval count must be greater than 0 for recurring prices"
        }
    }

    return problems
}

// PriceUpdateDTO represents the data that can be updated for a price
type PriceUpdateDTO struct {
	Name          string `json:"name,omitempty"`
	Active        *bool  `json:"active,omitempty"`
	Type          string `json:"type,omitempty"`   // one_time|recurring
	Interval      string `json:"interval,omitempty"`       // week|month|year (for recurring prices)
	IntervalCount int    `json:"interval_count,omitempty"` // Number of intervals (for recurring prices)
}

// Valid validates the PriceUpdateDTO
func (p *PriceUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	// Validate type if provided
	if p.Type != "" && p.Type != "one_time" && p.Type != "recurring" {
		problems["type"] = "type must be 'one_time' or 'recurring'"
	}

	// If type is recurring and interval is provided, validate it
	if p.Type == "recurring" && p.Interval != "" {
		if p.Interval != "week" && p.Interval != "month" && p.Interval != "year" {
			problems["interval"] = "interval must be 'week', 'month', or 'year'"
		}
	}

	// If type is recurring and interval_count is provided, validate it
	if p.Type == "recurring" && p.IntervalCount != 0 {
		if p.IntervalCount <= 0 {
			problems["interval_count"] = "interval count must be greater than 0 for recurring prices"
		}
	}
	
	return problems
}

// PriceResponseDTO represents the data returned to the client
type PriceResponseDTO struct {
	ID            string `json:"id"`
	ProductID     string `json:"product_id"`
	Name          string `json:"name"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count"`
	Active        bool   `json:"active"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
