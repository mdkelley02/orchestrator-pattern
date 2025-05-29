package orchestrator

import (
	"context"
)

// Journaler is a service that can be used to journal events.
type Journaler interface {
	Journal(ctx context.Context, event *Event) error
}
