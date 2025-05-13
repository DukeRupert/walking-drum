// internal/domain/dto/subscription_dto.go
package dto

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dukerupert/walking-drum/internal/domain/models"
)

// SubscriptionCreateDTO represents the data needed to create a new subscription
type SubscriptionCreateDTO struct {
	CustomerID         uuid.UUID            `json:"customer_id"`
	ProductID          uuid.UUID            `json:"product_id"`
	PriceID            uuid.UUID            `json:"price_id"`
	AddressID          *uuid.UUID           `json:"address_id,omitempty"` // Optional
	StripeID           string               `json:"stripe_id"`
	StripeItemID       string               `json:"stripe_item_id"`
	Status             string               `json:"status"`
	Quantity           int                  `json:"quantity"`
	CurrentPeriodStart time.Time            `json:"current_period_start"`
	CurrentPeriodEnd   time.Time            `json:"current_period_end"`
	NextDeliveryDate   time.Time            `json:"next_delivery_date"`
	Metadata           map[string]string    `json:"metadata,omitempty"`
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

	// AddressID is optional, so no validation needed

	if s.StripeID == "" {
		problems["stripe_id"] = "Stripe subscription ID is required"
	}

	if s.StripeItemID == "" {
		problems["stripe_item_id"] = "Stripe subscription item ID is required"
	}

	if s.Status == "" {
		problems["status"] = "status is required"
	} else if !isValidSubscriptionStatus(s.Status) {
		problems["status"] = "invalid subscription status"
	}

	if s.Quantity <= 0 {
		problems["quantity"] = "quantity must be greater than 0"
	}

	if s.CurrentPeriodStart.IsZero() {
		problems["current_period_start"] = "current period start is required"
	}

	if s.CurrentPeriodEnd.IsZero() {
		problems["current_period_end"] = "current period end is required"
	}

	if s.NextDeliveryDate.IsZero() {
		problems["next_delivery_date"] = "next delivery date is required"
	}

	return problems
}

// ToSubscription converts the DTO to a Subscription model
func (s *SubscriptionCreateDTO) ToSubscription() *models.Subscription {
	return &models.Subscription{
		ID:                 uuid.New(),
		CustomerID:         s.CustomerID,
		ProductID:          s.ProductID,
		PriceID:            s.PriceID,
		AddressID:          s.AddressID,
		StripeID:           s.StripeID,
		StripeItemID:       s.StripeItemID,
		Status:             s.Status,
		Quantity:           s.Quantity,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		NextDeliveryDate:   s.NextDeliveryDate,
		Metadata:           s.Metadata,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// isValidSubscriptionStatus checks if the status is a valid subscription status
func isValidSubscriptionStatus(status string) bool {
	validStatuses := []string{
		"active",
		"past_due",
		"incomplete",
		"incomplete_expired",
		"trialing",
		"canceled",
		"unpaid",
		"paused",
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	return false
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
