package dta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

func TestDecode(t *testing.T) {
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
