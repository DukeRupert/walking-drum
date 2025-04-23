// handlers/invoice_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type InvoiceHandler struct {
	invoiceRepo      repository.InvoiceRepository
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewInvoiceHandler(
	invoiceRepo repository.InvoiceRepository,
	userRepo repository.UserRepository,
	subscriptionRepo repository.SubscriptionRepository,
) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceRepo:      invoiceRepo,
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// Request/Response structs

type CreateInvoiceRequest struct {
	UserID          uuid.UUID              `json:"user_id"`
	SubscriptionID  *uuid.UUID             `json:"subscription_id,omitempty"`
	Status          string                 `json:"status"`
	AmountDue       int64                  `json:"amount_due"`
	AmountPaid      int64                  `json:"amount_paid"`
	Currency        string                 `json:"currency"`
	InvoicePDF      *string                `json:"invoice_pdf,omitempty"`
	StripeInvoiceID string                 `json:"stripe_invoice_id"`
	PaymentIntentID *string                `json:"payment_intent_id,omitempty"`
	PeriodStart     *string                `json:"period_start,omitempty"`
	PeriodEnd       *string                `json:"period_end,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateInvoiceRequest struct {
	Status          *string                 `json:"status,omitempty"`
	AmountDue       *int64                  `json:"amount_due,omitempty"`
	AmountPaid      *int64                  `json:"amount_paid,omitempty"`
	Currency        *string                 `json:"currency,omitempty"`
	InvoicePDF      *string                 `json:"invoice_pdf,omitempty"`
	PaymentIntentID *string                 `json:"payment_intent_id,omitempty"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
}

