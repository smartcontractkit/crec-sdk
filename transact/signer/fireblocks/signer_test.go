package fireblocks_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/crec-sdk/transact/signer/fireblocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "test-api-key"
const testVaultAccountID = "0"
const testAssetID = "ETH_TEST5"

func generateTestKey(t *testing.T) (string, *rsa.PrivateKey) {
	t.Helper()
	key, pem, err := fireblocks.GenerateTestPrivateKey()
	require.NoError(t, err)
	return pem, key
}

func TestSigner_NewSigner(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	tests := []struct {
		name           string
		apiKey         string
		privateKeyPEM  string
		vaultAccountID string
		assetID        string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid parameters",
			apiKey:         testAPIKey,
			privateKeyPEM:  pemKey,
			vaultAccountID: testVaultAccountID,
			assetID:        testAssetID,
			wantErr:        false,
		},
		{
			name:           "empty api key",
			apiKey:         "",
			privateKeyPEM:  pemKey,
			vaultAccountID: testVaultAccountID,
			assetID:        testAssetID,
			wantErr:        true,
			errContains:    "apiKey cannot be empty",
		},
		{
			name:           "empty private key",
			apiKey:         testAPIKey,
			privateKeyPEM:  "",
			vaultAccountID: testVaultAccountID,
			assetID:        testAssetID,
			wantErr:        true,
			errContains:    "privateKeyPEM cannot be empty",
		},
		{
			name:           "empty vault account ID",
			apiKey:         testAPIKey,
			privateKeyPEM:  pemKey,
			vaultAccountID: "",
			assetID:        testAssetID,
			wantErr:        true,
			errContains:    "vaultAccountID cannot be empty",
		},
		{
			name:           "empty asset ID",
			apiKey:         testAPIKey,
			privateKeyPEM:  pemKey,
			vaultAccountID: testVaultAccountID,
			assetID:        "",
			wantErr:        true,
			errContains:    "assetID cannot be empty",
		},
		{
			name:           "invalid private key PEM",
			apiKey:         testAPIKey,
			privateKeyPEM:  "not a valid PEM",
			vaultAccountID: testVaultAccountID,
			assetID:        testAssetID,
			wantErr:        true,
			errContains:    "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := fireblocks.NewSigner(tt.apiKey, tt.privateKeyPEM, tt.vaultAccountID, tt.assetID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, signer)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, signer)
			}
		})
	}
}

func TestSigner_NewSignerWithOptions(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	tests := []struct {
		name    string
		opts    []fireblocks.Option
		wantErr bool
	}{
		{
			name:    "with custom timeout",
			opts:    []fireblocks.Option{fireblocks.WithTimeout(30 * time.Second)},
			wantErr: false,
		},
		{
			name:    "with custom polling interval",
			opts:    []fireblocks.Option{fireblocks.WithPollingInterval(time.Second)},
			wantErr: false,
		},
		{
			name:    "with custom base URL",
			opts:    []fireblocks.Option{fireblocks.WithBaseURL("https://sandbox-api.fireblocks.io")},
			wantErr: false,
		},
		{
			name: "with multiple options",
			opts: []fireblocks.Option{
				fireblocks.WithTimeout(30 * time.Second),
				fireblocks.WithPollingInterval(time.Second),
				fireblocks.WithBaseURL("https://sandbox-api.fireblocks.io"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := fireblocks.NewSigner(testAPIKey, pemKey, testVaultAccountID, testAssetID, tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, signer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, signer)
			}
		})
	}
}

