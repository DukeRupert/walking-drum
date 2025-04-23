// handlers/order_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

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

// Handlers

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "CreateOrder").
        Logger()

    logger.Debug().Msg("Processing order creation request")

    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("userId", req.UserID.String()).
        Str("cartId", req.CartID.String()).
        Str("currency", req.Currency).
        Str("paymentIntentId", req.PaymentIntentID).
        Bool("hasMetadata", len(req.Metadata) > 0).
        Msg("Received order creation request")

    // Basic validation
    if req.UserID == uuid.Nil {
        logger.Error().Msg("User ID is required")
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }

    if req.CartID == uuid.Nil {
        logger.Error().Msg("Cart ID is required")
        http.Error(w, "Cart ID is required", http.StatusBadRequest)
        return
    }

    if req.ShippingAddress.Name == "" || req.ShippingAddress.Line1 == "" ||
        req.ShippingAddress.City == "" || req.ShippingAddress.State == "" ||
        req.ShippingAddress.PostalCode == "" || req.ShippingAddress.Country == "" {
        logger.Error().Interface("shippingAddress", req.ShippingAddress).Msg("Complete shipping address is required")
        http.Error(w, "Complete shipping address is required", http.StatusBadRequest)
        return
    }

    if req.BillingAddress.Name == "" || req.BillingAddress.Line1 == "" ||
        req.BillingAddress.City == "" || req.BillingAddress.State == "" ||
        req.BillingAddress.PostalCode == "" || req.BillingAddress.Country == "" {
        logger.Error().Interface("billingAddress", req.BillingAddress).Msg("Complete billing address is required")
        http.Error(w, "Complete billing address is required", http.StatusBadRequest)
        return
    }

    if req.PaymentIntentID == "" {
        logger.Error().Msg("Payment intent ID is required")
        http.Error(w, "Payment intent ID is required", http.StatusBadRequest)
        return
    }

    // Verify user exists
    logger.Debug().Str("userId", req.UserID.String()).Msg("Verifying user exists")
    user, err := h.userRepo.GetByID(r.Context(), req.UserID)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            logger.Error().Err(err).Str("userId", req.UserID.String()).Msg("User not found")
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("userId", req.UserID.String()).Msg("Failed to verify user")
        http.Error(w, "Failed to verify user", http.StatusInternalServerError)
        return
    }

    // Get cart and cart items
    logger.Debug().Str("cartId", req.CartID.String()).Msg("Retrieving cart")
    _, err = h.cartRepo.GetByID(r.Context(), req.CartID)
    if err != nil {
        if errors.Is(err, repository.ErrCartNotFound) {
            logger.Error().Err(err).Str("cartId", req.CartID.String()).Msg("Cart not found")
            http.Error(w, "Cart not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("cartId", req.CartID.String()).Msg("Failed to get cart")
        http.Error(w, "Failed to get cart", http.StatusInternalServerError)
        return
    }

    logger.Debug().Str("cartId", req.CartID.String()).Msg("Retrieving cart items")
    cartItems, err := h.cartItemRepo.ListByCartID(r.Context(), req.CartID)
    if err != nil {
        logger.Error().Err(err).Str("cartId", req.CartID.String()).Msg("Failed to get cart items")
        http.Error(w, "Failed to get cart items", http.StatusInternalServerError)
        return
    }

    if len(cartItems) == 0 {
        logger.Error().Str("cartId", req.CartID.String()).Msg("Cart is empty")
        http.Error(w, "Cart is empty", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("cartId", req.CartID.String()).
        Int("itemCount", len(cartItems)).
        Msg("Retrieved cart items")

    // Calculate total amount
    var totalAmount int64 = 0
    for _, item := range cartItems {
        totalAmount += item.UnitPrice * int64(item.Quantity)
    }

    logger.Debug().
        Str("cartId", req.CartID.String()).
        Int64("totalAmount", totalAmount).
        Str("currency", req.Currency).
        Msg("Calculated order total amount")

    // Create the order
    logger.Debug().Msg("Creating order object")
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
        hasStripeID := user.StripeCustomerID != nil && *user.StripeCustomerID != ""
        logger.Debug().Bool("hasStripeCustomerId", hasStripeID).Msg("Setting Stripe customer ID")
        order.StripeCustomerID = user.StripeCustomerID
    }

    if len(req.Metadata) > 0 {
        logger.Debug().Interface("metadata", req.Metadata).Msg("Setting order metadata")
        order.Metadata = &req.Metadata
    }

    // Create the order in the database
    logger.Debug().Msg("Creating order in database")
    err = h.orderRepo.Create(r.Context(), order)
    if err != nil {
        logger.Error().Err(err).Msg("Failed to create order in database")
        http.Error(w, "Failed to create order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", order.ID.String()).
        Str("userId", req.UserID.String()).
        Int64("totalAmount", totalAmount).
        Str("currency", req.Currency).
        Msg("Order created, now creating order items")

    // Create order items
    for i, cartItem := range cartItems {
        logger.Debug().
            Str("orderId", order.ID.String()).
            Str("productId", cartItem.ProductID.String()).
            Int("quantity", cartItem.Quantity).
            Int64("unitPrice", cartItem.UnitPrice).
            Bool("isSubscription", cartItem.IsSubscription).
            Int("itemIndex", i).
            Msg("Creating order item")

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
            logger.Debug().
                Str("orderId", order.ID.String()).
                Str("productId", cartItem.ProductID.String()).
                Str("priceId", cartItem.PriceID.String()).
                Msg("Creating subscription for order item")

            // Creating a subscription would normally involve Stripe API calls
            // Here we'll just create a basic record in our database
            subscription := &models.Subscription{
                UserID:               req.UserID,
                PriceID:              *cartItem.PriceID,
                Quantity:             cartItem.Quantity,
                Status:               models.SubscriptionStatusActive,
                CurrentPeriodStart:   time.Now(),
                CurrentPeriodEnd:     time.Now().AddDate(0, 1, 0), // Default to 1 month
                StripeSubscriptionID: "sub_" + uuid.New().String(), // Placeholder for real Stripe ID
                StripeCustomerID:     *user.StripeCustomerID,
                CollectionMethod:     "charge_automatically",
            }

            err = h.subscriptionRepo.Create(r.Context(), subscription)
            if err != nil {
                logger.Error().
                    Err(err).
                    Str("orderId", order.ID.String()).
                    Str("userId", req.UserID.String()).
                    Str("priceId", cartItem.PriceID.String()).
                    Msg("Failed to create subscription")
                http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
                return
            }

            logger.Debug().
                Str("subscriptionId", subscription.ID.String()).
                Str("userId", req.UserID.String()).
                Str("priceId", cartItem.PriceID.String()).
                Msg("Subscription created, linking to order item")

            // Link the order item to the subscription
            orderItem.SubscriptionID = &subscription.ID
        }

        err = h.orderItemRepo.Create(r.Context(), orderItem)
        if err != nil {
            logger.Error().
                Err(err).
                Str("orderId", order.ID.String()).
                Str("productId", cartItem.ProductID.String()).
                Msg("Failed to create order item")
            http.Error(w, "Failed to create order item", http.StatusInternalServerError)
            return
        }

        logger.Debug().
            Str("orderItemId", orderItem.ID.String()).
            Str("orderId", order.ID.String()).
            Str("productId", cartItem.ProductID.String()).
            Msg("Order item created successfully")
    }

    // Clear the cart after successful order creation
    logger.Debug().Str("cartId", req.CartID.String()).Msg("Clearing cart after successful order creation")
    err = h.cartItemRepo.DeleteByCartID(r.Context(), req.CartID)
    if err != nil {
        // Log the error but don't fail the request
        // The order has been created successfully
        logger.Warn().
            Err(err).
            Str("cartId", req.CartID.String()).
            Str("orderId", order.ID.String()).
            Msg("Failed to clear cart after order creation")
    }

    logger.Debug().
        Str("orderId", order.ID.String()).
        Str("userId", req.UserID.String()).
        Int64("totalAmount", totalAmount).
        Msg("Order created successfully, generating response")

    response, err := h.orderToResponse(r.Context(), order, true, true)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", order.ID.String()).
            Msg("Failed to generate order response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Str("orderId", order.ID.String()).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("orderId", order.ID.String()).
        Str("userId", req.UserID.String()).
        Int64("totalAmount", totalAmount).
        Str("currency", req.Currency).
        Int("itemCount", len(cartItems)).
        Msg("Order created successfully")
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "GetOrder").
        Logger()

    logger.Debug().Msg("Processing order retrieval request")

    vars := mux.Vars(r)
    orderID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("orderIdRaw", vars["id"]).Msg("Invalid order ID")
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Msg("Parsed order ID")

    includeItems := r.URL.Query().Get("include_items") != "false" // Default to true
    includeUser := r.URL.Query().Get("include_user") == "true"    // Default to false

    logger.Debug().
        Str("orderId", orderID.String()).
        Bool("includeItems", includeItems).
        Bool("includeUser", includeUser).
        Msg("Retrieving order with query parameters")

    order, err := h.orderRepo.GetByID(r.Context(), orderID)
    if err != nil {
        if errors.Is(err, repository.ErrOrderNotFound) {
            logger.Error().
                Err(err).
                Str("orderId", orderID.String()).
                Msg("Order not found")
            http.Error(w, "Order not found", http.StatusNotFound)
            return
        }
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to get order")
        http.Error(w, "Failed to get order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Int64("totalAmount", order.TotalAmount).
        Str("currency", order.Currency).
        Msg("Order retrieved successfully, generating response")

    response, err := h.orderToResponse(r.Context(), order, includeItems, includeUser)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Bool("includeItems", includeItems).
        Bool("includeUser", includeUser).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("orderId", orderID.String()).
        Bool("includeItems", includeItems).
        Bool("includeUser", includeUser).
        Msg("Order retrieved successfully")
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "UpdateOrder").
        Logger()

    logger.Debug().Msg("Processing order update request")

    vars := mux.Vars(r)
    orderID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("orderIdRaw", vars["id"]).Msg("Invalid order ID")
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("orderId", orderID.String()).Msg("Parsed order ID")

    var req UpdateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Bool("hasStatus", req.Status != nil).
        Bool("hasShippingAddress", req.ShippingAddress != nil).
        Bool("hasBillingAddress", req.BillingAddress != nil).
        Bool("hasPaymentIntentID", req.PaymentIntentID != nil).
        Bool("hasCompletedAt", req.CompletedAt != nil).
        Bool("hasMetadata", req.Metadata != nil && len(*req.Metadata) > 0).
        Msg("Received order update request")

    // Get the existing order
    logger.Debug().Str("orderId", orderID.String()).Msg("Retrieving order")
    order, err := h.orderRepo.GetByID(r.Context(), orderID)
    if err != nil {
        if errors.Is(err, repository.ErrOrderNotFound) {
            logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Order not found")
            http.Error(w, "Order not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Failed to get order")
        http.Error(w, "Failed to get order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("currentStatus", string(order.Status)).
        Bool("hasCurrentCompletedAt", order.CompletedAt != nil).
        Msg("Order retrieved, applying updates")

    // Update fields if provided
    if req.Status != nil {
        prevStatus := order.Status
        order.Status = models.OrderStatus(*req.Status)
        
        logger.Debug().
            Str("orderId", orderID.String()).
            Str("prevStatus", string(prevStatus)).
            Str("newStatus", string(order.Status)).
            Msg("Updating order status")

        // If status is completed and there's no completed_at timestamp, set it
        if order.Status == models.OrderStatusCompleted && order.CompletedAt == nil {
            now := time.Now()
            order.CompletedAt = &now
            logger.Debug().
                Str("orderId", orderID.String()).
                Time("completedAt", now).
                Msg("Setting completed_at timestamp")
        }
    }

    if req.ShippingAddress != nil {
        logger.Debug().
            Str("orderId", orderID.String()).
            Str("name", req.ShippingAddress.Name).
            Str("country", req.ShippingAddress.Country).
            Msg("Updating shipping address")

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
        logger.Debug().
            Str("orderId", orderID.String()).
            Str("name", req.BillingAddress.Name).
            Str("country", req.BillingAddress.Country).
            Msg("Updating billing address")

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
        logger.Debug().
            Str("orderId", orderID.String()).
            Str("paymentIntentId", *req.PaymentIntentID).
            Msg("Updating payment intent ID")

        order.PaymentIntentID = req.PaymentIntentID
    }

    if req.CompletedAt != nil {
        logger.Debug().
            Str("orderId", orderID.String()).
            Str("completedAtRaw", *req.CompletedAt).
            Msg("Parsing completed_at timestamp")

        completedAt, err := h.parseTimeStr(*req.CompletedAt)
        if err != nil {
            logger.Error().
                Err(err).
                Str("orderId", orderID.String()).
                Str("completedAtRaw", *req.CompletedAt).
                Msg("Invalid date format for completed_at")
            http.Error(w, "Invalid date format for completed_at", http.StatusBadRequest)
            return
        }
        
        logger.Debug().
            Str("orderId", orderID.String()).
            Time("completedAt", completedAt).
            Msg("Setting completed_at timestamp")

        order.CompletedAt = &completedAt
    }

    if req.Metadata != nil {
        logger.Debug().
            Str("orderId", orderID.String()).
            Interface("metadata", *req.Metadata).
            Msg("Updating order metadata")

        order.Metadata = req.Metadata
    }

    // Update the order
    logger.Debug().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Msg("Updating order in database")

    err = h.orderRepo.Update(r.Context(), order)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to update order")
        http.Error(w, "Failed to update order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Msg("Order updated successfully, generating response")

    response, err := h.orderToResponse(r.Context(), order, true, false)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Bool("hasPaymentIntentID", order.PaymentIntentID != nil).
        Bool("hasCompletedAt", order.CompletedAt != nil).
        Msg("Order updated successfully")
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "CancelOrder").
        Logger()

    logger.Debug().Msg("Processing order cancellation request")

    vars := mux.Vars(r)
    orderID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("orderIdRaw", vars["id"]).Msg("Invalid order ID")
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("orderId", orderID.String()).Msg("Parsed order ID")

    // Get the existing order
    logger.Debug().Str("orderId", orderID.String()).Msg("Retrieving order")
    order, err := h.orderRepo.GetByID(r.Context(), orderID)
    if err != nil {
        if errors.Is(err, repository.ErrOrderNotFound) {
            logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Order not found")
            http.Error(w, "Order not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Failed to get order")
        http.Error(w, "Failed to get order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("currentStatus", string(order.Status)).
        Msg("Order retrieved, checking if cancellation is allowed")

    // Can only cancel pending or processing orders
    if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusProcessing {
        logger.Warn().
            Str("orderId", orderID.String()).
            Str("currentStatus", string(order.Status)).
            Msg("Cannot cancel order with current status")
        http.Error(w, "Only pending or processing orders can be canceled", http.StatusBadRequest)
        return
    }

    // Update the order status
    prevStatus := order.Status
    order.Status = models.OrderStatusCanceled
    order.UpdatedAt = time.Now()

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("prevStatus", string(prevStatus)).
        Str("newStatus", string(order.Status)).
        Time("updatedAt", order.UpdatedAt).
        Msg("Updating order status to canceled")

    // Update the order
    err = h.orderRepo.Update(r.Context(), order)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to update order status to canceled")
        http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Msg("Order status updated to canceled, checking for associated subscriptions")

    // Cancel any subscriptions associated with this order
    orderItems, err := h.orderItemRepo.ListByOrderID(r.Context(), orderID)
    if err != nil {
        logger.Warn().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to retrieve order items, will skip subscription cancellation")
    } else {
        logger.Debug().
            Str("orderId", orderID.String()).
            Int("itemCount", len(orderItems)).
            Msg("Retrieved order items, checking for subscriptions")

        for i, item := range orderItems {
            if item.SubscriptionID != nil {
                logger.Debug().
                    Str("orderId", orderID.String()).
                    Str("orderItemId", item.ID.String()).
                    Str("subscriptionId", item.SubscriptionID.String()).
                    Int("itemIndex", i).
                    Msg("Found subscription to cancel")

                // In a real application, you would call Stripe API here
                // For now, we'll just update our database
                subscription, err := h.subscriptionRepo.GetByID(r.Context(), *item.SubscriptionID)
                if err != nil {
                    logger.Warn().
                        Err(err).
                        Str("orderId", orderID.String()).
                        Str("subscriptionId", item.SubscriptionID.String()).
                        Msg("Failed to retrieve subscription, skipping cancellation")
                } else {
                    now := time.Now()
                    prevStatus := subscription.Status
                    subscription.Status = models.SubscriptionStatusCanceled
                    subscription.CanceledAt = &now
                    subscription.EndedAt = &now

                    logger.Debug().
                        Str("orderId", orderID.String()).
                        Str("subscriptionId", subscription.ID.String()).
                        Str("prevStatus", string(prevStatus)).
                        Str("newStatus", string(subscription.Status)).
                        Time("canceledAt", now).
                        Msg("Canceling subscription")

                    updateErr := h.subscriptionRepo.Update(r.Context(), subscription)
                    if updateErr != nil {
                        logger.Warn().
                            Err(updateErr).
                            Str("orderId", orderID.String()).
                            Str("subscriptionId", subscription.ID.String()).
                            Msg("Failed to update subscription status")
                    } else {
                        logger.Debug().
                            Str("orderId", orderID.String()).
                            Str("subscriptionId", subscription.ID.String()).
                            Msg("Subscription canceled successfully")
                    }
                }
            }
        }
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Msg("Order canceled, generating response")

    response, err := h.orderToResponse(r.Context(), order, true, false)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to generate response")
        http.Error(w, "Failed to generate response", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("orderId", orderID.String()).
        Msg("Order canceled successfully")
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "DeleteOrder").
        Logger()

    logger.Debug().Msg("Processing order deletion request")

    vars := mux.Vars(r)
    orderID, err := uuid.Parse(vars["id"])
    if err != nil {
        logger.Error().Err(err).Str("orderIdRaw", vars["id"]).Msg("Invalid order ID")
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("orderId", orderID.String()).Msg("Parsed order ID")

    // Get the order to check its status
    logger.Debug().Str("orderId", orderID.String()).Msg("Retrieving order")
    order, err := h.orderRepo.GetByID(r.Context(), orderID)
    if err != nil {
        if errors.Is(err, repository.ErrOrderNotFound) {
            logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Order not found")
            http.Error(w, "Order not found", http.StatusNotFound)
            return
        }
        logger.Error().Err(err).Str("orderId", orderID.String()).Msg("Failed to get order")
        http.Error(w, "Failed to get order", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Msg("Order retrieved, checking if deletion is allowed")

    // Only allow deletion of pending or canceled orders
    if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusCanceled {
        logger.Warn().
            Str("orderId", orderID.String()).
            Str("status", string(order.Status)).
            Msg("Cannot delete order with current status")
        http.Error(w, "Only pending or canceled orders can be deleted", http.StatusBadRequest)
        return
    }

    logger.Debug().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Msg("Deleting order")

    // Delete the order (this will cascade to delete order items due to foreign key constraint)
    err = h.orderRepo.Delete(r.Context(), orderID)
    if err != nil {
        logger.Error().
            Err(err).
            Str("orderId", orderID.String()).
            Msg("Failed to delete order")
        http.Error(w, "Failed to delete order", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("orderId", orderID.String()).
        Str("status", string(order.Status)).
        Msg("Order deleted successfully")

    w.WriteHeader(http.StatusNoContent)
}

func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "ListUserOrders").
        Logger()

    logger.Debug().Msg("Processing list user orders request")

    vars := mux.Vars(r)
    userID, err := uuid.Parse(vars["user_id"])
    if err != nil {
        logger.Error().Err(err).Str("userIdRaw", vars["user_id"]).Msg("Invalid user ID")
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    logger.Debug().Str("userId", userID.String()).Msg("Parsed user ID")

    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

    logger.Debug().
        Str("userId", userID.String()).
        Str("limitRaw", limitStr).
        Str("offsetRaw", offsetStr).
        Bool("includeItems", includeItems).
        Msg("Parsed query parameters")

    limit := 10 // Default
    offset := 0 // Default

    if limitStr != "" {
        parsedLimit, err := strconv.Atoi(limitStr)
        if err == nil && parsedLimit > 0 {
            limit = parsedLimit
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("userId", userID.String()).
                Str("limitRaw", limitStr).
                Msg("Invalid limit parameter, using default")
        }
    }

    if offsetStr != "" {
        parsedOffset, err := strconv.Atoi(offsetStr)
        if err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("userId", userID.String()).
                Str("offsetRaw", offsetStr).
                Msg("Invalid offset parameter, using default")
        }
    }

    logger.Debug().
        Str("userId", userID.String()).
        Int("limit", limit).
        Int("offset", offset).
        Bool("includeItems", includeItems).
        Msg("Using pagination parameters")

    // Verify user exists
    logger.Debug().Str("userId", userID.String()).Msg("Verifying user exists")
    _, err = h.userRepo.GetByID(r.Context(), userID)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            logger.Error().
                Err(err).
                Str("userId", userID.String()).
                Msg("User not found")
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        logger.Error().
            Err(err).
            Str("userId", userID.String()).
            Msg("Failed to verify user")
        http.Error(w, "Failed to verify user", http.StatusInternalServerError)
        return
    }

    // Get the user's orders
    logger.Debug().
        Str("userId", userID.String()).
        Int("limit", limit).
        Int("offset", offset).
        Msg("Retrieving user orders")

    orders, err := h.orderRepo.ListByUserID(r.Context(), userID, limit, offset)
    if err != nil {
        logger.Error().
            Err(err).
            Str("userId", userID.String()).
            Int("limit", limit).
            Int("offset", offset).
            Msg("Failed to list user orders")
        http.Error(w, "Failed to list user orders", http.StatusInternalServerError)
        return
    }

    logger.Debug().
        Str("userId", userID.String()).
        Int("orderCount", len(orders)).
        Msg("Retrieved user orders, generating response")

    // Convert to response objects
    var responses []OrderResponse
    for i, order := range orders {
        logger.Debug().
            Str("userId", userID.String()).
            Str("orderId", order.ID.String()).
            Int("orderIndex", i).
            Bool("includeItems", includeItems).
            Msg("Converting order to response")

        response, err := h.orderToResponse(r.Context(), order, includeItems, false)
        if err != nil {
            logger.Error().
                Err(err).
                Str("userId", userID.String()).
                Str("orderId", order.ID.String()).
                Msg("Failed to generate response for order")
            http.Error(w, "Failed to generate response", http.StatusInternalServerError)
            return
        }
        responses = append(responses, response)
    }

    logger.Debug().
        Str("userId", userID.String()).
        Int("responseCount", len(responses)).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(responses); err != nil {
        logger.Error().
            Err(err).
            Str("userId", userID.String()).
            Int("responseCount", len(responses)).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Str("userId", userID.String()).
        Int("orderCount", len(orders)).
        Int("limit", limit).
        Int("offset", offset).
        Bool("includeItems", includeItems).
        Msg("User orders listed successfully")
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "OrderHandler").
        Str("method", "ListOrders").
        Logger()

    logger.Debug().Msg("Processing list orders request")

    // Parse query parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    status := r.URL.Query().Get("status")
    includeItems := r.URL.Query().Get("include_items") != "false" // Default to true

    logger.Debug().
        Str("limitRaw", limitStr).
        Str("offsetRaw", offsetStr).
        Str("status", status).
        Bool("includeItems", includeItems).
        Msg("Parsed query parameters")

    limit := 10 // Default
    offset := 0 // Default

    if limitStr != "" {
        parsedLimit, err := strconv.Atoi(limitStr)
        if err == nil && parsedLimit > 0 {
            limit = parsedLimit
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("limitRaw", limitStr).
                Msg("Invalid limit parameter, using default")
        }
    }

    if offsetStr != "" {
        parsedOffset, err := strconv.Atoi(offsetStr)
        if err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        } else if err != nil {
            logger.Warn().
                Err(err).
                Str("offsetRaw", offsetStr).
                Msg("Invalid offset parameter, using default")
        }
    }

    logger.Debug().
        Int("limit", limit).
        Int("offset", offset).
        Str("status", status).
        Bool("includeItems", includeItems).
        Msg("Using pagination parameters")

    var orders []*models.Order
    var err error

    // Filter by status if provided
    if status != "" {
        logger.Debug().
            Str("status", status).
            Int("limit", limit).
            Int("offset", offset).
            Msg("Retrieving orders filtered by status")

        orders, err = h.orderRepo.ListByStatus(r.Context(), models.OrderStatus(status), limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Str("status", status).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list orders by status")
            http.Error(w, "Failed to list orders by status", http.StatusInternalServerError)
            return
        }
    } else {
        // List all orders
        logger.Debug().
            Int("limit", limit).
            Int("offset", offset).
            Msg("Retrieving all orders")

        orders, err = h.orderRepo.List(r.Context(), limit, offset)
        if err != nil {
            logger.Error().
                Err(err).
                Int("limit", limit).
                Int("offset", offset).
                Msg("Failed to list orders")
            http.Error(w, "Failed to list orders", http.StatusInternalServerError)
            return
        }
    }

    logger.Debug().
        Int("orderCount", len(orders)).
        Str("statusFilter", status).
        Msg("Retrieved orders, generating response")

    // Convert to response objects
    var responses []OrderResponse
    for i, order := range orders {
        logger.Debug().
            Str("orderId", order.ID.String()).
            Str("status", string(order.Status)).
            Int("orderIndex", i).
            Bool("includeItems", includeItems).
            Msg("Converting order to response")

        response, err := h.orderToResponse(r.Context(), order, includeItems, false)
        if err != nil {
            logger.Error().
                Err(err).
                Str("orderId", order.ID.String()).
                Msg("Failed to generate response for order")
            http.Error(w, "Failed to generate response", http.StatusInternalServerError)
            return
        }
        responses = append(responses, response)
    }

    logger.Debug().
        Int("responseCount", len(responses)).
        Msg("Response generated, sending to client")

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(responses); err != nil {
        logger.Error().
            Int("responseCount", len(responses)).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }

    logger.Info().
        Int("orderCount", len(orders)).
        Int("limit", limit).
        Int("offset", offset).
        Str("statusFilter", status).
        Bool("includeItems", includeItems).
        Msg("Orders listed successfully")
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