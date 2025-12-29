package wallets

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
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

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
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
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
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
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWallet
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)
			assert.Equal(t, walletName, createReq.Name)
			assert.Equal(t, chainSelector, createReq.ChainSelector)
			assert.Equal(t, ownerAddress, createReq.WalletOwnerAddress)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          &walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               walletName,
			ChainSelector:      chainSelector,
			WalletOwnerAddress: ownerAddress,
			WalletType:         "ecdsa",
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, walletID, wallet.WalletId)
		assert.Equal(t, walletName, *wallet.Name)
		assert.Equal(t, walletAddress, wallet.Address)
		assert.Equal(t, chainSelector, wallet.ChainSelector)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "ecdsa",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrNameRequired))
	})

	t.Run("EmptyChainSelector", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty chain selector")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "ecdsa",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrChainSelectorRequired))
	})

	t.Run("EmptyOwnerAddress", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty owner address")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "",
			WalletType:         "ecdsa",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletOwnerAddressRequired))
	})

	t.Run("EmptyWalletType", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty wallet type")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletTypeRequired))
	})

	t.Run("NameTooLong", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with name too long")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		longName := string(make([]byte, MaxWalletNameLength+1))
		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               longName,
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "ecdsa",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrNameTooLong))
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "ecdsa",
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrCreateWallet))
	})

	t.Run("EcdsaWalletWithRsaSigners", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with mismatched signers")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		rsaSigners := []struct {
			E string `json:"e"`
			N string `json:"n"`
		}{
			{E: "AQAB", N: "abc123"},
		}

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "ecdsa",
			AllowedRsaSigners:  &rsaSigners,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidSignersForEcdsa))
	})

	t.Run("RsaWalletWithEcdsaSigners", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with mismatched signers")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ecdsaSigners := []string{"0x123...", "0x456..."}

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                "test-wallet",
			ChainSelector:       "ethereum-sepolia",
			WalletOwnerAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:          "rsa",
			AllowedEcdsaSigners: &ecdsaSigners,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidSignersForRsa))
	})

	t.Run("EcdsaWalletWithEcdsaSigners", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          &walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ecdsaSigners := []string{"0x123...", "0x456..."}

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                walletName,
			ChainSelector:       chainSelector,
			WalletOwnerAddress:  ownerAddress,
			WalletType:          "ecdsa",
			AllowedEcdsaSigners: &ecdsaSigners,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
	})

	t.Run("RsaWalletWithRsaSigners", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          &walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		rsaSigners := []struct {
			E string `json:"e"`
			N string `json:"n"`
		}{
			{E: "AQAB", N: "abc123"},
		}

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               walletName,
			ChainSelector:      chainSelector,
			WalletOwnerAddress: ownerAddress,
			WalletType:         "rsa",
			AllowedRsaSigners:  &rsaSigners,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets/"+walletID.String(), r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          &walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Get(context.Background(), walletID)

		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, walletID, wallet.WalletId)
		assert.Equal(t, walletName, *wallet.Name)
		assert.Equal(t, walletAddress, wallet.Address)
	})

	t.Run("NotFound", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Get(context.Background(), walletID)

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletNotFound))
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallet, err := client.Get(context.Background(), walletID)

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrGetWallet))
	})
}

func TestClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		walletID1 := uuid.New()
		walletID2 := uuid.New()
		walletName1 := "wallet-1"
		walletName2 := "wallet-2"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.WalletList{
				Data: []apiClient.Wallet{
					{
						WalletId:      walletID1,
						Name:          &walletName1,
						Address:       "0x1111111111111111111111111111111111111111",
						ChainSelector: "ethereum-sepolia",
					},
					{
						WalletId:      walletID2,
						Name:          &walletName2,
						Address:       "0x2222222222222222222222222222222222222222",
						ChainSelector: "ethereum-mainnet",
					},
				},
				HasMore: false,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallets, hasMore, err := client.List(context.Background(), ListInput{})

		require.NoError(t, err)
		assert.NotNil(t, wallets)
		assert.Len(t, wallets, 2)
		assert.False(t, hasMore)
		assert.Equal(t, walletID1, wallets[0].WalletId)
		assert.Equal(t, walletID2, wallets[1].WalletId)
	})

	t.Run("WithFilters", func(t *testing.T) {
		walletName := "production"
		chainSelector := "ethereum-mainnet"
		limit := 10
		offset := int64(5)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, walletName, r.URL.Query().Get("name"))
			assert.Equal(t, chainSelector, r.URL.Query().Get("chain_selector"))
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "5", r.URL.Query().Get("offset"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.WalletList{
				Data:    []apiClient.Wallet{},
				HasMore: true,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallets, hasMore, err := client.List(context.Background(), ListInput{
			Name:          &walletName,
			ChainSelector: &chainSelector,
			Limit:         &limit,
			Offset:        &offset,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallets)
		assert.True(t, hasMore)
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		wallets, hasMore, err := client.List(context.Background(), ListInput{})

		require.Error(t, err)
		assert.Nil(t, wallets)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrListWallets))
	})
}

func TestClient_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		walletID := uuid.New()
		newName := "updated-wallet"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets/"+walletID.String(), r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var updateReq apiClient.UpdateWallet
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			assert.Equal(t, newName, updateReq.Name)

			w.WriteHeader(http.StatusOK)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Update(context.Background(), walletID, UpdateInput{
			Name: newName,
		})

		require.NoError(t, err)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Update(context.Background(), uuid.New(), UpdateInput{
			Name: "",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNameRequired))
	})

	t.Run("NameTooLong", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with name too long")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		longName := string(make([]byte, MaxWalletNameLength+1))
		err := client.Update(context.Background(), uuid.New(), UpdateInput{
			Name: longName,
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNameTooLong))
	})

	t.Run("NotFound", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Update(context.Background(), walletID, UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletNotFound))
	})

	t.Run("UnexpectedStatusCode", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Update(context.Background(), walletID, UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrUpdateWallet))
	})
}
