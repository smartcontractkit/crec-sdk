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
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
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
			name: "operation rejected",
			setupServer: func(t *testing.T) *httptest.Server {
				return createMockFireblocksServer(t, signingKey, pubKeyBytes, fireblocks.StatusRejected)
			},
			wantErr:     true,
			errContains: "REJECTED",
		},
		{
			name: "operation failed",
			setupServer: func(t *testing.T) *httptest.Server {
				return createMockFireblocksServer(t, signingKey, pubKeyBytes, fireblocks.StatusFailed)
			},
			wantErr:     true,
			errContains: "FAILED",
		},
		{
			name: "create operation error",
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

	// Create a server that never completes the operation
	opCreated := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			opCreated = true
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "op-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/op-123" {
			// Always return pending status
			json.NewEncoder(w).Encode(map[string]any{
				"id":     "op-123",
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
	// Either "timeout" (from our message) or "deadline exceeded" (from cancelled HTTP request) is valid
	errMsg := err.Error()
	assert.True(t, strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded"),
		"expected timeout or deadline exceeded error, got: %s", errMsg)
	assert.True(t, opCreated)
}

func TestSigner_Sign_ContextCancellation(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "op-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/op-123" {
			// Simulate slow response
			time.Sleep(500 * time.Millisecond)
			json.NewEncoder(w).Encode(map[string]any{
				"id":     "op-123",
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

func TestSigner_SignTypedData(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	// Generate a test signing key
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubKeyBytes := crypto.FromECDSAPub(&signingKey.PublicKey)

	tests := []struct {
		name        string
		typedData   *signer.TypedData
		wantErr     bool
		errContains string
	}{
		{
			name: "successful EIP-712 signing",
			typedData: &signer.TypedData{
				Types: map[string][]signer.TypedDataField{
					"Person": {
						{Name: "name", Type: "string"},
						{Name: "wallet", Type: "address"},
					},
					"Mail": {
						{Name: "from", Type: "Person"},
						{Name: "to", Type: "Person"},
						{Name: "contents", Type: "string"},
					},
				},
				PrimaryType: "Mail",
				Domain: signer.TypedDataDomain{
					Name:              "Ether Mail",
					Version:           "1",
					ChainID:           1,
					VerifyingContract: "0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
				},
				Message: map[string]any{
					"from": map[string]any{
						"name":   "Cow",
						"wallet": "0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826",
					},
					"to": map[string]any{
						"name":   "Bob",
						"wallet": "0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB",
					},
					"contents": "Hello, Bob!",
				},
			},
			wantErr: false,
		},
		{
			name: "simple permit message",
			typedData: &signer.TypedData{
				Types: map[string][]signer.TypedDataField{
					"Permit": {
						{Name: "owner", Type: "address"},
						{Name: "spender", Type: "address"},
						{Name: "value", Type: "uint256"},
						{Name: "nonce", Type: "uint256"},
						{Name: "deadline", Type: "uint256"},
					},
				},
				PrimaryType: "Permit",
				Domain: signer.TypedDataDomain{
					Name:              "MyToken",
					Version:           "1",
					ChainID:           11155111, // Sepolia
					VerifyingContract: "0x1234567890123456789012345678901234567890",
				},
				Message: map[string]any{
					"owner":    "0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826",
					"spender":  "0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB",
					"value":    "1000000000000000000",
					"nonce":    0,
					"deadline": 1893456000,
				},
			},
			wantErr: false,
		},
		{
			name:        "nil typed data",
			typedData:   nil,
			wantErr:     true,
			errContains: "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockTypedMessageServer(t, signingKey, pubKeyBytes, fireblocks.StatusCompleted)
			defer server.Close()

			s, err := fireblocks.NewSigner(
				testAPIKey, pemKey, testVaultAccountID, testAssetID,
				fireblocks.WithBaseURL(server.URL),
				fireblocks.WithTimeout(5*time.Second),
				fireblocks.WithPollingInterval(10*time.Millisecond),
			)
			require.NoError(t, err)

			ctx := context.Background()
			signature, err := s.SignTypedData(ctx, tt.typedData)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, signature)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, signature)
				assert.Len(t, signature, 65)
			}
		})
	}
}

func TestSigner_SignTypedData_OperationRejected(t *testing.T) {
	pemKey, _ := generateTestKey(t)

	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	pubKeyBytes := crypto.FromECDSAPub(&signingKey.PublicKey)

	server := createMockTypedMessageServer(t, signingKey, pubKeyBytes, fireblocks.StatusRejected)
	defer server.Close()

	s, err := fireblocks.NewSigner(
		testAPIKey, pemKey, testVaultAccountID, testAssetID,
		fireblocks.WithBaseURL(server.URL),
		fireblocks.WithTimeout(5*time.Second),
		fireblocks.WithPollingInterval(10*time.Millisecond),
	)
	require.NoError(t, err)

	typedData := &signer.TypedData{
		Types: map[string][]signer.TypedDataField{
			"Test": {{Name: "value", Type: "uint256"}},
		},
		PrimaryType: "Test",
		Domain: signer.TypedDataDomain{
			Name:    "Test",
			Version: "1",
			ChainID: 1,
		},
		Message: map[string]any{"value": 123},
	}

	ctx := context.Background()
	signature, err := s.SignTypedData(ctx, typedData)
	require.Error(t, err)
	require.Nil(t, signature)
	assert.Contains(t, err.Error(), "REJECTED")
}

// Helper functions

func clearFireblocksEnvVars() {
	os.Unsetenv("FIREBLOCKS_API_KEY")
	os.Unsetenv("FIREBLOCKS_API_SECRET")
	os.Unsetenv("FIREBLOCKS_VAULT_ACCOUNT_ID")
	os.Unsetenv("FIREBLOCKS_ASSET_ID")
	os.Unsetenv("FIREBLOCKS_BASE_URL")
}

func createMockFireblocksServer(t *testing.T, signingKey *ecdsa.PrivateKey, pubKeyBytes []byte, finalStatus fireblocks.OperationStatus) *httptest.Server {
	t.Helper()

	operationPolls := 0
	var storedHash []byte

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			var payload map[string]any
			json.NewDecoder(r.Body).Decode(&payload)

			// Extract the hash from the request
			if extra, ok := payload["extraParameters"].(map[string]any); ok {
				if rawData, ok := extra["rawMessageData"].(map[string]any); ok {
					if messages, ok := rawData["messages"].([]any); ok && len(messages) > 0 {
						if msg, ok := messages[0].(map[string]any); ok {
							if content, ok := msg["content"].(string); ok {
								storedHash, _ = hex.DecodeString(content)
							}
						}
					}
				}
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "op-mock-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/op-mock-123" {
			operationPolls++

			if operationPolls < 2 {
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "op-mock-123",
					"status": "PENDING_SIGNATURE",
				})
				return
			}

			if finalStatus != fireblocks.StatusCompleted {
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "op-mock-123",
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

			json.NewEncoder(w).Encode(map[string]any{
				"id":     "op-mock-123",
				"status": "COMPLETED",
				"signedMessages": []map[string]any{
					{
						"content":   hex.EncodeToString(hashToSign),
						"algorithm": "MPC_ECDSA_SECP256K1",
						"signature": map[string]any{
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

func createMockTypedMessageServer(t *testing.T, signingKey *ecdsa.PrivateKey, pubKeyBytes []byte, finalStatus fireblocks.OperationStatus) *httptest.Server {
	t.Helper()

	operationPolls := 0
	var storedTypedData *signer.TypedData

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/transactions" {
			var payload map[string]any
			json.NewDecoder(r.Body).Decode(&payload)

			// Verify this is a TYPED_MESSAGE operation
			operation, _ := payload["operation"].(string)
			if operation != "TYPED_MESSAGE" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "expected TYPED_MESSAGE operation"})
				return
			}

			// Parse the typed message data to reconstruct TypedData
			if extra, ok := payload["extraParameters"].(map[string]any); ok {
				if tmd, ok := extra["typedMessageData"].(map[string]any); ok {
					storedTypedData = parseTypedMessageData(tmd)
				}
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"id":     "op-typed-123",
				"status": "SUBMITTED",
			})
			return
		}

		if r.Method == "GET" && r.URL.Path == "/v1/transactions/op-typed-123" {
			operationPolls++

			if operationPolls < 2 {
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "op-typed-123",
					"status": "PENDING_SIGNATURE",
				})
				return
			}

			if finalStatus != fireblocks.StatusCompleted {
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "op-typed-123",
					"status": string(finalStatus),
				})
				return
			}

			// Compute the EIP-712 hash using the same function as the implementation
			var hashToSign []byte
			if storedTypedData != nil {
				var err error
				hashToSign, err = fireblocks.HashTypedData(storedTypedData)
				if err != nil {
					hashToSign = crypto.Keccak256([]byte("fallback-hash"))
				}
			} else {
				hashToSign = crypto.Keccak256([]byte("fallback-hash"))
			}

			// Create mock signature
			sigBytes, _ := crypto.Sign(hashToSign, signingKey)

			rVal := hex.EncodeToString(sigBytes[:32])
			sVal := hex.EncodeToString(sigBytes[32:64])

			json.NewEncoder(w).Encode(map[string]any{
				"id":     "op-typed-123",
				"status": "COMPLETED",
				"signedMessages": []map[string]any{
					{
						"content":   hex.EncodeToString(hashToSign),
						"algorithm": "MPC_ECDSA_SECP256K1",
						"signature": map[string]any{
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

// parseTypedMessageData reconstructs TypedData from the Fireblocks payload format
func parseTypedMessageData(tmd map[string]any) *signer.TypedData {
	td := &signer.TypedData{
		Types:   make(map[string][]signer.TypedDataField),
		Message: make(map[string]any),
	}

	// Parse types
	if types, ok := tmd["types"].(map[string]any); ok {
		for typeName, fields := range types {
			if fieldArr, ok := fields.([]any); ok {
				var parsedFields []signer.TypedDataField
				for _, f := range fieldArr {
					if field, ok := f.(map[string]any); ok {
						name, _ := field["name"].(string)
						typ, _ := field["type"].(string)
						parsedFields = append(parsedFields, signer.TypedDataField{Name: name, Type: typ})
					}
				}
				td.Types[typeName] = parsedFields
			}
		}
	}

	// Parse primaryType
	if pt, ok := tmd["primaryType"].(string); ok {
		td.PrimaryType = pt
	}

	// Parse domain
	if domain, ok := tmd["domain"].(map[string]any); ok {
		if name, ok := domain["name"].(string); ok {
			td.Domain.Name = name
		}
		if version, ok := domain["version"].(string); ok {
			td.Domain.Version = version
		}
		if chainID, ok := domain["chainId"].(float64); ok {
			td.Domain.ChainID = int64(chainID)
		}
		if vc, ok := domain["verifyingContract"].(string); ok {
			td.Domain.VerifyingContract = vc
		}
		if salt, ok := domain["salt"].(string); ok {
			td.Domain.Salt = salt
		}
	}

	// Parse message
	if msg, ok := tmd["message"].(map[string]any); ok {
		td.Message = msg
	}

	return td
}
