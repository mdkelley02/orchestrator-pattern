package example

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
	"github.com/mdkelley02/orchestrator-pattern/internal/orchestrator"
	"github.com/mdkelley02/orchestrator-pattern/internal/services"
	"github.com/stretchr/testify/require"
)

func Request() *orchestrator.Request {
	return &orchestrator.Request{
		RequestId: strconv.Itoa(rand.Intn(1000000)),
		Payload: map[string]map[string]any{
			"device": {
				"userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
				"ipAddress": "127.0.0.1",
			},
			"billingInformation": {
				"email": "test@test.com",
				"phone": "1234567890",
				"address": map[string]any{
					"line1": "123 Main St",
					"city":  "Anytown",
					"state": "CA",
					"zip":   "12345",
				},
			},
		},
	}
}

func Test_SimpleDependencyGraph(t *testing.T) {
	paymentService := services.Config{
		Name:        "payments",
		RetryConfig: services.DefaultRetryConfig,
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "PAYMENTS", nil
		},
	}

	analyticsService := services.Config{
		Name:         "analytics",
		Dependencies: []string{paymentService.Name},
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "ANALYTICS", nil
		},
	}

	reportingService := services.Config{
		Name:         "reporting",
		Dependencies: []string{analyticsService.Name},
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "REPORTING", nil
		},
	}

	o := orchestrator.New(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		nil,
		paymentService,
		analyticsService,
		reportingService,
	)

	ctx := context.Background()

	for range 1000 {
		event, err := o.Orchestrate(ctx, Request())
		require.NoError(t, err)
		require.NotNil(t, event)
		require.Equal(t, event.Responses[paymentService.Name], "PAYMENTS")
		require.Equal(t, event.Responses[analyticsService.Name], "ANALYTICS")
		require.Equal(t, event.Responses[reportingService.Name], "REPORTING")
	}
}

func Test_FlatDependencyGraph(t *testing.T) {
	paymentService := services.Config{
		Name: "payments",
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "PAYMENTS", nil
		},
	}

	analyticsService := services.Config{
		Name: "analytics",
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "ANALYTICS", nil
		},
	}

	o := orchestrator.New(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		nil,
		paymentService,
		analyticsService,
	)

	ctx := context.Background()

	for range 1000 {
		event, err := o.Orchestrate(ctx, Request())
		require.NoError(t, err)
		require.NotNil(t, event)
		require.Equal(t, event.Responses[paymentService.Name], "PAYMENTS")
		require.Equal(t, event.Responses[analyticsService.Name], "ANALYTICS")
	}
}

func Test_DependentServiceThrowsError(t *testing.T) {
	authorizationService := services.Config{
		Name: "authorization",
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return nil, errors.New("authorization failed")
		},
	}

	paymentService := services.Config{
		Name:         "payment",
		Dependencies: []string{authorizationService.Name},
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "PAYMENT", nil
		},
	}

	o := orchestrator.New(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		nil,
		authorizationService,
		paymentService,
	)

	ctx := context.Background()

	for range 1000 {
		event, err := o.Orchestrate(ctx, Request())
		require.NoError(t, err)
		require.NotNil(t, event)
		require.Equal(t, []string{"authorization failed"}, event.Warnings)
		require.Equal(t, map[string]any{paymentService.Name: "PAYMENT"}, event.Responses)
	}
}
