package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// HandlerFunc is a function that handles an event.
type HandlerFunc func(ctx context.Context, request *Event) (any, error)

// Service is a service that can be orchestrated.
type Service struct {
	// The name of the service.
	Name string
	// Optional dependencies for the service. If nil, the service will not depend on any other services.
	Dependencies []string
	// Optional retry configuration for the service. If nil, the service will not be retried.
	RetryConfig *RetryConfig
	// The handler for the service.
	Handler HandlerFunc
}

// RetryConfig is a configuration for retrying a service
type RetryConfig struct {
	MaxRetries    int              // The maximum number of retries for the service.
	InitialDelay  time.Duration    // The initial delay between retries.
	MaxRetryDelay time.Duration    // The maximum delay between retries.
	JitterFactor  float64          // The jitter factor for the retries.
	ShouldRetry   func(error) bool // A function that determines if the service should retry.
}

// DefaultRetryConfig is a reasonable default retry configuration for a service
var DefaultRetryConfig = &RetryConfig{
	MaxRetries:    5,
	InitialDelay:  100 * time.Millisecond,
	MaxRetryDelay: 3 * time.Second,
	JitterFactor:  0.2,
	ShouldRetry: func(err error) bool {
		return err != nil
	},
}

// newRetryHandler wraps the provided service handler in a retry handler and returns a new handler.
func newRetryHandler(service Service) HandlerFunc {
	handler := service.Handler

	return func(ctx context.Context, request *Event) (any, error) {
		result, err := handler(ctx, request)

		for attempt := range service.RetryConfig.MaxRetries {
			if err != nil && service.RetryConfig.ShouldRetry(err) {
				baseDelay := math.Min(
					float64(service.RetryConfig.InitialDelay)*math.Pow(2, float64(attempt)),
					float64(service.RetryConfig.MaxRetryDelay),
				)
				delayWithJitter := time.Duration(
					baseDelay * (1 + (rand.Float64()*2-1)*service.RetryConfig.JitterFactor),
				)
				time.Sleep(delayWithJitter)
				result, err = handler(ctx, request)
			} else {
				return result, nil
			}
		}

		return nil, errors.Join(
			err,
			fmt.Errorf(
				"service %s failed after %d attempts",
				service.Name,
				service.RetryConfig.MaxRetries,
			),
		)
	}
}
