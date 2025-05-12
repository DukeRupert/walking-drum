// internal/handlers/price_handler.go
package handlers

import (
	"net/http"
	"strings"
	"time"

    "github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// PriceHandler handles HTTP requests for prices
type PriceHandler struct {
	priceService services.PriceService
	logger         zerolog.Logger
}

// NewPriceHandler creates a new price handler
func NewPriceHandler(priceService services.PriceService, logger *zerolog.Logger) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
		logger: logger.With().Str("component", "price_handler").Logger(),
	}
}

// Create handles POST /api/prices
func (h *PriceHandler) Create(c echo.Context) error {
    // Define the request/response types inline
    type request = dto.PriceCreateDTO
    type response struct {
        ID            string    `json:"id"`
        ProductID     string    `json:"product_id"`
        Name          string    `json:"name"`
        Amount        int64     `json:"amount"`
        Currency      string    `json:"currency"`
        Interval      string    `json:"interval"`
        IntervalCount int       `json:"interval_count"`
        Active        bool      `json:"active"`
        CreatedAt     time.Time `json:"created_at"`
        UpdatedAt     time.Time `json:"updated_at"`
    }

    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling price creation request")
    
    // 1. Bind request to PriceCreateDTO
    var req request
    if err := c.Bind(&req); err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Create").
            Str("request_id", requestID).
            Err(err).
            Msg("Failed to bind request body")
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("product_id", req.ProductID.String()).
        Str("name", req.Name).
        Int64("amount", req.Amount).
        Str("currency", req.Currency).
        Str("interval", req.Interval).
        Int("interval_count", req.IntervalCount).
        Bool("active", req.Active).
        Msg("Request body successfully bound")
    
    // 2. Validate DTO
    if problems := req.Valid(ctx); len(problems) > 0 {
        h.logger.Error().
            Str("handler", "PriceHandler.Create").
            Str("request_id", requestID).
            Interface("problems", problems).
            Str("product_id", req.ProductID.String()).
            Msg("Price validation failed")
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error":    "Validation failed",
            "problems": problems,
        })
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("product_id", req.ProductID.String()).
        Msg("Price validation passed")
    
    // 3. Call priceService.Create
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Msg("Calling priceService.Create")
        
    price, err := h.priceService.Create(ctx, &req)
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Create").
            Str("request_id", requestID).
            Err(err).
            Str("product_id", req.ProductID.String()).
            Str("name", req.Name).
            Msg("Service layer failed to create price")
            
        // Check for specific error types
        if strings.Contains(err.Error(), "product exists") {
            return echo.NewHTTPError(http.StatusNotFound, "Associated product not found")
        }
        
        if strings.Contains(err.Error(), "validation") {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }
        
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create price")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("price_id", price.ID.String()).
        Str("stripe_id", price.StripeID).
        Msg("Price successfully created by service layer")
    
    // 4. Map to response
    resp := response{
        ID:            price.ID.String(),
        ProductID:     price.ProductID.String(),
        Name:          price.Name,
        Amount:        price.Amount,
        Currency:      price.Currency,
        Interval:      price.Interval,
        IntervalCount: price.IntervalCount,
        Active:        price.Active,
        CreatedAt:     price.CreatedAt,
        UpdatedAt:     price.UpdatedAt,
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("response_status", http.StatusText(http.StatusCreated)).
        Msg("Preparing response")
    
    h.logger.Info().
        Str("handler", "PriceHandler.Create").
        Str("request_id", requestID).
        Str("price_id", price.ID.String()).
        Str("product_id", price.ProductID.String()).
        Str("name", price.Name).
        Str("stripe_id", price.StripeID).
        Int("status_code", http.StatusCreated).
        Msg("Price created successfully")
    
    return c.JSON(http.StatusCreated, resp)
}

