// services/subscription_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
	"github.com/dukerupert/walking-drum/services/payment"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
)

// SubscriptionService handles subscription operations
type SubscriptionService struct {
	processor        payment.Processor
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	priceRepo        repository.PriceRepository
	invoiceRepo 	repository.InvoiceRepository
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	processor payment.Processor,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	priceRepo repository.PriceRepository,
	invoiceRepo repository.InvoiceRepository,
) *SubscriptionService {
	return &SubscriptionService{
		processor:        processor,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		priceRepo:        priceRepo,
		invoiceRepo: invoiceRepo,
	}
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	UserID          uuid.UUID         `json:"user_id"`
	PriceID         uuid.UUID         `json:"price_id"`
	Quantity        int               `json:"quantity"`
	PaymentMethodID string            `json:"payment_method_id,omitempty"`
	Description     string            `json:"description,omitempty"`
	OrderID         string            `json:"order_id,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// CreateSubscription creates a new subscription
func (s *SubscriptionService) CreateSubscription(req CreateSubscriptionRequest) (*models.Subscription, error) {
	// Fetch the user to get the Stripe customer ID
	user, err := s.userRepo.GetByID(context.Background(), req.UserID)
	if err != nil {
		return nil, err
	}

	if user.StripeCustomerID == nil {
		return nil, errors.New("user does not have a Stripe customer ID")
	}

	// Fetch the price to get the Stripe price ID
	price, err := s.priceRepo.GetByID(context.Background(), req.PriceID)
	if err != nil {
		return nil, err
	}

	if price.StripePriceID == nil {
		return nil, errors.New("price does not have a Stripe price ID")
	}

	// Create the payment processor request
	processorReq := payment.SubscriptionRequest{
		CustomerID:      *user.StripeCustomerID,
		PriceID:         *price.StripePriceID,
		Quantity:        int64(req.Quantity),
		PaymentMethodID: req.PaymentMethodID,
		Description:     req.Description,
		OrderID:         req.OrderID,
		Metadata:        req.Metadata,
	}

	// Create the subscription in the payment processor
	processorResp, err := s.processor.CreateSubscription(processorReq)
	if err != nil {
		return nil, err
	}

	// Create the subscription in our database
	now := time.Now()
	periodStart := time.Unix(processorResp.CurrentPeriodStart, 0)
	periodEnd := time.Unix(processorResp.CurrentPeriodEnd, 0)

	subscription := &models.Subscription{
		UserID:               req.UserID,
		PriceID:              req.PriceID,
		Quantity:             req.Quantity,
		Status:               models.SubscriptionStatus(processorResp.Status),
		CurrentPeriodStart:   periodStart,
		CurrentPeriodEnd:     periodEnd,
		StripeSubscriptionID: processorResp.ProcessorID,
		StripeCustomerID:     *user.StripeCustomerID,
		CollectionMethod:     "charge_automatically", // Default for Stripe
		CancelAtPeriodEnd:    processorResp.CancelAtPeriodEnd,
		CreatedAt:            now, // Set the creation timestamp
		UpdatedAt:            now, // Set the update timestamp
	}

	// Save to database
	err = s.subscriptionRepo.Create(context.Background(), subscription)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(subscriptionID uuid.UUID) (*models.Subscription, error) {
	// Fetch the subscription
	subscription, err := s.subscriptionRepo.GetByID(context.Background(), subscriptionID)
	if err != nil {
		return nil, err
	}

	// Cancel the subscription in the payment processor
	_, err = s.processor.CancelSubscription(subscription.StripeSubscriptionID)
	if err != nil {
		return nil, err
	}

	// Update the subscription in our database
	now := time.Now()
	subscription.Status = models.SubscriptionStatusCanceled
	subscription.CanceledAt = &now
	subscription.EndedAt = &now
	subscription.UpdatedAt = now

	// Save to database
	err = s.subscriptionRepo.Update(context.Background(), subscription)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// ProcessWebhook processes a webhook event from the payment processor
func (s *SubscriptionService) ProcessWebhook(body []byte, signature string) (interface{}, error) {
    // Pass the webhook to the payment processor for verification and parsing
    event, err := s.processor.HandleWebhook(body, signature)
    if err != nil {
        return nil, err
    }
    
    // Type assertion to get the Stripe event
    stripeEvent, ok := event.(stripe.Event)
    if !ok {
        return nil, errors.New("unexpected event type from processor")
    }
    
    // Process different event types
    switch stripeEvent.Type {
    case "customer.subscription.created":
        var subscription stripe.Subscription
        err := json.Unmarshal(stripeEvent.Data.Raw, &subscription)
        if err != nil {
            return nil, errors.New("error parsing webhook JSON")
        }
        
        // Handle the subscription created event
        err = s.handleSubscriptionCreated(&subscription)
        if err != nil {
            return nil, err
        }
        
    case "customer.subscription.updated":
        var subscription stripe.Subscription
        err := json.Unmarshal(stripeEvent.Data.Raw, &subscription)
        if err != nil {
            return nil, errors.New("error parsing webhook JSON")
        }
        
        // Handle the subscription updated event
        err = s.handleSubscriptionUpdated(&subscription)
        if err != nil {
            return nil, err
        }
        
    case "customer.subscription.deleted":
        var subscription stripe.Subscription
        err := json.Unmarshal(stripeEvent.Data.Raw, &subscription)
        if err != nil {
            return nil, errors.New("error parsing webhook JSON")
        }
        
        // Handle the subscription deleted event
        err = s.handleSubscriptionDeleted(&subscription)
        if err != nil {
            return nil, err
        }
        
    case "invoice.paid":
        var invoice stripe.Invoice
        err := json.Unmarshal(stripeEvent.Data.Raw, &invoice)
        if err != nil {
            return nil, errors.New("error parsing webhook JSON")
        }
        
        // Handle the invoice paid event
        err = s.handleInvoicePaid(&invoice)
        if err != nil {
            return nil, err
        }
        
    case "invoice.payment_failed":
        var invoice stripe.Invoice
        err := json.Unmarshal(stripeEvent.Data.Raw, &invoice)
        if err != nil {
            return nil, errors.New("error parsing webhook JSON")
        }
        
        // Handle the invoice payment failed event
        err = s.handleInvoicePaymentFailed(&invoice)
        if err != nil {
            return nil, err
        }
        
    // Add more event types as needed
        
    default:
        // We'll log unexpected event types but not treat them as errors
        log.Printf("Unhandled event type: %s\n", stripeEvent.Type)
    }
    
    return event, nil
}

// Handler methods for specific events
func (s *SubscriptionService) handleSubscriptionCreated(subscription *stripe.Subscription) error {
    // First check if we already have this subscription in our database
    existingSub, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
    if err == nil && existingSub != nil {
        // We already have this subscription, nothing to do
        return nil
    }
    
    // Find the user by Stripe customer ID
    user, err := s.userRepo.GetByStripeCustomerID(context.Background(), subscription.Customer.ID)
    if err != nil {
        return fmt.Errorf("user with Stripe ID %s not found: %w", subscription.Customer.ID, err)
    }
    
    // Create a new subscription record
    now := time.Now()
    newSubscription := &models.Subscription{
        UserID:               user.ID,
        Status:               models.SubscriptionStatus(subscription.Status),
        CurrentPeriodStart:   time.Unix(subscription.CurrentPeriodStart, 0),
        CurrentPeriodEnd:     time.Unix(subscription.CurrentPeriodEnd, 0),
        StripeSubscriptionID: subscription.ID,
        StripeCustomerID:     subscription.Customer.ID,
        CollectionMethod:     string(subscription.CollectionMethod),
        CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
        CreatedAt:            now,
        UpdatedAt:            now,
    }
    
    // Handle trial periods if present
    if subscription.TrialStart > 0 {
        trialStart := time.Unix(subscription.TrialStart, 0)
        newSubscription.TrialStart = &trialStart
    }
    
    if subscription.TrialEnd > 0 {
        trialEnd := time.Unix(subscription.TrialEnd, 0)
        newSubscription.TrialEnd = &trialEnd
    }
    
    // Handle cancellation data if present
    if subscription.CanceledAt > 0 {
        canceledAt := time.Unix(subscription.CanceledAt, 0)
        newSubscription.CanceledAt = &canceledAt
    }
    
    // For each item in the subscription, set the PriceID
    // Note: We're just using the first item for simplicity
    if len(subscription.Items.Data) > 0 {
        stripePriceID := subscription.Items.Data[0].Price.ID
        
        // Look up our internal price ID based on the Stripe price ID
        price, err := s.priceRepo.GetByStripePriceID(context.Background(), stripePriceID)
        if err != nil {
            return fmt.Errorf("price with Stripe ID %s not found: %w", stripePriceID, err)
        }
        
        newSubscription.PriceID = price.ID
        existingSub.Quantity = int(subscription.Items.Data[0].Quantity)
    }
    
    // Create the subscription in our database
    return s.subscriptionRepo.Create(context.Background(), newSubscription)
}

func (s *SubscriptionService) handleSubscriptionUpdated(subscription *stripe.Subscription) error {
    // Find the subscription in our database
    existingSub, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
    if err != nil {
        return fmt.Errorf("subscription with Stripe ID %s not found: %w", subscription.ID, err)
    }
    
    // Update the subscription
    existingSub.Status = models.SubscriptionStatus(subscription.Status)
    existingSub.CurrentPeriodStart = time.Unix(subscription.CurrentPeriodStart, 0)
    existingSub.CurrentPeriodEnd = time.Unix(subscription.CurrentPeriodEnd, 0)
    existingSub.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd
    existingSub.UpdatedAt = time.Now()
    
    // Handle cancellation
    if subscription.CanceledAt > 0 {
        canceledAt := time.Unix(subscription.CanceledAt, 0)
        existingSub.CanceledAt = &canceledAt
    }
    
    // Handle cancel_at
    if subscription.CancelAt > 0 {
        cancelAt := time.Unix(subscription.CancelAt, 0)
        existingSub.CancelAt = &cancelAt
    }
    
    // Handle ended_at
    if subscription.EndedAt > 0 {
        endedAt := time.Unix(subscription.EndedAt, 0)
        existingSub.EndedAt = &endedAt
    }
    
    // If items changed, update the price and quantity
    if len(subscription.Items.Data) > 0 {
        stripePriceID := subscription.Items.Data[0].Price.ID
        
        // Check if the price changed
        price, err := s.priceRepo.GetByStripePriceID(context.Background(), stripePriceID)
        if err == nil && price.ID != existingSub.PriceID {
            existingSub.PriceID = price.ID
        }
        
        // Update quantity
        existingSub.Quantity = int(subscription.Items.Data[0].Quantity)
    }
    
    // Update the subscription in our database
    return s.subscriptionRepo.Update(context.Background(), existingSub)
}

func (s *SubscriptionService) handleSubscriptionDeleted(subscription *stripe.Subscription) error {
    // Find the subscription in our database
    existingSub, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
    if err != nil {
        return fmt.Errorf("subscription with Stripe ID %s not found: %w", subscription.ID, err)
    }
    
    // Update the subscription status
    now := time.Now()
    existingSub.Status = models.SubscriptionStatusCanceled
    existingSub.CanceledAt = &now
    existingSub.EndedAt = &now
    existingSub.UpdatedAt = now
    
    // Update the subscription in our database
    return s.subscriptionRepo.Update(context.Background(), existingSub)
}

func (s *SubscriptionService) handleInvoicePaid(invoice *stripe.Invoice) error {
    // If this is a subscription invoice, update the subscription status
    if invoice.Subscription != nil {
        // Find the subscription in our database
        subscription, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
        if err != nil {
            return fmt.Errorf("subscription with Stripe ID %s not found: %w", invoice.Subscription.ID, err)
        }
        
        // Update the subscription if needed (e.g., ensure it's active)
        if subscription.Status != models.SubscriptionStatusActive {
            subscription.Status = models.SubscriptionStatusActive
            subscription.UpdatedAt = time.Now()
            
            err = s.subscriptionRepo.Update(context.Background(), subscription)
            if err != nil {
                return err
            }
        }
    }
    
    // Create a new invoice record in our database
    now := time.Now()
    newInvoice := &models.Invoice{
        UserID:          uuid.UUID{}, // Need to look up the user
        Status:          models.InvoiceStatusPaid,
        AmountDue:       invoice.AmountDue,
        AmountPaid:      invoice.AmountPaid,
        Currency:        string(invoice.Currency),
        StripeInvoiceID: invoice.ID,
        CreatedAt:       now,
        UpdatedAt:       now,
    }
    
    // Set subscription ID if applicable
    if invoice.Subscription != nil {
        subscription, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
        if err == nil {
            newInvoice.SubscriptionID = &subscription.ID
            newInvoice.UserID = subscription.UserID
        }
    }
    
    // Set period start/end if available
    if invoice.PeriodStart > 0 {
        periodStart := time.Unix(invoice.PeriodStart, 0)
        newInvoice.PeriodStart = &periodStart
    }
    
    if invoice.PeriodEnd > 0 {
        periodEnd := time.Unix(invoice.PeriodEnd, 0)
        newInvoice.PeriodEnd = &periodEnd
    }
    
    // Set payment intent ID if available
    if invoice.PaymentIntent != nil {
        newInvoice.PaymentIntentID = &invoice.PaymentIntent.ID
    }
    
    // Check if we already have this invoice
    existingInvoice, err := s.invoiceRepo.GetByStripeInvoiceID(context.Background(), invoice.ID)
    if err == nil && existingInvoice != nil {
        // Invoice exists, update it instead
        existingInvoice.Status = models.InvoiceStatusPaid
        existingInvoice.AmountPaid = invoice.AmountPaid
        existingInvoice.UpdatedAt = now
        
        return s.invoiceRepo.Update(context.Background(), existingInvoice)
    }
    
    // Create new invoice
    return s.invoiceRepo.Create(context.Background(), newInvoice)
}

func (s *SubscriptionService) handleInvoicePaymentFailed(invoice *stripe.Invoice) error {
    // If this is a subscription invoice, update the subscription status
    if invoice.Subscription != nil {
        // Find the subscription in our database
        subscription, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
        if err != nil {
            return fmt.Errorf("subscription with Stripe ID %s not found: %w", invoice.Subscription.ID, err)
        }
        
        // Update the subscription status
        subscription.Status = models.SubscriptionStatusPastDue
        subscription.UpdatedAt = time.Now()
        
        err = s.subscriptionRepo.Update(context.Background(), subscription)
        if err != nil {
            return err
        }
    }
    
    // Update or create the invoice record
    existingInvoice, err := s.invoiceRepo.GetByStripeInvoiceID(context.Background(), invoice.ID)
    now := time.Now()
    
    if err == nil && existingInvoice != nil {
        // Invoice exists, update it
        existingInvoice.Status = models.InvoiceStatusUncollectible
        existingInvoice.UpdatedAt = now
        
        return s.invoiceRepo.Update(context.Background(), existingInvoice)
    }
    
    // Create a new invoice record
    newInvoice := &models.Invoice{
        UserID:          uuid.UUID{}, // Need to look up the user
        Status:          models.InvoiceStatusUncollectible,
        AmountDue:       invoice.AmountDue,
        AmountPaid:      invoice.AmountPaid,
        Currency:        string(invoice.Currency),
        StripeInvoiceID: invoice.ID,
        CreatedAt:       now,
        UpdatedAt:       now,
    }
    
    // Set subscription ID if applicable
    if invoice.Subscription != nil {
        subscription, err := s.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
        if err == nil {
            newInvoice.SubscriptionID = &subscription.ID
            newInvoice.UserID = subscription.UserID
        }
    }
    
    // Create new invoice
    return s.invoiceRepo.Create(context.Background(), newInvoice)
}
