package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/dukerupert/walking-drum/internal/events"
	"github.com/rs/zerolog"
)

func main() {
	// Set up logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	
	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	
	logger.Info().Str("url", natsURL).Msg("Connecting to NATS")
	eventBus, err := events.NewNATSEventBus(natsURL, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer eventBus.Close()
	
	// Subscribe to test topic
	_, err = eventBus.Subscribe("test", func(data []byte) {
		logger.Info().Str("data", string(data)).Msg("Received message")
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to subscribe")
	}
	
	// Publish a test message
	err = eventBus.Publish("test", map[string]interface{}{
		"message": "Hello, NATS!",
		"time":    time.Now(),
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to publish message")
	}
	
	logger.Info().Msg("Test message published")
	
	// Wait for signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	
	<-sigCh
	logger.Info().Msg("Shutting down")
}