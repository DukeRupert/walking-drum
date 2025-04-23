// services/payment/payment_method.go
package payment

import (
	"errors"
)

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	CardBrand    string            `json:"card_brand,omitempty"`
	CardLast4    string            `json:"card_last4,omitempty"`
	CardExpMonth int               `json:"card_exp_month,omitempty"`
	CardExpYear  int               `json:"card_exp_year,omitempty"`
	BillingName  string            `json:"billing_name,omitempty"`
	IsDefault    bool              `json:"is_default"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PaymentMethodService handles payment method operations
type PaymentMethodService struct {
	processor Processor
}

// NewPaymentMethodService creates a new payment method service
func NewPaymentMethodService(processor Processor) *PaymentMethodService {
	return &PaymentMethodService{
		processor: processor,
	}
}

// UpdatePaymentMethodInput defines the input for updating a payment method
type UpdatePaymentMethodInput struct {
	PaymentMethodID string
	BillingName     string
	Metadata        map[string]string
}

// CreatePaymentMethodInput defines the input for creating a payment method
type CreatePaymentMethodInput struct {
	CustomerID string
	CardToken  string
	Metadata   map[string]string
}

// CreatePaymentMethod creates a new payment method and sets it as default
func (s *PaymentMethodService) CreatePaymentMethod(input CreatePaymentMethodInput) (string, error) {
	// Create payment method request
	req := PaymentMethodRequest{
		CustomerID: input.CustomerID,
		Type:       "card",
		Token:      input.CardToken,
	}

	// Create the payment method and attach it to the customer
	paymentMethodID, err := s.processor.CreatePaymentMethod(req)
	if err != nil {
		return "", err
	}

	// Set as default payment method for the customer
	err = s.processor.SetDefaultPaymentMethod(input.CustomerID, paymentMethodID)
	if err != nil {
		return "", err
	}

	return paymentMethodID, nil
}

// ListPaymentMethods retrieves all payment methods for a customer
func (s *PaymentMethodService) ListPaymentMethods(customerID string) ([]PaymentMethod, error) {
	if customerID == "" {
		return nil, errors.New("customer ID cannot be empty")
	}

	return s.processor.ListPaymentMethods(customerID)
}

// GetPaymentMethod retrieves a specific payment method
func (s *PaymentMethodService) GetPaymentMethod(paymentMethodID string) (*PaymentMethod, error) {
	if paymentMethodID == "" {
		return nil, errors.New("payment method ID cannot be empty")
	}

	return s.processor.GetPaymentMethod(paymentMethodID)
}

// UpdatePaymentMethod updates a payment method's details
func (s *PaymentMethodService) UpdatePaymentMethod(input UpdatePaymentMethodInput) error {
	if input.PaymentMethodID == "" {
		return errors.New("payment method ID cannot be empty")
	}

	return s.processor.UpdatePaymentMethod(input.PaymentMethodID, input.BillingName, input.Metadata)
}

// SetDefaultPaymentMethod sets a payment method as the default for a customer
func (s *PaymentMethodService) SetDefaultPaymentMethod(customerID, paymentMethodID string) error {
	if customerID == "" {
		return errors.New("customer ID cannot be empty")
	}

	if paymentMethodID == "" {
		return errors.New("payment method ID cannot be empty")
	}

	return s.processor.SetDefaultPaymentMethod(customerID, paymentMethodID)
}

// DetachPaymentMethod detaches a payment method from a customer
func (s *PaymentMethodService) DetachPaymentMethod(paymentMethodID string) error {
	if paymentMethodID == "" {
		return errors.New("payment method ID cannot be empty")
	}

	return s.processor.DetachPaymentMethod(paymentMethodID)
}
