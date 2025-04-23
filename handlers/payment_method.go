// handlers/payment_method_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/dukerupert/walking-drum/services/payment"
    "github.com/gorilla/mux"
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
        Metadata:   req.Metadata,
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

// ListPaymentMethods returns all payment methods for a customer
func (h *PaymentMethodHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    customerID := vars["customerID"]
    
    if customerID == "" {
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    paymentMethods, err := h.paymentMethodService.ListPaymentMethods(customerID)
    if err != nil {
        http.Error(w, "Failed to list payment methods: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(paymentMethods)
}

// GetPaymentMethod returns a specific payment method
func (h *PaymentMethodHandler) GetPaymentMethod(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    paymentMethod, err := h.paymentMethodService.GetPaymentMethod(paymentMethodID)
    if err != nil {
        http.Error(w, "Failed to get payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(paymentMethod)
}

// UpdatePaymentMethod updates a payment method
func (h *PaymentMethodHandler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    var req UpdatePaymentMethodRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    input := payment.UpdatePaymentMethodInput{
        PaymentMethodID: paymentMethodID,
        BillingName:     req.BillingName,
        Metadata:        req.Metadata,
    }
    
    if err := h.paymentMethodService.UpdatePaymentMethod(input); err != nil {
        http.Error(w, "Failed to update payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

// SetDefaultPaymentMethod sets a payment method as default for a customer
func (h *PaymentMethodHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    customerID := vars["customerID"]
    
    if paymentMethodID == "" {
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    if customerID == "" {
        http.Error(w, "Customer ID is required", http.StatusBadRequest)
        return
    }
    
    if err := h.paymentMethodService.SetDefaultPaymentMethod(customerID, paymentMethodID); err != nil {
        http.Error(w, "Failed to set default payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

// DetachPaymentMethod removes a payment method from a customer
func (h *PaymentMethodHandler) DetachPaymentMethod(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    paymentMethodID := vars["id"]
    
    if paymentMethodID == "" {
        http.Error(w, "Payment method ID is required", http.StatusBadRequest)
        return
    }
    
    if err := h.paymentMethodService.DetachPaymentMethod(paymentMethodID); err != nil {
        http.Error(w, "Failed to detach payment method: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}