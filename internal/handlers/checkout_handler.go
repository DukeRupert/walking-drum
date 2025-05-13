package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/internal/services/stripe"

    gostripe "github.com/stripe/stripe-go/v82"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// CheckoutHandler handles checkout-related requests
type CheckoutHandler struct {
	logger             *zerolog.Logger
	stripeClient       stripe.StripeService
	productService     services.ProductService
	priceService       services.PriceService
	customerService    services.CustomerService
	subscriptionService services.SubscriptionService
}

// NewCheckoutHandler creates a new checkout handler
func NewCheckoutHandler(
	logger *zerolog.Logger,
	stripeClient stripe.StripeService,
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
	Quantity   int    `json:"quantity"`
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
		req.Quantity,
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

// CreateMultiItemSessionRequest is the request structure for creating a multi-item checkout session
type CreateMultiItemSessionRequest struct {
	Items      []CheckoutItem `json:"items"`
	CustomerID string         `json:"customer_id"`
	ReturnURL  string         `json:"return_url"`
}

// CheckoutItem represents a single item in the checkout
type CheckoutItem struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

// CreateMultiItemSessionResponse is the response body for creating a multi-item checkout session
type CreateMultiItemSessionResponse struct {
	ClientSecret string `json:"client_secret"`
	ID           string `json:"id"`
}

// CreateMultiItemSession creates a Stripe checkout session with multiple subscription items
func (h *CheckoutHandler) CreateMultiItemSession(c echo.Context) error {
	h.logger.Debug().Msg("CreateMultiItemSession called")

	// Check dependencies
	if h.stripeClient == nil || h.productService == nil || h.priceService == nil || h.customerService == nil {
		h.logger.Error().
			Bool("stripeClient_nil", h.stripeClient == nil).
			Bool("productService_nil", h.productService == nil).
			Bool("priceService_nil", h.priceService == nil).
			Bool("customerService_nil", h.customerService == nil).
			Msg("Service dependencies missing")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Required services not initialized",
		})
	}

	h.logger.Debug().Msg("Binding request")
	var req CreateMultiItemSessionRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	h.logger.Debug().
		Interface("items", req.Items).
		Str("customerID", req.CustomerID).
		Msg("Request parsed")

	// Validate request
	if len(req.Items) == 0 || req.CustomerID == "" || req.ReturnURL == "" {
		h.logger.Error().Msg("Missing required fields")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required fields",
		})
	}

	// Parse customer UUID
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		h.logger.Error().Err(err).Str("customerID", req.CustomerID).Msg("Invalid customer ID")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

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

	// Check if customer has Stripe ID
	if customer.StripeID == "" {
		h.logger.Error().Str("customerID", req.CustomerID).Msg("Customer has no Stripe ID")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Customer not synchronized with Stripe",
		})
	}

	// Build line items for Stripe checkout
	lineItems := []*gostripe.CheckoutSessionLineItemParams{}
	metadata := make(map[string]string)

	for i, item := range req.Items {
		// Parse price UUID
		priceID, err := uuid.Parse(item.PriceID)
		if err != nil {
			h.logger.Error().Err(err).Str("priceID", item.PriceID).Msg("Invalid price ID")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Invalid price ID for item %d", i),
			})
		}

		// Get price from database
		price, err := h.priceService.GetByID(c.Request().Context(), priceID)
		if err != nil {
			h.logger.Error().Err(err).Str("priceID", item.PriceID).Msg("Failed to get price")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to get price details for item %d", i),
			})
		}

		if price == nil {
			h.logger.Error().Str("priceID", item.PriceID).Msg("Price not found")
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("Price not found for item %d", i),
			})
		}

		// Check if price has Stripe ID
		if price.StripeID == "" {
			h.logger.Error().Str("priceID", item.PriceID).Msg("Price has no Stripe ID")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Price not synchronized with Stripe for item %d", i),
			})
		}

		// Get product from database
		product, err := h.productService.GetByID(c.Request().Context(), price.ProductID)
		if err != nil {
			h.logger.Error().Err(err).Str("productID", price.ProductID.String()).Msg("Failed to get product")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to get product details for item %d", i),
			})
		}

		if product == nil {
			h.logger.Error().Str("productID", price.ProductID.String()).Msg("Product not found")
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("Product not found for item %d", i),
			})
		}

		// Ensure quantity is valid
		quantity := item.Quantity
		if quantity < 1 {
			quantity = 1
		}

		// Add line item
		lineItems = append(lineItems, &gostripe.CheckoutSessionLineItemParams{
			Price:    gostripe.String(price.StripeID),
			Quantity: gostripe.Int64(int64(quantity)),
		})

		// Add product details to metadata
		metadata[fmt.Sprintf("product_%d_name", i)] = product.Name
		metadata[fmt.Sprintf("product_%d_id", i)] = product.ID.String()
		metadata[fmt.Sprintf("price_%d_id", i)] = price.ID.String()
	}

	h.logger.Debug().
		Str("customerStripeID", customer.StripeID).
		Int("lineItemCount", len(lineItems)).
		Msg("Creating multi-item checkout session")

	// Convert the line items to our service-specific format
	checkoutItems := make([]stripe.CheckoutItem, 0, len(lineItems))
	for _, item := range lineItems {
		priceID := ""
		if item.Price != nil {
			priceID = *item.Price
		}
		
		quantity := 1
		if item.Quantity != nil {
			quantity = int(*item.Quantity)
		}
		
		checkoutItems = append(checkoutItems, stripe.CheckoutItem{
			PriceID:  priceID,
			Quantity: quantity,
		})
	}

	// Create multi-item checkout session using your stripeClient
	session, err := h.stripeClient.CreateMultiItemCheckoutSession(
		customer.StripeID, 
		checkoutItems,
		metadata,
		req.ReturnURL,
	)

	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create Stripe multi-item checkout session")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to create checkout session: %v", err),
		})
	}

	h.logger.Debug().Msg("Multi-item checkout session created successfully")

	// Return the client secret
	return c.JSON(http.StatusOK, CreateMultiItemSessionResponse{
		ClientSecret: session.ClientSecret,
		ID:           session.ID,
	})
}