func TestSigner_NewSignerFromEnv(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid environment variables with PEM content",
			envVars: map[string]string{
				"FIREBLOCKS_API_KEY":          testAPIKey,
				"FIREBLOCKS_API_SECRET":       pemKey,
				"FIREBLOCKS_VAULT_ACCOUNT_ID": testVaultAccountID,
				"FIREBLOCKS_ASSET_ID":         testAssetID,
			},
			wantErr: false,
		},
		{
			name: "valid environment variables with custom base URL",
			envVars: map[string]string{
				"FIREBLOCKS_API_KEY":          testAPIKey,
				"FIREBLOCKS_API_SECRET":       pemKey,
				"FIREBLOCKS_VAULT_ACCOUNT_ID": testVaultAccountID,
				"FIREBLOCKS_ASSET_ID":         testAssetID,
				"FIREBLOCKS_BASE_URL":         "https://sandbox-api.fireblocks.io",
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			envVars: map[string]string{
				"FIREBLOCKS_API_SECRET":       pemKey,
				"FIREBLOCKS_VAULT_ACCOUNT_ID": testVaultAccountID,
				"FIREBLOCKS_ASSET_ID":         testAssetID,
			},
			wantErr:     true,
			errContains: "FIREBLOCKS_API_KEY",
		},
		{
			name: "missing api secret",
			envVars: map[string]string{
				"FIREBLOCKS_API_KEY":          testAPIKey,
				"FIREBLOCKS_VAULT_ACCOUNT_ID": testVaultAccountID,
				"FIREBLOCKS_ASSET_ID":         testAssetID,
			},
			wantErr:     true,
			errContains: "FIREBLOCKS_API_SECRET",
		},
		{
			name: "missing vault account ID",
			envVars: map[string]string{
				"FIREBLOCKS_API_KEY":    testAPIKey,
				"FIREBLOCKS_API_SECRET": pemKey,
				"FIREBLOCKS_ASSET_ID":   testAssetID,
			},
			wantErr:     true,
			errContains: "FIREBLOCKS_VAULT_ACCOUNT_ID",
		},
		{
			name: "missing asset ID",
			envVars: map[string]string{
				"FIREBLOCKS_API_KEY":          testAPIKey,
				"FIREBLOCKS_API_SECRET":       pemKey,
				"FIREBLOCKS_VAULT_ACCOUNT_ID": testVaultAccountID,
			},
			wantErr:     true,
			errContains: "FIREBLOCKS_ASSET_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearFireblocksEnvVars()
			defer clearFireblocksEnvVars()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			signer, err := fireblocks.NewSignerFromEnv()

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, signer)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, signer)
			}
		})
	}
}

func TestSigner_NewSignerFromEnv_FileRead(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	// Create a temporary file with the PEM key
	tmpFile, err := os.CreateTemp("", "fireblocks-test-key-*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(pemKey)
	require.NoError(t, err)
	tmpFile.Close()

	clearFireblocksEnvVars()
	defer clearFireblocksEnvVars()

	os.Setenv("FIREBLOCKS_API_KEY", testAPIKey)
	os.Setenv("FIREBLOCKS_API_SECRET", tmpFile.Name())
	os.Setenv("FIREBLOCKS_VAULT_ACCOUNT_ID", testVaultAccountID)
	os.Setenv("FIREBLOCKS_ASSET_ID", testAssetID)

	signer, err := fireblocks.NewSignerFromEnv()
	require.NoError(t, err)
	require.NotNil(t, signer)
}

func TestSigner_Sign(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	// Generate a test signing key to produce valid signatures
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubKeyBytes := crypto.FromECDSAPub(&signingKey.PublicKey)

	tests := []struct {
		name        string
		setupServer func(t *testing.T) *httptest.Server
		wantErr     bool
		errContains string
	}{
		{
			name: "successful signing",
			setupServer: func(t *testing.T) *httptest.Server {
				return createMockFireblocksServer(t, signingKey, pubKeyBytes, fireblocks.StatusCompleted)
			},
			wantErr: false,
		},
		{
			name: "transaction rejected",
			setupServer: func(t *testing.T) *httptest.Server {
				return createMockFireblocksServer(t, signingKey, pubKeyBytes, fireblocks.StatusRejected)
			},
			wantErr:     true,
			errContains: "REJECTED",
		},
		{
			name: "transaction failed",
			setupServer: func(t *testing.T) *httptest.Server {
				return createMockFireblocksServer(t, signingKey, pubKeyBytes, fireblocks.StatusFailed)
			},
			wantErr:     true,
			errContains: "FAILED",
		},
		{
			name: "create transaction error",
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal error"}`))
				}))
			},
			wantErr:     true,
			errContains: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer(t)
			defer server.Close()

			signer, err := fireblocks.NewSigner(
				testAPIKey, pemKey, testVaultAccountID, testAssetID,
				fireblocks.WithBaseURL(server.URL),
				fireblocks.WithTimeout(5*time.Second),
				fireblocks.WithPollingInterval(10*time.Millisecond),
			)
			require.NoError(t, err)

			ctx := context.Background()
			hash := crypto.Keccak256([]byte("test message"))

			signature, err := signer.Sign(ctx, hash)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, signature)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, signature)
				assert.Len(t, signature, 65) // Ethereum signature is 65 bytes
			}
		})
	}
}

