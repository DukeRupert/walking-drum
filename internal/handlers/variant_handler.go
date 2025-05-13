// internal/handlers/variant_handler.go
package handlers

import (
	"net/http"
	"strings"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// VariantHandler handles HTTP requests for product variants
type VariantHandler struct {
	variantService services.VariantService
	productService services.ProductService
	logger         zerolog.Logger
}

// NewVariantHandler creates a new variant handler
func NewVariantHandler(
	variantService services.VariantService,
	productService services.ProductService,
	logger *zerolog.Logger,
) *VariantHandler {
	return &VariantHandler{
		variantService: variantService,
		productService: productService,
		logger:         logger.With().Str("component", "variant_handler").Logger(),
	}
}

// ListByProduct handles GET /api/variants/product/:productId
func (h *VariantHandler) ListByProduct(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.ListByProduct").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling variant listing by product request")
	
	// 1. Parse product ID from URL
	productIDParam := c.Param("productId")
	
	h.logger.Debug().
		Str("handler", "VariantHandler.ListByProduct").
		Str("request_id", requestID).
		Str("product_id_param", productIDParam).
		Msg("Parsing product ID from URL")
	
	// Validate and convert the ID
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.ListByProduct").
			Str("request_id", requestID).
			Str("product_id_param", productIDParam).
			Err(err).
			Msg("Invalid product ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID format")
	}
	
	// 2. Check if product exists
	product, err := h.productService.GetByID(ctx, productID)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.ListByProduct").
			Str("request_id", requestID).
			Str("product_id", productID.String()).
			Err(err).
			Msg("Failed to get product")
			
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve product")
	}
	
	// 3. Get variants for the product
	variants, err := h.variantService.GetVariantsByProductID(ctx, productID)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.ListByProduct").
			Str("request_id", requestID).
			Str("product_id", productID.String()).
			Err(err).
			Msg("Failed to get variants for product")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variants for product")
	}
	
	// 4. Create the response structure with product details
	type variantResponse struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
	}
	
	response := struct {
		Product  struct {
			ID                string              `json:"id"`
			Name              string              `json:"name"`
			Description       string              `json:"description"`
			Options           map[string][]string `json:"options"`
			AllowSubscription bool                `json:"allow_subscription"`
		} `json:"product"`
		Variants []variantResponse `json:"variants"`
	}{
		Product: struct {
			ID                string              `json:"id"`
			Name              string              `json:"name"`
			Description       string              `json:"description"`
			Options           map[string][]string `json:"options"`
			AllowSubscription bool                `json:"allow_subscription"`
		}{
			ID:                product.ID.String(),
			Name:              product.Name,
			Description:       product.Description,
			Options:           product.Options,
			AllowSubscription: product.AllowSubscription,
		},
		Variants: make([]variantResponse, len(variants)),
	}
	
	// Map variants to response format
	for i, variant := range variants {
		response.Variants[i] = variantResponse{
			ID:            variant.ID.String(),
			ProductID:     variant.ProductID.String(),
			Weight:        variant.Weight,
			Grind:         variant.Grind,
			Active:        variant.Active,
			StockLevel:    variant.StockLevel,
			PriceID:       variant.PriceID.String(),
			StripePriceID: variant.StripePriceID,
		}
	}
	
	h.logger.Info().
		Str("handler", "VariantHandler.ListByProduct").
		Str("request_id", requestID).
		Str("product_id", productID.String()).
		Int("variants_count", len(variants)).
		Msg("Successfully retrieved variants for product")
	
	return c.JSON(http.StatusOK, response)
}

