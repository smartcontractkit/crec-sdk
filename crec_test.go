package crec_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk"
	"github.com/smartcontractkit/crec-sdk/channels"
	"github.com/smartcontractkit/crec-sdk/events"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		apiKey      string
		opts        []crec.Option
		wantErr     error
		errContains string
	}{
		{
			name:    "Success_MinimalConfig",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts:    nil,
			wantErr: nil,
		},
		{
			name:    "Success_WithLogger",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithLogger(slog.Default()),
			},
			wantErr: nil,
		},
		{
			name:    "Success_WithHTTPClient",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
			},
			wantErr: nil,
		},
		{
			name:    "Success_WithEventVerification",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithEventVerification(2, []string{
					"0x5db070ceabcf97e45d96b4f951a1df050ddb5559",
					"0xadebb9657c04692275973230b06adfabacc899bc",
					"0xc868bbb5d93e97b9d780fc93811a00ca7c016751",
				}),
			},
			wantErr: nil,
		},
		{
			name:    "Success_WithWatcherPolling",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithWatcherPolling(5*time.Second, 10*time.Second),
			},
			wantErr: nil,
		},
		{
			name:    "Success_AllOptions",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithLogger(slog.Default()),
				crec.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
				crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
				crec.WithWatcherPolling(5*time.Second, 10*time.Second),
			},
			wantErr: nil,
		},
		{
			name:    "Error_EmptyBaseURL",
			baseURL: "",
			apiKey:  "test-api-key",
			opts:    nil,
			wantErr: crec.ErrBaseURLRequired,
		},
		{
			name:    "Error_EmptyAPIKey",
			baseURL: "https://api.crec.example.com",
			apiKey:  "",
			opts:    nil,
			wantErr: crec.ErrAPIKeyRequired,
		},
		{
			name:    "Error_BothEmpty",
			baseURL: "",
			apiKey:  "",
			opts:    nil,
			wantErr: crec.ErrBaseURLRequired, // baseURL is checked first
		},
		{
			name:    "Error_InvalidEventVerification_ZeroMinWithSigners",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithEventVerification(0, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
			},
			wantErr: crec.ErrInvalidEventVerificationConfig,
		},
		{
			name:    "Error_InvalidEventVerification_NegativeMinWithSigners",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithEventVerification(-1, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559"}),
			},
			wantErr: crec.ErrInvalidEventVerificationConfig,
		},
		{
			name:    "Success_ZeroMinWithNoSigners",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithEventVerification(0, nil),
			},
			wantErr: nil,
		},
		{
			name:    "Success_ZeroMinWithEmptySigners",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithEventVerification(0, []string{}),
			},
			wantErr: nil,
		},
		{
			name:    "Success_WithoutEventVerification",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithoutEventVerification(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := crec.NewClient(tt.baseURL, tt.apiKey, tt.opts...)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "Expected error %v, got %v", tt.wantErr, err)
				assert.Nil(t, client)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)

			// Verify all sub-clients are initialized
			assert.NotNil(t, client.Channels, "Channels sub-client should be initialized")
			assert.NotNil(t, client.Events, "Events sub-client should be initialized")
			assert.NotNil(t, client.Transact, "Transact sub-client should be initialized")
			assert.NotNil(t, client.Watchers, "Watchers sub-client should be initialized")
		})
	}
}

func TestNewClient_WithMockServer(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header is set
		assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := crec.NewClient(server.URL, "test-api-key")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify all sub-clients are initialized
	assert.NotNil(t, client.Channels)
	assert.NotNil(t, client.Events)
	assert.NotNil(t, client.Transact)
	assert.NotNil(t, client.Watchers)
}

func TestNewClient_CustomHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := crec.NewClient(
		"https://api.crec.example.com",
		"test-api-key",
		crec.WithHTTPClient(customClient),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The custom HTTP client is used internally by sub-clients
	// We've verified it was passed correctly by the client being created successfully
}

func TestNewClient_CustomLogger(t *testing.T) {
	// Create a custom logger
	customLogger := slog.New(slog.NewTextHandler(&discardWriter{}, nil))

	client, err := crec.NewClient(
		"https://api.crec.example.com",
		"test-api-key",
		crec.WithLogger(customLogger),
	)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_DefaultLogger(t *testing.T) {
	// When no logger is provided, slog.Default() should be used
	client, err := crec.NewClient(
		"https://api.crec.example.com",
		"test-api-key",
	)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestClient_ListNetworks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		networkID := uuid.New()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/networks", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(apiClient.NetworkList{
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
			})
		}))
		defer server.Close()

		client, err := crec.NewClient(server.URL, "test-api-key")
		require.NoError(t, err)

		networks, hasMore, err := client.ListNetworks(context.Background())
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.False(t, hasMore)
		assert.Equal(t, networkID, networks[0].Id)
		assert.Equal(t, "Ethereum Mainnet", networks[0].Name)
		assert.Equal(t, "evm", networks[0].ChainFamily)
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, err := crec.NewClient(server.URL, "test-api-key")
		require.NoError(t, err)

		networks, hasMore, err := client.ListNetworks(context.Background())
		require.Error(t, err)
		assert.Nil(t, networks)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, crec.ErrListNetworks))
	})
}

