package events

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"
)

const (
	testAPIKey        = "test-api-key"
	testWorkflowOwner = "0x853d51d5d9935964267a5050aC53aa63ECA39bc5"
	testOrgID         = "test-org-123"
)

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{Logger: logger, CRECClient: crecClient})
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, crecClient, c.crecClient)
		assert.Equal(t, logger, c.logger)
	})

	t.Run("NilOptions", func(t *testing.T) {
		c, err := NewClient(nil)
		require.Error(t, err)
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrOptionsRequired)
	})

	t.Run("NilCRECClient", func(t *testing.T) {
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{Logger: logger})
		require.Error(t, err)
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrCRECClientRequired)
	})

	t.Run("DefaultLogger", func(t *testing.T) {
		crecClient := newCRECClient(t, "http://localhost:8080")
		c, err := NewClient(&Options{CRECClient: crecClient})
		require.NoError(t, err)
		assert.NotNil(t, c.logger)
	})
}

func TestClient_ListEvents(t *testing.T) {
	privKeys, addresses := generateTestKeys(t, 2)
	channelID := uuid.New()

	t.Run("Success", func(t *testing.T) {

		events := createTestEventsWithKeys(t, 3, privKeys)
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String()+"/events", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Server returns apiClient.EventList with events and has_more
			response := apiClient.EventList{
				Events:  events,
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})
		defer server.Close()

		eventsList, hasMore, err := c.Poll(context.Background(), channelID, nil)
		require.NoError(t, err)
		assert.Len(t, eventsList, 3)
		assert.False(t, hasMore)

		isEventVerified, err := c.VerifyWithWorkflowOwner(&eventsList[0], testWorkflowOwner)
		require.NoError(t, err)
		require.True(t, isEventVerified)
	})

	t.Run("WithParams", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		limit := 2
		offset := int64(10)

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "2", q.Get("limit"))
			assert.Equal(t, "10", q.Get("offset"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Server returns apiClient.EventList
			response := apiClient.EventList{
				Events:  events,
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsParams{
			Limit:  &limit,
			Offset: &offset,
		}
		eventsList, hasMore, err := c.Poll(context.Background(), channelID, params)
		require.NoError(t, err)
		// response unpacked
		assert.Len(t, eventsList, 2)
		assert.False(t, hasMore)
	})

	t.Run("NilChannelID", func(t *testing.T) {
		c := setupLocalClient(t)
		_, _, err := c.Poll(context.Background(), uuid.Nil, nil)
		require.Error(t, err)
		// nil response checked
		assert.ErrorIs(t, err, ErrChannelIDRequired)
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.Poll(context.Background(), channelID, nil)
		require.Error(t, err)
		// nil response checked
		assert.ErrorIs(t, err, ErrChannelNotFound)
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.Poll(context.Background(), channelID, nil)
		require.Error(t, err)
		// nil response checked
		assert.ErrorIs(t, err, ErrPollEvents)
		assert.ErrorIs(t, err, ErrUnexpectedStatusCode)
	})

	t.Run("NilResponseBody", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.Poll(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNilResponseBody)
	})

	t.Run("WithPagination", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		limit := 10
		offset := int64(5)

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "10", q.Get("limit"))
			assert.Equal(t, "5", q.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: true,
			}
			_ = json.NewEncoder(w).Encode(response)
		}

		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsParams{
			Limit:  &limit,
			Offset: &offset,
		}
		eventsList, hasMore, err := c.Poll(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 2)
		assert.True(t, hasMore)
	})
}

