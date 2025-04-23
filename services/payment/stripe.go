// services/payment/stripe_processor.go
package payment

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/client"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
	"github.com/stripe/stripe-go/v74/webhook"
)

// StripeProcessor implements the Processor interface for Stripe
type StripeProcessor struct {
	client *client.API
}

// NewStripeProcessor creates a new Stripe processor
func NewStripeProcessor() *StripeProcessor {
	// Log that we're creating a new Stripe processor
	log.Debug().Msg("Creating new Stripe processor")

	// Get the API key and log whether it was found
	apiKey := os.Getenv("STRIPE_SECRET_KEY")
	if apiKey == "" {
		log.Warn().Msg("STRIPE_SECRET_KEY environment variable is not set or empty")
		// For development, you could fall back to a hardcoded test key
		apiKey = "sk_test_your_test_key_here"
		log.Debug().Msg("Using fallback test key")
	} else {
		// Don't log the full key for security, just the first few characters
		maskedKey := apiKey[:8] + "..." // Only show first 8 chars
		log.Debug().Str("key_prefix", maskedKey).Msg("Found Stripe API key in environment")
	}

	// Initialize the Stripe package-level API key first
	// This is important for many Stripe operations
	stripe.Key = apiKey
	log.Debug().Msg("Set Stripe package-level API key")

	// Then initialize the client
	log.Debug().Msg("Initializing Stripe client")
	sc := &client.API{}
	sc.Init(apiKey, nil)

	// Log successful initialization
	log.Debug().Msg("Stripe client initialized successfully")

	return &StripeProcessor{
		client: sc,
	}
}

// CreateSubscription creates a new subscription in Stripe
func (p *StripeProcessor) CreateSubscription(request SubscriptionRequest) (*SubscriptionResponse, error) {
	// Create subscription parameters
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(request.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(request.PriceID),
				Quantity: stripe.Int64(request.Quantity),
			},
		},
	}

	// Set payment method if provided
	if request.PaymentMethodID != "" {
		params.DefaultPaymentMethod = stripe.String(request.PaymentMethodID)

		// Add payment settings
		params.PaymentSettings = &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		}
	}

	// Set description if provided
	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	// Set metadata
	if request.OrderID != "" {
		params.AddMetadata("order_id", request.OrderID)
	}

	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	// Generate a UUID as idempotency key
	params.SetIdempotencyKey(uuid.New().String())

	// Create the subscription
	sub, err := subscription.New(params)
	if err != nil {
		return nil, err
	}

	// Convert to generic response
	response := &SubscriptionResponse{
		ID:                 sub.ID,
		CustomerID:         sub.Customer.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		ProcessorID:        sub.ID,
	}

	if sub.LatestInvoice != nil {
		response.LatestInvoiceID = sub.LatestInvoice.ID
	}

	// Convert metadata
	if len(sub.Metadata) > 0 {
		response.Metadata = make(map[string]string)
		for k, v := range sub.Metadata {
			response.Metadata[k] = v
		}
	}

	return response, nil
}

// CancelSubscription cancels a subscription in Stripe
func (p *StripeProcessor) CancelSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	params := &stripe.SubscriptionCancelParams{
		Prorate: stripe.Bool(true),
	}

	sub, err := subscription.Cancel(subscriptionID, params)
	if err != nil {
		return nil, err
	}

	// Convert to generic response
	response := &SubscriptionResponse{
		ID:                 sub.ID,
		CustomerID:         sub.Customer.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		ProcessorID:        sub.ID,
	}

	// Convert metadata
	if len(sub.Metadata) > 0 {
		response.Metadata = make(map[string]string)
		for k, v := range sub.Metadata {
			response.Metadata[k] = v
		}
	}

	return response, nil
}

