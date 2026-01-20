package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	// Add API key header to all requests
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
		Logger:       logger,
		APIClient:    crecAPIClient,
		PollInterval: 10 * time.Millisecond, // Fast polling for tests
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
		assert.True(t, errors.Is(err, ErrOptionsRequired), "Expected ErrOptionsRequired, got: %v", err)
	})

	t.Run("NilAPIClient", func(t *testing.T) {
		logger := slog.New(slog.DiscardHandler)
		client, err := NewClient(&Options{
			Logger:    logger,
			APIClient: nil,
		})

		require.Error(t, err)
		assert.Nil(t, client)
		assert.True(t, errors.Is(err, ErrAPIClientRequired), "Expected ErrAPIClientRequired, got: %v", err)
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

	t.Run("DefaultPollInterval", func(t *testing.T) {
		crecAPIClient, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		logger := slog.New(slog.DiscardHandler)
		client, err := NewClient(&Options{
			Logger:    logger,
			APIClient: crecAPIClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, 2*time.Second, client.pollInterval)
	})

	t.Run("CustomPollInterval", func(t *testing.T) {
		crecAPIClient, err := apiClient.NewClientWithResponses("http://localhost:8080")
		require.NoError(t, err)

		logger := slog.New(slog.DiscardHandler)
		customInterval := 500 * time.Millisecond
		client, err := NewClient(&Options{
			Logger:       logger,
			APIClient:    crecAPIClient,
			PollInterval: customInterval,
		})

		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, customInterval, client.pollInterval)
	})
}

func TestClient_CreateWithDomain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "test-watcher"
		domain := "dvp"
		chainSelector := "1337"
		address := "0x1234567890abcdef"

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)

			domainWatcher, err := createReq.AsCreateWatcherWithDomain()
			require.NoError(t, err)
			require.NotNil(t, domainWatcher.Name)
			assert.Equal(t, watcherName, *domainWatcher.Name)
			assert.Equal(t, domain, domainWatcher.Domain)
			assert.Equal(t, chainSelector, domainWatcher.ChainSelector)
			assert.Equal(t, address, domainWatcher.Address)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &watcherName,
				ChainSelector: chainSelector,
				Address:       address,
				Status:        "pending",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithDomain(context.Background(), channelID, CreateWithDomainInput{
			Name:          &watcherName,
			Domain:        domain,
			ChainSelector: chainSelector,
			Address:       address,
			Events:        []string{"TestEvent"},
		})

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		require.NotNil(t, watcher.Name)
		assert.Equal(t, watcherName, *watcher.Name)
		assert.Equal(t, chainSelector, watcher.ChainSelector)
		assert.Equal(t, address, watcher.Address)
		assert.Equal(t, "pending", watcher.Status)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithDomain(context.Background(), uuid.Nil, CreateWithDomainInput{
			Name:          &name,
			Domain:        "dvp",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyDomain", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty domain")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithDomain(context.Background(), uuid.New(), CreateWithDomainInput{
			Name:          &name,
			Domain:        "",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrDomainRequired), "Expected ErrDomainRequired, got: %v", err)
	})

	t.Run("EmptyAddress", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty address")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithDomain(context.Background(), uuid.New(), CreateWithDomainInput{
			Name:          &name,
			Domain:        "dvp",
			ChainSelector: "1337",
			Address:       "",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrAddressRequired), "Expected ErrAddressRequired, got: %v", err)
	})

	t.Run("BadRequest", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid watcher configuration",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithDomain(context.Background(), channelID, CreateWithDomainInput{
			Name:          &name,
			Domain:        "dvp",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrCreateWatcherDomain), "Expected ErrCreateWatcherDomain, got: %v", err)
	})
}

