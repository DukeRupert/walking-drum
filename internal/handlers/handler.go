package handlers

import (
	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Product      *ProductHandler
	Variant      *VariantHandler
	Price        *PriceHandler
	Customer     *CustomerHandler
	Checkout     *CheckoutHandler
	Subscription *SubscriptionHandler
	Webhook      *WebhookHandler
}

func CreateHandlers(cfg *config.Config, s *services.Services, l *zerolog.Logger) *Handlers {
	return &Handlers{
		Product:      NewProductHandler(s.Product, l),
		Variant:      NewVariantHandler(s.Variant, s.Product, l),
		Price:        NewPriceHandler(s.Price, l),
		Customer:     NewCustomerHandler(s.Customer, l),
		Checkout:     NewCheckoutHandler(s.Stripe, s.Product, s.Price, s.Customer, s.Subscription, l),
		Subscription: NewSubscriptionHandler(s.Subscription, l),
		Webhook:      NewWebhookHandler(s.Stripe, cfg.Stripe.WebhookSecret, l),
	}
}
