// internal/interfaces/events.go
package interfaces

import "github.com/nats-io/nats.go"

// EventBus provides methods for publishing and subscribing to events
type EventBus interface {
	// Publish sends an event to the specified topic
	Publish(topic string, payload interface{}) error
	
	// Subscribe registers a handler for events on the specified topic
	Subscribe(topic string, handler func([]byte)) (*nats.Subscription, error)
	
	// Close closes the connection to the message bus
	Close()
}