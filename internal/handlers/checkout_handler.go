package handlers

import (
	"fmt"
	"net/http"

	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// CheckoutHandler handles checkout-related requests
type CheckoutHandler struct {
	logger             *zerolog.Logger
	stripeClient       *stripe.Client
	productService     services.ProductService
	priceService       services.PriceService
	customerService    services.CustomerService
	subscriptionService services.SubscriptionService
}

// NewCheckoutHandler creates a new checkout handler
func NewCheckoutHandler(
	logger *zerolog.Logger,
	stripeClient *stripe.Client,
	productService services.ProductService,
	priceService services.PriceService,
	customerService services.CustomerService,
	subscriptionService services.SubscriptionService,
) *CheckoutHandler {
	return &CheckoutHandler{
		logger:             logger,
		stripeClient:       stripeClient,
		productService:     productService,
		priceService:       priceService,
		customerService:    customerService,
		subscriptionService: subscriptionService,
	}
}

// CreateSessionRequest is the request body for creating a checkout session
type CreateSessionRequest struct {
	PriceID    string `json:"price_id"`
	CustomerID string `json:"customer_id"`
	ReturnURL  string `json:"return_url"`
}

// CreateSessionResponse is the response body for creating a checkout session
type CreateSessionResponse struct {
	ClientSecret string `json:"client_secret"`
}

func (h *CheckoutHandler) CreateSession(c echo.Context) error {
    h.logger.Debug().Msg("CreateSession called")
    
    // Check if handler itself is nil (shouldn't happen, but let's be thorough)
    if h == nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Handler not initialized",
        })
    }
    
    // Check dependencies
    if h.stripeClient == nil {
        h.logger.Error().Msg("stripeClient is nil")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Stripe client not initialized",
        })
    }
    
    if h.productService == nil || h.priceService == nil || h.customerService == nil {
        h.logger.Error().
            Bool("productService_nil", h.productService == nil).
            Bool("priceService_nil", h.priceService == nil).
            Bool("customerService_nil", h.customerService == nil).
            Msg("Service dependencies missing")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Required services not initialized",
        })
    }
    
    h.logger.Debug().Msg("Binding request")
    var req CreateSessionRequest
    if err := c.Bind(&req); err != nil {
        h.logger.Error().Err(err).Msg("Failed to bind request")
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid request body",
        })
    }
    
    h.logger.Debug().
        Str("priceID", req.PriceID).
        Str("customerID", req.CustomerID).
        Msg("Request parsed")
    
    // Validate request
    if req.PriceID == "" || req.CustomerID == "" || req.ReturnURL == "" {
        h.logger.Error().Msg("Missing required fields")
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Missing required fields",
        })
    }
    
    h.logger.Debug().Msg("Parsing UUIDs")
    // Parse UUIDs
    priceID, err := uuid.Parse(req.PriceID)
    if err != nil {
        h.logger.Error().Err(err).Str("priceID", req.PriceID).Msg("Invalid price ID")
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid price ID",
        })
    }
    
    customerID, err := uuid.Parse(req.CustomerID)
    if err != nil {
        h.logger.Error().Err(err).Str("customerID", req.CustomerID).Msg("Invalid customer ID")
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid customer ID",
        })
    }
    
    h.logger.Debug().Msg("Getting price from database")
    // Get price from database
    price, err := h.priceService.GetByID(c.Request().Context(), priceID)
    if err != nil {
        h.logger.Error().Err(err).Str("priceID", req.PriceID).Msg("Failed to get price")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to get price details",
        })
    }
    
    if price == nil {
        h.logger.Error().Str("priceID", req.PriceID).Msg("Price not found")
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "Price not found",
        })
    }
    
    h.logger.Debug().Msg("Getting product from database")
    // Get product from database
    product, err := h.productService.GetByID(c.Request().Context(), price.ProductID)
    if err != nil {
        h.logger.Error().Err(err).Str("productID", price.ProductID.String()).Msg("Failed to get product")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to get product details",
        })
    }
    
    if product == nil {
        h.logger.Error().Str("productID", price.ProductID.String()).Msg("Product not found")
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "Product not found",
        })
    }
    
    h.logger.Debug().Msg("Getting customer from database")
    // Get customer from database
    customer, err := h.customerService.GetByID(c.Request().Context(), customerID)
    if err != nil {
        h.logger.Error().Err(err).Str("customerID", req.CustomerID).Msg("Failed to get customer")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to get customer details",
        })
    }
    
    if customer == nil {
        h.logger.Error().Str("customerID", req.CustomerID).Msg("Customer not found")
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "Customer not found",
        })
    }
    
    // Check if customer.StripeID or price.StripeID is empty
    if customer.StripeID == "" {
        h.logger.Error().Str("customerID", req.CustomerID).Msg("Customer has no Stripe ID")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Customer not synchronized with Stripe",
        })
    }
    
    if price.StripeID == "" {
        h.logger.Error().Str("priceID", req.PriceID).Msg("Price has no Stripe ID")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Price not synchronized with Stripe",
        })
    }
    
    h.logger.Debug().
        Str("customerStripeID", customer.StripeID).
        Str("priceStripeID", price.StripeID).
        Str("productName", product.Name).
        Msg("Creating checkout session")
    
    // Create checkout session
    session, err := h.stripeClient.CreateEmbeddedCheckoutSession(
        customer.StripeID, 
        price.StripeID,
        product.Name,
        req.ReturnURL,
    )
    
    if err != nil {
        h.logger.Error().Err(err).Msg("Failed to create Stripe checkout session")
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": fmt.Sprintf("Failed to create checkout session: %v", err),
        })
    }
    
    h.logger.Debug().Msg("Checkout session created successfully")
    
    // Return the client secret
    return c.JSON(http.StatusOK, CreateSessionResponse{
        ClientSecret: session.ClientSecret,
    })
}

// VerifySession verifies a checkout session
func (h *CheckoutHandler) VerifySession(c echo.Context) error {
	sessionID := c.QueryParam("session_id")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing session ID",
		})
	}

	// Retrieve the session from Stripe
	session, err := h.stripeClient.RetrieveCheckoutSession(sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("sessionID", sessionID).Msg("Failed to retrieve checkout session")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify session",
		})
	}

	// For a real implementation, you'd create a subscription in your database
	// and return the details. This is simplified for the POC.
	
	// Return the session details
	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscription": map[string]interface{}{
			"id":                session.ID,
			"product_name":      "Cloud 9 Espresso", // In production, fetch this from DB
			"amount":            session.AmountTotal,
			"currency":          session.Currency,
			"interval":          "week", // In production, fetch this from DB
			"next_delivery_date": "2025-05-15T00:00:00Z", // Example date
		},
	})
}

// Register registers the checkout handler routes
func (h *CheckoutHandler) Register(g *echo.Group) {
	g.POST("/checkout/create-session", h.CreateSession)
	g.GET("/checkout/verify-session", h.VerifySession)
}