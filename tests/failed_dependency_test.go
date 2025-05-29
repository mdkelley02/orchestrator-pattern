package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/mdkelley02/orchestrator-pattern/orchestrator"
	"github.com/stretchr/testify/require"
)

func Test_FailedDependency(t *testing.T) {
	authorizationService := orchestrator.Service{
		Name: "authorization",
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return nil, errors.New("authorization failed")
		},
	}

	paymentService := orchestrator.Service{
		Name:         "payment",
		Dependencies: []string{authorizationService.Name},
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			return "PAYMENT", nil
		},
	}

	o := orchestrator.New(
		discardLogger,
		nil,
		authorizationService,
		paymentService,
	)

	ctx := context.Background()

	event, err := o.Orchestrate(ctx, newRequest())
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, []string{"authorization failed"}, event.Warnings)
	require.Equal(t, map[string]any{paymentService.Name: "PAYMENT"}, event.Responses)
}
