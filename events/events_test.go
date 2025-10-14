package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/client"
	mockdata "github.com/smartcontractkit/crec-sdk/mocks/events"
	mockserver "github.com/smartcontractkit/crec-sdk/mocks/server"
)

func TestReadEvent(t *testing.T) {
	ctx := context.Background()

	mockServer := mockserver.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := client.NewCREcClient(
		&client.ClientOptions{
			BaseURL: mockServer.TestServer.URL,
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	r, err := NewClient(
		&ClientOptions{
			CREcClient:  c,
			EventsAfter: 0,
		},
	)
	if err != nil {
		t.Fatalf("failed to create event reader: %v", err)
	}

	eventList, err := r.GetEvents(ctx, nil)
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	require.Equal(t, 3, len(eventList)) // must match the number of events in mockdata

	event := (eventList)[0]
	require.Equal(t, "dvp", event.Service)             // must match the service in the first mockdata event
	require.Equal(t, "SettlementAccepted", event.Name) // must match the name in the first mockdata event
}

func TestVerifyEvent(t *testing.T) {
	event, err := mockdata.LoadMockEvent("valid_event.json")
	require.NoError(t, err)

	mockServer := mockserver.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := client.NewCREcClient(
		&client.ClientOptions{
			BaseURL: mockServer.TestServer.URL,
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	v, err := NewClient(
		&ClientOptions{
			CREcClient: c,
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

func TestCreateListener(t *testing.T) {
	ctx := context.Background()

	mockServer := mockserver.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := client.NewCREcClient(
		&client.ClientOptions{
			BaseURL: mockServer.TestServer.URL,
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	r, err := NewClient(
		&ClientOptions{
			CREcClient: c,
		},
	)
	if err != nil {
		t.Fatalf("failed to create event reader: %v", err)
	}

	listener, err := r.CreateListener(
		ctx, &apiClient.CreateListener{
			Service: "dvp",
			Name:    "SettlementAccepted",
			ChainId: "1337",
			Address: "0x1234567890abcdef1234567890abcdef12345678",
		},
	)

	if err != nil {
		t.Fatalf("failed to create event listener: %v", err)
	}

	require.NotNil(t, listener)
	require.Equal(t, "dvp", listener.Service)             // must match the service in the first mockdata event
	require.Equal(t, "SettlementAccepted", listener.Name) // must match the name in the first mockdata event
}
