package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
	"github.com/mdkelley02/orchestrator-pattern/internal/journaler"
	"github.com/mdkelley02/orchestrator-pattern/internal/services"
)

type Orchestrator struct {
	log          *slog.Logger
	journaler    journaler.Journaler
	serviceMap   map[string]services.Config
	serviceCount int
}

func New(
	log *slog.Logger,
	journaler journaler.Journaler,
	s ...services.Config,
) *Orchestrator {
	if len(s) == 0 {
		panic("no services provided")
	}

	o := &Orchestrator{
		log:          log,
		journaler:    journaler,
		serviceMap:   make(map[string]services.Config),
		serviceCount: len(s),
	}

	for _, service := range s {
		if service.RetryConfig != nil {
			handler := service.Handler
			service.Handler = func(ctx context.Context, request *event.Event) (any, error) {
				result, err := handler(ctx, request)
				for attempt := range service.RetryConfig.MaxRetries {
					if err != nil {
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

		o.serviceMap[service.Name] = service
	}

	return o
}

type Request struct {
	RequestId string
	Payload   any
}

func (o *Orchestrator) Orchestrate(
	ctx context.Context,
	request *Request,
) (_ *event.Event, err error) {
	// Create the event
	event := &event.Event{
		EventId:   request.RequestId,
		CreatedAt: time.Now().Format(time.RFC3339),
		Request:   request,
		Warnings:  []string{},
		Responses: map[string]any{},
	}

	// Initialize channels for each service
	serviceChannels := make(map[string]chan struct{})
	for name := range o.serviceMap {
		serviceChannels[name] = make(chan struct{})
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(o.serviceCount)

	// Start all services
	for _, service := range o.serviceMap {
		go func(service services.Config) {
			defer wg.Done()

			// Wait for dependencies using channels
			for _, dependency := range o.serviceMap[service.Name].Dependencies {
				<-serviceChannels[dependency]
			}

			// Execute service and add result or error to event
			result, err := service.Handler(ctx, event)
			mu.Lock()
			if err == nil {
				event.Responses[service.Name] = result
			} else {
				event.Warnings = append(event.Warnings, err.Error())
			}
			mu.Unlock()

			// Signal completion to dependent services
			close(serviceChannels[service.Name])
		}(service)
	}

	// Wait for all services to complete
	wg.Wait()

	// Journal the event
	if o.journaler != nil {
		go func() {
			if err := o.journaler.Journal(ctx, event); err != nil {
				o.log.Error("failed to journal event", "error", err)
			}
		}()
	}

	// Return the response
	return event, nil
}
