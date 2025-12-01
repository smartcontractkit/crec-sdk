package events

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
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
)

const (
	testAPIKey     = "test-api-key"
	testWorkflowID = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
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
		assert.True(t, errors.Is(err, ErrOptionsRequired))
	})

	t.Run("NilCRECClient", func(t *testing.T) {
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{Logger: logger})
		require.Error(t, err)
		assert.Nil(t, c)
		assert.True(t, errors.Is(err, ErrCRECClientRequired))
	})

	t.Run("DefaultLogger", func(t *testing.T) {
		crecClient := newCRECClient(t, "http://localhost:8080")
		c, err := NewClient(&Options{CRECClient: crecClient})
		require.NoError(t, err)
		assert.NotNil(t, c.logger)
	})
}

func TestClient_ListEvents(t *testing.T) {
	// Helper to create test events programmatically
	privKeys, addresses := generateTestKeys(t, 2)

	// Helper to create test events with specific keys
	createTestEventsWithKeys := func(t *testing.T, count int, keys []*ecdsa.PrivateKey) []apiClient.Event {
		t.Helper()
		events := make([]apiClient.Event, count)
		for i := 0; i < count; i++ {
			eventPayload := createTestEventPayload(t)
			event := createValidEventWithSignatures(t, keys, &eventPayload)
			events[i] = *event
		}
		return events
	}

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

		isEventVerified, err := c.Verify(&eventsList[0], testWorkflowID)
		require.NoError(t, err)
		require.True(t, isEventVerified)
	})

	t.Run("WithParams", func(t *testing.T) {
		events := createTestEventsWithKeys(t, 2, privKeys)
		limit := 2
		offset := int64(10)
		domain := "dvp"
		eventName := "Transfer"
		typeVal := apiClient.GetChannelsChannelIdEventsParamsTypeWatcherEvent

		handler := func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "2", q.Get("limit"))
			assert.Equal(t, "10", q.Get("offset"))
			assert.Equal(t, "dvp", q.Get("domain"))
			assert.Equal(t, "Transfer", q.Get("event_name"))
			assert.Equal(t, "watcher.event", q.Get("type"))
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
			Limit:     &limit,
			Offset:    offset,
			Domain:    &domain,
			EventName: &eventName,
			Type:      &typeVal,
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
		assert.True(t, errors.Is(err, ErrChannelIDRequired))
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
		assert.True(t, errors.Is(err, ErrChannelNotFound))
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
		assert.True(t, errors.Is(err, ErrPollEvents))
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode))
	})

	t.Run("NilResponseBody", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		c, server := setupTestClient(t, handler)
		defer server.Close()

		_, _, err := c.Poll(context.Background(), channelID, nil)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNilResponseBody))
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
			Offset: offset,
		}
		eventsList, hasMore, err := c.Poll(context.Background(), channelID, params)
		require.NoError(t, err)
		assert.Len(t, eventsList, 2)
		assert.True(t, hasMore)
	})
}

