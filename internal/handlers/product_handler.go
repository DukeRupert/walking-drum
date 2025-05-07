// internal/handlers/product_handler.go
package handlers

import (
	"net/http"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/pkg/pagination"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productService services.ProductService
	logger         zerolog.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService services.ProductService, logger zerolog.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger.With().Str("component", "product_handler").Logger(),
	}
}

// Create handles POST /api/products
func (h *ProductHandler) Create(c echo.Context) error {
	// Define the request/response types inline, following Ryer's pattern
	type request = dto.ProductCreateDTO
	type response struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		Active      bool   `json:"active"`
		StockLevel  int    `json:"stock_level"`
		Weight      int    `json:"weight"`
		Origin      string `json:"origin"`
		RoastLevel  string `json:"roast_level"`
		FlavorNotes string `json:"flavor_notes"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	ctx := c.Request().Context()
	
	// 1. Bind request
	var req request
	if err := c.Bind(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	
	// 2. Validate using our validator interface (Mat Ryer style)
	if problems := req.Valid(ctx); len(problems) > 0 {
		h.logger.Error().
			Interface("problems", problems).
			Msg("Validation failed")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":    "Validation failed",
			"problems": problems,
		})
	}
	
	// 3. Call service
	product, err := h.productService.Create(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).
			Str("product_name", req.Name).
			Msg("Failed to create product")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create product")
	}
	
	// 4. Map to response
	resp := response{
		ID:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		ImageURL:    product.ImageURL,
		Active:      product.Active,
		StockLevel:  product.StockLevel,
		Weight:      product.Weight,
		Origin:      product.Origin,
		RoastLevel:  product.RoastLevel,
		FlavorNotes: product.FlavorNotes,
		CreatedAt:   product.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:   product.UpdatedAt.Format(http.TimeFormat),
	}
	
	h.logger.Info().
		Str("product_id", product.ID.String()).
		Str("product_name", product.Name).
		Msg("Product created successfully")
	
	return c.JSON(http.StatusCreated, resp)
}

// Get handles GET /api/products/:id
func (h *ProductHandler) Get(c echo.Context) error {
	// TODO: Implement product retrieval by ID
	// 1. Parse ID from URL
	// 2. Call productService.GetByID
	// 3. Return appropriate response
	return nil
}

// List handles GET /api/products
func (h *ProductHandler) List(c echo.Context) error {
	// TODO: Implement product listing with pagination
	ctx := c.Request().Context()

	// 1. Parse pagination parameters
	params := pagination.NewParams(c)

	// Parse additional filtering parameters
	includeInactive := false
	if c.QueryParam("include_inactive") == "true" {
		// todo: Only admins to see inactive products
		includeInactive = true
	}

	// 2. Call productService.List
	products, total, err := h.productService.List(ctx, params.Offset, params.PerPage, includeInactive)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve products")
	}

	// 3. Return paginated response
	meta := pagination.NewMeta(params, total)
	response := pagination.Response(products, meta)
	return c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/products/:id
func (h *ProductHandler) Update(c echo.Context) error {
	// TODO: Implement product update
	// 1. Parse ID from URL
	// 2. Bind request to ProductUpdateDTO
	// 3. Validate DTO
	// 4. Call productService.Update
	// 5. Return appropriate response
	return nil
}

// Delete handles DELETE /api/products/:id
func (h *ProductHandler) Delete(c echo.Context) error {
	// TODO: Implement product deletion
	// 1. Parse ID from URL
	// 2. Call productService.Delete
	// 3. Return appropriate response
	return nil
}

// UpdateStockLevel handles PATCH /api/products/:id/stock
func (h *ProductHandler) UpdateStockLevel(c echo.Context) error {
	// TODO: Implement stock level update
	// 1. Parse ID from URL
	// 2. Bind request body with quantity
	// 3. Validate quantity (non-negative)
	// 4. Call productService.UpdateStockLevel
	// 5. Return appropriate response
	return nil
}

// Helper functions for mapping between models and DTOs can be added here
