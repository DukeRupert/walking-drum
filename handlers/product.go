// handlers/product.go
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

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

// Handlers

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "ProductHandler").
		Str("method", "CreateProduct").
		Logger()

	logger.Debug().Msg("Processing product creation request")

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("name", req.Name).
		Str("description", truncateString(req.Description, 50)).
		Bool("isActive", req.IsActive).
		Msg("Received product creation request")

	// Basic validation
	if req.Name == "" {
		logger.Error().Msg("Product name is required")
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
		logger.Debug().Interface("metadata", req.Metadata).Msg("Product has metadata")
	}

	logger.Debug().
		Str("name", product.Name).
		Bool("isActive", product.IsActive).
		Msg("Creating product in database")

	err := h.productRepo.Create(r.Context(), product)
	if err != nil {
		if errors.Is(err, repository.ErrProductExists) {
			logger.Error().
				Err(err).
				Str("name", product.Name).
				Msg("Product with this name already exists")
			http.Error(w, "Product with this name already exists", http.StatusConflict)
			return
		}
		logger.Error().
			Err(err).
			Str("name", product.Name).
			Msg("Failed to create product in database")
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("productId", product.ID.String()).
		Str("name", product.Name).
		Msg("Product created successfully, generating response")

	response, err := h.modelToResponse(product)
	if err != nil {
		logger.Error().
			Err(err).
			Str("productId", product.ID.String()).
			Msg("Failed to generate product response")
		http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
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
		Str("productId", product.ID.String()).
		Str("name", product.Name).
		Msg("Product created successfully")
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "ProductHandler").
        Str("method", "GetProduct").
        Logger()

    logger.Debug().Msg("Processing get product request")

    vars := mux.Vars(r)
    productID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("productId", vars["id"]).Msg("Invalid product ID format")
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("productId", productID.String()).Msg("Looking up product in database")

    product, err := h.productRepo.GetByID(r.Context(), productID)
    if err != nil {
        if errors.Is(err, repository.ErrProductNotFound) {
            logger.Error().Err(err).Str("productId", productID.String()).Msg("Product not found")
            http.Error(w, "Product not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("productId", productID.String()).Msg("Failed to get product from database")
        http.Error(w, "Failed to get product", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("productId", product.ID.String()).
        Str("name", product.Name).
        Msg("Product found, generating response")

    response, err := h.modelToResponse(product)
    if err != nil {
        logger.Error().
            Err(err).
            Str("productId", product.ID.String()).
            Msg("Failed to generate product response")
        http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Str("productId", product.ID.String()).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("productId", product.ID.String()).
        Str("name", product.Name).
        Msg("Product retrieved successfully")
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "ProductHandler").
        Str("method", "UpdateProduct").
        Logger()

    logger.Debug().Msg("Processing product update request")

    vars := mux.Vars(r)
    productID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("productId", vars["id"]).Msg("Invalid product ID format")
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    var req UpdateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("productId", productID.String()).
        Interface("request", req).
        Msg("Received product update request")

    // Get the existing product
    logger.Debug().Str("productId", productID.String()).Msg("Looking up product in database")
    product, err := h.productRepo.GetByID(r.Context(), productID)
    if err != nil {
        if errors.Is(err, repository.ErrProductNotFound) {
            logger.Error().Err(err).Str("productId", productID.String()).Msg("Product not found")
            http.Error(w, "Product not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("productId", productID.String()).Msg("Failed to get product from database")
        http.Error(w, "Failed to get product", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("productId", productID.String()).
        Str("originalName", product.Name).
        Bool("originalIsActive", product.IsActive).
        Msg("Found product to update")

    // Update fields if provided
    if req.Name != nil {
        logger.Debug().Str("productId", productID.String()).Str("oldName", product.Name).Str("newName", *req.Name).Msg("Updating product name")
        product.Name = *req.Name
    }

    if req.Description != nil {
        logger.Debug().Str("productId", productID.String()).Msg("Updating product description")
        product.Description = *req.Description
    }

    if req.IsActive != nil {
        logger.Debug().Str("productId", productID.String()).Bool("oldIsActive", product.IsActive).Bool("newIsActive", *req.IsActive).Msg("Updating product active status")
        product.IsActive = *req.IsActive
    }

    if req.Metadata != nil {
        logger.Debug().Str("productId", productID.String()).Interface("metadata", req.Metadata).Msg("Updating product metadata")
        product.Metadata = req.Metadata
    }

    // Perform the update
    logger.Debug().
        Str("productId", product.ID.String()).
        Str("name", product.Name).
        Bool("isActive", product.IsActive).
        Msg("Updating product in database")

    err = h.productRepo.Update(r.Context(), product)
    if err != nil {
        if errors.Is(err, repository.ErrProductExists) {
            logger.Error().
                Err(err).
                Str("productId", product.ID.String()).
                Str("name", product.Name).
                Msg("Product with this name already exists")
            http.Error(w, "Product with this name already exists", http.StatusConflict)
            return
        }
        logger.Error().
            Err(err).
            Str("productId", product.ID.String()).
            Msg("Failed to update product in database")
        http.Error(w, "Failed to update product", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("productId", product.ID.String()).
        Str("name", product.Name).
        Msg("Product updated successfully, generating response")

    response, err := h.modelToResponse(product)
    if err != nil {
        logger.Error().
            Err(err).
            Str("productId", product.ID.String()).
            Msg("Failed to generate product response")
        http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Str("productId", product.ID.String()).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("productId", product.ID.String()).
        Str("name", product.Name).
        Msg("Product updated successfully")
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "ProductHandler").
        Str("method", "DeleteProduct").
        Logger()

    logger.Debug().Msg("Processing product deletion request")

    vars := mux.Vars(r)
    productID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("productId", vars["id"]).Msg("Invalid product ID format")
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("productId", productID.String()).Msg("Attempting to delete product from database")

    err = h.productRepo.Delete(r.Context(), productID)
    if err != nil {
        if errors.Is(err, repository.ErrProductNotFound) {
            logger.Error().Err(err).Str("productId", productID.String()).Msg("Product not found for deletion")
            http.Error(w, "Product not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("productId", productID.String()).Msg("Failed to delete product from database")
        http.Error(w, "Failed to delete product", http.StatusInternalServerError)
        return
    }

    logger.Info().Str("productId", productID.String()).Msg("Product deleted successfully")
    w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "ProductHandler").
        Str("method", "ListProducts").
        Logger()

    logger.Debug().Msg("Processing list products request")

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
        } else if err != nil {
            logger.Debug().Err(err).Str("limitStr", limitStr).Msg("Invalid limit parameter, using default")
        }
    }

    if offsetStr != "" {
        parsedOffset, err := strconv.Atoi(offsetStr)
        if err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        } else if err != nil {
            logger.Debug().Err(err).Str("offsetStr", offsetStr).Msg("Invalid offset parameter, using default")
        }
    }

    logger.Debug().
        Int("limit", limit).
        Int("offset", offset).
        Bool("activeOnly", activeOnly).
        Msg("Listing products with parameters")

    var products []*models.Product
    var err error

    if activeOnly {
        logger.Debug().Msg("Listing active products only")
        products, err = h.productRepo.ListActive(r.Context(), limit, offset)
    } else {
        logger.Debug().Msg("Listing all products")
        products, err = h.productRepo.List(r.Context(), limit, offset)
    }

    if err != nil {
        logger.Error().Err(err).Msg("Failed to list products from database")
        http.Error(w, "Failed to list products", http.StatusInternalServerError)
        return
    }

    logger.Debug().Int("count", len(products)).Msg("Retrieved products from database, generating response")

    // Convert to response objects
    var responses []ProductResponse
    for _, product := range products {
        response, err := h.modelToResponse(product)
        if err != nil {
            logger.Error().
                Err(err).
                Str("productId", product.ID.String()).
                Msg("Failed to generate product response")
            http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
            return
        }
        responses = append(responses, response)
    }

    logger.Debug().Int("count", len(responses)).Msg("Generated product responses, encoding JSON")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(responses); err != nil {
        logger.Error().Err(err).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Int("count", len(responses)).
        Int("limit", limit).
        Int("offset", offset).
        Bool("activeOnly", activeOnly).
        Msg("Products listed successfully")
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

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
