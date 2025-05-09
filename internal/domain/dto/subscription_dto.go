// internal/domain/dto/subscription_dto.go
package dto

import (
	"context"

	"github.com/google/uuid"
)

// SubscriptionCreateDTO represents the data needed to create a new subscription
type SubscriptionCreateDTO struct {
	CustomerID uuid.UUID `json:"customer_id"`
	ProductID  uuid.UUID `json:"product_id"`
	PriceID    uuid.UUID `json:"price_id"`
	AddressID  uuid.UUID `json:"address_id"`
	Quantity   int       `json:"quantity"`
}

// Valid validates the SubscriptionCreateDTO
func (s *SubscriptionCreateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if s.CustomerID == uuid.Nil {
		problems["customer_id"] = "customer ID is required"
	}

	if s.ProductID == uuid.Nil {
		problems["product_id"] = "product ID is required"
	}

	if s.PriceID == uuid.Nil {
		problems["price_id"] = "price ID is required"
	}

	if s.AddressID == uuid.Nil {
		problems["address_id"] = "address ID is required"
	}

	if s.Quantity <= 0 {
		problems["quantity"] = "quantity must be greater than 0"
	}

	return problems
}

// SubscriptionUpdateDTO represents the data that can be updated for a subscription
type SubscriptionUpdateDTO struct {
	Quantity int `json:"quantity,omitempty"`
}

// Valid validates the SubscriptionUpdateDTO
func (s *SubscriptionUpdateDTO) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if s.Quantity <= 0 {
		problems["quantity"] = "quantity must be greater than 0"
	}

	return problems
}

// SubscriptionResponseDTO represents the data returned to the client
type SubscriptionResponseDTO struct {
	ID                 string     `json:"id"`
	CustomerID         string     `json:"customer_id"`
	ProductID          string     `json:"product_id"`
	PriceID            string     `json:"price_id"`
	AddressID          string     `json:"address_id"`
	Quantity           int        `json:"quantity"`
	Status             string     `json:"status"`
	CurrentPeriodStart string     `json:"current_period_start"`
	CurrentPeriodEnd   string     `json:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CancelledAt        *string    `json:"cancelled_at,omitempty"`
	CreatedAt          string     `json:"created_at"`
	UpdatedAt          string     `json:"updated_at"`
}

// SubscriptionDetailDTO represents the detailed subscription data with related entities
type SubscriptionDetailDTO struct {
	ID                 string              `json:"id"`
	Customer           CustomerResponseDTO `json:"customer"`
	Product            ProductResponseDTO  `json:"product"`
	Price              PriceResponseDTO    `json:"price"`
	Address            AddressResponseDTO  `json:"address"`
	Quantity           int                 `json:"quantity"`
	Status             string              `json:"status"`
	CurrentPeriodStart string              `json:"current_period_start"`
	CurrentPeriodEnd   string              `json:"current_period_end"`
	CancelAtPeriodEnd  bool                `json:"cancel_at_period_end"`
	CancelledAt        *string             `json:"cancelled_at,omitempty"`
	CreatedAt          string              `json:"created_at"`
	UpdatedAt          string              `json:"updated_at"`
}

// ChangeProductRequest represents the request to change a subscription's product
type ChangeProductRequest struct {
	ProductID uuid.UUID `json:"product_id"`
}

// Valid validates the ChangeProductRequest
func (c *ChangeProductRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.ProductID == uuid.Nil {
		problems["product_id"] = "product ID is required"
	}

	return problems
}

// ChangePriceRequest represents the request to change a subscription's price
type ChangePriceRequest struct {
	PriceID uuid.UUID `json:"price_id"`
}

// Valid validates the ChangePriceRequest
func (c *ChangePriceRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.PriceID == uuid.Nil {
		problems["price_id"] = "price ID is required"
	}

	return problems
}

// ChangeQuantityRequest represents the request to change a subscription's quantity
type ChangeQuantityRequest struct {
	Quantity int `json:"quantity"`
}

// Valid validates the ChangeQuantityRequest
func (c *ChangeQuantityRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.Quantity <= 0 {
		problems["quantity"] = "quantity must be greater than 0"
	}

	return problems
}

// ChangeAddressRequest represents the request to change a subscription's shipping address
type ChangeAddressRequest struct {
	AddressID uuid.UUID `json:"address_id"`
}

// Valid validates the ChangeAddressRequest
func (c *ChangeAddressRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if c.AddressID == uuid.Nil {
		problems["address_id"] = "address ID is required"
	}

	return problems
}

// CancelRequest represents the request to cancel a subscription
type CancelRequest struct {
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
}
