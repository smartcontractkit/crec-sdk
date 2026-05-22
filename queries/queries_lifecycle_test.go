package queries

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"
)

func TestClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		expected := makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusSent)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String()+"/queries/"+queryID.String(), r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			writeJSON(t, w, http.StatusOK, expected)
		}

		client, server := setupQueriesTestClient(t, handler)
		defer server.Close()

		query, err := client.Get(context.Background(), channelID, queryID)

		require.NoError(t, err)
		require.NotNil(t, query)
		assert.Equal(t, queryID, query.QueryId)
		assert.Equal(t, apiClient.QueryStatusSent, query.Status)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("request should not be sent for validation failures")
		})
		defer server.Close()

		query, err := client.Get(context.Background(), uuid.Nil, uuid.New())
		require.Error(t, err)
		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrChannelIDRequired)

		query, err = client.Get(context.Background(), uuid.New(), uuid.Nil)
		require.Error(t, err)
		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryIDRequired)
	})

	t.Run("NotFound", func(t *testing.T) {
		client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(t, w, http.StatusNotFound, nil)
		})
		defer server.Close()

		query, err := client.Get(context.Background(), uuid.New(), uuid.New())

		require.Error(t, err)
		assert.Nil(t, query)
		assert.ErrorIs(t, err, ErrQueryNotFound)
	})
}

func TestClient_List(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/channels/"+channelID.String()+"/queries", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		q := r.URL.Query()
		assert.ElementsMatch(t, []string{"completed", "failed"}, q["status"])
		assert.Equal(t, "10", q.Get("limit"))
		assert.Equal(t, "5", q.Get("offset"))

		writeJSON(t, w, http.StatusOK, apiClient.QueryList{
			Data: []apiClient.Query{
				makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusCompleted),
			},
			HasMore: true,
		})
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	statuses := []apiClient.QueryStatus{apiClient.QueryStatusCompleted, apiClient.QueryStatusFailed}
	limit := 10
	offset := int64(5)

	queries, hasMore, err := client.List(context.Background(), ListInput{
		ChannelID: channelID,
		Status:    &statuses,
		Limit:     &limit,
		Offset:    &offset,
	})

	require.NoError(t, err)
	assert.True(t, hasMore)
	require.Len(t, queries, 1)
	assert.Equal(t, queryID, queries[0].QueryId)
}

func TestClient_List_ValidationAndErrors(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("request should not be sent for validation failures")
		})
		defer server.Close()

		_, _, err := client.List(context.Background(), ListInput{ChannelID: uuid.Nil})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrChannelIDRequired)

		limit := 101
		_, _, err = client.List(context.Background(), ListInput{ChannelID: uuid.New(), Limit: &limit})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidLimit)

		offset := int64(-1)
		_, _, err = client.List(context.Background(), ListInput{ChannelID: uuid.New(), Offset: &offset})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidOffset)
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(t, w, http.StatusNotFound, nil)
		})
		defer server.Close()

		queries, hasMore, err := client.List(context.Background(), ListInput{ChannelID: uuid.New()})

		require.Error(t, err)
		assert.Nil(t, queries)
		assert.False(t, hasMore)
		assert.ErrorIs(t, err, ErrChannelNotFound)
	})
}

func TestClient_Wait(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()
	callCount := 0

	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++

		status := apiClient.QueryStatusAccepted
		if callCount == 2 {
			status = apiClient.QueryStatusSent
		}
		if callCount >= 3 {
			status = apiClient.QueryStatusCompleted
		}

		writeJSON(t, w, http.StatusOK, makeAcceptedQuery(channelID, queryID, status))
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	query, err := client.Wait(ctx, channelID, queryID)

	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, apiClient.QueryStatusCompleted, query.Status)
	assert.GreaterOrEqual(t, callCount, 3)
}

func TestClient_Wait_ContextDeadline(t *testing.T) {
	client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, makeAcceptedQuery(uuid.New(), uuid.New(), apiClient.QueryStatusSent))
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	query, err := client.Wait(ctx, uuid.New(), uuid.New())

	require.Error(t, err)
	assert.Nil(t, query)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected context deadline exceeded, got %v", err)
}