// VerifyMultiItemSession verifies a checkout session with multiple items
func (h *CheckoutHandler) VerifyMultiItemSession(c echo.Context) error {
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

	// Check session status
	if session.Status != "complete" && session.PaymentStatus != "paid" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": session.Status,
		})
	}

	// Extract subscription details
	subscriptionItems := []map[string]interface{}{}

	// In a production environment, you'd query your database for the actual subscription details
	// using the Stripe subscription ID and/or the metadata from the session.
	// This is a simplified version for demonstration purposes.

	// Extract product information from metadata
	productCount := 0
	for key := range session.Metadata {
		if strings.HasPrefix(key, "product_") && strings.HasSuffix(key, "_name") {
			productCount++
		}
	}

	// Create sample subscription items (in production, get these from your database)
	for i := 0; i < productCount; i++ {
		productName := session.Metadata[fmt.Sprintf("product_%d_name", i)]
		subscriptionItems = append(subscriptionItems, map[string]interface{}{
			"id":                fmt.Sprintf("sub_item_%d", i+1),
			"product_name":      productName,
			"amount":            session.AmountTotal / int64(productCount), // Simplified division
			"currency":          session.Currency,
			"interval":          "week", // In production, fetch this from DB
			"quantity":          1,      // In production, fetch this from DB
			"next_delivery_date": time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
		})
	}

	// If no products were found in metadata, create a default subscription item
	if len(subscriptionItems) == 0 {
		subscriptionItems = append(subscriptionItems, map[string]interface{}{
			"id":                session.ID,
			"product_name":      "Coffee Subscription",
			"amount":            session.AmountTotal,
			"currency":          session.Currency,
			"interval":          "week",
			"quantity":          1,
			"next_delivery_date": time.Now().AddDate(0, 0, 7).Format(time.RFC3339),
		})
	}

	// Return the subscription details
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       session.Status,
		"subscriptions": subscriptionItems,
	})
}

// Register registers the checkout handler routes
func (h *CheckoutHandler) Register(g *echo.Group) {
	g.POST("/checkout/create-session", h.CreateSession)
    g.POST("/checkout/create-multi-item-session", h.CreateMultiItemSession)
	g.GET("/checkout/verify-session", h.VerifySession)
}