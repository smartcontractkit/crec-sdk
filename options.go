package crec

import (
	"log/slog"
	"net/http"
	"time"
)

// clientConfig holds the internal configuration for the Client.
type clientConfig struct {
	httpClient                       *http.Client
	logger                           *slog.Logger
	minRequiredSignatures            int
	validSigners                     []string
	watcherPollInterval              time.Duration
	watcherEventualConsistencyWindow time.Duration
}

// Option is a functional option for configuring the Client.
type Option func(*clientConfig)

// WithHTTPClient sets a custom HTTP client for API requests.
// If not provided, http.DefaultClient is used.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}

// WithLogger sets a custom logger for the SDK.
// If not provided, slog.Default() is used.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *clientConfig) {
		cfg.logger = logger
	}
}

// WithEventVerification configures the Events client for verifying event signatures.
//
// Parameters:
//   - minRequiredSignatures: Minimum number of valid signatures required to verify an event
//   - validSigners: List of valid signer addresses (as hex strings)
func WithEventVerification(minRequiredSignatures int, validSigners []string) Option {
	return func(cfg *clientConfig) {
		cfg.minRequiredSignatures = minRequiredSignatures
		cfg.validSigners = validSigners
	}
}

// WithWatcherPolling configures the Watchers client polling behavior.
//
// Parameters:
//   - pollInterval: Duration between polling attempts when waiting for watcher state changes
//   - eventualConsistencyWindow: Duration to tolerate 404 errors after creation due to eventual consistency
func WithWatcherPolling(pollInterval, eventualConsistencyWindow time.Duration) Option {
	return func(cfg *clientConfig) {
		cfg.watcherPollInterval = pollInterval
		cfg.watcherEventualConsistencyWindow = eventualConsistencyWindow
	}
}
