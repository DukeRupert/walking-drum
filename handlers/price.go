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
	"github.com/rs/zerolog/log"

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

// Handlers

func (h *PriceHandler) CreatePrice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PriceHandler").
        Str("method", "CreatePrice").
        Logger()

    logger.Debug().Msg("Processing price creation request")

    var req CreatePriceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("productId", req.ProductID.String()).
        Int64("amount", req.Amount).
        Str("currency", req.Currency).
        Str("intervalType", req.IntervalType).
        Int("intervalCount", req.IntervalCount).
        Bool("isActive", req.IsActive).
        Msg("Received price creation request")

    // Basic validation
    if req.ProductID == uuid.Nil {
        logger.Error().Msg("Product ID is required")
        http.Error(w, "Product ID is required", http.StatusBadRequest)
        return
    }

    if req.Amount <= 0 {
        logger.Error().Int64("amount", req.Amount).Msg("Amount must be greater than zero")
        http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
        return
    }

    if req.Currency == "" {
        logger.Error().Msg("Currency is required")
        http.Error(w, "Currency is required", http.StatusBadRequest)
        return
    }

    if req.IntervalType == "" {
        logger.Error().Msg("Interval type is required")
        http.Error(w, "Interval type is required", http.StatusBadRequest)
        return
    }

    if req.IntervalCount <= 0 {
        logger.Error().Int("intervalCount", req.IntervalCount).Msg("Interval count must be greater than zero")
        http.Error(w, "Interval count must be greater than zero", http.StatusBadRequest)
        return
    }

    // Verify product exists
    logger.Debug().Str("productId", req.ProductID.String()).Msg("Verifying product exists")
    _, err := h.productRepo.GetByID(r.Context(), req.ProductID)
    if err != nil {
        if errors.Is(err, repository.ErrProductNotFound) {
            logger.Error().Err(err).Str("productId", req.ProductID.String()).Msg("Product not found")
            http.Error(w, "Product not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("productId", req.ProductID.String()).Msg("Failed to verify product")
        http.Error(w, "Failed to verify product", http.StatusInternalServerError)
        return
    }

    // Create the price
    logger.Debug().Str("productId", req.ProductID.String()).Msg("Creating price object")
    price := &models.Price{
        ProductID:     req.ProductID,
        Amount:        req.Amount,
        Currency:      req.Currency,
        IntervalType:  models.BillingInterval(req.IntervalType),
        IntervalCount: req.IntervalCount,
        IsActive:      req.IsActive,
    }

    if req.TrialPeriodDays != nil {
        trialDays := *req.TrialPeriodDays
        logger.Debug().Int("trialPeriodDays", trialDays).Msg("Setting trial period days")
        price.TrialPeriodDays = req.TrialPeriodDays
    }

    if req.Nickname != nil {
        nickname := *req.Nickname
        logger.Debug().Str("nickname", nickname).Msg("Setting price nickname")
        price.Nickname = req.Nickname
    }

    if len(req.Metadata) > 0 {
        logger.Debug().Interface("metadata", req.Metadata).Msg("Setting price metadata")
        price.Metadata = &req.Metadata
    }

    logger.Debug().
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Str("intervalType", string(price.IntervalType)).
        Int("intervalCount", price.IntervalCount).
        Bool("isActive", price.IsActive).
        Msg("Creating price in database")

    err = h.priceRepo.Create(r.Context(), price)
    if err != nil {
        if errors.Is(err, repository.ErrPriceExists) {
            logger.Error().
                Err(err).
                Str("productId", price.ProductID.String()).
                Msg("Price with this Stripe price ID already exists")
            http.Error(w, "Price with this Stripe price ID already exists", http.StatusConflict)
            return
        }
        logger.Error().
            Err(err).
            Str("productId", price.ProductID.String()).
            Msg("Failed to create price in database")
        http.Error(w, "Failed to create price", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Msg("Price created successfully, generating response")

    response, err := h.modelToResponse(price, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("priceId", price.ID.String()).
            Msg("Failed to generate price response")
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
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Str("intervalType", string(price.IntervalType)).
        Int("intervalCount", price.IntervalCount).
        Bool("isActive", price.IsActive).
        Msg("Price created successfully")
}

func (h *PriceHandler) GetPrice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PriceHandler").
        Str("method", "GetPrice").
        Logger()

    logger.Debug().Msg("Processing get price request")

    vars := mux.Vars(r)
    priceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("priceId", vars["id"]).Msg("Invalid price ID format")
        http.Error(w, "Invalid price ID", http.StatusBadRequest)
        return
    }

    includeProduct := r.URL.Query().Get("include_product") == "true"
    logger.Debug().
        Str("priceId", priceID.String()).
        Bool("includeProduct", includeProduct).
        Msg("Looking up price in database")

    price, err := h.priceRepo.GetByID(r.Context(), priceID)
    if err != nil {
        if errors.Is(err, repository.ErrPriceNotFound) {
            logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Price not found")
            http.Error(w, "Price not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Failed to get price from database")
        http.Error(w, "Failed to get price", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Bool("includeProduct", includeProduct).
        Msg("Price found, generating response")

    response, err := h.modelToResponse(price, includeProduct)
    if err != nil {
        logger.Error().
            Err(err).
            Str("priceId", price.ID.String()).
            Msg("Failed to generate price response")
        http.Error(w, "Failed to generate response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Str("priceId", price.ID.String()).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Bool("includeProduct", includeProduct).
        Msg("Price retrieved successfully")
}

func (h *PriceHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PriceHandler").
        Str("method", "UpdatePrice").
        Logger()

    logger.Debug().Msg("Processing price update request")

    vars := mux.Vars(r)
    priceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("priceId", vars["id"]).Msg("Invalid price ID format")
        http.Error(w, "Invalid price ID", http.StatusBadRequest)
        return
    }

    var req UpdatePriceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("priceId", priceID.String()).
        Interface("request", req).
        Msg("Received price update request")

    // Get the existing price
    logger.Debug().Str("priceId", priceID.String()).Msg("Looking up price in database")
    price, err := h.priceRepo.GetByID(r.Context(), priceID)
    if err != nil {
        if errors.Is(err, repository.ErrPriceNotFound) {
            logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Price not found")
            http.Error(w, "Price not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Failed to get price from database")
        http.Error(w, "Failed to get price", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("priceId", priceID.String()).
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Msg("Found price to update")

    // Update fields if provided
    if req.ProductID != nil {
        newProductID := *req.ProductID
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldProductId", price.ProductID.String()).
            Str("newProductId", newProductID.String()).
            Msg("Verifying new product exists")
            
        // Verify product exists
        _, err := h.productRepo.GetByID(r.Context(), newProductID)
        if err != nil {
            if errors.Is(err, repository.ErrProductNotFound) {
                logger.Error().Err(err).Str("productId", newProductID.String()).Msg("Product not found")
                http.Error(w, "Product not found", http.StatusNotFound)
                return
            }
            logger.Error().Err(err).Str("productId", newProductID.String()).Msg("Failed to verify product")
            http.Error(w, "Failed to verify product", http.StatusInternalServerError)
            return
        }
        
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldProductId", price.ProductID.String()).
            Str("newProductId", newProductID.String()).
            Msg("Updating price product ID")
        price.ProductID = newProductID
    }

    if req.Amount != nil {
        amount := *req.Amount
        if amount <= 0 {
            logger.Error().Int64("amount", amount).Msg("Amount must be greater than zero")
            http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
            return
        }
        logger.Debug().
            Str("priceId", priceID.String()).
            Int64("oldAmount", price.Amount).
            Int64("newAmount", amount).
            Msg("Updating price amount")
        price.Amount = amount
    }

    if req.Currency != nil {
        currency := *req.Currency
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldCurrency", price.Currency).
            Str("newCurrency", currency).
            Msg("Updating price currency")
        price.Currency = currency
    }

    if req.IntervalType != nil {
        intervalType := *req.IntervalType
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldIntervalType", string(price.IntervalType)).
            Str("newIntervalType", intervalType).
            Msg("Updating price interval type")
        price.IntervalType = models.BillingInterval(intervalType)
    }

    if req.IntervalCount != nil {
        intervalCount := *req.IntervalCount
        if intervalCount <= 0 {
            logger.Error().Int("intervalCount", intervalCount).Msg("Interval count must be greater than zero")
            http.Error(w, "Interval count must be greater than zero", http.StatusBadRequest)
            return
        }
        logger.Debug().
            Str("priceId", priceID.String()).
            Int("oldIntervalCount", price.IntervalCount).
            Int("newIntervalCount", intervalCount).
            Msg("Updating price interval count")
        price.IntervalCount = intervalCount
    }

    if req.TrialPeriodDays != nil {
        trialDays := *req.TrialPeriodDays
        oldTrialDays := "nil"
        if price.TrialPeriodDays != nil {
            oldTrialDays = fmt.Sprintf("%d", *price.TrialPeriodDays)
        }
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldTrialPeriodDays", oldTrialDays).
            Int("newTrialPeriodDays", trialDays).
            Msg("Updating price trial period days")
        price.TrialPeriodDays = req.TrialPeriodDays
    }

    if req.IsActive != nil {
        isActive := *req.IsActive
        logger.Debug().
            Str("priceId", priceID.String()).
            Bool("oldIsActive", price.IsActive).
            Bool("newIsActive", isActive).
            Msg("Updating price active status")
        price.IsActive = isActive
    }

    if req.Nickname != nil {
        nickname := *req.Nickname
        oldNickname := "nil"
        if price.Nickname != nil {
            oldNickname = *price.Nickname
        }
        logger.Debug().
            Str("priceId", priceID.String()).
            Str("oldNickname", oldNickname).
            Str("newNickname", nickname).
            Msg("Updating price nickname")
        price.Nickname = req.Nickname
    }

    if req.Metadata != nil {
        logger.Debug().
            Str("priceId", priceID.String()).
            Interface("newMetadata", *req.Metadata).
            Msg("Updating price metadata")
        price.Metadata = req.Metadata
    }

    // Perform the update
    logger.Debug().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Msg("Updating price in database")

    err = h.priceRepo.Update(r.Context(), price)
    if err != nil {
        if errors.Is(err, repository.ErrPriceExists) {
            logger.Error().
                Err(err).
                Str("priceId", price.ID.String()).
                Msg("Price with this Stripe price ID already exists")
            http.Error(w, "Price with this Stripe price ID already exists", http.StatusConflict)
            return
        }
        logger.Error().
            Err(err).
            Str("priceId", price.ID.String()).
            Msg("Failed to update price in database")
        http.Error(w, "Failed to update price", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Msg("Price updated successfully, generating response")

    response, err := h.modelToResponse(price, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("priceId", price.ID.String()).
            Msg("Failed to generate price response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Str("priceId", price.ID.String()).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("priceId", price.ID.String()).
        Str("productId", price.ProductID.String()).
        Int64("amount", price.Amount).
        Str("currency", price.Currency).
        Msg("Price updated successfully")
}

func (h *PriceHandler) DeletePrice(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PriceHandler").
        Str("method", "DeletePrice").
        Logger()

    logger.Debug().Msg("Processing price deletion request")

    vars := mux.Vars(r)
    priceID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("priceId", vars["id"]).Msg("Invalid price ID format")
        http.Error(w, "Invalid price ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("priceId", priceID.String()).Msg("Attempting to delete price from database")

    err = h.priceRepo.Delete(r.Context(), priceID)
    if err != nil {
        if errors.Is(err, repository.ErrPriceNotFound) {
            logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Price not found for deletion")
            http.Error(w, "Price not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("priceId", priceID.String()).Msg("Failed to delete price from database")
        http.Error(w, "Failed to delete price", http.StatusInternalServerError)
        return
    }

    logger.Info().Str("priceId", priceID.String()).Msg("Price deleted successfully")
    w.WriteHeader(http.StatusNoContent)
}

func (h *PriceHandler) ListPrices(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PriceHandler").
        Str("method", "ListPrices").
        Logger()

    logger.Debug().Msg("Processing list prices request")

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
        Bool("includeProduct", includeProduct).
        Str("productIDStr", productIDStr).
        Msg("Parsed query parameters")

    var prices []*models.Price
    var err error

    // If product ID is provided, list only prices for that product
    if productIDStr != "" {
        productID, err := uuid.Parse(productIDStr)
        if err != nil {
            logger.Error().Err(err).Str("productId", productIDStr).Msg("Invalid product ID format")
            http.Error(w, "Invalid product ID", http.StatusBadRequest)
            return
        }

        // Verify product exists
        logger.Debug().Str("productId", productID.String()).Msg("Verifying product exists")
        _, err = h.productRepo.GetByID(r.Context(), productID)
        if err != nil {
            if errors.Is(err, repository.ErrProductNotFound) {
                logger.Error().Err(err).Str("productId", productID.String()).Msg("Product not found")
                http.Error(w, "Product not found", http.StatusNotFound)
                return
            }
            logger.Error().Err(err).Str("productId", productID.String()).Msg("Failed to verify product")
            http.Error(w, "Failed to verify product", http.StatusInternalServerError)
            return
        }

        logger.Debug().
            Str("productId", productID.String()).
            Bool("activeOnly", activeOnly).
            Msg("Listing prices for specific product")

        prices, err = h.priceRepo.ListByProductID(r.Context(), productID, activeOnly)
        if err != nil {
            logger.Error().
                Err(err).
                Str("productId", productID.String()).
                Bool("activeOnly", activeOnly).
                Msg("Failed to list prices by product ID")
            http.Error(w, "Failed to list prices by product ID", http.StatusInternalServerError)
            return
        }
    } else {
        // List all prices
        if activeOnly {
            logger.Debug().
                Int("limit", limit).
                Int("offset", offset).
                Msg("Listing active prices")
            prices, err = h.priceRepo.ListActive(r.Context(), limit, offset)
        } else {
            logger.Debug().
                Int("limit", limit).
                Int("offset", offset).
                Msg("Listing all prices")
            prices, err = h.priceRepo.List(r.Context(), limit, offset)
        }

        if err != nil {
            logger.Error().
                Err(err).
                Int("limit", limit).
                Int("offset", offset).
                Bool("activeOnly", activeOnly).
                Msg("Failed to list prices from database")
            http.Error(w, "Failed to list prices", http.StatusInternalServerError)
            return
        }
    }

    logger.Debug().
        Int("count", len(prices)).
        Bool("includeProduct", includeProduct).
        Msg("Retrieved prices from database, generating responses")

    // Convert to response objects
    var responses []PriceResponse
    for _, price := range prices {
        response, err := h.modelToResponse(price, includeProduct)
        if err != nil {
            logger.Error().
                Err(err).
                Str("priceId", price.ID.String()).
                Str("productId", price.ProductID.String()).
                Msg("Failed to generate price response")
            http.Error(w, "Failed to generate response", http.StatusInternalServerError)
            return
        }
        responses = append(responses, response)
    }

    logger.Debug().Int("count", len(responses)).Msg("Generated price responses, encoding JSON")

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
        Bool("includeProduct", includeProduct).
        Bool("byProduct", productIDStr != "").
        Msg("Prices listed successfully")
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