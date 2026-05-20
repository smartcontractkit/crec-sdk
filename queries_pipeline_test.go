package crec_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"

	"github.com/smartcontractkit/crec-sdk"
	eventsPkg "github.com/smartcontractkit/crec-sdk/events"
	"github.com/smartcontractkit/crec-sdk/queries"
)

const queryPipelineZeroAddress = "0x0000000000000000000000000000000000000000"

func TestClient_QueriesCompletePipelineWithVerifiedStatusEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	orgID := "query-pipeline-org"
	channelID := uuid.New()
	queryID := uuid.New()
	eventID := uuid.New()

	chainSelector := "16015286601757825753"
	token := "0x1234567890123456789012345678901234567890"
	idempotencyKey := "total-supply-finalized-demo"
	totalSupply, ok := new(big.Int).SetString("1000000000000000000000000", 10)
	require.True(t, ok)

	rawReturnData := queryPipelineUint256Hex(totalSupply)
	verifiableResult := queryPipelineVerifiableResult(t, channelID, queryID, chainSelector, token, rawReturnData)

	privateKeys, validSigners := queryPipelineGenerateKeys(t, 3)
	workflowOwner := queryPipelineWorkflowOwner(t, orgID)
	proof, eventHash := queryPipelineOCRProof(t, privateKeys[:2], workflowOwner, verifiableResult)
	queryStatusEvent := queryPipelineStatusEvent(t, eventID, queryID, verifiableResult, proof)
	terminalQuery := queryPipelineCompletedQuery(channelID, queryID, chainSelector, verifiableResult, eventHash, proof)

	var createQueryCount int
	var getQueryCount int
	var searchEventsCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

		switch {
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/channels/%s/queries", channelID):
			createQueryCount++

			var req apiClient.CreateQuery
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, idempotencyKey, req.IdempotencyKey)
			assert.Equal(t, apiClient.QueryKindEVMCall, req.QueryKind)
			assert.Equal(t, chainSelector, string(req.ChainSelector))
			assert.Equal(t, strings.ToLower(token), string(req.Params.ContractAddress))
			assert.Equal(t, "0x18160ddd", req.Params.CallData)
			require.NotNil(t, req.Params.FromAddress)
			assert.Equal(t, queryPipelineZeroAddress, string(*req.Params.FromAddress))

			discriminator, err := req.Params.BlockSelection.Discriminator()
			require.NoError(t, err)
			assert.Equal(t, "finalized", discriminator)

			queryPipelineWriteJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
				QueryId: queryID,
				Status:  apiClient.QueryStatusAccepted,
			})

		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/channels/%s/queries/%s", channelID, queryID):
			getQueryCount++
			queryPipelineWriteJSON(t, w, http.StatusOK, terminalQuery)

		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/channels/%s/events/search", channelID):
			searchEventsCount++

			q := r.URL.Query()
			assert.Equal(t, "100", q.Get("limit"))
			assert.Equal(t, "0", q.Get("offset"))
			assert.Equal(t, string(apiClient.EventTypeQueryStatus), q.Get("type"))

			queryPipelineWriteJSON(t, w, http.StatusOK, apiClient.EventList{
				Events:  []apiClient.Event{*queryStatusEvent},
				HasMore: false,
			})

		default:
			http.Error(w, fmt.Sprintf("unexpected request %s %s", r.Method, r.URL.Path), http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := crec.NewClient(
		server.URL,
		"test-api-key",
		crec.WithOrgID(orgID),
		crec.WithEventVerification(2, validSigners),
	)
	require.NoError(t, err)

	accepted, err := client.Queries.CreateEVMCall(
		ctx,
		queries.CallContractInput{
			ChannelID:       channelID,
			ChainSelector:   chainSelector,
			ContractAddress: token,
			CallData:        []byte{0x18, 0x16, 0x0d, 0xdd},
			BlockSelection:  queries.Finalized(),
			IdempotencyKey:  idempotencyKey,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, accepted)
	assert.Equal(t, queryID, accepted.QueryId)

	query, err := client.Queries.Wait(ctx, channelID, accepted.QueryId)
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, apiClient.QueryStatusCompleted, query.Status)

	result, err := client.Queries.ResultFromQuery(query)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Nil(t, result.Error)
	assert.Equal(t, verifiableResult, result.VerifiableResult)
	assert.Equal(t, eventHash, result.EventHash)
	assert.Equal(t, strings.ToLower(token), result.Target)

	queryStatusEventFromAPI, err := findQueryStatusEventForPipelineTest(ctx, client, channelID, accepted.QueryId)
	require.NoError(t, err)
	require.NotNil(t, queryStatusEventFromAPI)

	verified, err := client.Events.VerifyQueryStatus(queryStatusEventFromAPI)
	require.NoError(t, err)
	assert.True(t, verified, "query.status event should have enough valid DON signatures")

	queryStatusPayload, err := queryStatusEventFromAPI.Payload.AsQueryStatusPayload()
	require.NoError(t, err)
	require.NotNil(t, queryStatusPayload.VerifiableResult)
	assert.Equal(t, result.VerifiableResult, *queryStatusPayload.VerifiableResult)

	decodedQueryEvent, err := client.Events.DecodeQueryStatusVerifiableEvent(&queryStatusPayload)
	require.NoError(t, err)
	require.NotNil(t, decodedQueryEvent)
	assert.Equal(t, accepted.QueryId, decodedQueryEvent.Data.QueryId)
	assert.Equal(t, channelID, decodedQueryEvent.Data.ChannelId)
	assert.Equal(t, chainSelector, decodedQueryEvent.ChainSelector)

	decodedTotalSupply := new(big.Int).SetBytes(result.RawReturnData)
	assert.Equal(t, 0, decodedTotalSupply.Cmp(totalSupply), "decoded totalSupply should match mocked return data")

	assert.Equal(t, 1, createQueryCount)
	assert.Equal(t, 1, getQueryCount)
	assert.Equal(t, 1, searchEventsCount)
}

