// handlers/price_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type PriceHandler struct {
	priceRepo   repository.PriceRepository
	productRepo repository.ProductRepository
}

func NewPriceHandler(priceRepo repository.PriceRepository, productRepo repository.ProductRepository) *PriceHandler {
	return &PriceHandler{
		priceRepo:   priceRepo,
		productRepo: productRepo,
	}
}

// Request/Response structs

type CreatePriceRequest struct {
	ProductID       uuid.UUID              `json:"product_id"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	IntervalType    string                 `json:"interval_type"`
	IntervalCount   int                    `json:"interval_count"`
	TrialPeriodDays *int                   `json:"trial_period_days,omitempty"`
	IsActive        bool                   `json:"is_active"`
	Nickname        *string                `json:"nickname,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpdatePriceRequest struct {
	ProductID       *uuid.UUID              `json:"product_id,omitempty"`
	Amount          *int64                  `json:"amount,omitempty"`
	Currency        *string                 `json:"currency,omitempty"`
	IntervalType    *string                 `json:"interval_type,omitempty"`
	IntervalCount   *int                    `json:"interval_count,omitempty"`
	TrialPeriodDays *int                    `json:"trial_period_days,omitempty"`
	IsActive        *bool                   `json:"is_active,omitempty"`
	Nickname        *string                 `json:"nickname,omitempty"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
}

type PriceResponse struct {
	ID              uuid.UUID              `json:"id"`
	ProductID       uuid.UUID              `json:"product_id"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	IntervalType    string                 `json:"interval_type"`
	IntervalCount   int                    `json:"interval_count"`
	TrialPeriodDays *int                   `json:"trial_period_days,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	StripePriceID   *string                `json:"stripe_price_id,omitempty"`
	IsActive        bool                   `json:"is_active"`
	Nickname        *string                `json:"nickname,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Product         *ProductResponse       `json:"product,omitempty"`
}

// Helper functions
func (h *PriceHandler) modelToResponse(price *models.Price, includeProduct bool) (PriceResponse, error) {
	response := PriceResponse{
		ID:            price.ID,
		ProductID:     price.ProductID,
		Amount:        price.Amount,
		Currency:      price.Currency,
		IntervalType:  string(price.IntervalType),
		IntervalCount: price.IntervalCount,
		CreatedAt:     price.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:     price.UpdatedAt.Format(http.TimeFormat),
		IsActive:      price.IsActive,
	}

	if price.TrialPeriodDays != nil {
		response.TrialPeriodDays = price.TrialPeriodDays
	}

	if price.StripePriceID != nil {
		response.StripePriceID = price.StripePriceID
	}

	if price.Nickname != nil {
		response.Nickname = price.Nickname
	}

	if price.Metadata != nil {
		response.Metadata = *price.Metadata
	}

	// Include product details if requested
	if includeProduct && price.ProductID != uuid.Nil {
		product, err := h.productRepo.GetByID(context.Background(), price.ProductID)
		if err != nil {
			return response, fmt.Errorf("failed to get product details: %w", err)
		}

		// Create a temporary ProductHandler
		productHandler := NewProductHandler(h.productRepo)
		productResponse, err := productHandler.modelToResponse(product)
		if err != nil {
			return response, fmt.Errorf("failed to convert product to response: %w", err)
		}

		response.Product = &productResponse
	}

	return response, nil
}

// Handlers

