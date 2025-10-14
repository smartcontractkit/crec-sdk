package server

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/client"
)

func TestMockServer_Health_Events_Listeners_Accounts(t *testing.T) {
	s := NewMockServer()
	defer s.Close()

	// Health via raw HTTP
	resp, err := http.Get(s.TestServer.URL + "/health-check")
	if err != nil {
		t.Fatalf("health-check request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health-check status: %d", resp.StatusCode)
	}

	// Use the generated CREc client against the mock server
	c, err := client.NewCREcClient(
		&client.ClientOptions{
			BaseURL: s.TestServer.URL,
			APIKey:  "test-key",
		},
	)
	if err != nil {
		t.Fatalf("NewCREcClient: %v", err)
	}

	// List events
	evs, err := c.GetEventsWithResponse(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetEvents: %v", err)
	}
	if evs.StatusCode() != http.StatusOK {
		t.Fatalf("GetEvents status: %d", evs.StatusCode())
	}

	// Get single event by ID
	eid := uuid.New()
	ev, err := c.GetEventsEventIdWithResponse(context.Background(), eid)
	if err != nil {
		t.Fatalf("GetEventsEventId: %v", err)
	}
	if ev.StatusCode() != http.StatusOK {
		t.Fatalf("GetEventsEventId status: %d", ev.StatusCode())
	}
	if ev.JSON200 == nil || ev.JSON200.EventId != eid {
		t.Fatalf("unexpected event id: %+v", ev.JSON200)
	}

	// Create listener
	l, err := c.PostListenersWithResponse(
		context.Background(), apiClient.CreateListener{
			Service: "dvp",
			Name:    "SettlementAccepted",
			ChainId: "1337",
			Address: "0x1234567890abcdef1234567890abcdef12345678",
		},
	)
	if err != nil {
		t.Fatalf("PostListeners: %v", err)
	}
	if l.StatusCode() != http.StatusCreated || l.JSON201 == nil {
		t.Fatalf("unexpected listener creation response: status=%d body=%+v", l.StatusCode(), l.JSON201)
	}

	// Accounts: create, list, get by id
	testAccountName := "Test Account"
	acc, err := c.PostAccountsWithResponse(
		context.Background(), apiClient.CreateAccount{
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			ChainId: "1337",
			Name:    &testAccountName,
		},
	)
	if err != nil {
		t.Fatalf("PostAccounts: %v", err)
	}
	if acc.StatusCode() != http.StatusCreated || acc.JSON201 == nil {
		t.Fatalf("unexpected account creation response")
	}
	list, err := c.GetAccountsWithResponse(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}
	if list.StatusCode() != http.StatusOK || list.JSON200 == nil {
		t.Fatalf("unexpected accounts list response")
	}
	one, err := c.GetAccountsAccountIdWithResponse(context.Background(), acc.JSON201.AccountId)
	if err != nil {
		t.Fatalf("GetAccountsAccountId: %v", err)
	}
	if one.StatusCode() != http.StatusOK || one.JSON200 == nil {
		t.Fatalf("unexpected account get response")
	}
}
