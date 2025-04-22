// services/payment/processor.go
package payment

// SubscriptionRequest represents a request to create a subscription
type SubscriptionRequest struct {
	CustomerID      string
	PriceID         string
	Quantity        int64
	PaymentMethodID string
	Description     string
	OrderID         string
	Metadata        map[string]string
}

// SubscriptionResponse represents a generic subscription response
type SubscriptionResponse struct {
	ID                 string
	CustomerID         string
	Status             string
	CurrentPeriodStart int64
	CurrentPeriodEnd   int64
	CancelAtPeriodEnd  bool
	LatestInvoiceID    string
	ProcessorID        string // The ID in the payment processor's system
	Metadata           map[string]string
}

// Processor interface defines methods that any payment processor must implement
type Processor interface {
	// Subscription operations
	CreateSubscription(request SubscriptionRequest) (*SubscriptionResponse, error)
	CancelSubscription(subscriptionID string) (*SubscriptionResponse, error)
	UpdateSubscription(subscriptionID string, request SubscriptionRequest) (*SubscriptionResponse, error)
	
	// Webhook handling
	HandleWebhook(body []byte, signature string) (interface{}, error)
}