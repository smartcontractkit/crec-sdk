package channels

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		Logger:    logger,
		APIClient: crecAPIClient,
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
}

func TestClient_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		channelName := "test-channel"
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateChannel
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)
			assert.Equal(t, channelName, createReq.Name)
			assert.Nil(t, createReq.Description)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Channel{
				ChannelId: channelID,
				Name:      channelName,
				CreatedAt: createdAt,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Create(context.Background(), CreateInput{
			Name: channelName,
		})

		require.NoError(t, err)
		assert.NotNil(t, channel)
		assert.Equal(t, channelID, channel.ChannelId)
		assert.Equal(t, channelName, channel.Name)
		assert.Equal(t, createdAt, channel.CreatedAt)
	})

	t.Run("Success_WithDescription", func(t *testing.T) {
		channelID := uuid.New()
		channelName := "test-channel"
		channelDescription := "Test channel description"
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateChannel
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)
			assert.Equal(t, channelName, createReq.Name)
			assert.NotNil(t, createReq.Description)
			assert.Equal(t, channelDescription, *createReq.Description)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Channel{
				ChannelId:   channelID,
				Name:        channelName,
				Description: &channelDescription,
				CreatedAt:   createdAt,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Create(context.Background(), CreateInput{
			Name:        channelName,
			Description: &channelDescription,
		})

		require.NoError(t, err)
		assert.NotNil(t, channel)
		assert.Equal(t, channelID, channel.ChannelId)
		assert.Equal(t, channelName, channel.Name)
		assert.NotNil(t, channel.Description)
		assert.Equal(t, channelDescription, *channel.Description)
		assert.Equal(t, createdAt, channel.CreatedAt)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Create(context.Background(), CreateInput{
			Name: "",
		})

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrChannelNameRequired), "Expected ErrChannelNameRequired, got: %v", err)
	})

	t.Run("NameTooLong", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with name too long")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		longName := make([]byte, MaxChannelNameLength+1)
		for i := range longName {
			longName[i] = 'a'
		}

		channel, err := client.Create(context.Background(), CreateInput{
			Name: string(longName),
		})

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrChannelNameTooLong), "Expected ErrChannelNameTooLong, got: %v", err)
	})

	t.Run("BadRequest", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid channel name",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Create(context.Background(), CreateInput{
			Name: "test-channel",
		})

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrCreateChannel), "Expected ErrCreateChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Channel already exists",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Create(context.Background(), CreateInput{
			Name: "existing-channel",
		})

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrCreateChannel), "Expected ErrCreateChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channelID := uuid.New()
		channelName := "test-channel"
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String(), r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Channel{
				ChannelId: channelID,
				Name:      channelName,
				CreatedAt: createdAt,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Get(context.Background(), channelID)

		require.NoError(t, err)
		assert.NotNil(t, channel)
		assert.Equal(t, channelID, channel.ChannelId)
		assert.Equal(t, channelName, channel.Name)
		assert.Equal(t, createdAt, channel.CreatedAt)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Channel not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channel, err := client.Get(context.Background(), channelID)

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrChannelNotFound), "Expected ErrChannelNotFound, got: %v", err)
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

		channel, err := client.Get(context.Background(), channelID)

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrGetChannel), "Expected ErrGetChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channel1ID := uuid.New()
		channel2ID := uuid.New()
		createdAt := time.Now().Unix()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, "20", query.Get("limit"))
			assert.Equal(t, "0", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.ChannelList{
				Data: []apiClient.Channel{
					{
						ChannelId: channel1ID,
						Name:      "channel-1",
						CreatedAt: createdAt,
					},
					{
						ChannelId: channel2ID,
						Name:      "channel-2",
						CreatedAt: createdAt,
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		limit := 20
		offset := int64(0)
		channels, hasMore, err := client.List(context.Background(), ListInput{
			Limit:  &limit,
			Offset: &offset,
		})

		require.NoError(t, err)
		assert.Len(t, channels, 2)
		assert.False(t, hasMore)
		assert.Equal(t, channel1ID, channels[0].ChannelId)
		assert.Equal(t, "channel-1", channels[0].Name)
		assert.Equal(t, channel2ID, channels[1].ChannelId)
		assert.Equal(t, "channel-2", channels[1].Name)
	})

	t.Run("WithNameFilter", func(t *testing.T) {
		channelID := uuid.New()
		createdAt := time.Now().Unix()
		filterName := "test"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, filterName, query.Get("name"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.ChannelList{
				Data: []apiClient.Channel{
					{
						ChannelId: channelID,
						Name:      "test-channel",
						CreatedAt: createdAt,
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channels, hasMore, err := client.List(context.Background(), ListInput{
			Name: &filterName,
		})

		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.False(t, hasMore)
		assert.Equal(t, "test-channel", channels[0].Name)
	})

	t.Run("WithPagination", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, "10", query.Get("limit"))
			assert.Equal(t, "5", query.Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.ChannelList{
				Data:    []apiClient.Channel{},
				HasMore: true,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		limit := 10
		offset := int64(5)
		channels, hasMore, err := client.List(context.Background(), ListInput{
			Limit:  &limit,
			Offset: &offset,
		})

		require.NoError(t, err)
		assert.Len(t, channels, 0)
		assert.True(t, hasMore)
	})

	t.Run("ServerError", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		channels, hasMore, err := client.List(context.Background(), ListInput{})

		require.Error(t, err)
		assert.Nil(t, channels)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListChannels), "Expected ErrListChannels, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("SuccessAsync", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String(), r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.WriteHeader(http.StatusAccepted) // 202 for async deletion
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID)

		require.NoError(t, err)
	})

	t.Run("SuccessSync", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String(), r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.WriteHeader(http.StatusNoContent) // 204 for sync deletion
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID)

		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Channel not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Delete(context.Background(), channelID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrChannelNotFound), "Expected ErrChannelNotFound, got: %v", err)
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

		err := client.Delete(context.Background(), channelID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDeleteChannel), "Expected ErrDeleteChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}
