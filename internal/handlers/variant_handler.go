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

// VariantHandler handles HTTP requests for variants
type VariantHandler struct {
	variantService services.VariantService
	logger         zerolog.Logger
}

// NewVariantHandler creates a new variant handler
func NewVariantHandler(variantService services.VariantService, logger *zerolog.Logger) *VariantHandler {
	return &VariantHandler{
		variantService: variantService,
		logger:         logger.With().Str("component", "variant_handler").Logger(),
	}
}

// Create handles POST /api/variants
func (h *VariantHandler) Create(c echo.Context) error {
	// Define the request/response types inline
	type request = dto.VariantCreateDTO
	type response struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling variant creation request")

	// 1. Bind request to VariantCreateDTO
	var req request
	if err := c.Bind(&req); err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Create").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Str("product_id", req.ProductID.String()).
		Str("price_id", req.PriceID.String()).
		Str("weight", req.Weight).
		Str("grind", req.Grind).
		Bool("active", req.Active).
		Int("stock_level", req.StockLevel).
		Msg("Request body successfully bound")

	// 2. Validate DTO
	if problems := req.Valid(ctx); len(problems) > 0 {
		h.logger.Error().
			Str("handler", "VariantHandler.Create").
			Str("request_id", requestID).
			Interface("problems", problems).
			Msg("Variant validation failed")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":    "Validation failed",
			"problems": problems,
		})
	}

	h.logger.Debug().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Str("product_id", req.ProductID.String()).
		Msg("Variant validation passed")

	// 3. Call variantService.Create
	h.logger.Debug().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Msg("Calling variantService.Create")

	variant, err := h.variantService.Create(ctx, &req)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Create").
			Str("request_id", requestID).
			Err(err).
			Msg("Service layer failed to create variant")

		// Check for specific error types
		if strings.Contains(err.Error(), "already exists") {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}

		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		if strings.Contains(err.Error(), "validation") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create variant")
	}

	h.logger.Debug().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Str("variant_id", variant.ID.String()).
		Msg("Variant successfully created by service layer")

	// 4. Map to response
	resp := response{
		ID:            variant.ID.String(),
		ProductID:     variant.ProductID.String(),
		PriceID:       variant.PriceID.String(),
		StripePriceID: variant.StripePriceID,
		Weight:        variant.Weight,
		Grind:         variant.Grind,
		Active:        variant.Active,
		StockLevel:    variant.StockLevel,
		CreatedAt:     variant.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:     variant.UpdatedAt.Format(http.TimeFormat),
	}

	h.logger.Info().
		Str("handler", "VariantHandler.Create").
		Str("request_id", requestID).
		Str("variant_id", variant.ID.String()).
		Int("status_code", http.StatusCreated).
		Msg("Variant created successfully")

	return c.JSON(http.StatusCreated, resp)
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
	h.logger.Debug().
		Str("handler", "VariantHandler.Get").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Msg("Calling variantService.GetByID")

	variant, err := h.variantService.GetByID(ctx, id)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Get").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Err(err).
			Msg("Failed to retrieve variant")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variant")
	}

	// 3. Return response
	type response struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	resp := response{
		ID:            variant.ID.String(),
		ProductID:     variant.ProductID.String(),
		PriceID:       variant.PriceID.String(),
		StripePriceID: variant.StripePriceID,
		Weight:        variant.Weight,
		Grind:         variant.Grind,
		Active:        variant.Active,
		StockLevel:    variant.StockLevel,
		CreatedAt:     variant.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:     variant.UpdatedAt.Format(http.TimeFormat),
	}

	h.logger.Info().
		Str("handler", "VariantHandler.Get").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("status_code", http.StatusOK).
		Msg("Successfully returned variant details")

	return c.JSON(http.StatusOK, resp)
}

