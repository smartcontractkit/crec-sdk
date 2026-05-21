package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func TestClient_WaitForArchived(t *testing.T) {
	t.Run("SuccessAfterPolling", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		callCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			name := "test-watcher"
			if callCount < 3 {
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &name,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusArchiving,
				}
				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					return
				}
			} else {
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &name,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusArchived,
				}
				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					return
				}
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, callCount, 3)
	})

	t.Run("SuccessImmediate", func(t *testing.T) {
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
				Status:        apiClient.WatcherStatusArchived,
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
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
				Status:        apiClient.WatcherStatusArchiving,
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 100*time.Millisecond)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWaitForArchivedTimeout), "Expected ErrWaitForArchivedTimeout, got: %v", err)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
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
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := client.WaitForArchived(ctx, channelID, watcherID, 5*time.Second)

		require.Error(t, err)
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
				Status:        apiClient.WatcherStatusActive,
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 5*time.Second)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrWatcherArchiveFailed)
	})

	t.Run("EmptyChannelID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty channel ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), uuid.Nil, uuid.New(), 5*time.Second)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChannelIDRequired), "Expected ErrChannelIDRequired, got: %v", err)
	})

	t.Run("EmptyWatcherID", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty watcher ID")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), uuid.New(), uuid.Nil, 5*time.Second)

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

			if attemptCount <= 2 {
				w.WriteHeader(http.StatusBadGateway)
				_, writeErr := w.Write([]byte(`{"error": "bad gateway"}`))
				require.NoError(t, writeErr)
				return
			}

			w.WriteHeader(http.StatusOK)
			name := "test-watcher"
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &name,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        apiClient.WatcherStatusArchived,
			}
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 5*time.Second)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried after transient errors")
	})

	t.Run("TransientErrorRetry_EventuallyTimesOut", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		attemptCount := 0

		handler := func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := w.Write([]byte(`{"error": "internal server error"}`))
			require.NoError(t, writeErr)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.WaitForArchived(context.Background(), channelID, watcherID, 200*time.Millisecond)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWaitForArchivedTimeout), "Expected timeout error, got: %v", err)
		assert.Greater(t, attemptCount, 1, "Should have retried multiple times before timeout")
	})
}
