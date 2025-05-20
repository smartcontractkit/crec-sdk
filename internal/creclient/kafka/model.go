package kafka

type ResourceLogs struct {
	Resource  Resource   `json:"resource"`
	ScopeLogs []ScopeLog `json:"scopeLogs"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value Value  `json:"value"`
}

type Value struct {
	StringValue string `json:"stringValue,omitempty"`
	BytesValue  string `json:"bytesValue,omitempty"`
}

type ScopeLog struct {
	Scope      Scope       `json:"scope"`
	LogRecords []LogRecord `json:"logRecords"`
}

type Scope struct {
	Name string `json:"name"`
}

type LogRecord struct {
	ObservedTimeUnixNano string      `json:"observedTimeUnixNano"`
	Body                 Value       `json:"body"`
	Attributes           []Attribute `json:"attributes"`
}

type Body struct {
	ResourceLogs []ResourceLogs `json:"resourceLogs"`
}
type CREEvent struct {
	ID                string         `json:"id" mapstructure:"id"`
	BusinessEventID   string         `json:"business_event_id" mapstructure:"business_event_id"`
	TriggerEventId    string         `json:"trigger_event_id" mapstructure:"trigger_event_id"`
	EventTypeLabel    string         `json:"event_type_label" mapstructure:"event_type_label"`
	EventLabel        string         `json:"event_label" mapstructure:"event_label"`
	EventTimestamp    string         `json:"event_timestamp" mapstructure:"event_timestamp"`
	Participant       string         `json:"participant" mapstructure:"participant"`
	ParticipantRole   string         `json:"participant_role" mapstructure:"participant_role"`
	Model             string         `json:"model" mapstructure:"model"`
	Component         string         `json:"component" mapstructure:"component"`
	Title             string         `json:"title" mapstructure:"title"`
	ProcessLabels     []string       `json:"process_labels" mapstructure:"process_labels"`
	RawData           string         `json:"raw_data" mapstructure:"raw_data"`
	Attributes        map[string]any `json:"attributes" mapstructure:"attributes"`
	OnChainAttributes map[string]any `json:"on_chain_attributes" mapstructure:"on_chain_attributes"`
	Failed            bool           `json:"failed" mapstructure:"failed"`
	FinalEvent        bool           `json:"final_event" mapstructure:"final_event"`
}
