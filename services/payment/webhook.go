// services/payment/webhook_handler.go
package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	
	"github.com/dukerupert/walking-drum/models"
	"github.com/dukerupert/walking-drum/repository"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	processor        Processor
	subscriptionRepo repository.SubscriptionRepository
	invoiceRepo      repository.InvoiceRepository
	userRepo         repository.UserRepository
	productRepo      repository.ProductRepository
	priceRepo        repository.PriceRepository
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	processor Processor,
	subscriptionRepo repository.SubscriptionRepository,
	invoiceRepo repository.InvoiceRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	priceRepo repository.PriceRepository,
) *WebhookHandler {
	return &WebhookHandler{
		processor:        processor,
		subscriptionRepo: subscriptionRepo,
		invoiceRepo:      invoiceRepo,
		userRepo:         userRepo,
		productRepo:      productRepo,
		priceRepo:        priceRepo,
	}
}

// HandleWebhook processes Stripe webhook events
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading webhook request body")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Get the webhook secret from environment
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Error().Msg("Stripe webhook secret not set")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Verify the webhook signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying webhook signature")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log the event type
	log.Info().Str("event_type", event.Type).Str("event_id", event.ID).Msg("Received Stripe webhook event")

	// Process the event based on its type
	var webhookErr error
	
	switch event.Type {
	// Customer events
	case "customer.created":
		webhookErr = h.handleCustomerCreated(event)
	case "customer.updated":
		webhookErr = h.handleCustomerUpdated(event)
	case "customer.deleted":
		webhookErr = h.handleCustomerDeleted(event)
		
	// Subscription events
	case "customer.subscription.created":
		webhookErr = h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		webhookErr = h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		webhookErr = h.handleSubscriptionDeleted(event)
	case "customer.subscription.paused":
		webhookErr = h.handleSubscriptionPaused(event)
	case "customer.subscription.resumed":
		webhookErr = h.handleSubscriptionResumed(event)
	case "customer.subscription.trial_will_end":
		webhookErr = h.handleSubscriptionTrialWillEnd(event)
		
	// Invoice events
	case "invoice.created":
		webhookErr = h.handleInvoiceCreated(event)
	case "invoice.finalized":
		webhookErr = h.handleInvoiceFinalized(event)
	case "invoice.paid":
		webhookErr = h.handleInvoicePaid(event)
	case "invoice.payment_failed":
		webhookErr = h.handleInvoicePaymentFailed(event)
	case "invoice.upcoming":
		webhookErr = h.handleInvoiceUpcoming(event)
		
	// Payment events
	case "payment_method.attached":
		webhookErr = h.handlePaymentMethodAttached(event)
	case "payment_method.detached":
		webhookErr = h.handlePaymentMethodDetached(event)
	case "payment_intent.succeeded":
		webhookErr = h.handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		webhookErr = h.handlePaymentIntentFailed(event)
		
	// Product and price events
	case "product.created":
		webhookErr = h.handleProductCreated(event)
	case "product.updated":
		webhookErr = h.handleProductUpdated(event)
	case "price.created":
		webhookErr = h.handlePriceCreated(event)
	case "price.updated":
		webhookErr = h.handlePriceUpdated(event)
		
	default:
		log.Info().Str("event_type", event.Type).Msg("Unhandled event type")
	}

	if webhookErr != nil {
		log.Error().Err(webhookErr).Str("event_type", event.Type).Msg("Error handling webhook event")
		// We still return 200 OK to Stripe so they don't retry
		// In a production environment, you might want to queue the event for retry
	}

	// Return a 200 success response to Stripe
	w.WriteHeader(http.StatusOK)
}

// Event handlers - implement the important ones, stub the rest

// Customer events
func (h *WebhookHandler) handleCustomerCreated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Customer created event received")
	// Nothing to do here as we create customers in our system first
	return nil
}

func (h *WebhookHandler) handleCustomerUpdated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Customer updated event received")
	// We could sync customer data if needed
	return nil
}

func (h *WebhookHandler) handleCustomerDeleted(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Customer deleted event received")
	// We might want to mark the customer as inactive in our system
	return nil
}

// Subscription events
func (h *WebhookHandler) handleSubscriptionCreated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription created event received")
	
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("error parsing subscription JSON: %w", err)
	}
	
	// Check if we already have this subscription in our database
	existingSub, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
	if err == nil && existingSub != nil {
		// We already have this subscription, nothing to do
		return nil
	}
	
	// Find the user by Stripe customer ID
	user, err := h.userRepo.GetByStripeCustomerID(context.Background(), subscription.Customer.ID)
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
		price, err := h.priceRepo.GetByStripePriceID(context.Background(), stripePriceID)
		if err != nil {
			return fmt.Errorf("price with Stripe ID %s not found: %w", stripePriceID, err)
		}
		
		newSubscription.PriceID = price.ID
		newSubscription.Quantity = int(subscription.Items.Data[0].Quantity)
	}
	
	// Create the subscription in our database
	return h.subscriptionRepo.Create(context.Background(), newSubscription)
}

