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

	apiClient "github.com/smartcontractkit/crec-api-go/client"
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

func TestParseScientificNotationToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // string representation of expected big.Int
		wantOk   bool
	}{
		// Regular decimal numbers
		{
			name:     "Simple integer",
			input:    "123456789",
			expected: "123456789",
			wantOk:   true,
		},
		{
			name:     "Zero",
			input:    "0",
			expected: "0",
			wantOk:   true,
		},
		{
			name:     "Large integer",
			input:    "123456789012345678901234567890",
			expected: "123456789012345678901234567890",
			wantOk:   true,
		},

		// Scientific notation - positive exponents
		{
			name:     "Simple scientific notation",
			input:    "1e18",
			expected: "1000000000000000000",
			wantOk:   true,
		},
		{
			name:     "Scientific notation with decimal",
			input:    "1.2e+21",
			expected: "1200000000000000000000",
			wantOk:   true,
		},
		{
			name:     "Scientific notation uppercase E",
			input:    "5E20",
			expected: "500000000000000000000",
			wantOk:   true,
		},
		{
			name:     "Scientific notation with plus sign",
			input:    "3.14e+5",
			expected: "314000",
			wantOk:   true,
		},
		{
			name:     "Large scientific notation",
			input:    "1.23456789e+30",
			expected: "1234567890000000000000000000000",
			wantOk:   true,
		},

		// Scientific notation - negative exponents (should truncate to integer)
		{
			name:     "Scientific notation negative exponent small",
			input:    "1.5e-1",
			expected: "0", // truncated to integer
			wantOk:   true,
		},
		{
			name:     "Scientific notation negative exponent zero result",
			input:    "5e-10",
			expected: "0", // truncated to integer
			wantOk:   true,
		},
		{
			name:     "Scientific notation that becomes integer",
			input:    "1.23e2",
			expected: "123",
			wantOk:   true,
		},

		// Edge cases
		{
			name:     "Very large scientific notation",
			input:    "1e100",
			expected: "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			wantOk:   true,
		},
		{
			name:     "Scientific notation with many decimal places",
			input:    "1.23456789123456789e+20",
			expected: "123456789123456789000",
			wantOk:   true,
		},

		// Invalid inputs
		{
			name:     "Invalid format",
			input:    "not-a-number",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Invalid scientific notation",
			input:    "1e",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Multiple decimal points",
			input:    "1.2.3e10",
			expected: "",
			wantOk:   false,
		},
		{
			name:     "Invalid exponent",
			input:    "1eabc",
			expected: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseScientificNotationToBigInt(tt.input)

			require.Equal(t, tt.wantOk, ok, "parseScientificNotationToBigInt(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)

			if tt.wantOk && ok {
				expected := new(big.Int)
				expected.SetString(tt.expected, 10)
				require.Equal(t, expected, result, "parseScientificNotationToBigInt(%q) = %v, want %v", tt.input, result, expected)
			}
		})
	}
}

func TestScientificNotationInEventParsing(t *testing.T) {
	tests := []struct {
		name                   string
		eventType              string
		amountValue            string
		sharesValue            string
		expectError            bool
		expectedAmountOrShares string
	}{
		{
			name:                   "SubscriptionRequested with scientific notation amount",
			eventType:              EventSubscriptionRequested.String(),
			amountValue:            "1.2e+21",
			expectedAmountOrShares: "1200000000000000000000",
			expectError:            false,
		},
		{
			name:                   "SubscriptionRequested with large scientific notation",
			eventType:              EventSubscriptionRequested.String(),
			amountValue:            "5e18",
			expectedAmountOrShares: "5000000000000000000",
			expectError:            false,
		},
		{
			name:                   "RedemptionRequested with scientific notation shares",
			eventType:              EventRedemptionRequested.String(),
			sharesValue:            "3.14159e+20",
			expectedAmountOrShares: "314159000000000000000",
			expectError:            false,
		},
		{
			name:        "SubscriptionRequested with invalid scientific notation",
			eventType:   EventSubscriptionRequested.String(),
			amountValue: "invalid-amount",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attrs map[string]string

			switch tt.eventType {
			case EventSubscriptionRequested.String():
				attrs = map[string]string{
					"event_type":       tt.eventType,
					"fund_token_id":    common.HexToHash("0x02").Hex(),
					"distributor_addr": common.HexToAddress("0x00000000000000000000000000000000000000aa").Hex(),
					"request_id":       common.HexToHash("0x03").Hex(),
					"amount":           tt.amountValue,
					"created_at":       "12345",
				}
			case EventRedemptionRequested.String():
				attrs = map[string]string{
					"event_type":       tt.eventType,
					"fund_token_id":    common.HexToHash("0x02").Hex(),
					"distributor_addr": common.HexToAddress("0x00000000000000000000000000000000000000aa").Hex(),
					"request_id":       common.HexToHash("0x03").Hex(),
					"shares":           tt.sharesValue,
					"created_at":       "12345",
				}
			}

			ve := buildEnvelope(attrs, "")
			ev, err := Decode(context.Background(), encodeEvent(t, ve))

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify the parsed value matches expected
			expected := new(big.Int)
			expected.SetString(tt.expectedAmountOrShares, 10)

			switch concreteEvent := ev.ConcreteEvent.(type) {
			case *SubscriptionRequested:
				require.Equal(t, expected, concreteEvent.Amount, "Amount should match expected value")
			case *RedemptionRequested:
				require.Equal(t, expected, concreteEvent.Shares, "Shares should match expected value")
			default:
				t.Fatalf("Unexpected concrete event type: %T", concreteEvent)
			}
		})
	}
}
