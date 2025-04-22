// server.go
//
// Use this sample code to handle webhook events in your integration.
//
// 1) Create a new Go module
//   go mod init example.com/stripe/webhooks/example
//
// 2) Paste this code into a new file (server.go)
//
// 3) Install dependencies
//   go get -u github.com/stripe/stripe-go
//
// 4) Run the server on http://localhost:4242
//   go run server.go

package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"

  "github.com/stripe/stripe-go"
  "github.com/stripe/stripe-go/webhook"
)

// The library needs to be configured with your account's secret key.
// Ensure the key is kept out of any version control system you might be using.
stripe.Key = "sk_test_..."

func main() {
  http.HandleFunc("/webhook", handleWebhook)
  addr := "localhost:4242"
  log.Printf("Listening on %s", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}

func handleWebhook(w http.ResponseWriter, req *http.Request) {
  const MaxBodyBytes = int64(65536)
  req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
  payload, err := ioutil.ReadAll(req.Body)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
    w.WriteHeader(http.StatusServiceUnavailable)
    return
  }

  // This is your Stripe CLI webhook secret for testing your endpoint locally.
  endpointSecret := "whsec_9f43eadf4077d339ef631702cd7f378fc95ff269dcd90dba4410e3b42fbfe301";
  // Pass the request body and Stripe-Signature header to ConstructEvent, along
  // with the webhook signing key.
  event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
    endpointSecret)

  if err != nil {
    fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
    w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
    return
  }

  // Unmarshal the event data into an appropriate struct depending on its Type
  switch event.Type {
  case "charge.captured":
    // Then define and call a function to handle the event charge.captured
  case "charge.expired":
    // Then define and call a function to handle the event charge.expired
  case "charge.failed":
    // Then define and call a function to handle the event charge.failed
  case "charge.pending":
    // Then define and call a function to handle the event charge.pending
  case "charge.refunded":
    // Then define and call a function to handle the event charge.refunded
  case "charge.succeeded":
    // Then define and call a function to handle the event charge.succeeded
  case "charge.updated":
    // Then define and call a function to handle the event charge.updated
  case "charge.dispute.closed":
    // Then define and call a function to handle the event charge.dispute.closed
  case "charge.dispute.created":
    // Then define and call a function to handle the event charge.dispute.created
  case "charge.dispute.funds_reinstated":
    // Then define and call a function to handle the event charge.dispute.funds_reinstated
  case "charge.dispute.funds_withdrawn":
    // Then define and call a function to handle the event charge.dispute.funds_withdrawn
  case "charge.dispute.updated":
    // Then define and call a function to handle the event charge.dispute.updated
  case "charge.refund.updated":
    // Then define and call a function to handle the event charge.refund.updated
  case "checkout.session.async_payment_failed":
    // Then define and call a function to handle the event checkout.session.async_payment_failed
  case "checkout.session.async_payment_succeeded":
    // Then define and call a function to handle the event checkout.session.async_payment_succeeded
  case "checkout.session.completed":
    // Then define and call a function to handle the event checkout.session.completed
  case "checkout.session.expired":
    // Then define and call a function to handle the event checkout.session.expired
  case "customer.created":
    // Then define and call a function to handle the event customer.created
  case "customer.deleted":
    // Then define and call a function to handle the event customer.deleted
  case "customer.updated":
    // Then define and call a function to handle the event customer.updated
  case "customer.discount.created":
    // Then define and call a function to handle the event customer.discount.created
  case "customer.discount.deleted":
    // Then define and call a function to handle the event customer.discount.deleted
  case "customer.discount.updated":
    // Then define and call a function to handle the event customer.discount.updated
  case "customer.source.created":
    // Then define and call a function to handle the event customer.source.created
  case "customer.source.deleted":
    // Then define and call a function to handle the event customer.source.deleted
  case "customer.source.expiring":
    // Then define and call a function to handle the event customer.source.expiring
  case "customer.source.updated":
    // Then define and call a function to handle the event customer.source.updated
  case "customer.subscription.created":
    // Then define and call a function to handle the event customer.subscription.created
  case "customer.subscription.deleted":
    // Then define and call a function to handle the event customer.subscription.deleted
  case "customer.subscription.paused":
    // Then define and call a function to handle the event customer.subscription.paused
  case "customer.subscription.pending_update_applied":
    // Then define and call a function to handle the event customer.subscription.pending_update_applied
  case "customer.subscription.pending_update_expired":
    // Then define and call a function to handle the event customer.subscription.pending_update_expired
  case "customer.subscription.resumed":
    // Then define and call a function to handle the event customer.subscription.resumed
  case "customer.subscription.trial_will_end":
    // Then define and call a function to handle the event customer.subscription.trial_will_end
  case "customer.subscription.updated":
    // Then define and call a function to handle the event customer.subscription.updated
  case "customer.tax_id.created":
    // Then define and call a function to handle the event customer.tax_id.created
  case "customer.tax_id.deleted":
    // Then define and call a function to handle the event customer.tax_id.deleted
  case "customer.tax_id.updated":
    // Then define and call a function to handle the event customer.tax_id.updated
  case "identity.verification_session.canceled":
    // Then define and call a function to handle the event identity.verification_session.canceled
  case "identity.verification_session.created":
    // Then define and call a function to handle the event identity.verification_session.created
  case "identity.verification_session.processing":
    // Then define and call a function to handle the event identity.verification_session.processing
  case "identity.verification_session.redacted":
    // Then define and call a function to handle the event identity.verification_session.redacted
  case "identity.verification_session.requires_input":
    // Then define and call a function to handle the event identity.verification_session.requires_input
  case "identity.verification_session.verified":
    // Then define and call a function to handle the event identity.verification_session.verified
  case "invoice.created":
    // Then define and call a function to handle the event invoice.created
  case "invoice.deleted":
    // Then define and call a function to handle the event invoice.deleted
  case "invoice.finalization_failed":
    // Then define and call a function to handle the event invoice.finalization_failed
  case "invoice.finalized":
    // Then define and call a function to handle the event invoice.finalized
  case "invoice.marked_uncollectible":
    // Then define and call a function to handle the event invoice.marked_uncollectible
  case "invoice.overdue":
    // Then define and call a function to handle the event invoice.overdue
  case "invoice.overpaid":
    // Then define and call a function to handle the event invoice.overpaid
  case "invoice.paid":
    // Then define and call a function to handle the event invoice.paid
  case "invoice.payment_action_required":
    // Then define and call a function to handle the event invoice.payment_action_required
  case "invoice.payment_failed":
    // Then define and call a function to handle the event invoice.payment_failed
  case "invoice.payment_succeeded":
    // Then define and call a function to handle the event invoice.payment_succeeded
  case "invoice.sent":
    // Then define and call a function to handle the event invoice.sent
  case "invoice.upcoming":
    // Then define and call a function to handle the event invoice.upcoming
  case "invoice.updated":
    // Then define and call a function to handle the event invoice.updated
  case "invoice.voided":
    // Then define and call a function to handle the event invoice.voided
  case "invoice.will_be_due":
    // Then define and call a function to handle the event invoice.will_be_due
  case "payment_intent.amount_capturable_updated":
    // Then define and call a function to handle the event payment_intent.amount_capturable_updated
  case "payment_intent.canceled":
    // Then define and call a function to handle the event payment_intent.canceled
  case "payment_intent.created":
    // Then define and call a function to handle the event payment_intent.created
  case "payment_intent.partially_funded":
    // Then define and call a function to handle the event payment_intent.partially_funded
  case "payment_intent.payment_failed":
    // Then define and call a function to handle the event payment_intent.payment_failed
  case "payment_intent.processing":
    // Then define and call a function to handle the event payment_intent.processing
  case "payment_intent.requires_action":
    // Then define and call a function to handle the event payment_intent.requires_action
  case "payment_intent.succeeded":
    // Then define and call a function to handle the event payment_intent.succeeded
  case "payment_method.attached":
    // Then define and call a function to handle the event payment_method.attached
  case "payment_method.automatically_updated":
    // Then define and call a function to handle the event payment_method.automatically_updated
  case "payment_method.detached":
    // Then define and call a function to handle the event payment_method.detached
  case "payment_method.updated":
    // Then define and call a function to handle the event payment_method.updated
  case "plan.created":
    // Then define and call a function to handle the event plan.created
  case "plan.deleted":
    // Then define and call a function to handle the event plan.deleted
  case "plan.updated":
    // Then define and call a function to handle the event plan.updated
  case "price.created":
    // Then define and call a function to handle the event price.created
  case "price.deleted":
    // Then define and call a function to handle the event price.deleted
  case "price.updated":
    // Then define and call a function to handle the event price.updated
  case "product.created":
    // Then define and call a function to handle the event product.created
  case "product.deleted":
    // Then define and call a function to handle the event product.deleted
  case "product.updated":
    // Then define and call a function to handle the event product.updated
  case "refund.created":
    // Then define and call a function to handle the event refund.created
  case "refund.failed":
    // Then define and call a function to handle the event refund.failed
  case "refund.updated":
    // Then define and call a function to handle the event refund.updated
  case "subscription_schedule.aborted":
    // Then define and call a function to handle the event subscription_schedule.aborted
  case "subscription_schedule.canceled":
    // Then define and call a function to handle the event subscription_schedule.canceled
  case "subscription_schedule.completed":
    // Then define and call a function to handle the event subscription_schedule.completed
  case "subscription_schedule.created":
    // Then define and call a function to handle the event subscription_schedule.created
  case "subscription_schedule.expiring":
    // Then define and call a function to handle the event subscription_schedule.expiring
  case "subscription_schedule.released":
    // Then define and call a function to handle the event subscription_schedule.released
  case "subscription_schedule.updated":
    // Then define and call a function to handle the event subscription_schedule.updated
  case "tax.settings.updated":
    // Then define and call a function to handle the event tax.settings.updated
  case "tax_rate.created":
    // Then define and call a function to handle the event tax_rate.created
  case "tax_rate.updated":
    // Then define and call a function to handle the event tax_rate.updated
  // ... handle other event types
  default:
      fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
  }

  w.WriteHeader(http.StatusOK)
}