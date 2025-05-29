# Orchestrator Pattern

A robust and flexible Go library for orchestrating complex service workflows with dependency management, retry mechanisms, and event journaling.

## Features

- **Dependency Management**: Define service dependencies and execute them in the correct order
- **Concurrent Execution**: Services run concurrently when possible, improving performance
- **Retry Mechanism**: Built-in support for configurable retry policies with exponential backoff and jitter
- **Event Journaling**: Optional event logging for audit trails and debugging
- **Error Handling**: Graceful error handling with service-level error isolation
- **Context Support**: Full context.Context integration for cancellation and timeouts

## Installation

```bash
go get github.com/mdkelley02/orchestrator-pattern
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/mdkelley02/orchestrator-pattern/orchestrator"
)

func main() {
    // Create a logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Define services
    authService := orchestrator.Service{
        Name: "authorization",
        Handler: func(ctx context.Context, event *orchestrator.Event) (any, error) {
            // Implement authorization logic
            return "AUTHORIZED", nil
        },
    }

    paymentService := orchestrator.Service{
        Name:         "payment",
        Dependencies: []string{authService.Name},
        RetryConfig:  orchestrator.DefaultRetryConfig,
        Handler: func(ctx context.Context, event *orchestrator.Event) (any, error) {
            // Implement payment logic
            return "PAID", nil
        },
    }

    // Create orchestrator
    o := orchestrator.New(logger, nil, authService, paymentService)

    // Create request
    request := &orchestrator.Request{
        RequestId: "123",
        Payload:   map[string]any{"amount": 100},
    }

    // Execute workflow
    event, err := o.Orchestrate(context.Background(), request)
    if err != nil {
        logger.Error("orchestration failed", "error", err)
        return
    }

    // Process results
    logger.Info("orchestration completed", "responses", event.Responses)
}
```

## Key Concepts

### Service

A service represents a single unit of work in your workflow. Each service has:

- A unique name
- Optional dependencies on other services
- A handler function that implements the service logic
- Optional retry configuration

### Retry Configuration

The library provides a flexible retry mechanism with:

- Configurable maximum retry attempts
- Exponential backoff
- Jitter to prevent thundering herd problems
- Custom retry conditions

```go
retryConfig := &orchestrator.RetryConfig{
    MaxRetries:    5,
    InitialDelay:  100 * time.Millisecond,
    MaxRetryDelay: 3 * time.Second,
    JitterFactor:  0.2,
    ShouldRetry: func(err error) bool {
        return err != nil
    },
}
```

### Event Journaling

Implement the `Journaler` interface to log events for auditing or debugging:

```go
type MyJournaler struct{}

func (j *MyJournaler) Journal(ctx context.Context, event *orchestrator.Event) error {
    // Implement your journaling logic
    return nil
}
```

## Best Practices

1. **Service Isolation**: Keep services focused on a single responsibility
2. **Error Handling**: Use the retry mechanism for transient failures
3. **Dependency Management**: Keep the dependency graph as simple as possible
4. **Context Usage**: Always respect context cancellation in service handlers
5. **Logging**: Use the provided logger for consistent logging across services

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