func TestClient_EventHash(t *testing.T) {
	crecClient := newCRECClient(t, "http://localhost:8080")
	logger := slog.New(slog.DiscardHandler)
	c, err := NewClient(&Options{
		Logger:     logger,
		CRECClient: crecClient,
	})
	require.NoError(t, err)

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

	t.Run("DifferentEventNameProducesDifferentHash", func(t *testing.T) {
		eventPayload1 := createTestEventPayload(t)
		eventPayload2 := createTestEventPayload(t)
		eventPayload2.Event.EventName = "Approval"

		hash1, err := c.EventHash(&eventPayload1)
		require.NoError(t, err)

		hash2, err := c.EventHash(&eventPayload2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "different event names should produce different hashes")
	})

	t.Run("DifferentDomainProducesDifferentHash", func(t *testing.T) {
		eventPayload1 := createTestEventPayload(t)
		eventPayload2 := createTestEventPayload(t)
		differentDomain := "dta"
		eventPayload2.Event.Domain = &differentDomain

		hash1, err := c.EventHash(&eventPayload1)
		require.NoError(t, err)

		hash2, err := c.EventHash(&eventPayload2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "different domains should produce different hashes")
	})

	t.Run("DifferentDataProducesDifferentHash", func(t *testing.T) {
		eventPayload1 := createTestEventPayload(t)
		eventPayload2 := createTestEventPayload(t)
		eventPayload2.Event.Data = map[string]interface{}{
			"from":  "0x0000000000000000000000000000000000000000",
			"to":    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			"value": "2000000000000000000", // Different value
		}

		hash1, err := c.EventHash(&eventPayload1)
		require.NoError(t, err)

		hash2, err := c.EventHash(&eventPayload2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "different event data should produce different hashes")
	})

	t.Run("VerifyHashFormat", func(t *testing.T) {
		eventPayload := createTestEventPayload(t)

		hash, err := c.EventHash(&eventPayload)
		require.NoError(t, err)

		// Manually compute the expected hash to verify the algorithm
		dataBytes, err := json.Marshal(eventPayload.Event.Data)
		require.NoError(t, err)
		dataStr := base64.StdEncoding.EncodeToString(dataBytes)
		expectedHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))

		assert.Equal(t, expectedHash, hash, "hash should match expected Keccak256 computation")
	})

	t.Run("ErrEventDomainIsNil", func(t *testing.T) {
		eventPayload := createTestEventPayload(t)
		eventPayload.Event.Domain = nil // Set domain to nil

		hash, err := c.EventHash(&eventPayload)
		require.Error(t, err)
		assert.Equal(t, common.Hash{}, hash, "hash should be empty when domain is nil")
		assert.True(t, errors.Is(err, ErrEventDomainIsNil))
	})
}

