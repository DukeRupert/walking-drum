// internal/handlers/customer_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/pkg/pagination"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// CustomerHandler handles HTTP requests for customers
type CustomerHandler struct {
	customerService services.CustomerService
	logger         zerolog.Logger
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService services.CustomerService, logger *zerolog.Logger) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
		logger: logger.With().Str("component", "customer_handler").Logger(),
	}
}

// Create handles POST /api/customers
func (h *CustomerHandler) Create(c echo.Context) error {
    // Define the request/response types inline
    type request = dto.CustomerCreateDTO
    type response struct {
        ID          string    `json:"id"`
        Email       string    `json:"email"`
        FirstName   string    `json:"first_name"`
        LastName    string    `json:"last_name"`
        PhoneNumber string    `json:"phone_number"`
        Active      bool      `json:"active"`
        CreatedAt   time.Time `json:"created_at"`
        UpdatedAt   time.Time `json:"updated_at"`
    }

    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling customer creation request")
    
    // 1. Bind request to CustomerCreateDTO
    var req request
    if err := c.Bind(&req); err != nil {
        h.logger.Error().
            Str("handler", "CustomerHandler.Create").
            Str("request_id", requestID).
            Err(err).
            Msg("Failed to bind request body")
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("email", req.Email).
        Str("first_name", req.FirstName).
        Str("last_name", req.LastName).
        Msg("Request body successfully bound")
    
    // 2. Validate DTO
    if problems := req.Valid(ctx); len(problems) > 0 {
        h.logger.Error().
            Str("handler", "CustomerHandler.Create").
            Str("request_id", requestID).
            Interface("problems", problems).
            Str("email", req.Email).
            Msg("Customer validation failed")
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error":    "Validation failed",
            "problems": problems,
        })
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("email", req.Email).
        Msg("Customer validation passed")
    
    // 3. Call customerService.Create
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Msg("Calling customerService.Create")
        
    customer, err := h.customerService.Create(ctx, &req)
    if err != nil {
        h.logger.Error().
            Str("handler", "CustomerHandler.Create").
            Str("request_id", requestID).
            Err(err).
            Str("email", req.Email).
            Msg("Service layer failed to create customer")
            
        // Check for specific error types
        if strings.Contains(err.Error(), "already exists") {
            return echo.NewHTTPError(http.StatusConflict, "Customer with this email already exists")
        }
        
        if strings.Contains(err.Error(), "validation") {
            return echo.NewHTTPError(http.StatusBadRequest, err.Error())
        }
        
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create customer")
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("customer_id", customer.ID.String()).
        Str("stripe_id", customer.StripeID).
        Msg("Customer successfully created by service layer")
    
    // 4. Map to response
    resp := response{
        ID:          customer.ID.String(),
        Email:       customer.Email,
        FirstName:   customer.FirstName,
        LastName:    customer.LastName,
        PhoneNumber: customer.PhoneNumber,
        Active:      customer.Active,
        CreatedAt:   customer.CreatedAt,
        UpdatedAt:   customer.UpdatedAt,
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("response_status", http.StatusText(http.StatusCreated)).
        Msg("Preparing response")
    
    h.logger.Info().
        Str("handler", "CustomerHandler.Create").
        Str("request_id", requestID).
        Str("customer_id", customer.ID.String()).
        Str("email", customer.Email).
        Str("stripe_id", customer.StripeID).
        Int("status_code", http.StatusCreated).
        Msg("Customer created successfully")
    
    return c.JSON(http.StatusCreated, resp)
}

// Get handles GET /api/customers/:id
func (h *CustomerHandler) Get(c echo.Context) error {
    // 1. Parse ID from URL
    idParam := c.Param("id")
    if idParam == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Missing customer ID",
        })
    }

    // Parse ID to UUID
    id, err := uuid.Parse(idParam)
    if err != nil {
        h.logger.Error().Err(err).Str("id", idParam).Msg("Invalid customer ID format")
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid customer ID format",
        })
    }

    // 2. Call customerService.GetByID
    customer, err := h.customerService.GetByID(c.Request().Context(), id)
    if err != nil {
        h.logger.Error().Err(err).Str("id", idParam).Msg("Failed to get customer")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve customer",
        })
    }

    // Check if customer was found
    if customer == nil {
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "Customer not found",
        })
    }

    // 3. Return appropriate response
    return c.JSON(http.StatusOK, map[string]interface{}{
        "data": customer,
    })
}

// GetByEmail handles GET /api/customers/email/:email
func (h *CustomerHandler) GetByEmail(c echo.Context) error {
	// TODO: Implement customer retrieval by email
	// 1. Parse email from URL
	// 2. Call customerService.GetByEmail
	// 3. Return appropriate response
	return nil
}

