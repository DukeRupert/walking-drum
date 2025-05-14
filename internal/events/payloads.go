// internal/events/payloads.go
package events

import (
    "time"
)

// ProductCreatedPayload represents the data in a product.created event
type ProductCreatedPayload struct {
    ProductID   string    `json:"product_id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    StockLevel  int       `json:"stock_level"`
    Origin      string    `json:"origin"`
    RoastLevel  string    `json:"roast_level"`
    Active      bool      `json:"active"`
    StripeID    string    `json:"stripe_id"`
    CreatedAt   time.Time `json:"created_at"`
}

// ProductStockUpdatedPayload represents the data in a product.stock_updated event
type ProductStockUpdatedPayload struct {
    ProductID     string    `json:"product_id"`
    Name          string    `json:"name"`
    OldStockLevel int       `json:"old_stock_level"`
    NewStockLevel int       `json:"new_stock_level"`
    IsLowStock    bool      `json:"is_low_stock"`
    UpdatedAt     time.Time `json:"updated_at"`
}