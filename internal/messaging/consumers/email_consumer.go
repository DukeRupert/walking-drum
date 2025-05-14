// internal/messaging/consumers/email_consumer.go
package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dukerupert/walking-drum/internal/messaging/messages"
	"github.com/dukerupert/walking-drum/internal/messaging/rabbitmq"
	"github.com/dukerupert/walking-drum/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

// EmailConsumer processes email notification events
type EmailConsumer struct {
	client      *rabbitmq.Client
	logger      zerolog.Logger
	emailSvc    services.EmailService
	queue       string
	consumerTag string
	channel 	*amqp.Channel
	done        chan struct{}
}

// NewEmailConsumer creates a new email consumer
func NewEmailConsumer(
	client *rabbitmq.Client,
	logger *zerolog.Logger,
	emailSvc services.EmailService,
) *EmailConsumer {
	return &EmailConsumer{
		client:      client,
		logger:      logger.With().Str("component", "email_consumer").Logger(),
		emailSvc:    emailSvc,
		queue:       "email_notifications",
		consumerTag: "email_consumer",
		done:        make(chan struct{}),
	}
}

// Start starts the consumer
func (c *EmailConsumer) Start() error {
    // Get a fresh channel for this consumer
    ch := c.client.Channel()
    if ch == nil {
        return fmt.Errorf("failed to get channel for consumer")
    }
    
    // Store the channel for cleanup
    c.channel = ch
    
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
        c.channel = nil
        return fmt.Errorf("failed to declare queue: %w", err)
    }

    // Bind the queue to the exchange with multiple patterns
    bindingPatterns := []string{
        "email.welcome",
        "email.renewal",
        "email.cancelled",
        "email.payment_failed",
        "email.shipment_notification",
    }
    
    for _, pattern := range bindingPatterns {
        err = ch.QueueBind(
            q.Name,                     // queue name
            pattern,                    // routing key
            c.client.Config().ExchangeName, // exchange
            false,
            nil,
        )
        if err != nil {
            c.channel = nil
            return fmt.Errorf("failed to bind queue to %s: %w", pattern, err)
        }
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
        c.channel = nil
        return fmt.Errorf("failed to register consumer: %w", err)
    }

    // Reset the done channel if it was closed
    if c.done == nil {
        c.done = make(chan struct{})
    }
    
    // Register with client for reconnection handling
    c.client.RegisterConsumer(c.consumerTag, c)

    go func() {
        c.logger.Info().
            Str("queue", q.Name).
            Str("consumer_tag", c.consumerTag).
            Msg("Email consumer started")

        for {
            select {
            case <-c.done:
                c.logger.Info().Msg("Consumer stopping: received done signal")
                if c.channel != nil {
                    c.channel.Cancel(c.consumerTag, false)
                }
                return
            case msg, ok := <-msgs:
                if !ok {
                    c.logger.Warn().Msg("Email notification channel closed unexpectedly")
                    // The channel will be reconnected by the client
                    return
                }
                
                // Process the message
                if err := c.processMessage(msg); err != nil {
                    c.logger.Error().
                        Err(err).
                        Msg("Failed to process message")
                        
                    // Handle message failure - either reject or requeue
                    if c.isPermanentError(err) {
                        msg.Reject(false) // Don't requeue
                    } else {
                        msg.Reject(true) // Requeue for retry
                    }
                }
            }
        }
    }()

    return nil
}

// Stop stops the consumer
func (c *EmailConsumer) Stop() error {
	c.logger.Info().Msg("Stopping email consumer")
	close(c.done)
	return nil
}

// processMessage processes an email notification message
func (c *EmailConsumer) processMessage(msg amqp.Delivery) error {
    var message messages.EmailNotificationMessage
    
    // Create a context for processing
    ctx := context.Background()
    
    c.logger.Debug().
        RawJSON("body", msg.Body).
        Msg("Processing email notification message")
    
    // Unmarshal the message
    if err := json.Unmarshal(msg.Body, &message); err != nil {
        c.logger.Error().
            Err(err).
            RawJSON("body", msg.Body).
            Msg("Failed to unmarshal email notification message")
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }
    
    // Send the email
    err := c.emailSvc.SendEmail(ctx, message.Email, message.Subject, message.Type, message.Data)
    if err != nil {
        c.logger.Error().
            Err(err).
            Str("customer_id", message.CustomerID.String()).
            Str("email", message.Email).
            Str("type", message.Type).
            Msg("Failed to send email")
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    c.logger.Info().
        Str("customer_id", message.CustomerID.String()).
        Str("email", message.Email).
        Str("type", message.Type).
        Msg("Successfully sent email")
    
    // Acknowledge the message
    if err := msg.Ack(false); err != nil {
        return fmt.Errorf("failed to acknowledge message: %w", err)
    }
    
    return nil
}

// Add the helper function for determining permanent errors
func (c *EmailConsumer) isPermanentError(err error) bool {
    // Implement logic to determine if an email error is permanent
    // For example, invalid email address, template not found, etc.
    
    if strings.Contains(err.Error(), "invalid email") {
        return true
    }
    
    if strings.Contains(err.Error(), "template not found") {
        return true
    }
    
    // Default to temporary error for most cases, allowing retries
    return false
}