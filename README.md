# Orchestrator Pattern

A Go implementation of the Orchestrator pattern for managing complex service workflows with dependency management, retry mechanisms, and event journaling.

## Overview

This project provides a robust framework for orchestrating multiple services in a distributed system. It handles service dependencies, implements retry mechanisms with exponential backoff and jitter, and provides event journaling capabilities.

## Features

- **Service Orchestration**: Manage multiple services with defined dependencies
- **Retry Mechanism**: Configurable retry policies with exponential backoff and jitter
- **Event Journaling**: Track and log service execution events
- **Concurrent Execution**: Services run concurrently while respecting dependencies
- **Error Handling**: Comprehensive error handling and propagation
- **Structured Logging**: Built-in support for structured logging using `slog`

## Installation

```bash
go get github.com/mdkelley02/orchestrator-pattern
```

## Usage

Here's a basic example of how to use the orchestrator:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/mdkelley02/orchestrator-pattern/internal/event"
    "github.com/mdkelley02/orchestrator-pattern/internal/journaler"
    "github.com/mdkelley02/orchestrator-pattern/internal/orchestrator"
    "github.com/mdkelley02/orchestrator-pattern/internal/services"
)

func main() {
    // Create services
    paymentService := services.Config{
        Name:        "payments",
        RetryConfig: services.DefaultRetryConfig,
        Handler: func(ctx context.Context, request *event.Event) (any, error) {
            // Implement payment service logic
            return "PAYMENT_SUCCESS", nil
        },
    }

    analyticsService := services.Config{
        Name:         "analytics",
        Dependencies: []string{paymentService.Name},
        Handler: func(ctx context.Context, request *event.Event) (any, error) {
            // Implement analytics service logic
            return "ANALYTICS_COMPLETE", nil
        },
    }

    // Create orchestrator
    o := orchestrator.New(
        slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })),
        journaler.New(),
        paymentService,
        analyticsService,
    )

    // Create request
    request := &orchestrator.Request{
        EventId:        "123",
        TenantId:       "tenant-1",
        OrganizationId: "org-1",
        Payload: map[string]any{
            "amount": 100.00,
            "currency": "USD",
        },
    }

    // Execute orchestration
    response, err := o.Orchestrate(context.Background(), request)
    if err != nil {
        // Handle error
    }
}
```

## Service Configuration

### Retry Configuration

The orchestrator supports configurable retry policies:

```go
retryConfig := &services.RetryConfig{
    MaxRetries:    5,
    InitialDelay:  100 * time.Millisecond,
    MaxRetryDelay: 3 * time.Second,
    JitterFactor:  0.2,
}
```

### Service Dependencies

Services can declare dependencies on other services:

```go
service := services.Config{
    Name:         "service-name",
    Dependencies: []string{"dependency-1", "dependency-2"},
    Handler: func(ctx context.Context, request *event.Event) (any, error) {
        // Service implementation
        return result, nil
    },
}
```

## Event Journaling

The orchestrator automatically journals events, including:

- Service execution results
- Warnings and errors
- Execution timestamps
- Request and response data

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