func TestSigner_Sign_Timeout(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	// Create a server that never completes the transaction
	txCreated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			txCreated = true
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "tx-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/tx-123" {
			// Always return pending status
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     "tx-123",
				"status": "PENDING_SIGNATURE",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	signer, err := fireblocks.NewSigner(
		testAPIKey, pemKey, testVaultAccountID, testAssetID,
		fireblocks.WithBaseURL(server.URL),
		fireblocks.WithTimeout(100*time.Millisecond),
		fireblocks.WithPollingInterval(10*time.Millisecond),
	)
	require.NoError(t, err)

	ctx := context.Background()
	hash := crypto.Keccak256([]byte("test message"))

	signature, err := signer.Sign(ctx, hash)
	require.Error(t, err)
	require.Nil(t, signature)
	assert.Contains(t, err.Error(), "timeout")
	assert.True(t, txCreated)
}

func TestSigner_Sign_ContextCancellation(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "tx-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/tx-123" {
			// Simulate slow response
			time.Sleep(500 * time.Millisecond)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     "tx-123",
				"status": "PENDING_SIGNATURE",
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	signer, err := fireblocks.NewSigner(
		testAPIKey, pemKey, testVaultAccountID, testAssetID,
		fireblocks.WithBaseURL(server.URL),
		fireblocks.WithTimeout(5*time.Second),
		fireblocks.WithPollingInterval(10*time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	hash := crypto.Keccak256([]byte("test message"))

	signature, err := signer.Sign(ctx, hash)
	require.Error(t, err)
	require.Nil(t, signature)
}

func TestSigner_GetVaultAccountAddress(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	tests := []struct {
		name            string
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		expectedAddress string
		wantErr         bool
		errContains     string
	}{
		{
			name: "successful get address",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(map[string]string{
					"address": "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4",
				})
			},
			expectedAddress: "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4",
			wantErr:         false,
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal error"}`))
			},
			wantErr:     true,
			errContains: "500",
		},
		{
			name: "not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "vault account not found"}`))
			},
			wantErr:     true,
			errContains: "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			signer, err := fireblocks.NewSigner(
				testAPIKey, pemKey, testVaultAccountID, testAssetID,
				fireblocks.WithBaseURL(server.URL),
			)
			require.NoError(t, err)

			ctx := context.Background()
			address, err := signer.GetVaultAccountAddress(ctx)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, address)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedAddress, address)
			}
		})
	}
}

func TestSigner_GenerateTestPrivateKey(t *testing.T) {
	key, pem, err := fireblocks.GenerateTestPrivateKey()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.NotEmpty(t, pem)
	assert.Contains(t, pem, "-----BEGIN RSA PRIVATE KEY-----")
	assert.Contains(t, pem, "-----END RSA PRIVATE KEY-----")
}

// Helper functions

func clearFireblocksEnvVars() {
	os.Unsetenv("FIREBLOCKS_API_KEY")
	os.Unsetenv("FIREBLOCKS_API_SECRET")
	os.Unsetenv("FIREBLOCKS_VAULT_ACCOUNT_ID")
	os.Unsetenv("FIREBLOCKS_ASSET_ID")
	os.Unsetenv("FIREBLOCKS_BASE_URL")
}

func createMockFireblocksServer(t *testing.T, signingKey *ecdsa.PrivateKey, pubKeyBytes []byte, finalStatus fireblocks.TransactionStatus) *httptest.Server {
	t.Helper()

	transactionPolls := 0
	var storedHash []byte

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)

			// Extract the hash from the request
			if extra, ok := payload["extraParameters"].(map[string]interface{}); ok {
				if rawData, ok := extra["rawMessageData"].(map[string]interface{}); ok {
					if messages, ok := rawData["messages"].([]interface{}); ok && len(messages) > 0 {
						if msg, ok := messages[0].(map[string]interface{}); ok {
							if content, ok := msg["content"].(string); ok {
								storedHash, _ = hex.DecodeString(content)
							}
						}
					}
				}
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "tx-mock-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/tx-mock-123" {
			transactionPolls++

			if transactionPolls < 2 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":     "tx-mock-123",
					"status": "PENDING_SIGNATURE",
				})
				return
			}

			if finalStatus != fireblocks.StatusCompleted {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":     "tx-mock-123",
					"status": string(finalStatus),
				})
				return
			}

			// Use the stored hash or a default
			hashToSign := storedHash
			if hashToSign == nil {
				hashToSign = crypto.Keccak256([]byte("test message"))
			}

			// Create mock signature with correct format
			sigBytes, _ := crypto.Sign(hashToSign, signingKey)

			rVal := hex.EncodeToString(sigBytes[:32])
			sVal := hex.EncodeToString(sigBytes[32:64])

			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     "tx-mock-123",
				"status": "COMPLETED",
				"signedMessages": []map[string]interface{}{
					{
						"content":   hex.EncodeToString(hashToSign),
						"algorithm": "MPC_ECDSA_SECP256K1",
						"signature": map[string]interface{}{
							"r":       rVal,
							"s":       sVal,
							"v":       int(sigBytes[64]),
							"fullSig": hex.EncodeToString(sigBytes),
						},
						"publicKey": hex.EncodeToString(pubKeyBytes),
					},
				},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
}