func TestClient_CreateWithABI(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "test-watcher-abi"
		chainSelector := "1337"
		address := "0x1234567890abcdef"

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)

			abiWatcher, err := createReq.AsCreateWatcherWithABI()
			require.NoError(t, err)
			require.NotNil(t, abiWatcher.Name)
			assert.Equal(t, watcherName, *abiWatcher.Name)
			assert.Equal(t, chainSelector, abiWatcher.ChainSelector)
			assert.Equal(t, address, abiWatcher.Address)
			assert.Len(t, abiWatcher.Abi, 1)
			assert.Equal(t, "Transfer", abiWatcher.Abi[0].Name)
			assert.Len(t, abiWatcher.Abi[0].Inputs, 3)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &watcherName,
				ChainSelector: chainSelector,
				Address:       address,
				Status:        "pending",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithABI(context.Background(), channelID, CreateWithABIInput{
			Name:          &watcherName,
			ChainSelector: chainSelector,
			Address:       address,
			Events:        []string{"Transfer"},
			ABI: []EventABI{
				{
					Name:      "Transfer",
					Type:      "event",
					Anonymous: false,
					Inputs: []EventABIInput{
						{
							Indexed: true,
							Name:    "from",
							Type:    "address",
						},
						{
							Indexed: true,
							Name:    "to",
							Type:    "address",
						},
						{
							Indexed: false,
							Name:    "value",
							Type:    "uint256",
						},
					},
				},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		require.NotNil(t, watcher.Name)
		assert.Equal(t, watcherName, *watcher.Name)
		assert.Equal(t, chainSelector, watcher.ChainSelector)
		assert.Equal(t, address, watcher.Address)
		assert.Equal(t, "pending", watcher.Status)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithABI(context.Background(), uuid.Nil, CreateWithABIInput{
			Name:          &name,
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
			ABI:           []EventABI{},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyABI", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithABI(context.Background(), uuid.New(), CreateWithABIInput{
			Name:          &name,
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
			ABI:           []EventABI{},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrABIRequired), "Expected ErrABIRequired, got: %v", err)
	})

	t.Run("InvalidABIType", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithABI(context.Background(), uuid.New(), CreateWithABIInput{
			Name:          &name,
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestFunction"},
			ABI: []EventABI{
				{
					Name:      "TestFunction",
					Type:      "function", // Invalid: only "event" is supported
					Anonymous: false,
					Inputs: []EventABIInput{
						{Name: "param1", Type: "uint256", Indexed: false},
					},
				},
			},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrInvalidABIType), "Expected ErrInvalidABIType, got: %v", err)
		assert.Contains(t, err.Error(), "function")
		assert.Contains(t, err.Error(), "TestFunction")
	})

	t.Run("EventNotInABI", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithABI(context.Background(), uuid.New(), CreateWithABIInput{
			Name:          &name,
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"Transfer", "Approval", "MissingEvent"}, // MissingEvent not in ABI
			ABI: []EventABI{
				{
					Name:      "Transfer",
					Type:      "event",
					Anonymous: false,
					Inputs: []EventABIInput{
						{Name: "from", Type: "address", Indexed: true},
						{Name: "to", Type: "address", Indexed: true},
						{Name: "value", Type: "uint256", Indexed: false},
					},
				},
				{
					Name:      "Approval",
					Type:      "event",
					Anonymous: false,
					Inputs: []EventABIInput{
						{Name: "owner", Type: "address", Indexed: true},
						{Name: "spender", Type: "address", Indexed: true},
						{Name: "value", Type: "uint256", Indexed: false},
					},
				},
			},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrEventNotInABI), "Expected ErrEventNotInABI, got: %v", err)
		assert.Contains(t, err.Error(), "MissingEvent")
		assert.Contains(t, err.Error(), "not found in ABI definitions")
	})
}

func TestClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcher1ID := uuid.New()
		watcher2ID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			query := r.URL.Query()
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "0", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name1 := "watcher-1"
			name2 := "watcher-2"
			response := apiClient.WatcherList{
				Data: []apiClient.Watcher{
					{
						WatcherId:     watcher1ID,
						Name:          &name1,
						ChainSelector: "1337",
						Address:       "0x1111",
						Status:        "active",
					},
					{
						WatcherId:     watcher2ID,
						Name:          &name2,
						ChainSelector: "1337",
						Address:       "0x2222",
						Status:        "active",
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		limit := 10
		offset := int64(0)
		result, err := client.List(context.Background(), channelID, ListFilters{
			Limit:  &limit,
			Offset: &offset,
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.False(t, result.HasMore)
		assert.Equal(t, watcher1ID, result.Data[0].WatcherId)
		require.NotNil(t, result.Data[0].Name)
		assert.Equal(t, "watcher-1", *result.Data[0].Name)
		assert.Equal(t, watcher2ID, result.Data[1].WatcherId)
		require.NotNil(t, result.Data[1].Name)
		assert.Equal(t, "watcher-2", *result.Data[1].Name)
	})

	t.Run("WithFilters", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		name := "test-watcher"
		status := StatusActive

		handler := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, name, query.Get("name"))
			assert.Equal(t, string(status), query.Get("status"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.WatcherList{
				Data: []apiClient.Watcher{
					{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1111",
						Status:        string(status),
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		result, err := client.List(context.Background(), channelID, ListFilters{
			Name:   &name,
			Status: &status,
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		require.NotNil(t, result.Data[0].Name)
		assert.Equal(t, name, *result.Data[0].Name)
		assert.Equal(t, string(status), result.Data[0].Status)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		result, err := client.List(context.Background(), uuid.Nil, ListFilters{})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
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

		client, server := setupTestClient(t, handler)
		defer server.Close()

		result, err := client.List(context.Background(), channelID, ListFilters{})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, ErrListWatchers), "Expected ErrListWatchers, got: %v", err)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers/" + watcherID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		require.NotNil(t, watcher.Name)
		assert.Equal(t, "test-watcher", *watcher.Name)
		assert.Equal(t, "active", watcher.Status)
	})

	t.Run("SuccessWithDONInfo", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				ChannelId:     channelID,
				Name:          &name,
				ChainSelector: "16015286601757825753",
				Address:       "0x1234567890123456789012345678901234567890",
				Status:        "active",
				CreatedAt:     1704067200, // 2024-01-01 00:00:00 UTC
				Events:        []string{"Transfer"},
				WorkflowId:    "00a52f385ef2c2ae57721370dbcef8b25ab406de2be190575c88e324c002e22f",
				DonFamily:     "zone-a",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		assert.Equal(t, channelID, watcher.ChannelId)
		assert.Equal(t, "zone-a", watcher.DonFamily)
		assert.Equal(t, "00a52f385ef2c2ae57721370dbcef8b25ab406de2be190575c88e324c002e22f", watcher.WorkflowId)
		assert.Equal(t, []string{"Transfer"}, watcher.Events)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), uuid.Nil, uuid.New())

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), uuid.New(), uuid.Nil)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherIDRequired), "Expected ErrWatcherIDRequired, got: %v", err)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Watcher not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})
}

func TestClient_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		newName := "updated-watcher"

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers/" + watcherID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var updateReq apiClient.UpdateWatcher
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			assert.Equal(t, newName, updateReq.Name)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &newName,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Update(context.Background(), channelID, watcherID, UpdateInput{
			Name: newName,
		})

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		require.NotNil(t, watcher.Name)
		assert.Equal(t, newName, *watcher.Name)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Update(context.Background(), uuid.Nil, uuid.New(), UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Update(context.Background(), uuid.New(), uuid.Nil, UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherIDRequired), "Expected ErrWatcherIDRequired, got: %v", err)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Update(context.Background(), uuid.New(), uuid.New(), UpdateInput{
			Name: "",
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrNameRequired), "Expected ErrNameRequired, got: %v", err)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Watcher not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Update(context.Background(), channelID, watcherID, UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("SuccessSync", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers/" + watcherID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.WriteHeader(http.StatusNoContent)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID, watcherID)

		require.NoError(t, err)
	})

	t.Run("SuccessAsync", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers/" + watcherID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.WriteHeader(http.StatusAccepted) // 202 for async deletion
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID, watcherID)

		require.NoError(t, err)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), uuid.Nil, uuid.New())

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), uuid.New(), uuid.Nil)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherIDRequired), "Expected ErrWatcherIDRequired, got: %v", err)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Watcher not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})

	t.Run("ServerError", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDeleteWatcher), "Expected ErrDeleteWatcher, got: %v", err)
	})
}

