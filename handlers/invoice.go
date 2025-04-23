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
	"github.com/rs/zerolog/log"

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

// Handlers

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "CreateInvoice").
        Logger()

    logger.Debug().Msg("Processing invoice creation request")

    var req CreateInvoiceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("userId", req.UserID.String()).
        Bool("hasSubscriptionId", req.SubscriptionID != nil).
        Str("status", req.Status).
        Int64("amountDue", req.AmountDue).
        Int64("amountPaid", req.AmountPaid).
        Str("currency", req.Currency).
        Bool("hasInvoicePDF", req.InvoicePDF != nil).
        Str("stripeInvoiceId", req.StripeInvoiceID).
        Bool("hasPaymentIntentId", req.PaymentIntentID != nil).
        Bool("hasPeriodStart", req.PeriodStart != nil).
        Bool("hasPeriodEnd", req.PeriodEnd != nil).
        Bool("hasMetadata", len(req.Metadata) > 0).
        Msg("Received invoice creation request")

    // Basic validation
    if req.UserID == uuid.Nil {
        logger.Error().Msg("User ID is required")
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }

    if req.Status == "" {
        logger.Error().Msg("Status is required")
        http.Error(w, "Status is required", http.StatusBadRequest)
        return
    }

    if req.AmountDue < 0 {
        logger.Error().Int64("amountDue", req.AmountDue).Msg("Amount due cannot be negative")
        http.Error(w, "Amount due cannot be negative", http.StatusBadRequest)
        return
    }

    if req.Currency == "" {
        logger.Error().Msg("Currency is required")
        http.Error(w, "Currency is required", http.StatusBadRequest)
        return
    }

    if req.StripeInvoiceID == "" {
        logger.Error().Msg("Stripe invoice ID is required")
        http.Error(w, "Stripe invoice ID is required", http.StatusBadRequest)
        return
    }

    // Verify user exists
    logger.Debug().Str("userId", req.UserID.String()).Msg("Verifying user exists")
    _, err := h.userRepo.GetByID(r.Context(), req.UserID)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            logger.Error().Err(err).Str("userId", req.UserID.String()).Msg("User not found")
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("userId", req.UserID.String()).Msg("Failed to verify user")
        http.Error(w, "Failed to verify user", http.StatusInternalServerError)
        return
    }

    // Verify subscription exists if provided
    if req.SubscriptionID != nil {
        logger.Debug().
            Str("userId", req.UserID.String()).
            Str("subscriptionId", req.SubscriptionID.String()).
            Msg("Verifying subscription exists")

        _, err := h.subscriptionRepo.GetByID(r.Context(), *req.SubscriptionID)
        if err != nil {
            if errors.Is(err, repository.ErrSubscriptionNotFound) {
                logger.Error().
                    Err(err).
                    Str("userId", req.UserID.String()).
                    Str("subscriptionId", req.SubscriptionID.String()).
                    Msg("Subscription not found")
                http.Error(w, "Subscription not found", http.StatusNotFound)
                return
            }
            logger.Error().
                Err(err).
                Str("userId", req.UserID.String()).
                Str("subscriptionId", req.SubscriptionID.String()).
                Msg("Failed to verify subscription")
            http.Error(w, "Failed to verify subscription", http.StatusInternalServerError)
            return
        }
    }

    // Parse dates if provided
    var periodStart, periodEnd *time.Time

    if req.PeriodStart != nil {
        logger.Debug().
            Str("userId", req.UserID.String()).
            Str("periodStartRaw", *req.PeriodStart).
            Msg("Parsing period start date")

        parsed, err := h.parseTimeStr(*req.PeriodStart)
        if err != nil {
            logger.Error().
                Err(err).
                Str("userId", req.UserID.String()).
                Str("periodStartRaw", *req.PeriodStart).
                Msg("Invalid date format for period start")
            http.Error(w, "Invalid date format for period start", http.StatusBadRequest)
            return
        }
        periodStart = &parsed
        
        logger.Debug().
            Str("userId", req.UserID.String()).
            Time("periodStart", parsed).
            Msg("Parsed period start date")
    }

    if req.PeriodEnd != nil {
        logger.Debug().
            Str("userId", req.UserID.String()).
            Str("periodEndRaw", *req.PeriodEnd).
            Msg("Parsing period end date")

        parsed, err := h.parseTimeStr(*req.PeriodEnd)
        if err != nil {
            logger.Error().
                Err(err).
                Str("userId", req.UserID.String()).
                Str("periodEndRaw", *req.PeriodEnd).
                Msg("Invalid date format for period end")
            http.Error(w, "Invalid date format for period end", http.StatusBadRequest)
            return
        }
        periodEnd = &parsed
        
        logger.Debug().
            Str("userId", req.UserID.String()).
            Time("periodEnd", parsed).
            Msg("Parsed period end date")
    }

    // Create the invoice
    logger.Debug().
        Str("userId", req.UserID.String()).
        Str("status", req.Status).
        Int64("amountDue", req.AmountDue).
        Str("currency", req.Currency).
        Msg("Creating invoice object")

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
        logger.Debug().
            Str("userId", req.UserID.String()).
            Interface("metadata", req.Metadata).
            Msg("Setting invoice metadata")
        invoice.Metadata = &req.Metadata
    }

    logger.Debug().
        Str("userId", req.UserID.String()).
        Str("stripeInvoiceId", req.StripeInvoiceID).
        Msg("Creating invoice in database")

    err = h.invoiceRepo.Create(r.Context(), invoice)
    if err != nil {
        if errors.Is(err, repository.ErrInvoiceExists) {
            logger.Error().
                Err(err).
                Str("userId", req.UserID.String()).
                Str("stripeInvoiceId", req.StripeInvoiceID).
                Msg("Invoice with this Stripe invoice ID already exists")
            http.Error(w, "Invoice with this Stripe invoice ID already exists", http.StatusConflict)
            return
        }
        logger.Error().
            Err(err).
            Str("userId", req.UserID.String()).
            Str("stripeInvoiceId", req.StripeInvoiceID).
            Msg("Failed to create invoice")
        http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoice.ID.String()).
        Str("userId", req.UserID.String()).
        Str("stripeInvoiceId", req.StripeInvoiceID).
        Msg("Invoice created successfully, generating response")

    response, err := h.modelToResponse(invoice, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoice.ID.String()).
            Str("userId", req.UserID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoice.ID.String()).
        Str("userId", req.UserID.String()).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoice.ID.String()).
            Str("userId", req.UserID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("invoiceId", invoice.ID.String()).
        Str("userId", req.UserID.String()).
        Bool("hasSubscriptionId", req.SubscriptionID != nil).
        Str("status", req.Status).
        Int64("amountDue", req.AmountDue).
        Str("currency", req.Currency).
        Str("stripeInvoiceId", req.StripeInvoiceID).
        Msg("Invoice created successfully")
}

