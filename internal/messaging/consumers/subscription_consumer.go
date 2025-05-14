// internal/messaging/consumers/subscription_consumer.go
package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dukerupert/walking-drum/internal/messaging/messages"
	"github.com/dukerupert/walking-drum/internal/messaging/rabbitmq"
	"github.com/dukerupert/walking-drum/internal/services"
	"github.com/rs/zerolog"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SubscriptionRenewalConsumer processes subscription renewal events
type SubscriptionRenewalConsumer struct {
	client             *rabbitmq.Client
	logger             zerolog.Logger
	subscriptionSvc    services.SubscriptionService
	productSvc         services.ProductService
	queue              string
	consumerTag        string
	done               chan struct{}
}

// NewSubscriptionRenewalConsumer creates a new subscription renewal consumer
func NewSubscriptionRenewalConsumer(
	client *rabbitmq.Client,
	logger *zerolog.Logger,
	subscriptionSvc services.SubscriptionService,
	productSvc services.ProductService,
) *SubscriptionRenewalConsumer {
	return &SubscriptionRenewalConsumer{
		client:          client,
		logger:          logger.With().Str("component", "subscription_renewal_consumer").Logger(),
		subscriptionSvc: subscriptionSvc,
		productSvc:      productSvc,
		queue:           "subscription_renewals",
		consumerTag:     "subscription_renewal_consumer",
		done:            make(chan struct{}),
	}
}

// Start starts the consumer
func (c *SubscriptionRenewalConsumer) Start() error {
	ch := c.client.Channel()
	
	// Declare a queue for this consumer
	q, err := ch.QueueDeclare(
		c.queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,                     // queue name
		"subscription.renewal",     // routing key
		c.client.Config().ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name,        // queue
		c.consumerTag, // consumer tag
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		c.logger.Info().
			Str("queue", q.Name).
			Str("consumer_tag", c.consumerTag).
			Msg("Subscription renewal consumer started")

		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Warn().Msg("Subscription renewal channel closed")
					return
				}
				c.processMessage(msg)
			}
		}
	}()

	return nil
}

// Stop stops the consumer
func (c *SubscriptionRenewalConsumer) Stop() error {
	c.logger.Info().Msg("Stopping subscription renewal consumer")
	close(c.done)
	return nil
}

// processMessage processes a subscription renewal message
func (c *SubscriptionRenewalConsumer) processMessage(msg amqp.Delivery) {
	var message messages.SubscriptionRenewalMessage
	
	// Create a context for processing
	ctx := context.Background()
	
	c.logger.Debug().
		RawJSON("body", msg.Body).
		Msg("Processing subscription renewal message")
	
	// Unmarshal the message
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		c.logger.Error().
			Err(err).
			RawJSON("body", msg.Body).
			Msg("Failed to unmarshal subscription renewal message")
		msg.Nack(false, false) // Don't requeue malformed messages
		return
	}
	
	// Process the renewal
	err := c.subscriptionSvc.ProcessRenewal(ctx, message.SubscriptionID)
	if err != nil {
		c.logger.Error().
			Err(err).
			Str("subscription_id", message.SubscriptionID.String()).
			Msg("Failed to process subscription renewal")
		
		// Requeue the message for retry, unless it's a permanent error
		if isPermanentError(err) {
			c.logger.Warn().
				Err(err).
				Str("subscription_id", message.SubscriptionID.String()).
				Msg("Permanent error - not requeuing message")
			msg.Nack(false, false)
		} else {
			c.logger.Info().
				Err(err).
				Str("subscription_id", message.SubscriptionID.String()).
				Msg("Temporary error - requeuing message")
			msg.Nack(false, true)
		}
		return
	}
	
	c.logger.Info().
		Str("subscription_id", message.SubscriptionID.String()).
		Msg("Successfully processed subscription renewal")
	
	// Acknowledge the message
	msg.Ack(false)
}

// isPermanentError determines if an error is permanent (should not be retried)
func isPermanentError(err error) bool {
	// Implement logic to determine if an error is permanent
	// For example, validation errors, not found errors, etc.
	return false
}