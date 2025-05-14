// internal/messaging/publishers/subscription_publisher.go
package publishers

import (
	"context"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/messaging/messages"
	"github.com/dukerupert/walking-drum/internal/messaging/rabbitmq"
	"github.com/rs/zerolog"
)

// SubscriptionPublisher publishes subscription-related events
type SubscriptionPublisher struct {
	client *rabbitmq.Client
	logger zerolog.Logger
}

// NewSubscriptionPublisher creates a new subscription publisher
func NewSubscriptionPublisher(client *rabbitmq.Client, logger *zerolog.Logger) *SubscriptionPublisher {
	return &SubscriptionPublisher{
		client: client,
		logger: logger.With().Str("component", "subscription_publisher").Logger(),
	}
}

// PublishRenewal publishes a subscription renewal event
func (p *SubscriptionPublisher) PublishRenewal(ctx context.Context, message messages.SubscriptionRenewalMessage) error {
	routingKey := "subscription.renewal"
	p.logger.Debug().
		Str("subscription_id", message.SubscriptionID.String()).
		Str("routing_key", routingKey).
		Msg("Publishing subscription renewal event")

	return p.client.Publish(ctx, routingKey, message)
}

// PublishStatusChange publishes a subscription status change event
func (p *SubscriptionPublisher) PublishStatusChange(ctx context.Context, message messages.SubscriptionStatusChangeMessage) error {
	routingKey := fmt.Sprintf("subscription.status.%s", message.NewStatus)
	p.logger.Debug().
		Str("subscription_id", message.SubscriptionID.String()).
		Str("old_status", message.OldStatus).
		Str("new_status", message.NewStatus).
		Str("routing_key", routingKey).
		Msg("Publishing subscription status change event")

	return p.client.Publish(ctx, routingKey, message)
}

// PublishEmailNotification publishes an email notification event
func (p *SubscriptionPublisher) PublishEmailNotification(ctx context.Context, message messages.EmailNotificationMessage) error {
	routingKey := fmt.Sprintf("email.%s", message.Type)
	p.logger.Debug().
		Str("customer_id", message.CustomerID.String()).
		Str("email_type", message.Type).
		Str("routing_key", routingKey).
		Msg("Publishing email notification event")

	return p.client.Publish(ctx, routingKey, message)
}

// PublishStockUpdate publishes a stock update event
func (p *SubscriptionPublisher) PublishStockUpdate(ctx context.Context, message messages.StockUpdateMessage) error {
	routingKey := fmt.Sprintf("stock.%s", message.Operation)
	p.logger.Debug().
		Str("product_id", message.ProductID.String()).
		Int("quantity", message.Quantity).
		Str("operation", message.Operation).
		Str("routing_key", routingKey).
		Msg("Publishing stock update event")

	return p.client.Publish(ctx, routingKey, message)
}