func (h *PriceHandler) CreatePrice(w http.ResponseWriter, r *http.Request) {
	var req CreatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.ProductID == uuid.Nil {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		http.Error(w, "Currency is required", http.StatusBadRequest)
		return
	}

	if req.IntervalType == "" {
		http.Error(w, "Interval type is required", http.StatusBadRequest)
		return
	}

	if req.IntervalCount <= 0 {
		http.Error(w, "Interval count must be greater than zero", http.StatusBadRequest)
		return
	}

	// Verify product exists
	_, err := h.productRepo.GetByID(r.Context(), req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify product", http.StatusInternalServerError)
		return
	}

	// Create the price
	price := &models.Price{
		ProductID:     req.ProductID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		IntervalType:  models.BillingInterval(req.IntervalType),
		IntervalCount: req.IntervalCount,
		IsActive:      req.IsActive,
	}

	if req.TrialPeriodDays != nil {
		price.TrialPeriodDays = req.TrialPeriodDays
	}

	if req.Nickname != nil {
		price.Nickname = req.Nickname
	}

	if len(req.Metadata) > 0 {
		price.Metadata = &req.Metadata
	}

	err = h.priceRepo.Create(r.Context(), price)
	if err != nil {
		if errors.Is(err, repository.ErrPriceExists) {
			http.Error(w, "Price with this Stripe price ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create price", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(price, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PriceHandler) GetPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	priceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid price ID", http.StatusBadRequest)
		return
	}

	includeProduct := r.URL.Query().Get("include_product") == "true"

	price, err := h.priceRepo.GetByID(r.Context(), priceID)
	if err != nil {
		if errors.Is(err, repository.ErrPriceNotFound) {
			http.Error(w, "Price not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get price", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(price, includeProduct)
	if err != nil {
		http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PriceHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	priceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid price ID", http.StatusBadRequest)
		return
	}

	var req UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing price
	price, err := h.priceRepo.GetByID(r.Context(), priceID)
	if err != nil {
		if errors.Is(err, repository.ErrPriceNotFound) {
			http.Error(w, "Price not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get price", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.ProductID != nil {
		// Verify product exists
		_, err := h.productRepo.GetByID(r.Context(), *req.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				http.Error(w, "Product not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify product", http.StatusInternalServerError)
			return
		}
		price.ProductID = *req.ProductID
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
			return
		}
		price.Amount = *req.Amount
	}

	if req.Currency != nil {
		price.Currency = *req.Currency
	}

	if req.IntervalType != nil {
		price.IntervalType = models.BillingInterval(*req.IntervalType)
	}

	if req.IntervalCount != nil {
		if *req.IntervalCount <= 0 {
			http.Error(w, "Interval count must be greater than zero", http.StatusBadRequest)
			return
		}
		price.IntervalCount = *req.IntervalCount
	}

	if req.TrialPeriodDays != nil {
		price.TrialPeriodDays = req.TrialPeriodDays
	}

	if req.IsActive != nil {
		price.IsActive = *req.IsActive
	}

	if req.Nickname != nil {
		price.Nickname = req.Nickname
	}

	if req.Metadata != nil {
		price.Metadata = req.Metadata
	}

	// Perform the update
	err = h.priceRepo.Update(r.Context(), price)
	if err != nil {
		if errors.Is(err, repository.ErrPriceExists) {
			http.Error(w, "Price with this Stripe price ID already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update price", http.StatusInternalServerError)
		return
	}

	response, err := h.modelToResponse(price, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PriceHandler) DeletePrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	priceID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid price ID", http.StatusBadRequest)
		return
	}

	err = h.priceRepo.Delete(r.Context(), priceID)
	if err != nil {
		if errors.Is(err, repository.ErrPriceNotFound) {
			http.Error(w, "Price not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete price", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PriceHandler) ListPrices(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeOnly := r.URL.Query().Get("active") == "true"
	includeProduct := r.URL.Query().Get("include_product") == "true"
	productIDStr := r.URL.Query().Get("product_id")

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

	var prices []*models.Price
	var err error

	// If product ID is provided, list only prices for that product
	if productIDStr != "" {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		// Verify product exists
		_, err = h.productRepo.GetByID(r.Context(), productID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				http.Error(w, "Product not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify product", http.StatusInternalServerError)
			return
		}

		prices, err = h.priceRepo.ListByProductID(r.Context(), productID, activeOnly)
		if err != nil {
			http.Error(w, "Failed to list prices by product ID", http.StatusInternalServerError)
			return
		}
	} else {
		// List all prices
		if activeOnly {
			prices, err = h.priceRepo.ListActive(r.Context(), limit, offset)
		} else {
			prices, err = h.priceRepo.List(r.Context(), limit, offset)
		}

		if err != nil {
			http.Error(w, "Failed to list prices", http.StatusInternalServerError)
			return
		}
	}

	// Convert to response objects
	var responses []PriceResponse
	for _, price := range prices {
		response, err := h.modelToResponse(price, includeProduct)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
