package tests

import (
	"context"
	"testing"

	"github.com/mdkelley02/orchestrator-pattern/orchestrator"
	"github.com/stretchr/testify/require"
)

func Test_SimpleGraph(t *testing.T) {
	authorizationService := orchestrator.Service{
		Name: "authorization",
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return "AUTHORIZATION", nil
		},
	}

	paymentService := orchestrator.Service{
		Name:         "payment",
		Dependencies: []string{authorizationService.Name},
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return "PAYMENT", nil
		},
	}

	analyticsService := orchestrator.Service{
		Name:         "analytics",
		Dependencies: []string{paymentService.Name},
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return "ANALYTICS", nil
		},
	}

	notificationService := orchestrator.Service{
		Name:         "notification",
		Dependencies: []string{analyticsService.Name},
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return "NOTIFICATION", nil
		},
	}

	o := orchestrator.New(
		discardLogger,
		nil,
		authorizationService,
		paymentService,
		analyticsService,
		notificationService,
	)

	ctx := context.Background()

	event, err := o.Orchestrate(ctx, newRequest())
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, map[string]any{
		authorizationService.Name: "AUTHORIZATION",
		paymentService.Name:       "PAYMENT",
		analyticsService.Name:     "ANALYTICS",
		notificationService.Name:  "NOTIFICATION",
	}, event.Responses)
}