func TestClient_WaitForActive(t *testing.T) {
	t.Run("SuccessImmediatelyActive", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, "active", watcher.Status)
	})

	t.Run("SuccessAfterPolling", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		callCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			var status string
			if callCount < 3 {
				status = "pending"
			} else {
				status = "active"
			}

			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        status,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 10*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, "active", watcher.Status)
		assert.GreaterOrEqual(t, callCount, 3)
	})

	t.Run("FailedStatus", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "failed",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherDeploymentFailed), "Expected ErrWatcherDeploymentFailed, got: %v", err)
	})

	t.Run("Timeout", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "pending",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 100*time.Millisecond)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWaitForActiveTimeout), "Expected ErrWaitForActiveTimeout, got: %v", err)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), uuid.Nil, uuid.New(), 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), uuid.New(), uuid.Nil, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherIDRequired), "Expected ErrWatcherIDRequired, got: %v", err)
	})

	t.Run("DeletingStatus", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "deleting",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherIsDeleting), "Expected ErrWatcherIsDeleting, got: %v", err)
	})

	t.Run("DeletedStatus", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "deleted",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherAlreadyDeleted), "Expected ErrWatcherAlreadyDeleted, got: %v", err)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow activation - always return pending status
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "pending",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel the context after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		watcher, err := client.WaitForActive(ctx, channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, err == context.Canceled || strings.Contains(err.Error(), "context canceled"),
			"Expected context cancellation error, got: %v", err)
	})

	t.Run("TransientErrorRetry_EventuallySucceeds", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.Header().Set("Content-Type", "application/json")

			// First 2 attempts return 503 (transient error)
			if attemptCount <= 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "service temporarily unavailable"}`))
				return
			}

			// Third attempt succeeds with active status
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, "active", watcher.Status)
		assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried after transient errors")
	})

	t.Run("TransientErrorRetry_EventuallyTimesOut", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			// Always return 503 (transient error)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service temporarily unavailable"}`))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 200*time.Millisecond)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWaitForActiveTimeout), "Expected timeout error, got: %v", err)
		assert.Greater(t, attemptCount, 1, "Should have retried multiple times before timeout")
	})

	t.Run("PermanentError_FailsImmediately", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			// Return 400 (permanent error - bad request)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "bad request"}`))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.Equal(t, 1, attemptCount, "Should NOT retry permanent errors")
		assert.True(t, errors.Is(err, ErrCheckWatcherStatus), "Expected ErrCheckWatcherStatus, got: %v", err)
	})

	t.Run("TransientError500_Retry_EventuallySucceeds", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.Header().Set("Content-Type", "application/json")

			// First 2 attempts return 500 (internal server error - transient)
			if attemptCount <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
				return
			}

			// Third attempt succeeds with active status
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, "active", watcher.Status)
		assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried after 500 errors")
	})
}

func TestClient_WaitForDeleted(t *testing.T) {
	t.Run("SuccessImmediatelyDeleted", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "deleted",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
	})

	t.Run("SuccessAfterPolling", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		callCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			var status string
			if callCount < 3 {
				status = "deleting"
			} else {
				status = "deleted"
			}

			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        status,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, callCount, 3)
	})

	t.Run("SuccessNotFound", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Watcher not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err) // 404 means it's been deleted
	})

	t.Run("Timeout", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "deleting",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		// Use a very short timeout to test timeout behavior
		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 100*time.Millisecond)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWaitForDeletedTimeout), "Expected ErrWaitForDeletedTimeout, got: %v", err)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow deletion - always return deleting status
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "deleting",
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel the context after a short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := client.WaitForDeleted(ctx, channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		// The error should be either context.Canceled directly or wrapped
		assert.True(t, err == context.Canceled || strings.Contains(err.Error(), "context canceled"),
			"Expected context cancellation error, got: %v", err)
	})

	t.Run("UnexpectedStatus", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        "active", // Unexpected status while waiting for deletion
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherDeletionFailed), "Expected ErrWatcherDeletionFailed, got: %v", err)
		assert.Contains(t, err.Error(), "active")
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), uuid.Nil, uuid.New(), 5*time.Second)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), uuid.New(), uuid.Nil, 5*time.Second)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherIDRequired), "Expected ErrWatcherIDRequired, got: %v", err)
	})

	t.Run("TransientErrorRetry_EventuallySucceeds", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.Header().Set("Content-Type", "application/json")

			// First 2 attempts return 502 (transient error)
			if attemptCount <= 2 {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(`{"error": "bad gateway"}`))
				return
			}

			// Third attempt succeeds with 404 (watcher deleted)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried after transient errors")
	})

	t.Run("TransientErrorRetry_EventuallyTimesOut", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			// Always return 500 (transient error)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForDeleted(context.Background(), channelID, watcherID, 200*time.Millisecond)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWaitForDeletedTimeout), "Expected timeout error, got: %v", err)
		assert.Greater(t, attemptCount, 1, "Should have retried multiple times before timeout")
	})
}

