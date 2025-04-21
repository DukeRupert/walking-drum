// models/order_item.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID              uuid.UUID  `json:"id"`
	OrderID         uuid.UUID  `json:"order_id"`
	ProductID       uuid.UUID  `json:"product_id"`
	PriceID         *uuid.UUID `json:"price_id,omitempty"`
	SubscriptionID  *uuid.UUID `json:"subscription_id,omitempty"`
	Quantity        int        `json:"quantity"`
	UnitPrice       int64      `json:"unit_price"` // In cents
	TotalPrice      int64      `json:"total_price"` // In cents
	IsSubscription  bool       `json:"is_subscription"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Options         *map[string]interface{} `json:"options,omitempty"` // Coffee-specific options
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	Product         *Product   `json:"product,omitempty"`
	Price           *Price     `json:"price,omitempty"`
	Subscription    *Subscription `json:"subscription,omitempty"`
}