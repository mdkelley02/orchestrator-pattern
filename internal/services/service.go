package services

import (
	"context"
	"time"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
)

// RetryConfig is a configuration for retrying a service
type RetryConfig struct {
	MaxRetries    int           // The maximum number of retries for the service.
	InitialDelay  time.Duration // The initial delay between retries.
	MaxRetryDelay time.Duration // The maximum delay between retries.
	JitterFactor  float64       // The jitter factor for the retries.
}

// DefaultRetryConfig is a reasonable default retry configuration for a service
var DefaultRetryConfig = &RetryConfig{
	MaxRetries:    5,
	InitialDelay:  100 * time.Millisecond,
	MaxRetryDelay: 3 * time.Second,
	JitterFactor:  0.2,
}

type Config struct {
	// The name of the service.
	Name string
	// Optional dependencies for the service. If nil, the service will not depend on any other services.
	Dependencies []string
	// Optional retry configuration for the service. If nil, the service will not be retried.
	RetryConfig *RetryConfig
	// The handler for the service.
	Handler func(ctx context.Context, request *event.Event) (any, error)
}
