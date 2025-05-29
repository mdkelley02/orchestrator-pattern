package tests

import (
	"io"
	"log/slog"
	"math/rand"
	"strconv"

	"github.com/mdkelley02/orchestrator-pattern/orchestrator"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newRequest() *orchestrator.Request {
	return &orchestrator.Request{
		RequestId: strconv.Itoa(rand.Intn(1000000)),
		Payload: map[string]map[string]any{
			"device": {
				"userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
				"ipAddress": "127.0.0.1",
			},
		},
	}
}