func (h *WebhookHandler) handleSubscriptionUpdated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription updated event received")
	
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("error parsing subscription JSON: %w", err)
	}
	
	// Find the subscription in our database
	existingSub, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
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
		price, err := h.priceRepo.GetByStripePriceID(context.Background(), stripePriceID)
		if err == nil && price.ID != existingSub.PriceID {
			existingSub.PriceID = price.ID
		}
		
		// Update quantity
		existingSub.Quantity = int(subscription.Items.Data[0].Quantity)
	}
	
	// Update the subscription in our database
	return h.subscriptionRepo.Update(context.Background(), existingSub)
}

func (h *WebhookHandler) handleSubscriptionDeleted(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription deleted event received")
	
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		return fmt.Errorf("error parsing subscription JSON: %w", err)
	}
	
	// Find the subscription in our database
	existingSub, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), subscription.ID)
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
	return h.subscriptionRepo.Update(context.Background(), existingSub)
}

func (h *WebhookHandler) handleSubscriptionPaused(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription paused event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handleSubscriptionResumed(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription resumed event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handleSubscriptionTrialWillEnd(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Subscription trial will end event received")
	// Stub - log only
	// In a real application, you might want to send a notification to the customer
	return nil
}

// Invoice events
func (h *WebhookHandler) handleInvoiceCreated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Invoice created event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handleInvoiceFinalized(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Invoice finalized event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handleInvoicePaid(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Invoice paid event received")
	
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		return fmt.Errorf("error parsing invoice JSON: %w", err)
	}
	
	// If this is a subscription invoice, update the subscription status
	if invoice.Subscription != nil {
		// Find the subscription in our database
		subscription, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
		if err != nil {
			return fmt.Errorf("subscription with Stripe ID %s not found: %w", invoice.Subscription.ID, err)
		}
		
		// Update the subscription if needed (e.g., ensure it's active)
		if subscription.Status != models.SubscriptionStatusActive {
			subscription.Status = models.SubscriptionStatusActive
			subscription.UpdatedAt = time.Now()
			
			err = h.subscriptionRepo.Update(context.Background(), subscription)
			if err != nil {
				return err
			}
		}
	}
	
	// Create a new invoice record in our database
	now := time.Now()
	newInvoice := &models.Invoice{
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
		subscription, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
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
	existingInvoice, err := h.invoiceRepo.GetByStripeInvoiceID(context.Background(), invoice.ID)
	if err == nil && existingInvoice != nil {
		// Invoice exists, update it instead
		existingInvoice.Status = models.InvoiceStatusPaid
		existingInvoice.AmountPaid = invoice.AmountPaid
		existingInvoice.UpdatedAt = now
		
		return h.invoiceRepo.Update(context.Background(), existingInvoice)
	}
	
	// Create new invoice
	return h.invoiceRepo.Create(context.Background(), newInvoice)
}

func (h *WebhookHandler) handleInvoicePaymentFailed(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Invoice payment failed event received")
	
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		return fmt.Errorf("error parsing invoice JSON: %w", err)
	}
	
	// If this is a subscription invoice, update the subscription status
	if invoice.Subscription != nil {
		// Find the subscription in our database
		subscription, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
		if err != nil {
			return fmt.Errorf("subscription with Stripe ID %s not found: %w", invoice.Subscription.ID, err)
		}
		
		// Update the subscription status
		subscription.Status = models.SubscriptionStatusPastDue
		subscription.UpdatedAt = time.Now()
		
		err = h.subscriptionRepo.Update(context.Background(), subscription)
		if err != nil {
			return err
		}
	}
	
	// Update or create the invoice record
	existingInvoice, err := h.invoiceRepo.GetByStripeInvoiceID(context.Background(), invoice.ID)
	now := time.Now()
	
	if err == nil && existingInvoice != nil {
		// Invoice exists, update it
		existingInvoice.Status = models.InvoiceStatusUncollectible
		existingInvoice.UpdatedAt = now
		
		return h.invoiceRepo.Update(context.Background(), existingInvoice)
	}
	
	// Create a new invoice record
	newInvoice := &models.Invoice{
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
		subscription, err := h.subscriptionRepo.GetByStripeSubscriptionID(context.Background(), invoice.Subscription.ID)
		if err == nil {
			newInvoice.SubscriptionID = &subscription.ID
			newInvoice.UserID = subscription.UserID
		}
	}
	
	// Create new invoice
	return h.invoiceRepo.Create(context.Background(), newInvoice)
}

func (h *WebhookHandler) handleInvoiceUpcoming(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Invoice upcoming event received")
	// Stub - log only
	// In a real application, you might want to send a notification to the customer
	return nil
}

// Payment events
func (h *WebhookHandler) handlePaymentMethodAttached(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Payment method attached event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handlePaymentMethodDetached(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Payment method detached event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handlePaymentIntentSucceeded(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Payment intent succeeded event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handlePaymentIntentFailed(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Payment intent failed event received")
	// Stub - log only
	return nil
}

// Product and price events
func (h *WebhookHandler) handleProductCreated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Product created event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handleProductUpdated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Product updated event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handlePriceCreated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Price created event received")
	// Stub - log only
	return nil
}

func (h *WebhookHandler) handlePriceUpdated(event stripe.Event) error {
	log.Info().Str("event_id", event.ID).Msg("Price updated event received")
	// Stub - log only
	return nil
}