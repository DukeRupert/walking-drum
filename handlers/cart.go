// handlers/cart_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type CartHandler struct {
	cartRepo     repository.CartRepository
	cartItemRepo repository.CartItemRepository
	productRepo  repository.ProductRepository
	priceRepo    repository.PriceRepository
}

func NewCartHandler(
	cartRepo repository.CartRepository,
	cartItemRepo repository.CartItemRepository,
	productRepo repository.ProductRepository,
	priceRepo repository.PriceRepository,
) *CartHandler {
	return &CartHandler{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		productRepo:  productRepo,
		priceRepo:    priceRepo,
	}
}

// Request/Response structs

type CreateCartRequest struct {
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	SessionID *string                `json:"session_id,omitempty"`
	ExpiresAt *string                `json:"expires_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AddToCartRequest struct {
	ProductID      uuid.UUID              `json:"product_id"`
	PriceID        *uuid.UUID             `json:"price_id,omitempty"`
	Quantity       int                    `json:"quantity"`
	IsSubscription bool                   `json:"is_subscription"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

type UpdateCartItemRequest struct {
	Quantity       *int                    `json:"quantity,omitempty"`
	IsSubscription *bool                   `json:"is_subscription,omitempty"`
	Options        *map[string]interface{} `json:"options,omitempty"`
}

type CartItemResponse struct {
	ID             uuid.UUID              `json:"id"`
	ProductID      uuid.UUID              `json:"product_id"`
	PriceID        *uuid.UUID             `json:"price_id,omitempty"`
	Quantity       int                    `json:"quantity"`
	UnitPrice      int64                  `json:"unit_price"`
	TotalPrice     int64                  `json:"total_price"`
	IsSubscription bool                   `json:"is_subscription"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	Options        map[string]interface{} `json:"options,omitempty"`
	Product        *ProductResponse       `json:"product,omitempty"`
	Price          *PriceResponse         `json:"price,omitempty"`
}

type CartResponse struct {
	ID         uuid.UUID              `json:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	SessionID  *string                `json:"session_id,omitempty"`
	ItemCount  int                    `json:"item_count"`
	TotalPrice int64                  `json:"total_price"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
	ExpiresAt  *string                `json:"expires_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Items      []CartItemResponse     `json:"items,omitempty"`
}

// Handlers

func (h *CartHandler) CreateCart(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "CreateCart").
		Logger()

	logger.Debug().Msg("Processing cart creation request")

	var req CreateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Bool("hasUserId", req.UserID != nil).
		Bool("hasSessionId", req.SessionID != nil).
		Bool("hasExpiresAt", req.ExpiresAt != nil).
		Bool("hasMetadata", len(req.Metadata) > 0).
		Msg("Received cart creation request")

	if req.UserID != nil {
		logger.Debug().Str("userId", req.UserID.String()).Msg("Cart creation with user ID")
	}

	if req.SessionID != nil {
		logger.Debug().Str("sessionId", *req.SessionID).Msg("Cart creation with session ID")
	}

	// Basic validation
	if req.UserID == nil && req.SessionID == nil {
		logger.Error().Msg("Neither user_id nor session_id provided")
		http.Error(w, "Either user_id or session_id is required", http.StatusBadRequest)
		return
	}

	// Create the cart
	logger.Debug().Msg("Creating cart object")
	cart := &models.Cart{
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}

	if req.ExpiresAt != nil {
		logger.Debug().
			Str("expiresAtRaw", *req.ExpiresAt).
			Msg("Parsing expires_at timestamp")

		expires, err := h.parseTimeStr(*req.ExpiresAt)
		if err != nil {
			logger.Error().
				Err(err).
				Str("expiresAtRaw", *req.ExpiresAt).
				Msg("Invalid date format for expires_at")
			http.Error(w, "Invalid date format for expires_at", http.StatusBadRequest)
			return
		}

		logger.Debug().
			Time("expiresAt", expires).
			Msg("Set cart expiration time")

		cart.ExpiresAt = &expires
	}

	if len(req.Metadata) > 0 {
		logger.Debug().
			Interface("metadata", req.Metadata).
			Msg("Setting cart metadata")

		cart.Metadata = &req.Metadata
	}

	logger.Debug().Msg("Creating cart in database")
	err := h.cartRepo.Create(r.Context(), cart)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to create cart")
		http.Error(w, "Failed to create cart", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cart.ID.String()).
		Msg("Cart created successfully, generating response")

	response, err := h.cartToResponse(r.Context(), cart, false)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cart.ID.String()).
			Msg("Failed to generate response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cart.ID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cart.ID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Construct a final success log with appropriate context
	logEvent := logger.Info().
		Str("cartId", cart.ID.String())

	if req.UserID != nil {
		logEvent = logEvent.Str("userId", req.UserID.String())
	}

	if req.SessionID != nil {
		logEvent = logEvent.Str("sessionId", *req.SessionID)
	}

	if cart.ExpiresAt != nil {
		logEvent = logEvent.Time("expiresAt", *cart.ExpiresAt)
	}

	logEvent.Msg("Cart created successfully")
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "GetCart").
		Logger()

	logger.Debug().Msg("Processing get cart request")

	vars := mux.Vars(r)
	cartIDStr := vars["id"]
	logger.Debug().Str("cartIdRaw", cartIDStr).Msg("Extracting cart ID from path")

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartIdRaw", cartIDStr).
			Msg("Invalid cart ID format")
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Parsed cart ID, fetching cart details")

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true
	logger.Debug().
		Bool("includeItems", includeItems).
		Msg("Include items in response")

	cart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Msg("Cart not found")
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to get cart from database")
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Bool("hasUserId", cart.UserID != nil).
		Bool("hasSessionId", cart.SessionID != nil).
		Msg("Cart retrieved successfully, generating response")

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to generate cart response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Construct a final success log with appropriate context
	logEvent := logger.Info().
		Str("cartId", cartID.String())

	if cart.UserID != nil {
		logEvent = logEvent.Str("userId", cart.UserID.String())
	}

	if cart.SessionID != nil {
		logEvent = logEvent.Str("sessionId", *cart.SessionID)
	}

	if cart.ExpiresAt != nil {
		logEvent = logEvent.Time("expiresAt", *cart.ExpiresAt)
	}

	logEvent.Msg("Cart retrieved successfully")
}

