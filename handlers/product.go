// handlers/product.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type ProductHandler struct {
	productRepo repository.ProductRepository
}

func NewProductHandler(productRepo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
}

// Request/Response structs

type CreateProductRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateProductRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	IsActive    *bool                   `json:"is_active,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
}

type ProductResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	IsActive        bool                   `json:"is_active"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	StripeProductID *string                `json:"stripe_product_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Helper functions
func (h *ProductHandler) modelToResponse(product *models.Product) (ProductResponse, error) {
    response := ProductResponse{
        ID:          product.ID,
        Name:        product.Name,
        Description: product.Description,
        IsActive:    product.IsActive,
        CreatedAt:   product.CreatedAt.Format(http.TimeFormat),
        UpdatedAt:   product.UpdatedAt.Format(http.TimeFormat),
    }

    if product.StripeProductID != nil {
        response.StripeProductID = product.StripeProductID
    }

    if product.Metadata != nil {
        response.Metadata = *product.Metadata
    }

    return response, nil
}

// Handlers
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Name == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)
		return
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if len(req.Metadata) > 0 {
		product.Metadata = &req.Metadata
	}

	err := h.productRepo.Create(r.Context(), product)
	if err != nil {
		if errors.Is(err, repository.ErrProductExists) {
			http.Error(w, "Product with this name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	response := h.modelToResponse(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	response := h.modelToResponse(product)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing product
	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = *req.Description
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if req.Metadata != nil {
		product.Metadata = req.Metadata
	}

	// Perform the update
	err = h.productRepo.Update(r.Context(), product)
	if err != nil {
		if errors.Is(err, repository.ErrProductExists) {
			http.Error(w, "Product with this name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	response := h.modelToResponse(product)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.productRepo.Delete(r.Context(), productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeOnly := r.URL.Query().Get("active") == "true"

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

	var products []*models.Product
	var err error

	if activeOnly {
		products, err = h.productRepo.ListActive(r.Context(), limit, offset)
	} else {
		products, err = h.productRepo.List(r.Context(), limit, offset)
	}

	if err != nil {
		http.Error(w, "Failed to list products", http.StatusInternalServerError)
		return
	}

	// Convert to response objects
	var responses []ProductResponse
	for _, product := range products {
		responses = append(responses, h.modelToResponse(product))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