func TestClient_SearchEvents(t *testing.T) {
	privKeys, addresses := generateTestKeys(t, 2)
	channelID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 3, privKeys)
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String()+"/events/search", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})
		defer server.Close()

		eventsList, hasMore, err := c.SearchEvents(context.Background(), channelID, nil)
		require.NoError(t, err)
		assert.Len(t, eventsList, 3)
		assert.False(t, hasMore)

		isEventVerified, err := c.VerifyWithWorkflowOwner(&eventsList[0], testWorkflowOwner)
		require.NoError(t, err)
		require.True(t, isEventVerified)
	})

	t.Run("WithAllParams", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		limit := 50
		offset := int64(10)
		typeVal := []apiClient.EventType{apiClient.EventTypeWatcherEvent}
		createdLt := int64(1700000000)
		createdLte := int64(1700000001)
		createdGt := int64(1600000000)
		createdGte := int64(1600000001)
		chainSelectors := []string{"5009297550715157269"}
		status := []string{string(apiClient.OperationStatusConfirmed)}
		watcherID := uuid.New()
		address := []string{"0x1234567890123456789012345678901234567890"}
		walletOperationID := "op-123"
		operationID := uuid.New().String()
		eventName := "Transfer"
		service := []string{"dvp"}

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "50", q.Get("limit"))
			assert.Equal(t, "10", q.Get("offset"))
			assert.Equal(t, "watcher.event", q.Get("type"))
			assert.Equal(t, "1700000000", q.Get("created.lt"))
			assert.Equal(t, "1700000001", q.Get("created.lte"))
			assert.Equal(t, "1600000000", q.Get("created.gt"))
			assert.Equal(t, "1600000001", q.Get("created.gte"))
			assert.Equal(t, "5009297550715157269", q.Get("chain_selector"))
			assert.Equal(t, string(apiClient.OperationStatusConfirmed), q.Get("status"))
			assert.Equal(t, watcherID.String(), q.Get("watcher_id"))
			assert.Equal(t, "0x1234567890123456789012345678901234567890", q.Get("address"))
			assert.Equal(t, "op-123", q.Get("wallet_operation_id"))
			assert.Equal(t, operationID, q.Get("operation_id"))
			assert.Equal(t, "Transfer", q.Get("event_name"))
			assert.Equal(t, "dvp", q.Get("service"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: true,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsSearchParams{
			Limit:             &limit,
			Offset:            &offset,
			Type:              &typeVal,
			CreatedLt:         &createdLt,
			CreatedLte:        &createdLte,
			CreatedGt:         &createdGt,
			CreatedGte:        &createdGte,
			ChainSelector:     &chainSelectors,
			Status:            &status,
			WatcherId:         &watcherID,
			Address:           &address,
			WalletOperationId: &walletOperationID,
			OperationId:       &operationID,
			EventName:         &eventName,
			Service:           &service,
		}
		eventsList, hasMore, err := c.SearchEvents(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 2)
		assert.True(t, hasMore)
	})

	t.Run("WithTimestampFilters", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 1, privKeys)
		createdGte := int64(1600000000)
		createdLte := int64(1700000000)

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "1600000000", q.Get("created.gte"))
			assert.Equal(t, "1700000000", q.Get("created.lte"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsSearchParams{
			CreatedGte: &createdGte,
			CreatedLte: &createdLte,
		}
		eventsList, hasMore, err := c.SearchEvents(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 1)
		assert.False(t, hasMore)
	})

	t.Run("WithTypeFilter", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		typeVal := []apiClient.EventType{apiClient.EventTypeOperationStatus}

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "operation.status", q.Get("type"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsSearchParams{
			Type: &typeVal,
		}
		eventsList, hasMore, err := c.SearchEvents(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 2)
		assert.False(t, hasMore)
	})

	t.Run("NilChannelID", func(t *testing.T) {
		c := setupLocalClient(t)
		_, _, err := c.SearchEvents(context.Background(), uuid.Nil, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrChannelIDRequired)
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.SearchEvents(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrChannelNotFound)
	})

	t.Run("BadRequest", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := apiClient.ApplicationError{
				Type:    "Bad request",
				Message: "Invalid parameter combination",
			}
			_ = json.NewEncoder(w).Encode(response)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.SearchEvents(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSearchEvents)
		assert.ErrorIs(t, err, ErrBadRequest)
		assert.Contains(t, err.Error(), "Invalid parameter combination")
	})

	t.Run("BadRequest_NoApplicationError", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			// No JSON body, so JSON400 will be nil
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.SearchEvents(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSearchEvents)
		assert.ErrorIs(t, err, ErrBadRequest)
		assert.Contains(t, err.Error(), "Invalid request parameters")
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.SearchEvents(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSearchEvents)
		assert.ErrorIs(t, err, ErrUnexpectedStatusCode)
	})

	t.Run("NilResponseBody", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.SearchEvents(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNilResponseBody)
	})

	t.Run("WithPagination", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		limit := 10
		offset := int64(5)

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "10", q.Get("limit"))
			assert.Equal(t, "5", q.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.EventList{
				Events:  events,
				HasMore: true,
			}
			_ = json.NewEncoder(w).Encode(response)
		}

		c, server := setupTestClient(t, handler)
		defer server.Close()

		params := &apiClient.GetChannelsChannelIdEventsSearchParams{
			Limit:  &limit,
			Offset: &offset,
		}
		eventsList, hasMore, err := c.SearchEvents(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 2)
		assert.True(t, hasMore)
	})
}

func TestClient_EventHash(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("ComputesValidHash", func(t *testing.T) {
		eventPayload := createTestEventPayload(t)

		hash, err := c.EventHash(&eventPayload)
		require.NoError(t, err)
		assert.NotEqual(t, common.Hash{}, hash)

		// Verify the hash is 32 bytes (256 bits)
		assert.Equal(t, 32, len(hash.Bytes()))
	})

	t.Run("DeterministicHash", func(t *testing.T) {
		// Same payload should produce the same hash
		eventPayload := createTestEventPayload(t)

		hash1, err := c.EventHash(&eventPayload)
		require.NoError(t, err)

		hash2, err := c.EventHash(&eventPayload)
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2, "same event payload should produce same hash")
	})

	t.Run("DifferentDataProducesDifferentHash", func(t *testing.T) {
		eventPayload1 := createTestEventPayload(t)
		eventPayload2 := createTestEventPayload(t)
		differentData := map[string]interface{}{
			"from":  "0x0000000000000000000000000000000000000000",
			"to":    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			"value": "2000000000000000000",
		}
		differentDataBytes, _ := json.Marshal(differentData)
		eventPayload2.VerifiableEvent = base64.StdEncoding.EncodeToString(differentDataBytes)

		hash1, err := c.EventHash(&eventPayload1)
		require.NoError(t, err)

		hash2, err := c.EventHash(&eventPayload2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("VerifyHashFormat", func(t *testing.T) {
		eventPayload := createTestEventPayload(t)

		hash, err := c.EventHash(&eventPayload)
		require.NoError(t, err)

		// Hash is now just keccak256(verifiableEvent)
		expectedHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))

		assert.Equal(t, expectedHash, hash, "hash should match expected Keccak256 computation")
	})
}

func TestClient_OperationStatusHash(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("ComputesValidHash", func(t *testing.T) {
		eventPayload := createTestOperationStatusPayload(t)

		hash, err := c.OperationStatusHash(&eventPayload)
		require.NoError(t, err)
		assert.NotEqual(t, common.Hash{}, hash)

		// Verify the hash is 32 bytes (256 bits)
		assert.Equal(t, 32, len(hash.Bytes()))
	})

	t.Run("DeterministicHash", func(t *testing.T) {
		// Same payload should produce the same hash
		eventPayload := createTestOperationStatusPayload(t)

		hash1, err := c.OperationStatusHash(&eventPayload)
		require.NoError(t, err)

		hash2, err := c.OperationStatusHash(&eventPayload)
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2, "same operation status payload should produce same hash")
	})

	t.Run("DifferentVerifiableEventProducesDifferentHash", func(t *testing.T) {
		eventPayload1 := createTestOperationStatusPayload(t)
		eventPayload2 := createTestOperationStatusPayload(t)

		// Create different verifiable event data
		differentData := base64.StdEncoding.EncodeToString([]byte(`{"operationId":"test-op-456","status":"failed"}`))
		eventPayload2.VerifiableEvent = &differentData

		hash1, err := c.OperationStatusHash(&eventPayload1)
		require.NoError(t, err)

		hash2, err := c.OperationStatusHash(&eventPayload2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "different verifiable events should produce different hashes")
	})

	t.Run("VerifyHashFormat", func(t *testing.T) {
		eventPayload := createTestOperationStatusPayload(t)

		hash, err := c.OperationStatusHash(&eventPayload)
		require.NoError(t, err)

		// Hash is now just keccak256(verifiableEvent)
		expectedHash := crypto.Keccak256Hash([]byte(*eventPayload.VerifiableEvent))

		assert.Equal(t, expectedHash, hash, "hash should match expected Keccak256 computation")
	})

	t.Run("NilVerifiableEvent_ReturnsError", func(t *testing.T) {
		eventPayload := apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusAccepted,
			StatusReason:      "Operation pending",
			Timestamp:         1700000000,
			VerifiableEvent:   nil,
		}

		hash, err := c.OperationStatusHash(&eventPayload)
		require.Error(t, err)
		assert.Equal(t, common.Hash{}, hash)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.Contains(t, err.Error(), "verifiable event is required")
	})

	t.Run("EmptyVerifiableEvent_ReturnsError", func(t *testing.T) {
		emptyVerifiableEvent := ""
		eventPayload := apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusAccepted,
			StatusReason:      "Operation pending",
			Timestamp:         1700000000,
			VerifiableEvent:   &emptyVerifiableEvent,
		}

		hash, err := c.OperationStatusHash(&eventPayload)
		require.Error(t, err)
		assert.Equal(t, common.Hash{}, hash)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.Contains(t, err.Error(), "verifiable event is required")
	})
}

