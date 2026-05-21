package queries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"
)

const (
	testAPIKey          = "test-api-key"
	testChainSelector   = "16015286601757825753"
	testContractAddress = "0x1234567890123456789012345678901234567890"
	testFromAddress     = "0x000000000000000000000000000000000000dEaD"
	testCallData        = "0x18160ddd"
)

func setupQueriesTestClient(t *testing.T, handler http.HandlerFunc, modify ...func(*Options)) (*Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)

	api, err := apiClient.NewClientWithResponses(
		server.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey "+testAPIKey)
			return nil
		}),
	)
	require.NoError(t, err)

	opts := &Options{
		Logger:       slog.New(slog.DiscardHandler),
		APIClient:    api,
		PollInterval: time.Millisecond,
	}
	for _, m := range modify {
		m(opts)
	}

	client, err := NewClient(opts)
	require.NoError(t, err)

	return client, server
}

func writeJSON(t *testing.T, w http.ResponseWriter, statusCode int, body any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}
}

func validEVMParams() apiClient.EVMCallQueryParams {
	return apiClient.EVMCallQueryParams{
		ContractAddress: apiClient.EthereumAddress(testContractAddress),
		CallData:        testCallData,
		BlockSelection:  LatestBlockSelection(),
	}
}

func timestampPtr(v int64) *apiClient.Timestamp {
	ts := apiClient.Timestamp(v)
	return &ts
}

func makeAcceptedQuery(channelID, queryID uuid.UUID, status apiClient.QueryStatus) apiClient.Query {
	return apiClient.Query{
		QueryId:       queryID,
		ChannelId:     channelID,
		Status:        status,
		QueryKind:     apiClient.QueryKindEVMCall,
		ChainSelector: apiClient.ChainSelector(testChainSelector),
		CreatedAt:     1700000000,
		UpdatedAt:     1700000001,
	}
}

func makeVerifiableResult(t *testing.T, channelID, queryID uuid.UUID, rawReturnData *string, queryErr *models.ChainQueryError) string {
	t.Helper()

	var requested models.ChainQueryRequestedBlockSelection
	require.NoError(t, requested.FromChainQueryLatestBlockSelection(models.ChainQueryLatestBlockSelection{}))

	var result *models.ChainQueryExecutionResult
	if rawReturnData != nil {
		result = &models.ChainQueryExecutionResult{
			RawReturnData: *rawReturnData,
		}
	}

	event := models.ChainQueryVerifiableEvent{
		Service:       models.ChainQueryVerifiableEventServiceCREC,
		Name:          models.ChainQueryVerifiableEventNameChainQuery,
		ChainSelector: testChainSelector,
		Timestamp:     time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		Data: models.ChainQueryData{
			QueryId:   queryID,
			ChannelId: channelID,
			QueryKind: models.ChainQueryKindEVMCall,
			Target: models.ChainQueryTarget{
				FromAddress:     zeroAddress,
				ContractAddress: testContractAddress,
				CallData:        testCallData,
			},
			BlockSelection: models.ChainQueryBlockSelection{
				Requested: requested,
				Resolved: &models.ChainQueryResolvedBlock{
					BlockNumber:    "12345678",
					BlockHash:      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					BlockTimestamp: 1777334412,
				},
			},
			Result: result,
			Error:  queryErr,
		},
	}

	encoded, err := json.Marshal(event)
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(encoded)
}

func makeCompletedQuery(t *testing.T, channelID, queryID uuid.UUID, rawReturnData string) *apiClient.Query {
	t.Helper()

	eventHash := "0x5c7a"
	verifiableResult := makeVerifiableResult(t, channelID, queryID, &rawReturnData, nil)
	proof := apiClient.OCRProof{
		Alg:        "ocr",
		OcrReport:  "0x01",
		OcrContext: "0x02",
		Signatures: []string{"0x03"},
	}

	query := makeAcceptedQuery(channelID, queryID, apiClient.QueryStatusCompleted)
	query.EventHash = &eventHash
	query.VerifiableResult = &verifiableResult
	query.Proof = &proof
	query.CompletedAt = timestampPtr(1700000002)

	return &query
}

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		api, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		logger := slog.New(slog.DiscardHandler)
		client, err := NewClient(&Options{
			Logger:       logger,
			APIClient:    api,
			PollInterval: 123 * time.Millisecond,
		})

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, logger, client.logger)
		assert.Equal(t, api, client.apiClient)
		assert.Equal(t, 123*time.Millisecond, client.pollInterval)
	})

	t.Run("NilOptions", func(t *testing.T) {
		client, err := NewClient(nil)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrOptionsRequired)
	})

	t.Run("NilAPIClient", func(t *testing.T) {
		client, err := NewClient(&Options{Logger: slog.New(slog.DiscardHandler)})

		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrAPIClientRequired)
	})

	t.Run("DefaultLoggerAndPollInterval", func(t *testing.T) {
		api, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		client, err := NewClient(&Options{APIClient: api})

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.logger)
		assert.Equal(t, defaultPollInterval, client.pollInterval)
	})
}