// UpdateSubscription updates a subscription in Stripe
func (p *StripeProcessor) UpdateSubscription(subscriptionID string, request SubscriptionRequest) (*SubscriptionResponse, error) {
	params := &stripe.SubscriptionParams{}

	// If quantity is changing, we need to update the items
	if request.Quantity > 0 {
		// First get the subscription to find the subscription item ID
		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			return nil, err
		}

		// Find the subscription item ID for the given price or use the first item
		var subscriptionItemID string
		for _, item := range sub.Items.Data {
			if request.PriceID == "" || item.Price.ID == request.PriceID {
				subscriptionItemID = item.ID
				break
			}
		}

		if subscriptionItemID != "" {
			params.Items = []*stripe.SubscriptionItemsParams{
				{
					ID:       stripe.String(subscriptionItemID),
					Quantity: stripe.Int64(request.Quantity),
				},
			}
		}
	}

	// Set other fields if provided
	if request.PaymentMethodID != "" {
		params.DefaultPaymentMethod = stripe.String(request.PaymentMethodID)
	}

	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	// Update metadata
	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	// Update the subscription
	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, err
	}

	// Convert to generic response
	response := &SubscriptionResponse{
		ID:                 sub.ID,
		CustomerID:         sub.Customer.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		ProcessorID:        sub.ID,
	}

	if sub.LatestInvoice != nil {
		response.LatestInvoiceID = sub.LatestInvoice.ID
	}

	// Convert metadata
	if len(sub.Metadata) > 0 {
		response.Metadata = make(map[string]string)
		for k, v := range sub.Metadata {
			response.Metadata[k] = v
		}
	}

	return response, nil
}

// CreateCustomer creates a new customer in Stripe
func (p *StripeProcessor) CreateCustomer(request CustomerRequest) (string, error) {
	params := &stripe.CustomerParams{}

	if request.Email != "" {
		params.Email = stripe.String(request.Email)
	}

	if request.Name != "" {
		params.Name = stripe.String(request.Name)
	}

	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	// Add metadata
	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	customer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return customer.ID, nil
}

func (p *StripeProcessor) RetrieveCustomer(customerID string, params interface{}) (interface{}, error) {
	// Assert params to the correct type
	customerParams, ok := params.(*stripe.CustomerParams)
	if !ok {
		customerParams = &stripe.CustomerParams{}
	}

	return customer.Get(customerID, customerParams)
}

func (p *StripeProcessor) CreatePrice(request PriceRequest) (string, error) {
	// First, check if we need to create the product
	productID := request.ProductID

	// If product doesn't exist and isn't a Stripe ID, create it
	if !strings.HasPrefix(productID, "prod_") {
		// Create a product first
		productParams := &stripe.ProductParams{
			Name: stripe.String(productID), // Use the provided ID as name
			Type: stripe.String("service"),
		}

		for k, v := range request.Metadata {
			productParams.AddMetadata(k, v)
		}

		product, err := product.New(productParams)
		if err != nil {
			return "", fmt.Errorf("failed to create product: %w", err)
		}

		productID = product.ID
	}

	// Now create the price
	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(request.UnitAmount),
		Currency:   stripe.String(request.Currency),
	}

	if request.Nickname != "" {
		params.Nickname = stripe.String(request.Nickname)
	}

	// Add metadata
	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	// Set up recurring if needed
	if request.Recurring {
		params.Recurring = &stripe.PriceRecurringParams{
			Interval:      stripe.String(request.IntervalType),
			IntervalCount: stripe.Int64(request.IntervalCount),
		}
	}

	price, err := price.New(params)
	if err != nil {
		return "", err
	}

	return price.ID, nil
}

func (p *StripeProcessor) RetrievePrice(priceID string, params interface{}) (interface{}, error) {
	// Assert params to the correct type
	priceParams, ok := params.(*stripe.PriceParams)
	if !ok {
		priceParams = &stripe.PriceParams{}
	}

	return price.Get(priceID, priceParams)
}