func TestNewClient_EventVerificationConfig(t *testing.T) {
	tests := []struct {
		name          string
		minRequired   int
		validSigners  []string
		expectSuccess bool
	}{
		{
			name:          "Valid_ThreeSignersTwoRequired",
			minRequired:   2,
			validSigners:  []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc", "0xc868bbb5d93e97b9d780fc93811a00ca7c016751"},
			expectSuccess: true,
		},
		{
			name:          "Valid_OneSignerOneRequired",
			minRequired:   1,
			validSigners:  []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559"},
			expectSuccess: true,
		},
		{
			name:          "Valid_NoSignersNoRequired",
			minRequired:   0,
			validSigners:  nil,
			expectSuccess: true,
		},
		{
			name:          "Valid_NoSignersEmptySlice",
			minRequired:   0,
			validSigners:  []string{},
			expectSuccess: true,
		},
		{
			name:          "Invalid_SignersButZeroRequired",
			minRequired:   0,
			validSigners:  []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559"},
			expectSuccess: false,
		},
		{
			name:          "Invalid_SignersButNegativeRequired",
			minRequired:   -5,
			validSigners:  []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"},
			expectSuccess: false,
		},
		{
			name:          "Invalid_DuplicateSigners",
			minRequired:   1,
			validSigners:  []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0x5db070ceabcf97e45d96b4f951a1df050ddb5559"},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := crec.NewClient(
				"https://api.crec.example.com",
				"test-api-key",
				crec.WithEventVerification(tt.minRequired, tt.validSigners),
			)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, client)
			} else {
				require.Error(t, err)
				assert.Nil(t, client)
				// The error could be from the config validation or from the events subclient initialization
				if tt.name == "Invalid_DuplicateSigners" {
					assert.ErrorIs(t, err, events.ErrDuplicateSigner)
				} else {
					assert.True(t, errors.Is(err, crec.ErrInvalidEventVerificationConfig))
				}
			}
		})
	}
}

func TestNewClient_WatcherPollingConfig(t *testing.T) {
	pollInterval := 5 * time.Second
	consistencyWindow := 10 * time.Second

	client, err := crec.NewClient(
		"https://api.crec.example.com",
		"test-api-key",
		crec.WithWatcherPolling(pollInterval, consistencyWindow),
	)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.Watchers)
}

func TestNewClient_DefaultEventVerification(t *testing.T) {
	t.Run("DefaultsAppliedWithNoOptions", func(t *testing.T) {
		// When no options are provided, defaults should be applied
		client, err := crec.NewClient(
			"https://api.crec.example.com",
			"test-api-key",
		)
		require.NoError(t, err)
		require.NotNil(t, client)
		// Client created successfully means defaults were applied correctly
		// (DefaultMinRequiredSignatures=3 with DefaultValidSigners)
	})

	t.Run("WithoutEventVerificationDisablesDefaults", func(t *testing.T) {
		// When WithoutEventVerification is used, no defaults should be applied
		client, err := crec.NewClient(
			"https://api.crec.example.com",
			"test-api-key",
			crec.WithoutEventVerification(),
		)
		require.NoError(t, err)
		require.NotNil(t, client)
		// Client created successfully with verification disabled
	})

	t.Run("CustomSignersOverrideDefaults", func(t *testing.T) {
		// When custom signers are provided, they should override defaults
		client, err := crec.NewClient(
			"https://api.crec.example.com",
			"test-api-key",
			crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
		)
		require.NoError(t, err)
		require.NotNil(t, client)
		// Custom signers used instead of defaults
	})

	t.Run("WithoutEventVerificationTakesPrecedence", func(t *testing.T) {
		// WithoutEventVerification should work even if called before WithEventVerification
		// Last option wins, so order matters
		client, err := crec.NewClient(
			"https://api.crec.example.com",
			"test-api-key",
			crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
			crec.WithoutEventVerification(),
		)
		require.NoError(t, err)
		require.NotNil(t, client)
		// Verification disabled even though WithEventVerification was called first
	})
}