func (h *CartHandler) GetUserCart(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "GetUserCart").
		Logger()

	logger.Debug().Msg("Processing get user cart request")

	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	logger.Debug().Str("userIdRaw", userIDStr).Msg("Extracting user ID from path")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userIdRaw", userIDStr).
			Msg("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("userId", userID.String()).
		Msg("Parsed user ID, fetching user's cart")

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true
	logger.Debug().
		Bool("includeItems", includeItems).
		Msg("Include items in response")

	cart, err := h.cartRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Info().
				Str("userId", userID.String()).
				Msg("Cart not found for user, creating a new cart")

			// Create a new cart for the user
			cart = &models.Cart{
				UserID: &userID,
			}

			logger.Debug().
				Str("userId", userID.String()).
				Msg("Creating new cart in database")

			err = h.cartRepo.Create(r.Context(), cart)
			if err != nil {
				logger.Error().
					Err(err).
					Str("userId", userID.String()).
					Msg("Failed to create cart for user")
				http.Error(w, "Failed to create cart for user", http.StatusInternalServerError)
				return
			}

			logger.Info().
				Str("userId", userID.String()).
				Str("cartId", cart.ID.String()).
				Msg("New cart created successfully for user")
		} else {
			logger.Error().
				Err(err).
				Str("userId", userID.String()).
				Msg("Failed to get user cart from database")
			http.Error(w, "Failed to get user cart", http.StatusInternalServerError)
			return
		}
	} else {
		logger.Debug().
			Str("userId", userID.String()).
			Str("cartId", cart.ID.String()).
			Msg("Found existing cart for user")
	}

	logger.Debug().
		Str("userId", userID.String()).
		Str("cartId", cart.ID.String()).
		Bool("includeItems", includeItems).
		Msg("Generating cart response")

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userId", userID.String()).
			Str("cartId", cart.ID.String()).
			Msg("Failed to generate cart response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("userId", userID.String()).
		Str("cartId", cart.ID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("userId", userID.String()).
			Str("cartId", cart.ID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Construct a final success log with appropriate context
	logEvent := logger.Info().
		Str("userId", userID.String()).
		Str("cartId", cart.ID.String())

	if cart.SessionID != nil {
		logEvent = logEvent.Str("sessionId", *cart.SessionID)
	}

	if cart.ExpiresAt != nil {
		logEvent = logEvent.Time("expiresAt", *cart.ExpiresAt)
	}

	logEvent.Bool("includeItems", includeItems).Msg("User cart retrieved successfully")
}

func (h *CartHandler) GetSessionCart(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "GetSessionCart").
		Logger()

	logger.Debug().Msg("Processing get session cart request")

	vars := mux.Vars(r)
	sessionID := vars["session_id"]
	logger.Debug().Str("sessionId", sessionID).Msg("Extracting session ID from path")

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true
	logger.Debug().
		Bool("includeItems", includeItems).
		Msg("Include items in response")

	cart, err := h.cartRepo.GetBySessionID(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Info().
				Str("sessionId", sessionID).
				Msg("Cart not found for session, creating a new cart")

			// Create a new cart for the session
			cart = &models.Cart{
				SessionID: &sessionID,
			}

			logger.Debug().
				Str("sessionId", sessionID).
				Msg("Creating new cart in database")

			err = h.cartRepo.Create(r.Context(), cart)
			if err != nil {
				logger.Error().
					Err(err).
					Str("sessionId", sessionID).
					Msg("Failed to create cart for session")
				http.Error(w, "Failed to create cart for session", http.StatusInternalServerError)
				return
			}

			logger.Info().
				Str("sessionId", sessionID).
				Str("cartId", cart.ID.String()).
				Msg("New cart created successfully for session")
		} else {
			logger.Error().
				Err(err).
				Str("sessionId", sessionID).
				Msg("Failed to get session cart from database")
			http.Error(w, "Failed to get session cart", http.StatusInternalServerError)
			return
		}
	} else {
		logger.Debug().
			Str("sessionId", sessionID).
			Str("cartId", cart.ID.String()).
			Msg("Found existing cart for session")
	}

	logger.Debug().
		Str("sessionId", sessionID).
		Str("cartId", cart.ID.String()).
		Bool("includeItems", includeItems).
		Msg("Generating cart response")

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		logger.Error().
			Err(err).
			Str("sessionId", sessionID).
			Str("cartId", cart.ID.String()).
			Msg("Failed to generate cart response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("sessionId", sessionID).
		Str("cartId", cart.ID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("sessionId", sessionID).
			Str("cartId", cart.ID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Construct a final success log with appropriate context
	logEvent := logger.Info().
		Str("sessionId", sessionID).
		Str("cartId", cart.ID.String())

	if cart.UserID != nil {
		logEvent = logEvent.Str("userId", cart.UserID.String())
	}

	if cart.ExpiresAt != nil {
		logEvent = logEvent.Time("expiresAt", *cart.ExpiresAt)
	}

	logEvent.Bool("includeItems", includeItems).Msg("Session cart retrieved successfully")
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "AddToCart").
		Logger()

	logger.Debug().Msg("Processing add to cart request")

	vars := mux.Vars(r)
	cartIDStr := vars["id"]
	logger.Debug().Str("cartIdRaw", cartIDStr).Msg("Extracting cart ID from path")

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartIdRaw", cartIDStr).
			Msg("Invalid cart ID format")
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Parsed cart ID, decoding request body")

	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Bool("hasPriceId", req.PriceID != nil).
		Int("quantity", req.Quantity).
		Bool("isSubscription", req.IsSubscription).
		Bool("hasOptions", len(req.Options) > 0).
		Msg("Received add to cart request")

	if req.PriceID != nil {
		logger.Debug().Str("priceId", req.PriceID.String()).Msg("Price ID specified")
	}

	// Basic validation
	if req.ProductID == uuid.Nil {
		logger.Error().
			Str("cartId", cartID.String()).
			Msg("Product ID is required")
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Int("quantity", req.Quantity).
			Msg("Quantity must be greater than zero")
		http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
		return
	}

	// Verify cart exists
	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Verifying cart exists")
	_, err = h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Msg("Cart not found")
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to verify cart")
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Verify product exists and get price
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Msg("Verifying product exists")
	product, err := h.productRepo.GetByID(r.Context(), req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("productId", req.ProductID.String()).
				Msg("Product not found")
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Msg("Failed to verify product")
		http.Error(w, "Failed to verify product", http.StatusInternalServerError)
		return
	}

	// Determine price to use
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Msg("Determining price to use")
	var price *models.Price
	var unitPrice int64 = 0

	if req.PriceID != nil {
		// Use the specified price if provided
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Str("priceId", req.PriceID.String()).
			Msg("Using specified price")
		price, err = h.priceRepo.GetByID(r.Context(), *req.PriceID)
		if err != nil {
			if errors.Is(err, repository.ErrPriceNotFound) {
				logger.Error().
					Err(err).
					Str("cartId", cartID.String()).
					Str("productId", req.ProductID.String()).
					Str("priceId", req.PriceID.String()).
					Msg("Price not found")
				http.Error(w, "Price not found", http.StatusNotFound)
				return
			}
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("productId", req.ProductID.String()).
				Str("priceId", req.PriceID.String()).
				Msg("Failed to verify price")
			http.Error(w, "Failed to verify price", http.StatusInternalServerError)
			return
		}
		unitPrice = price.Amount
	} else if req.IsSubscription {
		// For subscription items, we need a price
		logger.Error().
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Bool("isSubscription", req.IsSubscription).
			Msg("Price ID is required for subscription items")
		http.Error(w, "Price ID is required for subscription items", http.StatusBadRequest)
		return
	} else {
		// For one-time purchases, we can use a default price if it exists
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Msg("Looking for default price for one-time purchase")
		prices, err := h.priceRepo.ListByProductID(r.Context(), product.ID, true)
		if err == nil && len(prices) > 0 {
			price = prices[0]
			unitPrice = price.Amount
			logger.Debug().
				Str("cartId", cartID.String()).
				Str("productId", req.ProductID.String()).
				Str("defaultPriceId", price.ID.String()).
				Int64("unitPrice", unitPrice).
				Msg("Using default price")
		} else {
			// If no price found (which shouldn't happen in practice), use a fallback
			logger.Error().
				Str("cartId", cartID.String()).
				Str("productId", req.ProductID.String()).
				Msg("No price available for this product")
			http.Error(w, "No price available for this product", http.StatusBadRequest)
			return
		}
	}

	// Create the cart item
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Int64("unitPrice", unitPrice).
		Int("quantity", req.Quantity).
		Bool("isSubscription", req.IsSubscription).
		Msg("Creating cart item")
	cartItem := &models.CartItem{
		CartID:         cartID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		UnitPrice:      unitPrice,
		IsSubscription: req.IsSubscription,
	}

	if price != nil && price.ID != uuid.Nil {
		cartItem.PriceID = &price.ID
	}

	if len(req.Options) > 0 {
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Interface("options", req.Options).
			Msg("Adding options to cart item")
		cartItem.Options = &req.Options
	}

	err = h.cartItemRepo.Create(r.Context(), cartItem)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("productId", req.ProductID.String()).
			Msg("Failed to add item to cart")
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Str("cartItemId", cartItem.ID.String()).
		Msg("Item added to cart successfully, retrieving updated cart")

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to get updated cart")
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Generating cart response")
	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to generate response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Final success log with context
	totalItems := len(updatedCart.Items)
	var totalQuantity int
	var totalValue int64

	for _, item := range updatedCart.Items {
		totalQuantity += item.Quantity
		totalValue += item.UnitPrice * int64(item.Quantity)
	}

	logger.Info().
		Str("cartId", cartID.String()).
		Str("productId", req.ProductID.String()).
		Str("cartItemId", cartItem.ID.String()).
		Int("totalItems", totalItems).
		Int("totalQuantity", totalQuantity).
		Int64("totalValue", totalValue).
		Msg("Item added to cart successfully")
}

func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "UpdateCartItem").
		Logger()

	logger.Debug().Msg("Processing update cart item request")

	vars := mux.Vars(r)
	cartIDStr := vars["id"]
	itemIDStr := vars["item_id"]

	logger.Debug().
		Str("cartIdRaw", cartIDStr).
		Str("itemIdRaw", itemIDStr).
		Msg("Extracting IDs from path")

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartIdRaw", cartIDStr).
			Msg("Invalid cart ID format")
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		logger.Error().
			Err(err).
			Str("itemIdRaw", itemIDStr).
			Str("cartId", cartID.String()).
			Msg("Invalid item ID format")
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Parsing request body")

	var req UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log what fields are being updated
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Bool("updatingQuantity", req.Quantity != nil).
		Bool("updatingSubscriptionStatus", req.IsSubscription != nil).
		Bool("updatingOptions", req.Options != nil).
		Msg("Received update cart item request")

	if req.Quantity != nil {
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Int("newQuantity", *req.Quantity).
			Msg("Updating quantity")
	}

	if req.IsSubscription != nil {
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Bool("newIsSubscription", *req.IsSubscription).
			Msg("Updating subscription status")
	}

	// Verify cart exists
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Verifying cart exists")

	_, err = h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Msg("Cart not found")
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Failed to verify cart")
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Get the cart item
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Retrieving cart item")

	cartItem, err := h.cartItemRepo.GetByID(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Msg("Cart item not found")
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Failed to get cart item")
		http.Error(w, "Failed to get cart item", http.StatusInternalServerError)
		return
	}

	// Verify the item belongs to the cart
	if cartItem.CartID != cartID {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Str("itemCartId", cartItem.CartID.String()).
			Msg("Cart item does not belong to the specified cart")
		http.Error(w, "Cart item does not belong to the specified cart", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Str("productId", cartItem.ProductID.String()).
		Int("currentQuantity", cartItem.Quantity).
		Int64("unitPrice", cartItem.UnitPrice).
		Bool("currentIsSubscription", cartItem.IsSubscription).
		Msg("Found cart item, applying updates")

	// Update fields if provided
	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			// If quantity is set to 0 or negative, remove the item
			logger.Info().
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Str("productId", cartItem.ProductID.String()).
				Int("newQuantity", *req.Quantity).
				Msg("Removing item from cart due to zero or negative quantity")

			err = h.cartItemRepo.Delete(r.Context(), itemID)
			if err != nil {
				logger.Error().
					Err(err).
					Str("cartId", cartID.String()).
					Str("itemId", itemID.String()).
					Msg("Failed to remove cart item")
				http.Error(w, "Failed to remove cart item", http.StatusInternalServerError)
				return
			}

			logger.Info().
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Str("productId", cartItem.ProductID.String()).
				Msg("Cart item removed successfully")

			// Return the updated cart
			logger.Debug().
				Str("cartId", cartID.String()).
				Msg("Retrieving updated cart after item removal")

			updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
			if err != nil {
				logger.Error().
					Err(err).
					Str("cartId", cartID.String()).
					Msg("Failed to get updated cart")
				http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
				return
			}

			logger.Debug().
				Str("cartId", cartID.String()).
				Int("remainingItems", len(updatedCart.Items)).
				Msg("Generating cart response after item removal")

			response, err := h.cartToResponse(r.Context(), updatedCart, true)
			if err != nil {
				logger.Error().
					Err(err).
					Str("cartId", cartID.String()).
					Msg("Failed to generate response")
				http.Error(w, "Failed to generate response", http.StatusInternalServerError)
				return
			}

			logger.Debug().
				Str("cartId", cartID.String()).
				Msg("Response generated, sending to client")

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Error().
					Err(err).
					Str("cartId", cartID.String()).
					Msg("Failed to encode JSON response")
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}

			// Log final success for item removal path
			logger.Info().
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Str("productId", cartItem.ProductID.String()).
				Int("remainingItems", len(updatedCart.Items)).
				Msg("Cart updated successfully after item removal")
			return
		}

		logger.Debug().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Int("oldQuantity", cartItem.Quantity).
			Int("newQuantity", *req.Quantity).
			Msg("Updating item quantity")

		cartItem.Quantity = *req.Quantity
	}

	if req.IsSubscription != nil {
		// If changing to subscription and no price ID
		if *req.IsSubscription && cartItem.PriceID == nil {
			logger.Error().
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Str("productId", cartItem.ProductID.String()).
				Bool("newIsSubscription", *req.IsSubscription).
				Msg("Price ID is required for subscription items")
			http.Error(w, "Price ID is required for subscription items", http.StatusBadRequest)
			return
		}

		logger.Debug().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Bool("oldIsSubscription", cartItem.IsSubscription).
			Bool("newIsSubscription", *req.IsSubscription).
			Msg("Updating subscription status")

		cartItem.IsSubscription = *req.IsSubscription
	}

	if req.Options != nil {
		logger.Debug().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Interface("newOptions", *req.Options).
			Msg("Updating item options")

		cartItem.Options = req.Options
	}

	// Update the cart item
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Updating cart item in database")

	err = h.cartItemRepo.Update(r.Context(), cartItem)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Failed to update cart item")
		http.Error(w, "Failed to update cart item", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Cart item updated successfully, retrieving updated cart")

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to get updated cart")
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Int("totalItems", len(updatedCart.Items)).
		Msg("Generating cart response")

	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to generate response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
		Str("cartId", cartID.String()).
		Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Msg("Failed to encode JSON response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Final success log with context
	var totalQuantity int
	var totalValue int64

	for _, item := range updatedCart.Items {
		totalQuantity += item.Quantity
		totalValue += item.UnitPrice * int64(item.Quantity)
	}

	logger.Info().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Str("productId", cartItem.ProductID.String()).
		Int("quantity", cartItem.Quantity).
		Bool("isSubscription", cartItem.IsSubscription).
		Int("totalItems", len(updatedCart.Items)).
		Int("totalQuantity", totalQuantity).
		Int64("totalValue", totalValue).
		Msg("Cart item updated successfully")
}

func (h *CartHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("handler", "CartHandler").
		Str("method", "RemoveCartItem").
		Logger()

	logger.Debug().Msg("Processing remove cart item request")

	vars := mux.Vars(r)
	cartIDStr := vars["id"]
	itemIDStr := vars["item_id"]

	logger.Debug().
		Str("cartIdRaw", cartIDStr).
		Str("itemIdRaw", itemIDStr).
		Msg("Extracting IDs from path")

	cartID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(vars["item_id"])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Verify cart exists
	logger.Debug().
		Str("cartId", cartID.String()).
		Str("itemId", itemID.String()).
		Msg("Verifying cart exists")

	_, err = h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Msg("Cart not found")
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Failed to verify cart")
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Get the cart item to verify ownership
	cartItem, err := h.cartItemRepo.GetByID(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			logger.Error().
				Err(err).
				Str("cartId", cartID.String()).
				Str("itemId", itemID.String()).
				Msg("Cart item not found")
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}
		logger.Error().
			Err(err).
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Msg("Failed to get cart item")
		http.Error(w, "Failed to get cart item", http.StatusInternalServerError)
		return
	}

	// Verify the item belongs to the cart
	if cartItem.CartID != cartID {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Str("itemCartId", cartItem.CartID.String()).
			Msg("Cart item does not belong to the specified cart")
		http.Error(w, "Cart item does not belong to the specified cart", http.StatusBadRequest)
		return
	}

	// Delete the cart item
	err = h.cartItemRepo.Delete(r.Context(), itemID)
	if err != nil {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Str("itemCartId", cartItem.CartID.String()).
			Msg("Failed to remove cart item")
		http.Error(w, "Failed to remove cart item", http.StatusInternalServerError)
		return
	}

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Str("itemCartId", cartItem.CartID.String()).
			Msg("Failed to get updated cart")
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		logger.Error().
			Str("cartId", cartID.String()).
			Str("itemId", itemID.String()).
			Str("itemCartId", cartItem.CartID.String()).
			Msg("Failed to generate response")
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	logger.Debug().
	Str("cartId", cartID.String()).
	Msg("Response generated, sending to client")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	// Verify cart exists
	cart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Delete all items in the cart
	err = h.cartItemRepo.DeleteByCartID(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), cart, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	// Verify cart exists
	_, err = h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Delete the cart (this will cascade to delete cart items due to foreign key constraint)
	err = h.cartRepo.Delete(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Failed to delete cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) CleanExpiredCarts(w http.ResponseWriter, r *http.Request) {
	// This endpoint should be protected by admin authentication

	count, err := h.cartRepo.CleanExpiredCarts(r.Context())
	if err != nil {
		http.Error(w, "Failed to clean expired carts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"deleted_count": count})
}

// Helper functions

func (h *CartHandler) cartToResponse(ctx context.Context, cart *models.Cart, includeItems bool) (CartResponse, error) {
	response := CartResponse{
		ID:        cart.ID,
		UserID:    cart.UserID,
		SessionID: cart.SessionID,
		CreatedAt: cart.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: cart.UpdatedAt.Format(http.TimeFormat),
	}

	if cart.ExpiresAt != nil {
		formatted := cart.ExpiresAt.Format(http.TimeFormat)
		response.ExpiresAt = &formatted
	}

	if cart.Metadata != nil {
		response.Metadata = *cart.Metadata
	}

	// Include cart items if requested
	if includeItems {
		items, err := h.cartItemRepo.ListByCartID(ctx, cart.ID)
		if err != nil {
			return response, err
		}

		var totalPrice int64 = 0
		var itemResponses []CartItemResponse

		for _, item := range items {
			itemResponse, err := h.cartItemToResponse(ctx, item, true)
			if err != nil {
				return response, err
			}

			totalPrice += itemResponse.TotalPrice
			itemResponses = append(itemResponses, itemResponse)
		}

		response.Items = itemResponses
		response.ItemCount = len(itemResponses)
		response.TotalPrice = totalPrice
	}

	return response, nil
}

func (h *CartHandler) cartItemToResponse(ctx context.Context, item *models.CartItem, includeProduct bool) (CartItemResponse, error) {
	response := CartItemResponse{
		ID:             item.ID,
		ProductID:      item.ProductID,
		PriceID:        item.PriceID,
		Quantity:       item.Quantity,
		UnitPrice:      item.UnitPrice,
		TotalPrice:     item.UnitPrice * int64(item.Quantity),
		IsSubscription: item.IsSubscription,
		CreatedAt:      item.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:      item.UpdatedAt.Format(http.TimeFormat),
	}

	if item.Options != nil {
		response.Options = *item.Options
	}

	// Include product details if requested
	if includeProduct {
		product, err := h.productRepo.GetByID(ctx, item.ProductID)
		if err == nil {
			productHandler := NewProductHandler(h.productRepo)
			productResponse, err := productHandler.modelToResponse(product)
			if err == nil {
				response.Product = &productResponse
			}
		}

		// Include price details if available
		if item.PriceID != nil {
			price, err := h.priceRepo.GetByID(ctx, *item.PriceID)
			if err == nil {
				priceHandler := NewPriceHandler(h.priceRepo, h.productRepo)
				priceResponse, err := priceHandler.modelToResponse(price, false)
				if err == nil {
					response.Price = &priceResponse
				}
			}
		}
	}

	return response, nil
}

func (h *CartHandler) parseTimeStr(timeStr string) (time.Time, error) {
	return time.Parse(http.TimeFormat, timeStr)
}
