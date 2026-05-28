package transact

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/hashicorp/vault/api"
	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	vaultcontainer "github.com/testcontainers/testcontainers-go/modules/vault"

	"github.com/smartcontractkit/crec-sdk/mocks/server"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	vaultSigner "github.com/smartcontractkit/crec-sdk/transact/signer/vault"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

func TestHashOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	// chainId := "31337"
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := apiClient.NewClientWithResponses(
		mockServer.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey some-api-key")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transactClient, err := NewClient(
		&Options{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:       big.NewInt(1),
		Account:  account,
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{
				To:    to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	hash, err := transactClient.HashOperation(operation, chainSelector)
	if err != nil {
		t.Fatalf("Failed to hash operation: %v", err)
	}

	// check for pre-computed hash for the operation based on the above to/account
	require.Equal(t, "0x411a1f5cc217aae5e44d3145aa7967a19ac788b1decd082c6ac5ccae2d2e6d98", hash.Hex())
}

func TestSignOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	// chainId := "31337"
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := apiClient.NewClientWithResponses(
		mockServer.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey some-api-key")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transactClient, err := NewClient(
		&Options{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:       big.NewInt(1),
		Account:  account,
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{
				To:    to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	privateKey, err := ethcrypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	localSigner := local.NewSigner(privateKey)
	opHash, sig, err := transactClient.SignOperation(context.Background(), operation, localSigner, chainSelector)
	require.NoError(t, err)

	// check for pre-computed signature for the operation based on the above to/account and private key
	require.Equal(
		t, "0x411a1f5cc217aae5e44d3145aa7967a19ac788b1decd082c6ac5ccae2d2e6d98", opHash.Hex(),
	)

	// check for pre-computed signature for the operation based on the above to/account and private key
	require.Equal(
		t,
		"0c600835117288b37f33e0f100886c1180ddb5b2d0b8e2dcb1f5ac87f2e62b53683cbc90600f4267aa0bbe695d813c531b36ecb0eff8030cb53d29df5b4cc60b1c",
		common.Bytes2Hex(sig),
	)
}

func TestSendSignedOperation(t *testing.T) {
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	channelID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name        string
		operation   *types.Operation
		signature   []byte
		expectError bool
		errorIs     error
	}{
		{
			name: "Success",
			operation: &types.Operation{
				ID:       big.NewInt(1),
				Account:  account,
				Deadline: big.NewInt(0),
				Transactions: []types.Transaction{
					{To: to, Value: big.NewInt(0), Data: []byte("")},
				},
			},
			signature:   []byte("test-signature"),
			expectError: false,
		},
		{
			name: "NilOperationID",
			operation: &types.Operation{
				ID:       nil,
				Account:  account,
				Deadline: big.NewInt(0),
				Transactions: []types.Transaction{
					{To: to, Value: big.NewInt(0), Data: []byte("")},
				},
			},
			signature:   []byte("test-signature"),
			expectError: true,
			errorIs:     ErrWalletOperationIDRequired,
		},
		{
			name:        "NilOperation",
			operation:   nil,
			signature:   []byte("test-signature"),
			expectError: true,
			errorIs:     ErrOperationRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := server.NewMockServer()
			defer mockServer.Close()

			c, err := apiClient.NewClientWithResponses(
				mockServer.TestServer.URL,
				apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
					req.Header.Set("Authorization", "Apikey some-api-key")
					return nil
				}),
			)
			require.NoError(t, err)

			transactClient, err := NewClient(&Options{CRECClient: c})
			require.NoError(t, err)

			parsedChannelID := uuid.MustParse(channelID)
			op, err := transactClient.SendSignedOperation(
				context.Background(),
				parsedChannelID,
				tt.operation,
				tt.signature,
				chainSelector,
			)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, op)
				if tt.errorIs != nil {
					require.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, op)
			}
		})
	}
}

