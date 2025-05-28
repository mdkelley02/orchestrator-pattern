package example

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
	"github.com/mdkelley02/orchestrator-pattern/internal/journaler"
	"github.com/mdkelley02/orchestrator-pattern/internal/orchestrator"
	"github.com/mdkelley02/orchestrator-pattern/internal/services"
	"github.com/stretchr/testify/require"
)

func Request() *orchestrator.Request {
	return &orchestrator.Request{
		EventId:        strconv.Itoa(rand.Intn(1000000)),
		TenantId:       strconv.Itoa(rand.Intn(1000000)),
		OrganizationId: strconv.Itoa(rand.Intn(1000000)),
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

func TestLightGraphOrchestration(t *testing.T) {
	var PaymentService = services.Config{
		Name:        "payments",
		RetryConfig: services.DefaultRetryConfig,
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "PAYMENTS", nil
		},
	}

	var AnalyticsService = services.Config{
		Name:         "analytics",
		Dependencies: []string{PaymentService.Name},
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "ANALYTICS", nil
		},
	}

	var ReportingService = services.Config{
		Name:         "reporting",
		Dependencies: []string{AnalyticsService.Name},
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "REPORTING", nil
		},
	}

	o := orchestrator.New(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		journaler.New(),
		PaymentService,
		AnalyticsService,
		ReportingService,
	)

	ctx := context.Background()

	for range 1000 {
		response, err := o.Orchestrate(ctx, Request())
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, response.Event.Responses[PaymentService.Name], "PAYMENTS")
		require.Equal(t, response.Event.Responses[AnalyticsService.Name], "ANALYTICS")
		require.Equal(t, response.Event.Responses[ReportingService.Name], "REPORTING")
	}
}

func TestFlatOrchestration(t *testing.T) {
	var PaymentService = services.Config{
		Name: "payments",
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "PAYMENTS", nil
		},
	}

	var AnalyticsService = services.Config{
		Name: "analytics",
		Handler: func(ctx context.Context, request *event.Event) (any, error) {
			return "ANALYTICS", nil
		},
	}

	o := orchestrator.New(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		journaler.New(),
		PaymentService,
		AnalyticsService,
	)

	ctx := context.Background()

	for range 1000 {
		response, err := o.Orchestrate(ctx, Request())
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, response.Event.Responses[PaymentService.Name], "PAYMENTS")
		require.Equal(t, response.Event.Responses[AnalyticsService.Name], "ANALYTICS")
	}
}
