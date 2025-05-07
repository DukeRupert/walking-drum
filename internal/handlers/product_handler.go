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
	// TODO: Implement product creation
	ctx := c.Request().Context()
	// 1. Bind request to ProductCreateDTO
	var req dto.ProductCreateDTO
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	// 2. Validate DTO
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	
	// 3. Call productService.Create
	product, err := h.productService.Create(ctx, &req); 
	if err != nil {
		h.logger.Error().Err(err).
			Str("product_id", product.ID.String()).
			Str("product_name", product.Name).
			Msg("Failed to create product")

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create product")
	}
	// 4. Return appropriate response
	h.logger.Info().
		Str("product_id", product.ID.String()).
		Str("product_name", product.Name).
		Msg("Product created successfully")

	return c.JSON(http.StatusCreated, product)
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
