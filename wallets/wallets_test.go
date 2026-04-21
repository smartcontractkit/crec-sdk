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
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

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
				Name:          walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ecdsaSigners := []string{}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                walletName,
			ChainSelector:       chainSelector,
			WalletOwnerAddress:  ownerAddress,
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, walletID, wallet.WalletId)
		assert.Equal(t, walletName, wallet.Name)
		assert.Equal(t, walletAddress, wallet.Address)
		assert.Equal(t, chainSelector, wallet.ChainSelector)
	})

	t.Run("DuplicateEcdsaSigners", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		ecdsaSigners := []string{
			"0x1234567890abcdef1234567890abcdef12345678",
			"0x1234567890abcdef1234567890abcdef12345678", // duplicate
		}
		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                "test",
			ChainSelector:       "ethereum-sepolia",
			WalletOwnerAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate ecdsa signer")
		assert.Nil(t, wallet)
	})

	t.Run("DuplicateRsaSigners", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		rsaSigners := []apiClient.RSAPublicKey{
			{E: "010001", N: "abc"},
			{E: "010001", N: "abc"}, // duplicate
		}
		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test",
			ChainSelector:      "ethereum-sepolia",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Rsa,
			AllowedRsaSigners:  &rsaSigners,
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate rsa signer")
		assert.Nil(t, wallet)
	})

	t.Run("SuccessWithStatusChannelId", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet-with-status-channel"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "ethereum-sepolia"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
		statusChannelID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var createReq apiClient.CreateWallet
			err = json.Unmarshal(body, &createReq)
			require.NoError(t, err)
			assert.Equal(t, walletName, createReq.Name)
			assert.Equal(t, chainSelector, createReq.ChainSelector)
			assert.Equal(t, ownerAddress, createReq.WalletOwnerAddress)
			assert.NotNil(t, createReq.StatusChannelId)
			assert.Equal(t, statusChannelID, *createReq.StatusChannelId)

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:        walletID,
				Name:            walletName,
				Address:         walletAddress,
				ChainSelector:   chainSelector,
				StatusChannelId: &statusChannelID,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ecdsaSigners := []string{}

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                walletName,
			ChainSelector:       chainSelector,
			WalletOwnerAddress:  ownerAddress,
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, walletID, wallet.WalletId)
		assert.Equal(t, walletName, wallet.Name)
		assert.Equal(t, walletAddress, wallet.Address)
		assert.Equal(t, chainSelector, wallet.ChainSelector)
		assert.NotNil(t, wallet.StatusChannelId)
		assert.Equal(t, statusChannelID, *wallet.StatusChannelId)
	})

	t.Run("EmptyName", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty name")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Ecdsa,
			StatusChannelId:    &statusChannelID,
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

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Ecdsa,
			StatusChannelId:    &statusChannelID,
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

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "",
			WalletType:         apiClient.Ecdsa,
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletOwnerAddressRequired))
	})

	t.Run("InvalidOwnerAddress", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with invalid owner address")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "not-a-valid-hex-address",
			WalletType:         apiClient.Ecdsa,
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidWalletOwnerAddress))
	})

	t.Run("EmptyWalletType", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with empty wallet type")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "",
			StatusChannelId:    &statusChannelID,
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

		statusChannelID := uuid.New()
		longName := string(make([]byte, MaxWalletNameLength+1))
		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               longName,
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Ecdsa,
			StatusChannelId:    &statusChannelID,
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

		ecdsaSigners := []string{}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                "test-wallet",
			ChainSelector:       "5009297550715157269",
			WalletOwnerAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrCreateWallet))
	})

	t.Run("UnsupportedWalletType", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         "unsupported",
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrUnsupportedWalletType))
	})

	t.Run("EcdsaWalletWithRsaSigners", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request with mismatched signers")
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		rsaSigners := apiClient.RSASignersList{{E: "AQAB", N: "abc123"}}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Ecdsa,
			AllowedRsaSigners:  &rsaSigners,
			StatusChannelId:    &statusChannelID,
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

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                "test-wallet",
			ChainSelector:       "5009297550715157269",
			WalletOwnerAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:          apiClient.Rsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidSignersForRsa))
	})

	t.Run("EcdsaWalletWithInvalidSigner", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		ecdsaSigners := []string{"0x1234567890abcdef1234567890abcdef12345678", "not-a-valid-address"}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                "test-wallet",
			ChainSelector:       "5009297550715157269",
			WalletOwnerAddress:  "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidEcdsaSigner))
	})

	t.Run("EcdsaWalletWithEcdsaSigners", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "5009297550715157269"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		ecdsaSigners := []string{"0x1234567890abcdef1234567890abcdef12345678", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:                walletName,
			ChainSelector:       chainSelector,
			WalletOwnerAddress:  ownerAddress,
			WalletType:          apiClient.Ecdsa,
			AllowedEcdsaSigners: &ecdsaSigners,
			StatusChannelId:     &statusChannelID,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
	})

	t.Run("RsaWalletWithRsaSigners", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "5009297550715157269"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          walletName,
				Address:       walletAddress,
				ChainSelector: chainSelector,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		rsaSigners := apiClient.RSASignersList{{E: "AQAB", N: "abc123"}}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               walletName,
			ChainSelector:      chainSelector,
			WalletOwnerAddress: ownerAddress,
			WalletType:         apiClient.Rsa,
			AllowedRsaSigners:  &rsaSigners,
			StatusChannelId:    &statusChannelID,
		})

		require.NoError(t, err)
		assert.NotNil(t, wallet)
	})

	t.Run("RsaWalletWithInvalidSigner_EmptyE", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		rsaSigners := apiClient.RSASignersList{{E: "", N: "abc123"}}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Rsa,
			AllowedRsaSigners:  &rsaSigners,
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidRsaSigner))
	})

	t.Run("RsaWalletWithInvalidSigner_EmptyN", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		rsaSigners := apiClient.RSASignersList{{E: "AQAB", N: ""}}

		statusChannelID := uuid.New()

		wallet, err := client.Create(context.Background(), CreateInput{
			Name:               "test-wallet",
			ChainSelector:      "5009297550715157269",
			WalletOwnerAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			WalletType:         apiClient.Rsa,
			AllowedRsaSigners:  &rsaSigners,
			StatusChannelId:    &statusChannelID,
		})

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidRsaSigner))
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("NilWalletID", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		wallet, err := client.Get(context.Background(), uuid.Nil)

		require.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletIDRequired))
	})

	t.Run("Success", func(t *testing.T) {
		walletID := uuid.New()
		walletName := "test-wallet"
		walletAddress := "0x1234567890abcdef1234567890abcdef12345678"
		chainSelector := "5009297550715157269"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets/"+walletID.String(), r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Wallet{
				WalletId:      walletID,
				Name:          walletName,
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
		assert.Equal(t, walletName, wallet.Name)
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
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.WalletList{
				Data: []apiClient.Wallet{
					{
						WalletId:      walletID1,
						Name:          walletName1,
						Address:       "0x1111111111111111111111111111111111111111",
						ChainSelector: "ethereum-sepolia",
					},
					{
						WalletId:      walletID2,
						Name:          walletName2,
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
		chainSelector := "5009297550715157269"
		ownerAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
		walletType := apiClient.Ecdsa
		walletStatus := []apiClient.WalletStatus{apiClient.WalletStatusDeployed}
		limit := 10
		offset := int64(5)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets", r.URL.Path)
			assert.Equal(t, walletName, r.URL.Query().Get("name"))
			assert.Equal(t, chainSelector, r.URL.Query().Get("chain_selector"))
			assert.Equal(t, ownerAddress, r.URL.Query().Get("owner"))
			assert.Equal(t, "ecdsa", r.URL.Query().Get("type"))
			assert.Equal(t, string(apiClient.WalletStatusDeployed), r.URL.Query().Get("status"))
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
			Owner:         &ownerAddress,
			Type:          &walletType,
			Status:        &walletStatus,
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

	t.Run("InvalidLimit_Zero", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		limit := 0
		wallets, hasMore, err := client.List(context.Background(), ListInput{
			Limit: &limit,
		})

		require.Error(t, err)
		assert.Nil(t, wallets)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrInvalidLimit))
	})

	t.Run("InvalidLimit_Negative", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		limit := -1
		wallets, hasMore, err := client.List(context.Background(), ListInput{
			Limit: &limit,
		})

		require.Error(t, err)
		assert.Nil(t, wallets)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrInvalidLimit))
	})

	t.Run("InvalidOffset_Negative", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		offset := int64(-1)
		wallets, hasMore, err := client.List(context.Background(), ListInput{
			Offset: &offset,
		})

		require.Error(t, err)
		assert.Nil(t, wallets)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrInvalidOffset))
	})

	t.Run("InvalidOwner_InvalidAddress", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		invalidOwner := "not-a-valid-address"
		wallets, hasMore, err := client.List(context.Background(), ListInput{
			Owner: &invalidOwner,
		})

		require.Error(t, err)
		assert.Nil(t, wallets)
		assert.False(t, hasMore)
		assert.True(t, errors.Is(err, ErrInvalidOwnerAddress))
	})
}

