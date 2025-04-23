// handlers/subscription_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/services"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// Request/Response structs

type SubscriptionResponse struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             uuid.UUID              `json:"user_id"`
	PriceID            uuid.UUID              `json:"price_id"`
	Quantity           int                    `json:"quantity"`
	Status             string                 `json:"status"`
	CollectionMethod   string                 `json:"collection_method"`
	Currency           string                 `json:"currency"`
	CustomerId         string                 `json:"customer_id"`
	CurrentPeriodStart time.Time              `json:"current_period_start"`
	CurrentPeriodEnd   time.Time              `json:"current_period_end"`
	CancelAt           *time.Time             `json:"cancel_at,omitempty"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	CanceledAt         *time.Time             `json:"canceled_at,omitempty"`
	EndedAt            *time.Time             `json:"ended_at,omitempty"`
	TrialStart         *time.Time             `json:"trial_start,omitempty"`
	TrialEnd           *time.Time             `json:"trial_end,omitempty"`
	ResumeAt           *time.Time             `json:"resume_at,omitempty"`
	LatestInvoiceID    string                 `json:"latest_invoice_id,omitempty"`
	PaymentMethodID    string                 `json:"payment_method_id,omitempty"`
	StripeID           string                 `json:"stripe_id"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

type CreateSubscriptionRequest struct {
	UserID          uuid.UUID         `json:"user_id"`
	PriceID         uuid.UUID         `json:"price_id"`
	Quantity        int64             `json:"quantity"`
	PaymentMethodID string            `json:"payment_method_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	OrderID         string            `json:"order_id,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type UpdateSubscriptionRequest struct {
	Quantity        int64             `json:"quantity,omitempty"`
	PriceID         uuid.UUID         `json:"price_id,omitempty"`
	PaymentMethodID string            `json:"payment_method_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type PauseSubscriptionRequest struct {
	ResumeAt *string `json:"resume_at,omitempty"` // ISO format date string, optional
}

// CreateSubscription handles the creation of a new subscription
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "CreateSubscription").
		Logger()

	logger.Debug().Msg("Processing subscription creation request")

	var req services.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log request data
	logger.Debug().
		Str("userID", req.UserID.String()).
		Str("priceID", req.PriceID.String()).
		Int("quantity", req.Quantity).
		Msg("Received subscription creation request")

	// Basic validation
	if req.UserID == uuid.Nil {
		logger.Error().Msg("User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if req.PriceID == uuid.Nil {
		logger.Error().Msg("Price ID is required")
		http.Error(w, "Price ID is required", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		logger.Debug().Int("quantity", req.Quantity).Msg("Invalid quantity, defaulting to 1")
		req.Quantity = 1 // Default to 1 if not specified or invalid
	}

	// Create subscription
	logger.Debug().
		Str("userID", req.UserID.String()).
		Str("priceID", req.PriceID.String()).
		Int("quantity", req.Quantity).
		Msg("Creating subscription")

	subscription, err := h.subscriptionService.CreateSubscription(req)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userID", req.UserID.String()).
			Str("priceID", req.PriceID.String()).
			Msg("Failed to create subscription")
		http.Error(w, "Failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", subscription.ID.String()).
		Str("userID", req.UserID.String()).
		Msg("Successfully created subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", subscription.ID.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", subscription.ID.String()).
		Str("userID", req.UserID.String()).
		Msg("Subscription created successfully")
}

// GetSubscription retrieves a subscription by ID
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "GetSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Msg("Retrieving subscription")

	subscription, err := h.subscriptionService.GetSubscription(id)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to retrieve subscription")
		http.Error(w, "Failed to retrieve subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Msg("Successfully retrieved subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Msg("Subscription returned successfully")
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (h *SubscriptionHandler) GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "GetUserSubscriptions").
		Logger()

	vars := mux.Vars(r)
	userIDStr := vars["userID"]

	logger.Debug().Str("userID", userIDStr).Msg("Processing request")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error().Err(err).Str("userID", userIDStr).Msg("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get pagination parameters from query string
	limit := 10 // Default limit
	offset := 0 // Default offset

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 {
			limit = limitVal
		} else if err != nil {
			logger.Debug().Err(err).Str("limit", limitStr).Msg("Invalid limit parameter")
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offsetVal, err := strconv.Atoi(offsetStr); err == nil && offsetVal >= 0 {
			offset = offsetVal
		} else if err != nil {
			logger.Debug().Err(err).Str("offset", offsetStr).Msg("Invalid offset parameter")
		}
	}

	// Get status filter from query string
	status := r.URL.Query().Get("status")

	logger.Debug().
		Str("userID", userID.String()).
		Str("status", status).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Retrieving subscriptions")

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(userID, status, limit, offset)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userID", userID.String()).
			Str("status", status).
			Msg("Failed to retrieve subscriptions")
		http.Error(w, "Failed to retrieve subscriptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Int("count", len(subscriptions)).
		Str("userID", userID.String()).
		Msg("Successfully retrieved subscriptions")

	// Convert to response
	var responses []map[string]interface{}
	for _, sub := range subscriptions {
		resp, err := NewSubscriptionResponse(sub)
		if err != nil {
			logger.Error().
				Err(err).
				Str("subscriptionID", sub.ID.String()).
				Msg("Failed to generate subscription response")
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("userID", userID.String()).
		Int("count", len(responses)).
		Msg("Subscriptions returned successfully")
}

// UpdateSubscription handles updating a subscription
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "UpdateSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing update request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Str("subscriptionID", id.String()).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log update details
	logger.Debug().
		Str("subscriptionID", id.String()).
		Int64("quantity", req.Quantity).
		Str("priceID", req.PriceID.String()).
		Str("paymentMethodID", string(req.PaymentMethodID)).
		Msg("Updating subscription")

	// Convert to service request
	updateReq := services.UpdateSubscriptionRequest{
		Quantity:        req.Quantity,
		PriceID:         req.PriceID,
		PaymentMethodID: req.PaymentMethodID,
		Description:     req.Description,
		Metadata:        req.Metadata,
	}

	subscription, err := h.subscriptionService.UpdateSubscription(id, updateReq)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to update subscription")
		http.Error(w, "Failed to update subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Msg("Successfully updated subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Msg("Subscription updated successfully")
}

// CancelSubscription handles canceling a subscription
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "CancelSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing cancellation request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	// Check if immediate cancellation is requested
	immediate := false
	if r.URL.Query().Get("immediate") == "true" {
		immediate = true
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Bool("immediate", immediate).
		Msg("Canceling subscription")

	// Cancel subscription
	subscription, err := h.subscriptionService.CancelSubscription(id, immediate)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Bool("immediate", immediate).
			Msg("Failed to cancel subscription")
		http.Error(w, "Failed to cancel subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Str("status", string(subscription.Status)).
		Msg("Successfully canceled subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Bool("immediate", immediate).
		Msg("Subscription canceled successfully")
}

// ReactivateSubscription reactivates a canceled subscription that was set to cancel at period end
func (h *SubscriptionHandler) ReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "ReactivateSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing reactivation request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Msg("Reactivating subscription")

	subscription, err := h.subscriptionService.ReactivateSubscription(id)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to reactivate subscription")
		http.Error(w, "Failed to reactivate subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Str("status", string(subscription.Status)).
		Msg("Successfully reactivated subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Msg("Subscription reactivated successfully")
}

// PauseSubscription pauses a subscription
func (h *SubscriptionHandler) PauseSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "PauseSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing pause request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var req PauseSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Str("subscriptionID", id.String()).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the resumeAt time if provided
	resumeAtStr := "not specified"
	if req.ResumeAt != nil {
		resumeAtStr = *req.ResumeAt
	}
	
	logger.Debug().
		Str("subscriptionID", id.String()).
		Str("resumeAt", resumeAtStr).
		Msg("Pausing subscription")

	subscription, err := h.subscriptionService.PauseSubscription(id, req.ResumeAt)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Str("resumeAt", resumeAtStr).
			Msg("Failed to pause subscription")
		http.Error(w, "Failed to pause subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Str("status", string(subscription.Status)).
		Msg("Successfully paused subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Str("resumeAt", resumeAtStr).
		Msg("Subscription paused successfully")
}

// ResumeSubscription resumes a paused subscription
func (h *SubscriptionHandler) ResumeSubscription(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "SubscriptionHandler").
		Str("method", "ResumeSubscription").
		Logger()

	vars := mux.Vars(r)
	idStr := vars["id"]

	logger.Debug().Str("subscriptionID", idStr).Msg("Processing resume request")

	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error().Err(err).Str("subscriptionID", idStr).Msg("Invalid subscription ID format")
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Msg("Resuming subscription")

	subscription, err := h.subscriptionService.ResumeSubscription(id)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to resume subscription")
		http.Error(w, "Failed to resume subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("subscriptionID", id.String()).
		Str("status", string(subscription.Status)).
		Msg("Successfully resumed subscription")

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		logger.Error().
			Err(err).
			Str("subscriptionID", id.String()).
			Msg("Failed to generate subscription response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().Err(err).Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Info().
		Str("subscriptionID", id.String()).
		Msg("Subscription resumed successfully")
}

// Helper functions for response conversion
func NewSubscriptionResponse(subscription *models.Subscription) (map[string]interface{}, error) {
	response := map[string]interface{}{
		"id":                     subscription.ID,
		"user_id":                subscription.UserID,
		"price_id":               subscription.PriceID,
		"quantity":               subscription.Quantity,
		"status":                 subscription.Status,
		"current_period_start":   subscription.CurrentPeriodStart,
		"current_period_end":     subscription.CurrentPeriodEnd,
		"cancel_at_period_end":   subscription.CancelAtPeriodEnd,
		"stripe_subscription_id": subscription.StripeSubscriptionID,
		"stripe_customer_id":     subscription.StripeCustomerID,
		"created_at":             subscription.CreatedAt,
		"updated_at":             subscription.UpdatedAt,
	}

	if subscription.CancelAt != nil {
		response["cancel_at"] = subscription.CancelAt
	}

	if subscription.CanceledAt != nil {
		response["canceled_at"] = subscription.CanceledAt
	}

	if subscription.EndedAt != nil {
		response["ended_at"] = subscription.EndedAt
	}

	if subscription.TrialStart != nil {
		response["trial_start"] = subscription.TrialStart
	}

	if subscription.TrialEnd != nil {
		response["trial_end"] = subscription.TrialEnd
	}

	// Add this block for the new field
	if subscription.ResumeAt != nil {
		response["resume_at"] = subscription.ResumeAt
	}

	return response, nil
}
