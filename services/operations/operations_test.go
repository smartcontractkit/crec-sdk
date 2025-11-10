package operations

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

func setupTestService(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	server := httptest.NewServer(handler)

	crecClient, err := client.NewCRECClient(&client.ClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})
	require.NoError(t, err)

	logger := zerolog.Nop()
	service, err := NewService(&ServiceOptions{
		Logger:     &logger,
		CRECClient: crecClient,
	})
	require.NoError(t, err)

	return service, server
}

func TestNewService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		crecClient, err := client.NewCRECClient(&client.ClientOptions{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-api-key",
		})
		require.NoError(t, err)

		logger := zerolog.Nop()
		service, err := NewService(&ServiceOptions{
			Logger:     &logger,
			CRECClient: crecClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
		assert.NotNil(t, service.crecClient)
	})

	t.Run("NilOptions", func(t *testing.T) {
		service, err := NewService(nil)

		require.Error(t, err)
		assert.Nil(t, service)
		assert.True(t, errors.Is(err, ErrServiceOptionsRequired), "Expected ErrServiceOptionsRequired, got: %v", err)
	})

	t.Run("NilCRECClient", func(t *testing.T) {
		logger := zerolog.Nop()
		service, err := NewService(&ServiceOptions{
			Logger:     &logger,
			CRECClient: nil,
		})

		require.Error(t, err)
		assert.Nil(t, service)
		assert.True(t, errors.Is(err, ErrCRECClientRequired), "Expected ErrCRECClientRequired, got: %v", err)
	})

	t.Run("DefaultLogger", func(t *testing.T) {
		crecClient, err := client.NewCRECClient(&client.ClientOptions{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-api-key",
		})
		require.NoError(t, err)

		service, err := NewService(&ServiceOptions{
			Logger:     nil,
			CRECClient: crecClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
	})
}

func TestService_CreateOperation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/operations"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateOperation
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)
			assert.Equal(t, uint64(1337), createReq.ChainSelector)
			assert.Equal(t, "0x1234", createReq.Address)
			assert.Equal(t, "op-123", createReq.WalletOperationId)
			assert.Len(t, createReq.Transactions, 1)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.OperationResponse{
				OperationId: operationID,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		returnedOperationID, err := service.CreateOperation(context.Background(), CreateOperationInput{
			ChannelID:         channelID,
			ChainSelector:     1337,
			Address:           "0x1234",
			WalletOperationID: "op-123",
			Transactions: []TransactionRequest{
				{
					To:    "0x5678",
					Value: "0",
					Data:  "0xabcd",
				},
			},
			Signature: "0xsignature",
		})

		require.NoError(t, err)
		assert.NotNil(t, returnedOperationID)
		assert.Equal(t, operationID, *returnedOperationID)
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with invalid input")
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		testCases := []struct {
			name          string
			input         CreateOperationInput
			expectedError error
		}{
			{
				name: "EmptyChannelID",
				input: CreateOperationInput{
					ChannelID:         uuid.Nil,
					ChainSelector:     1337,
					Address:           "0x1234",
					WalletOperationID: "op-123",
					Transactions:      []TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
					Signature:         "0xsig",
				},
				expectedError: ErrChannelIDRequired,
			},
			{
				name: "EmptyChainSelector",
				input: CreateOperationInput{
					ChannelID:         uuid.New(),
					ChainSelector:     0,
					Address:           "0x1234",
					WalletOperationID: "op-123",
					Transactions:      []TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
					Signature:         "0xsig",
				},
				expectedError: ErrChainSelectorRequired,
			},
			{
				name: "EmptyAddress",
				input: CreateOperationInput{
					ChannelID:         uuid.New(),
					ChainSelector:     1337,
					Address:           "",
					WalletOperationID: "op-123",
					Transactions:      []TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
					Signature:         "0xsig",
				},
				expectedError: ErrAddressRequired,
			},
			{
				name: "EmptyWalletOperationID",
				input: CreateOperationInput{
					ChannelID:         uuid.New(),
					ChainSelector:     1337,
					Address:           "0x1234",
					WalletOperationID: "",
					Transactions:      []TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
					Signature:         "0xsig",
				},
				expectedError: ErrWalletOperationIDRequired,
			},
			{
				name: "EmptyTransactions",
				input: CreateOperationInput{
					ChannelID:         uuid.New(),
					ChainSelector:     1337,
					Address:           "0x1234",
					WalletOperationID: "op-123",
					Transactions:      []TransactionRequest{},
					Signature:         "0xsig",
				},
				expectedError: ErrAtLeastOneTransactionRequired,
			},
			{
				name: "EmptySignature",
				input: CreateOperationInput{
					ChannelID:         uuid.New(),
					ChainSelector:     1337,
					Address:           "0x1234",
					WalletOperationID: "op-123",
					Transactions:      []TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
					Signature:         "",
				},
				expectedError: ErrSignatureRequired,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				opID, err := service.CreateOperation(context.Background(), tc.input)

				require.Error(t, err)
				assert.Nil(t, opID)
				assert.True(t, errors.Is(err, tc.expectedError), "Expected %v, got: %v", tc.expectedError, err)
			})
		}
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Channel not found",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		opID, err := service.CreateOperation(context.Background(), CreateOperationInput{
			ChannelID:         channelID,
			ChainSelector:     1337,
			Address:           "0x1234",
			WalletOperationID: "op-123",
			Transactions: []TransactionRequest{
				{To: "0x5678", Value: "0", Data: "0xabcd"},
			},
			Signature: "0xsig",
		})

		require.Error(t, err)
		assert.Nil(t, opID)
		assert.True(t, errors.Is(err, ErrChannelNotFound), "Expected ErrChannelNotFound, got: %v", err)
	})

	t.Run("BadRequest", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid operation data",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		opID, err := service.CreateOperation(context.Background(), CreateOperationInput{
			ChannelID:         channelID,
			ChainSelector:     1337,
			Address:           "0x1234",
			WalletOperationID: "op-123",
			Transactions: []TransactionRequest{
				{To: "0x5678", Value: "0", Data: "0xabcd"},
			},
			Signature: "0xsig",
		})

		require.Error(t, err)
		assert.Nil(t, opID)
		assert.True(t, errors.Is(err, ErrCreateOperation), "Expected ErrCreateOperation, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestService_GetOperation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/operations/" + operationID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Operation{
				OperationId:       operationID,
				Status:            "pending",
				ChainSelector:     uint64(1337),
				Address:           "0x1234",
				WalletOperationId: "op-123",
				Transactions: []apiClient.TransactionRequest{
					{To: "0x5678", Value: "0", Data: "0xabcd"},
				},
				Signature: "0xsig",
				CreatedAt: createdAt,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operation, err := service.GetOperation(context.Background(), channelID, operationID)

		require.NoError(t, err)
		assert.NotNil(t, operation)
		assert.Equal(t, operationID, operation.OperationId)
		assert.Equal(t, "pending", operation.Status)
		assert.Equal(t, uint64(1337), operation.ChainSelector)
		assert.Equal(t, "0x1234", operation.Address)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Operation not found",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operation, err := service.GetOperation(context.Background(), channelID, operationID)

		require.Error(t, err)
		assert.Nil(t, operation)
		assert.True(t, errors.Is(err, ErrOperationNotFound), "Expected ErrOperationNotFound, got: %v", err)
	})

	t.Run("ServerError", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operation, err := service.GetOperation(context.Background(), channelID, operationID)

		require.Error(t, err)
		assert.Nil(t, operation)
		assert.True(t, errors.Is(err, ErrGetOperation), "Expected ErrGetOperation, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestService_ListOperations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		operation1ID := uuid.New()
		operation2ID := uuid.New()
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/operations"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, "20", query.Get("limit"))
			assert.Equal(t, "0", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.OperationList{
				Data: []apiClient.Operation{
					{
						OperationId:       operation1ID,
						Status:            "pending",
						ChainSelector:     uint64(1337),
						Address:           "0x1234",
						WalletOperationId: "op-1",
						Transactions:      []apiClient.TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
						Signature:         "0xsig1",
						CreatedAt:         createdAt,
					},
					{
						OperationId:       operation2ID,
						Status:            "confirmed",
						ChainSelector:     uint64(1337),
						Address:           "0x1234",
						WalletOperationId: "op-2",
						Transactions:      []apiClient.TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
						Signature:         "0xsig2",
						CreatedAt:         createdAt,
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		limit := 20
		offset := 0
		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: channelID,
			Limit:     &limit,
			Offset:    &offset,
		})

		require.NoError(t, err)
		assert.Len(t, operations, 2)
		assert.False(t, hasMore)
		assert.Equal(t, operation1ID, operations[0].OperationId)
		assert.Equal(t, "pending", operations[0].Status)
		assert.Equal(t, operation2ID, operations[1].OperationId)
		assert.Equal(t, "confirmed", operations[1].Status)
	})

	t.Run("WithFilters", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()
		createdAt := time.Now().Unix()
		status := "pending"
		chainSelector := "1337"
		address := "0x1234"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)

			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, status, query.Get("status"))
			assert.Equal(t, chainSelector, query.Get("chain_selector"))
			assert.Equal(t, address, query.Get("address"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.OperationList{
				Data: []apiClient.Operation{
					{
						OperationId:       operationID,
						Status:            status,
						ChainSelector:     uint64(1337),
						Address:           address,
						WalletOperationId: "op-1",
						Transactions:      []apiClient.TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
						Signature:         "0xsig",
						CreatedAt:         createdAt,
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID:     channelID,
			Status:        &status,
			ChainSelector: &chainSelector,
			Address:       &address,
		})

		require.NoError(t, err)
		assert.Len(t, operations, 1)
		assert.False(t, hasMore)
		assert.Equal(t, status, operations[0].Status)
		assert.Equal(t, uint64(1337), operations[0].ChainSelector)
		assert.Equal(t, address, operations[0].Address)
	})

	t.Run("WithWalletIDFilter", func(t *testing.T) {
		channelID := uuid.New()
		operationID := uuid.New()
		walletID := uuid.New()
		createdAt := time.Now().Unix()
		status := "pending"
		address := "0x1234"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)

			// Check that wallet_id query parameter is passed correctly
			query := r.URL.Query()
			assert.Equal(t, walletID.String(), query.Get("wallet_id"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.OperationList{
				Data: []apiClient.Operation{
					{
						OperationId:       operationID,
						Status:            status,
						ChainSelector:     uint64(1337),
						Address:           address,
						WalletOperationId: "op-1",
						Transactions:      []apiClient.TransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
						Signature:         "0xsig",
						CreatedAt:         createdAt,
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: channelID,
			WalletID:  &walletID,
		})

		require.NoError(t, err)
		assert.Len(t, operations, 1)
		assert.False(t, hasMore)
		assert.Equal(t, operationID, operations[0].OperationId)
	})

	t.Run("WithPagination", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "5", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.OperationList{
				Data:    []apiClient.Operation{},
				HasMore: true,
			}
			json.NewEncoder(w).Encode(response)
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		limit := 10
		offset := 5
		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: channelID,
			Limit:     &limit,
			Offset:    &offset,
		})

		require.NoError(t, err)
		assert.Len(t, operations, 0)
		assert.True(t, hasMore)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: uuid.Nil,
		})

		require.Error(t, err)
		assert.Nil(t, operations)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Channel not found",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: channelID,
		})

		require.Error(t, err)
		assert.Nil(t, operations)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrChannelNotFound), "Expected ErrChannelNotFound, got: %v", err)
	})

	t.Run("ServerError", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		operations, hasMore, err := service.ListOperations(context.Background(), ListOperationsInput{
			ChannelID: channelID,
		})

		require.Error(t, err)
		assert.Nil(t, operations)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListOperations), "Expected ErrListOperations, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}