func TestExecuteOperation(t *testing.T) {
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	channelID := "550e8400-e29b-41d4-a716-446655440000"

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := apiClient.NewClientWithResponses(
		mockServer.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey some-api-key")
			return nil
		}),
	)
	require.NoError(t, err)

	transactClient, err := NewClient(&Options{CRECClient: c})
	require.NoError(t, err)

	privateKey, err := ethcrypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	localSigner := local.NewSigner(privateKey)

	operation := &types.Operation{
		ID:       big.NewInt(time.Now().Unix()),
		Account:  account,
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{To: to, Value: big.NewInt(0), Data: []byte("")},
		},
	}

	parsedChannelID := uuid.MustParse(channelID)
	op, err := transactClient.ExecuteOperation(
		context.Background(),
		parsedChannelID,
		localSigner,
		operation,
		chainSelector,
	)

	require.NoError(t, err)
	require.NotNil(t, op)
}

func TestExecuteTransactions(t *testing.T) {
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	channelID := "550e8400-e29b-41d4-a716-446655440000"

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := apiClient.NewClientWithResponses(
		mockServer.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey some-api-key")
			return nil
		}),
	)
	require.NoError(t, err)

	transactClient, err := NewClient(&Options{CRECClient: c})
	require.NoError(t, err)

	privateKey, err := ethcrypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	localSigner := local.NewSigner(privateKey)

	txs := []types.Transaction{
		{To: to, Value: big.NewInt(0), Data: []byte("")},
		{To: to, Value: big.NewInt(100), Data: []byte("0x1234")},
	}

	parsedChannelID := uuid.MustParse(channelID)
	op, err := transactClient.ExecuteTransactions(
		context.Background(),
		parsedChannelID,
		localSigner,
		account,
		txs,
		big.NewInt(0),
		chainSelector,
	)

	require.NoError(t, err)
	require.NotNil(t, op)
}

func TestSignOperationWithVaultTransit(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vaultcontainer.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vaultcontainer.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	vaultClient, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	require.NoError(t, vaultClient.SetAddress(vaultURL))
	vaultClient.SetToken("myroot")

	// Enable transit secrets engine
	err = vaultClient.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create RSA key for signing
	keyName := "test-bank-rsa-key"
	_, err = vaultClient.Logical().Write(
		fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
			"type": "rsa-2048",
		},
	)
	require.NoError(t, err)

	// Set up the same test scenario as TestSignOperation
	// chainId := "31337"
	chainSelector := "7759470850252068959"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := apiClient.NewClientWithResponses(
		mockServer.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey some-api-key")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transactClient, err := NewClient(
		&Options{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:       big.NewInt(1),
		Account:  account,
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{
				To:    to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	// Create our Vault Signer
	vaultSignerInst, err := vaultSigner.NewSigner(
		vaultURL,
		"myroot",
		"transit",
		keyName,
	)
	require.NoError(t, err)

	// Test signing the operation
	_, sig, err := transactClient.SignOperation(context.Background(), operation, vaultSignerInst, chainSelector)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	require.Greater(t, len(sig), 100, "Vault signature should be reasonably sized")
	require.Less(t, len(sig), 400, "Vault signature shouldn't be too large for RSA-2048")

	t.Logf("Vault Transit signature: %s", common.Bytes2Hex(sig))
	t.Logf("Vault Transit signature length: %d bytes", len(sig))

	// Get the public key from Vault to verify the signature
	pubKeyInterface, err := vaultSignerInst.Public()
	require.NoError(t, err)
	require.NotNil(t, pubKeyInterface)

	// Verify it's an RSA public key
	rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	require.True(t, ok, "Public key should be an RSA key")
	require.NotNil(t, rsaPubKey)

	// Get the operation hash for verification
	operationHash, err := transactClient.HashOperation(operation, chainSelector)
	require.NoError(t, err)

	// Verify the signature using the public key
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, operationHash.Bytes(), sig, nil)
	require.NoError(t, err, "Vault signature should be valid")

	// Test that we can sign the same operation multiple times
	opHash, sig2, err := transactClient.SignOperation(context.Background(), operation, vaultSignerInst, chainSelector)
	require.NoError(t, err)
	require.NotEmpty(t, sig2)

	// Verify the second signature as well
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, opHash.Bytes(), sig2, nil)
	require.NoError(t, err, "Second Vault signature should also be valid")

	// Signatures might be different due to RSA-PSS randomness
	t.Logf("Second Vault Transit signature: %s", common.Bytes2Hex(sig2))
}

