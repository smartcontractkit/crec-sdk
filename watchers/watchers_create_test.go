package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func TestClient_CreateWithService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "test-watcher"
		service := "dvp"
		chainSelector := "1337"
		address := "0x1234567890abcdef"

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)

			serviceWatcher, err := createReq.AsCreateWatcherWithService()
			require.NoError(t, err)
			assert.Equal(t, watcherName, serviceWatcher.Name)
			assert.Equal(t, service, serviceWatcher.Service)
			assert.Equal(t, chainSelector, serviceWatcher.ChainSelector)
			assert.Equal(t, address, serviceWatcher.Address)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &watcherName,
				ChainSelector: chainSelector,
				Address:       address,
				Status:        apiClient.WatcherStatusPending,
			}
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithService(context.Background(), channelID, CreateWithServiceInput{
			Name:          watcherName,
			Service:       service,
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
		assert.Equal(t, apiClient.WatcherStatusPending, watcher.Status)
	})

	t.Run("Success_with_confidence_level", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "test-watcher"
		service := "dvp"
		chainSelector := "1337"
		address := "0x1234567890abcdef"
		conf := apiClient.Safe

		handler := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			require.NoError(t, json.Unmarshal(body, &createReq))

			serviceWatcher, err := createReq.AsCreateWatcherWithService()
			require.NoError(t, err)
			require.NotNil(t, serviceWatcher.ConfidenceLevel)
			assert.Equal(t, conf, *serviceWatcher.ConfidenceLevel)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Watcher{
				WatcherId:       watcherID,
				Name:            &watcherName,
				ChainSelector:   chainSelector,
				Address:         address,
				Status:          apiClient.WatcherStatusPending,
				ConfidenceLevel: conf,
			}
			_ = json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithService(context.Background(), channelID, CreateWithServiceInput{
			Name:            watcherName,
			Service:         service,
			ChainSelector:   chainSelector,
			Address:         address,
			Events:          []string{"TestEvent"},
			ConfidenceLevel: &conf,
		})

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, conf, watcher.ConfidenceLevel)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithService(context.Background(), uuid.Nil, CreateWithServiceInput{
			Name:          name,
			Service:       "dvp",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyService", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty service")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithService(context.Background(), uuid.New(), CreateWithServiceInput{
			Name:          name,
			Service:       "",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrServiceRequired), "Expected ErrServiceRequired, got: %v", err)
	})

	t.Run("EmptyAddress", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty address")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithService(context.Background(), uuid.New(), CreateWithServiceInput{
			Name:          name,
			Service:       "dvp",
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
			err := json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid watcher configuration",
			})
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		name := "test-watcher"
		watcher, err := client.CreateWithService(context.Background(), channelID, CreateWithServiceInput{
			Name:          name,
			Service:       "dvp",
			ChainSelector: "1337",
			Address:       "0x1234",
			Events:        []string{"TestEvent"},
		})

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrCreateWatcherService), "Expected ErrCreateWatcherService, got: %v", err)
	})
}

func TestClient_CreateWithService_watcherNameValidation(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected HTTP request during name validation")
	}
	client, server := setupTestClient(t, handler)
	defer server.Close()

	channelID := uuid.New()
	tests := []struct {
		name        string
		watcherName string
		wantErr     error
	}{
		{"empty", "", ErrNameRequired},
		{"whitespace_only", "   ", ErrNameRequired},
		{"too_short", "abc", ErrWatcherNameTooShort},
		{"trim_then_short", "  ab ", ErrWatcherNameTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateWithService(context.Background(), channelID, CreateWithServiceInput{
				Name:          tt.watcherName,
				Service:       "dvp",
				ChainSelector: "1337",
				Address:       "0x1234",
				Events:        []string{"TestEvent"},
			})
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr), "want %v, got %v", tt.wantErr, err)
		})
	}
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
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)

			abiWatcher, err := createReq.AsCreateWatcherWithABI()
			require.NoError(t, err)
			assert.Equal(t, watcherName, abiWatcher.Name)
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
				Status:        apiClient.WatcherStatusPending,
			}
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithABI(context.Background(), channelID, CreateWithABIInput{
			Name:          watcherName,
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
		assert.Equal(t, apiClient.WatcherStatusPending, watcher.Status)
	})

	t.Run("Success_with_confidence_level", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "test-watcher-abi"
		chainSelector := "1337"
		address := "0x1234567890abcdef"
		conf := apiClient.Finalized

		abiFixture := EventABI{
			Name:      "Transfer",
			Type:      "event",
			Anonymous: false,
			Inputs: []EventABIInput{
				{Indexed: true, Name: "from", Type: "address"},
				{Indexed: true, Name: "to", Type: "address"},
				{Indexed: false, Name: "value", Type: "uint256"},
			},
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWatcher
			require.NoError(t, json.Unmarshal(body, &createReq))

			abiWatcher, err := createReq.AsCreateWatcherWithABI()
			require.NoError(t, err)
			require.NotNil(t, abiWatcher.ConfidenceLevel)
			assert.Equal(t, conf, *abiWatcher.ConfidenceLevel)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Watcher{
				WatcherId:       watcherID,
				Name:            &watcherName,
				ChainSelector:   chainSelector,
				Address:         address,
				Status:          apiClient.WatcherStatusPending,
				ConfidenceLevel: conf,
			}
			_ = json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithABI(context.Background(), channelID, CreateWithABIInput{
			Name:            watcherName,
			ChainSelector:   chainSelector,
			Address:         address,
			Events:          []string{"Transfer"},
			ABI:             []EventABI{abiFixture},
			ConfidenceLevel: &conf,
		})

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, conf, watcher.ConfidenceLevel)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.CreateWithABI(context.Background(), uuid.Nil, CreateWithABIInput{
			Name:          "test-watcher",
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
			Name:          name,
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

		watcher, err := client.CreateWithABI(context.Background(), uuid.New(), CreateWithABIInput{
			Name:          "test-watcher",
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
		assert.ErrorIs(t, err, ErrInvalidABIType)
	})

	t.Run("EventNotInABI", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		watcher, err := client.CreateWithABI(context.Background(), uuid.New(), CreateWithABIInput{
			Name:          "test-watcher",
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
		assert.ErrorIs(t, err, ErrEventNotInABI)
	})
}

func TestClient_CreateWithABI_watcherNameValidation(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("unexpected HTTP request during name validation")
	}
	client, server := setupTestClient(t, handler)
	defer server.Close()

	minimalABI := []EventABI{
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
	}

	channelID := uuid.New()
	tests := []struct {
		name        string
		watcherName string
		wantErr     error
	}{
		{"empty", "", ErrNameRequired},
		{"whitespace_only", "\t", ErrNameRequired},
		{"too_short", "abc", ErrWatcherNameTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateWithABI(context.Background(), channelID, CreateWithABIInput{
				Name:          tt.watcherName,
				ChainSelector: "1337",
				Address:       "0x1234",
				Events:        []string{"Transfer"},
				ABI:           minimalABI,
			})
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr), "want %v, got %v", tt.wantErr, err)
		})
	}
}