func TestEndToEnd_WatcherLifecycle(t *testing.T) {
	t.Run("CreateWithDomain_WaitActive_Update_Delete", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "integration-test-watcher"
		updatedName := "updated-watcher"

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			switch {
			// 1. Create watcher
			case r.Method == "POST" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers"):
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "pending",
				}
				json.NewEncoder(w).Encode(response)

			// 2-3. Wait for active (first pending, then active)
			case r.Method == "GET" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				status := "pending"
				name := watcherName
				if callCount > 2 {
					status = "active"
				}
				if callCount > 4 {
					name = updatedName
				}
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &name,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        status,
				}
				json.NewEncoder(w).Encode(response)

			// 4. Update watcher
			case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &updatedName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "active",
				}
				json.NewEncoder(w).Encode(response)

			// 5. Delete watcher
			case r.Method == "DELETE" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusNoContent)

			default:
				t.Fatalf("Unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Step 1: Create watcher with domain
		createInput := CreateWithDomainInput{
			Name:          &watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Domain:        "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithDomain(ctx, channelID, createInput)
		require.NoError(t, err)
		assert.Equal(t, watcherID, created.WatcherId)
		assert.Equal(t, "pending", created.Status)

		// Step 2: Wait for watcher to become active
		active, err := client.WaitForActive(ctx, channelID, watcherID, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, "active", active.Status)

		// Step 3: Find the watcher
		found, err := client.Get(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, watcherID, found.WatcherId)
		assert.Equal(t, watcherName, *found.Name)

		// Step 4: Update the watcher
		updateInput := UpdateInput{
			Name: updatedName,
		}
		updated, err := client.Update(ctx, channelID, watcherID, updateInput)
		require.NoError(t, err)
		assert.Equal(t, updatedName, *updated.Name)

		// Step 5: Verify update
		found, err = client.Get(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, updatedName, *found.Name)

		// Step 6: Delete the watcher
		err = client.Delete(ctx, channelID, watcherID)
		require.NoError(t, err)

		// Verify we made all expected calls
		assert.GreaterOrEqual(t, callCount, 6, "Should have made at least 6 API calls")
	})

	t.Run("CreateWithABI_WaitActive_List_Delete", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "abi-test-watcher"

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			switch {
			// 1. Create watcher with ABI
			case r.Method == "POST" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers"):
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x5678",
					Status:        "pending",
				}
				json.NewEncoder(w).Encode(response)

			// 2. Wait for active (immediately active)
			case r.Method == "GET" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x5678",
					Status:        "active",
				}
				json.NewEncoder(w).Encode(response)

			// 3. List watchers
			case r.Method == "GET" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers") && !strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				response := apiClient.WatcherList{
					Data: []apiClient.Watcher{
						{
							WatcherId:     watcherID,
							Name:          &watcherName,
							ChainSelector: "1337",
							Address:       "0x5678",
							Status:        "active",
						},
					},
					HasMore: false,
				}
				json.NewEncoder(w).Encode(response)

			// 4. Delete watcher (async)
			case r.Method == "DELETE" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusAccepted)

			default:
				t.Fatalf("Unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Step 1: Create watcher with custom ABI
		createInput := CreateWithABIInput{
			Name:          &watcherName,
			ChainSelector: "1337",
			Address:       "0x5678",
			Events:        []string{"Transfer"},
			ABI: []EventABI{
				{
					Name:      "Transfer",
					Type:      "event",
					Anonymous: false,
					Inputs: []EventABIInput{
						{Name: "from", Type: "address", Indexed: true},
						{Name: "to", Type: "address", Indexed: true},
						{Name: "value", Type: "uint256", Indexed: false},
					},
				},
			},
		}
		created, err := client.CreateWithABI(ctx, channelID, createInput)
		require.NoError(t, err)
		assert.Equal(t, watcherID, created.WatcherId)

		// Step 2: Wait for active
		active, err := client.WaitForActive(ctx, channelID, watcherID, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, "active", active.Status)

		// Step 3: List all watchers in the channel
		filters := ListFilters{}
		list, err := client.List(ctx, channelID, filters)
		require.NoError(t, err)
		assert.Len(t, list.Data, 1)
		assert.Equal(t, watcherID, list.Data[0].WatcherId)

		// Step 4: Delete the watcher (async)
		err = client.Delete(ctx, channelID, watcherID)
		require.NoError(t, err)

		assert.Equal(t, 4, callCount, "Should have made 4 API calls")
	})
}