// Get handles GET /api/variants/:id
func (h *VariantHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.Get").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling get variant by ID request")
	
	// 1. Parse ID from URL
	idParam := c.Param("id")
	
	h.logger.Debug().
		Str("handler", "VariantHandler.Get").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing variant ID from URL")
	
	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Get").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid variant ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid variant ID format")
	}
	
	// 2. Call variantService.GetByID
	variant, err := h.variantService.GetByID(ctx, id)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Get").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Err(err).
			Msg("Failed to get variant")
			
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variant")
	}
	
	if variant == nil {
		h.logger.Debug().
			Str("handler", "VariantHandler.Get").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Msg("Variant not found")
		return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
	}
	
	// Get related product information
	product, err := h.productService.GetByID(ctx, variant.ProductID)
	if err != nil {
		h.logger.Warn().
			Str("handler", "VariantHandler.Get").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Str("product_id", variant.ProductID.String()).
			Err(err).
			Msg("Failed to get related product")
	}
	
	// 3. Create the response
	response := struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		ProductName   string `json:"product_name,omitempty"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
	}{
		ID:            variant.ID.String(),
		ProductID:     variant.ProductID.String(),
		Weight:        variant.Weight,
		Grind:         variant.Grind,
		Active:        variant.Active,
		StockLevel:    variant.StockLevel,
		PriceID:       variant.PriceID.String(),
		StripePriceID: variant.StripePriceID,
	}
	
	// Add product name if we have it
	if product != nil {
		response.ProductName = product.Name
	}
	
	h.logger.Info().
		Str("handler", "VariantHandler.Get").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Str("product_id", variant.ProductID.String()).
		Msg("Successfully retrieved variant")
	
	return c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/variants/:id
func (h *VariantHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling variant update request")
	
	// 1. Parse ID from URL
	idParam := c.Param("id")
	
	h.logger.Debug().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing variant ID from URL")
	
	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid variant ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid variant ID format")
	}
	
	// 2. Bind request to VariantUpdateDTO
	var variantDTO dto.VariantUpdateDTO
	if err := c.Bind(&variantDTO); err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	
	// 3. Call variantService.Update
	variant, err := h.variantService.Update(ctx, id, &variantDTO)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Err(err).
			Msg("Failed to update variant")
			
		// Check for specific error types
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
		}
		
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update variant")
	}
	
	// 4. Create the response
	response := struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
	}{
		ID:            variant.ID.String(),
		ProductID:     variant.ProductID.String(),
		Weight:        variant.Weight,
		Grind:         variant.Grind,
		Active:        variant.Active,
		StockLevel:    variant.StockLevel,
		PriceID:       variant.PriceID.String(),
		StripePriceID: variant.StripePriceID,
	}
	
	h.logger.Info().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Str("product_id", variant.ProductID.String()).
		Msg("Successfully updated variant")
	
	return c.JSON(http.StatusOK, response)
}

// UpdateStockLevel handles PATCH /api/variants/:id/stock
func (h *VariantHandler) UpdateStockLevel(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.UpdateStockLevel").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling variant stock update request")
	
	// 1. Parse ID from URL
	idParam := c.Param("id")
	
	h.logger.Debug().
		Str("handler", "VariantHandler.UpdateStockLevel").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing variant ID from URL")
	
	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid variant ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid variant ID format")
	}
	
	// 2. Bind request body
	var stockRequest struct {
		Quantity int `json:"quantity"`
	}
	
	if err := c.Bind(&stockRequest); err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	
	// Validate quantity
	if stockRequest.Quantity < 0 {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Int("quantity", stockRequest.Quantity).
			Msg("Invalid stock quantity")
		return echo.NewHTTPError(http.StatusBadRequest, "Stock quantity cannot be negative")
	}
	
	// 3. Call variantService.UpdateStockLevel
	err = h.variantService.UpdateStockLevel(ctx, id, stockRequest.Quantity)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Int("quantity", stockRequest.Quantity).
			Err(err).
			Msg("Failed to update variant stock level")
			
		// Check for specific error types
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
		}
		
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update variant stock level")
	}
	
	// 4. Return success response
	h.logger.Info().
		Str("handler", "VariantHandler.UpdateStockLevel").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("quantity", stockRequest.Quantity).
		Msg("Successfully updated variant stock level")
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":          id.String(),
		"stock_level": stockRequest.Quantity,
		"message":     "Stock level updated successfully",
	})
}

// List handles GET /api/variants
func (h *VariantHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling variant listing request")
	
	// 1. Parse pagination parameters
	params := pagination.NewParams(c)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("offset", params.Offset).
		Int("per_page", params.PerPage).
		Int("page", params.Page).
		Msg("Pagination parameters parsed")
	
	// Parse additional filtering parameters
	activeOnly := true
	if c.QueryParam("include_inactive") == "true" {
		activeOnly = false
		h.logger.Debug().
			Str("handler", "VariantHandler.List").
			Str("request_id", requestID).
			Bool("active_only", activeOnly).
			Msg("Including inactive variants in results")
	}
	
	// 2. Call variantService.List
	variants, total, err := h.variantService.List(ctx, params.Offset, params.PerPage, activeOnly)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.List").
			Str("request_id", requestID).
			Err(err).
			Int("offset", params.Offset).
			Int("per_page", params.PerPage).
			Bool("active_only", activeOnly).
			Msg("Failed to retrieve variants from service")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variants")
	}
	
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("variants_count", len(variants)).
		Int("total_count", total).
		Msg("Variants retrieved successfully")
	
	// 3. Create the response structure
	type variantResponse struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		ProductName   string `json:"product_name,omitempty"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
	}
	
	// Map variants to response format
	variantResponses := make([]variantResponse, len(variants))
	
	// Create a map of product IDs to fetch product details in one pass
	productIDs := make(map[uuid.UUID]bool)
	for _, variant := range variants {
		productIDs[variant.ProductID] = true
	}
	
	// Fetch product details for all products in one pass (in a real implementation)
	// For now, we'll just log and use basic details
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("unique_products", len(productIDs)).
		Msg("Would fetch product details for all products in one pass")
	
	for i, variant := range variants {
		variantResponses[i] = variantResponse{
			ID:            variant.ID.String(),
			ProductID:     variant.ProductID.String(),
			Weight:        variant.Weight,
			Grind:         variant.Grind,
			Active:        variant.Active,
			StockLevel:    variant.StockLevel,
			PriceID:       variant.PriceID.String(),
			StripePriceID: variant.StripePriceID,
		}
		
		// In a real implementation, we would have product details from the bulk fetch
		// For now, just fetch individually as needed
		product, err := h.productService.GetByID(ctx, variant.ProductID)
		if err == nil && product != nil {
			variantResponses[i].ProductName = product.Name
		}
	}
	
	// 4. Create and return paginated response
	meta := pagination.NewMeta(params, total)
	
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("current_page", meta.Page).
		Int("total_pages", meta.TotalPages).
		Int("per_page", meta.PerPage).
		Int("total", meta.Total).
		Msg("Pagination metadata generated")
	
	response := pagination.Response(variantResponses, meta)
	
	h.logger.Info().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("variants_count", len(variants)).
		Int("total_count", total).
		Int("page", params.Page).
		Int("per_page", params.PerPage).
		Int("status_code", http.StatusOK).
		Msg("Variant listing successfully returned")
	
	return c.JSON(http.StatusOK, response)
}