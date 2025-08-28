package events

import (
	"testing"

	"github.com/smartcontractkit/cvn-api-go/client"
)

type eventListFile struct {
	Data    []client.Event `json:"data"`
	HasMore bool           `json:"has_more"`
}

func TestLoadMockEventList(t *testing.T) {
	var l eventListFile
	if err := LoadJson("event_list.json", &l); err != nil {
		t.Fatalf("LoadJson(event_list.json): %v", err)
	}
	if got := len(l.Data); got != 3 {
		t.Fatalf("expected 3 events in list, got %d", got)
	}
}

func TestLoadMockEvent_Valid(t *testing.T) {
	ev, err := LoadMockEvent("valid_event.json")
	if err != nil {
		t.Fatalf("LoadMockEvent(valid_event.json): %v", err)
	}
	if ev.Service != "dvp" {
		t.Fatalf("expected service 'dvp', got %q", ev.Service)
	}
	if ev.Name == "" {
		t.Fatalf("event name should not be empty")
	}
}
