// services/payment/processor.go
package payment

type SubscriptionRequest struct {
	CustomerID      string
	PriceID         string
	Quantity        int64
	PaymentMethodID string
	Description     string
	OrderID         string
	Metadata        map[string]string
}

type PriceRequest struct {
    ProductID      string
    UnitAmount     int64
    Currency       string
    Recurring      bool
    IntervalType   string // "day", "week", "month", "year"
    IntervalCount  int64
    Nickname       string
    Metadata       map[string]string
}

type ProductRequest struct {
    Name        string
    Description string
    Active      bool
    Metadata    map[string]string
}

type CustomerRequest struct {
    Email       string
    Name        string
    Description string
    Metadata    map[string]string
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

	// Customer operations
	CreateCustomer(request CustomerRequest) (string, error)
	RetrieveCustomer(customerID string, params interface{}) (interface{}, error)

	// Price operations
	CreatePrice(request PriceRequest) (string, error)
    RetrievePrice(priceID string, params interface{}) (interface{}, error)

	// Product operations
	CreateProduct(request ProductRequest) (string, error)
    RetrieveProduct(productID string, params interface{}) (interface{}, error)
	
	// Webhook handling
	HandleWebhook(body []byte, signature string) (interface{}, error)
}