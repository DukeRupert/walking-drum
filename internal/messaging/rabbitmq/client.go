// internal/messaging/rabbitmq/client.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client represents a RabbitMQ client
type Client struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	logger     zerolog.Logger
	config     Config
	publishers map[string]Publisher
	consumers  map[string]Consumer
}

// Config represents RabbitMQ configuration
type Config struct {
	URL               string
	ReconnectInterval time.Duration
	ExchangeName      string
}

// NewClient creates a new RabbitMQ client
func NewClient(config Config, logger *zerolog.Logger) (*Client, error) {
	client := &Client{
		config:     config,
		logger:     logger.With().Str("component", "rabbitmq").Logger(),
		publishers: make(map[string]Publisher),
		consumers:  make(map[string]Consumer),
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	// Set up reconnect handling
	go client.handleReconnect()

	return client, nil
}

// connect establishes a connection to RabbitMQ
func (c *Client) connect() error {
	c.logger.Info().Msg("Connecting to RabbitMQ")
	
	conn, err := amqp.Dial(c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}
	c.channel = ch

	// Declare the topic exchange
	err = ch.ExchangeDeclare(
		c.config.ExchangeName, // name
		"topic",               // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	c.logger.Info().Msg("Successfully connected to RabbitMQ")
	return nil
}

// handleReconnect continuously monitors the connection and reconnects if needed
func (c *Client) handleReconnect() {
	for {
		// Wait for connection to close
		reason, ok := <-c.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			// Channel closed normally
			break
		}

		c.logger.Error().
			Err(reason).
			Msg("RabbitMQ connection closed unexpectedly")

		// Clear existing resources
		if c.channel != nil {
			c.channel.Close()
		}

		// Reconnect with exponential backoff
		backoff := c.config.ReconnectInterval
		for {
			c.logger.Info().
				Dur("backoff", backoff).
				Msg("Attempting to reconnect to RabbitMQ")

			if err := c.connect(); err != nil {
				c.logger.Error().
					Err(err).
					Msg("Failed to reconnect to RabbitMQ")

				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
				continue
			}

			// Recreate all publishers and consumers
			c.recreatePublishers()
			c.recreateConsumers()
			break
		}
	}
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// recreatePublishers recreates all registered publishers after reconnection
func (c *Client) recreatePublishers() {
	// Implementation depends on how publishers are registered
}

// recreateConsumers recreates all registered consumers after reconnection
func (c *Client) recreateConsumers() {
	// Implementation depends on how consumers are registered
}

// Publish publishes a message to RabbitMQ
func (c *Client) Publish(ctx context.Context, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		c.config.ExchangeName, // exchange
		routingKey,            // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	c.logger.Debug().
		Str("routing_key", routingKey).
		Int("body_size", len(body)).
		Msg("Message published successfully")

	return nil
}

// Consumer represents a message consumer
type Consumer interface {
	Start() error
	Stop() error
}

// Publisher represents a message publisher
type Publisher interface {
	Publish(ctx context.Context, message interface{}) error
}