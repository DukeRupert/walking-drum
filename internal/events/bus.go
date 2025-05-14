// internal/events/bus.go
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

// EventBus provides methods for publishing and subscribing to events
type EventBus interface {
	// Publish sends an event to the specified topic
	Publish(topic string, payload interface{}) error

	// Subscribe registers a handler for events on the specified topic
	Subscribe(topic string, handler func([]byte)) (*nats.Subscription, error)

	// Close closes the connection to the message bus
	Close()
}

// NATSEventBus implements EventBus using NATS
type NATSEventBus struct {
	conn      *nats.Conn
	jetStream nats.JetStreamContext
	logger    zerolog.Logger
}

// Event represents a message in the event bus
type Event struct {
	ID        string      `json:"id"`
	Topic     string      `json:"topic"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// NewNATSEventBus creates a new NATS-based event bus
func NewNATSEventBus(natsURL string, logger *zerolog.Logger) (*NATSEventBus, error) {
	subLogger := logger.With().Str("component", "nats_event_bus").Logger()

	// Connect to NATS
	subLogger.Info().Str("url", natsURL).Msg("Connecting to NATS")
	nc, err := nats.Connect(natsURL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			subLogger.Warn().Err(err).Msg("Disconnected from NATS")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			subLogger.Info().Str("url", nc.ConnectedUrl()).Msg("Reconnected to NATS")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			subLogger.Error().Err(err).Msg("NATS error")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	subLogger.Info().Msg("Successfully connected to NATS")

	return &NATSEventBus{
		conn:      nc,
		jetStream: js,
		logger:    subLogger,
	}, nil
}

// Publish sends an event to the specified topic
func (n *NATSEventBus) Publish(topic string, payload interface{}) error {
	// Create an event with metadata
	event := Event{
		ID:        uuid.New().String(),
		Topic:     topic,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	// Marshal the event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	n.logger.Debug().
		Str("topic", topic).
		Str("event_id", event.ID).
		Int("data_size", len(data)).
		Msg("Publishing event")

	// Publish the event
	return n.conn.Publish(topic, data)
}

// PublishPersistent publishes an event that will be stored in JetStream
func (n *NATSEventBus) PublishPersistent(topic string, payload interface{}) error {
    // Create an event with metadata
    event := Event{
        ID:        uuid.New().String(),
        Topic:     topic,
        Timestamp: time.Now(),
        Payload:   payload,
    }
    
    // Marshal the event to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    n.logger.Debug().
        Str("topic", topic).
        Str("event_id", event.ID).
        Int("data_size", len(data)).
        Msg("Publishing persistent event")
    
    // Create stream if needed
    streamName := "EVENTS"
    stream, err := n.jetStream.StreamInfo(streamName)
    if err != nil || stream == nil {
        n.logger.Info().Msg("Creating events stream")
        _, err = n.jetStream.AddStream(&nats.StreamConfig{
            Name:     streamName,
            Subjects: []string{"events.>"},
            Storage:  nats.FileStorage,
            MaxAge:   time.Hour * 24 * 30, // 30 days retention
        })
        if err != nil {
            return fmt.Errorf("failed to create stream: %w", err)
        }
    }
    
    // Publish the event with acknowledgment
    _, err = n.jetStream.Publish("events."+topic, data)
    return err
}

// Subscribe registers a handler for events on the specified topic
func (n *NATSEventBus) Subscribe(topic string, handler func([]byte)) (*nats.Subscription, error) {
	n.logger.Debug().
		Str("topic", topic).
		Msg("Subscribing to topic")

	// Subscribe to the topic
	return n.conn.Subscribe(topic, func(msg *nats.Msg) {
		n.logger.Debug().
			Str("topic", topic).
			Int("data_size", len(msg.Data)).
			Msg("Received message")

		// Call the handler with the message data
		handler(msg.Data)
	})
}

// Close closes the connection to NATS
func (n *NATSEventBus) Close() {
	if n.conn != nil {
		n.logger.Debug().Msg("Closing NATS connection")
		n.conn.Close()
	}
}
