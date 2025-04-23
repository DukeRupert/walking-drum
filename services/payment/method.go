// services/payment/payment_method.go
package payment

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

// CreatePaymentMethodInput defines the input for creating a payment method
type CreatePaymentMethodInput struct {
    CustomerID string
    CardToken  string
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