func (p *StripeProcessor) CreateProduct(request ProductRequest) (string, error) {
	params := &stripe.ProductParams{
		Name:   stripe.String(request.Name),
		Active: stripe.Bool(request.Active),
	}

	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	// Add metadata
	for k, v := range request.Metadata {
		params.AddMetadata(k, v)
	}

	product, err := product.New(params)
	if err != nil {
		return "", err
	}

	return product.ID, nil
}

// Add to Stripe processor
func (p *StripeProcessor) RetrieveProduct(productID string, params interface{}) (interface{}, error) {
	// Assert params to the correct type
	productParams, ok := params.(*stripe.ProductParams)
	if !ok {
		productParams = &stripe.ProductParams{}
	}

	return product.Get(productID, productParams)
}

func (p *StripeProcessor) CreatePaymentMethod(request PaymentMethodRequest) (string, error) {
    var pm *stripe.PaymentMethod
    var err error
    
    if request.Token != "" {
        // Create payment method from token
        params := &stripe.PaymentMethodParams{
            Type: stripe.String("card"),
            Card: &stripe.PaymentMethodCardParams{
                Token: stripe.String(request.Token),
            },
        }
        
        pm, err = paymentmethod.New(params)
        if err != nil {
            return "", err
        }
    } else {
        // This branch shouldn't be used in production as it would require
        // special Stripe account settings to accept raw card data
        return "", errors.New("creating payment methods with raw card data is not supported")
    }
    
    // Attach the payment method to the customer
    attachParams := &stripe.PaymentMethodAttachParams{
        Customer: stripe.String(request.CustomerID),
    }
    
    _, err = paymentmethod.Attach(pm.ID, attachParams)
    if err != nil {
        return "", fmt.Errorf("failed to attach payment method to customer: %w", err)
    }
    
    return pm.ID, nil
}

func (p *StripeProcessor) SetDefaultPaymentMethod(customerID string, paymentMethodID string) error {
    // Set the payment method as the default for the customer
    params := &stripe.CustomerParams{
        InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
            DefaultPaymentMethod: stripe.String(paymentMethodID),
        },
    }
    
    _, err := customer.Update(customerID, params)
    if err != nil {
        return fmt.Errorf("failed to set default payment method: %w", err)
    }
    
    return nil
}

func (p *StripeProcessor) AttachPaymentMethod(paymentMethodID string, customerID string) error {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	_, err := paymentmethod.Attach(paymentMethodID, params)
	return err
}

func (p *StripeProcessor) AttachPaymentMethodIfNeeded(paymentMethodID string, customerID string) error {
    // First, try to get the payment method to see if it exists and if it's already attached
    pm, err := paymentmethod.Get(paymentMethodID, nil)
    if err != nil {
        return fmt.Errorf("failed to get payment method: %w", err)
    }
    
    // Check if it's already attached to this customer
    if pm.Customer != nil && pm.Customer.ID == customerID {
        return nil // Already attached to this customer
    }
    
    // If it's attached to another customer or not attached at all, try to attach it
    attachParams := &stripe.PaymentMethodAttachParams{
        Customer: stripe.String(customerID),
    }
    
    _, err = paymentmethod.Attach(paymentMethodID, attachParams)
    if err != nil {
        return fmt.Errorf("failed to attach payment method to customer: %w", err)
    }
    
    return nil
}

