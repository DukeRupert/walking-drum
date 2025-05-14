// internal/events/handlers.go
package events

import (
	"encoding/json"

	"github.com/dukerupert/walking-drum/internal/events"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/rs/zerolog"
)

// EventHandler manages event subscriptions and processing
type EventHandler struct {
	bus             events.NATSEventBus
	logger          zerolog.Logger
	productService  services.ProductService
	customerService services.CustomerService
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	bus events.NATSEventBus,
	logger *zerolog.Logger,
	productService services.ProductService,
	customerService services.CustomerService,
) *EventHandler {
	return &EventHandler{
		bus:             bus,
		logger:          logger.With().Str("component", "event_handler").Logger(),
		productService:  productService,
		customerService: customerService,
	}
}

// RegisterHandlers sets up all event subscriptions
func (h *EventHandler) RegisterHandlers() error {
	// Subscribe to product events
	if _, err := h.bus.Subscribe("products.created", h.handleProductCreated); err != nil {
		return err
	}
	if _, err := h.bus.Subscribe("products.stock_updated", h.handleProductStockUpdated); err != nil {
		return err
	}
	
	// Subscribe to customer events
	if _, err := h.bus.Subscribe("customers.created", h.handleCustomerCreated); err != nil {
		return err
	}
	
	// Subscribe to subscription events
	if _, err := h.bus.Subscribe("subscriptions.created", h.handleSubscriptionCreated); err != nil {
		return err
	}
	
	h.logger.Info().Msg("Event handlers registered successfully")
	return nil
}

// Event handler implementations
func (h *EventHandler) handleProductCreated(data []byte) {
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal product.created event")
		return
	}
	
	h.logger.Info().
		Str("event_id", event.ID).
		Interface("payload", event.Payload).
		Msg("Processing product.created event")
	
	// Process the event...
	// This might involve notifying other systems, updating caches, etc.
}

func (h *EventHandler) handleProductStockUpdated(data []byte) {
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal product.stock_updated event")
		return
	}
	
	h.logger.Info().
		Str("event_id", event.ID).
		Interface("payload", event.Payload).
		Msg("Processing product.stock_updated event")
	
	// Check if stock is low and notify if needed
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		h.logger.Error().Msg("Invalid payload format")
		return
	}
	
	newStock, ok := payload["new_stock_level"].(float64)
	if !ok {
		h.logger.Error().Msg("Invalid stock level format")
		return
	}
	
	// If stock is low, publish a low stock event
	if newStock < 10 {
		h.logger.Info().
			Float64("stock_level", newStock).
			Str("product_name", payload["product_name"].(string)).
			Msg("Product has low stock")
		
		h.bus.Publish("products.low_stock", payload)
	}
}

func (h *EventHandler) handleCustomerCreated(data []byte) {
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal customer.created event")
		return
	}
	
	h.logger.Info().
		Str("event_id", event.ID).
		Interface("payload", event.Payload).
		Msg("Processing customer.created event")
	
	// Process the event...
	// This might involve sending welcome emails, setting up analytics, etc.
}

func (h *EventHandler) handleSubscriptionCreated(data []byte) {
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal subscription.created event")
		return
	}
	
	h.logger.Info().
		Str("event_id", event.ID).
		Interface("payload", event.Payload).
		Msg("Processing subscription.created event")
	
	// Process the event...
	// This might involve scheduling initial deliveries, sending confirmations, etc.
}