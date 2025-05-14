// internal/services/subscription_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/dukerupert/walking-drum/internal/domain/dto"
	"github.com/dukerupert/walking-drum/internal/domain/models"
	"github.com/dukerupert/walking-drum/internal/messaging/messages"
	"github.com/dukerupert/walking-drum/internal/messaging/publishers"
	"github.com/dukerupert/walking-drum/internal/repositories/interfaces"
	"github.com/dukerupert/walking-drum/internal/services/stripe"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SubscriptionService defines the interface for subscription business logic
type SubscriptionService interface {
	Create(ctx context.Context, subscription *models.Subscription) (*models.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetByStripeItemID(ctx context.Context, stripeItemID string) (*models.Subscription, error)
	GetWithRelatedData(ctx context.Context, id uuid.UUID) (*dto.SubscriptionDetailDTO, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Subscription, int, error)
	ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error)
	ListActiveByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error)
	ListByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.Subscription, int, error)
	ListDueForRenewal(ctx context.Context, before time.Time) ([]*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, subscriptionDTO *dto.SubscriptionUpdateDTO) (*models.Subscription, error)
	Cancel(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error
	Pause(ctx context.Context, id uuid.UUID) error
	Resume(ctx context.Context, id uuid.UUID) error
	ChangeProduct(ctx context.Context, id uuid.UUID, productID uuid.UUID) error
	ChangePrice(ctx context.Context, id uuid.UUID, priceID uuid.UUID) error
	ChangeQuantity(ctx context.Context, id uuid.UUID, quantity int) error
	ChangeAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) error
	ProcessRenewal(ctx context.Context, id uuid.UUID) error
}

// subscriptionService implements the SubscriptionService interface
type subscriptionService struct {
	subscriptionRepo interfaces.SubscriptionRepository
	customerRepo     interfaces.CustomerRepository
	productRepo      interfaces.ProductRepository
	priceRepo        interfaces.PriceRepository
	stripeClient     stripe.StripeService
	publisher        *publishers.SubscriptionPublisher
	logger           zerolog.Logger
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	subscriptionRepo interfaces.SubscriptionRepository,
	customerRepo interfaces.CustomerRepository,
	productRepo interfaces.ProductRepository,
	priceRepo interfaces.PriceRepository,
	stripeClient stripe.StripeService,
	publisher *publishers.SubscriptionPublisher,
	logger *zerolog.Logger,
) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		customerRepo:     customerRepo,
		productRepo:      productRepo,
		priceRepo:        priceRepo,
		stripeClient:     stripeClient,
		publisher:        publisher,
		logger:           logger.With().Str("component", "subscription_service").Logger(),
	}
}