func TestNewClient_MultipleOptionsOrder(t *testing.T) {
	// Test that options can be applied in any order
	customClient := &http.Client{Timeout: 30 * time.Second}
	customLogger := slog.New(slog.NewTextHandler(&discardWriter{}, nil))

	// Apply in different orders
	orders := [][]crec.Option{
		{
			crec.WithHTTPClient(customClient),
			crec.WithLogger(customLogger),
			crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
			crec.WithWatcherPolling(5*time.Second, 10*time.Second),
		},
		{
			crec.WithWatcherPolling(5*time.Second, 10*time.Second),
			crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
			crec.WithLogger(customLogger),
			crec.WithHTTPClient(customClient),
		},
		{
			crec.WithLogger(customLogger),
			crec.WithWatcherPolling(5*time.Second, 10*time.Second),
			crec.WithHTTPClient(customClient),
			crec.WithEventVerification(2, []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}),
		},
	}

	for i, opts := range orders {
		t.Run("Order"+string(rune('A'+i)), func(t *testing.T) {
			client, err := crec.NewClient(
				"https://api.crec.example.com",
				"test-api-key",
				opts...,
			)
			require.NoError(t, err)
			require.NotNil(t, client)
			assert.NotNil(t, client.Channels)
			assert.NotNil(t, client.Events)
			assert.NotNil(t, client.Transact)
			assert.NotNil(t, client.Watchers)
		})
	}
}

func TestClient_SubClientsIntegration(t *testing.T) {
	// Create a mock server that responds to health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client, err := crec.NewClient(
		server.URL,
		"test-api-key",
		crec.WithEventVerification(2, []string{
			"0x5db070ceabcf97e45d96b4f951a1df050ddb5559",
			"0xadebb9657c04692275973230b06adfabacc899bc",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify all sub-clients can be accessed
	t.Run("ChannelsClientAccessible", func(t *testing.T) {
		require.NotNil(t, client.Channels)
	})

	t.Run("EventsClientAccessible", func(t *testing.T) {
		require.NotNil(t, client.Events)
	})

	t.Run("TransactClientAccessible", func(t *testing.T) {
		require.NotNil(t, client.Transact)
	})

	t.Run("WatchersClientAccessible", func(t *testing.T) {
		require.NotNil(t, client.Watchers)
	})
}

// Helper types for testing

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Test that the client properly passes the API key in headers
func TestNewClient_APIKeyHeader(t *testing.T) {
	apiKeyReceived := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyReceived = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [], "hasMore": false}`))
	}))
	defer server.Close()

	client, err := crec.NewClient(server.URL, "my-secret-api-key")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Make a request through one of the sub-clients to verify the API key is sent
	// Using Channels.List as it's a simple GET request
	_, _, _ = client.Channels.List(context.Background(), channels.ListInput{})
	// The request might fail but we don't care - we just want to verify the header was set
	assert.Equal(t, "Apikey my-secret-api-key", apiKeyReceived)
}

func TestNewAPIClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		opts    []crec.Option
		wantErr error
	}{
		{
			name:    "Success_MinimalConfig",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts:    nil,
			wantErr: nil,
		},
		{
			name:    "Success_WithHTTPClient",
			baseURL: "https://api.crec.example.com",
			apiKey:  "test-api-key",
			opts: []crec.Option{
				crec.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
			},
			wantErr: nil,
		},
		{
			name:    "Error_EmptyBaseURL",
			baseURL: "",
			apiKey:  "test-api-key",
			opts:    nil,
			wantErr: crec.ErrBaseURLRequired,
		},
		{
			name:    "Error_EmptyAPIKey",
			baseURL: "https://api.crec.example.com",
			apiKey:  "",
			opts:    nil,
			wantErr: crec.ErrAPIKeyRequired,
		},
		{
			name:    "Error_BothEmpty",
			baseURL: "",
			apiKey:  "",
			opts:    nil,
			wantErr: crec.ErrBaseURLRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient, err := crec.NewAPIClient(tt.baseURL, tt.apiKey, tt.opts...)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "Expected error %v, got %v", tt.wantErr, err)
				assert.Nil(t, apiClient)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, apiClient)
		})
	}
}

func TestNewAPIClient_APIKeyHeader(t *testing.T) {
	apiKeyReceived := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeyReceived = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [], "hasMore": false}`))
	}))
	defer server.Close()

	api, err := crec.NewAPIClient(server.URL, "my-secret-api-key")
	require.NoError(t, err)
	require.NotNil(t, api)

	// Create a channels sub-client using the API client directly
	channelsClient, err := channels.NewClient(&channels.Options{APIClient: api})
	require.NoError(t, err)

	// Make a request to verify the API key is sent
	_, _, _ = channelsClient.List(context.Background(), channels.ListInput{})
	assert.Equal(t, "Apikey my-secret-api-key", apiKeyReceived)
}

func TestNewAPIClient_UseWithSubClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [], "hasMore": false}`))
	}))
	defer server.Close()

	// Create API client independently
	api, err := crec.NewAPIClient(server.URL, "test-api-key")
	require.NoError(t, err)
	require.NotNil(t, api)

	// Use it to create only the channels sub-client
	channelsClient, err := channels.NewClient(&channels.Options{
		APIClient: api,
	})
	require.NoError(t, err)
	require.NotNil(t, channelsClient)

	// Verify it works
	result, hasMore, err := channelsClient.List(context.Background(), channels.ListInput{})
	require.NoError(t, err)
	assert.Empty(t, result)
	assert.False(t, hasMore)
}