func findQueryStatusEventForPipelineTest(
	ctx context.Context,
	sdkClient *crec.Client,
	channelID uuid.UUID,
	queryID uuid.UUID,
) (*apiClient.Event, error) {
	limit := 100
	eventTypes := []apiClient.EventType{apiClient.EventTypeQueryStatus}

	for offset := int64(0); ; offset += int64(limit) {
		channelEvents, hasMore, err := sdkClient.Events.SearchEvents(
			ctx,
			channelID,
			&apiClient.SearchChannelEventsParams{
				Limit:  &limit,
				Offset: &offset,
				Type:   &eventTypes,
			},
		)
		if err != nil {
			return nil, err
		}

		for i := range channelEvents {
			if channelEvents[i].Headers.Type != apiClient.EventTypeQueryStatus {
				continue
			}

			payload, err := channelEvents[i].Payload.AsQueryStatusPayload()
			if err != nil {
				return nil, fmt.Errorf("decode query.status payload: %w", err)
			}
			if payload.QueryId == queryID {
				return &channelEvents[i], nil
			}
		}

		if !hasMore {
			break
		}
	}

	return nil, fmt.Errorf("query.status event for query %s not found", queryID)
}

func queryPipelineVerifiableResult(
	t *testing.T,
	channelID uuid.UUID,
	queryID uuid.UUID,
	chainSelector string,
	contractAddress string,
	rawReturnData string,
) string {
	t.Helper()

	var requested models.ChainQueryRequestedBlockSelection
	require.NoError(t, requested.FromChainQueryFinalizedBlockSelection(models.ChainQueryFinalizedBlockSelection{}))

	event := models.ChainQueryVerifiableEvent{
		Service:       models.ChainQueryVerifiableEventServiceCREC,
		Name:          models.ChainQueryVerifiableEventNameChainQuery,
		ChainSelector: chainSelector,
		Timestamp:     time.Date(2026, time.May, 14, 12, 0, 0, 0, time.UTC),
		Data: models.ChainQueryData{
			QueryId:   queryID,
			ChannelId: channelID,
			QueryKind: models.ChainQueryKindEVMCall,
			Target: models.ChainQueryTarget{
				FromAddress:     queryPipelineZeroAddress,
				ContractAddress: strings.ToLower(contractAddress),
				CallData:        "0x18160ddd",
			},
			BlockSelection: models.ChainQueryBlockSelection{
				Requested: requested,
				Resolved: &models.ChainQueryResolvedBlock{
					BlockNumber:    "22446688",
					BlockHash:      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					BlockTimestamp: 1778760000,
				},
			},
			Result: &models.ChainQueryExecutionResult{
				RawReturnData: rawReturnData,
			},
		},
	}

	encoded, err := json.Marshal(event)
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(encoded)
}

