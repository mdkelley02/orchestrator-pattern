package orchestrator

// Request is a struct that represents a request to the orchestrator.
type Request struct {
	RequestId string `json:"requestId"`
	Payload   any    `json:"payload"`
}

// Event is a struct that represents an event in the orchestrator.
type Event struct {
	EventId   string         `json:"eventId"`
	CreatedAt string         `json:"createdAt"`
	Request   any            `json:"request"`
	Warnings  []string       `json:"warnings"`
	Responses map[string]any `json:"responses"`
}
