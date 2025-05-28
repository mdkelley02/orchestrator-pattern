package journaler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mdkelley02/orchestrator-pattern/internal/event"
)

type Journaler interface {
	Journal(ctx context.Context, event *event.Event) error
}

type journaler struct{}

func New() Journaler {
	return &journaler{}
}

func (j *journaler) Journal(ctx context.Context, event *event.Event) error {
	json, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}
