// internal/handlers/product_handler.go
package handlers

import (
	"net/http"
	"strings"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productService services.ProductService
	logger         zerolog.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService services.ProductService, logger *zerolog.Logger) *ProductHandler {
	sublogger := logger.With().Str("component", "product_handler").Logger()
	return &ProductHandler{
		productService: productService,
		logger:         sublogger,
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
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling product creation request")

	// 1. Bind request
	var req request
	if err := c.Bind(&req); err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Create").
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to bind request body")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("product_name", req.Name).
		Str("description", req.Description).
		Int("stock_level", req.StockLevel).
		Int("weight", req.Weight).
		Str("origin", req.Origin).
		Str("roast_level", req.RoastLevel).
		Bool("active", req.Active).
		Msg("Request body successfully bound")

	// 2. Validate using our validator interface (Mat Ryer style)
	if problems := req.Valid(ctx); len(problems) > 0 {
		h.logger.Error().
			Str("handler", "ProductHandler.Create").
			Str("request_id", requestID).
			Interface("problems", problems).
			Str("product_name", req.Name).
			Msg("Product validation failed")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":    "Validation failed",
			"problems": problems,
		})
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("product_name", req.Name).
		Msg("Product validation passed")

	// 3. Call service
	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Msg("Calling productService.Create")

	product, err := h.productService.Create(ctx, &req)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Create").
			Str("request_id", requestID).
			Err(err).
			Str("product_name", req.Name).
			Msg("Service layer failed to create product")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create product")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("product_id", product.ID.String()).
		Str("stripe_id", product.StripeID).
		Msg("Product successfully created by service layer")

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

	h.logger.Debug().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("response_status", http.StatusText(http.StatusCreated)).
		Msg("Preparing response")

	h.logger.Info().
		Str("handler", "ProductHandler.Create").
		Str("request_id", requestID).
		Str("product_id", product.ID.String()).
		Str("product_name", product.Name).
		Str("stripe_id", product.StripeID).
		Int("status_code", http.StatusCreated).
		Msg("Product created successfully")

	return c.JSON(http.StatusCreated, resp)
}

// Get handles GET /api/products/:id
func (h *ProductHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling get product by ID request")

	// 1. Parse ID from URL
	idParam := c.Param("id")

	h.logger.Debug().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing product ID from URL")

	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Get").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid product ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID format")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Msg("Product ID parsed successfully")

	// 2. Call productService.GetByID
	h.logger.Debug().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Msg("Calling productService.GetByID")

	product, err := h.productService.GetByID(ctx, id)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Get").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Err(err).
			Msg("Failed to retrieve product")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve product")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Str("product_name", product.Name).
		Msg("Product retrieved successfully")

	// 3. Return appropriate response
	resp := models.Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		ImageURL:    product.ImageURL,
		Active:      product.Active,
		StockLevel:  product.StockLevel,
		Weight:      product.Weight,
		Origin:      product.Origin,
		RoastLevel:  product.RoastLevel,
		FlavorNotes: product.FlavorNotes,
		StripeID:    product.StripeID,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	h.logger.Info().
		Str("handler", "ProductHandler.Get").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Str("product_name", product.Name).
		Int("status_code", http.StatusOK).
		Msg("Successfully returned product details")

	return c.JSON(http.StatusOK, resp)
}

// List handles GET /api/products
func (h *ProductHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling product listing request")

	// 1. Parse pagination parameters
	params := pagination.NewParams(c)

	h.logger.Debug().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Int("offset", params.Offset).
		Int("per_page", params.PerPage).
		Int("page", params.Page).
		Msg("Pagination parameters parsed")

	// Parse additional filtering parameters
	includeInactive := false
	if c.QueryParam("include_inactive") == "true" {
		// todo: Only admins to see inactive products
		includeInactive = true
		h.logger.Debug().
			Str("handler", "ProductHandler.List").
			Str("request_id", requestID).
			Bool("include_inactive", includeInactive).
			Msg("Including inactive products in results")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Int("offset", params.Offset).
		Int("per_page", params.PerPage).
		Bool("include_inactive", includeInactive).
		Msg("Calling productService.List")

	// 2. Call productService.List
	products, total, err := h.productService.List(ctx, params.Offset, params.PerPage, includeInactive)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.List").
			Str("request_id", requestID).
			Err(err).
			Int("offset", params.Offset).
			Int("per_page", params.PerPage).
			Bool("include_inactive", includeInactive).
			Msg("Failed to retrieve products from service")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve products")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Int("products_count", len(products)).
		Int("total_count", total).
		Msg("Products retrieved successfully")

	// 3. Return paginated response
	meta := pagination.NewMeta(params, total)

	h.logger.Debug().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Int("current_page", meta.Page).
		Int("total_pages", meta.TotalPages).
		Int("per_page", meta.PerPage).
		Int("total", meta.Total).
		Msg("Pagination metadata generated")

	response := pagination.Response(products, meta)

	// Log product IDs for easier debugging (limited to first 5 to avoid excessively long logs)
	if len(products) > 0 {
		logProducts := products
		if len(products) > 5 {
			logProducts = products[:5]
		}

		productIds := make([]string, len(logProducts))
		productNames := make([]string, len(logProducts))

		for i, p := range logProducts {
			productIds[i] = p.ID.String()
			productNames[i] = p.Name
		}

		h.logger.Debug().
			Str("handler", "ProductHandler.List").
			Str("request_id", requestID).
			Strs("product_ids", productIds).
			Strs("product_names", productNames).
			Int("total_results", len(products)).
			Msg("Sample of products being returned")
	}

	h.logger.Info().
		Str("handler", "ProductHandler.List").
		Str("request_id", requestID).
		Int("products_count", len(products)).
		Int("total_count", total).
		Int("page", params.Page).
		Int("per_page", params.PerPage).
		Int("status_code", http.StatusOK).
		Msg("Product listing successfully returned")

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
	ctx := c.Request().Context()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug().
		Str("handler", "ProductHandler.Delete").
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Request().URL.Path).
		Str("remote_addr", c.Request().RemoteAddr).
		Msg("Handling product deletion request")

	// 1. Parse ID from URL
	idParam := c.Param("id")

	h.logger.Debug().
		Str("handler", "ProductHandler.Delete").
		Str("request_id", requestID).
		Str("id_param", idParam).
		Msg("Parsing product ID from URL")

	// Validate and convert the ID
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Delete").
			Str("request_id", requestID).
			Str("id_param", idParam).
			Err(err).
			Msg("Invalid product ID format")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID format")
	}

	h.logger.Debug().
		Str("handler", "ProductHandler.Delete").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Msg("Product ID parsed successfully")

	// 2. Call productService.Delete
	h.logger.Debug().
		Str("handler", "ProductHandler.Delete").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Msg("Calling productService.Delete")

	err = h.productService.Delete(ctx, id)
	if err != nil {
		h.logger.Error().
			Str("handler", "ProductHandler.Delete").
			Str("request_id", requestID).
			Str("product_id", id.String()).
			Err(err).
			Msg("Failed to delete product")

		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}

		// For other errors, return internal server error
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete product")
	}

	h.logger.Info().
		Str("handler", "ProductHandler.Delete").
		Str("request_id", requestID).
		Str("product_id", id.String()).
		Int("status_code", http.StatusNoContent).
		Msg("Product deleted successfully")

	// 3. Return appropriate response (204 No Content for successful deletion)
	return c.NoContent(http.StatusNoContent)
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