// Create adds a new subscription to the system
func (s *subscriptionService) Create(ctx context.Context, subscription *models.Subscription) (*models.Subscription, error) {
	s.logger.Debug().
		Str("function", "subscriptionService.Create").
		Str("customer_id", subscription.CustomerID.String()).
		Str("product_id", subscription.ProductID.String()).
		Str("price_id", subscription.PriceID.String()).
		Str("stripe_id", subscription.StripeID).
		Str("stripe_item_id", subscription.StripeItemID).
		Int("quantity", subscription.Quantity).
		Msg("Starting subscription creation")

	// Validate the subscription
	if subscription.ID == uuid.Nil {
		subscription.ID = uuid.New()
	}

	if subscription.CustomerID == uuid.Nil {
		return nil, fmt.Errorf("customer ID is required")
	}

	if subscription.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product ID is required")
	}

	if subscription.PriceID == uuid.Nil {
		return nil, fmt.Errorf("price ID is required")
	}

	if subscription.StripeID == "" {
		return nil, fmt.Errorf("stripe subscription ID is required")
	}

	if subscription.Quantity <= 0 {
		subscription.Quantity = 1 // Default to 1 if not specified
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, subscription.CustomerID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "subscriptionService.Create").
			Str("customer_id", subscription.CustomerID.String()).
			Msg("Error retrieving customer")
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	if customer == nil {
		s.logger.Error().
			Str("function", "subscriptionService.Create").
			Str("customer_id", subscription.CustomerID.String()).
			Msg("Customer not found")
		return nil, fmt.Errorf("customer not found")
	}

	// Verify product exists
	product, err := s.productRepo.GetByID(ctx, subscription.ProductID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "subscriptionService.Create").
			Str("product_id", subscription.ProductID.String()).
			Msg("Error retrieving product")
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	if product == nil {
		s.logger.Error().
			Str("function", "subscriptionService.Create").
			Str("product_id", subscription.ProductID.String()).
			Msg("Product not found")
		return nil, fmt.Errorf("product not found")
	}

	// Verify price exists
	price, err := s.priceRepo.GetByID(ctx, subscription.PriceID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "subscriptionService.Create").
			Str("price_id", subscription.PriceID.String()).
			Msg("Error retrieving price")
		return nil, fmt.Errorf("failed to retrieve price: %w", err)
	}

	if price == nil {
		s.logger.Error().
			Str("function", "subscriptionService.Create").
			Str("price_id", subscription.PriceID.String()).
			Msg("Price not found")
		return nil, fmt.Errorf("price not found")
	}

	// Check product stock level (if applicable)
	if product.StockLevel >= 0 && subscription.Quantity > product.StockLevel {
		s.logger.Error().
			Str("function", "subscriptionService.Create").
			Str("product_id", subscription.ProductID.String()).
			Int("requested_quantity", subscription.Quantity).
			Int("available_stock", product.StockLevel).
			Msg("Insufficient stock")
		return nil, fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
			product.Name, subscription.Quantity, product.StockLevel)
	}

	// Set timestamps if not already set
	if subscription.CreatedAt.IsZero() {
		subscription.CreatedAt = time.Now()
	}
	if subscription.UpdatedAt.IsZero() {
		subscription.UpdatedAt = time.Now()
	}

	// Set default status if not set
	if subscription.Status == "" {
		subscription.Status = models.SubscriptionStatusActive
	}

	// Create subscription in database
	err = s.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "subscriptionService.Create").
			Str("subscription_id", subscription.ID.String()).
			Msg("Error creating subscription in database")
		return nil, fmt.Errorf("failed to create subscription in database: %w", err)
	}

	// Update product stock if applicable
	if product.StockLevel > 0 {
		updatedStock := product.StockLevel - subscription.Quantity
		err = s.productRepo.UpdateStockLevel(ctx, subscription.ProductID, updatedStock)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("function", "subscriptionService.Create").
				Str("product_id", subscription.ProductID.String()).
				Int("new_stock", updatedStock).
				Msg("Failed to update product stock, but subscription was created")
			// Don't fail the subscription creation if stock update fails
			// Just log a warning
		}
	}

	s.logger.Info().
		Str("function", "subscriptionService.Create").
		Str("subscription_id", subscription.ID.String()).
		Str("customer_id", subscription.CustomerID.String()).
		Str("product_id", subscription.ProductID.String()).
		Str("stripe_id", subscription.StripeID).
		Str("status", subscription.Status).
		Msg("Subscription created successfully")

	return subscription, nil
}

// GetByID retrieves a subscription by its ID
func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	// TODO: Implement get subscription by ID
	// 1. Call repository to fetch subscription
	return nil, nil
}

