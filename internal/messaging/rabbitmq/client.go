// internal/messaging/rabbitmq/client.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client represents a RabbitMQ client
type Client struct {
	conn            *amqp.Connection
	connLock        sync.Mutex
	config          Config
	logger          zerolog.Logger
	reconnectNotify chan struct{}
	isConnected     bool
	done            chan struct{}
	
	// Channel pool related fields
	channelPool     []*amqp.Channel
	channelLock     sync.Mutex
	maxChannels     int
	
	// Consumer related fields
	consumers       map[string]Consumer
	consumersLock   sync.Mutex
}

// Config represents RabbitMQ configuration
type Config struct {
	URL               string
	ReconnectInterval time.Duration
	ExchangeName      string
	MaxChannels       int
}

// NewClient creates a new RabbitMQ client
func NewClient(config Config, logger *zerolog.Logger) (*Client, error) {
	if config.MaxChannels <= 0 {
		config.MaxChannels = 10 // Default value
	}
	
	client := &Client{
		config:          config,
		logger:          logger.With().Str("component", "rabbitmq").Logger(),
		reconnectNotify: make(chan struct{}),
		isConnected:     false,
		done:            make(chan struct{}),
		maxChannels:     config.MaxChannels,
		channelPool:     make([]*amqp.Channel, 0, config.MaxChannels),
		consumers:       make(map[string]Consumer),
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
	c.connLock.Lock()
	defer c.connLock.Unlock()

	c.logger.Info().Msg("Connecting to RabbitMQ")
	
	// Close existing connection if it exists
	if c.conn != nil {
		c.conn.Close()
	}
	
	// Clear the channel pool
	c.channelLock.Lock()
	c.channelPool = make([]*amqp.Channel, 0, c.maxChannels)
	c.channelLock.Unlock()
	
	// Connect with heartbeat and locale settings
	conn, err := amqp.DialConfig(c.config.URL, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Properties: amqp.Table{
			"connection_name": "coffee-subscription-service",
		},
	})
	
	if err != nil {
		c.isConnected = false
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	c.conn = conn
	c.isConnected = true

	// Initialize the first channel and declare exchange
	ch, err := c.createChannel()
	if err != nil {
		c.isConnected = false
		conn.Close()
		return fmt.Errorf("failed to create initial channel: %w", err)
	}

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
		c.isConnected = false
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Add the channel to the pool
	c.channelLock.Lock()
	c.channelPool = append(c.channelPool, ch)
	c.channelLock.Unlock()

	c.logger.Info().Msg("Successfully connected to RabbitMQ")
	
	// Signal reconnect if this was a reconnection
	if c.reconnectNotify != nil {
		select {
		case c.reconnectNotify <- struct{}{}:
			// Notified successfully
		default:
			// Channel full or closed, create a new one
			c.reconnectNotify = make(chan struct{}, 1)
			c.reconnectNotify <- struct{}{}
		}
	}
	
	return nil
}

// handleReconnect continuously monitors the connection and reconnects if needed
func (c *Client) handleReconnect() {
	for {
		select {
		case <-c.done:
			return
		case err := <-c.conn.NotifyClose(make(chan *amqp.Error)):
			if err == nil {
				// Normal closure, no error
				c.logger.Info().Msg("RabbitMQ connection closed gracefully")
				return
			}

			c.logger.Error().
				Err(err).
				Msg("RabbitMQ connection closed unexpectedly")

			c.isConnected = false

			// Reconnect with exponential backoff
			backoff := c.config.ReconnectInterval
			maxBackoff := 1 * time.Minute
			
			for {
				select {
				case <-c.done:
					return
				default:
					c.logger.Info().
						Dur("backoff", backoff).
						Msg("Attempting to reconnect to RabbitMQ")

					if err := c.connect(); err != nil {
						c.logger.Error().
							Err(err).
							Msg("Failed to reconnect to RabbitMQ")

						time.Sleep(backoff)
						backoff *= 2 // Exponential backoff
						if backoff > maxBackoff {
							backoff = maxBackoff
						}
						continue
					}

					// Successfully reconnected, restart consumers
					c.restartConsumers()
					break
				}
				
				// Exit the retry loop once connected
				if c.isConnected {
					break
				}
			}
		}
	}
}

// restartConsumers restarts all registered consumers
func (c *Client) restartConsumers() {
	c.consumersLock.Lock()
	defer c.consumersLock.Unlock()

	for name, consumer := range c.consumers {
		c.logger.Info().
			Str("consumer", name).
			Msg("Restarting consumer after reconnection")
			
		if err := consumer.Start(); err != nil {
			c.logger.Error().
				Err(err).
				Str("consumer", name).
				Msg("Failed to restart consumer")
		}
	}
}

// Channel returns a channel from the pool or creates a new one
func (c *Client) Channel() *amqp.Channel {
	c.channelLock.Lock()
	defer c.channelLock.Unlock()

	// Return an existing channel if available
	if len(c.channelPool) > 0 {
		ch := c.channelPool[0]
		c.channelPool = c.channelPool[1:]
		return ch
	}

	// Create a new channel if the pool is empty
	ch, err := c.createChannel()
	if err != nil {
		c.logger.Error().
			Err(err).
			Msg("Failed to create channel")
		return nil
	}

	return ch
}

// createChannel creates a new AMQP channel
func (c *Client) createChannel() (*amqp.Channel, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Set QoS/prefetch for better load balancing
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return ch, nil
}

// ReturnChannel returns a channel to the pool
func (c *Client) ReturnChannel(ch *amqp.Channel) {
	if ch == nil {
		return
	}

	c.channelLock.Lock()
	defer c.channelLock.Unlock()

	// Only return to pool if we're under the limit
	if len(c.channelPool) < c.maxChannels {
		c.channelPool = append(c.channelPool, ch)
	} else {
		ch.Close()
	}
}

// Connection returns the underlying AMQP connection
func (c *Client) Connection() *amqp.Connection {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	return c.conn
}

// Config returns the client's configuration
func (c *Client) Config() Config {
	return c.config
}

// RegisterConsumer registers a consumer with the client
func (c *Client) RegisterConsumer(name string, consumer Consumer) {
	c.consumersLock.Lock()
	defer c.consumersLock.Unlock()
	
	c.consumers[name] = consumer
	c.logger.Info().
		Str("consumer", name).
		Msg("Registered consumer")
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	close(c.done)
	
	c.connLock.Lock()
	defer c.connLock.Unlock()
	
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Publish publishes a message to RabbitMQ
func (c *Client) Publish(ctx context.Context, routingKey string, message interface{}) error {
	if !c.isConnected {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Get a channel for publishing
	ch := c.Channel()
	if ch == nil {
		return fmt.Errorf("failed to get channel for publishing")
	}
	defer c.ReturnChannel(ch)

	err = ch.PublishWithContext(
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