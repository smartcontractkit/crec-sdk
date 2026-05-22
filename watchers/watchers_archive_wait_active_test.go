package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func TestClient_Archive(t *testing.T) {
	t.Run("SuccessAsync", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers/" + watcherID.String()
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var updateReq apiClient.UpdateWatcher
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			require.NotNil(t, updateReq.Status)
			assert.Equal(t, apiClient.WatcherStatusArchived, *updateReq.Status)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        apiClient.WatcherStatusArchiving,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Archive(context.Background(), channelID, watcherID)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, apiClient.WatcherStatusArchiving, watcher.Status)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Archive(context.Background(), uuid.Nil, uuid.New())

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

		watcher, err := client.Archive(context.Background(), uuid.New(), uuid.Nil)

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
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"error": "Watcher not found",
			}))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Archive(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})

	t.Run("ServerError", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			}))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Archive(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrArchiveWatcher), "Expected ErrArchiveWatcher, got: %v", err)
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
				Status:        apiClient.WatcherStatusActive,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, apiClient.WatcherStatusActive, watcher.Status)
	})

	t.Run("SuccessAfterPolling", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		callCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			var status apiClient.WatcherStatus
			if callCount < 3 {
				status = apiClient.WatcherStatusPending
			} else {
				status = apiClient.WatcherStatusActive
			}

			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        status,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 10*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, apiClient.WatcherStatusActive, watcher.Status)
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
				Status:        apiClient.WatcherStatusFailed,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
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
				Status:        apiClient.WatcherStatusPending,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
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

	t.Run("ArchivingStatus", func(t *testing.T) {
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
				Status:        apiClient.WatcherStatusArchiving,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherIsArchiving), "Expected ErrWatcherIsArchiving, got: %v", err)
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
				Status:        apiClient.WatcherStatusPending,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
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
				_, writeErr := w.Write([]byte(`{"error": "service temporarily unavailable"}`))
				require.NoError(t, writeErr)
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
				Status:        apiClient.WatcherStatusActive,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, apiClient.WatcherStatusActive, watcher.Status)
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
			_, writeErr := w.Write([]byte(`{"error": "service temporarily unavailable"}`))
			require.NoError(t, writeErr)
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
			_, writeErr := w.Write([]byte(`{"error": "bad request"}`))
			require.NoError(t, writeErr)
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
				_, writeErr := w.Write([]byte(`{"error": "internal server error"}`))
				require.NoError(t, writeErr)
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
				Status:        apiClient.WatcherStatusActive,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.WaitForActive(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, apiClient.WatcherStatusActive, watcher.Status)
		assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried after 500 errors")
	})
}