// GetByStripeItemID retrieves a subscription by its Stripe item ID
func (s *subscriptionService) GetByStripeItemID(ctx context.Context, stripeItemID string) (*models.Subscription, error) {
	// Add debug logging
	s.logger.Debug().
		Str("function", "subscriptionService.GetByStripeItemID").
		Str("stripe_item_id", stripeItemID).
		Msg("Starting subscription retrieval by Stripe item ID")

	// Validate input
	if stripeItemID == "" {
		s.logger.Error().
			Str("function", "subscriptionService.GetByStripeItemID").
			Msg("Stripe item ID is empty")
		return nil, fmt.Errorf("stripe item ID cannot be empty")
	}

	// Call repository to fetch subscription
	subscription, err := s.subscriptionRepo.GetByStripeID(ctx, stripeItemID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "subscriptionService.GetByStripeItemID").
			Str("stripe_item_id", stripeItemID).
			Msg("Error retrieving subscription from repository")
		return nil, fmt.Errorf("failed to retrieve subscription by Stripe item ID: %w", err)
	}

	// Check if subscription was found
	if subscription == nil {
		s.logger.Warn().
			Str("function", "subscriptionService.GetByStripeItemID").
			Str("stripe_item_id", stripeItemID).
			Msg("Subscription not found with the given Stripe item ID")
		return nil, nil
	}

	// Log success
	s.logger.Info().
		Str("function", "subscriptionService.GetByStripeItemID").
		Str("subscription_id", subscription.ID.String()).
		Str("customer_id", subscription.CustomerID.String()).
		Str("product_id", subscription.ProductID.String()).
		Str("stripe_id", subscription.StripeID).
		Str("stripe_item_id", stripeItemID).
		Msg("Subscription successfully retrieved by Stripe item ID")

	return subscription, nil
}

// GetWithRelatedData retrieves a subscription with its related product, price, customer and address data
func (s *subscriptionService) GetWithRelatedData(ctx context.Context, id uuid.UUID) (*dto.SubscriptionDetailDTO, error) {
	// TODO: Implement get subscription with related data
	// 1. Call repository to fetch subscription with related data
	// 2. Map to DTO
	return nil, nil
}

// List retrieves all subscriptions with optional pagination
func (s *subscriptionService) List(ctx context.Context, page, pageSize int) ([]*models.Subscription, int, error) {
	// TODO: Implement subscription listing
	// 1. Calculate offset from page and pageSize
	// 2. Call repository to list subscriptions
	return nil, 0, nil
}

// ListByCustomerID retrieves all subscriptions for a specific customer
func (s *subscriptionService) ListByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error) {
	// TODO: Implement subscription listing by customer ID
	// 1. Call repository to list subscriptions by customer ID
	return nil, nil
}

// ListActiveByCustomerID retrieves active subscriptions for a specific customer
func (s *subscriptionService) ListActiveByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*models.Subscription, error) {
	// TODO: Implement active subscription listing by customer ID
	// 1. Call repository to list active subscriptions by customer ID
	return nil, nil
}

// ListByStatus retrieves subscriptions filtered by status
func (s *subscriptionService) ListByStatus(ctx context.Context, status string, page, pageSize int) ([]*models.Subscription, int, error) {
	// TODO: Implement subscription listing by status
	// 1. Validate status
	// 2. Calculate offset from page and pageSize
	// 3. Call repository to list subscriptions by status
	return nil, 0, nil
}

// ListDueForRenewal retrieves subscriptions due for renewal in the given time period
func (s *subscriptionService) ListDueForRenewal(ctx context.Context, before time.Time) ([]*models.Subscription, error) {
	// TODO: Implement listing subscriptions due for renewal
	// 1. Call repository to list subscriptions due for renewal
	return nil, nil
}

// Update updates an existing subscription
func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, subscriptionDTO *dto.SubscriptionUpdateDTO) (*models.Subscription, error) {
	// TODO: Implement subscription update
	// 1. Get existing subscription
	// 2. Update fields from DTO
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil, nil
}

// Pause pauses a subscription
func (s *subscriptionService) Pause(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement subscription pausing
	// 1. Get existing subscription
	// 2. Pause in Stripe
	// 3. Update status in database
	// 4. Handle errors
	return nil
}

// Resume resumes a paused subscription
func (s *subscriptionService) Resume(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement subscription resuming
	// 1. Get existing subscription
	// 2. Resume in Stripe
	// 3. Update status in database
	// 4. Handle errors
	return nil
}

