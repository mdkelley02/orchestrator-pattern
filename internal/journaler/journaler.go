package journaler

import (
	"context"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
)

type Journaler interface {
	Journal(ctx context.Context, event *event.Event) error
}