type InvoiceResponse struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	SubscriptionID  *uuid.UUID             `json:"subscription_id,omitempty"`
	Status          string                 `json:"status"`
	AmountDue       int64                  `json:"amount_due"`
	AmountPaid      int64                  `json:"amount_paid"`
	Currency        string                 `json:"currency"`
	InvoicePDF      *string                `json:"invoice_pdf,omitempty"`
	StripeInvoiceID string                 `json:"stripe_invoice_id"`
	PaymentIntentID *string                `json:"payment_intent_id,omitempty"`
	PeriodStart     *string                `json:"period_start,omitempty"`
	PeriodEnd       *string                `json:"period_end,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	User            *UserResponse          `json:"user,omitempty"`
	Subscription    *SubscriptionResponse  `json:"subscription,omitempty"`
}

// Helper functions

func (h *InvoiceHandler) modelToResponse(invoice *models.Invoice, includeRelations bool) (InvoiceResponse, error) {
	response := InvoiceResponse{
		ID:              invoice.ID,
		UserID:          invoice.UserID,
		Status:          string(invoice.Status),
		AmountDue:       invoice.AmountDue,
		AmountPaid:      invoice.AmountPaid,
		Currency:        invoice.Currency,
		StripeInvoiceID: invoice.StripeInvoiceID,
		CreatedAt:       invoice.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:       invoice.UpdatedAt.Format(http.TimeFormat),
	}

	if invoice.SubscriptionID != nil {
		response.SubscriptionID = invoice.SubscriptionID
	}

	if invoice.InvoicePDF != nil {
		response.InvoicePDF = invoice.InvoicePDF
	}

	if invoice.PaymentIntentID != nil {
		response.PaymentIntentID = invoice.PaymentIntentID
	}

	if invoice.PeriodStart != nil {
		formatted := invoice.PeriodStart.Format(http.TimeFormat)
		response.PeriodStart = &formatted
	}

	if invoice.PeriodEnd != nil {
		formatted := invoice.PeriodEnd.Format(http.TimeFormat)
		response.PeriodEnd = &formatted
	}

	if invoice.Metadata != nil {
		response.Metadata = *invoice.Metadata
	}

	// Include related entities if requested
	if includeRelations {
		// Include user details
		user, err := h.userRepo.GetByID(context.Background(), invoice.UserID)
		if err == nil {
			userHandler := NewUserHandler(h.userRepo)
			userResponse, err := userHandler.modelToResponse(user)
			if err == nil {
				response.User = &userResponse
			}
		}

		// Include subscription details if available
		if invoice.SubscriptionID != nil {
			subscription, err := h.subscriptionRepo.GetByID(context.Background(), *invoice.SubscriptionID)
			if err == nil {
				// Convert subscription model to response type
				subscriptionResponse := formatSubscriptionResponse(subscription)
				response.Subscription = &subscriptionResponse
			}
		}
	}

	return response, nil
}

// Helper function to format subscription model as response
func formatSubscriptionResponse(subscription *models.Subscription) SubscriptionResponse {
    response := SubscriptionResponse{
        ID:                 subscription.ID,
        UserID:             subscription.UserID,
        PriceID:            subscription.PriceID,
        Quantity:           subscription.Quantity,
        Status:             string(subscription.Status),
        CurrentPeriodStart: subscription.CurrentPeriodStart,
        CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
        CancelAtPeriodEnd:  subscription.CancelAtPeriodEnd,
        CreatedAt:          subscription.CreatedAt,
        UpdatedAt:          subscription.UpdatedAt,
    }
    
    // Add optional fields
    if subscription.CancelAt != nil {
        formatted := subscription.CancelAt
        response.CancelAt = formatted
    }
    
    if subscription.CanceledAt != nil {
        formatted := subscription.CanceledAt
        response.CanceledAt = formatted
    }
    
    if subscription.EndedAt != nil {
        formatted := subscription.EndedAt
        response.EndedAt = formatted
    }
    
    if subscription.TrialStart != nil {
        formatted := subscription.TrialStart
        response.TrialStart = formatted
    }
    
    if subscription.TrialEnd != nil {
        formatted := subscription.TrialEnd
        response.TrialEnd = formatted
    }
    
    return response
}

func (h *InvoiceHandler) parseTimeStr(timeStr string) (time.Time, error) {
	return time.Parse(http.TimeFormat, timeStr)
}

// Handlers

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID == uuid.Nil {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	if req.AmountDue < 0 {
		http.Error(w, "Amount due cannot be negative", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		http.Error(w, "Currency is required", http.StatusBadRequest)
		return
	}

	if req.StripeInvoiceID == "" {
		http.Error(w, "Stripe invoice ID is required", http.StatusBadRequest)
		return
	}

	// Verify user exists
	_, err := h.userRepo.GetByID(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	// Verify subscription exists if provided
	if req.SubscriptionID != nil {
		_, err := h.subscriptionRepo.GetByID(r.Context(), *req.SubscriptionID)
		if err != nil {
			if errors.Is(err, repository.ErrSubscriptionNotFound) {
				http.Error(w, "Subscription not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify subscription", http.StatusInternalServerError)
			return
		}
	}

	// Parse dates if provided
	var periodStart, periodEnd *time.Time

	if req.PeriodStart != nil {
		parsed, err := h.parseTimeStr(*req.PeriodStart)
		if err != nil {
			http.Error(w, "Invalid date format for period start", http.StatusBadRequest)
			return
		}
		periodStart = &parsed
	}

	if req.PeriodEnd != nil {
		parsed, err := h.parseTimeStr(*req.PeriodEnd)
		if err != nil {
			http.Error(w, "Invalid date format for period end", http.StatusBadRequest)
			return
		}
		periodEnd = &parsed
	}

	// Create the invoice
	invoice := &models.Invoice{
		UserID:          req.UserID,
		SubscriptionID:  req.SubscriptionID,
		Status:          models.InvoiceStatus(req.Status),
		AmountDue:       req.AmountDue,
		AmountPaid:      req.AmountPaid,
		Currency:        req.Currency,
		InvoicePDF:      req.InvoicePDF,
		StripeInvoiceID: req.StripeInvoiceID,
		PaymentIntentID: req.PaymentIntentID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
	}

	if len(req.Metadata) > 0 {
		invoice.Metadata = &req.Metadata
	}

	err = h.invoiceRepo.Create(r.Context(), invoice)
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceExists) {
			http.Error(w, "Invoice with this Stripe invoice ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(invoice, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	includeRelations := r.URL.Query().Get("include_relations") == "true"

	invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(invoice, includeRelations)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var req UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing invoice
	invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.Status != nil {
		invoice.Status = models.InvoiceStatus(*req.Status)
	}

	if req.AmountDue != nil {
		if *req.AmountDue < 0 {
			http.Error(w, "Amount due cannot be negative", http.StatusBadRequest)
			return
		}
		invoice.AmountDue = *req.AmountDue
	}

	if req.AmountPaid != nil {
		invoice.AmountPaid = *req.AmountPaid
	}

	if req.Currency != nil {
		invoice.Currency = *req.Currency
	}

	if req.InvoicePDF != nil {
		invoice.InvoicePDF = req.InvoicePDF
	}

	if req.PaymentIntentID != nil {
		invoice.PaymentIntentID = req.PaymentIntentID
	}

	if req.Metadata != nil {
		invoice.Metadata = req.Metadata
	}

	// Perform the update
	err = h.invoiceRepo.Update(r.Context(), invoice)
	if err != nil {
		http.Error(w, "Failed to update invoice", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(invoice, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InvoiceHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	// Get the existing invoice
	invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
		return
	}

	// Mark as paid
	invoice.Status = models.InvoiceStatusPaid
	invoice.AmountPaid = invoice.AmountDue

	// Update the invoice
	err = h.invoiceRepo.Update(r.Context(), invoice)
	if err != nil {
		http.Error(w, "Failed to mark invoice as paid", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(invoice, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	err = h.invoiceRepo.Delete(r.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, repository.ErrInvoiceNotFound) {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete invoice", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")
	userIDStr := r.URL.Query().Get("user_id")
	subscriptionIDStr := r.URL.Query().Get("subscription_id")
	includeRelations := r.URL.Query().Get("include_relations") == "true"

	limit := 10 // Default
	offset := 0 // Default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var invoices []*models.Invoice
	var err error

	// List by user ID
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Verify user exists
		_, err = h.userRepo.GetByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify user", http.StatusInternalServerError)
			return
		}

		invoices, err = h.invoiceRepo.ListByUserID(r.Context(), userID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to list invoices by user ID", http.StatusInternalServerError)
			return
		}
	} else if subscriptionIDStr != "" {
		// List by subscription ID
		subscriptionID, err := uuid.Parse(subscriptionIDStr)
		if err != nil {
			http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
			return
		}

		// Verify subscription exists
		_, err = h.subscriptionRepo.GetByID(r.Context(), subscriptionID)
		if err != nil {
			if errors.Is(err, repository.ErrSubscriptionNotFound) {
				http.Error(w, "Subscription not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify subscription", http.StatusInternalServerError)
			return
		}

		invoices, err = h.invoiceRepo.ListBySubscriptionID(r.Context(), subscriptionID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to list invoices by subscription ID", http.StatusInternalServerError)
			return
		}
	} else if status != "" {
		// List by status
		invoices, err = h.invoiceRepo.ListByStatus(r.Context(), models.InvoiceStatus(status), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list invoices by status", http.StatusInternalServerError)
			return
		}
	} else {
		// List all
		invoices, err = h.invoiceRepo.List(r.Context(), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list invoices", http.StatusInternalServerError)
			return
		}
	}

	// Convert to response objects
	var responses []InvoiceResponse
	for _, invoice := range invoices {
		response, err := h.modelToResponse(invoice, includeRelations)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
