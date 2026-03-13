package eip712

import (
	"context"
	"log/slog"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with default logger", func(t *testing.T) {
		handler, err := NewHandler(nil)
		require.NoError(t, err)
		require.NotNil(t, handler)
		assert.NotNil(t, handler.logger)
	})

	t.Run("creates handler with custom logger", func(t *testing.T) {
		logger := slog.Default()
		handler, err := NewHandler(&Options{Logger: logger})
		require.NoError(t, err)
		require.NotNil(t, handler)
		assert.Equal(t, logger, handler.logger)
	})
}

func TestHashOperation(t *testing.T) {
	handler, err := NewHandler(nil)
	require.NoError(t, err)

	// Base Sepolia chain selector
	chainSelector := "10344971235874465080"

	operation := &types.Operation{
		ID:      big.NewInt(12345),
		Account: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Transactions: []types.Transaction{
			{
				To:    common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				Value: big.NewInt(0),
				Data:  []byte{0x01, 0x02, 0x03},
			},
		},
	}

	t.Run("successfully hashes operation", func(t *testing.T) {
		hash, err := handler.HashOperation(operation, chainSelector)
		require.NoError(t, err)
		assert.NotEqual(t, common.Hash{}, hash)
	})

	t.Run("returns error for nil operation", func(t *testing.T) {
		_, err := handler.HashOperation(nil, chainSelector)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrOperationRequired)
	})

	t.Run("returns error for invalid chain selector", func(t *testing.T) {
		_, err := handler.HashOperation(operation, "invalid")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrParseChainSelector)
	})
}

func TestSignOperation(t *testing.T) {
	handler, err := NewHandler(nil)
	require.NoError(t, err)

	// Base Sepolia chain selector
	chainSelector := "10344971235874465080"

	// Create a test private key
	privateKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	require.NoError(t, err)
	localSigner := local.NewSigner(privateKey)

	operation := &types.Operation{
		ID:      big.NewInt(12345),
		Account: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Transactions: []types.Transaction{
			{
				To:    common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				Value: big.NewInt(0),
				Data:  []byte{0x01, 0x02, 0x03},
			},
		},
	}

	t.Run("successfully signs operation", func(t *testing.T) {
		hash, sig, err := handler.SignOperation(context.Background(), operation, localSigner, chainSelector)
		require.NoError(t, err)
		assert.NotEqual(t, common.Hash{}, hash)
		assert.NotEmpty(t, sig)
		assert.Equal(t, 65, len(sig)) // Ethereum signature length
	})

	t.Run("returns error for nil signer", func(t *testing.T) {
		_, _, err := handler.SignOperation(context.Background(), operation, nil, chainSelector)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSignerRequired)
	})

	t.Run("returns error for nil operation", func(t *testing.T) {
		_, _, err := handler.SignOperation(context.Background(), nil, localSigner, chainSelector)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrOperationRequired)
	})
}

func TestSignOperationHash(t *testing.T) {
	handler, err := NewHandler(nil)
	require.NoError(t, err)

	// Create a test private key
	privateKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	require.NoError(t, err)
	localSigner := local.NewSigner(privateKey)

	testHash := common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234")

	t.Run("successfully signs hash", func(t *testing.T) {
		sig, err := handler.SignOperationHash(context.Background(), testHash, localSigner)
		require.NoError(t, err)
		assert.NotEmpty(t, sig)
		assert.Equal(t, 65, len(sig)) // Ethereum signature length
	})

	t.Run("returns error for nil signer", func(t *testing.T) {
		_, err := handler.SignOperationHash(context.Background(), testHash, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSignerRequired)
	})
}

func TestHashOperationWithCustomDomainName(t *testing.T) {
	handler, err := NewHandler(nil)
	require.NoError(t, err)

	chainSelector := "10344971235874465080"

	operation := &types.Operation{
		ID:      big.NewInt(12345),
		Account: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Transactions: []types.Transaction{
			{
				To:    common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				Value: big.NewInt(0),
				Data:  []byte{0x01, 0x02, 0x03},
			},
		},
	}

	t.Run("default and custom domain produce different hashes", func(t *testing.T) {
		defaultHash, err := handler.HashOperation(operation, chainSelector)
		require.NoError(t, err)

		customHash, err := handler.HashOperation(operation, chainSelector, "CustomDomain")
		require.NoError(t, err)

		assert.NotEqual(t, defaultHash, customHash)
	})

	t.Run("same custom domain produces consistent hashes", func(t *testing.T) {
		hash1, err := handler.HashOperation(operation, chainSelector, "CustomDomain")
		require.NoError(t, err)

		hash2, err := handler.HashOperation(operation, chainSelector, "CustomDomain")
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("no domain name uses default", func(t *testing.T) {
		hashNoDomain, err := handler.HashOperation(operation, chainSelector)
		require.NoError(t, err)

		hashExplicitDefault, err := handler.HashOperation(operation, chainSelector, types.EIP712DomainName)
		require.NoError(t, err)

		assert.Equal(t, hashNoDomain, hashExplicitDefault)
	})
}

func TestGetChainIDFromSelector(t *testing.T) {
	t.Run("successfully extracts chain ID", func(t *testing.T) {
		// Base Sepolia chain selector
		chainSelector := "10344971235874465080"
		chainID, err := GetChainIDFromSelector(chainSelector)
		require.NoError(t, err)
		assert.NotNil(t, chainID)
		// Base Sepolia chain ID is 84532
		assert.Equal(t, big.NewInt(84532), chainID)
	})

	t.Run("returns error for invalid selector", func(t *testing.T) {
		_, err := GetChainIDFromSelector("invalid")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrParseChainSelector)
	})
}