// ChangeProduct changes the product of a subscription
func (s *subscriptionService) ChangeProduct(ctx context.Context, id uuid.UUID, productID uuid.UUID) error {
	// TODO: Implement changing subscription product
	// 1. Get existing subscription
	// 2. Check if product exists and has stock
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil
}

// ChangePrice changes the price of a subscription
func (s *subscriptionService) ChangePrice(ctx context.Context, id uuid.UUID, priceID uuid.UUID) error {
	// TODO: Implement changing subscription price
	// 1. Get existing subscription
	// 2. Check if price exists and belongs to the product
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil
}

// ChangeQuantity changes the quantity of a subscription
func (s *subscriptionService) ChangeQuantity(ctx context.Context, id uuid.UUID, quantity int) error {
	// TODO: Implement changing subscription quantity
	// 1. Get existing subscription
	// 2. Check if product has enough stock
	// 3. Update in Stripe
	// 4. Update in database
	// 5. Handle errors
	return nil
}

// ChangeAddress changes the shipping address of a subscription
func (s *subscriptionService) ChangeAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) error {
	// TODO: Implement changing subscription address
	// 1. Get existing subscription
	// 2. Check if address exists and belongs to the customer
	// 3. Update in database
	// 4. Handle errors
	return nil
}

// QueueRenewal queues a subscription for renewal processing
func (s *subscriptionService) QueueRenewal(ctx context.Context, subscriptionID uuid.UUID) error {
    subscription, err := s.subscriptionRepo.GetByID(ctx, subscriptionID)
    if err != nil {
        s.logger.Error().
            Err(err).
            Str("function", "subscriptionService.QueueRenewal").
            Str("subscription_id", subscriptionID.String()).
            Msg("Failed to get subscription for renewal")
        return fmt.Errorf("failed to get subscription: %w", err)
    }

    if subscription == nil {
        s.logger.Error().
            Str("function", "subscriptionService.QueueRenewal").
            Str("subscription_id", subscriptionID.String()).
            Msg("Subscription not found for renewal")
        return fmt.Errorf("subscription not found: %s", subscriptionID)
    }

    // Create a renewal message
    renewalMsg := messages.SubscriptionRenewalMessage{
        SubscriptionID: subscription.ID,
        CustomerID:     subscription.CustomerID,
        ProductID:      subscription.ProductID,
        PriceID:        subscription.PriceID,
        Quantity:       subscription.Quantity,
        RenewalDate:    time.Now(),
    }

    // Publish the renewal message
    err = s.publisher.PublishRenewal(ctx, renewalMsg)
    if err != nil {
        s.logger.Error().
            Err(err).
            Str("function", "subscriptionService.QueueRenewal").
            Str("subscription_id", subscriptionID.String()).
            Msg("Failed to publish renewal message")
        return fmt.Errorf("failed to queue renewal: %w", err)
    }

    s.logger.Info().
        Str("function", "subscriptionService.QueueRenewal").
        Str("subscription_id", subscriptionID.String()).
        Msg("Subscription renewal queued successfully")

    return nil
}

// ProcessRenewal processes a subscription renewal
// This is now called by the consumer when processing a renewal message
func (s *subscriptionService) ProcessRenewal(ctx context.Context, id uuid.UUID) error {
    // Get the subscription from the database
    subscription, err := s.subscriptionRepo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error().
            Err(err).
            Str("function", "subscriptionService.ProcessRenewal").
            Str("subscription_id", id.String()).
            Msg("Failed to retrieve subscription")
        return fmt.Errorf("failed to get subscription: %w", err)
    }

    if subscription == nil {
        s.logger.Error().
            Str("function", "subscriptionService.ProcessRenewal").
            Str("subscription_id", id.String()).
            Msg("Subscription not found")
        return fmt.Errorf("subscription not found: %s", id)
    }
    
    // After successful renewal, publish email notification
    if err := s.sendRenewalEmail(ctx, subscription); err != nil {
        // Log the error but continue, as the renewal itself was successful
        s.logger.Error().
            Err(err).
            Str("function", "subscriptionService.ProcessRenewal").
            Str("subscription_id", id.String()).
            Msg("Failed to send renewal email")
    }
    
    // After successful renewal, update product stock levels
    if err := s.updateProductStock(ctx, subscription); err != nil {
        // Log the error but continue, as the renewal itself was successful
        s.logger.Error().
            Err(err).
            Str("function", "subscriptionService.ProcessRenewal").
            Str("subscription_id", id.String()).
            Msg("Failed to update product stock levels")
    }
    
    return nil
}

