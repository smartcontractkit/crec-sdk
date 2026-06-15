package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func TestClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		watcher1ID := uuid.New()
		watcher2ID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/channels/" + channelID.String() + "/watchers"
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			query := r.URL.Query()
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "0", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name1 := "watcher-1"
			name2 := "watcher-2"
			response := apiClient.WatcherList{
				Data: []apiClient.WatcherSummary{
					{
						WatcherId:     watcher1ID,
						Name:          &name1,
						ChainSelector: "1337",
						Address:       "0x1111",
						Status:        apiClient.WatcherStatusActive,
						ChannelId:     channelID,
						CreatedAt:     time.Now().Unix(),
						DonFamily:     "zone-a",
					},
					{
						WatcherId:     watcher2ID,
						Name:          &name2,
						ChainSelector: "1337",
						Address:       "0x2222",
						Status:        apiClient.WatcherStatusActive,
						ChannelId:     channelID,
						CreatedAt:     time.Now().Unix(),
						DonFamily:     "zone-a",
					},
				},
				HasMore: false,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
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
		status := apiClient.WatcherStatusActive

		handler := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, name, query.Get("name"))
			assert.Equal(t, string(status), query.Get("status"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.WatcherList{
				Data: []apiClient.WatcherSummary{
					{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1111",
						Status:        apiClient.WatcherStatusActive,
						ChannelId:     channelID,
						CreatedAt:     time.Now().Unix(),
						DonFamily:     "zone-a",
					},
				},
				HasMore: false,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		statusFilter := []apiClient.WatcherStatus{status}
		result, err := client.List(context.Background(), channelID, ListFilters{
			Name:   &name,
			Status: &statusFilter,
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		require.NotNil(t, result.Data[0].Name)
		assert.Equal(t, name, *result.Data[0].Name)
		assert.Equal(t, apiClient.WatcherStatusActive, result.Data[0].Status)
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
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			}))
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
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

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

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		require.NotNil(t, watcher.Name)
		assert.Equal(t, "test-watcher", *watcher.Name)
		assert.Equal(t, apiClient.WatcherStatusActive, watcher.Status)
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
				Status:        apiClient.WatcherStatusActive,
				CreatedAt:     1704067200, // 2024-01-01 00:00:00 UTC
				Events:        []string{"Transfer"},
				DonFamily:     "zone-a",
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.NoError(t, err)
		assert.NotNil(t, watcher)
		assert.Equal(t, watcherID, watcher.WatcherId)
		assert.Equal(t, channelID, watcher.ChannelId)
		assert.Equal(t, "zone-a", watcher.DonFamily)
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
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"message": "watcher with ID " + watcherID.String() + " not found",
				"type":    "NOT_FOUND",
				"code":    "WATCHER_NOT_FOUND",
			}))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrWatcherNotFound), "Expected ErrWatcherNotFound, got: %v", err)
	})

	t.Run("ChannelNotFound", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"message": "channel with ID " + channelID.String() + " not found",
				"type":    "NOT_FOUND",
				"code":    "CHANNEL_NOT_FOUND",
			}))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		watcher, err := client.Get(context.Background(), channelID, watcherID)

		require.Error(t, err)
		assert.Nil(t, watcher)
		assert.True(t, errors.Is(err, ErrChannelNotFound), "Expected ErrChannelNotFound, got: %v", err)
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
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var updateReq apiClient.UpdateWatcher
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			require.NotNil(t, updateReq.Name)
			assert.Equal(t, newName, *updateReq.Name)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Watcher{
				WatcherId:     watcherID,
				Name:          &newName,
				ChainSelector: "1337",
				Address:       "0x1234",
				Status:        apiClient.WatcherStatusActive,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
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
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"message": "watcher with ID " + watcherID.String() + " not found",
				"type":    "NOT_FOUND",
				"code":    "WATCHER_NOT_FOUND",
			}))
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
