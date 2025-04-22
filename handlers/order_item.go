// handlers/order_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	
	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

type OrderHandler struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	cartRepo      repository.CartRepository
	cartItemRepo  repository.CartItemRepository
	userRepo      repository.UserRepository
	productRepo   repository.ProductRepository
	priceRepo     repository.PriceRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewOrderHandler(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	cartRepo repository.CartRepository,
	cartItemRepo repository.CartItemRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	priceRepo repository.PriceRepository,
	subscriptionRepo repository.SubscriptionRepository,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		cartRepo:      cartRepo,
		cartItemRepo:  cartItemRepo,
		userRepo:      userRepo,
		productRepo:   productRepo,
		priceRepo:     priceRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// Request/Response structs

type AddressRequest struct {
	Name        string `json:"name"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2,omitempty"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type CreateOrderRequest struct {
	UserID          uuid.UUID      `json:"user_id"`
	CartID          uuid.UUID      `json:"cart_id"`
	ShippingAddress AddressRequest `json:"shipping_address"`
	BillingAddress  AddressRequest `json:"billing_address"`
	PaymentIntentID string         `json:"payment_intent_id"`
	Currency        string         `json:"currency"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateOrderRequest struct {
	Status          *string                 `json:"status,omitempty"`
	ShippingAddress *AddressRequest         `json:"shipping_address,omitempty"`
	BillingAddress  *AddressRequest         `json:"billing_address,omitempty"`
	PaymentIntentID *string                 `json:"payment_intent_id,omitempty"`
	CompletedAt     *string                 `json:"completed_at,omitempty"`
	Metadata        *map[string]interface{} `json:"metadata,omitempty"`
}

type OrderItemResponse struct {
	ID             uuid.UUID               `json:"id"`
	ProductID      uuid.UUID               `json:"product_id"`
	PriceID        *uuid.UUID              `json:"price_id,omitempty"`
	SubscriptionID *uuid.UUID              `json:"subscription_id,omitempty"`
	Quantity       int                     `json:"quantity"`
	UnitPrice      int64                   `json:"unit_price"`
	TotalPrice     int64                   `json:"total_price"`
	IsSubscription bool                    `json:"is_subscription"`
	Options        map[string]interface{}  `json:"options,omitempty"`
	Product        *ProductResponse        `json:"product,omitempty"`
	Price          *PriceResponse          `json:"price,omitempty"`
	Subscription   *SubscriptionResponse   `json:"subscription,omitempty"`
}

type OrderResponse struct {
	ID               uuid.UUID               `json:"id"`
	UserID           *uuid.UUID              `json:"user_id,omitempty"`
	Status           string                  `json:"status"`
	TotalAmount      int64                   `json:"total_amount"`
	Currency         string                  `json:"currency"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
	CompletedAt      *string                 `json:"completed_at,omitempty"`
	ShippingAddress  *models.Address         `json:"shipping_address,omitempty"`
	BillingAddress   *models.Address         `json:"billing_address,omitempty"`
	PaymentIntentID  *string                 `json:"payment_intent_id,omitempty"`
	StripeCustomerID *string                 `json:"stripe_customer_id,omitempty"`
	Metadata         map[string]interface{}  `json:"metadata,omitempty"`
	Items            []OrderItemResponse     `json:"items,omitempty"`
	User             *UserResponse           `json:"user,omitempty"`
}

// Helper functions

func (h *OrderHandler) orderToResponse(ctx context.Context, order *models.Order, includeItems bool, includeUser bool) (OrderResponse, error) {
	response := OrderResponse{
		ID:               order.ID,
		UserID:           order.UserID,
		Status:           string(order.Status),
		TotalAmount:      order.TotalAmount,
		Currency:         order.Currency,
		CreatedAt:        order.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:        order.UpdatedAt.Format(http.TimeFormat),
		ShippingAddress:  order.ShippingAddress,
		BillingAddress:   order.BillingAddress,
		PaymentIntentID:  order.PaymentIntentID,
		StripeCustomerID: order.StripeCustomerID,
	}

	if order.CompletedAt != nil {
		formatted := order.CompletedAt.Format(http.TimeFormat)
		response.CompletedAt = &formatted
	}

	if order.Metadata != nil {
		response.Metadata = *order.Metadata
	}

	// Include order items if requested
	if includeItems {
		items, err := h.orderItemRepo.ListByOrderID(ctx, order.ID)
		if err != nil {
			return response, err
		}

		var itemResponses []OrderItemResponse

		for _, item := range items {
			itemResponse, err := h.orderItemToResponse(ctx, item)
			if err != nil {
				return response, err
			}

			itemResponses = append(itemResponses, itemResponse)
		}

		response.Items = itemResponses
	}

	// Include user if requested
	if includeUser && order.UserID != nil {
		user, err := h.userRepo.GetByID(ctx, *order.UserID)
		if err == nil {
			userHandler := NewUserHandler(h.userRepo)
			userResponse, err := userHandler.modelToResponse(user)
			if err == nil {
				response.User = &userResponse
			}
		}
	}

	return response, nil
}

func (h *OrderHandler) orderItemToResponse(ctx context.Context, item *models.OrderItem) (OrderItemResponse, error) {
	response := OrderItemResponse{
		ID:             item.ID,
		ProductID:      item.ProductID,
		PriceID:        item.PriceID,
		SubscriptionID: item.SubscriptionID,
		Quantity:       item.Quantity,
		UnitPrice:      item.UnitPrice,
		TotalPrice:     item.TotalPrice,
		IsSubscription: item.IsSubscription,
	}

	if item.Options != nil {
		response.Options = *item.Options
	}

	// Include product details
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

	// Include subscription details if available
	if item.SubscriptionID != nil {
		subscription, err := h.subscriptionRepo.GetByID(context.Background(), *item.SubscriptionID)
		if err == nil {
			// Convert subscription model to response type
			subscriptionResponse := formatSubscriptionResponse(subscription)
			response.Subscription = &subscriptionResponse
		}
	}

	return response, nil
}

func (h *OrderHandler) parseTimeStr(timeStr string) (time.Time, error) {
	return time.Parse(http.TimeFormat, timeStr)
}

// Handlers

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID == uuid.Nil {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if req.CartID == uuid.Nil {
		http.Error(w, "Cart ID is required", http.StatusBadRequest)
		return
	}

	if req.ShippingAddress.Name == "" || req.ShippingAddress.Line1 == "" ||
		req.ShippingAddress.City == "" || req.ShippingAddress.State == "" ||
		req.ShippingAddress.PostalCode == "" || req.ShippingAddress.Country == "" {
		http.Error(w, "Complete shipping address is required", http.StatusBadRequest)
		return
	}

	if req.BillingAddress.Name == "" || req.BillingAddress.Line1 == "" ||
		req.BillingAddress.City == "" || req.BillingAddress.State == "" ||
		req.BillingAddress.PostalCode == "" || req.BillingAddress.Country == "" {
		http.Error(w, "Complete billing address is required", http.StatusBadRequest)
		return
	}

	if req.PaymentIntentID == "" {
		http.Error(w, "Payment intent ID is required", http.StatusBadRequest)
		return
	}

	// Verify user exists
	user, err := h.userRepo.GetByID(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	// Get cart and cart items
	_, err = h.cartRepo.GetByID(r.Context(), req.CartID)
	if err != nil {
		if errors.Is(err, repository.ErrCartNotFound) {
			http.Error(w, "Cart not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}

	cartItems, err := h.cartItemRepo.ListByCartID(r.Context(), req.CartID)
	if err != nil {
		http.Error(w, "Failed to get cart items", http.StatusInternalServerError)
		return
	}

	if len(cartItems) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	// Calculate total amount
	var totalAmount int64 = 0
	for _, item := range cartItems {
		totalAmount += item.UnitPrice * int64(item.Quantity)
	}

	// Create the order
	order := &models.Order{
		UserID:      &req.UserID,
		Status:      models.OrderStatusPending,
		TotalAmount: totalAmount,
		Currency:    req.Currency,
		ShippingAddress: &models.Address{
			Name:        req.ShippingAddress.Name,
			Line1:       req.ShippingAddress.Line1,
			Line2:       req.ShippingAddress.Line2,
			City:        req.ShippingAddress.City,
			State:       req.ShippingAddress.State,
			PostalCode:  req.ShippingAddress.PostalCode,
			Country:     req.ShippingAddress.Country,
			PhoneNumber: req.ShippingAddress.PhoneNumber,
		},
		BillingAddress: &models.Address{
			Name:        req.BillingAddress.Name,
			Line1:       req.BillingAddress.Line1,
			Line2:       req.BillingAddress.Line2,
			City:        req.BillingAddress.City,
			State:       req.BillingAddress.State,
			PostalCode:  req.BillingAddress.PostalCode,
			Country:     req.BillingAddress.Country,
			PhoneNumber: req.BillingAddress.PhoneNumber,
		},
		PaymentIntentID: &req.PaymentIntentID,
	}

	// Set stripe customer ID if available
	if user.StripeCustomerID != nil {
		order.StripeCustomerID = user.StripeCustomerID
	}

	if len(req.Metadata) > 0 {
		order.Metadata = &req.Metadata
	}

	// Create the order in the database
	err = h.orderRepo.Create(r.Context(), order)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Create order items
	for _, cartItem := range cartItems {
		orderItem := &models.OrderItem{
			OrderID:        order.ID,
			ProductID:      cartItem.ProductID,
			PriceID:        cartItem.PriceID,
			Quantity:       cartItem.Quantity,
			UnitPrice:      cartItem.UnitPrice,
			TotalPrice:     cartItem.UnitPrice * int64(cartItem.Quantity),
			IsSubscription: cartItem.IsSubscription,
			Options:        cartItem.Options,
		}

		// If this is a subscription item, create a subscription
		if cartItem.IsSubscription && cartItem.PriceID != nil {
			// Creating a subscription would normally involve Stripe API calls
			// Here we'll just create a basic record in our database
			subscription := &models.Subscription{
				UserID:            req.UserID,
				PriceID:           *cartItem.PriceID,
				Quantity:          cartItem.Quantity,
				Status:            models.SubscriptionStatusActive,
				CurrentPeriodStart: time.Now(),
				CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0), // Default to 1 month
				StripeSubscriptionID: "sub_" + uuid.New().String(), // Placeholder for real Stripe ID
				StripeCustomerID:   *user.StripeCustomerID,
				CollectionMethod:   "charge_automatically",
			}

			err = h.subscriptionRepo.Create(r.Context(), subscription)
			if err != nil {
				http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
				return
			}

			// Link the order item to the subscription
			orderItem.SubscriptionID = &subscription.ID
		}

		err = h.orderItemRepo.Create(r.Context(), orderItem)
		if err != nil {
			http.Error(w, "Failed to create order item", http.StatusInternalServerError)
			return
		}
	}

	// Clear the cart after successful order creation
	err = h.cartItemRepo.DeleteByCartID(r.Context(), req.CartID)
	if err != nil {
		// Log the error but don't fail the request
		// The order has been created successfully
		// TODO: Implement proper logging
		fmt.Printf("Warning: Failed to clear cart after order creation: %v\n", err)
	}

	response, err := h.orderToResponse(r.Context(), order, true, true)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true
	includeUser := r.URL.Query().Get("include_user") == "true" // Default to false

	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	response, err := h.orderToResponse(r.Context(), order, includeItems, includeUser)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing order
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Update fields if provided
	if req.Status != nil {
		order.Status = models.OrderStatus(*req.Status)

		// If status is completed and there's no completed_at timestamp, set it
		if order.Status == models.OrderStatusCompleted && order.CompletedAt == nil {
			now := time.Now()
			order.CompletedAt = &now
		}
	}

	if req.ShippingAddress != nil {
		order.ShippingAddress = &models.Address{
			Name:        req.ShippingAddress.Name,
			Line1:       req.ShippingAddress.Line1,
			Line2:       req.ShippingAddress.Line2,
			City:        req.ShippingAddress.City,
			State:       req.ShippingAddress.State,
			PostalCode:  req.ShippingAddress.PostalCode,
			Country:     req.ShippingAddress.Country,
			PhoneNumber: req.ShippingAddress.PhoneNumber,
		}
	}

	if req.BillingAddress != nil {
		order.BillingAddress = &models.Address{
			Name:        req.BillingAddress.Name,
			Line1:       req.BillingAddress.Line1,
			Line2:       req.BillingAddress.Line2,
			City:        req.BillingAddress.City,
			State:       req.BillingAddress.State,
			PostalCode:  req.BillingAddress.PostalCode,
			Country:     req.BillingAddress.Country,
			PhoneNumber: req.BillingAddress.PhoneNumber,
		}
	}

	if req.PaymentIntentID != nil {
		order.PaymentIntentID = req.PaymentIntentID
	}

	if req.CompletedAt != nil {
		completedAt, err := h.parseTimeStr(*req.CompletedAt)
		if err != nil {
			http.Error(w, "Invalid date format for completed_at", http.StatusBadRequest)
			return
		}
		order.CompletedAt = &completedAt
	}

	if req.Metadata != nil {
		order.Metadata = req.Metadata
	}

	// Update the order
	err = h.orderRepo.Update(r.Context(), order)
	if err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	response, err := h.orderToResponse(r.Context(), order, true, false)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Get the existing order
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Can only cancel pending or processing orders
	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusProcessing {
		http.Error(w, "Only pending or processing orders can be canceled", http.StatusBadRequest)
		return
	}

	// Update the order status
	order.Status = models.OrderStatusCanceled
	order.UpdatedAt = time.Now()

	// Update the order
	err = h.orderRepo.Update(r.Context(), order)
	if err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	// Cancel any subscriptions associated with this order
	orderItems, err := h.orderItemRepo.ListByOrderID(r.Context(), orderID)
	if err == nil {
		for _, item := range orderItems {
			if item.SubscriptionID != nil {
				// In a real application, you would call Stripe API here
				// For now, we'll just update our database
				subscription, err := h.subscriptionRepo.GetByID(r.Context(), *item.SubscriptionID)
				if err == nil {
					now := time.Now()
					subscription.Status = models.SubscriptionStatusCanceled
					subscription.CanceledAt = &now
					subscription.EndedAt = &now
					h.subscriptionRepo.Update(r.Context(), subscription)
				}
			}
		}
	}

	response, err := h.orderToResponse(r.Context(), order, true, false)
	if err != nil {
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Get the order to check its status
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	// Only allow deletion of pending or canceled orders
	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusCanceled {
		http.Error(w, "Only pending or canceled orders can be deleted", http.StatusBadRequest)
		return
	}

	// Delete the order (this will cascade to delete order items due to foreign key constraint)
	err = h.orderRepo.Delete(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

	limit := 10 // Default
	offset := 0 // Default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Verify user exists
	_, err = h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	// Get the user's orders
	orders, err := h.orderRepo.ListByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to list user orders", http.StatusInternalServerError)
		return
	}

	// Convert to response objects
	var responses []OrderResponse
	for _, order := range orders {
		response, err := h.orderToResponse(r.Context(), order, includeItems, false)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	status := r.URL.Query().Get("status")
	includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

	limit := 10 // Default
	offset := 0 // Default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var orders []*models.Order
	var err error

	// Filter by status if provided
	if status != "" {
		orders, err = h.orderRepo.ListByStatus(r.Context(), models.OrderStatus(status), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list orders by status", http.StatusInternalServerError)
			return
		}
	} else {
		// List all orders
		orders, err = h.orderRepo.List(r.Context(), limit, offset)
		if err != nil {
			http.Error(w, "Failed to list orders", http.StatusInternalServerError)
			return
		}
	}

	// Convert to response objects
	var responses []OrderResponse
	for _, order := range orders {
		response, err := h.orderToResponse(r.Context(), order, includeItems, false)
		if err != nil {
			http.Error(w, "Failed to generate response", http.StatusInternalServerError)
			return
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}