// services/stripe/webhooks.go
package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// WebhookHandler handles Stripe webhook events
type WebhookHandler struct {
	// Dependencies would go here, such as your repositories
	// subscriptionRepo repository.SubscriptionRepository
	// invoiceRepo      repository.InvoiceRepository
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		// Initialize dependencies
	}
}

// HandleWebhook handles Stripe webhook events
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the webhook payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		http.Error(w, fmt.Sprintf("Webhook error: %v", err), http.StatusBadRequest)
		return
	}

	// Handle different event types
	switch event.Type {
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		
		// Process the subscription created event
		// This would typically involve updating your database
		// For example:
		// h.handleSubscriptionCreated(&subscription)
		
	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		
		// Process the subscription updated event
		// h.handleSubscriptionUpdated(&subscription)
		
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		
		// Process the subscription deleted event
		// h.handleSubscriptionDeleted(&subscription)
		
	case "invoice.paid":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		
		// Process the invoice paid event
		// h.handleInvoicePaid(&invoice)
		
	case "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
			return
		}
		
		// Process the invoice payment failed event
		// h.handleInvoicePaymentFailed(&invoice)
		
	// Add more event types as needed
		
	default:
		// Unexpected event type
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	// Return a 200 success response to Stripe
	w.WriteHeader(http.StatusOK)
}

// Add implementation of event handlers below
// func (h *WebhookHandler) handleSubscriptionCreated(subscription *stripe.Subscription) {
//     // Implementation
// }