// Get handles GET /api/prices/:id
func (h *PriceHandler) Get(c echo.Context) error {
    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling get price by ID request")
    
    // 1. Parse ID from URL
    idParam := c.Param("id")
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("id_param", idParam).
        Msg("Parsing price ID from URL")
    
    // Validate and convert the ID
    id, err := uuid.Parse(idParam)
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Get").
            Str("request_id", requestID).
            Str("id_param", idParam).
            Err(err).
            Msg("Invalid price ID format")
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid price ID format")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Msg("Price ID parsed successfully")
    
    // 2. Call priceService.GetByID
    h.logger.Debug().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Msg("Calling priceService.GetByID")
    
    price, err := h.priceService.GetByID(ctx, id)
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Get").
            Str("request_id", requestID).
            Str("price_id", id.String()).
            Err(err).
            Msg("Failed to retrieve price")
        
        // Check if it's a "not found" error
        if strings.Contains(err.Error(), "not found") {
            return echo.NewHTTPError(http.StatusNotFound, "Price not found")
        }
        
        // For other errors, return internal server error
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve price")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Str("product_id", price.ProductID.String()).
        Msg("Price retrieved successfully")
    
    // 3. Return appropriate response
    // Define a response structure for consistency
    type response struct {
        ID            string    `json:"id"`
        ProductID     string    `json:"product_id"`
        Name          string    `json:"name"`
        Amount        int64     `json:"amount"`
        Currency      string    `json:"currency"`
        Interval      string    `json:"interval"`
        IntervalCount int       `json:"interval_count"`
        Active        bool      `json:"active"`
        CreatedAt     time.Time `json:"created_at"`
        UpdatedAt     time.Time `json:"updated_at"`
    }
    
    resp := response{
        ID:            price.ID.String(),
        ProductID:     price.ProductID.String(),
        Name:          price.Name,
        Amount:        price.Amount,
        Currency:      price.Currency,
        Interval:      price.Interval,
        IntervalCount: price.IntervalCount,
        Active:        price.Active,
        CreatedAt:     price.CreatedAt,
        UpdatedAt:     price.UpdatedAt,
    }
    
    h.logger.Info().
        Str("handler", "PriceHandler.Get").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Str("name", price.Name).
        Int("status_code", http.StatusOK).
        Msg("Successfully returned price details")
    
    return c.JSON(http.StatusOK, resp)
}

// Get handles GET /api/prices/product/:productId
func (h *PriceHandler) ListByProduct(c echo.Context) error {
	// TODO: Implement price retrieval by productID
	// 1. Parse ID from URL
	// 2. Call priceService.GetByProductID
	// 3. Return appropriate response
	return nil
}

