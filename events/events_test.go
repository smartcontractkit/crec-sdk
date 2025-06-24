package events

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/internal/mockdata"
	"github.com/smartcontractkit/cvn-sdk/internal/mockserver"
)

func TestReadEvent(t *testing.T) {
	ctx := context.Background()

	mockServer := mockserver.NewMockServer(t)
	defer mockServer.Close()

	c, err := client.NewClientWithResponses(mockServer.TestServer.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	r, err := NewClient(
		c, &ClientOptions{
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

func TestVerifyEvent(t *testing.T) {
	event, err := mockdata.LoadMockEvent("valid_event.json")
	require.NoError(t, err)

	mockServer := mockserver.NewMockServer(t)
	defer mockServer.Close()

	c, err := client.NewClientWithResponses(mockServer.TestServer.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	v, err := NewClient(
		c, &ClientOptions{
			ValidSigners: []string{
				"0xFD528f7bd7a6eB8d6605BF944d122da7665C69A1",
				"0x4f4ddd274635D014C4584118a0fdD6cf89B25d3b",
			},
			MinRequiredSignatures: 2,
		},
	)
	require.NoError(t, err)

	verified, err := v.Verify(event)
	require.NoError(t, err)
	require.True(t, verified)
}
