package wallets

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/transact/adminop"
	"github.com/smartcontractkit/crec-sdk/transact/eip712"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// mockSepoliaSelector is the chain selector the mock server seeds real
// deployments on, so the signing path can resolve a chain ID.
const mockSepoliaSelector = "16015286601757825753"

// mockTestOwnerKeyHex is a throwaway ECDSA key for mock-backed tests.
const mockTestOwnerKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

// mockTestEnv bundles a wallets client against the mock server with a test
// owner key and the EIP-712 handler.
type mockTestEnv struct {
	client  *Client
	handler *eip712.Handler
	signer  *local.Signer
	address common.Address
}

func setupMockEnv(t *testing.T) *mockTestEnv {
	t.Helper()
	client, _ := setupMockClient(t)
	handler, err := eip712.NewHandler(nil)
	require.NoError(t, err)
	key, err := crypto.HexToECDSA(mockTestOwnerKeyHex)
	require.NoError(t, err)
	return &mockTestEnv{
		client:  client,
		handler: handler,
		signer:  local.NewSigner(key),
		address: crypto.PubkeyToAddress(key.PublicKey),
	}
}

// createWallet creates an ECDSA wallet owned and signable by the test key on
// the mock's sepolia deployment.
func (e *mockTestEnv) createWallet(t *testing.T) *apiClient.Wallet {
	t.Helper()
	wallet, err := e.client.Create(context.Background(), CreateInput{
		Name:               "test-wallet-" + uuid8(),
		ChainSelector:      mockSepoliaSelector,
		WalletOwnerAddress: e.address.Hex(),
		WalletType:         apiClient.WalletTypeECDSA,
		Configuration: apiClient.WalletConfiguration{
			"allowed_signers": []string{e.address.Hex()},
		},
	})
	require.NoError(t, err)
	return wallet
}

var uuidCounter int64

func uuid8() string {
	uuidCounter++
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d-%x", uuidCounter, b[:])
}

// buildUpgradeInput builds a valid signed upgrade submission.
func (e *mockTestEnv) buildUpgradeInput(
	t *testing.T,
	wallet *apiClient.Wallet,
	targetVersion, implementationAddress string,
	authOpID, operationOpID int64,
) CreateUpgradeInput {
	t.Helper()
	ctx := context.Background()
	deadline := big.NewInt(time.Now().Unix() + 3600)
	walletAddress := common.HexToAddress(wallet.Address)

	auth := &eip712.AdminAuthorization{
		EIP712Type: eip712.UpgradeAuthorizationType,
		Arguments:  map[string]any{"newImplementation": implementationAddress},
		OpID:       big.NewInt(authOpID),
		Deadline:   deadline,
	}
	_, authSig, err := e.handler.SignAdminAuthorization(ctx, auth, e.signer, walletAddress, wallet.ChainSelector)
	require.NoError(t, err)
	calldata, err := adminop.PackUpgradeToWithAuthorization(
		common.HexToAddress(implementationAddress), auth.OpID, deadline, authSig)
	require.NoError(t, err)
	operation := &transactTypes.Operation{
		ID:       big.NewInt(operationOpID),
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := e.handler.SignOperation(ctx, operation, e.signer, wallet.ChainSelector)
	require.NoError(t, err)
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	return CreateUpgradeInput{
		TargetVersion:         targetVersion,
		ImplementationAddress: implementationAddress,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: auth.OpID.String(),
			Deadline:          deadline.Int64(),
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadline.Int64(),
			Signature:         &opSigHex,
			WalletOperationId: big.NewInt(operationOpID).String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	}
}

// buildAdminOperationInput builds a valid signed admin operation submission.
func (e *mockTestEnv) buildAdminOperationInput(
	t *testing.T,
	wallet *apiClient.Wallet,
	descriptor apiClient.WalletCapabilityDescriptor,
	capability string,
	arguments map[string]any,
	authOpID, operationOpID int64,
) CreateAdminOperationInput {
	t.Helper()
	ctx := context.Background()
	deadline := big.NewInt(time.Now().Unix() + 3600)
	walletAddress := common.HexToAddress(wallet.Address)

	auth := &eip712.AdminAuthorization{
		EIP712Type: descriptor.Eip712Type,
		Arguments:  arguments,
		OpID:       big.NewInt(authOpID),
		Deadline:   deadline,
	}
	_, authSig, err := e.handler.SignAdminAuthorization(ctx, auth, e.signer, walletAddress, wallet.ChainSelector)
	require.NoError(t, err)
	calldata, err := adminop.PackCapabilityCall(descriptor.AbiFragment, arguments, auth.OpID, deadline, authSig)
	require.NoError(t, err)
	operation := &transactTypes.Operation{
		ID:       big.NewInt(operationOpID),
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := e.handler.SignOperation(ctx, operation, e.signer, wallet.ChainSelector)
	require.NoError(t, err)
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	return CreateAdminOperationInput{
		Capability: capability,
		Arguments:  arguments,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: auth.OpID.String(),
			Deadline:          deadline.Int64(),
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadline.Int64(),
			Signature:         &opSigHex,
			WalletOperationId: big.NewInt(operationOpID).String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	}
}

// setSignerAllowedDescriptor returns the descriptor the mock's 2025.09.0
// version publishes for the ecdsa type.
func setSignerAllowedDescriptor(t *testing.T, client *Client) apiClient.WalletCapabilityDescriptor {
	t.Helper()
	version, err := client.GetTypeVersion(context.Background(), apiClient.WalletTypeECDSA, "2025.09.0")
	require.NoError(t, err)
	for _, c := range version.Capabilities {
		if c.Name == "setSignerAllowed" {
			return c
		}
	}
	t.Fatal("mock 2025.09.0 ecdsa version publishes no setSignerAllowed capability")
	return apiClient.WalletCapabilityDescriptor{}
}
