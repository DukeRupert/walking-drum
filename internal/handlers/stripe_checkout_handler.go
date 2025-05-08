package handlers

import (
	"net/http"

	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type StripeCheckoutHandler struct {
	logger         *zerolog.Logger
	stripeClient   *stripe.Client
	productService services.ProductService
	priceService   services.PriceService
	customerService services.CustomerService
}

func NewStripeCheckoutHandler(
	logger *zerolog.Logger,
	stripeClient *stripe.Client,
	productService services.ProductService,
	priceService services.PriceService,
	customerService services.CustomerService,
) *StripeCheckoutHandler {
	return &StripeCheckoutHandler{
		logger:         logger,
		stripeClient:   stripeClient,
		productService: productService,
		priceService:   priceService,
		customerService: customerService,
	}
}

// Request structure for creating a checkout session
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	CustomerID string `json:"customer_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// Response structure for a checkout session
type CheckoutSessionResponse struct {
	ClientSecret string `json:"client_secret"`
}

// CreateCheckoutSession creates a new Stripe checkout session
func (h *StripeCheckoutHandler) CreateCheckoutSession(c echo.Context) error {
	var req CreateCheckoutSessionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	// Validate request
	if req.PriceID == "" || req.CustomerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Price ID and Customer ID are required",
		})
	}

	// Convert string IDs to UUIDs
	priceID, err := uuid.Parse(req.PriceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid Price ID",
		})
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid Customer ID",
		})
	}

	// Get price from database
	price, err := h.priceService.GetByID(c.Request().Context(), priceID)
	if err != nil {
		h.logger.Error().Err(err).Str("priceID", req.PriceID).Msg("Failed to get price")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get price",
		})
	}

	// Get customer from database
	customer, err := h.customerService.GetByID(c.Request().Context(), customerID)
	if err != nil {
		h.logger.Error().Err(err).Str("customerID", req.CustomerID).Msg("Failed to get customer")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get customer",
		})
	}

	// Create checkout session
	session, err := h.stripeClient.CreateSubscriptionCheckoutSession(
		customer.StripeID,
		price.StripeID,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create checkout session")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create checkout session",
		})
	}

	// Return the client secret
	return c.JSON(http.StatusOK, CheckoutSessionResponse{
		ClientSecret: session.ClientSecret,
	})
}

func (h *StripeCheckoutHandler) Register(g *echo.Group) {
	g.POST("/checkout/create-session", h.CreateCheckoutSession)
}