func TestEvents_WorkflowOwnerFromOrgID(t *testing.T) {
	c := setupLocalClient(t)

	tests := []struct {
		name  string
		orgID string
	}{
		{name: "SimpleOrgID", orgID: "test-org-123"},
		{name: "UUIDOrgID", orgID: "550e8400-e29b-41d4-a716-446655440000"},
		{name: "EmptyOrgID", orgID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := c.WorkflowOwnerFromOrgID(tt.orgID)
			require.NoError(t, err)
			assert.NotEmpty(t, addr)
			assert.Regexp(t, `^0x[0-9a-fA-F]{40}$`, addr)
		})
	}

	t.Run("Deterministic", func(t *testing.T) {
		addr1, err := c.WorkflowOwnerFromOrgID("my-org")
		require.NoError(t, err)
		addr2, err := c.WorkflowOwnerFromOrgID("my-org")
		require.NoError(t, err)
		assert.Equal(t, addr1, addr2)
	})

	t.Run("DifferentOrgIDsProduceDifferentAddresses", func(t *testing.T) {
		addr1, err := c.WorkflowOwnerFromOrgID("org-alpha")
		require.NoError(t, err)
		addr2, err := c.WorkflowOwnerFromOrgID("org-beta")
		require.NoError(t, err)
		assert.NotEqual(t, addr1, addr2)
	})
}

func TestClient_Verify(t *testing.T) {
	c := setupLocalClient(t)
	orgOwner, err := c.WorkflowOwnerFromOrgID(testOrgID)
	require.NoError(t, err)

	t.Run("VerifyWithDefaultOrgID", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.OrgID = testOrgID
		})

		ok, err := c.Verify(event)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("VerifyWithDefaultWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.WorkflowOwner = testWorkflowOwner
		})

		ok, err := c.Verify(event)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ErrOrgIDOrWorkflowOwnerReq", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.Verify(event)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOrgIDOrWorkflowOwnerReq)
	})

	t.Run("VerifyWithOrgID_HappyPath", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithOrgID(event, testOrgID)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("VerifyWithOrgID_WrongOrgID", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithOrgID(event, "wrong-org-id")
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})
}

