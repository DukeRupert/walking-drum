package services

import (
	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/dukerupert/walking-drum/internal/repositories/postgres"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/rs/zerolog"
)

type Services struct {
	Product      ProductService
	Variant      VariantService
	Price        PriceService
	Customer     CustomerService
	Subscription SubscriptionService
	Stripe       stripe.StripeService
}

func CreateServices(cfg *config.Config, repos *postgres.Repositories, logger *zerolog.Logger) *Services {
	// Initialize Stripe client
	stripeService := stripe.NewClient(cfg.Stripe.SecretKey, logger)

	return &Services{
		Product:  NewProductService(repos.Product, stripeService, logger),
		Variant:  NewVariantService(repos.Variant, repos.Product, repos.Price, logger),
		Price:    NewPriceService(repos.Price, repos.Product, stripeService, logger),
		Customer: NewCustomerService(repos.Customer, stripeService, logger),
		Subscription: NewSubscriptionService(
			repos.Subscription,
			repos.Customer,
			repos.Product,
			repos.Price,
			stripeService,
			logger,
		),
		Stripe: stripe.NewClient(cfg.Stripe.SecretKey, logger),
	}
}