// sendRenewalEmail sends a renewal notification email
func (s *subscriptionService) sendRenewalEmail(ctx context.Context, subscription *models.Subscription) error {
    // Get customer details
    customer, err := s.customerRepo.GetByID(ctx, subscription.CustomerID)
    if err != nil {
        return fmt.Errorf("failed to get customer: %w", err)
    }

    // Get product details
    product, err := s.productRepo.GetByID(ctx, subscription.ProductID)
    if err != nil {
        return fmt.Errorf("failed to get product: %w", err)
    }

    // Create email notification message
    emailMsg := messages.EmailNotificationMessage{
        Type:       "renewal",
        CustomerID: customer.ID,
        Email:      customer.Email,
        Subject:    "Your subscription has been renewed",
        Data: map[string]interface{}{
            "customer_name":       customer.FirstName + " " + customer.LastName,
            "subscription_id":     subscription.ID.String(),
            "product_name":        product.Name,
            "renewal_date":        subscription.CurrentPeriodStart.Format("2006-01-02"),
            "next_delivery_date":  subscription.NextDeliveryDate.Format("2006-01-02"),
            "next_billing_date":   subscription.CurrentPeriodEnd.Format("2006-01-02"),
        },
    }

    // Publish the email notification
    return s.publisher.PublishEmailNotification(ctx, emailMsg)
}

// updateProductStock updates the product stock levels after a renewal
func (s *subscriptionService) updateProductStock(ctx context.Context, subscription *models.Subscription) error {
    // Create stock update message
    stockMsg := messages.StockUpdateMessage{
        ProductID: subscription.ProductID,
        Quantity:  subscription.Quantity,
        Operation: "decrement",
    }

    // Publish the stock update message
    return s.publisher.PublishStockUpdate(ctx, stockMsg)
}

// Cancel cancels a subscription
func (s *subscriptionService) Cancel(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error {
    // Get existing subscription
    subscription, err := s.subscriptionRepo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get subscription: %w", err)
    }

    if subscription == nil {
        return fmt.Errorf("subscription not found: %s", id)
    }

    oldStatus := subscription.Status

    // Existing cancellation logic...

    // Publish status change event
    statusChangeMsg := messages.SubscriptionStatusChangeMessage{
        SubscriptionID: subscription.ID,
        CustomerID:     subscription.CustomerID,
        OldStatus:      oldStatus,
        NewStatus:      subscription.Status,
        ChangeDate:     time.Now(),
    }

    if err := s.publisher.PublishStatusChange(ctx, statusChangeMsg); err != nil {
        s.logger.Error().
            Err(err).
            Str("subscription_id", id.String()).
            Msg("Failed to publish status change event, but subscription was cancelled")
    }

    // Send cancellation email
    if err := s.sendCancellationEmail(ctx, subscription); err != nil {
        s.logger.Error().
            Err(err).
            Str("subscription_id", id.String()).
            Msg("Failed to send cancellation email, but subscription was cancelled")
    }

    return nil
}

// sendCancellationEmail sends a cancellation notification email
func (s *subscriptionService) sendCancellationEmail(ctx context.Context, subscription *models.Subscription) error {
    // Similar to sendRenewalEmail but for cancellation notifications
    // ...
    return nil
}