func TestClient_CreateUnsignedDraftOperation_Success(t *testing.T) {
	channelID := uuid.New()
	operationID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/channels/" + channelID.String() + "/operations"
		require.Equal(t, expectedPath, r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var createReq apiClient.CreateOperation
		err = json.Unmarshal(body, &createReq)
		require.NoError(t, err)

		require.Equal(t, "1337", createReq.ChainSelector)
		require.Equal(t, "0x1234", createReq.Address)
		require.Equal(t, "op-123", createReq.WalletOperationId)
		require.Len(t, createReq.Transactions, 1)
		require.Nil(t, createReq.Transactions[0].Preview)
		require.Nil(t, createReq.Signature)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(apiClient.OperationResponse{OperationId: operationID}))
	}

	client, server := setupTestClient(t, handler)
	defer server.Close()

	opID, err := client.CreateUnsignedDraftOperation(context.Background(), channelID, CreateDraftOperationInput{
		ChainSelector:     "1337",
		Address:           "0x1234",
		WalletOperationID: "op-123",
		Deadline:          0,
		Transactions: []DraftTransactionRequest{
			{To: "0x5678", Value: "0", Data: "0xabcd"},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, opID)
	require.Equal(t, operationID, *opID)
}

func TestClient_CreateUnsignedDraftOperation_ValidationErrors(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be called for invalid input")
	}
	client, server := setupTestClient(t, handler)
	defer server.Close()

	testCases := []struct {
		name          string
		channelID     uuid.UUID
		input         CreateDraftOperationInput
		expectedError error
	}{
		{
			name:      "NilChannelID",
			channelID: uuid.Nil,
			input: CreateDraftOperationInput{
				ChainSelector:     "1337",
				Address:           "0x1234",
				WalletOperationID: "op-123",
				Transactions:      []DraftTransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
			},
			expectedError: ErrChannelIDRequired,
		},
		{
			name:      "EmptyChainSelector",
			channelID: uuid.New(),
			input: CreateDraftOperationInput{
				ChainSelector:     "",
				Address:           "0x1234",
				WalletOperationID: "op-123",
				Transactions:      []DraftTransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
			},
			expectedError: ErrChainSelectorRequired,
		},
		{
			name:      "EmptyAddress",
			channelID: uuid.New(),
			input: CreateDraftOperationInput{
				ChainSelector:     "1337",
				Address:           "",
				WalletOperationID: "op-123",
				Transactions:      []DraftTransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
			},
			expectedError: ErrAddressRequired,
		},
		{
			name:      "EmptyWalletOperationID",
			channelID: uuid.New(),
			input: CreateDraftOperationInput{
				ChainSelector:     "1337",
				Address:           "0x1234",
				WalletOperationID: "",
				Transactions:      []DraftTransactionRequest{{To: "0x5678", Value: "0", Data: "0xabcd"}},
			},
			expectedError: ErrWalletOperationIDRequired,
		},
		{
			name:      "EmptyTransactions",
			channelID: uuid.New(),
			input: CreateDraftOperationInput{
				ChainSelector:     "1337",
				Address:           "0x1234",
				WalletOperationID: "op-123",
				Transactions:      []DraftTransactionRequest{},
			},
			expectedError: ErrAtLeastOneTransactionRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.CreateUnsignedDraftOperation(context.Background(), tc.channelID, tc.input)
			require.Error(t, err)
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestClient_SendDraftOperation_SuccessWithPreview(t *testing.T) {
	channelID := uuid.New()
	operationID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var createReq apiClient.CreateOperation
		err = json.Unmarshal(body, &createReq)
		require.NoError(t, err)

		require.Len(t, createReq.Transactions, 1)
		require.NotNil(t, createReq.Transactions[0].Preview)
		require.Equal(t, "transfer(address,uint256)", createReq.Transactions[0].Preview.FunctionSignature)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(apiClient.OperationResponse{OperationId: operationID}))
	}

	client, server := setupTestClient(t, handler)
	defer server.Close()

	op := &types.Operation{
		ID:       big.NewInt(42),
		Account:  common.HexToAddress("0x1234"),
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{To: common.HexToAddress("0x5678"), Value: big.NewInt(0), Data: []byte{0xab, 0xcd}},
		},
	}

	preview := &DraftTransactionPreview{FunctionSignature: "transfer(address,uint256)"}

	opID, err := client.SendDraftOperation(context.Background(), channelID, op, "1337", []*DraftTransactionPreview{preview})
	require.NoError(t, err)
	require.NotNil(t, opID)
	require.Equal(t, operationID, *opID)
}

