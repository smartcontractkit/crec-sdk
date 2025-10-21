package transact

import (
	"context"
	"crypto"
	"crypto/rsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	vaultcontainer "github.com/testcontainers/testcontainers-go/modules/vault"

	"github.com/smartcontractkit/crec-sdk/client"
	"github.com/smartcontractkit/crec-sdk/mocks/server"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	"github.com/smartcontractkit/crec-sdk/transact/signer/vault"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

func TestHashOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	chainId := "31337"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := client.NewCRECClient(
		&client.ClientOptions{
			BaseURL: mockServer.TestServer.URL,
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transact, err := NewClient(
		&ClientOptions{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:      big.NewInt(1),
		Account: account,
		Transactions: []types.Transaction{
			{
				To:    to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	hash, err := transact.HashOperation(operation, chainId)
	if err != nil {
		t.Fatalf("Failed to hash operation: %v", err)
	}

	// check for pre-computed hash for the operation based on the above to/account
	require.Equal(t, "0xcd4308149652087bf9621b30e3d7781c475abb327b12b4e257966e88fa4a1ada", hash.Hex())
}

func TestSignOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	chainId := "31337"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	c, err := client.NewCRECClient(
		&client.ClientOptions{
			BaseURL: "http://localhost:8080",
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transact, err := NewClient(
		&ClientOptions{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:      big.NewInt(1),
		Account: account,
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
	opHash, sig, err := transact.SignOperation(context.Background(), operation, localSigner, chainId)
	require.NoError(t, err)

	// check for pre-computed signature for the operation based on the above to/account and private key
	require.Equal(
		t, "0xcd4308149652087bf9621b30e3d7781c475abb327b12b4e257966e88fa4a1ada", opHash.Hex(),
	)

	// check for pre-computed signature for the operation based on the above to/account and private key
	require.Equal(
		t,
		"5e1d5b835e963051f75e33bb8d20dd6464afe89268d53cfc06f3223ffcc1357b30f5fe9f75ceddf99792d9e1c877a3824bef0f79d522985723df46f3185ec75f1b",
		common.Bytes2Hex(sig),
	)
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
	vaultClient.SetAddress(vaultURL)
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
	chainId := "31337"
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := server.NewMockServer()
	t.Logf("Mock server started at URL: %s", mockServer.TestServer.URL)
	defer mockServer.Close()

	c, err := client.NewCRECClient(
		&client.ClientOptions{
			BaseURL: mockServer.TestServer.URL,
			APIKey:  "some-api-key",
		},
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transact, err := NewClient(
		&ClientOptions{
			CRECClient: c,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:      big.NewInt(1),
		Account: account,
		Transactions: []types.Transaction{
			{
				To:    to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	// Create our Vault Signer
	vaultSigner, err := vault.NewSigner(
		vaultURL,
		"myroot",
		"transit",
		keyName,
	)
	require.NoError(t, err)

	// Test signing the operation
	_, sig, err := transact.SignOperation(context.Background(), operation, vaultSigner, chainId)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	require.Greater(t, len(sig), 100, "Vault signature should be reasonably sized")
	require.Less(t, len(sig), 400, "Vault signature shouldn't be too large for RSA-2048")

	t.Logf("Vault Transit signature: %s", common.Bytes2Hex(sig))
	t.Logf("Vault Transit signature length: %d bytes", len(sig))

	// Get the public key from Vault to verify the signature
	pubKeyInterface, err := vaultSigner.Public()
	require.NoError(t, err)
	require.NotNil(t, pubKeyInterface)

	// Verify it's an RSA public key
	rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	require.True(t, ok, "Public key should be an RSA key")
	require.NotNil(t, rsaPubKey)

	// Get the operation hash for verification
	operationHash, err := transact.HashOperation(operation, chainId)
	require.NoError(t, err)

	// Verify the signature using the public key
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, operationHash.Bytes(), sig, nil)
	require.NoError(t, err, "Vault signature should be valid")

	// Test that we can sign the same operation multiple times
	opHash, sig2, err := transact.SignOperation(context.Background(), operation, vaultSigner, chainId)
	require.NoError(t, err)
	require.NotEmpty(t, sig2)

	// Verify the second signature as well
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, opHash.Bytes(), sig2, nil)
	require.NoError(t, err, "Second Vault signature should also be valid")

	// Signatures might be different due to RSA-PSS randomness
	t.Logf("Second Vault Transit signature: %s", common.Bytes2Hex(sig2))
}
