// services/subscription_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
	"github.com/dukerupert/walking-drum/services/payment"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v74"
)

// SubscriptionService handles subscription operations
type SubscriptionService struct {
	processor        payment.Processor
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	priceRepo        repository.PriceRepository
	productRepo      repository.ProductRepository
	invoiceRepo      repository.InvoiceRepository
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	processor payment.Processor,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	priceRepo repository.PriceRepository,
	productRepo repository.ProductRepository,
	invoiceRepo repository.InvoiceRepository,
) *SubscriptionService {
	return &SubscriptionService{
		processor:        processor,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		priceRepo:        priceRepo,
		productRepo:      productRepo,
		invoiceRepo:      invoiceRepo,
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
    // Sync the user with Stripe to ensure they have a Stripe customer ID
    customerID, err := s.SyncUserWithStripe(req.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to sync user with Stripe: %w", err)
    }

    // Sync the price with Stripe
    stripePriceID, err := s.SyncPriceWithStripe(req.PriceID)
    if err != nil {
        return nil, fmt.Errorf("failed to sync price with Stripe: %w", err)
    }

    // Check if a payment method was provided
    var paymentMethodID string
    if req.PaymentMethodID != "" {
        log.Info().Str("payment_method_id", req.PaymentMethodID).Msg("Payment method ID provided")
        
        // For Stripe test mode, we can use tok_visa directly
        if req.PaymentMethodID == "tok_visa" {
            // Create a payment method using the token
            pmReq := payment.PaymentMethodRequest{
                CustomerID: customerID,
                Type:       "card",
                Token:      req.PaymentMethodID,  // Use the token
            }
            
            log.Debug().
                Str("customer_id", customerID).
                Str("token", req.PaymentMethodID).
                Msg("Creating payment method from token")
                
            createdPM, err := s.processor.CreatePaymentMethod(pmReq)
            if err != nil {
                return nil, fmt.Errorf("failed to create payment method from token: %w", err)
            }
            
            paymentMethodID = createdPM
            
            log.Info().
                Str("payment_method_id", paymentMethodID).
                Str("customer_id", customerID).
                Msg("Created payment method from token")
                
        } else if strings.HasPrefix(req.PaymentMethodID, "pm_") {
            // This is already a Stripe payment method ID, use it directly
            log.Info().Msg("Using existing payment method ID")
            paymentMethodID = req.PaymentMethodID
            
            // Ensure it's attached to the customer
            err := s.processor.AttachPaymentMethodIfNeeded(paymentMethodID, customerID)
            if err != nil {
                log.Warn().
                    Err(err).
                    Str("payment_method_id", paymentMethodID).
                    Str("customer_id", customerID).
                    Msg("Failed to attach payment method to customer")
                    
                // Continue anyway - it might already be attached
            }
        } else {
            // Not a valid payment method ID format, just log and continue
            log.Warn().
                Str("payment_method_id", req.PaymentMethodID).
                Msg("Invalid payment method ID format, proceeding without payment method")
        }
    } else {
        log.Info().Msg("No payment method ID provided")
    }

    // Create the payment processor request
    processorReq := payment.SubscriptionRequest{
        CustomerID:      customerID,
        PriceID:         stripePriceID,
        Quantity:        int64(req.Quantity),
        PaymentMethodID: paymentMethodID,
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
        StripeCustomerID:     customerID,
        CollectionMethod:     "charge_automatically", // Default for Stripe
        CancelAtPeriodEnd:    processorResp.CancelAtPeriodEnd,
        CreatedAt:            now,
        UpdatedAt:            now,
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

// SyncUserWithStripe ensures a user has a Stripe customer ID
func (s *SubscriptionService) SyncUserWithStripe(userID uuid.UUID) (string, error) {
	log.Debug().Str("user_id", userID.String()).Msg("Starting SyncUserWithStripe")

	// Get the user
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user")
		return "", err
	}

	log.Debug().
		Str("user_id", userID.String()).
		Str("email", user.Email).
		Str("name", user.Name).
		Msg("Found user")

	// If user already has a Stripe customer ID and it's valid, return it
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		// Check if the ID starts with "cus_" which is the Stripe prefix for customers
		if strings.HasPrefix(*user.StripeCustomerID, "cus_") {
			log.Debug().
				Str("user_id", userID.String()).
				Str("stripe_customer_id", *user.StripeCustomerID).
				Msg("User already has a Stripe customer ID")

			// Try to retrieve the customer from Stripe to verify it exists
			params := &stripe.CustomerParams{}
			_, err := s.processor.RetrieveCustomer(*user.StripeCustomerID, params)
			if err == nil {
				log.Debug().
					Str("user_id", userID.String()).
					Str("stripe_customer_id", *user.StripeCustomerID).
					Msg("Verified customer exists in Stripe")
				return *user.StripeCustomerID, nil
			}

			log.Warn().
				Err(err).
				Str("user_id", userID.String()).
				Str("stripe_customer_id", *user.StripeCustomerID).
				Msg("Customer ID exists in database but not in Stripe, creating new one")
		} else {
			log.Warn().
				Str("user_id", userID.String()).
				Str("stripe_customer_id", *user.StripeCustomerID).
				Msg("User has invalid Stripe customer ID format, creating new one")
		}
	} else {
		log.Debug().
			Str("user_id", userID.String()).
			Msg("User does not have a Stripe customer ID, creating one")
	}

	// User doesn't have a valid Stripe customer ID, create one
	customerReq := payment.CustomerRequest{
		Email:       user.Email,
		Name:        user.Name,
		Description: "Walking-Drum customer",
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
	}

	log.Debug().
		Str("user_id", userID.String()).
		Str("email", user.Email).
		Msg("Creating new Stripe customer")

	customerID, err := s.processor.CreateCustomer(customerReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("Failed to create Stripe customer")
		return "", err
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("stripe_customer_id", customerID).
		Msg("Successfully created Stripe customer")

	// Update the user with the new Stripe customer ID
	user.StripeCustomerID = &customerID
	err = s.userRepo.Update(context.Background(), user)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Str("stripe_customer_id", customerID).
			Msg("Failed to update user with Stripe customer ID")
		return "", err
	}

	log.Debug().
		Str("user_id", userID.String()).
		Str("stripe_customer_id", customerID).
		Msg("Updated user with Stripe customer ID")

	return customerID, nil
}

func (s *SubscriptionService) SyncPriceWithStripe(priceID uuid.UUID) (string, error) {
	log.Debug().Str("price_id", priceID.String()).Msg("Starting SyncPriceWithStripe")

	// Get the price
	price, err := s.priceRepo.GetByID(context.Background(), priceID)
	if err != nil {
		log.Error().Err(err).Str("price_id", priceID.String()).Msg("Failed to get price")
		return "", err
	}

	log.Debug().
		Str("price_id", priceID.String()).
		Int64("amount", price.Amount).
		Str("currency", price.Currency).
		Msg("Found price")

	// If price already has a Stripe price ID and it's valid, return it
	if price.StripePriceID != nil && *price.StripePriceID != "" {
		// Check if the ID starts with "price_" which is the Stripe prefix for prices
		if strings.HasPrefix(*price.StripePriceID, "price_") {
			log.Debug().
				Str("price_id", priceID.String()).
				Str("stripe_price_id", *price.StripePriceID).
				Msg("Price already has a Stripe price ID")

			// Try to retrieve the price from Stripe to verify it exists
			params := &stripe.PriceParams{}
			_, err := s.processor.RetrievePrice(*price.StripePriceID, params)
			if err == nil {
				log.Debug().
					Str("price_id", priceID.String()).
					Str("stripe_price_id", *price.StripePriceID).
					Msg("Verified price exists in Stripe")
				return *price.StripePriceID, nil
			}

			log.Warn().
				Err(err).
				Str("price_id", priceID.String()).
				Str("stripe_price_id", *price.StripePriceID).
				Msg("Price ID exists in database but not in Stripe, creating new one")
		} else {
			log.Warn().
				Str("price_id", priceID.String()).
				Str("stripe_price_id", *price.StripePriceID).
				Msg("Price has invalid Stripe price ID format, creating new one")
		}
	} else {
		log.Debug().
			Str("price_id", priceID.String()).
			Msg("Price does not have a Stripe price ID, creating one")
	}

	// Get the product
	product, err := s.productRepo.GetByID(context.Background(), price.ProductID)
	if err != nil {
		log.Error().Err(err).Str("product_id", price.ProductID.String()).Msg("Failed to get product")
		return "", err
	}

	// Check if product has Stripe ID, if not, use name
	// Sync the product with Stripe
	stripeProductID, err := s.SyncProductWithStripe(price.ProductID)
	if err != nil {
		log.Error().
			Err(err).
			Str("price_id", priceID.String()).
			Str("product_id", price.ProductID.String()).
			Msg("Failed to sync product with Stripe")
		return "", fmt.Errorf("failed to sync product with Stripe: %w", err)
	}

	// Create price request
	priceReq := payment.PriceRequest{
		ProductID:  stripeProductID, // Use the synced Stripe product ID
		UnitAmount: price.Amount,
		Currency:   price.Currency,
		Recurring:  price.IntervalType != "",
		Metadata: map[string]string{
			"price_id":   price.ID.String(),
			"product_id": product.ID.String(),
		},
	}

	// Set nickname based on availability
	if price.Nickname != nil {
		priceReq.Nickname = *price.Nickname
	} else {
		priceReq.Nickname = product.Name
	}

	// Set recurring details if applicable
	if priceReq.Recurring {
		priceReq.IntervalType = string(price.IntervalType)
		priceReq.IntervalCount = int64(price.IntervalCount)
	}

	log.Debug().
		Str("price_id", priceID.String()).
		Str("product_id", stripeProductID).
		Int64("amount", price.Amount).
		Msg("Creating new Stripe price")

	stripePriceID, err := s.processor.CreatePrice(priceReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("price_id", priceID.String()).
			Msg("Failed to create Stripe price")
		return "", err
	}

	log.Info().
		Str("price_id", priceID.String()).
		Str("stripe_price_id", stripePriceID).
		Msg("Successfully created Stripe price")

	// Update the price with the new Stripe price ID
	price.StripePriceID = &stripePriceID
	err = s.priceRepo.Update(context.Background(), price)
	if err != nil {
		log.Error().
			Err(err).
			Str("price_id", priceID.String()).
			Str("stripe_price_id", stripePriceID).
			Msg("Failed to update price with Stripe price ID")
		return "", err
	}

	log.Debug().
		Str("price_id", priceID.String()).
		Str("stripe_price_id", stripePriceID).
		Msg("Updated price with Stripe price ID")

	return stripePriceID, nil
}

func (s *SubscriptionService) SyncProductWithStripe(productID uuid.UUID) (string, error) {
	log.Debug().Str("product_id", productID.String()).Msg("Starting SyncProductWithStripe")

	// Get the product
	product, err := s.productRepo.GetByID(context.Background(), productID)
	if err != nil {
		log.Error().Err(err).Str("product_id", productID.String()).Msg("Failed to get product")
		return "", err
	}

	log.Debug().
		Str("product_id", productID.String()).
		Str("name", product.Name).
		Msg("Found product")

	// If product already has a Stripe product ID and it's valid, return it
	if product.StripeProductID != nil && *product.StripeProductID != "" {
		// Check if the ID starts with "prod_" which is the Stripe prefix for products
		if strings.HasPrefix(*product.StripeProductID, "prod_") {
			log.Debug().
				Str("product_id", productID.String()).
				Str("stripe_product_id", *product.StripeProductID).
				Msg("Product already has a Stripe product ID")

			// Try to retrieve the product from Stripe to verify it exists
			params := &stripe.ProductParams{}
			_, err := s.processor.RetrieveProduct(*product.StripeProductID, params)
			if err == nil {
				log.Debug().
					Str("product_id", productID.String()).
					Str("stripe_product_id", *product.StripeProductID).
					Msg("Verified product exists in Stripe")
				return *product.StripeProductID, nil
			}

			log.Warn().
				Err(err).
				Str("product_id", productID.String()).
				Str("stripe_product_id", *product.StripeProductID).
				Msg("Product ID exists in database but not in Stripe, creating new one")
		} else {
			log.Warn().
				Str("product_id", productID.String()).
				Str("stripe_product_id", *product.StripeProductID).
				Msg("Product has invalid Stripe product ID format, creating new one")
		}
	} else {
		log.Debug().
			Str("product_id", productID.String()).
			Msg("Product does not have a Stripe product ID, creating one")
	}

	// Create product request
	productReq := payment.ProductRequest{
		Name:        product.Name,
		Description: product.Description,
		Active:      product.IsActive,
		Metadata: map[string]string{
			"product_id": product.ID.String(),
		},
	}

	log.Debug().
		Str("product_id", productID.String()).
		Str("name", product.Name).
		Msg("Creating new Stripe product")

	stripeProductID, err := s.processor.CreateProduct(productReq)
	if err != nil {
		log.Error().
			Err(err).
			Str("product_id", productID.String()).
			Msg("Failed to create Stripe product")
		return "", err
	}

	log.Info().
		Str("product_id", productID.String()).
		Str("stripe_product_id", stripeProductID).
		Msg("Successfully created Stripe product")

	// Update the product with the new Stripe product ID
	product.StripeProductID = &stripeProductID
	err = s.productRepo.Update(context.Background(), product)
	if err != nil {
		log.Error().
			Err(err).
			Str("product_id", productID.String()).
			Str("stripe_product_id", stripeProductID).
			Msg("Failed to update product with Stripe product ID")
		return "", err
	}

	log.Debug().
		Str("product_id", productID.String()).
		Str("stripe_product_id", stripeProductID).
		Msg("Updated product with Stripe product ID")

	return stripeProductID, nil
}