func TestClient_VerifyWithWorkflowOwner(t *testing.T) {
	t.Run("VerifyWithExplicitWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.WorkflowOwner = testWorkflowOwner
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ErrVerificationNotConfigured", func(t *testing.T) {
		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 0
			opts.ValidSigners = nil
		})

		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerificationNotConfigured)
	})

	t.Run("ErrVerificationNotConfigured_EmptySigners", func(t *testing.T) {
		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 0
			opts.ValidSigners = []string{}
		})

		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerificationNotConfigured)
	})

	t.Run("HappyPath_TwoValidSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("NotEnoughSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 3)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys[:2], &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 3
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("IncorrectSigners", func(t *testing.T) {
		signingKeys, _ := generateTestKeys(t, 2)
		_, validAddresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, signingKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = validAddresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("ErrNoOCRProofs", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestEventPayload(t)
		payloadUnion := apiClient.Event_Payload{}
		err := payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.ErrorIs(t, err, ErrNoOCRProofs)
	})

	t.Run("ErrOCRReportTooShort", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestEventPayload(t)

		// Create a short OCR report (less than 141 bytes)
		shortOcrReport := make([]byte, 50)
		ocrContext := []byte("test")

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(shortOcrReport),
			Signatures: []string{"0x1234"},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOCRReportTooShort)
	})

	t.Run("ErrInvalidWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		// Place a wrong workflow owner at offset 87 (20 bytes)
		wrongWorkflowOwner := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		copy(ocrReport[87:107], wrongWorkflowOwner.Bytes())

		eventHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))
		copy(ocrReport[109:], eventHash.Bytes())

		ocrContext := []byte("test-context")
		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		sig, _ := crypto.Sign(reportHash.Bytes(), privKeys[0])
		sig[64] += 27

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x" + common.Bytes2Hex(sig)},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 1
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})

	t.Run("ErrInvalidEventHash", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		// Place correct workflow owner at offset 87 (20 bytes)
		workflowOwner := common.HexToAddress(testWorkflowOwner)
		copy(ocrReport[87:107], workflowOwner.Bytes())

		wrongHash := crypto.Keccak256Hash([]byte("wrong-data"))
		copy(ocrReport[109:], wrongHash.Bytes()) // Wrong hash!

		ocrContext := []byte("test-context")
		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		sig, _ := crypto.Sign(reportHash.Bytes(), privKeys[0])
		sig[64] += 27

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x" + common.Bytes2Hex(sig)},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 1
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})

	t.Run("ErrMultipleOCRProofs", func(t *testing.T) {
		c := setupLocalClient(t)
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create two identical OCR proofs
		ocrReport := make([]byte, 141)

		// Place workflow owner at offset 87 (20 bytes)
		workflowOwner := common.HexToAddress(testWorkflowOwner)
		copy(ocrReport[87:107], workflowOwner.Bytes())

		ocrContext := []byte("test")
		eventHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))
		copy(ocrReport[109:], eventHash.Bytes())

		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))
		sig, _ := crypto.Sign(reportHash.Bytes(), privKeys[0])
		sig[64] += 27

		ocrProof1 := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x" + common.Bytes2Hex(sig)},
		}
		ocrProof2 := ocrProof1 // Duplicate

		proof1 := apiClient.EventHeaders_Proofs_Item{}
		err := proof1.FromOCRProof(ocrProof1)
		require.NoError(t, err)
		proof2 := apiClient.EventHeaders_Proofs_Item{}
		err = proof2.FromOCRProof(ocrProof2)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proof1, proof2}, // Multiple proofs
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.ErrorIs(t, err, ErrMultipleOCRProofs)
	})

	t.Run("ErrParseOCRReport", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestEventPayload(t)

		// Create OCR proof with invalid hex string for report (odd length hex will fail)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x1234",
			OcrReport:  "0x123", // Invalid - odd length hex!
			Signatures: []string{"0x1234"},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrParseOCRReport)
	})

	t.Run("ErrParseOCRContext", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestEventPayload(t)

		// Create valid OCR report but invalid context (odd length hex will fail)
		ocrReport := make([]byte, 141)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0xabc", // Invalid - odd length hex!
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x1234"},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrParseOCRContext)
	})

	t.Run("ErrParseSignature", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create valid OCR report
		ocrReport := make([]byte, 141)

		// Place workflow owner at offset 87 (20 bytes)
		workflowOwner := common.HexToAddress(testWorkflowOwner)
		copy(ocrReport[87:107], workflowOwner.Bytes())

		eventHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))
		copy(ocrReport[109:], eventHash.Bytes())

		ocrContext := []byte("test")

		// Create OCR proof with invalid signature hex (odd length will fail)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0xabc"}, // Invalid - odd length hex!
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 1
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrParseSignature)
	})

	t.Run("ErrRecoverPubKeyFromSignature", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create valid OCR report
		ocrReport := make([]byte, 141)

		// Place workflow owner at offset 87 (20 bytes)
		workflowOwner := common.HexToAddress(testWorkflowOwner)
		copy(ocrReport[87:107], workflowOwner.Bytes())

		eventHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))
		copy(ocrReport[109:], eventHash.Bytes())

		ocrContext := []byte("test")

		// Create a signature with correct length (65 bytes) but invalid data
		invalidSig := make([]byte, 65) // Correct length but all zeros = invalid signature
		invalidSig[64] = 27            // Set v value properly

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x" + common.Bytes2Hex(invalidSig)}, // Valid length but invalid signature data
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromWatcherEventPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 1
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrRecoverPubKeyFromSignature)
	})

	// Test that wrong header type returns error
	t.Run("ErrOnlyWatcherEventsSupported_WrongHeaderType", func(t *testing.T) {
		c := setupLocalClient(t)

		// Provide properly sized OCR report (141 bytes minimum) to pass validation
		ocrReport := make([]byte, 141)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x01",
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{},
		}
		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		require.NoError(t, proofUnion.FromOCRProof(ocrProof))

		// Build a valid WatcherStatusPayload (not WatcherEventPayload)
		statusPayload := apiClient.WatcherStatusPayload{
			ChainSelector: "16015286601757825753",
			WatcherId:     "550e8400-e29b-41d4-a716-446655440000",
			Status:        apiClient.WatcherEventStatusPending,
			StatusReason:  "Watcher is pending",
			Timestamp:     1700000000,
		}

		payloadUnion := apiClient.Event_Payload{}
		require.NoError(t, payloadUnion.FromWatcherStatusPayload(statusPayload))

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherStatus, // Wrong type for Verify()
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOnlyWatcherEventsSupported)
	})
}

func TestClient_VerifyOperationStatus(t *testing.T) {
	c := setupLocalClient(t)
	orgOwner, err := c.WorkflowOwnerFromOrgID(testOrgID)
	require.NoError(t, err)

	t.Run("VerifyWithDefaultOrgID", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.OrgID = testOrgID
		})

		ok, err := c.VerifyOperationStatus(event)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("VerifyOperationStatusWithDefaultWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.WorkflowOwner = testWorkflowOwner
		})

		ok, err := c.VerifyOperationStatus(event)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ErrOrgIDOrWorkflowOwnerReq", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatus(event)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOrgIDOrWorkflowOwnerReq)
	})

	t.Run("VerifyOperationStatusWithOrgID_HappyPath", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithOrgID(event, testOrgID)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("VerifyOperationStatusWithOrgID_WrongOrgID", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, orgOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithOrgID(event, "wrong-org-id")
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})
}