// ListPaymentMethods retrieves all payment methods for a customer
func (p *StripeProcessor) ListPaymentMethods(customerID string) ([]PaymentMethod, error) {
	// Create parameters for listing payment methods
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"), // Filter to card payment methods
	}
	
	// Create an iterator
	i := paymentmethod.List(params)
	
	// Get customer to check default payment method
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	
	// Default payment method ID (might be nil)
	var defaultPMID string
	if cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil {
		defaultPMID = *&cust.InvoiceSettings.DefaultPaymentMethod.ID
	}
	
	// Collect payment methods
	var paymentMethods []PaymentMethod
	for i.Next() {
		pm := i.PaymentMethod()
		
		// Create our model
		method := PaymentMethod{
			ID:        pm.ID,
			Type:      string(pm.Type),
			IsDefault: pm.ID == defaultPMID,
		}
		
		// Extract card details if present
		if pm.Card != nil {
			method.CardBrand = string(pm.Card.Brand)
			method.CardLast4 = pm.Card.Last4
			method.CardExpMonth = int(pm.Card.ExpMonth)
			method.CardExpYear = int(pm.Card.ExpYear)
		}
		
		// Extract billing details
		if pm.BillingDetails != nil && pm.BillingDetails.Name != "" {
			method.BillingName = pm.BillingDetails.Name
		}
		
		// Extract metadata
		if len(pm.Metadata) > 0 {
			method.Metadata = make(map[string]string)
			for k, v := range pm.Metadata {
				method.Metadata[k] = v
			}
		}
		
		paymentMethods = append(paymentMethods, method)
	}
	
	// Check for errors from the iterator
	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment methods: %w", err)
	}
	
	return paymentMethods, nil
}

// GetPaymentMethod retrieves a specific payment method
func (p *StripeProcessor) GetPaymentMethod(paymentMethodID string) (*PaymentMethod, error) {
	// Retrieve the payment method from Stripe
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	
	// Create our model
	method := &PaymentMethod{
		ID:   pm.ID,
		Type: string(pm.Type),
	}
	
	// Extract card details if present
	if pm.Card != nil {
		method.CardBrand = string(pm.Card.Brand)
		method.CardLast4 = pm.Card.Last4
		method.CardExpMonth = int(pm.Card.ExpMonth)
		method.CardExpYear = int(pm.Card.ExpYear)
	}
	
	// Extract billing details
	if pm.BillingDetails != nil && pm.BillingDetails.Name != "" {
		method.BillingName = pm.BillingDetails.Name
	}
	
	// Extract metadata
	if len(pm.Metadata) > 0 {
		method.Metadata = make(map[string]string)
		for k, v := range pm.Metadata {
			method.Metadata[k] = v
		}
	}
	
	// Check if this is the default payment method (if attached to a customer)
	if pm.Customer != nil && pm.Customer.ID != "" {
		cust, err := customer.Get(pm.Customer.ID, nil)
		if err == nil && cust.InvoiceSettings != nil && 
           cust.InvoiceSettings.DefaultPaymentMethod != nil && 
           *&cust.InvoiceSettings.DefaultPaymentMethod.ID == pm.ID {
			method.IsDefault = true
		}
	}
	
	return method, nil
}

// UpdatePaymentMethod updates a payment method's details
func (p *StripeProcessor) UpdatePaymentMethod(paymentMethodID, billingName string, metadata map[string]string) error {
	params := &stripe.PaymentMethodParams{}
	
	// Update billing details if provided
	if billingName != "" {
		params.BillingDetails = &stripe.PaymentMethodBillingDetailsParams {
			Name: stripe.String(billingName),
		}
	}
	
	// Update metadata if provided
	if metadata != nil {
		params.Metadata = make(map[string]string)
		for k, v := range metadata {
			params.AddMetadata(k, v)
		}
	}
	
	// Only make the API call if we have something to update
	if billingName != "" || metadata != nil {
		_, err := paymentmethod.Update(paymentMethodID, params)
		if err != nil {
			return fmt.Errorf("failed to update payment method: %w", err)
		}
	}
	
	return nil
}

// DetachPaymentMethod detaches a payment method from a customer
func (p *StripeProcessor) DetachPaymentMethod(paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}
	
	return nil
}

// HandleWebhook processes Stripe webhooks
func (p *StripeProcessor) HandleWebhook(body []byte, signature string) (interface{}, error) {
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(body, signature, endpointSecret)
	if err != nil {
		return nil, fmt.Errorf("webhook error: %w", err)
	}

	return event, nil
}
