// models/invoice.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft        InvoiceStatus = "draft"
	InvoiceStatusOpen         InvoiceStatus = "open"
	InvoiceStatusPaid         InvoiceStatus = "paid"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusVoid         InvoiceStatus = "void"
)

type Invoice struct {
	ID               uuid.UUID      `json:"id"`
	UserID           uuid.UUID      `json:"user_id"`
	SubscriptionID   *uuid.UUID     `json:"subscription_id,omitempty"`
	Status           InvoiceStatus  `json:"status"`
	AmountDue        int64          `json:"amount_due"` // Amount in cents
	AmountPaid       int64          `json:"amount_paid"` // Amount in cents
	Currency         string         `json:"currency"`
	InvoicePDF       *string        `json:"invoice_pdf,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	StripeInvoiceID  string         `json:"stripe_invoice_id"`
	PaymentIntentID  *string        `json:"payment_intent_id,omitempty"`
	PeriodStart      *time.Time     `json:"period_start,omitempty"`
	PeriodEnd        *time.Time     `json:"period_end,omitempty"`
	Metadata         *map[string]interface{} `json:"metadata,omitempty"`
	
	// Relations
	User             *User          `json:"user,omitempty"`
	Subscription     *Subscription  `json:"subscription,omitempty"`
}