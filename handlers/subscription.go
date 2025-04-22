// handlers/subscription_handler.go
package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/services"

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

type CreateSubscriptionRequest struct {
	CustomerID      string `json:"customer_id"`
	PriceID         string `json:"price_id"`
	Quantity        int64  `json:"quantity"`
	PaymentMethodID string `json:"payment_method_id,omitempty"`
	Description     string `json:"description,omitempty"`
	OrderID         string `json:"order_id,omitempty"`
}

type SubscriptionResponse struct {
    ID                 uuid.UUID  `json:"id"`
    UserID             uuid.UUID  `json:"user_id"`
    PriceID            uuid.UUID  `json:"price_id"`
    Quantity           int        `json:"quantity"`
    Status             string     `json:"status"`
    CurrentPeriodStart string     `json:"current_period_start"` // Formatted time string
    CurrentPeriodEnd   string     `json:"current_period_end"`   // Formatted time string
    CancelAt           *string    `json:"cancel_at,omitempty"`
    CanceledAt         *string    `json:"canceled_at,omitempty"`
    EndedAt            *string    `json:"ended_at,omitempty"`
    TrialStart         *string    `json:"trial_start,omitempty"`
    TrialEnd           *string    `json:"trial_end,omitempty"`
    CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
    StripeSubscriptionID string    `json:"stripe_subscription_id"`
    StripeCustomerID   string     `json:"stripe_customer_id"`
    CreatedAt          string     `json:"created_at"`
    UpdatedAt          string     `json:"updated_at"`
}

// CreateSubscription handles the creation of a new subscription
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req services.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID == uuid.Nil {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if req.PriceID == uuid.Nil {
		http.Error(w, "Price ID is required", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1 // Default to 1 if not specified or invalid
	}

	// Create subscription
	subscription, err := h.subscriptionService.CreateSubscription(req)
	if err != nil {
		http.Error(w, "Failed to create subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CancelSubscription handles canceling a subscription
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	// Cancel subscription
	subscription, err := h.subscriptionService.CancelSubscription(id)
	if err != nil {
		http.Error(w, "Failed to cancel subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response
	response, err := NewSubscriptionResponse(subscription)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleWebhook processes payment processor webhooks
func (h *SubscriptionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Handle the webhook event
	_, err = h.subscriptionService.ProcessWebhook(body, r.Header.Get("Stripe-Signature"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}

// Helper functions for response conversion
func NewSubscriptionResponse(subscription *models.Subscription) (map[string]interface{}, error) {
	response := map[string]interface{}{
		"id":                   subscription.ID,
		"user_id":              subscription.UserID,
		"price_id":             subscription.PriceID,
		"quantity":             subscription.Quantity,
		"status":               subscription.Status,
		"current_period_start": subscription.CurrentPeriodStart,
		"current_period_end":   subscription.CurrentPeriodEnd,
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
		"created_at":           subscription.CreatedAt,
		"updated_at":           subscription.UpdatedAt,
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

	return response, nil
}