func TestClient_VerifyOperationStatusWithWorkflowOwner(t *testing.T) {
	t.Run("VerifyWithExplicitWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
			opts.WorkflowOwner = testWorkflowOwner
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ErrVerificationNotConfigured", func(t *testing.T) {
		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 0
			opts.ValidSigners = nil
		})

		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerificationNotConfigured)
	})

	t.Run("ErrVerificationNotConfigured_EmptySigners", func(t *testing.T) {
		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 0
			opts.ValidSigners = []string{}
		})

		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerificationNotConfigured)
	})

	t.Run("HappyPath_TwoValidSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("NotEnoughSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 3)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, privKeys[:2], &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 3
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("IncorrectSigners", func(t *testing.T) {
		signingKeys, _ := generateTestKeys(t, 2)
		_, validAddresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)
		event := createValidOperationStatusEventForOwner(t, signingKeys, &eventPayload, testWorkflowOwner)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = validAddresses
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("ErrNoOCRProofs", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestOperationStatusPayload(t)
		payloadUnion := apiClient.Event_Payload{}
		err := payloadUnion.FromOperationStatusPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.ErrorIs(t, err, ErrNoOCRProofs)
	})

	t.Run("ErrOCRReportTooShort", func(t *testing.T) {
		c := setupLocalClient(t)
		eventPayload := createTestOperationStatusPayload(t)

		shortOcrReport := make([]byte, 50)
		ocrContext := []byte("test")

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(shortOcrReport),
			Signatures: []string{"0x1234"},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromOperationStatusPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeOperationStatus,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOCRReportTooShort)
	})

	t.Run("ErrInvalidWorkflowOwner", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		// Place a wrong workflow owner at offset 87 (20 bytes)
		wrongWorkflowOwner := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		copy(ocrReport[87:107], wrongWorkflowOwner.Bytes())

		// Compute event hash using OperationStatusHash pattern
		eventHash := crypto.Keccak256Hash([]byte(*eventPayload.VerifiableEvent))
		copy(ocrReport[109:], eventHash.Bytes())

		ocrContext := []byte("test-context")
		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		sig, _ := crypto.Sign(reportHash.Bytes(), privKeys[0])
		sig[64] += 27

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{"0x" + common.Bytes2Hex(sig)},
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromOperationStatusPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeOperationStatus,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 1
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})

	t.Run("ErrInvalidEventHash", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestOperationStatusPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		// Place correct workflow owner at offset 87 (20 bytes)
		workflowOwner := common.HexToAddress(testWorkflowOwner)
		copy(ocrReport[87:107], workflowOwner.Bytes())

		// Use wrong event hash
		wrongHash := common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		copy(ocrReport[109:], wrongHash.Bytes())

		ocrContext := []byte("test-context")
		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		var signatures []string
		for _, privKey := range privKeys {
			sig, _ := crypto.Sign(reportHash.Bytes(), privKey)
			sig[64] += 27
			signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
		}

		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x" + common.Bytes2Hex(ocrContext),
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: signatures,
		}

		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		err := proofUnion.FromOCRProof(ocrProof)
		require.NoError(t, err)

		payloadUnion := apiClient.Event_Payload{}
		err = payloadUnion.FromOperationStatusPayload(eventPayload)
		require.NoError(t, err)

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeOperationStatus,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrInvalidEventHash)
	})

	t.Run("ErrOnlyOperationStatusSupported_WrongHeaderType", func(t *testing.T) {
		c := setupLocalClient(t)

		ocrReport := make([]byte, 141)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x01",
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{},
		}
		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		require.NoError(t, proofUnion.FromOCRProof(ocrProof))

		// Build a valid WatcherEventPayload (not OperationStatusPayload)
		eventPayload := createTestEventPayload(t)

		payloadUnion := apiClient.Event_Payload{}
		require.NoError(t, payloadUnion.FromWatcherEventPayload(eventPayload))

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeWatcherEvent, // Wrong type for VerifyOperationStatus()
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrOnlyOperationStatusSupported)
	})

	t.Run("ErrVerifiableEventRequired", func(t *testing.T) {
		c := setupLocalClient(t)

		// Create operation status payload without VerifiableEvent
		operationStatusPayload := apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   nil,
		}

		ocrReport := make([]byte, 141)
		ocrProof := apiClient.OCRProof{
			Alg:        "ecdsa-secp256k1",
			OcrContext: "0x01",
			OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
			Signatures: []string{},
		}
		proofUnion := apiClient.EventHeaders_Proofs_Item{}
		require.NoError(t, proofUnion.FromOCRProof(ocrProof))

		payloadUnion := apiClient.Event_Payload{}
		require.NoError(t, payloadUnion.FromOperationStatusPayload(operationStatusPayload))

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Type:   apiClient.EventTypeOperationStatus,
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.VerifyOperationStatusWithWorkflowOwner(event, testWorkflowOwner)
		require.Error(t, err)
		assert.False(t, ok)
		assert.ErrorIs(t, err, ErrVerifyEvent)
		assert.Contains(t, err.Error(), "verifiable event is required")
	})
}

func TestEvents_VerifyOCRSignatures(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 3)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01
		ocrContext := []byte("test-context-data")

		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		var signatures []string
		for _, privKey := range privKeys {
			sig, err := crypto.Sign(reportHash.Bytes(), privKey)
			require.NoError(t, err)
			sig[64] += 27
			signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		valid, err := c.VerifyOCRSignatures(
			"0x"+common.Bytes2Hex(ocrReport),
			"0x"+common.Bytes2Hex(ocrContext),
			signatures,
		)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("ErrVerificationNotConfigured", func(t *testing.T) {
		c := setupLocalClient(t, func(opts *Options) {
			opts.ValidSigners = nil
		})

		valid, err := c.VerifyOCRSignatures("0x01", "0x01", []string{})
		require.Error(t, err)
		assert.False(t, valid)
		assert.ErrorIs(t, err, ErrVerificationNotConfigured)
	})

	t.Run("NotEnoughValidSignatures", func(t *testing.T) {
		privKeys, _ := generateTestKeys(t, 2)
		_, otherAddresses := generateTestKeys(t, 2)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01
		ocrContext := []byte("test-context-data")

		reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

		var signatures []string
		for _, privKey := range privKeys {
			sig, err := crypto.Sign(reportHash.Bytes(), privKey)
			require.NoError(t, err)
			sig[64] += 27
			signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
		}

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = otherAddresses
		})

		valid, err := c.VerifyOCRSignatures(
			"0x"+common.Bytes2Hex(ocrReport),
			"0x"+common.Bytes2Hex(ocrContext),
			signatures,
		)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("ErrOCRReportTooShort", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		valid, err := c.VerifyOCRSignatures("0x01", "0x01", []string{})
		require.Error(t, err)
		assert.False(t, valid)
		assert.ErrorIs(t, err, ErrOCRReportTooShort)
	})

	t.Run("InvalidOCRContextFormat", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)

		c := setupLocalClient(t, func(opts *Options) {
			opts.MinRequiredSignatures = 2
			opts.ValidSigners = addresses
		})

		ocrReport := make([]byte, 141)
		valid, err := c.VerifyOCRSignatures("0x"+common.Bytes2Hex(ocrReport), "0xZZZ", []string{})
		require.Error(t, err)
		assert.False(t, valid)
		assert.ErrorIs(t, err, ErrParseOCRContext)
	})
}

func TestClient_Decode(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("Success", func(t *testing.T) {
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		var decoded map[string]interface{}
		err := c.Decode(event, &decoded)
		require.NoError(t, err)
		assert.NotNil(t, decoded)
	})
}

