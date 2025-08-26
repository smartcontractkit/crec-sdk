package kms

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

const (
	envKMSKeyARN = "KMS_KEY_ARN"
)

func TestKMSSignerIntegration(t *testing.T) {
	// Check if KMS_KEY_ARN environment variable is set
	kmsKeyARN := os.Getenv(envKMSKeyARN)
	if kmsKeyARN == "" {
		t.Skipf("Skipping KMS integration test: %s environment variable not set", envKMSKeyARN)
	}

	ctx := t.Context()

	kmsSigner, err := NewSigner(ctx, kmsKeyARN)
	require.NoError(t, err, "Failed to create KMS signer")

	pubKey, err := GetPubKeyCtx(ctx, kmsSigner.client, kmsSigner.keyID)
	require.NoError(t, err, "Failed to get public key from KMS")

	kmsAddress := ethcrypto.PubkeyToAddress(*pubKey)
	t.Logf("KMS-derived Ethereum address: %s", kmsAddress.Hex())

	// Create a simulated blockchain with the KMS address pre-funded
	alloc := make(types.GenesisAlloc)
	alloc[kmsAddress] = types.Account{Balance: big.NewInt(1000000000000000000)} // 1 ETH

	sim := simulated.NewBackend(alloc, simulated.WithBlockGasLimit(10e6))
	defer sim.Close()

	targetAddress := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")

	nonce, err := sim.Client().PendingNonceAt(ctx, kmsAddress)
	require.NoError(t, err, "Failed to get nonce")

	gasPrice, err := sim.Client().SuggestGasPrice(ctx)
	require.NoError(t, err, "Failed to get gas price")

	tx := types.NewTransaction(
		nonce,
		targetAddress,
		big.NewInt(50000000000000000), // 0.05 ETH
		21000,
		gasPrice,
		nil,
	)

	chainID := big.NewInt(1337)
	signer := types.NewEIP155Signer(chainID)

	signedTx, err := signTransactionWithKMS(tx, signer, kmsSigner, ctx)
	require.NoError(t, err, "Failed to sign transaction with KMS")

	err = sim.Client().SendTransaction(ctx, signedTx)
	require.NoError(t, err, "Failed to send transaction to simulated blockchain")

	sim.Commit()

	receipt, err := sim.Client().TransactionReceipt(ctx, signedTx.Hash())
	require.NoError(t, err, "Failed to get transaction receipt")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "Transaction should be successful")

	targetBalance, err := sim.Client().BalanceAt(ctx, targetAddress, nil)
	require.NoError(t, err, "Failed to get target balance")

	require.Equal(t, big.NewInt(50000000000000000), targetBalance, "Target should have received 0.05 ETH")
}

func signTransactionWithKMS(tx *types.Transaction, signer types.Signer, kmsSigner *Signer, ctx context.Context) (*types.Transaction, error) {
	hash := signer.Hash(tx)

	signature, err := kmsSigner.Sign(ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}

	return tx.WithSignature(signer, signature)
}
