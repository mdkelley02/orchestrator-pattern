package orchestrator

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Orchestrator struct {
	log          *slog.Logger
	journaler    Journaler
	services     map[string]Service
	serviceCount int
}

func New(
	log *slog.Logger,
	journaler Journaler,
	services ...Service,
) *Orchestrator {
	if len(services) == 0 {
		panic("no services provided")
	}

	orchestrator := &Orchestrator{
		log:          log,
		journaler:    journaler,
		services:     make(map[string]Service),
		serviceCount: len(services),
	}

	for _, serviceConfig := range services {
		if serviceConfig.RetryConfig != nil {
			serviceConfig.Handler = newRetryHandler(serviceConfig)
		}
		orchestrator.services[serviceConfig.Name] = serviceConfig
	}

	return orchestrator
}

func (o *Orchestrator) Orchestrate(
	ctx context.Context,
	request *Request,
) (_ *Event, err error) {
	// Create the event
	event := &Event{
		EventId:   request.RequestId,
		CreatedAt: time.Now().Format(time.RFC3339),
		Request:   request,
		Warnings:  []string{},
		Responses: map[string]any{},
	}

	// Initialize channels for each service
	serviceChannels := make(map[string]chan struct{})
	for name := range o.services {
		serviceChannels[name] = make(chan struct{})
	}

	// Initialize mutex and wait group
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(o.serviceCount)

	// Start all services
	for _, service := range o.services {
		go func(service Service) {
			defer wg.Done()

			// Wait for dependencies using channels
			for _, dependency := range o.services[service.Name].Dependencies {
				<-serviceChannels[dependency]
			}

			// Execute service
			result, err := service.Handler(ctx, event)

			// Add result or error to event
			mu.Lock()
			if err == nil {
				event.Responses[service.Name] = result
			} else {
				o.log.Error("service failed", "error", err)
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
