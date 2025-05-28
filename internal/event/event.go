package event

type Event struct {
	EventId   string         `json:"eventId"`
	CreatedAt string         `json:"createdAt"`
	Request   any            `json:"request"`
	Warnings  []string       `json:"warnings"`
	Responses map[string]any `json:"responses"`
}
