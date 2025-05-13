// internal/handlers/webhook_handler.go
package handlers

import (
	"io"
	"net/http"

	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	stripeService stripe.StripeService
	webhookSecret string
	logger        zerolog.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	stripeService stripe.StripeService,
	webhookSecret string,
	logger *zerolog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		stripeService: stripeService,
		webhookSecret: webhookSecret,
		logger:        logger.With().Str("component", "webhook_handler").Logger(),
	}
}

// HandleWebhook processes incoming Stripe webhook events
func (h *WebhookHandler) HandleWebhook(c echo.Context) error {
	const MaxBodyBytes = int64(65536)

	// Create a request body reader with a size limit
	body := http.MaxBytesReader(c.Response().Writer, c.Request().Body, MaxBodyBytes)

	// Read the payload
	payload, err := io.ReadAll(body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Error reading webhook request body")
		return c.NoContent(http.StatusServiceUnavailable)
	}

	// Get the Stripe signature from headers
	signature := c.Request().Header.Get("Stripe-Signature")

	// Process the webhook
	if err := h.stripeService.ProcessWebhook(c.Request().Context(), payload, signature, h.webhookSecret); err != nil {
		h.logger.Error().Err(err).Msg("Failed to process webhook")
		return c.NoContent(http.StatusBadRequest)
	}

	return c.NoContent(http.StatusOK)
}
