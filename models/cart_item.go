package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID             uuid.UUID  `json:"id"`
	CartID         uuid.UUID  `json:"cart_id"`
	ProductID      uuid.UUID  `json:"product_id"`
	PriceID        *uuid.UUID `json:"price_id,omitempty"`
	Quantity       int        `json:"quantity"`
	UnitPrice      int64      `json:"unit_price"` // In cents
	IsSubscription bool       `json:"is_subscription"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Options        *map[string]interface{} `json:"options,omitempty"` // Coffee-specific options
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	Product        *Product   `json:"product,omitempty"`
	Price          *Price     `json:"price,omitempty"`
}

// Coffee option type for better type safety
type CoffeeOptions struct {
	GrindType    string `json:"grind_type,omitempty"`    // whole_bean, coarse, medium, fine, etc.
	RoastLevel   string `json:"roast_level,omitempty"`   // light, medium, dark, etc.
	Size         string `json:"size,omitempty"`          // 12oz, 1lb, 2lb, etc.
	Decaf        bool   `json:"decaf,omitempty"`
	SingleOrigin bool   `json:"single_origin,omitempty"`
}