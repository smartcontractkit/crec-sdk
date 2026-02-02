// Package crec provides a unified SDK for interacting with the CREC system.
//
// The SDK follows a resource-oriented design with sub-clients for each domain:
//
//	client, err := crec.NewClient("https://api.crec.example.com", "your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a channel
//	channel, err := client.Channels.Create(ctx, channels.CreateInput{Name: "my-channel"})
//
//	// Sign and send an operation
//	op, err := client.Transact.ExecuteOperation(ctx, signer, operation, chainSelector)
//
//	// Poll for events
//	events, hasMore, err := client.Events.Poll(ctx, channelID, params)
package crec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/channels"
	"github.com/smartcontractkit/crec-sdk/events"
	"github.com/smartcontractkit/crec-sdk/transact"
	"github.com/smartcontractkit/crec-sdk/wallets"
	"github.com/smartcontractkit/crec-sdk/watchers"
)

// APIClient is the underlying HTTP client for the CREC API.
// This type is exported to allow users to create individual sub-clients
// without using the full Client.
type APIClient = apiClient.ClientWithResponses

// Client initialization errors
var (
	// ErrBaseURLRequired is returned when the base URL is empty.
	ErrBaseURLRequired = errors.New("base URL is required")

	// ErrAPIKeyRequired is returned when the API key is empty.
	ErrAPIKeyRequired = errors.New("API key is required")

	// ErrInvalidEventVerificationConfig is returned when event verification is misconfigured.
	ErrInvalidEventVerificationConfig = errors.New("minRequiredSignatures must be > 0 when validSigners are provided")

	// ErrListNetworks is returned when listing networks fails.
	ErrListNetworks = errors.New("failed to list networks")
)

// NewAPIClient creates an authenticated CREC API client that can be used
// to create individual sub-clients without instantiating the full Client.
//
// This is useful when you only need a subset of the SDK's functionality.
//
// Example:
//
//	api, err := crec.NewAPIClient("https://api.crec.example.com", "your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create only the channels sub-client
//	channelsClient, err := channels.NewClient(&channels.Options{APIClient: api})
func NewAPIClient(baseURL, apiKey string, opts ...Option) (*APIClient, error) {
	if baseURL == "" {
		return nil, ErrBaseURLRequired
	}
	if apiKey == "" {
		return nil, ErrAPIKeyRequired
	}

	cfg := &clientConfig{
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	apiKeyHeaderEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", apiKey)
		return nil
	}

	return apiClient.NewClientWithResponses(
		baseURL,
		apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
		apiClient.WithHTTPClient(cfg.httpClient),
	)
}

// Client is the main entry point for the CREC SDK.
// It provides access to all sub-clients for interacting with different parts of the CREC system.
type Client struct {
	// Channels provides operations for managing CREC channels.
	Channels *channels.Client

	// Events provides operations for polling and verifying events from CREC.
	Events *events.Client

	// Transact provides operations for signing and sending operations to CREC.
	Transact *transact.Client

	// Wallets provides operations for managing CREC Smart Wallets.
	Wallets *wallets.Client

	// Watchers provides operations for managing CREC watchers.
	Watchers *watchers.Client

	// apiClient is the underlying CREC API client
	apiClient *apiClient.ClientWithResponses

	// logger is used for logging throughout the SDK
	logger *slog.Logger
}

// NewClient creates a new CREC SDK client with the provided base URL and API key.
//
// Parameters:
//   - baseURL: The base URL of the CREC API (e.g., "https://api.crec.example.com")
//   - apiKey: The API key for authenticating with the CREC API
//   - opts: Optional configuration options (see Option for available options)
//
// Returns a configured Client or an error if initialization fails.
func NewClient(baseURL, apiKey string, opts ...Option) (*Client, error) {
	// Apply default configuration
	cfg := &clientConfig{
		httpClient: http.DefaultClient,
	}

	// Apply provided options
	for _, opt := range opts {
		opt(cfg)
	}

	// Apply default event verification if not disabled and not custom configured
	if !cfg.disableEventVerification && len(cfg.validSigners) == 0 {
		cfg.validSigners = DefaultValidSigners
		cfg.minRequiredSignatures = DefaultMinRequiredSignatures
	}

	// Validate event verification configuration
	if len(cfg.validSigners) > 0 && cfg.minRequiredSignatures <= 0 {
		return nil, ErrInvalidEventVerificationConfig
	}

	api, err := NewAPIClient(baseURL, apiKey, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	logger := cfg.logger
	if logger == nil {
		logger = slog.Default()
	}

	client := &Client{
		apiClient: api,
		logger:    logger,
	}

	// Initialize sub-clients
	if err := client.initSubClients(cfg); err != nil {
		return nil, err
	}

	return client, nil
}

// initSubClients initializes all sub-clients with the appropriate configuration.
func (c *Client) initSubClients(cfg *clientConfig) error {
	var err error

	// Initialize Channels client
	c.Channels, err = channels.NewClient(&channels.Options{
		Logger:    c.logger,
		APIClient: c.apiClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create channels client: %w", err)
	}

	// Initialize Events client
	c.Events, err = events.NewClient(&events.Options{
		Logger:                c.logger,
		CRECClient:            c.apiClient,
		MinRequiredSignatures: cfg.minRequiredSignatures,
		ValidSigners:          cfg.validSigners,
	})
	if err != nil {
		return fmt.Errorf("failed to create events client: %w", err)
	}

	// Initialize Transact client
	c.Transact, err = transact.NewClient(&transact.Options{
		Logger:     c.logger,
		CRECClient: c.apiClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create transact client: %w", err)
	}

	// Initialize Wallets client
	c.Wallets, err = wallets.NewClient(&wallets.Options{
		Logger:    c.logger,
		APIClient: c.apiClient,
	})
	if err != nil {
		return fmt.Errorf("failed to create wallets client: %w", err)
	}

	// Initialize Watchers client
	c.Watchers, err = watchers.NewClient(&watchers.Options{
		Logger:                    c.logger,
		APIClient:                 c.apiClient,
		PollInterval:              cfg.watcherPollInterval,
		EventualConsistencyWindow: cfg.watcherEventualConsistencyWindow,
	})
	if err != nil {
		return fmt.Errorf("failed to create watchers client: %w", err)
	}

	return nil
}

// ListNetworks returns the list of available networks supported by the CREC platform.
// It delegates to the underlying API client (GET /networks) with no extra SDK logic.
//
// Parameters:
//   - ctx: The context for the request.
//
// Returns the list of networks, a boolean indicating if there are more results (HasMore),
// and an error if the request fails.
func (c *Client) ListNetworks(ctx context.Context) ([]apiClient.Network, bool, error) {
	resp, err := c.apiClient.GetNetworksWithResponse(ctx)
	if err != nil {
		c.logger.Error("Failed to list networks", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListNetworks, err)
	}
	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when listing networks",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w (status code %d)", ErrListNetworks, resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, false, ErrListNetworks
	}
	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}
