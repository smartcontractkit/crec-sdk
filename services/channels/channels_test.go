package channels

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

func setupTestService(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	server := httptest.NewServer(handler)

	crecClient, err := client.NewCRECClient(&client.ClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})
	require.NoError(t, err)

	logger := zerolog.Nop()
	service, err := NewService(&ServiceOptions{
		Logger:     &logger,
		CRECClient: crecClient,
	})
	require.NoError(t, err)

	return service, server
}

func TestNewService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		crecClient, err := client.NewCRECClient(&client.ClientOptions{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-api-key",
		})
		require.NoError(t, err)

		logger := zerolog.Nop()
		service, err := NewService(&ServiceOptions{
			Logger:     &logger,
			CRECClient: crecClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
		assert.NotNil(t, service.crecClient)
	})

	t.Run("NilOptions", func(t *testing.T) {
		service, err := NewService(nil)

		require.Error(t, err)
		assert.Nil(t, service)
		assert.True(t, errors.Is(err, ErrServiceOptionsRequired), "Expected ErrServiceOptionsRequired, got: %v", err)
	})

	t.Run("NilCRECClient", func(t *testing.T) {
		logger := zerolog.Nop()
		service, err := NewService(&ServiceOptions{
			Logger:     &logger,
			CRECClient: nil,
		})

		require.Error(t, err)
		assert.Nil(t, service)
		assert.True(t, errors.Is(err, ErrCRECClientRequired), "Expected ErrCRECClientRequired, got: %v", err)
	})

	t.Run("DefaultLogger", func(t *testing.T) {
		crecClient, err := client.NewCRECClient(&client.ClientOptions{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-api-key",
		})
		require.NoError(t, err)

		service, err := NewService(&ServiceOptions{
			Logger:     nil,
			CRECClient: crecClient,
		})

		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
	})
}

func TestService_CreateChannel(t *testing.T) {
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.CreateChannel(context.Background(), CreateChannelInput{
			Name: channelName,
		})

		require.NoError(t, err)
		assert.NotNil(t, channel)
		assert.Equal(t, channelID, channel.ChannelId)
		assert.Equal(t, channelName, channel.Name)
		assert.Equal(t, createdAt, channel.CreatedAt)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.CreateChannel(context.Background(), CreateChannelInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		longName := make([]byte, MaxChannelNameLength+1)
		for i := range longName {
			longName[i] = 'a'
		}

		channel, err := service.CreateChannel(context.Background(), CreateChannelInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.CreateChannel(context.Background(), CreateChannelInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.CreateChannel(context.Background(), CreateChannelInput{
			Name: "existing-channel",
		})

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrCreateChannel), "Expected ErrCreateChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestService_GetChannel(t *testing.T) {
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.GetChannel(context.Background(), channelID)

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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.GetChannel(context.Background(), channelID)

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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channel, err := service.GetChannel(context.Background(), channelID)

		require.Error(t, err)
		assert.Nil(t, channel)
		assert.True(t, errors.Is(err, ErrGetChannel), "Expected ErrGetChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestService_ListChannels(t *testing.T) {
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		limit := 20
		offset := 0
		channels, hasMore, err := service.ListChannels(context.Background(), ListChannelsInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channels, hasMore, err := service.ListChannels(context.Background(), ListChannelsInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		limit := 10
		offset := 5
		channels, hasMore, err := service.ListChannels(context.Background(), ListChannelsInput{
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

		service, server := setupTestService(t, handler)
		defer server.Close()

		channels, hasMore, err := service.ListChannels(context.Background(), ListChannelsInput{})

		require.Error(t, err)
		assert.Nil(t, channels)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListChannels), "Expected ErrListChannels, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}

func TestService_DeleteChannel(t *testing.T) {
	t.Run("SuccessAsync", func(t *testing.T) {
		channelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/channels/"+channelID.String(), r.URL.Path)
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.WriteHeader(http.StatusAccepted) // 202 for async deletion
		}

		service, server := setupTestService(t, handler)
		defer server.Close()

		err := service.DeleteChannel(context.Background(), channelID)

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

		service, server := setupTestService(t, handler)
		defer server.Close()

		err := service.DeleteChannel(context.Background(), channelID)

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

		service, server := setupTestService(t, handler)
		defer server.Close()

		err := service.DeleteChannel(context.Background(), channelID)

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

		service, server := setupTestService(t, handler)
		defer server.Close()

		err := service.DeleteChannel(context.Background(), channelID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDeleteChannel), "Expected ErrDeleteChannel, got: %v", err)
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode), "Expected ErrUnexpectedStatusCode, got: %v", err)
	})
}