func TestClient_ToJson(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("Success", func(t *testing.T) {
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventForOwner(t, privKeys, &eventPayload, testWorkflowOwner)

		jsonBytes, err := c.ToJSON(*event)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonBytes)

		// Verify it's valid JSON
		var decoded map[string]interface{}
		err = json.Unmarshal(jsonBytes, &decoded)
		require.NoError(t, err)
	})
}

func TestClient_DecodeVerifiableEvent(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("Success_WithEVMChainEvent", func(t *testing.T) {
		// Watcher events always have chain events - create a proper EVMEvent
		evmEvent := models.EVMEvent{
			Address:        "0x1234567890123456789012345678901234567890",
			BlockNumber:    12345678,
			BlockTimestamp: 1700000000,
			ChainId:        "1",
			EventSignature: "Transfer(address,address,uint256)",
			LogIndex:       5,
			TopicHash:      "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
			TxHash:         "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Params: &map[string]interface{}{
				"from":  "0x0000000000000000000000000000000000000000",
				"to":    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
				"value": "1000000000000000000",
			},
		}

		chainEvent := &models.VerifiableEvent_ChainEvent{}
		err := chainEvent.FromEVMEvent(evmEvent)
		require.NoError(t, err)

		chainFamily := "evm"
		chainSelector := "5009297550715157269"
		service := "watcher"
		verifiableEvent := models.VerifiableEvent{
			Name:          "Transfer",
			Timestamp:     time.Now().UTC().Truncate(time.Second),
			ChainFamily:   &chainFamily,
			ChainSelector: &chainSelector,
			Service:       &service,
			ChainEvent:    chainEvent,
		}

		verifiableEventBytes, err := json.Marshal(verifiableEvent)
		require.NoError(t, err)
		verifiableEventBase64 := base64.StdEncoding.EncodeToString(verifiableEventBytes)

		payload := &apiClient.WatcherEventPayload{
			WatcherId:       "550e8400-e29b-41d4-a716-446655440000",
			VerifiableEvent: verifiableEventBase64,
			EventHash:       crypto.Keccak256Hash([]byte(verifiableEventBase64)).Hex(),
		}

		decoded, err := c.DecodeVerifiableEvent(payload)
		require.NoError(t, err)
		require.NotNil(t, decoded)

		// Verify metadata
		assert.Equal(t, "Transfer", decoded.Name)
		assert.Equal(t, verifiableEvent.Timestamp, decoded.Timestamp)
		require.NotNil(t, decoded.ChainFamily)
		assert.Equal(t, chainFamily, *decoded.ChainFamily)
		require.NotNil(t, decoded.ChainSelector)
		assert.Equal(t, chainSelector, *decoded.ChainSelector)
		require.NotNil(t, decoded.Service)
		assert.Equal(t, service, *decoded.Service)
		assert.Nil(t, decoded.Data) // Watcher events don't use Data field

		// Verify ChainEvent
		require.NotNil(t, decoded.ChainEvent)
		decodedEVMEvent, err := decoded.ChainEvent.AsEVMEvent()
		require.NoError(t, err)

		assert.Equal(t, evmEvent.Address, decodedEVMEvent.Address)
		assert.Equal(t, evmEvent.BlockNumber, decodedEVMEvent.BlockNumber)
		assert.Equal(t, evmEvent.BlockTimestamp, decodedEVMEvent.BlockTimestamp)
		assert.Equal(t, evmEvent.ChainId, decodedEVMEvent.ChainId)
		assert.Equal(t, evmEvent.EventSignature, decodedEVMEvent.EventSignature)
		assert.Equal(t, evmEvent.LogIndex, decodedEVMEvent.LogIndex)
		assert.Equal(t, evmEvent.TopicHash, decodedEVMEvent.TopicHash)
		assert.Equal(t, evmEvent.TxHash, decodedEVMEvent.TxHash)
		require.NotNil(t, decodedEVMEvent.Params)
		assert.Equal(t, "0x0000000000000000000000000000000000000000", (*decodedEVMEvent.Params)["from"])
	})

	t.Run("Error_NilPayload", func(t *testing.T) {
		decoded, err := c.DecodeVerifiableEvent(nil)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "payload is nil")
	})

	t.Run("Error_EmptyVerifiableEvent", func(t *testing.T) {
		payload := &apiClient.WatcherEventPayload{
			WatcherId:       "550e8400-e29b-41d4-a716-446655440000",
			VerifiableEvent: "",
			EventHash:       "0x1234",
		}

		decoded, err := c.DecodeVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "verifiable event is empty")
	})

	t.Run("Error_InvalidBase64", func(t *testing.T) {
		payload := &apiClient.WatcherEventPayload{
			WatcherId:       "550e8400-e29b-41d4-a716-446655440000",
			VerifiableEvent: "not-valid-base64!!!",
			EventHash:       "0x1234",
		}

		decoded, err := c.DecodeVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "invalid base64")
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		// Valid base64 but invalid JSON
		invalidJSON := base64.StdEncoding.EncodeToString([]byte("not valid json {{{"))
		payload := &apiClient.WatcherEventPayload{
			WatcherId:       "550e8400-e29b-41d4-a716-446655440000",
			VerifiableEvent: invalidJSON,
			EventHash:       "0x1234",
		}

		decoded, err := c.DecodeVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}

