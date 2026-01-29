package networks

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)

	apiKeyEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", "test-api-key")
		return nil
	}

	crecAPIClient, err := apiClient.NewClientWithResponses(
		server.URL,
		apiClient.WithRequestEditorFn(apiKeyEditor),
	)
	require.NoError(t, err)

	logger := slog.New(slog.DiscardHandler)
	client, err := NewClient(&Options{
		Logger:    logger,
		APIClient: crecAPIClient,
	})
	require.NoError(t, err)

	return client, server
}

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		crecAPIClient, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		logger := slog.New(slog.DiscardHandler)
		client, err := NewClient(&Options{
			Logger:    logger,
			APIClient: crecAPIClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.logger)
		assert.NotNil(t, client.apiClient)
	})

	t.Run("NilOptions", func(t *testing.T) {
		client, err := NewClient(nil)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.True(t, errors.Is(err, ErrOptionsRequired))
	})

	t.Run("NilAPIClient", func(t *testing.T) {
		logger := slog.New(slog.DiscardHandler)
		client, err := NewClient(&Options{
			Logger:    logger,
			APIClient: nil,
		})

		require.Error(t, err)
		assert.Nil(t, client)
		assert.True(t, errors.Is(err, ErrAPIClientRequired))
	})

	t.Run("DefaultLogger", func(t *testing.T) {
		crecAPIClient, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		client, err := NewClient(&Options{
			Logger:    nil,
			APIClient: crecAPIClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.logger)
	})
}

func TestClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		networkID := uuid.New()
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/networks", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.NetworkList{
				Data: []apiClient.Network{
					{
						Id:            networkID,
						Name:          "Ethereum Mainnet",
						ChainFamily:   "evm",
						ChainId:       "1",
						ChainSelector: "5009297550715157269",
						CreatedAt:     1700000000,
						UpdatedAt:     1700000000,
					},
				},
				HasMore: false,
			}
			_ = json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		networks, hasMore, err := client.List(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, networks)
		assert.Len(t, networks, 1)
		assert.False(t, hasMore)
		assert.Equal(t, networkID, networks[0].Id)
		assert.Equal(t, "Ethereum Mainnet", networks[0].Name)
		assert.Equal(t, "evm", networks[0].ChainFamily)
		assert.Equal(t, "1", networks[0].ChainId)
		assert.Equal(t, "5009297550715157269", networks[0].ChainSelector)
	})

	t.Run("SuccessEmpty", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(apiClient.NetworkList{
				Data:    []apiClient.Network{},
				HasMore: false,
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		networks, hasMore, err := client.List(context.Background())

		require.NoError(t, err)
		assert.NotNil(t, networks)
		assert.Len(t, networks, 0)
		assert.False(t, hasMore)
	})

	t.Run("SuccessWithHasMore", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(apiClient.NetworkList{
				Data: []apiClient.Network{
					{Id: uuid.New(), Name: "Network 1", ChainFamily: "evm", ChainSelector: "1"},
				},
				HasMore: true,
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		networks, hasMore, err := client.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.True(t, hasMore)
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		networks, hasMore, err := client.List(context.Background())

		require.Error(t, err)
		assert.Nil(t, networks)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListNetworks))
		assert.Contains(t, err.Error(), ErrUnexpectedStatusCode.Error())
	})

	t.Run("NotFound", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		networks, hasMore, err := client.List(context.Background())

		require.Error(t, err)
		assert.Nil(t, networks)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListNetworks))
	})
}
