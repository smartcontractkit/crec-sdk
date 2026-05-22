package watchers

import (
	"context"
	"encoding/json"
	"errors"
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

// TestEndToEnd_WatcherLifecycle tests complete watcher workflows
func TestEndToEnd_WatcherLifecycle(t *testing.T) {
	t.Run("CreateWithService_WaitActive_Update_Archive", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "integration-test-watcher"
		updatedName := "updated-watcher"

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			switch {
			case r.Method == "POST" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers"):
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusPending,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))

			case r.Method == "GET" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				status := apiClient.WatcherStatusPending
				name := watcherName
				if callCount > 2 {
					status = apiClient.WatcherStatusActive
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
				require.NoError(t, json.NewEncoder(w).Encode(response))

			case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				body, _ := io.ReadAll(r.Body)
				var updateReq apiClient.UpdateWatcher
				require.NoError(t, json.Unmarshal(body, &updateReq))

				if updateReq.Status != nil && *updateReq.Status == apiClient.WatcherStatusArchived {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &updatedName,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusArchived,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				} else {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &updatedName,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusActive,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				}

			default:
				t.Fatalf("Unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		createInput := CreateWithServiceInput{
			Name:          watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Service:       "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithService(ctx, channelID, createInput)
		require.NoError(t, err)
		assert.Equal(t, watcherID, created.WatcherId)
		assert.Equal(t, apiClient.WatcherStatusPending, created.Status)

		active, err := client.WaitForActive(ctx, channelID, watcherID, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WatcherStatusActive, active.Status)

		found, err := client.Get(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, watcherID, found.WatcherId)
		assert.Equal(t, watcherName, *found.Name)

		updateInput := UpdateInput{
			Name: updatedName,
		}
		updated, err := client.Update(ctx, channelID, watcherID, updateInput)
		require.NoError(t, err)
		assert.Equal(t, updatedName, *updated.Name)

		found, err = client.Get(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, updatedName, *found.Name)

		archived, err := client.Archive(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WatcherStatusArchived, archived.Status)

		assert.GreaterOrEqual(t, callCount, 6, "Should have made at least 6 API calls")
	})

	t.Run("CreateWithABI_WaitActive_List_Archive", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "abi-test-watcher"

		callCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			switch {
			case r.Method == "POST" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers"):
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x5678",
					Status:        apiClient.WatcherStatusPending,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))

			case r.Method == "GET" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x5678",
					Status:        apiClient.WatcherStatusActive,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))

			case r.Method == "GET" && strings.Contains(r.URL.Path, "/channels/"+channelID.String()+"/watchers") && !strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusOK)
				response := apiClient.WatcherList{
					Data: []apiClient.WatcherSummary{
						{
							WatcherId:     watcherID,
							Name:          &watcherName,
							ChainSelector: "1337",
							Address:       "0x5678",
							Status:        apiClient.WatcherStatusActive,
							ChannelId:     channelID,
							CreatedAt:     time.Now().Unix(),
							DonFamily:     "zone-a",
						},
					},
					HasMore: false,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))

			case r.Method == "PATCH" && strings.Contains(r.URL.Path, "/watchers/"+watcherID.String()):
				w.WriteHeader(http.StatusAccepted)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x5678",
					Status:        apiClient.WatcherStatusArchiving,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))

			default:
				t.Fatalf("Unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		createInput := CreateWithABIInput{
			Name:          watcherName,
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

		active, err := client.WaitForActive(ctx, channelID, watcherID, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WatcherStatusActive, active.Status)

		filters := ListFilters{}
		list, err := client.List(ctx, channelID, filters)
		require.NoError(t, err)
		assert.Len(t, list.Data, 1)
		assert.Equal(t, watcherID, list.Data[0].WatcherId)

		archived, err := client.Archive(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WatcherStatusArchiving, archived.Status)

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
				_, writeErr := w.Write([]byte(`{"error": "invalid chain selector"}`))
				require.NoError(t, writeErr)
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Try to create with invalid data
		createInput := CreateWithServiceInput{
			Name:          "chain-invalid-test",
			ChainSelector: "0", // Invalid
			Address:       "0x1234",
			Service:       "dvp",
			Events:        []string{"TestEvent"},
		}
		_, err := client.CreateWithService(ctx, channelID, createInput)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChainSelectorRequired), "Expected ErrChainSelectorRequired, got: %v", err)
	})

	t.Run("CreateSucceeds_ButFailsToDeploy", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "failing-watcher"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusPending,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))
			case "GET":
				// Watcher failed to deploy
				w.WriteHeader(http.StatusOK)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusFailed,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		// Create watcher
		createInput := CreateWithServiceInput{
			Name:          watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Service:       "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithService(ctx, channelID, createInput)
		require.NoError(t, err)

		// Wait for active - should fail because watcher deployment failed
		_, err = client.WaitForActive(ctx, channelID, created.WatcherId, 5*time.Second)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherDeploymentFailed), "Expected ErrWatcherDeploymentFailed, got: %v", err)
	})

	t.Run("WatcherIsArchived_WhileWaitingForActive", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()
		watcherName := "archived-watcher"

		getCallCount := 0
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.Method {
			case "POST":
				w.WriteHeader(http.StatusCreated)
				response := apiClient.Watcher{
					WatcherId:     watcherID,
					Name:          &watcherName,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusPending,
				}
				require.NoError(t, json.NewEncoder(w).Encode(response))
			case "GET":
				getCallCount++
				if getCallCount == 1 {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &watcherName,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusPending,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				} else {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &watcherName,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusArchived,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				}
			}
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

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
			Logger:                    logger,
			APIClient:                 crecAPIClient,
			PollInterval:              10 * time.Millisecond,
			EventualConsistencyWindow: 1 * time.Millisecond,
		})
		require.NoError(t, err)

		ctx := context.Background()

		createInput := CreateWithServiceInput{
			Name:          watcherName,
			ChainSelector: "1337",
			Address:       "0x1234",
			Service:       "dvp",
			Events:        []string{"TestEvent"},
		}
		created, err := client.CreateWithService(ctx, channelID, createInput)
		require.NoError(t, err)

		_, err = client.WaitForActive(ctx, channelID, created.WatcherId, 5*time.Second)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWatcherAlreadyArchived), "Expected ErrWatcherAlreadyArchived, got: %v", err)
	})

	t.Run("UpdateNonExistentWatcher", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PATCH" {
				w.WriteHeader(http.StatusNotFound)
				_, writeErr := w.Write([]byte(`{"error": "watcher not found"}`))
				require.NoError(t, writeErr)
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

	t.Run("ArchiveDuringWaitForArchived_CompletesSuccessfully", func(t *testing.T) {
		channelID := uuid.New()
		watcherID := uuid.New()

		getCallCount := 0
		patchDone := false
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.Method == "PATCH" && !patchDone {
				patchDone = true
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
			} else if r.Method == "GET" {
				getCallCount++
				name := "test-watcher"
				if getCallCount <= 2 {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusArchiving,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				} else {
					w.WriteHeader(http.StatusOK)
					response := apiClient.Watcher{
						WatcherId:     watcherID,
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
						Status:        apiClient.WatcherStatusArchived,
					}
					require.NoError(t, json.NewEncoder(w).Encode(response))
				}
			}
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()

		archived, err := client.Archive(ctx, channelID, watcherID)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WatcherStatusArchiving, archived.Status)

		err = client.WaitForArchived(ctx, channelID, watcherID, 5*time.Second)
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
			assert.Equal(t, "dvp", query.Get("service"))
			assert.Equal(t, string(apiClient.WatcherStatusActive), query.Get("status"))
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "5", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			name := "my-watcher"
			response := apiClient.WatcherList{
				Data: []apiClient.WatcherSummary{
					{
						WatcherId:     uuid.New(),
						Name:          &name,
						ChainSelector: "1337",
						Address:       "0x1234",
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

		ctx := context.Background()

		// Search with all filters
		chainSelector := "1337"
		name := "my-watcher"
		address := "0x1234"
		service := "dvp"
		status := apiClient.WatcherStatusActive
		limit := 10
		offset := int64(5)

		serviceFilter := []string{service}
		statusFilter := []apiClient.WatcherStatus{status}
		filters := ListFilters{
			Name:          &name,
			ChainSelector: &chainSelector,
			Address:       &address,
			Service:       &serviceFilter,
			Status:        &statusFilter,
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
			watchers := []apiClient.WatcherSummary{}
			for i := 0; i < 5 && offset+i < 15; i++ {
				name := "watcher-" + strconv.Itoa(offset+i)
				watchers = append(watchers, apiClient.WatcherSummary{
					WatcherId:     uuid.New(),
					Name:          &name,
					ChainSelector: "1337",
					Address:       "0x1234",
					Status:        apiClient.WatcherStatusActive,
					ChannelId:     channelID,
					CreatedAt:     time.Now().Unix(),
					DonFamily:     "zone-a",
				})
			}

			hasMore := offset+5 < 15
			response := apiClient.WatcherList{
				Data:    watchers,
				HasMore: hasMore,
			}
			require.NoError(t, json.NewEncoder(w).Encode(response))
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ctx := context.Background()
		limit := 5

		// Fetch pages
		allWatchers := []apiClient.WatcherSummary{}
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