// GetByProductID handles GET /api/variants/product/:productId
func (h *VariantHandler) GetByProductID(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "VariantHandler.GetByProductID").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Msg("Handling get variants by product ID request")

	// 1. Parse product ID from URL
	productIDParam := c.Param("productId")

	h.logger.Debug().
		Str("handler", "VariantHandler.GetByProductID").
		Str("request_id", requestID).
		Str("product_id_param", productIDParam).
		Msg("Parsing product ID from URL")

	// Validate and convert the ID
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.GetByProductID").
			Str("request_id", requestID).
			Str("product_id_param", productIDParam).
			Err(err).
			Msg("Invalid product ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID format")
	}

	// 2. Call variantService.GetByProductID
	h.logger.Debug().
		Str("handler", "VariantHandler.GetByProductID").
		Str("request_id", requestID).
		Str("product_id", productID.String()).
		Msg("Calling variantService.GetByProductID")

	variants, err := h.variantService.GetByProductID(ctx, productID)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.GetByProductID").
			Str("request_id", requestID).
			Str("product_id", productID.String()).
			Err(err).
			Msg("Failed to retrieve variants")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variants")
	}

	// 3. Map to response
	type variantResponse struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	variantResponses := make([]variantResponse, len(variants))
	for i, variant := range variants {
		variantResponses[i] = variantResponse{
			ID:            variant.ID.String(),
			ProductID:     variant.ProductID.String(),
			PriceID:       variant.PriceID.String(),
			StripePriceID: variant.StripePriceID,
			Weight:        variant.Weight,
			Grind:         variant.Grind,
			Active:        variant.Active,
			StockLevel:    variant.StockLevel,
			CreatedAt:     variant.CreatedAt.Format(http.TimeFormat),
			UpdatedAt:     variant.UpdatedAt.Format(http.TimeFormat),
		}
	}

	h.logger.Info().
		Str("handler", "VariantHandler.GetByProductID").
		Str("request_id", requestID).
		Str("product_id", productID.String()).
		Int("variants_count", len(variants)).
		Int("status_code", http.StatusOK).
		Msg("Successfully returned variants for product")

	return c.JSON(http.StatusOK, variantResponses)
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
			Bool("activeOnly", activeOnly).
			Msg("Including inactive variants in results")
	}

	// 2. Call variantService.List
	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("offset", params.Offset).
		Int("per_page", params.PerPage).
		Bool("activeOnly", activeOnly).
		Msg("Calling variantService.List")

	variants, total, err := h.variantService.List(ctx, params.Offset, params.PerPage, activeOnly)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.List").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to retrieve variants from service")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variants")
	}

	h.logger.Debug().
		Str("handler", "VariantHandler.List").
		Str("request_id", requestID).
		Int("variants_count", len(variants)).
		Int("total_count", total).
		Msg("Variants retrieved successfully")

	// 3. Map to response
	type variantResponse struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		ProductName   string `json:"product_name"`
		ProductImage  string `json:"product_image,omitempty"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
		Origin        string `json:"origin,omitempty"`
		RoastLevel    string `json:"roast_level,omitempty"`
		FlavorNotes   string `json:"flavor_notes,omitempty"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	variantResponses := make([]variantResponse, len(variants))
	for i, variant := range variants {
		variantResponses[i] = variantResponse{
			ID:            variant.ID.String(),
			ProductID:     variant.ProductID.String(),
			ProductName:   variant.ProductName,
			ProductImage:  variant.ProductImage,
			PriceID:       variant.PriceID.String(),
			StripePriceID: variant.StripePriceID,
			Weight:        variant.Weight,
			Grind:         variant.Grind,
			Active:        variant.Active,
			StockLevel:    variant.StockLevel,
			Amount:        variant.Amount,
			Currency:      variant.Currency,
			Origin:        variant.Origin,
			RoastLevel:    variant.RoastLevel,
			FlavorNotes:   variant.FlavorNotes,
			CreatedAt:     variant.CreatedAt.Format(http.TimeFormat),
			UpdatedAt:     variant.UpdatedAt.Format(http.TimeFormat),
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

// Update handles PUT /api/variants/:id
func (h *VariantHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
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
	var req dto.VariantUpdateDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 3. Validate DTO
	if problems := req.Valid(ctx); len(problems) > 0 {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Interface("problems", problems).
			Msg("Variant validation failed")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":    "Validation failed",
			"problems": problems,
		})
	}

	// 4. Call variantService.Update
	h.logger.Debug().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Msg("Calling variantService.Update")

	variant, err := h.variantService.Update(ctx, id, &req)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Update").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Err(err).
			Msg("Service layer failed to update variant")

		// Check for specific error types
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		if strings.Contains(err.Error(), "already exists") {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}

		if strings.Contains(err.Error(), "validation") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update variant")
	}

	// 5. Map to response
	type response struct {
		ID            string `json:"id"`
		ProductID     string `json:"product_id"`
		PriceID       string `json:"price_id"`
		StripePriceID string `json:"stripe_price_id"`
		Weight        string `json:"weight"`
		Grind         string `json:"grind"`
		Active        bool   `json:"active"`
		StockLevel    int    `json:"stock_level"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	resp := response{
		ID:            variant.ID.String(),
		ProductID:     variant.ProductID.String(),
		PriceID:       variant.PriceID.String(),
		StripePriceID: variant.StripePriceID,
		Weight:        variant.Weight,
		Grind:         variant.Grind,
		Active:        variant.Active,
		StockLevel:    variant.StockLevel,
		CreatedAt:     variant.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:     variant.UpdatedAt.Format(http.TimeFormat),
	}

	h.logger.Info().
		Str("handler", "VariantHandler.Update").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("status_code", http.StatusOK).
		Msg("Variant updated successfully")

	return c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /api/variants/:id
func (h *VariantHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "VariantHandler.Delete").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Msg("Handling variant deletion request")

	// 1. Parse ID from URL
	idParam := c.Param("id")

	h.logger.Debug().
		Str("handler", "VariantHandler.Delete").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing variant ID from URL")

	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Delete").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid variant ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid variant ID format")
	}

	// 2. Call variantService.Delete
	h.logger.Debug().
		Str("handler", "VariantHandler.Delete").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Msg("Calling variantService.Delete")

	err = h.variantService.Delete(ctx, id)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.Delete").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Err(err).
			Msg("Failed to delete variant")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete variant")
	}

	h.logger.Info().
		Str("handler", "VariantHandler.Delete").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("status_code", http.StatusNoContent).
		Msg("Variant deleted successfully")

	// 3. Return appropriate response (204 No Content for successful deletion)
	return c.NoContent(http.StatusNoContent)
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
		Msg("Handling variant stock level update request")

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

	// 2. Bind request body with quantity
	type StockUpdateRequest struct {
		Quantity int `json:"quantity"`
	}

	var req StockUpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// 3. Validate quantity (non-negative)
	if req.Quantity < 0 {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Int("quantity", req.Quantity).
			Msg("Negative quantity provided")
		return echo.NewHTTPError(http.StatusBadRequest, "Stock level cannot be negative")
	}

	// 4. Call variantService.UpdateStockLevel
	h.logger.Debug().
		Str("handler", "VariantHandler.UpdateStockLevel").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("quantity", req.Quantity).
		Msg("Calling variantService.UpdateStockLevel")

	err = h.variantService.UpdateStockLevel(ctx, id, req.Quantity)
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.UpdateStockLevel").
			Str("request_id", requestID).
			Str("variant_id", id.String()).
			Int("quantity", req.Quantity).
			Err(err).
			Msg("Failed to update stock level")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Variant not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update stock level")
	}

	h.logger.Info().
		Str("handler", "VariantHandler.UpdateStockLevel").
		Str("request_id", requestID).
		Str("variant_id", id.String()).
		Int("quantity", req.Quantity).
		Int("status_code", http.StatusOK).
		Msg("Stock level updated successfully")

	// 5. Return appropriate response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":          id.String(),
		"stock_level": req.Quantity,
		"message":     "Stock level updated successfully",
	})
}

// GetOptions handles GET /api/variants/options
func (h *VariantHandler) GetOptions(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "VariantHandler.GetOptions").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Msg("Handling get variant options request")

	// Call variantService to get available options
	options, err := h.variantService.GetAvailableOptions()
	if err != nil {
		h.logger.Error().
			Str("handler", "VariantHandler.GetOptions").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to retrieve variant options")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve variant options")
	}

	h.logger.Info().
		Str("handler", "VariantHandler.GetOptions").
		Str("request_id", requestID).
		Int("weight_options_count", len(options.Weights)).
		Int("grind_options_count", len(options.Grinds)).
		Int("status_code", http.StatusOK).
		Msg("Variant options retrieved successfully")

	return c.JSON(http.StatusOK, options)
}
