package client

type Event struct {
	// Context The context of the event, can contain any JSON data.
	Context []ContextElement `json:"context"`

	// Message The message associated with the event (events with the same message are grouped together)
	Message string `json:"message"`

	// Source The source of the event (GitHub APP, CI, etc..)
	Source string `json:"source"`

	// Tags A map of tags associated with the event.
	Tags map[string]string `json:"tags"`
}

// ContextElement defines model for ContextElement.
type ContextElement struct {
	Data interface{} `json:"data"`
	Type string      `json:"type"`
}
