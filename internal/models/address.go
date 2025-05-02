package models

import (
	"time"
)

// CustomerAddress represents a shipping address for a customer
type CustomerAddress struct {
	ID         int64     `json:"id" db:"id"`
	CustomerID int64     `json:"customer_id" db:"customer_id"`
	Line1      string    `json:"line1" db:"line1"`
	Line2      string    `json:"line2" db:"line2"`
	City       string    `json:"city" db:"city"`
	State      string    `json:"state" db:"state"`
	PostalCode string    `json:"postal_code" db:"postal_code"`
	Country    string    `json:"country" db:"country"`
	IsDefault  bool      `json:"is_default" db:"is_default"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	
	// Relations (not stored in DB)
	Customer *Customer `json:"customer,omitempty" db:"-"`
}