// models/cart.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID *string    `json:"session_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	Items     []*CartItem `json:"items,omitempty"`
}