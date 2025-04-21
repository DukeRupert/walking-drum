// models/order.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusCanceled   OrderStatus = "canceled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type Address struct {
	Name        string `json:"name"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2,omitempty"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type Order struct {
	ID               uuid.UUID       `json:"id"`
	UserID           *uuid.UUID      `json:"user_id,omitempty"`
	Status           OrderStatus     `json:"status"`
	TotalAmount      int64           `json:"total_amount"` // In cents
	Currency         string          `json:"currency"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	CompletedAt      *time.Time      `json:"completed_at,omitempty"`
	ShippingAddress  *Address        `json:"shipping_address,omitempty"`
	BillingAddress   *Address        `json:"billing_address,omitempty"`
	PaymentIntentID  *string         `json:"payment_intent_id,omitempty"`
	StripeCustomerID *string         `json:"stripe_customer_id,omitempty"`
	Metadata         *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	Items            []*OrderItem    `json:"items,omitempty"`
	User             *User           `json:"user,omitempty"`
}