package reader

import (
	"context"
	"log"
	"testing"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/internal/mockserver"
)

func TestEventReader(t *testing.T) {
	ctx := context.Background()

	mockServer := mockserver.NewMockServer(t)
	defer mockServer.Close()

	c, err := client.NewClientWithResponses(mockServer.TestServer.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	r, err := NewEventReader(
		c, &EventReaderOptions{
			EventsAfter: 0,
		},
	)
	if err != nil {
		t.Fatalf("failed to create event reader: %v", err)
	}

	eventList1, err := r.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	log.Printf("Events 1 received: %v", eventList1)

	eventList2, err := r.Read(ctx)
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	log.Printf("Events 2 received: %v", eventList2)
}
