package dta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/smartcontractkit/cvn-sdk/client"
)

// VerifiableEvent represents an event structure that encapsulates data about the event, its metadata, and associated blockchain transaction details.
type VerifiableEvent struct {
	CreatedAt time.Time `json:"createdAt"`
	Event     struct {
		Type      string `json:"type"`
		Name      string `json:"name"`
		Address   string `json:"address"`
		RequestId string `json:"requestId"`
		TopicHash string `json:"topicHash"`
	} `json:"event"`
	Metadata struct {
		ChainId       string `json:"chainId"`
		Network       string `json:"network"`
		WorkflowEvent struct {
			Attributes      map[string]Attribute `json:"attributes"`
			BusinessEventId string               `json:"business_event_id"`
			Component       string               `json:"component"`
			EventTimestamp  time.Time            `json:"event_timestamp"`
			EventTypeLabel  string               `json:"event_type_label"`
			Failed          bool                 `json:"failed"`
			FinalEvent      bool                 `json:"final_event"`
			Id              string               `json:"id"`
			Participant     string               `json:"participant"`
			ParticipantRole string               `json:"participant_role"`
			ProcessLabels   []string             `json:"process_labels"`
			RawData         string               `json:"raw_data"`
			Title           string               `json:"title"`
		} `json:"workflowEvent"`
	} `json:"metadata"`
	Transaction struct {
		Timestamp int    `json:"timestamp"`
		ChainId   string `json:"chainId"`
		Hash      string `json:"hash"`
	} `json:"transaction"`
}

type Attribute struct {
	Key        string `json:"key"`
	OnChain    bool   `json:"on_chain"`
	Value      string `json:"value"`
	Visibility string `json:"visibility"`
}

func Decode(ctx context.Context, event client.Event) (VerifiableEvent, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		return VerifiableEvent{}, err
	}

	var verifiableEvent VerifiableEvent
	if err = json.Unmarshal(decodedBytes, &verifiableEvent); err != nil {
		return VerifiableEvent{}, err
	}

	return verifiableEvent, nil
}

// EventName is a method that returns the event name associated with this VerifiableEvent
func (v VerifiableEvent) EventName() EventName {
	eventType, ok := v.Metadata.WorkflowEvent.Attributes["event_type"]
	if !ok {
		return ""
	}
	eventName, b := parseEvent(eventType.Value)
	if !b {
		return ""
	}
	return eventName
}
