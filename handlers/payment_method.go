// handlers/payment_method_handler.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dukerupert/walking-drum/services/payment"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type PaymentMethodHandler struct {
    paymentMethodService *payment.PaymentMethodService
}

func NewPaymentMethodHandler(paymentMethodService *payment.PaymentMethodService) *PaymentMethodHandler {
    return &PaymentMethodHandler{
        paymentMethodService: paymentMethodService,
    }
}

// Request/Response structs

type CreatePaymentMethodRequest struct {
    CustomerID string            `json:"customer_id"`
    CardToken  string            `json:"card_token"`
    Metadata   map[string]string `json:"metadata,omitempty"`
}

type UpdatePaymentMethodRequest struct {
    BillingName string            `json:"billing_name,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type PaymentMethodResponse struct {
    ID string `json:"id"`
}

type SetDefaultPaymentMethodRequest struct {
    CustomerID      string `json:"customer_id"`
    PaymentMethodID string `json:"payment_method_id"`
}

// Handlers

func (h *PaymentMethodHandler) CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "CreatePaymentMethod").
        Logger()

    logger.Debug().Msg("Processing payment method creation request")

    var req CreatePaymentMethodRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Mask sensitive data in logs
    maskedCardToken := "REDACTED"
    if len(req.CardToken) > 4 {
        maskedCardToken = "..." + req.CardToken[len(req.CardToken)-4:]
    }
    
    logger.Debug().
        Str("customerId", req.CustomerID).
        Str("cardToken", maskedCardToken).
        Bool("hasMetadata", len(req.Metadata) > 0).
        Msg("Received payment method creation request")
    
    // Basic validation
    if req.CustomerID == "" {
        logger.Error().Msg("Customer ID is required")
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    if req.CardToken == "" {
        logger.Error().Msg("Card token is required")
        http.Error(w, "Card token is required", http.StatusBadRequest)
        return
    }
    
    // Create the payment method
    logger.Debug().
        Str("customerId", req.CustomerID).
        Str("cardToken", maskedCardToken).
        Msg("Creating payment method with payment service")
        
    input := payment.CreatePaymentMethodInput{
        CustomerID: req.CustomerID,
        CardToken:  req.CardToken,
        Metadata:   req.Metadata,
    }
    
    paymentMethodID, err := h.paymentMethodService.CreatePaymentMethod(input)
    if err != nil {
        logger.Error().
            Err(err).
            Str("customerId", req.CustomerID).
            Msg("Failed to create payment method with payment service")
        http.Error(w, "Failed to create payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Str("customerId", req.CustomerID).
        Msg("Payment method created successfully, generating response")
    
    // Return the payment method ID
    response := PaymentMethodResponse{
        ID: paymentMethodID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        logger.Error().Err(err).Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("paymentMethodId", paymentMethodID).
        Str("customerId", req.CustomerID).
        Msg("Payment method created successfully")
}

func (h *PaymentMethodHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "ListPaymentMethods").
        Logger()

    logger.Debug().Msg("Processing list payment methods request")

    vars := mux.Vars(r)
    customerID := vars["customerID"]
    
    if customerID == "" {
        logger.Error().Msg("Customer ID is required")
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    logger.Debug().Str("customerId", customerID).Msg("Listing payment methods for customer")
    
    paymentMethods, err := h.paymentMethodService.ListPaymentMethods(customerID)
    if err != nil {
        logger.Error().
            Err(err).
            Str("customerId", customerID).
            Msg("Failed to list payment methods from payment service")
        http.Error(w, "Failed to list payment methods: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Debug().
        Str("customerId", customerID).
        Int("count", len(paymentMethods)).
        Msg("Retrieved payment methods, encoding response")
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(paymentMethods); err != nil {
        logger.Error().
            Err(err).
            Str("customerId", customerID).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("customerId", customerID).
        Int("count", len(paymentMethods)).
        Msg("Payment methods listed successfully")
}

func (h *PaymentMethodHandler) GetPaymentMethod(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "GetPaymentMethod").
        Logger()

    logger.Debug().Msg("Processing get payment method request")

    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        logger.Error().Msg("Payment method ID is required")
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    logger.Debug().Str("paymentMethodId", paymentMethodID).Msg("Looking up payment method")
    
    paymentMethod, err := h.paymentMethodService.GetPaymentMethod(paymentMethodID)
    if err != nil {
        logger.Error().
            Err(err).
            Str("paymentMethodId", paymentMethodID).
            Msg("Failed to get payment method from payment service")
        http.Error(w, "Failed to get payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Str("type", paymentMethod.Type).
        Msg("Payment method found, encoding response")
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(paymentMethod); err != nil {
        logger.Error().
            Err(err).
            Str("paymentMethodId", paymentMethodID).
            Msg("Failed to encode JSON response")
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("paymentMethodId", paymentMethodID).
        Msg("Payment method retrieved successfully")
}

func (h *PaymentMethodHandler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "UpdatePaymentMethod").
        Logger()

    logger.Debug().Msg("Processing update payment method request")

    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        logger.Error().Msg("Payment method ID is required")
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    var req UpdatePaymentMethodRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Error().Err(err).Str("paymentMethodId", paymentMethodID).Msg("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Str("billingName", req.BillingName).
        Bool("hasMetadata", len(req.Metadata) > 0).
        Msg("Received payment method update request")
    
    input := payment.UpdatePaymentMethodInput{
        PaymentMethodID: paymentMethodID,
        BillingName:     req.BillingName,
        Metadata:        req.Metadata,
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Msg("Updating payment method with payment service")
    
    if err := h.paymentMethodService.UpdatePaymentMethod(input); err != nil {
        logger.Error().
            Err(err).
            Str("paymentMethodId", paymentMethodID).
            Msg("Failed to update payment method with payment service")
        http.Error(w, "Failed to update payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("paymentMethodId", paymentMethodID).
        Msg("Payment method updated successfully")
    
    w.WriteHeader(http.StatusOK)
}

func (h *PaymentMethodHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "SetDefaultPaymentMethod").
        Logger()

    logger.Debug().Msg("Processing set default payment method request")

    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    customerID := vars["customerID"]
    
    if paymentMethodID == "" {
        logger.Error().Msg("Payment method ID is required")
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    if customerID == "" {
        logger.Error().Msg("Customer ID is required")
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Str("customerId", customerID).
        Msg("Setting default payment method for customer")
    
    if err := h.paymentMethodService.SetDefaultPaymentMethod(customerID, paymentMethodID); err != nil {
        logger.Error().
            Err(err).
            Str("paymentMethodId", paymentMethodID).
            Str("customerId", customerID).
            Msg("Failed to set default payment method with payment service")
        http.Error(w, "Failed to set default payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("paymentMethodId", paymentMethodID).
        Str("customerId", customerID).
        Msg("Default payment method set successfully")
    
    w.WriteHeader(http.StatusOK)
}

func (h *PaymentMethodHandler) DetachPaymentMethod(w http.ResponseWriter, r *http.Request) {
    logger := log.With().
        Str("handler", "PaymentMethodHandler").
        Str("method", "DetachPaymentMethod").
        Logger()

    logger.Debug().Msg("Processing detach payment method request")

    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        logger.Error().Msg("Payment method ID is required")
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    logger.Debug().
        Str("paymentMethodId", paymentMethodID).
        Msg("Detaching payment method")
    
    if err := h.paymentMethodService.DetachPaymentMethod(paymentMethodID); err != nil {
        logger.Error().
            Err(err).
            Str("paymentMethodId", paymentMethodID).
            Msg("Failed to detach payment method with payment service")
        http.Error(w, "Failed to detach payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    logger.Info().
        Str("paymentMethodId", paymentMethodID).
        Msg("Payment method detached successfully")
    
    w.WriteHeader(http.StatusOK)
}