func TestClient_CallContract(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()
	rawReturnData := "0x00000000000000000000000000000000000000000000000000000000000003e8"
	getCount := 0

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			assert.Equal(t, "/channels/"+channelID.String()+"/queries", r.URL.Path)

			var req apiClient.CreateQuery
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "call-contract-1", req.IdempotencyKey)
			assert.Equal(t, "0x18160ddd", req.Params.CallData)
			require.NotNil(t, req.Params.FromAddress)
			assert.Equal(t, "0x000000000000000000000000000000000000dead", string(*req.Params.FromAddress))

			writeJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
				QueryId: queryID,
				Status:  apiClient.QueryStatusAccepted,
			})
		case http.MethodGet:
			getCount++
			if getCount == 1 {
				writeJSON(t, w, http.StatusOK, makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusSent))
				return
			}
			writeJSON(t, w, http.StatusOK, makeCompletedQuery(t, channelID, queryID, rawReturnData))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := client.CallContract(
		ctx,
		channelID,
		testChainSelector,
		testContractAddress,
		[]byte{0x18, 0x16, 0x0d, 0xdd},
		Latest(),
		"call-contract-1",
		WithFromAddress(testFromAddress),
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, queryID.String(), result.QueryID)
	assert.Equal(t, string(apiClient.QueryStatusCompleted), result.Status)
	assert.Equal(t, testChainSelector, result.ChainSelector)
	assert.Equal(t, "0x5c7a", result.EventHash)
	assert.Equal(t, []byte{0x03, 0xe8}, bytesTrimLeftZeroes(result.RawReturnData))
	require.NotNil(t, result.Block)
	assert.Equal(t, "12345678", result.Block.BlockNumber)
	assert.Equal(t, int64(1777334412), result.Block.BlockTimestamp)
	assert.Equal(t, "ocr", result.Proof.Alg)
	assert.Nil(t, result.Error)
}

func TestClient_CreateEVMCall_Wait_ResultFromQuery(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()
	rawReturnData := "0x" + fmt.Sprintf("%064x", big.NewInt(1000))
	postSeen := false
	getCount := 0

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			assert.Equal(t, "/channels/"+channelID.String()+"/queries", r.URL.Path)

			var req apiClient.CreateQuery
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "async-read-1", req.IdempotencyKey)
			assert.Equal(t, apiClient.QueryKindEVMCall, req.QueryKind)
			assert.Equal(t, testChainSelector, string(req.ChainSelector))
			assert.Equal(t, testContractAddress, string(req.Params.ContractAddress))
			assert.Equal(t, "0x18160ddd", req.Params.CallData)

			discriminator, err := req.Params.BlockSelection.Discriminator()
			require.NoError(t, err)
			assert.Equal(t, "finalized", discriminator)

			postSeen = true
			writeJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
				QueryId: queryID,
				Status:  apiClient.QueryStatusAccepted,
			})
		case http.MethodGet:
			assert.True(t, postSeen, "query must be created before waiting for it")
			assert.Equal(t, "/channels/"+channelID.String()+"/queries/"+queryID.String(), r.URL.Path)

			getCount++
			if getCount == 1 {
				writeJSON(t, w, http.StatusOK, makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusSent))
				return
			}
			writeJSON(t, w, http.StatusOK, makeCompletedQuery(t, channelID, queryID, rawReturnData))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	accepted, err := client.CreateEVMCall(
		ctx,
		CallContractInput{
			ChannelID:       channelID,
			ChainSelector:   testChainSelector,
			ContractAddress: testContractAddress,
			CallData:        []byte{0x18, 0x16, 0x0d, 0xdd},
			BlockSelection:  Finalized(),
			IdempotencyKey:  "async-read-1",
		},
	)
	require.NoError(t, err)
	require.NotNil(t, accepted)
	assert.Equal(t, queryID, accepted.QueryId)

	query, err := client.Wait(ctx, channelID, accepted.QueryId)
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, apiClient.QueryStatusCompleted, query.Status)

	result, err := ResultFromQuery(query)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, queryID.String(), result.QueryID)
	assert.Equal(t, string(apiClient.QueryStatusCompleted), result.Status)
	assert.Equal(t, testChainSelector, result.ChainSelector)
	assert.Equal(t, big.NewInt(1000), new(big.Int).SetBytes(result.RawReturnData))
	assert.GreaterOrEqual(t, getCount, 2)
	assert.Nil(t, result.Error)
}

func TestClient_CallContractWithABI(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()
	rawReturnData := "0x" + fmt.Sprintf("%064x", big.NewInt(1000))

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req apiClient.CreateQuery
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, "abi-call-1", req.IdempotencyKey)
			assert.Equal(t, "0x18160ddd", req.Params.CallData)

			writeJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
				QueryId: queryID,
				Status:  apiClient.QueryStatusAccepted,
			})
		case http.MethodGet:
			writeJSON(t, w, http.StatusOK, makeCompletedQuery(t, channelID, queryID, rawReturnData))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := client.CallContractWithABI(
		ctx,
		channelID,
		testChainSelector,
		testContractAddress,
		"function totalSupply() view returns (uint256)",
		"totalSupply",
		nil,
		Latest(),
		"abi-call-1",
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Outputs, 1)

	output, ok := result.Outputs[0].(*big.Int)
	require.True(t, ok, "expected ABI output to be *big.Int, got %T", result.Outputs[0])
	assert.Equal(t, big.NewInt(1000), output)
}