func queryPipelineOCRProof(
	t *testing.T,
	privateKeys []*ecdsa.PrivateKey,
	workflowOwner string,
	verifiableResult string,
) (apiClient.OCRProof, string) {
	t.Helper()

	eventHash := crypto.Keccak256Hash([]byte(verifiableResult))

	ocrReport := make([]byte, 141)
	ocrReport[0] = 0x01
	copy(ocrReport[87:107], common.HexToAddress(workflowOwner).Bytes())
	copy(ocrReport[109:], eventHash.Bytes())

	ocrContext := []byte("query-status-pipeline-context")
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	signatures := make([]string, 0, len(privateKeys))
	for _, privateKey := range privateKeys {
		sig, err := crypto.Sign(reportHash.Bytes(), privateKey)
		require.NoError(t, err)
		sig[64] += 27
		signatures = append(signatures, "0x"+common.Bytes2Hex(sig))
	}

	return apiClient.OCRProof{
		Alg:        "ecdsa-secp256k1",
		OcrContext: "0x" + common.Bytes2Hex(ocrContext),
		OcrReport:  "0x" + common.Bytes2Hex(ocrReport),
		Signatures: signatures,
	}, eventHash.Hex()
}

func queryPipelineStatusEvent(
	t *testing.T,
	eventID uuid.UUID,
	queryID uuid.UUID,
	verifiableResult string,
	proof apiClient.OCRProof,
) *apiClient.Event {
	t.Helper()

	proofUnion := apiClient.EventHeaders_Proofs_Item{}
	require.NoError(t, proofUnion.FromOCRProof(proof))

	payload := apiClient.QueryStatusPayload{
		QueryId:          queryID,
		Status:           apiClient.QueryStatusCompleted,
		VerifiableResult: &verifiableResult,
	}
	payloadUnion := apiClient.Event_Payload{}
	require.NoError(t, payloadUnion.FromQueryStatusPayload(payload))

	return &apiClient.Event{
		EventId: &eventID,
		Headers: apiClient.EventHeaders{
			Type:   apiClient.EventTypeQueryStatus,
			Offset: 42,
			Proofs: []apiClient.EventHeaders_Proofs_Item{proofUnion},
		},
		Payload: payloadUnion,
	}
}

func queryPipelineCompletedQuery(
	channelID uuid.UUID,
	queryID uuid.UUID,
	chainSelector string,
	verifiableResult string,
	eventHash string,
	proof apiClient.OCRProof,
) *apiClient.Query {
	completedAt := apiClient.Timestamp(1778760001)

	return &apiClient.Query{
		QueryId:          queryID,
		ChannelId:        channelID,
		Status:           apiClient.QueryStatusCompleted,
		QueryKind:        apiClient.QueryKindEVMCall,
		ChainSelector:    apiClient.ChainSelector(chainSelector),
		EventHash:        &eventHash,
		VerifiableResult: &verifiableResult,
		Proof:            &proof,
		CreatedAt:        1778759900,
		UpdatedAt:        1778760001,
		CompletedAt:      &completedAt,
	}
}

func queryPipelineGenerateKeys(t *testing.T, count int) ([]*ecdsa.PrivateKey, []string) {
	t.Helper()

	privateKeys := make([]*ecdsa.PrivateKey, 0, count)
	addresses := make([]string, 0, count)

	for i := 0; i < count; i++ {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		privateKeys = append(privateKeys, privateKey)
		addresses = append(addresses, crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	}

	return privateKeys, addresses
}

func queryPipelineWorkflowOwner(t *testing.T, orgID string) string {
	t.Helper()

	ownerBytes, err := workflows.GenerateWorkflowOwnerAddress(eventsPkg.CreMainlineTenantID, orgID)
	require.NoError(t, err)

	return common.BytesToAddress(ownerBytes).Hex()
}

func queryPipelineUint256Hex(value *big.Int) string {
	hexValue := value.Text(16)
	if len(hexValue) < 64 {
		hexValue = strings.Repeat("0", 64-len(hexValue)) + hexValue
	}
	return "0x" + hexValue
}

func queryPipelineWriteJSON(t *testing.T, w http.ResponseWriter, statusCode int, body any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}
}