func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "GetInvoice").
        Logger()

    logger.Debug().Msg("Processing invoice retrieval request")

    vars := mux.Vars(r)
    invoiceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("invoiceIdRaw", vars["id"]).Msg("Invalid invoice ID")
        http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Parsed invoice ID")

    includeRelations := r.URL.Query().Get("include_relations") == "true"

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Bool("includeRelations", includeRelations).
        Msg("Retrieving invoice with query parameters")

    invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
    if err != nil {
        if errors.Is(err, repository.ErrInvoiceNotFound) {
            logger.Error().
                Err(err).
                Str("invoiceId", invoiceID.String()).
                Msg("Invoice not found")
            http.Error(w, "Invoice not found", http.StatusNotFound)
            return
        }
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to get invoice")
        http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("status", string(invoice.Status)).
        Int64("amountDue", invoice.AmountDue).
        Int64("amountPaid", invoice.AmountPaid).
        Str("currency", invoice.Currency).
        Str("stripeInvoiceId", invoice.StripeInvoiceID).
        Bool("hasSubscriptionId", invoice.SubscriptionID != nil).
        Msg("Invoice retrieved successfully, generating response")

    response, err := h.modelToResponse(invoice, includeRelations)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Bool("includeRelations", includeRelations).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("status", string(invoice.Status)).
        Bool("includeRelations", includeRelations).
        Msg("Invoice retrieved successfully")
}

