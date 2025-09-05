package dta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

func TestDecodeSimple(t *testing.T) {
	type testCase struct {
		name          string
		event         apiClient.Event
		expectErr     bool
		expectedEvent VerifiableEvent
	}

	validVerifiableEvent := VerifiableEvent{
		CreatedAt: time.Now(),
		Event: struct {
			Type      string "json:\"type\""
			Name      string "json:\"name\""
			Address   string "json:\"address\""
			RequestId string "json:\"requestId\""
			TopicHash string "json:\"topicHash\""
		}{
			Type:      "exampleType",
			Name:      "exampleName",
			Address:   "exampleAddress",
			RequestId: "exampleRequestId",
			TopicHash: "exampleTopicHash",
		},
		Metadata: Metadata{
			WorkflowEvent: WorkflowEvent{Attributes: map[string]Attribute{
				"event_type":       {Value: EventDistributorRegistered.String()},
				"request_id":       {Value: "exampleRequestId"},
				"distributor_addr": {Value: "exampleAddress"},
			}},
		},
	}

	validVerifiableEventBytes, _ := json.Marshal(validVerifiableEvent)
	validBase64 := base64.StdEncoding.EncodeToString(validVerifiableEventBytes)
	var expectedVerifiableEvent VerifiableEvent
	_ = json.Unmarshal(validVerifiableEventBytes, &expectedVerifiableEvent)

	cases := []testCase{
		{
			name: "Valid base64 and valid JSON",
			event: apiClient.Event{
				VerifiableEvent: validBase64,
			},
			expectErr:     false,
			expectedEvent: expectedVerifiableEvent,
		},
		{
			name: "Invalid base64 string",
			event: apiClient.Event{
				VerifiableEvent: "invalid-base64",
			},
			expectErr: true,
		},
		{
			name: "Valid base64 but invalid JSON",
			event: apiClient.Event{
				VerifiableEvent: base64.StdEncoding.EncodeToString([]byte("invalid-json")),
			},
			expectErr: true,
		},
		{
			name: "Empty verifiable event",
			event: apiClient.Event{
				VerifiableEvent: "",
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name, func(t *testing.T) {
				ctx := context.Background()
				result, err := Decode(ctx, tc.event)

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, tc.expectedEvent, result)
				}
			},
		)
	}
}

// helper to build a verifiable event envelope with attributes
func buildEnvelope(attrs map[string]string, overrideEventName string) VerifiableEvent {
	ve := VerifiableEvent{}
	ve.CreatedAt = time.Unix(0, 0)
	// set outer event name for fallback logic
	ve.Event.Name = overrideEventName
	ve.Metadata.WorkflowEvent.Attributes = make(Attrs)
	for k, v := range attrs {
		ve.Metadata.WorkflowEvent.Attributes[k] = Attribute{Key: k, Value: v}
	}
	return ve
}

func encodeEvent(t *testing.T, ve VerifiableEvent) apiClient.Event {
	b, err := json.Marshal(ve)
	require.NoError(t, err)
	return apiClient.Event{VerifiableEvent: base64.StdEncoding.EncodeToString(b)}
}

func TestDecodeUnmarshal(t *testing.T) {
	cases := []struct {
		name          string
		attrs         map[string]string
		overrideName  string
		wantEventName EventName
		wantEvent     ConcreteEvent
		wantErr       bool
	}{
		{
			name: "DistributorRequestProcessed_Success",
			attrs: map[string]string{
				"event_type": EventDistributorRequestProcessed.String(),
				"request_id": common.HexToHash("0x01").Hex(),
				"shares":     "12345678901234567890",
				"status":     "7",
				"error":      "some-bytes",
			},
			overrideName:  "",
			wantEventName: EventDistributorRequestProcessed,
			wantEvent: &DistributorRequestProcessed{
				RequestId: common.HexToHash("0x01"),
				Shares:    func() *big.Int { i, _ := new(big.Int).SetString("12345678901234567890", 10); return i }(),
				Status:    uint8(7),
				Error:     []byte("some-bytes"),
			},
		},
		{
			name: "SubscriptionRequested_Success_And_FallbackName",
			attrs: map[string]string{
				"fund_token_id":    common.HexToHash("0x02").Hex(),
				"distributor_addr": common.HexToAddress("0x00000000000000000000000000000000000000aa").Hex(),
				"request_id":       common.HexToHash("0x03").Hex(),
				"amount":           "42",
				"created_at":       "12345",
			},
			overrideName:  EventSubscriptionRequested.String(),
			wantEventName: EventSubscriptionRequested,
			wantEvent: &SubscriptionRequested{
				FundTokenId:     common.HexToHash("0x02"),
				DistributorAddr: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
				RequestId:       common.HexToHash("0x03"),
				Amount:          big.NewInt(42),
				CreatedAt:       uint64(12345),
			},
		},
		{
			name: "ErrorOnInvalidNumeric",
			attrs: map[string]string{
				"event_type": EventInitialized.String(),
				"version":    "not-a-number",
			},
			wantErr: true,
		},
		{
			name: "UnsupportedEvent",
			attrs: map[string]string{
				"event_type": "SomeUnknownEvent",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ve := buildEnvelope(tc.attrs, tc.overrideName)
			ev, err := Decode(t.Context(), encodeEvent(t, ve))

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantEventName, ev.EventName())
			require.Equal(t, tc.wantEvent, ev.ConcreteEvent)
		})
	}
}