// TestEndToEnd_ErrorScenarios tests various error scenarios in realistic workflows
func TestEndToEnd_ErrorScenarios(t *testing.T) {
	t.Run("CreateFails_NeverReachesActive", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "invalid chain selector"}`))
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Try to create with invalid data
		createInput := CreateWithDomainInput{
			ChainSelector: "0", // Invalid
			Address:       "0x1234",
			Domain:        "dvp",
			Events:        []string{"TestEvent"},
		}
		_, err := client.CreateWithDomain(ctx, channelID, createInput)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChainSelectorRequired), "Expected ErrChainSelectorRequired, got: %v", err)
	})

	t.Run("CreateSucceeds_ButFailsToDeploy", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "failing-watcher"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "pending",
				}
				json.NewEncoder(w).Encode(response)
			} else if r.Method == "GET" {
				// Watcher failed to deploy
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "failed",
				}
				json.NewEncoder(w).Encode(response)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Create watcher
		createInput := CreateWithDomainInput{
			Name:          &watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Domain:        "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithDomain(ctx, channelID, createInput)
		require.NoError(t, err)

		// Wait for active - should fail because watcher deployment failed
		_, err = client.WaitForActive(ctx, channelID, created.WatcherId, 5*time.Second)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherDeploymentFailed), "Expected ErrWatcherDeploymentFailed, got: %v", err)
	})

	t.Run("WatcherIsDeleted_WhileWaitingForActive", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "deleted-watcher"

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "pending",
				}
				json.NewEncoder(w).Encode(response)
			} else if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
				// First call: pending, then deleted
				status := "pending"
				if callCount > 2 {
					status = "deleted"
				}
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        status,
				}
				json.NewEncoder(w).Encode(response)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Create watcher
		createInput := CreateWithDomainInput{
			Name:          &watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Domain:        "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithDomain(ctx, channelID, createInput)
		require.NoError(t, err)

		// Wait for active - should fail because watcher was deleted
		_, err = client.WaitForActive(ctx, channelID, created.WatcherId, 5*time.Second)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherAlreadyDeleted), "Expected ErrWatcherAlreadyDeleted, got: %v", err)
	})

	t.Run("UpdateNonExistentWatcher", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PATCH" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "watcher not found"}`))
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Try to update non-existent watcher
		updateInput := UpdateInput{
			Name: "new-name",
		}
		_, err := client.Update(ctx, channelID, watcherID, updateInput)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})

	t.Run("DeleteDuringWaitForDeleted_CompletesSuccessfully", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			if r.Method == "DELETE" {
				// Async deletion
				w.WriteHeader(http.StatusAccepted)
			} else if r.Method == "GET" {
				// First call: deleting, then deleted
				if callCount <= 2 {
					w.WriteHeader(http.StatusOK)
					name := "test-watcher"
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        "deleting",
					}
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(http.StatusOK)
					name := "test-watcher"
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        "deleted",
					}
					json.NewEncoder(w).Encode(response)
				}
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Delete watcher (async)
		err := client.Delete(ctx, channelID, watcherID)
		require.NoError(t, err)

		// Wait for deletion to complete
		err = client.WaitForDeleted(ctx, channelID, watcherID, 5*time.Second)
		require.NoError(t, err)
	})
}