func TestBlockSelectionHelpers(t *testing.T) {
	t.Run("Latest", func(t *testing.T) {
		selection := LatestBlockSelection()

		discriminator, err := selection.Discriminator()
		require.NoError(t, err)
		assert.Equal(t, "latest", discriminator)
		require.NoError(t, validateBlockSelection(selection))

		body, err := json.Marshal(selection)
		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"latest"}`, string(body))
	})

	t.Run("Finalized", func(t *testing.T) {
		selection := FinalizedBlockSelection()

		discriminator, err := selection.Discriminator()
		require.NoError(t, err)
		assert.Equal(t, "finalized", discriminator)
		require.NoError(t, validateBlockSelection(selection))
	})

	t.Run("BlockNumber", func(t *testing.T) {
		selection := BlockNumber(12345)

		discriminator, err := selection.Discriminator()
		require.NoError(t, err)
		assert.Equal(t, "block_number", discriminator)

		blockSelection, err := selection.AsBlockNumberBlockSelection()
		require.NoError(t, err)
		assert.Equal(t, "12345", blockSelection.BlockNumber)
		require.NoError(t, validateBlockSelection(selection))
	})

	t.Run("InvalidBlockNumberString", func(t *testing.T) {
		_, err := BlockNumberBlockSelectionString("not-a-number")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidBlockSelection)
	})

	t.Run("MissingSelection", func(t *testing.T) {
		err := validateBlockSelection(BlockSelection{})

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBlockSelectionRequired)
	})

	t.Run("UnsupportedSelection", func(t *testing.T) {
		var selection BlockSelection
		require.NoError(t, selection.UnmarshalJSON([]byte(`{"type":"safe"}`)))

		err := validateBlockSelection(selection)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidBlockSelection)
	})
}

func TestClient_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		queryID := uuid.New()
		contractAddress := apiClient.EthereumAddress("0x123456789012345678901234567890123456ABCD")
		fromAddress := apiClient.EthereumAddress(testFromAddress)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String()+"/queries", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "Apikey "+testAPIKey, r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var req apiClient.CreateQuery
			require.NoError(t, json.Unmarshal(body, &req))
			assert.Equal(t, "idem-1", req.IdempotencyKey)
			assert.Equal(t, apiClient.QueryKindEVMCall, req.QueryKind)
			assert.Equal(t, apiClient.ChainSelector(testChainSelector), req.ChainSelector)
			assert.Equal(t, strings.ToLower(string(contractAddress)), string(req.Params.ContractAddress))
			assert.Equal(t, "0xabcd", req.Params.CallData)
			require.NotNil(t, req.Params.FromAddress)
			assert.Equal(t, "0x000000000000000000000000000000000000dead", string(*req.Params.FromAddress))

			discriminator, err := req.Params.BlockSelection.Discriminator()
			require.NoError(t, err)
			assert.Equal(t, "latest", discriminator)

			writeJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
				QueryId: queryID,
				Status:  apiClient.QueryStatusAccepted,
			})
		}

		client, server := setupQueriesTestClient(t, handler)
		defer server.Close()

		resp, err := client.Create(context.Background(), CreateInput{
			ChannelID:      channelID,
			IdempotencyKey: "idem-1",
			QueryKind:      apiClient.QueryKindEVMCall,
			ChainSelector:  testChainSelector,
			Params: apiClient.EVMCallQueryParams{
				ContractAddress: contractAddress,
				CallData:        "0xABCD",
				BlockSelection:  LatestBlockSelection(),
				FromAddress:     &fromAddress,
			},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, queryID, resp.QueryId)
		assert.Equal(t, apiClient.QueryStatusAccepted, resp.Status)
	})

	t.Run("StatusMappings", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
			wantErr    error
		}{
			{name: "ChannelNotFound", statusCode: http.StatusNotFound, wantErr: ErrChannelNotFound},
			{name: "IdempotencyConflict", statusCode: http.StatusConflict, wantErr: ErrIdempotencyConflict},
			{name: "RateLimitExceeded", statusCode: http.StatusTooManyRequests, wantErr: ErrRateLimitExceeded},
			{name: "Unexpected", statusCode: http.StatusInternalServerError, wantErr: ErrUnexpectedStatusCode},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				handler := func(w http.ResponseWriter, r *http.Request) {
					writeJSON(t, w, tt.statusCode, apiClient.ApplicationError{
						Message: "error",
					})
				}

				client, server := setupQueriesTestClient(t, handler)
				defer server.Close()

				resp, err := client.Create(context.Background(), CreateInput{
					ChannelID:      uuid.New(),
					IdempotencyKey: "idem",
					QueryKind:      apiClient.QueryKindEVMCall,
					ChainSelector:  testChainSelector,
					Params:         validEVMParams(),
				})

				require.Error(t, err)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.wantErr)
			})
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("request should not be sent for validation failures")
		}

		client, server := setupQueriesTestClient(t, handler)
		defer server.Close()

		tests := []struct {
			name   string
			mutate func(*CreateInput)
			want   error
		}{
			{name: "MissingChannelID", mutate: func(in *CreateInput) { in.ChannelID = uuid.Nil }, want: ErrChannelIDRequired},
			{name: "MissingIdempotencyKey", mutate: func(in *CreateInput) { in.IdempotencyKey = "" }, want: ErrIdempotencyKeyRequired},
			{name: "MissingQueryKind", mutate: func(in *CreateInput) { in.QueryKind = "" }, want: ErrQueryKindRequired},
			{name: "UnsupportedQueryKind", mutate: func(in *CreateInput) { in.QueryKind = apiClient.QueryKind("evm_logs") }, want: ErrUnsupportedQueryKind},
			{name: "MissingChainSelector", mutate: func(in *CreateInput) { in.ChainSelector = "" }, want: ErrChainSelectorRequired},
			{name: "InvalidContractAddress", mutate: func(in *CreateInput) { in.Params.ContractAddress = "not-address" }, want: ErrInvalidContractAddress},
			{name: "InvalidCallData", mutate: func(in *CreateInput) { in.Params.CallData = "0xabc" }, want: ErrInvalidCallData},
			{name: "MissingBlockSelection", mutate: func(in *CreateInput) { in.Params.BlockSelection = BlockSelection{} }, want: ErrBlockSelectionRequired},
			{name: "InvalidFromAddress", mutate: func(in *CreateInput) {
				from := apiClient.EthereumAddress("not-address")
				in.Params.FromAddress = &from
			}, want: ErrInvalidFromAddress},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input := CreateInput{
					ChannelID:      uuid.New(),
					IdempotencyKey: "idem",
					QueryKind:      apiClient.QueryKindEVMCall,
					ChainSelector:  testChainSelector,
					Params:         validEVMParams(),
				}
				tt.mutate(&input)

				resp, err := client.Create(context.Background(), input)

				require.Error(t, err)
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, tt.want)
			})
		}
	})
}

func TestClient_CreateEVMCall(t *testing.T) {
	channelID := uuid.New()
	queryID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/channels/"+channelID.String()+"/queries", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req apiClient.CreateQuery
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "raw-call-1", req.IdempotencyKey)
		assert.Equal(t, apiClient.QueryKindEVMCall, req.QueryKind)
		assert.Equal(t, "0x18160ddd", req.Params.CallData)
		require.NotNil(t, req.Params.FromAddress)
		assert.Equal(t, "0x000000000000000000000000000000000000dead", string(*req.Params.FromAddress))
		require.NotNil(t, req.Metadata)
		assert.Equal(t, "client-ref", (*req.Metadata)["client_reference_id"])

		writeJSON(t, w, http.StatusAccepted, apiClient.QueryAcceptedResponse{
			QueryId: queryID,
			Status:  apiClient.QueryStatusAccepted,
		})
	}

	client, server := setupQueriesTestClient(t, handler)
	defer server.Close()

	resp, err := client.CreateEVMCall(
		context.Background(),
		CallContractInput{
			ChannelID:       channelID,
			ChainSelector:   testChainSelector,
			ContractAddress: testContractAddress,
			CallData:        []byte{0x18, 0x16, 0x0d, 0xdd},
			BlockSelection:  LatestBlockSelection(),
			IdempotencyKey:  "raw-call-1",
		},
		WithFromAddress(testFromAddress),
		WithMetadata(map[string]interface{}{"client_reference_id": "client-ref"}),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, queryID, resp.QueryId)
}
