// internal/events/handlers/variant_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/events"
	"github.com/dukerupert/walking-drum/internal/interfaces"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// VariantHandler handles variant-related events
type VariantHandler struct {
	variantService services.VariantService
	eventBus       interfaces.EventBus
	logger         zerolog.Logger
}

// NewVariantHandler creates a new variant handler
func NewVariantHandler(
	variantService services.VariantService,
	eventBus interfaces.EventBus,
	logger *zerolog.Logger,
) *VariantHandler {
	return &VariantHandler{
		variantService: variantService,
		eventBus:       eventBus,
		logger:         logger.With().Str("component", "variant_handler").Logger(),
	}
}

// Register registers all event handlers
func (h *VariantHandler) Register() error {
	// Subscribe to products.created event
	_, err := h.eventBus.Subscribe("products.created", h.handleProductCreated)
	if err != nil {
		return fmt.Errorf("failed to subscribe to products.created event: %w", err)
	}

	h.logger.Info().Msg("Variant handler registered")
	return nil
}

// handleProductCreated handles the products.created event
func (h *VariantHandler) handleProductCreated(data []byte) {
	h.logger.Debug().
		Str("data", string(data)).
		Msg("Received products.created event")

	// Parse the event
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to unmarshal products.created event")
		return
	}

	// Parse the payload
	var payload map[string]interface{}
	payloadData, err := json.Marshal(event.Payload)
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to marshal event payload")
		return
	}

	if err := json.Unmarshal(payloadData, &payload); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to unmarshal event payload")
		return
	}

	// Extract product ID
	productIDStr, ok := payload["product_id"].(string)
	if !ok {
		h.logger.Error().
			Msg("Product ID missing from payload")
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("product_id_str", productIDStr).
			Msg("Invalid product ID format")
		return
	}

	// Extract options
	options := make(map[string][]string)
	
	// Default options if not provided
	options["weight"] = []string{"12oz", "3lb", "5lb"}
	options["grind"] = []string{"Whole Bean", "Drip Ground"}
	
	// Check if options are provided in the payload
	if optionsData, ok := payload["options"].(map[string]interface{}); ok {
		// Process weight options
		if weightData, ok := optionsData["weight"].([]interface{}); ok {
			weights := make([]string, 0, len(weightData))
			for _, w := range weightData {
				if weight, ok := w.(string); ok {
					weights = append(weights, weight)
				}
			}
			if len(weights) > 0 {
				options["weight"] = weights
			}
		}
		
		// Process grind options
		if grindData, ok := optionsData["grind"].([]interface{}); ok {
			grinds := make([]string, 0, len(grindData))
			for _, g := range grindData {
				if grind, ok := g.(string); ok {
					grinds = append(grinds, grind)
				}
			}
			if len(grinds) > 0 {
				options["grind"] = grinds
			}
		}
	}

	// Generate variants
	h.logger.Info().
		Str("product_id", productID.String()).
		Interface("options", options).
		Msg("Generating variants for product")

	ctx := context.Background()
	if err := h.variantService.GenerateVariantsForProduct(ctx, productID, options); err != nil {
		h.logger.Error().
			Err(err).
			Str("product_id", productID.String()).
			Msg("Failed to generate variants for product")
		return
	}

	h.logger.Info().
		Str("product_id", productID.String()).
		Msg("Successfully generated variants for product")
}