// TestEndToEnd_Filtering tests filtering and pagination in list operations
func TestEndToEnd_Filtering(t *testing.T) {
	t.Run("FilterByMultipleCriteria", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Fatalf("Expected GET, got %s", r.Method)
			}

			// Verify query parameters
			query := r.URL.Query()
			assert.Equal(t, "my-watcher", query.Get("name"))
			assert.Equal(t, "1337", query.Get("chain_selector"))
			assert.Equal(t, "0x1234", query.Get("address"))
			assert.Equal(t, "dvp", query.Get("domain"))
			assert.Equal(t, "active", query.Get("status"))
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "5", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "my-watcher"
			response := apiClient.WatcherList{
				Data: []apiClient.Watcher{
					{
						WatcherId:     uuid.New(),
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        "active",
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Search with all filters
		chainSelector := "1337"
		name := "my-watcher"
		address := "0x1234"
		domain := "dvp"
		status := StatusActive
		limit := 10
		offset := int64(5)

		filters := ListFilters{
			Name:          &name,
			ChainSelector: &chainSelector,
			Address:       &address,
			Domain:        &domain,
			Status:        &status,
			Limit:         &limit,
			Offset:        &offset,
		}

		list, err := client.List(ctx, channelID, filters)
		require.NoError(t, err)
		assert.Len(t, list.Data, 1)
	})

	t.Run("PaginationThroughResults", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			offset := 0
			if query.Get("offset") != "" {
				offset, _ = strconv.Atoi(query.Get("offset"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Simulate paginated results
			watchers := []apiClient.Watcher{}
			for i := 0; i < 5 && offset+i < 15; i++ {
				name := "watcher-" + strconv.Itoa(offset+i)
				watchers = append(watchers, apiClient.Watcher{
					WatcherId:     uuid.New(),
					Name:          &name,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        "active",
				})
			}

			hasMore := offset+5 < 15
			response := apiClient.WatcherList{
				Data:    watchers,
				HasMore: hasMore,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()
		limit := 5

		// Fetch pages
		allWatchers := []apiClient.Watcher{}
		for offset := int64(0); offset < 15; offset += 5 {
			filters := ListFilters{
				Limit:  &limit,
				Offset: &offset,
			}
			list, err := client.List(ctx, channelID, filters)
			require.NoError(t, err)
			allWatchers = append(allWatchers, list.Data...)

			// Break if no more results
			if !list.HasMore {
				break
			}
		}

		// Should have fetched all 15 watchers across 3 pages
		assert.Len(t, allWatchers, 15)
	})
}

func TestIsTransientError(t *testing.T) {
	t.Run("TransientHTTPStatusCodes", func(t *testing.T) {
		testCases := []struct {
			name      string
			err       error
			transient bool
		}{
			// 5xx errors are transient
			{
				name:      "500InternalServerError",
				err:       fmt.Errorf("failed to get watcher: unexpected status code 500"),
				transient: true,
			},
			{
				name:      "502BadGateway",
				err:       fmt.Errorf("request failed: status code: 502"),
				transient: true,
			},
			{
				name:      "503ServiceUnavailable",
				err:       fmt.Errorf("failed to delete watcher: unexpected status code: 503"),
				transient: true,
			},
			{
				name:      "504GatewayTimeout",
				err:       fmt.Errorf("status code 504"),
				transient: true,
			},
			// 429 is transient (rate limiting)
			{
				name:      "429TooManyRequests",
				err:       fmt.Errorf("rate limited: status code 429"),
				transient: true,
			},
			// 4xx errors (except 429) are permanent
			{
				name:      "400BadRequest",
				err:       fmt.Errorf("invalid request: status code 400"),
				transient: false,
			},
			{
				name:      "404NotFound",
				err:       fmt.Errorf("watcher not found: status code 404"),
				transient: false,
			},
			{
				name:      "409Conflict",
				err:       fmt.Errorf("conflict: status code 409"),
				transient: false,
			},
			// 2xx/3xx are not errors, but if they appear in error messages, treat as permanent
			{
				name:      "200OK",
				err:       fmt.Errorf("unexpected success: status code 200"),
				transient: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isTransientError(tc.err)
				assert.Equal(t, tc.transient, result,
					"Expected isTransientError(%v) to be %v", tc.err, tc.transient)
			})
		}
	})

	t.Run("NetworkErrors", func(t *testing.T) {
		testCases := []struct {
			name      string
			err       error
			transient bool
		}{
			{
				name:      "ConnectionRefused",
				err:       fmt.Errorf("dial tcp: connection refused"),
				transient: true,
			},
			{
				name:      "ConnectionReset",
				err:       fmt.Errorf("read tcp: connection reset by peer"),
				transient: true,
			},
			{
				name:      "Timeout",
				err:       fmt.Errorf("request timeout"),
				transient: true,
			},
			{
				name:      "EOF",
				err:       fmt.Errorf("unexpected EOF"),
				transient: true,
			},
			{
				name:      "BrokenPipe",
				err:       fmt.Errorf("write: broken pipe"),
				transient: true,
			},
			{
				name:      "NoSuchHost",
				err:       fmt.Errorf("dial tcp: lookup api.example.com: no such host"),
				transient: true,
			},
			{
				name:      "NetworkUnreachable",
				err:       fmt.Errorf("network is unreachable"),
				transient: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isTransientError(tc.err)
				assert.Equal(t, tc.transient, result,
					"Expected isTransientError(%v) to be %v", tc.err, tc.transient)
			})
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Validation errors should always be permanent
		testCases := []error{
			ErrChannelIDRequired,
			ErrWatcherIDRequired,
			ErrNameRequired,
			ErrDomainRequired,
			ErrAddressRequired,
			ErrEventsRequired,
			ErrABIRequired,
		}

		for _, err := range testCases {
			t.Run(err.Error(), func(t *testing.T) {
				result := isTransientError(err)
				assert.False(t, result, "Validation errors should be permanent")
			})
		}
	})

	t.Run("ContextErrors", func(t *testing.T) {
		// Context errors should always be permanent
		testCases := []struct {
			name string
			err  error
		}{
			{
				name: "ContextCanceled",
				err:  context.Canceled,
			},
			{
				name: "ContextDeadlineExceeded",
				err:  context.DeadlineExceeded,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isTransientError(tc.err)
				assert.False(t, result, "Context errors should be permanent")
			})
		}
	})

	t.Run("UnknownErrors", func(t *testing.T) {
		// Unknown errors should be treated as permanent (fail fast)
		result := isTransientError(fmt.Errorf("some unknown error"))
		assert.False(t, result, "Unknown errors should be permanent to fail fast")
	})
}

func TestIsTransientStatusCode(t *testing.T) {
	testCases := []struct {
		statusCode int
		transient  bool
		name       string
	}{
		// 5xx codes are transient
		{statusCode: 500, transient: true, name: "500 Internal Server Error"},
		{statusCode: 501, transient: true, name: "501 Not Implemented"},
		{statusCode: 502, transient: true, name: "502 Bad Gateway"},
		{statusCode: 503, transient: true, name: "503 Service Unavailable"},
		{statusCode: 504, transient: true, name: "504 Gateway Timeout"},
		{statusCode: 599, transient: true, name: "599 Edge case"},

		// 429 is transient (rate limiting)
		{statusCode: 429, transient: true, name: "429 Too Many Requests"},

		// 4xx codes (except 429) are permanent
		{statusCode: 400, transient: false, name: "400 Bad Request"},
		{statusCode: 401, transient: false, name: "401 Unauthorized"},
		{statusCode: 403, transient: false, name: "403 Forbidden"},
		{statusCode: 404, transient: false, name: "404 Not Found"},
		{statusCode: 409, transient: false, name: "409 Conflict"},
		{statusCode: 422, transient: false, name: "422 Unprocessable Entity"},

		// 2xx/3xx are not transient (shouldn't appear in errors anyway)
		{statusCode: 200, transient: false, name: "200 OK"},
		{statusCode: 201, transient: false, name: "201 Created"},
		{statusCode: 301, transient: false, name: "301 Moved Permanently"},
		{statusCode: 302, transient: false, name: "302 Found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTransientStatusCode(tc.statusCode)
			assert.Equal(t, tc.transient, result,
				"Expected isTransientStatusCode(%d) to be %v", tc.statusCode, tc.transient)
		})
	}
}