func TestClient_Verify(t *testing.T) {
	t.Run("ErrVerificationNotConfigured", func(t *testing.T) {
		// Create client WITHOUT configuring signers
		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 0,
			ValidSigners:          nil, // No signers configured
		})
		require.NoError(t, err)

		// Create a valid event
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys, &eventPayload)

		// Verify should fail because no signers are configured
		ok, err := c.Verify(event)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrVerificationNotConfigured))
	})

	t.Run("ErrVerificationNotConfigured_EmptySigners", func(t *testing.T) {
		// Create client with empty signers slice
		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 0,
			ValidSigners:          []string{}, // Empty signers slice
		})
		require.NoError(t, err)

		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys, &eventPayload)

		ok, err := c.Verify(event)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrVerificationNotConfigured))
	})

	t.Run("HappyPath_TwoValidSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys, &eventPayload)

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 2,
			ValidSigners:          addresses,
		})
		require.NoError(t, err)

		ok, err := c.Verify(event, testWorkflowID)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("NotEnoughSignatures", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 3)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys[:2], &eventPayload)

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 3,
			ValidSigners:          addresses,
		})
		require.NoError(t, err)

		ok, err := c.Verify(event, testWorkflowID)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("IncorrectSigners", func(t *testing.T) {
		signingKeys, _ := generateTestKeys(t, 2)
		_, validAddresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, signingKeys, &eventPayload)

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 2,
			ValidSigners:          validAddresses,
		})
		require.NoError(t, err)

		ok, err := c.Verify(event, testWorkflowID)
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{},
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrVerifyEvent))
		assert.True(t, errors.Is(err, ErrNoOCRProofs))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrOCRReportTooShort))
	})

	t.Run("ErrInvalidWorkflowCID", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		wrongWorkflowCid := common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		copy(ocrReport[45:77], wrongWorkflowCid.Bytes())

		dataBytes, _ := json.Marshal(eventPayload.Event.Data)
		dataStr := base64.StdEncoding.EncodeToString(dataBytes)
		eventHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := zerolog.Nop()
		c, err := NewClient(&ClientOptions{
			Logger:                &logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 1,
			ValidSigners:          addresses,
		})

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrInvalidEventHash))
	})

	t.Run("ErrInvalidEventHash", func(t *testing.T) {
		privKeys, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		ocrReport := make([]byte, 141)
		ocrReport[0] = 0x01

		workflowCid := common.HexToHash(testWorkflowID)
		copy(ocrReport[45:77], workflowCid.Bytes())

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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 1,
			ValidSigners:          addresses,
		})

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrInvalidEventHash))
	})

	t.Run("ErrMultipleOCRProofs", func(t *testing.T) {
		c := setupLocalClient(t)
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create two identical OCR proofs
		ocrReport := make([]byte, 141)

		// Place workflow_cid at offset 45 (32 bytes)
		workflowCid := common.HexToHash(testWorkflowID)
		copy(ocrReport[45:77], workflowCid.Bytes())

		ocrContext := []byte("test")
		dataBytes, _ := json.Marshal(eventPayload.Event.Data)
		dataStr := base64.StdEncoding.EncodeToString(dataBytes)
		eventHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proof1, proof2}, // Multiple proofs
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrVerifyEvent))
		assert.True(t, errors.Is(err, ErrMultipleOCRProofs))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrParseOCRReport))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrParseOCRContext))
	})

	t.Run("ErrParseSignature", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create valid OCR report
		ocrReport := make([]byte, 141)

		workflowCid := common.HexToHash(testWorkflowID)
		copy(ocrReport[45:77], workflowCid.Bytes())

		dataBytes, _ := json.Marshal(eventPayload.Event.Data)
		dataStr := base64.StdEncoding.EncodeToString(dataBytes)
		eventHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 1,
			ValidSigners:          addresses,
		})

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrParseSignature))
	})

	t.Run("ErrRecoverPubKeyFromSignature", func(t *testing.T) {
		_, addresses := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)

		// Create valid OCR report
		ocrReport := make([]byte, 141)

		workflowCid := common.HexToHash(testWorkflowID)
		copy(ocrReport[45:77], workflowCid.Bytes())

		dataBytes, _ := json.Marshal(eventPayload.Event.Data)
		dataStr := base64.StdEncoding.EncodeToString(dataBytes)
		eventHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))
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
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		crecClient := newCRECClient(t, "http://localhost:8080")
		logger := slog.New(slog.DiscardHandler)
		c, err := NewClient(&Options{
			Logger:                logger,
			CRECClient:            crecClient,
			MinRequiredSignatures: 1,
			ValidSigners:          addresses,
		})

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrRecoverPubKeyFromSignature))
	})

	// Test payload parsing errors with watcher.status type (not watcher.event)
	t.Run("ErrParseEventPayload_OnlyWatcherEventsSupported", func(t *testing.T) {
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
			Type:          apiClient.WatcherStatus,
			WatcherId:     "550e8400-e29b-41d4-a716-446655440000",
			ChainSelector: "5009297550715157269",
			Status:        apiClient.WatcherStatusPayloadStatusDeploying,
			StatusCode:    "DEPLOYING",
			StatusReason:  "Watcher is being deployed",
		}

		payloadUnion := apiClient.Event_Payload{}
		require.NoError(t, payloadUnion.FromWatcherStatusPayload(statusPayload))

		event := &apiClient.Event{
			Headers: apiClient.EventHeaders{
				Offset: int64(12345),
				Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
			},
			Payload: payloadUnion,
		}

		ok, err := c.Verify(event, testWorkflowID)
		require.Error(t, err)
		assert.False(t, ok)
		assert.True(t, errors.Is(err, ErrOnlyWatcherEventsSupported))
	})
}

func TestClient_Decode(t *testing.T) {
	crecClient := newCRECClient(t, "http://localhost:8080")
	logger := slog.New(slog.DiscardHandler)
	c, err := NewClient(&Options{
		Logger:     logger,
		CRECClient: crecClient,
	})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys, &eventPayload)

		var decoded map[string]interface{}
		err := c.Decode(event, &decoded)
		require.NoError(t, err)
		assert.NotNil(t, decoded)
	})
}

