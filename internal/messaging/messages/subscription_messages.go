// internal/messaging/messages/subscription_messages.go
package messages

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionRenewalMessage represents a subscription renewal event
type SubscriptionRenewalMessage struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	ProductID      uuid.UUID `json:"product_id"`
	PriceID        uuid.UUID `json:"price_id"`
	Quantity       int       `json:"quantity"`
	RenewalDate    time.Time `json:"renewal_date"`
}

// SubscriptionStatusChangeMessage represents a subscription status change event
type SubscriptionStatusChangeMessage struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	OldStatus      string    `json:"old_status"`
	NewStatus      string    `json:"new_status"`
	ChangeDate     time.Time `json:"change_date"`
}

// EmailNotificationMessage represents an email notification event
type EmailNotificationMessage struct {
	Type       string                 `json:"type"`
	CustomerID uuid.UUID              `json:"customer_id"`
	Email      string                 `json:"email"`
	Subject    string                 `json:"subject"`
	Data       map[string]interface{} `json:"data"`
}

// StockUpdateMessage represents a stock update event
type StockUpdateMessage struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Operation string    `json:"operation"` // "decrement" or "increment"
}
