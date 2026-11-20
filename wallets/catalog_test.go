package wallets

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
	"github.com/smartcontractkit/crec-sdk/mocks/server"
)

// setupMockClient creates a wallets Client backed by the in-memory mock server.
func setupMockClient(t *testing.T) (*Client, *server.MockServer) {
	t.Helper()
	mock := server.NewMockServer()
	t.Cleanup(mock.Close)

	crecAPIClient, err := apiClient.NewClientWithResponses(mock.TestServer.URL)
	require.NoError(t, err)

	client, err := NewClient(&Options{
		Logger:    slog.New(slog.DiscardHandler),
		APIClient: crecAPIClient,
	})
	require.NoError(t, err)
	return client, mock
}

func TestClient_ListTypeVersions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client, _ := setupMockClient(t)

		versions, hasMore, err := client.ListTypeVersions(context.Background(), apiClient.WalletTypeECDSA, ListTypeVersionsInput{})
		require.NoError(t, err)
		assert.False(t, hasMore)
		require.Len(t, versions, 2)

		var foundDefault bool
		for _, v := range versions {
			assert.Equal(t, apiClient.WalletTypeECDSA, v.WalletType)
			assert.Equal(t, apiClient.WalletTypeVersionStatusActive, v.Status)
			for _, d := range v.Deployments {
				if d.IsDefault {
					foundDefault = true
				}
			}
		}
		assert.True(t, foundDefault, "one version should be the default")
	})

	t.Run("ChainSelectorFilter", func(t *testing.T) {
		client, _ := setupMockClient(t)
		unknownChain := "999999"

		versions, _, err := client.ListTypeVersions(context.Background(), apiClient.WalletTypeECDSA, ListTypeVersionsInput{
			ChainSelector: &unknownChain,
		})
		require.NoError(t, err)
		require.Len(t, versions, 2)
		for _, v := range versions {
			assert.Empty(t, v.Deployments, "versions with no deployment on the chain still return, with empty deployments")
		}
	})

	t.Run("EmptyWalletType", func(t *testing.T) {
		client, _ := setupMockClient(t)

		versions, hasMore, err := client.ListTypeVersions(context.Background(), "", ListTypeVersionsInput{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletTypeRequired))
		assert.Nil(t, versions)
		assert.False(t, hasMore)
	})
}

func TestClient_GetTypeVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client, _ := setupMockClient(t)

		version, err := client.GetTypeVersion(context.Background(), apiClient.WalletTypeECDSA, "2025.09.0")
		require.NoError(t, err)
		require.NotNil(t, version)
		assert.Equal(t, "2025.09.0", version.ImplementationVersion)
		assert.True(t, version.IsUpgradeable)
		assert.NotEmpty(t, version.Capabilities, "the newer seeded version publishes capabilities")
	})

	t.Run("NotFound", func(t *testing.T) {
		client, _ := setupMockClient(t)

		version, err := client.GetTypeVersion(context.Background(), apiClient.WalletTypeECDSA, "1999.01.0")
		require.Error(t, err)
		assert.Nil(t, version)
		assert.ErrorIs(t, err, apierror.ErrWalletTypeVersionNotFound)
	})

	t.Run("EmptyWalletType", func(t *testing.T) {
		client, _ := setupMockClient(t)

		_, err := client.GetTypeVersion(context.Background(), "", "2025.09.0")
		require.ErrorIs(t, err, ErrWalletTypeRequired)
	})

	t.Run("EmptyVersion", func(t *testing.T) {
		client, _ := setupMockClient(t)

		_, err := client.GetTypeVersion(context.Background(), apiClient.WalletTypeECDSA, "")
		require.ErrorIs(t, err, ErrVersionRequired)
	})
}
