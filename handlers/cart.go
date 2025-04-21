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

// Handlers

func (h *CartHandler) CreateCart(w http.ResponseWriter, r *http.Request) {
	var req CreateCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID == nil && req.SessionID == nil {
		http.Error(w, "Either user_id or session_id is required", http.StatusBadRequest)
		return
	}

	// Create the cart
	cart := &models.Cart{
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}

	if req.ExpiresAt != nil {
		expires, err := h.parseTimeStr(*req.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid date format for expires_at", http.StatusBadRequest)
			return
		}
		cart.ExpiresAt = &expires
	}

	if len(req.Metadata) > 0 {
		cart.Metadata = &req.Metadata
	}

	err := h.cartRepo.Create(r.Context(), cart)
	if err != nil {
		http.Error(w, "Failed to create cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), cart, false)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

	cart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) GetUserCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

	cart, err := h.cartRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			// Create a new cart for the user
			cart = &models.Cart{
				UserID: &userID,
			}
			err = h.cartRepo.Create(r.Context(), cart)
			if err != nil {
				http.Error(w, "Failed to create cart for user", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to get user cart", http.StatusInternalServerError)
			return
		}
	}

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) GetSessionCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

	cart, err := h.cartRepo.GetBySessionID(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			// Create a new cart for the session
			cart = &models.Cart{
				SessionID: &sessionID,
			}
			err = h.cartRepo.Create(r.Context(), cart)
			if err != nil {
				http.Error(w, "Failed to create cart for session", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to get session cart", http.StatusInternalServerError)
			return
		}
	}

	response, err := h.cartToResponse(r.Context(), cart, includeItems)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cartID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid cart ID", http.StatusBadRequest)
		return
	}

	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.ProductID == uuid.Nil {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be greater than zero", http.StatusBadRequest)
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

	// Verify product exists and get price
	product, err := h.productRepo.GetByID(r.Context(), req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify product", http.StatusInternalServerError)
		return
	}

	// Determine price to use
	var price *models.Price
	var unitPrice int64 = 0

	if req.PriceID != nil {
		// Use the specified price if provided
		price, err = h.priceRepo.GetByID(r.Context(), *req.PriceID)
		if err != nil {
			if errors.Is(err, repository.ErrPriceNotFound) {
				http.Error(w, "Price not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to verify price", http.StatusInternalServerError)
			return
		}
		unitPrice = price.Amount
	} else if req.IsSubscription {
		// For subscription items, we need a price
		http.Error(w, "Price ID is required for subscription items", http.StatusBadRequest)
		return
	} else {
		// For one-time purchases, we can use a default price if it exists
		prices, err := h.priceRepo.ListByProductID(r.Context(), product.ID, true)
		if err == nil && len(prices) > 0 {
			price = prices[0]
			unitPrice = price.Amount
		} else {
			// If no price found (which shouldn't happen in practice), use a fallback
			http.Error(w, "No price available for this product", http.StatusBadRequest)
			return
		}
	}

	// Create the cart item
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
		cartItem.Options = &req.Options
	}

	err = h.cartItemRepo.Create(r.Context(), cartItem)
	if err != nil {
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
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

	var req UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	// Get the cart item
	cartItem, err := h.cartItemRepo.GetByID(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get cart item", http.StatusInternalServerError)
		return
	}

	// Verify the item belongs to the cart
	if cartItem.CartID != cartID {
		http.Error(w, "Cart item does not belong to the specified cart", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			// If quantity is set to 0 or negative, remove the item
			err = h.cartItemRepo.Delete(r.Context(), itemID)
			if err != nil {
				http.Error(w, "Failed to remove cart item", http.StatusInternalServerError)
				return
			}

			// Return the updated cart
			updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
			if err != nil {
				http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
				return
			}

			response, err := h.cartToResponse(r.Context(), updatedCart, true)
			if err != nil {
				http.Error(w, "Failed to generate response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		cartItem.Quantity = *req.Quantity
	}

	if req.IsSubscription != nil {
		// If changing to subscription and no price ID
		if *req.IsSubscription && cartItem.PriceID == nil {
			http.Error(w, "Price ID is required for subscription items", http.StatusBadRequest)
			return
		}
		cartItem.IsSubscription = *req.IsSubscription
	}

	if req.Options != nil {
		cartItem.Options = req.Options
	}

	// Update the cart item
	err = h.cartItemRepo.Update(r.Context(), cartItem)
	if err != nil {
		http.Error(w, "Failed to update cart item", http.StatusInternalServerError)
		return
	}

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CartHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
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
	_, err = h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify cart", http.StatusInternalServerError)
		return
	}

	// Get the cart item to verify ownership
	cartItem, err := h.cartItemRepo.GetByID(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get cart item", http.StatusInternalServerError)
		return
	}

	// Verify the item belongs to the cart
	if cartItem.CartID != cartID {
		http.Error(w, "Cart item does not belong to the specified cart", http.StatusBadRequest)
		return
	}

	// Delete the cart item
	err = h.cartItemRepo.Delete(r.Context(), itemID)
	if err != nil {
		http.Error(w, "Failed to remove cart item", http.StatusInternalServerError)
		return
	}

	// Return the updated cart
	updatedCart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Failed to get updated cart", http.StatusInternalServerError)
		return
	}

	response, err := h.cartToResponse(r.Context(), updatedCart, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

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
