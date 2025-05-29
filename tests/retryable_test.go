package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mdkelley02/orchestrator-pattern/orchestrator"
	"github.com/stretchr/testify/require"
)

func Test_Retryable(t *testing.T) {
	retryableError := errors.New("retryable error")
	retryConfig := &orchestrator.RetryConfig{
		MaxRetries:    orchestrator.DefaultRetryConfig.MaxRetries,
		InitialDelay:  time.Millisecond,
		MaxRetryDelay: time.Millisecond * 10,
		JitterFactor:  orchestrator.DefaultRetryConfig.JitterFactor,
		ShouldRetry: func(err error) bool {
			return err == retryableError
		},
	}

	i := 0
	authorizationService := orchestrator.Service{
		Name:        "authorization",
		RetryConfig: retryConfig,
		Handler: func(ctx context.Context, request *orchestrator.Event) (any, error) {
			i++
			if i < orchestrator.DefaultRetryConfig.MaxRetries {
				return nil, retryableError
			}
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
	require.Equal(t, map[string]any{
		authorizationService.Name: "AUTHORIZATION",
		paymentService.Name:       "PAYMENT",
	}, event.Responses)
}
