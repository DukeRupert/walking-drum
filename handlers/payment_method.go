// handlers/payment_method_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    
    "github.com/dukerupert/walking-drum/services/payment"
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
    CustomerID string `json:"customer_id"`
    CardToken  string `json:"card_token"`
}

type PaymentMethodResponse struct {
    ID string `json:"id"`
}

// Handlers

func (h *PaymentMethodHandler) CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
    var req CreatePaymentMethodRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Basic validation
    if req.CustomerID == "" {
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    if req.CardToken == "" {
        http.Error(w, "Card token is required", http.StatusBadRequest)
        return
    }
    
    // Create the payment method
    input := payment.CreatePaymentMethodInput{
        CustomerID: req.CustomerID,
        CardToken:  req.CardToken,
    }
    
    paymentMethodID, err := h.paymentMethodService.CreatePaymentMethod(input)
    if err != nil {
        http.Error(w, "Failed to create payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return the payment method ID
    response := PaymentMethodResponse{
        ID: paymentMethodID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}