// List handles GET /api/prices
func (h *PriceHandler) List(c echo.Context) error {
    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling price listing request")
    
    // 1. Parse pagination parameters
    params := pagination.NewParams(c)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Int("offset", params.Offset).
        Int("per_page", params.PerPage).
        Int("page", params.Page).
        Msg("Pagination parameters parsed")
    
    // Parse additional filtering parameters
    includeInactive := false
    if c.QueryParam("include_inactive") == "true" {
        // Only admins should see inactive prices
        includeInactive = true
        h.logger.Debug().
            Str("handler", "PriceHandler.List").
            Str("request_id", requestID).
            Bool("include_inactive", includeInactive).
            Msg("Including inactive prices in results")
    }
    
    // Parse product ID filter if present
    var productID *uuid.UUID
    productIDParam := c.QueryParam("product_id")
    if productIDParam != "" {
        h.logger.Debug().
            Str("handler", "PriceHandler.List").
            Str("request_id", requestID).
            Str("product_id_param", productIDParam).
            Msg("Parsing product ID filter")
            
        id, err := uuid.Parse(productIDParam)
        if err != nil {
            h.logger.Error().
                Str("handler", "PriceHandler.List").
                Str("request_id", requestID).
                Str("product_id_param", productIDParam).
                Err(err).
                Msg("Invalid product ID format")
            return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID format")
        }
        
        productID = &id
        h.logger.Debug().
            Str("handler", "PriceHandler.List").
            Str("request_id", requestID).
            Str("product_id", id.String()).
            Msg("Product ID filter parsed successfully")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Int("offset", params.Offset).
        Int("per_page", params.PerPage).
        Bool("include_inactive", includeInactive).
        Msg("Calling priceService.List")
    
    // 2. Call priceService.List
    var prices []*models.Price
    var total int
    var err error
    
    if productID != nil {
        // If product ID is provided, call ListByProductID
        h.logger.Debug().
            Str("handler", "PriceHandler.List").
            Str("request_id", requestID).
            Str("product_id", productID.String()).
            Msg("Calling priceService.ListByProductID")
            
        prices, err = h.priceService.ListByProductID(ctx, *productID, includeInactive)
        total = len(prices)
    } else {
        // Otherwise call the standard List method
        prices, total, err = h.priceService.List(ctx, params.Offset, params.PerPage, includeInactive)
    }
    
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.List").
            Str("request_id", requestID).
            Err(err).
            Msg("Failed to retrieve prices from service")
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve prices")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Int("prices_count", len(prices)).
        Int("total_count", total).
        Msg("Prices retrieved successfully")
    
    // 3. Create the response structure
    type priceResponse struct {
        ID            string    `json:"id"`
        ProductID     string    `json:"product_id"`
        Name          string    `json:"name"`
        Amount        int64     `json:"amount"`
        Currency      string    `json:"currency"`
        Type          string    `json:"type"`
        Interval      string    `json:"interval"`
        IntervalCount int       `json:"interval_count"`
        Active        bool      `json:"active"`
        CreatedAt     time.Time `json:"created_at"`
        UpdatedAt     time.Time `json:"updated_at"`
    }
    
    // Map the prices to response objects
    priceResponses := make([]priceResponse, len(prices))
    for i, price := range prices {
        priceResponses[i] = priceResponse{
            ID:            price.ID.String(),
            ProductID:     price.ProductID.String(),
            Name:          price.Name,
            Amount:        price.Amount,
            Currency:      price.Currency,
            Type:          price.Type,
            Interval:      price.Interval,
            IntervalCount: price.IntervalCount,
            Active:        price.Active,
            CreatedAt:     price.CreatedAt,
            UpdatedAt:     price.UpdatedAt,
        }
    }
    
    // Log a sample of prices (limited to first 5)
    if len(prices) > 0 {
        logCount := len(prices)
        if logCount > 5 {
            logCount = 5
        }
        
        for i := 0; i < logCount; i++ {
            h.logger.Debug().
                Str("handler", "PriceHandler.List").
                Str("request_id", requestID).
                Str("price_id", prices[i].ID.String()).
                Str("price_name", prices[i].Name).
                Int64("amount", prices[i].Amount).
                Str("currency", prices[i].Currency).
                Str("type", prices[i].Type).
                Str("interval", prices[i].Interval).
                Msgf("Price %d/%d in results", i+1, logCount)
        }
    }
    
    // 4. Return paginated response
    meta := pagination.NewMeta(params, total)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Int("current_page", meta.Page).
        Int("total_pages", meta.TotalPages).
        Int("per_page", meta.PerPage).
        Int("total", meta.Total).
        Msg("Pagination metadata generated")
    
    response := pagination.Response(priceResponses, meta)
    
    h.logger.Info().
        Str("handler", "PriceHandler.List").
        Str("request_id", requestID).
        Int("prices_count", len(prices)).
        Int("total_count", total).
        Int("page", params.Page).
        Int("per_page", params.PerPage).
        Int("status_code", http.StatusOK).
        Msg("Price listing successfully returned")
    
    return c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/prices/:id
func (h *PriceHandler) Update(c echo.Context) error {
	// TODO: Implement price update
	// 1. Parse ID from URL
	// 2. Bind request to PriceUpdateDTO
	// 3. Validate DTO
	// 4. Call priceService.Update
	// 5. Return appropriate response
	return nil
}

// Delete handles DELETE /api/prices/:id
func (h *PriceHandler) Delete(c echo.Context) error {
    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Delete").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling price deletion request")
    
    // 1. Parse ID from URL
    idParam := c.Param("id")
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Delete").
        Str("request_id", requestID).
        Str("id_param", idParam).
        Msg("Parsing price ID from URL")
    
    // Validate and convert the ID
    id, err := uuid.Parse(idParam)
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Delete").
            Str("request_id", requestID).
            Str("id_param", idParam).
            Err(err).
            Msg("Invalid price ID format")
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid price ID format")
    }
    
    h.logger.Debug().
        Str("handler", "PriceHandler.Delete").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Msg("Price ID parsed successfully")
    
    // 2. Call priceService.Delete
    h.logger.Debug().
        Str("handler", "PriceHandler.Delete").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Msg("Calling priceService.Delete")
    
    err = h.priceService.Delete(ctx, id)
    if err != nil {
        h.logger.Error().
            Str("handler", "PriceHandler.Delete").
            Str("request_id", requestID).
            Str("price_id", id.String()).
            Err(err).
            Msg("Failed to delete price")
        
        // Check if it's a "not found" error
        if strings.Contains(err.Error(), "not found") {
            return echo.NewHTTPError(http.StatusNotFound, "Price not found")
        }
        
        // Check if it's a "in use" error (if you implement that check)
        if strings.Contains(err.Error(), "in use") || strings.Contains(err.Error(), "active subscriptions") {
            return echo.NewHTTPError(http.StatusConflict, "Cannot delete price that is in use by active subscriptions")
        }
        
        // For other errors, return internal server error
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete price")
    }
    
    h.logger.Info().
        Str("handler", "PriceHandler.Delete").
        Str("request_id", requestID).
        Str("price_id", id.String()).
        Int("status_code", http.StatusNoContent).
        Msg("Price deleted successfully")
    
    // 3. Return appropriate response (204 No Content for successful deletion)
    return c.NoContent(http.StatusNoContent)
}

// Helper functions for mapping between models and DTOs can be added here
