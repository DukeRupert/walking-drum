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
	productService := NewProductService(repos.Product, stripeService, logger)
	variantService := NewVariantService(repos.Variant, repos.Product, repos.Price, stripeService, logger)
	priceService := NewPriceService(repos.Price, repos.Product, stripeService, logger)
	customerService := NewCustomerService(repos.Customer, stripeService, logger)
	subscriptionService := NewSubscriptionService(
		repos.Subscription,
		repos.Customer,
		repos.Product,
		repos.Price,
		stripeService,
		logger,
	)

	return &Services{
		Product:      productService,
		Variant:      variantService,
		Price:        priceService,
		Customer:     customerService,
		Subscription: subscriptionService,
		Stripe:       stripeService,
	}
}