func TestClient_SendDraftOperation_ValidationErrors(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be called for validation errors")
	})
	defer server.Close()

	testCases := []struct {
		name          string
		op            *types.Operation
		previews      []*DraftTransactionPreview
		expectedError error
	}{
		{
			name: "PreviewCountMismatch",
			op: &types.Operation{
				ID:       big.NewInt(1),
				Account:  common.HexToAddress("0x1234"),
				Deadline: big.NewInt(0),
				Transactions: []types.Transaction{
					{To: common.HexToAddress("0x5678"), Value: big.NewInt(0), Data: []byte{0xab}},
					{To: common.HexToAddress("0x9999"), Value: big.NewInt(0), Data: []byte{0xcd}},
				},
			},
			previews:      []*DraftTransactionPreview{{FunctionSignature: "f()"}},
			expectedError: ErrTransactionPreviewCountMismatch,
		},
		{
			name: "NilOperationID",
			op: &types.Operation{
				ID:       nil,
				Account:  common.HexToAddress("0x1234"),
				Deadline: big.NewInt(0),
				Transactions: []types.Transaction{
					{To: common.HexToAddress("0x5678"), Value: big.NewInt(0), Data: []byte{0xab}},
				},
			},
			previews:      nil,
			expectedError: ErrWalletOperationIDRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.SendDraftOperation(context.Background(), uuid.New(), tc.op, "1337", tc.previews)
			require.Error(t, err)
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

type stubSigner struct {
	lastDigest []byte
}

func (s *stubSigner) Sign(_ context.Context, digest []byte) ([]byte, error) {
	s.lastDigest = append([]byte{}, digest...)
	return []byte{0xaa, 0xbb}, nil
}

func TestClient_SendSignedDraftOperation_Success(t *testing.T) {
	channelID := uuid.New()
	operationID := uuid.New()
	digest := []byte{0x11, 0x22}
	signature := []byte{0x33, 0x44}

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var patchReq apiClient.PatchOperation
		err = json.Unmarshal(body, &patchReq)
		require.NoError(t, err)

		finalize, err := patchReq.AsFinalizeOperation()
		require.NoError(t, err)
		require.Equal(t, apiClient.FinalizeOperationStatusAccepted, finalize.Status)
		require.Equal(t, "0x"+common.Bytes2Hex(signature), finalize.Signature)
		require.Equal(t, "0x"+common.Bytes2Hex(digest), finalize.Digest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(apiClient.Operation{}))
	}

	client, server := setupTestClient(t, handler)
	defer server.Close()

	_, err := client.SendSignedDraftOperation(context.Background(), channelID, operationID, digest, signature)
	require.NoError(t, err)
}

func TestClient_ExecuteDraftOperation_SignsProvidedDigest(t *testing.T) {
	channelID := uuid.New()
	operationID := uuid.New()
	digest := []byte{0xde, 0xad}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(apiClient.Operation{}))
	}

	client, server := setupTestClient(t, handler)
	defer server.Close()

	signer := &stubSigner{}
	_, err := client.ExecuteDraftOperation(context.Background(), channelID, operationID, digest, signer)
	require.NoError(t, err)
	require.Equal(t, digest, signer.lastDigest)
}

func TestClient_CancelDraftOperation(t *testing.T) {
	testCases := []struct {
		name          string
		handler       func(t *testing.T, w http.ResponseWriter, r *http.Request)
		expectedError error
	}{
		{
			name: "Success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var patchReq apiClient.PatchOperation
				require.NoError(t, json.Unmarshal(body, &patchReq))

				cancel, err := patchReq.AsCancelOperation()
				require.NoError(t, err)
				require.Equal(t, apiClient.CancelOperationStatusCancelled, cancel.Status)

				w.WriteHeader(http.StatusNoContent)
			},
			expectedError: nil,
		},
		{
			name: "Conflict",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
			},
			expectedError: ErrDraftNotCancellable,
		},
		{
			name: "NotFound",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedError: ErrDraftNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channelID := uuid.New()
			operationID := uuid.New()

			client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			})
			defer server.Close()

			err := client.CancelDraftOperation(context.Background(), channelID, operationID)
			if tc.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.expectedError)
			}
		})
	}
}