// List handles GET /api/customers
func (h *CustomerHandler) List(c echo.Context) error {
    ctx := c.Request().Context()
    requestID := c.Response().Header().Get(echo.HeaderXRequestID)
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Str("method", c.Request().Method).
        Str("path", c.Request().URL.Path).
        Str("remote_addr", c.Request().RemoteAddr).
        Msg("Handling customer listing request")
    
    // 1. Parse pagination parameters
    params := pagination.NewParams(c)
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Int("offset", params.Offset).
        Int("per_page", params.PerPage).
        Int("page", params.Page).
        Msg("Pagination parameters parsed")
    
    // Parse additional filtering parameters
    includeInactive := false
    if c.QueryParam("include_inactive") == "true" {
        // Only admins should see inactive customers
        includeInactive = true
        h.logger.Debug().
            Str("handler", "CustomerHandler.List").
            Str("request_id", requestID).
            Bool("include_inactive", includeInactive).
            Msg("Including inactive customers in results")
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Int("offset", params.Offset).
        Int("per_page", params.PerPage).
        Bool("include_inactive", includeInactive).
        Msg("Calling customerService.List")
    
    // 2. Call customerService.List
    customers, total, err := h.customerService.List(ctx, params.Offset, params.PerPage, includeInactive)
    if err != nil {
        h.logger.Error().
            Str("handler", "CustomerHandler.List").
            Str("request_id", requestID).
            Err(err).
            Int("offset", params.Offset).
            Int("per_page", params.PerPage).
            Bool("include_inactive", includeInactive).
            Msg("Failed to retrieve customers from service")
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve customers")
    }
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Int("customers_count", len(customers)).
        Int("total_count", total).
        Msg("Customers retrieved successfully")
    
    // 3. Create the response structure
    type customerResponse struct {
        ID          string    `json:"id"`
        Email       string    `json:"email"`
        FirstName   string    `json:"first_name"`
        LastName    string    `json:"last_name"`
        PhoneNumber string    `json:"phone_number"`
        Active      bool      `json:"active"`
        CreatedAt   time.Time `json:"created_at"`
        UpdatedAt   time.Time `json:"updated_at"`
    }
    
    // Map the customers to response objects
    customerResponses := make([]customerResponse, len(customers))
    for i, customer := range customers {
        customerResponses[i] = customerResponse{
            ID:          customer.ID.String(),
            Email:       customer.Email,
            FirstName:   customer.FirstName,
            LastName:    customer.LastName,
            PhoneNumber: customer.PhoneNumber,
            Active:      customer.Active,
            CreatedAt:   customer.CreatedAt,
            UpdatedAt:   customer.UpdatedAt,
        }
    }
    
    // Log a sample of customers (limited to first 5)
    if len(customers) > 0 {
        logCount := len(customers)
        if logCount > 5 {
            logCount = 5
        }
        
        for i := 0; i < logCount; i++ {
            h.logger.Debug().
                Str("handler", "CustomerHandler.List").
                Str("request_id", requestID).
                Str("customer_id", customers[i].ID.String()).
                Str("email", customers[i].Email).
                Str("name", fmt.Sprintf("%s %s", customers[i].FirstName, customers[i].LastName)).
                Msgf("Customer %d/%d in results", i+1, logCount)
        }
    }
    
    // 4. Create and return paginated response
    meta := pagination.NewMeta(params, total)
    
    h.logger.Debug().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Int("current_page", meta.Page).
        Int("total_pages", meta.TotalPages).
        Int("per_page", meta.PerPage).
        Int("total", meta.Total).
        Msg("Pagination metadata generated")
    
    response := pagination.Response(customerResponses, meta)
    
    h.logger.Info().
        Str("handler", "CustomerHandler.List").
        Str("request_id", requestID).
        Int("customers_count", len(customers)).
        Int("total_count", total).
        Int("page", params.Page).
        Int("per_page", params.PerPage).
        Int("status_code", http.StatusOK).
        Msg("Customer listing successfully returned")
    
    return c.JSON(http.StatusOK, response)
}

// Update handles PUT /api/customers/:id
func (h *CustomerHandler) Update(c echo.Context) error {
	// TODO: Implement customer update
	// 1. Parse ID from URL
	// 2. Bind request to CustomerUpdateDTO
	// 3. Validate DTO
	// 4. Call customerService.Update
	// 5. Return appropriate response
	return nil
}

// Delete handles DELETE /api/customers/:id
func (h *CustomerHandler) Delete(c echo.Context) error {
	// TODO: Implement customer deletion
	// 1. Parse ID from URL
	// 2. Call customerService.Delete
	// 3. Return appropriate response
	return nil
}

// AddPaymentMethod handles POST /api/customers/:id/payment-methods
func (h *CustomerHandler) AddPaymentMethod(c echo.Context) error {
	// TODO: Implement adding payment method to customer
	// 1. Parse customer ID from URL
	// 2. Bind request with payment method ID
	// 3. Call customerService.AddPaymentMethod
	// 4. Return appropriate response
	return nil
}

// RemovePaymentMethod handles DELETE /api/customers/:id/payment-methods/:paymentMethodId
func (h *CustomerHandler) RemovePaymentMethod(c echo.Context) error {
	// TODO: Implement removing payment method from customer
	// 1. Parse customer ID and payment method ID from URL
	// 2. Call customerService.RemovePaymentMethod
	// 3. Return appropriate response
	return nil
}

// Helper functions for mapping between models and DTOs can be added here