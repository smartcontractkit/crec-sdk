package privy

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func MockPrivyServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/wallets/mock-wallet-id-123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != "test-app-id" || password != "test-app-secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		appID := r.Header.Get("privy-app-id")
		if appID != "test-app-id" {
			http.Error(w, "Missing or invalid privy-app-id header", http.StatusBadRequest)
			return
		}

		wallet := WalletResponse{
			ID:           "mock-wallet-id-123",
			Address:      "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4",
			ChainType:    "ethereum",
			WalletIndex:  0,
			PublicKey:    "0x04c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5ae0c2c2e9f1c6f3b8a6b6e7c8c9a8f7e9c8b6d3e5b7c9d8a6f7e9c8b6d5c7e8f9",
			WalletClient: "privy",
			Imported:     false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wallet)
	})

	mux.HandleFunc("/v1/wallets/mock-wallet-id-123/rpc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != "test-app-id" || password != "test-app-secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		appID := r.Header.Get("privy-app-id")
		if appID != "test-app-id" {
			http.Error(w, "Missing or invalid privy-app-id header", http.StatusBadRequest)
			return
		}

		var rpcReq RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&rpcReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if rpcReq.Method != "personal_sign" {
			http.Error(w, "Unsupported method", http.StatusBadRequest)
			return
		}

		// Generate a valid 65-byte Ethereum signature for a fixed message using a test private key
		testPrivKeyHex := "4c0883a69102937d6231471b5dbb6204fe5129617082796fe2b8b4b6d20b6e8a" // well-known test key
		testPrivKey, err := crypto.HexToECDSA(testPrivKeyHex)
		require.NoError(t, err)
		message := []byte("test message for signature")
		hash := crypto.Keccak256(
			[]byte("\x19Ethereum Signed Message:\n" + string(len(message))),
			message,
		)
		sig, err := crypto.Sign(hash, testPrivKey)
		require.NoError(t, err)
		mockSignature := "0x" + hex.EncodeToString(sig)

		rpcResp := RPCResponse{
			Method: "personal_sign",
			Data: struct {
				Signature string `json:"signature"`
				Encoding  string `json:"encoding"`
			}{
				Signature: mockSignature,
				Encoding:  "hex",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rpcResp)
	})

	return httptest.NewServer(mux)
}

func TestNewSigner(t *testing.T) {
	tests := []struct {
		name        string
		appID       string
		appSecret   string
		walletID    string
		baseURL     string
		expectError bool
	}{
		{
			name:        "valid parameters",
			appID:       "test-app-id",
			appSecret:   "test-app-secret",
			walletID:    "test-wallet-id",
			baseURL:     "https://api.privy.io",
			expectError: false,
		},
		{
			name:        "empty app ID",
			appID:       "",
			appSecret:   "test-app-secret",
			walletID:    "test-wallet-id",
			baseURL:     "https://api.privy.io",
			expectError: true,
		},
		{
			name:        "empty app secret",
			appID:       "test-app-id",
			appSecret:   "",
			walletID:    "test-wallet-id",
			baseURL:     "https://api.privy.io",
			expectError: true,
		},
		{
			name:        "empty wallet ID",
			appID:       "test-app-id",
			appSecret:   "test-app-secret",
			walletID:    "",
			baseURL:     "https://api.privy.io",
			expectError: true,
		},
		{
			name:        "empty base URL",
			appID:       "test-app-id",
			appSecret:   "test-app-secret",
			walletID:    "test-wallet-id",
			baseURL:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewSigner(tt.appID, tt.appSecret, tt.walletID, tt.baseURL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, signer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, signer)
				assert.Equal(t, tt.appID, signer.appID)
				assert.Equal(t, tt.appSecret, signer.appSecret)
				assert.Equal(t, tt.walletID, signer.walletID)
				assert.Equal(t, tt.baseURL, signer.baseURL)
			}
		})
	}
}

func TestNewSignerFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"PRIVY_APP_ID":     "test-app-id",
				"PRIVY_APP_SECRET": "test-app-secret",
				"PRIVY_WALLET_ID":  "test-wallet-id",
				"PRIVY_BASE_URL":   "https://api.privy.io",
			},
			expectError: false,
		},
		{
			name: "default base URL when not set",
			envVars: map[string]string{
				"PRIVY_APP_ID":     "test-app-id",
				"PRIVY_APP_SECRET": "test-app-secret",
				"PRIVY_WALLET_ID":  "test-wallet-id",
			},
			expectError: false,
		},
		{
			name: "missing app ID",
			envVars: map[string]string{
				"PRIVY_APP_SECRET": "test-app-secret",
				"PRIVY_WALLET_ID":  "test-wallet-id",
			},
			expectError: true,
			errorMsg:    "PRIVY_APP_ID environment variable not set",
		},
		{
			name: "missing app secret",
			envVars: map[string]string{
				"PRIVY_APP_ID":    "test-app-id",
				"PRIVY_WALLET_ID": "test-wallet-id",
			},
			expectError: true,
			errorMsg:    "PRIVY_APP_SECRET environment variable not set",
		},
		{
			name: "missing wallet ID",
			envVars: map[string]string{
				"PRIVY_APP_ID":     "test-app-id",
				"PRIVY_APP_SECRET": "test-app-secret",
			},
			expectError: true,
			errorMsg:    "PRIVY_WALLET_ID environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing environment variables
			os.Unsetenv("PRIVY_APP_ID")
			os.Unsetenv("PRIVY_APP_SECRET")
			os.Unsetenv("PRIVY_WALLET_ID")
			os.Unsetenv("PRIVY_BASE_URL")

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			signer, err := NewSignerFromEnv()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, signer)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, signer)
				assert.Equal(t, tt.envVars["PRIVY_APP_ID"], signer.appID)
				assert.Equal(t, tt.envVars["PRIVY_APP_SECRET"], signer.appSecret)
				assert.Equal(t, tt.envVars["PRIVY_WALLET_ID"], signer.walletID)

				expectedBaseURL := tt.envVars["PRIVY_BASE_URL"]
				if expectedBaseURL == "" {
					expectedBaseURL = "https://api.privy.io"
				}
				assert.Equal(t, expectedBaseURL, signer.baseURL)
			}

			// Clean up environment variables
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestSigner_Sign(t *testing.T) {
	server := MockPrivyServer(t)
	defer server.Close()

	signer, err := NewSignerWithCustomClient("test-app-id", "test-app-secret", "mock-wallet-id-123", &http.Client{}, server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	hash := crypto.Keccak256([]byte("hello world"))

	signature, err := signer.Sign(ctx, hash)

	require.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Greater(t, len(signature), 0)

	// Verify the signature is properly decoded from hex - should be 65 bytes
	expectedSig, _ := hex.DecodeString(expectedSignatureHex)
	assert.Equal(t, expectedSig, signature)
	assert.Equal(t, 65, len(signature))
}

func TestSigner_GetWalletAddress(t *testing.T) {
	server := MockPrivyServer(t)
	defer server.Close()

	signer, err := NewSignerWithCustomClient("test-app-id", "test-app-secret", "mock-wallet-id-123", &http.Client{}, server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	address, err := signer.GetWalletAddress(ctx)

	require.NoError(t, err)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4", address)
}

func TestSigner_AuthenticationFailure(t *testing.T) {
	server := MockPrivyServer(t)
	defer server.Close()

	// Create signer with wrong credentials
	signer, err := NewSignerWithCustomClient("wrong-app-id", "wrong-app-secret", "mock-wallet-id-123", &http.Client{}, server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	hash := crypto.Keccak256([]byte("hello world"))
	_, err = signer.Sign(ctx, hash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")

	_, err = signer.GetWalletAddress(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestSigner_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	signer, err := NewSignerWithCustomClient("test-app-id", "test-app-secret", "test-wallet-id", &http.Client{}, server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	hash := crypto.Keccak256([]byte("hello world"))

	_, err = signer.Sign(ctx, hash)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