func TestClient_CallContractWithABI_Validation(t *testing.T) {
	client, server := setupQueriesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be sent for ABI validation failures")
	})
	defer server.Close()

	_, err := client.CallContractWithABI(
		context.Background(),
		uuid.New(),
		testChainSelector,
		testContractAddress,
		"",
		"",
		nil,
		Latest(),
		"abi-validation-1",
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrABIRequired)

	_, err = client.CallContractWithABI(
		context.Background(),
		uuid.New(),
		testChainSelector,
		testContractAddress,
		"function balanceOf(address) view returns (uint256)",
		"balanceOf",
		nil,
		Latest(),
		"abi-validation-2",
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrABIArgumentCount)

	_, err = client.CallContractWithABI(
		context.Background(),
		uuid.New(),
		testChainSelector,
		testContractAddress,
		"function balanceOf(address) view returns (uint256)",
		"balanceOf",
		[]any{"not-an-address"},
		Latest(),
		"abi-validation-3",
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrABIArgumentType)
}

func TestResultFromQuery(t *testing.T) {
	t.Run("CompletedDecodesVerifiableResult", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		rawReturnData := "0x00000000000000000000000000000000000000000000000000000000000003e8"
		query := makeCompletedQuery(t, channelID, queryID, rawReturnData)

		result, err := ResultFromQuery(query)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, queryID.String(), result.QueryID)
		assert.Equal(t, string(apiClient.QueryStatusCompleted), result.Status)
		assert.Equal(t, testChainSelector, result.ChainSelector)
		assert.Equal(t, "0x5c7a", result.EventHash)
		assert.Equal(t, rawReturnData, "0x"+hex.EncodeToString(result.RawReturnData))
		require.NotNil(t, result.Block)
		assert.Equal(t, "12345678", result.Block.BlockNumber)
		assert.NotEmpty(t, result.VerifiableQuery)
		assert.NotEmpty(t, result.VerifiableResult)
		assert.Nil(t, result.Error)
	})

	t.Run("CompletedRequiresVerifiableResult", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		query := makeCompletedQuery(t, channelID, queryID, "0x")
		query.VerifiableResult = nil

		result, err := ResultFromQuery(query)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrDecodeVerifiableResult)
		assert.ErrorIs(t, err, ErrVerifiableResultRequired)
	})

	t.Run("FailedWithSignedError", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		rawRevertData := "0x08c379a0"
		verifiableResult := makeVerifiableResult(t, channelID, queryID, nil, &models.ChainQueryError{
			Code:          models.ChainQueryErrorCodeCallReverted,
			Message:       "execution reverted",
			RawRevertData: &rawRevertData,
		})

		query := makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusFailed)
		query.VerifiableResult = &verifiableResult

		result, err := ResultFromQuery(&query)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Error)
		assert.Equal(t, "CALL_REVERTED", result.Error.Code)
		assert.Equal(t, "execution reverted", result.Error.Message)
		assert.Equal(t, rawRevertData, result.Error.RawRevertDataHex)
		assert.Equal(t, []byte{0x08, 0xc3, 0x79, 0xa0}, result.Error.RawRevertData)
	})

	t.Run("FailedWithoutVerifiableResultUsesDefaultError", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()

		query := makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusFailed)

		result, err := ResultFromQuery(&query)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Error)
		assert.Equal(t, "CRE_WORKFLOW_FAILED", result.Error.Code)
	})

	t.Run("ExpiredUsesDefaultError", func(t *testing.T) {
		query := makeAcceptedQuery(uuid.New(), uuid.New(), apiClient.QueryStatusExpired)

		result, err := ResultFromQuery(&query)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Error)
		assert.Equal(t, "QUERY_EXPIRED", result.Error.Code)
	})

	t.Run("NilQuery", func(t *testing.T) {
		result, err := ResultFromQuery(nil)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrQueryRequired)
	})
}

func TestDecodeVerifiableResult(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		rawReturnData := "0x"
		encoded := makeVerifiableResult(t, channelID, queryID, &rawReturnData, nil)

		event, err := DecodeVerifiableResult(encoded)

		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, testChainSelector, event.ChainSelector)
		assert.Equal(t, queryID, event.Data.QueryId)
		assert.Equal(t, channelID, event.Data.ChannelId)
	})

	t.Run("Empty", func(t *testing.T) {
		event, err := DecodeVerifiableResult("")

		require.Error(t, err)
		assert.Nil(t, event)
		assert.ErrorIs(t, err, ErrDecodeVerifiableResult)
		assert.ErrorIs(t, err, ErrVerifiableResultRequired)
	})

	t.Run("InvalidBase64", func(t *testing.T) {
		event, err := DecodeVerifiableResult("!!!")

		require.Error(t, err)
		assert.Nil(t, event)
		assert.ErrorIs(t, err, ErrInvalidVerifiableResultBase64)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("{not-json"))

		event, err := DecodeVerifiableResult(encoded)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.ErrorIs(t, err, ErrInvalidVerifiableResultJSON)
	})
}

func bytesTrimLeftZeroes(b []byte) []byte {
	trimmed := strings.TrimLeft(hex.EncodeToString(b), "0")
	if trimmed == "" {
		return []byte{}
	}
	if len(trimmed)%2 != 0 {
		trimmed = "0" + trimmed
	}
	out, _ := hex.DecodeString(trimmed)
	return out
}