func TestClient_Update(t *testing.T) {
	t.Run("NilWalletID", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		err := client.Update(context.Background(), uuid.Nil, UpdateInput{
			Name: "new-name",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletIDRequired))
	})

	t.Run("Success", func(t *testing.T) {
		walletID := uuid.New()
		newName := "updated-wallet"

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets/"+walletID.String(), r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			// Read and validate request body
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var updateReq apiClient.UpdateWallet
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			require.NotNil(t, updateReq.Name)
			assert.Equal(t, newName, *updateReq.Name)

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

func TestClient_Archive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/wallets/"+walletID.String(), r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "Apikey test-api-key", r.Header.Get("Authorization"))

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var updateReq apiClient.UpdateWallet
			err = json.Unmarshal(body, &updateReq)
			require.NoError(t, err)
			assert.NotNil(t, updateReq.Status)
			assert.Equal(t, apiClient.WalletStatusArchived, *updateReq.Status)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := apiClient.Wallet{
				WalletId: walletID,
				Name:     "test-wallet",
				Status:   apiClient.WalletStatusArchived,
			}
			json.NewEncoder(w).Encode(response)
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Archive(context.Background(), walletID)

		require.NoError(t, err)
	})

	t.Run("NilWalletID", func(t *testing.T) {
		client, server := setupTestClient(t, nil)
		defer server.Close()

		err := client.Archive(context.Background(), uuid.Nil)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletIDRequired))
	})

	t.Run("NotFound", func(t *testing.T) {
		walletID := uuid.New()

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Wallet not found",
				"type":    "Not found",
			})
		}

		client, server := setupTestClient(t, handler)
		defer server.Close()

		err := client.Archive(context.Background(), walletID)

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

		err := client.Archive(context.Background(), walletID)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrArchiveWallet))
		assert.True(t, errors.Is(err, ErrUnexpectedStatusCode))
	})
}
