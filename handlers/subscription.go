// handlers/subscription_handler.go
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

type SubscriptionHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	priceRepo        repository.PriceRepository
}

func NewSubscriptionHandler(
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	priceRepo repository.PriceRepository,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		priceRepo:        priceRepo,
	}
}

// Request/Response structs

type CreateSubscriptionRequest struct {
	UserID           uuid.UUID              `json:"user_id"`
	PriceID          uuid.UUID              `json:"price_id"`
	Quantity         int                    `json:"quantity"`
	CurrentPeriodEnd string                 `json:"current_period_end"`
	TrialEnd         *string                `json:"trial_end,omitempty"`
	CollectionMethod string                 `json:"collection_method"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateSubscriptionRequest struct {
	PriceID           *uuid.UUID              `json:"price_id,omitempty"`
	Quantity          *int                    `json:"quantity,omitempty"`
	Status            *string                 `json:"status,omitempty"`
	CancelAt          *string                 `json:"cancel_at,omitempty"`
	CancelAtPeriodEnd *bool                   `json:"cancel_at_period_end,omitempty"`
	Metadata          *map[string]interface{} `json:"metadata,omitempty"`
}

type SubscriptionResponse struct {
	ID                   uuid.UUID              `json:"id"`
	UserID               uuid.UUID              `json:"user_id"`
	PriceID              uuid.UUID              `json:"price_id"`
	Quantity             int                    `json:"quantity"`
	Status               string                 `json:"status"`
	CurrentPeriodStart   string                 `json:"current_period_start"`
	CurrentPeriodEnd     string                 `json:"current_period_end"`
	CancelAt             *string                `json:"cancel_at,omitempty"`
	CanceledAt           *string                `json:"canceled_at,omitempty"`
	EndedAt              *string                `json:"ended_at,omitempty"`
	TrialStart           *string                `json:"trial_start,omitempty"`
	TrialEnd             *string                `json:"trial_end,omitempty"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
	StripeSubscriptionID string                 `json:"stripe_subscription_id"`
	StripeCustomerID     string                 `json:"stripe_customer_id"`
	CollectionMethod     string                 `json:"collection_method"`
	CancelAtPeriodEnd    bool                   `json:"cancel_at_period_end"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	User                 *UserResponse          `json:"user,omitempty"`
	Price                *PriceResponse         `json:"price,omitempty"`
}

// Helper functions

func (h *SubscriptionHandler) modelToResponse(subscription *models.Subscription, includeRelations bool) (SubscriptionResponse, error) {
	response := SubscriptionResponse{
		ID:                   subscription.ID,
		UserID:               subscription.UserID,
		PriceID:              subscription.PriceID,
		Quantity:             subscription.Quantity,
		Status:               string(subscription.Status),
		CurrentPeriodStart:   subscription.CurrentPeriodStart.Format(http.TimeFormat),
		CurrentPeriodEnd:     subscription.CurrentPeriodEnd.Format(http.TimeFormat),
		StripeSubscriptionID: subscription.StripeSubscriptionID,
		StripeCustomerID:     subscription.StripeCustomerID,
		CollectionMethod:     subscription.CollectionMethod,
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		CreatedAt:            subscription.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:            subscription.UpdatedAt.Format(http.TimeFormat),
	}

	if subscription.CancelAt != nil {
		formatted := subscription.CancelAt.Format(http.TimeFormat)
		response.CancelAt = &formatted
	}

	if subscription.CanceledAt != nil {
		formatted := subscription.CanceledAt.Format(http.TimeFormat)
		response.CanceledAt = &formatted
	}

	if subscription.EndedAt != nil {
		formatted := subscription.EndedAt.Format(http.TimeFormat)
		response.EndedAt = &formatted
	}

	if subscription.TrialStart != nil {
		formatted := subscription.TrialStart.Format(http.TimeFormat)
		response.TrialStart = &formatted
	}

	if subscription.TrialEnd != nil {
		formatted := subscription.TrialEnd.Format(http.TimeFormat)
		response.TrialEnd = &formatted
	}

	if subscription.Metadata != nil {
		response.Metadata = *subscription.Metadata
	}

	// Include related entities if requested
	if includeRelations {
		// Include user details
		user, err := h.userRepo.GetByID(context.Background(), subscription.UserID)
		if err == nil {
			userHandler := NewUserHandler(h.userRepo)
			userResponse, err := userHandler.modelToResponse(user)
			if err == nil {
				response.User = &userResponse
			}
		}

		// Include price details
		price, err := h.priceRepo.GetByID(context.Background(), subscription.PriceID)
		if err == nil {
			priceHandler := NewPriceHandler(h.priceRepo, nil) // No need for product repo here
			priceResponse, err := priceHandler.modelToResponse(price, false)
			if err == nil {
				response.Price = &priceResponse
			}
		}
	}

	return response, nil
}

func (h *SubscriptionHandler) parseTimeStr(timeStr string) (time.Time, error) {
	return time.Parse(http.TimeFormat, timeStr)
}

// Handlers

func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
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
		http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
		return
	}

	if req.CurrentPeriodEnd == "" {
		http.Error(w, "Current period end is required", http.StatusBadRequest)
		return
	}

	// Verify user exists
	user, err := h.userRepo.GetByID(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	// Verify price exists
	_, err = h.priceRepo.GetByID(r.Context(), req.PriceID)
	if err != nil {
		if errors.Is(err, repository.ErrPriceNotFound) {
			http.Error(w, "Price not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify price", http.StatusInternalServerError)
		return
	}

	// Parse dates
	currentPeriodStart := time.Now()
	currentPeriodEnd, err := h.parseTimeStr(req.CurrentPeriodEnd)
	if err != nil {
		http.Error(w, "Invalid date format for current period end", http.StatusBadRequest)
		return
	}

	var trialEnd *time.Time
	if req.TrialEnd != nil {
		parsed, err := h.parseTimeStr(*req.TrialEnd)
		if err != nil {
			http.Error(w, "Invalid date format for trial end", http.StatusBadRequest)
			return
		}
		trialEnd = &parsed
	}

	// Create stripe customer ID if not exists
	stripeCustomerID := "cus_" + uuid.New().String()
	if user.StripeCustomerID != nil {
		stripeCustomerID = *user.StripeCustomerID
	} else {
		// In a real application, we would create a customer in Stripe here
		// and update the user with the customer ID
		customerID := stripeCustomerID
		user.StripeCustomerID = &customerID
		err = h.userRepo.Update(r.Context(), user)
		if err != nil {
			http.Error(w, "Failed to update user with Stripe customer ID", http.StatusInternalServerError)
			return
		}
	}

	// Create the subscription
	subscription := &models.Subscription{
		UserID:               req.UserID,
		PriceID:              req.PriceID,
		Quantity:             req.Quantity,
		Status:               models.SubscriptionStatusActive,
		CurrentPeriodStart:   currentPeriodStart,
		CurrentPeriodEnd:     currentPeriodEnd,
		TrialEnd:             trialEnd,
		StripeSubscriptionID: "sub_" + uuid.New().String(), // In a real app, this would come from Stripe
		StripeCustomerID:     stripeCustomerID,
		CollectionMethod:     req.CollectionMethod,
	}

	if len(req.Metadata) > 0 {
		subscription.Metadata = &req.Metadata
	}

	err = h.subscriptionRepo.Create(r.Context(), subscription)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionExists) {
			http.Error(w, "Subscription with this Stripe subscription ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(subscription, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	includeRelations := r.URL.Query().Get("include_relations") == "true"

	subscription, err := h.subscriptionRepo.GetByID(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(subscription, includeRelations)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing subscription
	subscription, err := h.subscriptionRepo.GetByID(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.PriceID != nil {
		// Verify price exists
		_, err := h.priceRepo.GetByID(r.Context(), *req.PriceID)
		if err != nil {
			if errors.Is(err, repository.ErrPriceNotFound) {
				http.Error(w, "Price not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify price", http.StatusInternalServerError)
			return
		}
		subscription.PriceID = *req.PriceID
	}

	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
			return
		}
		subscription.Quantity = *req.Quantity
	}

	if req.Status != nil {
		subscription.Status = models.SubscriptionStatus(*req.Status)
	}

	if req.CancelAt != nil {
		cancelAt, err := h.parseTimeStr(*req.CancelAt)
		if err != nil {
			http.Error(w, "Invalid date format for cancel at", http.StatusBadRequest)
			return
		}
		subscription.CancelAt = &cancelAt
	}

	if req.CancelAtPeriodEnd != nil {
		subscription.CancelAtPeriodEnd = *req.CancelAtPeriodEnd

		// If canceling at period end, set the cancel_at time
		if *req.CancelAtPeriodEnd {
			subscription.CancelAt = &subscription.CurrentPeriodEnd
		}
	}

	if req.Metadata != nil {
		subscription.Metadata = req.Metadata
	}

	// Perform the update
	err = h.subscriptionRepo.Update(r.Context(), subscription)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionExists) {
			http.Error(w, "Subscription with this Stripe subscription ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(subscription, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	// Get the existing subscription
	subscription, err := h.subscriptionRepo.GetByID(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	// Only active subscriptions can be canceled
	if subscription.Status != models.SubscriptionStatusActive {
		http.Error(w, "Only active subscriptions can be canceled", http.StatusBadRequest)
		return
	}

	// Check if we should cancel at period end
	atPeriodEnd := r.URL.Query().Get("at_period_end") == "true"

	now := time.Now()
	subscription.CanceledAt = &now

	if atPeriodEnd {
		subscription.CancelAtPeriodEnd = true
		subscription.CancelAt = &subscription.CurrentPeriodEnd
	} else {
		subscription.Status = models.SubscriptionStatusCanceled
		subscription.EndedAt = &now
	}

	// Perform the update
	err = h.subscriptionRepo.Update(r.Context(), subscription)
	if err != nil {
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(subscription, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	err = h.subscriptionRepo.Delete(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")
	userIDStr := r.URL.Query().Get("user_id")
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

	var subscriptions []*models.Subscription
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

		subscriptions, err = h.subscriptionRepo.ListByUserID(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to list subscriptions by user ID", http.StatusInternalServerError)
			return
		}
	} else if status != "" {
		// List by status
		subscriptions, err = h.subscriptionRepo.ListByStatus(r.Context(), models.SubscriptionStatus(status), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list subscriptions by status", http.StatusInternalServerError)
			return
		}
	} else {
		// List all
		subscriptions, err = h.subscriptionRepo.List(r.Context(), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list subscriptions", http.StatusInternalServerError)
			return
		}
	}

	// Convert to response objects
	var responses []SubscriptionResponse
	for _, subscription := range subscriptions {
		response, err := h.modelToResponse(subscription, includeRelations)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
