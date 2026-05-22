package watchers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/internal/retry"
)

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	// Add API key header to all requests
	apiKeyEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Apikey test-api-key")
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
			ErrWatcherNameTooShort,
			ErrServiceRequired,
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
			result := retry.IsTransientStatusCode(tc.statusCode)
			assert.Equal(t, tc.transient, result,
				"Expected retry.IsTransientStatusCode(%d) to be %v", tc.statusCode, tc.transient)
		})
	}
}
