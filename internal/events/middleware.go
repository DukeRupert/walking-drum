// internal/events/middleware.go
package events

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// HandlerFunc is a function that handles an event
type HandlerFunc func(ctx context.Context, data []byte)

// Middleware wraps a HandlerFunc
type Middleware func(HandlerFunc) HandlerFunc

// WithLogging logs information about the event processing
func WithLogging(logger zerolog.Logger) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx context.Context, data []byte) {
            start := time.Now()
            
            topic, _ := ctx.Value("topic").(string)
            eventID, _ := ctx.Value("event_id").(string)
            
            logger.Debug().
                Str("topic", topic).
                Str("event_id", eventID).
                Msg("Processing event")
            
            // Call the next handler
            next(ctx, data)
            
            logger.Debug().
                Str("topic", topic).
                Str("event_id", eventID).
                Dur("duration", time.Since(start)).
                Msg("Event processed")
        }
    }
}

// WithRetry retries the handler if it panics
func WithRetry(maxRetries int) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx context.Context, data []byte) {
            var err error
            
            for attempt := 0; attempt <= maxRetries; attempt++ {
                func() {
                    defer func() {
                        if r := recover(); r != nil {
                            err = fmt.Errorf("panic in event handler: %v", r)
                        }
                    }()
                    
                    // If this is a retry, wait using exponential backoff
                    if attempt > 0 {
                        backoff := time.Duration(1<<uint(attempt-1)) * time.Second
                        time.Sleep(backoff)
                    }
                    
                    // Try to handle the event
                    next(ctx, data)
                    
                    // If we get here, we succeeded
                    err = nil
                }()
                
                // If no error, we're done
                if err == nil {
                    break
                }
            }
        }
    }
}