func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "UpdateInvoice").
        Logger()

    logger.Debug().Msg("Processing invoice update request")

    vars := mux.Vars(r)
    invoiceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("invoiceIdRaw", vars["id"]).Msg("Invalid invoice ID")
        http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Parsed invoice ID")

    var req UpdateInvoiceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Str("invoiceId", invoiceID.String()).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Bool("hasStatus", req.Status != nil).
        Bool("hasAmountDue", req.AmountDue != nil).
        Bool("hasAmountPaid", req.AmountPaid != nil).
        Bool("hasCurrency", req.Currency != nil).
        Bool("hasInvoicePDF", req.InvoicePDF != nil).
        Bool("hasPaymentIntentID", req.PaymentIntentID != nil).
        Bool("hasMetadata", req.Metadata != nil && len(*req.Metadata) > 0).
        Msg("Received invoice update request")

    // Get the existing invoice
    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Retrieving invoice")
    invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
    if err != nil {
        if errors.Is(err, repository.ErrInvoiceNotFound) {
            logger.Error().Err(err).Str("invoiceId", invoiceID.String()).Msg("Invoice not found")
            http.Error(w, "Invoice not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("invoiceId", invoiceID.String()).Msg("Failed to get invoice")
        http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("currentStatus", string(invoice.Status)).
        Int64("currentAmountDue", invoice.AmountDue).
        Int64("currentAmountPaid", invoice.AmountPaid).
        Str("currentCurrency", invoice.Currency).
        Msg("Invoice retrieved, applying updates")

    // Update fields if provided
    if req.Status != nil {
        prevStatus := invoice.Status
        invoice.Status = models.InvoiceStatus(*req.Status)
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Str("prevStatus", string(prevStatus)).
            Str("newStatus", string(invoice.Status)).
            Msg("Updating invoice status")
    }

    if req.AmountDue != nil {
        if *req.AmountDue < 0 {
            logger.Error().
                Str("invoiceId", invoiceID.String()).
                Int64("amountDue", *req.AmountDue).
                Msg("Amount due cannot be negative")
            http.Error(w, "Amount due cannot be negative", http.StatusBadRequest)
            return
        }
        prevAmountDue := invoice.AmountDue
        invoice.AmountDue = *req.AmountDue
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Int64("prevAmountDue", prevAmountDue).
            Int64("newAmountDue", invoice.AmountDue).
            Msg("Updating invoice amount due")
    }

    if req.AmountPaid != nil {
        prevAmountPaid := invoice.AmountPaid
        invoice.AmountPaid = *req.AmountPaid
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Int64("prevAmountPaid", prevAmountPaid).
            Int64("newAmountPaid", invoice.AmountPaid).
            Msg("Updating invoice amount paid")
    }

    if req.Currency != nil {
        prevCurrency := invoice.Currency
        invoice.Currency = *req.Currency
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Str("prevCurrency", prevCurrency).
            Str("newCurrency", invoice.Currency).
            Msg("Updating invoice currency")
    }

    if req.InvoicePDF != nil {
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Str("invoicePdf", *req.InvoicePDF).
            Msg("Updating invoice PDF link")
        invoice.InvoicePDF = req.InvoicePDF
    }

    if req.PaymentIntentID != nil {
        hasCurrentPaymentIntentID := invoice.PaymentIntentID != nil && *invoice.PaymentIntentID != ""
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Bool("hasCurrentPaymentIntentID", hasCurrentPaymentIntentID).
            Str("newPaymentIntentID", *req.PaymentIntentID).
            Msg("Updating payment intent ID")
        invoice.PaymentIntentID = req.PaymentIntentID
    }

    if req.Metadata != nil {
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Interface("metadata", *req.Metadata).
            Msg("Updating invoice metadata")
        invoice.Metadata = req.Metadata
    }

    // Perform the update
    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("status", string(invoice.Status)).
        Int64("amountDue", invoice.AmountDue).
        Int64("amountPaid", invoice.AmountPaid).
        Str("currency", invoice.Currency).
        Msg("Updating invoice in database")

    err = h.invoiceRepo.Update(r.Context(), invoice)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to update invoice")
        http.Error(w, "Failed to update invoice", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("status", string(invoice.Status)).
        Msg("Invoice updated successfully, generating response")

    response, err := h.modelToResponse(invoice, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("status", string(invoice.Status)).
        Int64("amountDue", invoice.AmountDue).
        Int64("amountPaid", invoice.AmountPaid).
        Str("currency", invoice.Currency).
        Msg("Invoice updated successfully")
}

func (h *InvoiceHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "MarkAsPaid").
        Logger()

    logger.Debug().Msg("Processing mark invoice as paid request")

    vars := mux.Vars(r)
    invoiceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("invoiceIdRaw", vars["id"]).Msg("Invalid invoice ID")
        http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Parsed invoice ID")

    // Get the existing invoice
    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Retrieving invoice")
    invoice, err := h.invoiceRepo.GetByID(r.Context(), invoiceID)
    if err != nil {
        if errors.Is(err, repository.ErrInvoiceNotFound) {
            logger.Error().Err(err).Str("invoiceId", invoiceID.String()).Msg("Invoice not found")
            http.Error(w, "Invoice not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("invoiceId", invoiceID.String()).Msg("Failed to get invoice")
        http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("currentStatus", string(invoice.Status)).
        Int64("amountDue", invoice.AmountDue).
        Int64("currentAmountPaid", invoice.AmountPaid).
        Msg("Invoice retrieved, marking as paid")

    // Mark as paid
    prevStatus := invoice.Status
    prevAmountPaid := invoice.AmountPaid
    invoice.Status = models.InvoiceStatusPaid
    invoice.AmountPaid = invoice.AmountDue

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("prevStatus", string(prevStatus)).
        Str("newStatus", string(invoice.Status)).
        Int64("prevAmountPaid", prevAmountPaid).
        Int64("newAmountPaid", invoice.AmountPaid).
        Msg("Updating invoice status and amount paid")

    // Update the invoice
    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Msg("Updating invoice in database")

    err = h.invoiceRepo.Update(r.Context(), invoice)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to mark invoice as paid")
        http.Error(w, "Failed to mark invoice as paid", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("status", string(invoice.Status)).
        Int64("amountPaid", invoice.AmountPaid).
        Msg("Invoice marked as paid successfully, generating response")

    response, err := h.modelToResponse(invoice, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("invoiceId", invoiceID.String()).
        Str("userId", invoice.UserID.String()).
        Str("prevStatus", string(prevStatus)).
        Str("newStatus", string(invoice.Status)).
        Int64("amountDue", invoice.AmountDue).
        Int64("amountPaid", invoice.AmountPaid).
        Msg("Invoice marked as paid successfully")
}

func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "DeleteInvoice").
        Logger()

    logger.Debug().Msg("Processing invoice deletion request")

    vars := mux.Vars(r)
    invoiceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("invoiceIdRaw", vars["id"]).Msg("Invalid invoice ID")
        http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("invoiceId", invoiceID.String()).Msg("Parsed invoice ID")

    // Optional: Get the invoice before deletion to include more context in logs
    invoice, getErr := h.invoiceRepo.GetByID(r.Context(), invoiceID)
    if getErr == nil {
        logger.Debug().
            Str("invoiceId", invoiceID.String()).
            Str("userId", invoice.UserID.String()).
            Str("status", string(invoice.Status)).
            Str("stripeInvoiceId", invoice.StripeInvoiceID).
            Msg("Found invoice to delete")
    } else if !errors.Is(getErr, repository.ErrInvoiceNotFound) {
        logger.Warn().
            Err(getErr).
            Str("invoiceId", invoiceID.String()).
            Msg("Could not retrieve invoice details before deletion, proceeding anyway")
    }

    logger.Debug().
        Str("invoiceId", invoiceID.String()).
        Msg("Deleting invoice from database")

    err = h.invoiceRepo.Delete(r.Context(), invoiceID)
    if err != nil {
        if errors.Is(err, repository.ErrInvoiceNotFound) {
            logger.Error().
                Err(err).
                Str("invoiceId", invoiceID.String()).
                Msg("Invoice not found")
            http.Error(w, "Invoice not found", http.StatusNotFound)
            return
        }
        logger.Error().
            Err(err).
            Str("invoiceId", invoiceID.String()).
            Msg("Failed to delete invoice")
        http.Error(w, "Failed to delete invoice", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("invoiceId", invoiceID.String()).
        Msg("Invoice deleted successfully")

    w.WriteHeader(http.StatusNoContent)
}

func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "InvoiceHandler").
        Str("method", "ListInvoices").
        Logger()

    logger.Debug().Msg("Processing list invoices request")

    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    status := r.URL.Query().Get("status")
    userIDStr := r.URL.Query().Get("user_id")
    subscriptionIDStr := r.URL.Query().Get("subscription_id")
    includeRelations := r.URL.Query().Get("include_relations") == "true"

    logger.Debug().
        Str("limitRaw", limitStr).
        Str("offsetRaw", offsetStr).
        Str("status", status).
        Str("userIdRaw", userIDStr).
        Str("subscriptionIdRaw", subscriptionIDStr).
        Bool("includeRelations", includeRelations).
        Msg("Parsed query parameters")

    limit := 10 // Default
    offset := 0 // Default

    if limitStr != "" {
        parsedLimit, err := strconv.Atoi(limitStr)
        if err == nil && parsedLimit > 0 {
            limit = parsedLimit
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("limitRaw", limitStr).
                Msg("Invalid limit parameter, using default")
        }
    }

    if offsetStr != "" {
        parsedOffset, err := strconv.Atoi(offsetStr)
        if err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("offsetRaw", offsetStr).
                Msg("Invalid offset parameter, using default")
        }
    }

    logger.Debug().
        Int("limit", limit).
        Int("offset", offset).
        Msg("Using pagination parameters")

    var invoices []*models.Invoice
    var err error

    // List by user ID
    if userIDStr != "" {
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            logger.Error().Err(err).Str("userIdRaw", userIDStr).Msg("Invalid user ID")
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }

        logger.Debug().
            Str("userId", userID.String()).
            Int("limit", limit).
            Int("offset", offset).
            Msg("Listing invoices by user ID")

        // Verify user exists
        logger.Debug().Str("userId", userID.String()).Msg("Verifying user exists")
        _, err = h.userRepo.GetByID(r.Context(), userID)
        if err != nil {
            if errors.Is(err, repository.ErrUserNotFound) {
                logger.Error().Err(err).Str("userId", userID.String()).Msg("User not found")
                http.Error(w, "User not found", http.StatusNotFound)
                return
            }
            logger.Error().Err(err).Str("userId", userID.String()).Msg("Failed to verify user")
            http.Error(w, "Failed to verify user", http.StatusInternalServerError)
            return
        }

        invoices, err = h.invoiceRepo.ListByUserID(r.Context(), userID, limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Str("userId", userID.String()).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list invoices by user ID")
            http.Error(w, "Failed to list invoices by user ID", http.StatusInternalServerError)
            return
        }
        
        logger.Debug().
            Str("userId", userID.String()).
            Int("invoiceCount", len(invoices)).
            Msg("Retrieved invoices by user ID")
            
    } else if subscriptionIDStr != "" {
        // List by subscription ID
        subscriptionID, err := uuid.Parse(subscriptionIDStr)
        if err != nil {
            logger.Error().Err(err).Str("subscriptionIdRaw", subscriptionIDStr).Msg("Invalid subscription ID")
            http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
            return
        }

        logger.Debug().
            Str("subscriptionId", subscriptionID.String()).
            Int("limit", limit).
            Int("offset", offset).
            Msg("Listing invoices by subscription ID")

        // Verify subscription exists
        logger.Debug().Str("subscriptionId", subscriptionID.String()).Msg("Verifying subscription exists")
        _, err = h.subscriptionRepo.GetByID(r.Context(), subscriptionID)
        if err != nil {
            if errors.Is(err, repository.ErrSubscriptionNotFound) {
                logger.Error().Err(err).Str("subscriptionId", subscriptionID.String()).Msg("Subscription not found")
                http.Error(w, "Subscription not found", http.StatusNotFound)
                return
            }
            logger.Error().Err(err).Str("subscriptionId", subscriptionID.String()).Msg("Failed to verify subscription")
            http.Error(w, "Failed to verify subscription", http.StatusInternalServerError)
            return
        }

        invoices, err = h.invoiceRepo.ListBySubscriptionID(r.Context(), subscriptionID, limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Str("subscriptionId", subscriptionID.String()).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list invoices by subscription ID")
            http.Error(w, "Failed to list invoices by subscription ID", http.StatusInternalServerError)
            return
        }
        
        logger.Debug().
            Str("subscriptionId", subscriptionID.String()).
            Int("invoiceCount", len(invoices)).
            Msg("Retrieved invoices by subscription ID")
            
    } else if status != "" {
        // List by status
        logger.Debug().
            Str("status", status).
            Int("limit", limit).
            Int("offset", offset).
            Msg("Listing invoices by status")
            
        invoices, err = h.invoiceRepo.ListByStatus(r.Context(), models.InvoiceStatus(status), limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Str("status", status).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list invoices by status")
            http.Error(w, "Failed to list invoices by status", http.StatusInternalServerError)
            return
        }
        
        logger.Debug().
            Str("status", status).
            Int("invoiceCount", len(invoices)).
            Msg("Retrieved invoices by status")
            
    } else {
        // List all
        logger.Debug().
            Int("limit", limit).
            Int("offset", offset).
            Msg("Listing all invoices")
            
        invoices, err = h.invoiceRepo.List(r.Context(), limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list invoices")
            http.Error(w, "Failed to list invoices", http.StatusInternalServerError)
            return
        }
        
        logger.Debug().
            Int("invoiceCount", len(invoices)).
            Msg("Retrieved all invoices")
    }

    logger.Debug().
        Int("invoiceCount", len(invoices)).
        Bool("includeRelations", includeRelations).
        Msg("Generating response for invoices")

    // Convert to response objects
    var responses []InvoiceResponse
    for i, invoice := range invoices {
        logger.Debug().
            Int("index", i).
            Str("invoiceId", invoice.ID.String()).
            Str("status", string(invoice.Status)).
            Bool("includeRelations", includeRelations).
            Msg("Converting invoice to response")
            
        response, err := h.modelToResponse(invoice, includeRelations)
        if err != nil {
            logger.Error().
                Err(err).
                Str("invoiceId", invoice.ID.String()).
                Msg("Failed to generate response for invoice")
            http.Error(w, "Failed to generate response", http.StatusInternalServerError)
            return
        }
        responses = append(responses, response)
    }

    logger.Debug().
        Int("responseCount", len(responses)).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(responses); err != nil {
        logger.Error().
            Err(err).
            Int("responseCount", len(responses)).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    // Log based on which filter was used
    if userIDStr != "" {
        userID, _ := uuid.Parse(userIDStr)
        logger.Info().
            Str("userId", userID.String()).
            Int("invoiceCount", len(invoices)).
            Int("limit", limit).
            Int("offset", offset).
            Bool("includeRelations", includeRelations).
            Msg("User invoices listed successfully")
    } else if subscriptionIDStr != "" {
        subscriptionID, _ := uuid.Parse(subscriptionIDStr)
        logger.Info().
            Str("subscriptionId", subscriptionID.String()).
            Int("invoiceCount", len(invoices)).
            Int("limit", limit).
            Int("offset", offset).
            Bool("includeRelations", includeRelations).
            Msg("Subscription invoices listed successfully")
    } else if status != "" {
        logger.Info().
            Str("status", status).
            Int("invoiceCount", len(invoices)).
            Int("limit", limit).
            Int("offset", offset).
            Bool("includeRelations", includeRelations).
            Msg("Status-filtered invoices listed successfully")
    } else {
        logger.Info().
            Int("invoiceCount", len(invoices)).
            Int("limit", limit).
            Int("offset", offset).
            Bool("includeRelations", includeRelations).
            Msg("All invoices listed successfully")
    }
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