func TestClient_DecodeOperationStatusVerifiableEvent(t *testing.T) {
	c := setupLocalClient(t)

	t.Run("Success_ConfirmedWithChainEvent", func(t *testing.T) {
		// Confirmed operation status events have chain events (the on-chain transaction)
		evmEvent := models.EVMEvent{
			Address:        "0x1234567890123456789012345678901234567890",
			BlockNumber:    12345678,
			BlockTimestamp: 1700000000,
			ChainId:        "1",
			EventSignature: "OperationExecuted(bytes32,address)",
			LogIndex:       0,
			TopicHash:      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			TxHash:         "0xdeadbeef1234567890abcdef1234567890abcdef1234567890abcdef12345678",
		}

		chainEvent := &models.VerifiableEvent_ChainEvent{}
		err := chainEvent.FromEVMEvent(evmEvent)
		require.NoError(t, err)

		chainFamily := "evm"
		chainSelector := "5009297550715157269"
		service := "_crec"
		verifiableEvent := models.VerifiableEvent{
			Name:          "OperationConfirmed",
			Timestamp:     time.Now().UTC().Truncate(time.Second),
			ChainFamily:   &chainFamily,
			ChainSelector: &chainSelector,
			Service:       &service,
			ChainEvent:    chainEvent,
		}

		verifiableEventBytes, err := json.Marshal(verifiableEvent)
		require.NoError(t, err)
		verifiableEventBase64 := base64.StdEncoding.EncodeToString(verifiableEventBytes)

		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "5009297550715157269",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   &verifiableEventBase64,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.NoError(t, err)
		require.NotNil(t, decoded)

		// Verify metadata
		assert.Equal(t, "OperationConfirmed", decoded.Name)
		assert.Equal(t, verifiableEvent.Timestamp, decoded.Timestamp)
		require.NotNil(t, decoded.ChainFamily)
		assert.Equal(t, chainFamily, *decoded.ChainFamily)
		require.NotNil(t, decoded.ChainSelector)
		assert.Equal(t, chainSelector, *decoded.ChainSelector)
		require.NotNil(t, decoded.Service)
		assert.Equal(t, service, *decoded.Service)
		assert.Nil(t, decoded.Data) // Confirmed operations have ChainEvent, not Data

		// Verify ChainEvent
		require.NotNil(t, decoded.ChainEvent)
		decodedEVMEvent, err := decoded.ChainEvent.AsEVMEvent()
		require.NoError(t, err)
		assert.Equal(t, evmEvent.TxHash, decodedEVMEvent.TxHash)
	})

	t.Run("Success_FailedWithOperationStatusData", func(t *testing.T) {
		// Failed operation status events have no chain event but contain OperationStatusData in Data
		service := "_crec"
		operationStatusData := models.OperationStatusData{
			Status:            models.Failed,
			StatusReason:      "Insufficient funds",
			WalletAddress:     "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			WalletOperationId: "wallet-op-456",
		}

		// Convert OperationStatusData to map for the Data field
		operationStatusDataBytes, err := json.Marshal(operationStatusData)
		require.NoError(t, err)
		var dataMap map[string]interface{}
		err = json.Unmarshal(operationStatusDataBytes, &dataMap)
		require.NoError(t, err)

		verifiableEvent := models.VerifiableEvent{
			Name:       "OperationFailed",
			Timestamp:  time.Now().UTC().Truncate(time.Second),
			Service:    &service,
			ChainEvent: nil, // No chain event for failed operations
			Data:       &dataMap,
		}

		verifiableEventBytes, err := json.Marshal(verifiableEvent)
		require.NoError(t, err)
		verifiableEventBase64 := base64.StdEncoding.EncodeToString(verifiableEventBytes)

		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-456",
			Status:            apiClient.OperationStatusFailed,
			StatusReason:      "Insufficient funds",
			Timestamp:         1700000000,
			VerifiableEvent:   &verifiableEventBase64,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.NoError(t, err)
		require.NotNil(t, decoded)

		// Verify metadata
		assert.Equal(t, "OperationFailed", decoded.Name)
		require.NotNil(t, decoded.Service)
		assert.Equal(t, service, *decoded.Service)

		// Failed operations have no ChainEvent
		assert.Nil(t, decoded.ChainEvent)

		// But they have OperationStatusData in Data field
		require.NotNil(t, decoded.Data)
		assert.Equal(t, "failed", (*decoded.Data)["status"])
		assert.Equal(t, "Insufficient funds", (*decoded.Data)["status_reason"])
		assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", (*decoded.Data)["wallet_address"])
		assert.Equal(t, "wallet-op-456", (*decoded.Data)["wallet_operation_id"])
	})

	t.Run("Error_NilPayload", func(t *testing.T) {
		decoded, err := c.DecodeOperationStatusVerifiableEvent(nil)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "payload is nil")
	})

	t.Run("Error_NilVerifiableEvent", func(t *testing.T) {
		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   nil,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "verifiable event is nil or empty")
	})

	t.Run("Error_EmptyVerifiableEvent", func(t *testing.T) {
		emptyStr := ""
		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   &emptyStr,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "verifiable event is nil or empty")
	})

	t.Run("Error_InvalidBase64", func(t *testing.T) {
		invalidBase64 := "not-valid-base64!!!"
		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   &invalidBase64,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "invalid base64")
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		// Valid base64 but invalid JSON
		invalidJSON := base64.StdEncoding.EncodeToString([]byte("not valid json {{{"))
		payload := &apiClient.OperationStatusPayload{
			Address:           "0x1234",
			ChainSelector:     "16015286601757825753",
			OperationId:       uuid.New(),
			WalletOperationId: "wallet-op-123",
			Status:            apiClient.OperationStatusConfirmed,
			StatusReason:      "Operation confirmed",
			Timestamp:         1700000000,
			VerifiableEvent:   &invalidJSON,
		}

		decoded, err := c.DecodeOperationStatusVerifiableEvent(payload)
		require.Error(t, err)
		assert.Nil(t, decoded)
		assert.ErrorIs(t, err, ErrDecodeVerifiableEvent)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}

func newCRECClient(t *testing.T, baseURL string) *apiClient.ClientWithResponses {
	t.Helper()
	crecClient, err := apiClient.NewClientWithResponses(
		baseURL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey "+testAPIKey)
			return nil
		}),
	)
	require.NoError(t, err)
	return crecClient
}

// newTestClient creates a CREC events Client with defaults and allows optional modifications.
func newTestClient(t *testing.T, baseURL string, modify ...func(*Options)) *Client {
	crecClient := newCRECClient(t, baseURL)
	logger := slog.New(slog.DiscardHandler)
	opts := &Options{
		Logger:                logger,
		CRECClient:            crecClient,
		MinRequiredSignatures: 1,
		ValidSigners:          []string{"0x742d35Cc6634C0532925a3b844Bc454e4438f44e"},
	}
	for _, m := range modify {
		m(opts)
	}
	c, err := NewClient(opts)
	require.NoError(t, err)
	return c
}

