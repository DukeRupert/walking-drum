// models/webhook_event.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type WebhookEvent struct {
	ID           uuid.UUID  `json:"id"`
	StripeEventID string     `json:"stripe_event_id"`
	EventType    string     `json:"event_type"`
	ObjectID     *string    `json:"object_id,omitempty"`
	ObjectType   *string    `json:"object_type,omitempty"`
	Data         map[string]interface{} `json:"data"`
	CreatedAt    time.Time  `json:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	IsProcessed  bool       `json:"is_processed"`
}