func TestClient_ToJson(t *testing.T) {
	crecClient := newCRECClient(t, "http://localhost:8080")
	logger := slog.New(slog.DiscardHandler)
	c, err := NewClient(&Options{
		Logger:     logger,
		CRECClient: crecClient,
	})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		privKeys, _ := generateTestKeys(t, 2)
		eventPayload := createTestEventPayload(t)
		event := createValidEventWithSignatures(t, privKeys, &eventPayload)

		jsonBytes, err := c.ToJSON(*event)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonBytes)

		// Verify it's valid JSON
		var decoded map[string]interface{}
		err = json.Unmarshal(jsonBytes, &decoded)
		require.NoError(t, err)
	})
}

func newCRECClient(t *testing.T, baseURL string) *apiClient.ClientWithResponses {
	t.Helper()
	crecClient, err := apiClient.NewClientWithResponses(
		baseURL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Api-Key", testAPIKey)
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

// mustParseTime parses time string or fails the test
func mustParseTime(t *testing.T, timeStr string) time.Time {
	t.Helper()
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsedTime
}

// createTestEventPayload creates a standard test event payload
func createTestEventPayload(t *testing.T) apiClient.WatcherEventPayload {
	t.Helper()

	domain := "dvp"
	metadata := map[string]interface{}{
		"block_number": 12345678,
		"block_hash":   "0xabcdef1234567890",
	}

	return apiClient.WatcherEventPayload{
		Type:          apiClient.WatcherEventPayloadType("watcher.event"),
		WatcherId:     "550e8400-e29b-41d4-a716-446655440000",
		Address:       "0x1234567890123456789012345678901234567890",
		ChainSelector: "5009297550715157269",
		Event: apiClient.WatcherEvent{
			EventName: "Transfer",
			TopicHash: "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
			Timestamp: mustParseTime(t, "2024-01-01T00:00:00Z"),
			LogIndex:  42,
			Domain:    &domain,
			Data: map[string]interface{}{
				"from":  "0x0000000000000000000000000000000000000000",
				"to":    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
				"value": "1000000000000000000",
			},
			Metadata: &metadata,
		},
	}
}

// createValidEventWithSignatures creates a valid event with proper OCR report and signatures
func createValidEventWithSignatures(t *testing.T, privateKeys []*ecdsa.PrivateKey, eventPayload *apiClient.WatcherEventPayload) *apiClient.Event {
	t.Helper()

	// Create a valid OCR report with the proper structure (141 bytes minimum)
	ocrReport := make([]byte, 141)
	ocrReport[0] = 0x01 // version

	workflowCid := common.HexToHash(testWorkflowID)
	copy(ocrReport[45:77], workflowCid.Bytes())

	// Compute event hash using base64 encoding (same as EventHash method)
	dataBytes, err := json.Marshal(eventPayload.Event.Data)
	require.NoError(t, err)
	dataStr := base64.StdEncoding.EncodeToString(dataBytes)
	eventHash := crypto.Keccak256Hash([]byte(*eventPayload.Event.Domain + "." + eventPayload.Event.EventName + "." + dataStr))

	// Place event hash at offset 109
	copy(ocrReport[109:], eventHash.Bytes())

	ocrContext := []byte("test-context-data")

	// Generate report hash for signing
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	// Generate signatures
	var signatures []string
	for _, privKey := range privateKeys {
		sig, err := crypto.Sign(reportHash.Bytes(), privKey)
		require.NoError(t, err)
		sig[64] += 27 // Adjust v value for Ethereum format
		signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
	}

	// Create OCR proof
	ocrProof := apiClient.OCRProof{
		Alg:        "ecdsa-secp256k1",
		OcrContext: "0x" + common.Bytes2Hex(ocrContext),
		OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
		Signatures: signatures,
	}

	proofUnion := apiClient.EventHeaders_Proofs_Item{}
	err = proofUnion.FromOCRProof(ocrProof)
	require.NoError(t, err)

	// Create event payload union
	payloadUnion := apiClient.Event_Payload{}
	err = payloadUnion.FromWatcherEventPayload(*eventPayload)
	require.NoError(t, err)

	event := &apiClient.Event{
		Headers: apiClient.EventHeaders{
			Offset: int64(12345),
			Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
		},
		Payload: payloadUnion,
	}

	return event
}