// setupTestClient spins up a test HTTP server and returns a client bound to it.
func setupTestClient(t *testing.T, handler http.HandlerFunc, modify ...func(*Options)) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	c := newTestClient(t, server.URL, modify...)
	return c, server
}

// setupLocalClient creates a client pointing to a local (non-started) endpoint.
func setupLocalClient(t *testing.T, modify ...func(*Options)) *Client {
	return newTestClient(t, "http://localhost:8080", modify...)
}

// generateTestKeys generates test private keys and returns them with their addresses
func generateTestKeys(t *testing.T, count int) ([]*ecdsa.PrivateKey, []string) {
	t.Helper()
	var keys []*ecdsa.PrivateKey
	var addresses []string

	for i := 0; i < count; i++ {
		privKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		keys = append(keys, privKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey)
		addresses = append(addresses, address.Hex())
	}

	return keys, addresses
}

func createTestEventsWithKeys(t *testing.T, count int, keys []*ecdsa.PrivateKey) []apiClient.Event {
	t.Helper()
	events := make([]apiClient.Event, count)
	for i := range count {
		eventPayload := createTestEventPayload(t)
		events[i] = *createValidEventForOwner(t, keys, &eventPayload, testWorkflowOwner)
	}
	return events
}

func createTestEventPayload(t *testing.T) apiClient.WatcherEventPayload {
	t.Helper()

	// Create verifiable event data (base64 encoded)
	verifiableEventData := map[string]interface{}{
		"from":  "0x0000000000000000000000000000000000000000",
		"to":    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"value": "1000000000000000000",
	}
	verifiableEventBytes, err := json.Marshal(verifiableEventData)
	require.NoError(t, err)
	verifiableEvent := base64.StdEncoding.EncodeToString(verifiableEventBytes)

	// Compute the event hash (keccak256 of verifiable event)
	eventHash := crypto.Keccak256Hash([]byte(verifiableEvent))

	return apiClient.WatcherEventPayload{
		WatcherId:       "550e8400-e29b-41d4-a716-446655440000",
		VerifiableEvent: verifiableEvent,
		EventHash:       eventHash.Hex(),
	}
}

// createTestOperationStatusPayload creates a standard test operation status payload
func createTestOperationStatusPayload(t *testing.T) apiClient.OperationStatusPayload {
	t.Helper()

	verifiableEvent := base64.StdEncoding.EncodeToString([]byte(`{"operationId":"test-op-123","status":"confirmed"}`))
	operationId := uuid.New()

	return apiClient.OperationStatusPayload{
		Address:           "0x1234",
		ChainSelector:     "16015286601757825753",
		OperationId:       operationId,
		WalletOperationId: "wallet-op-123",
		Status:            apiClient.OperationStatusConfirmed,
		StatusReason:      "Operation confirmed",
		Timestamp:         1700000000,
		VerifiableEvent:   &verifiableEvent,
	}
}

func createValidEventForOwner(t *testing.T, privateKeys []*ecdsa.PrivateKey, eventPayload *apiClient.WatcherEventPayload, owner string) *apiClient.Event {
	t.Helper()

	ocrReport := make([]byte, 141)
	ocrReport[0] = 0x01

	workflowOwner := common.HexToAddress(owner)
	copy(ocrReport[87:107], workflowOwner.Bytes())

	eventHash := crypto.Keccak256Hash([]byte(eventPayload.VerifiableEvent))
	copy(ocrReport[109:], eventHash.Bytes())
	ocrContext := []byte("test-context-data")

	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	var signatures []string
	for _, privKey := range privateKeys {
		sig, err := crypto.Sign(reportHash.Bytes(), privKey)
		require.NoError(t, err)
		sig[64] += 27
		signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
	}

	ocrProof := apiClient.OCRProof{
		Alg:        "ecdsa-secp256k1",
		OcrContext: "0x" + common.Bytes2Hex(ocrContext),
		OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
		Signatures: signatures,
	}

	proofUnion := apiClient.EventHeaders_Proofs_Item{}
	err := proofUnion.FromOCRProof(ocrProof)
	require.NoError(t, err)

	payloadUnion := apiClient.Event_Payload{}
	err = payloadUnion.FromWatcherEventPayload(*eventPayload)
	require.NoError(t, err)

	return &apiClient.Event{
		Headers: apiClient.EventHeaders{
			Type:   apiClient.EventTypeWatcherEvent,
			Offset: int64(12345),
			Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
		},
		Payload: payloadUnion,
	}
}

func createValidOperationStatusEventForOwner(t *testing.T, privateKeys []*ecdsa.PrivateKey, eventPayload *apiClient.OperationStatusPayload, owner string) *apiClient.Event {
	t.Helper()

	ocrReport := make([]byte, 141)
	ocrReport[0] = 0x01

	workflowOwner := common.HexToAddress(owner)
	copy(ocrReport[87:107], workflowOwner.Bytes())

	eventHash := crypto.Keccak256Hash([]byte(*eventPayload.VerifiableEvent))

	copy(ocrReport[109:], eventHash.Bytes())
	ocrContext := []byte("test-context-data")

	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	var signatures []string
	for _, privKey := range privateKeys {
		sig, err := crypto.Sign(reportHash.Bytes(), privKey)
		require.NoError(t, err)
		sig[64] += 27
		signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
	}

	ocrProof := apiClient.OCRProof{
		Alg:        "ecdsa-secp256k1",
		OcrContext: "0x" + common.Bytes2Hex(ocrContext),
		OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
		Signatures: signatures,
	}

	proofUnion := apiClient.EventHeaders_Proofs_Item{}
	err := proofUnion.FromOCRProof(ocrProof)
	require.NoError(t, err)

	payloadUnion := apiClient.Event_Payload{}
	err = payloadUnion.FromOperationStatusPayload(*eventPayload)
	require.NoError(t, err)

	return &apiClient.Event{
		Headers: apiClient.EventHeaders{
			Type:   apiClient.EventTypeOperationStatus,
			Offset: int64(12345),
			Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
		},
		Payload: payloadUnion,
	}
}

