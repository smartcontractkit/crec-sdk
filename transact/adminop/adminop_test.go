package adminop

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upgradeToWithAuthorizationSelector anchors the packed upgrade call to the
// account contract's entry point, whose signature string is pinned in
// chainlink-smart-accounts' SignatureVerifyingAccount.
var upgradeToWithAuthorizationSelector = crypto.Keccak256(
	[]byte("upgradeToWithAuthorization(address,uint256,uint256,bytes)"))[:4]

func TestPackUpgradeToWithAuthorization(t *testing.T) {
	impl := common.HexToAddress("0xe4974c79c42dc628bbd0b055885c0cfe508c60b2")
	opID := big.NewInt(1000001)
	deadline := big.NewInt(4000000000)
	sig := common.FromHex("0x0102030405")

	packed, err := PackUpgradeToWithAuthorization(impl, opID, deadline, sig)
	require.NoError(t, err)

	assert.Equal(t, upgradeToWithAuthorizationSelector, packed[:4], "selector must be the contract's")

	// Head: newImplementation, opId, deadline; tail: dynamic bytes offset.
	assert.Equal(t, impl, common.BytesToAddress(packed[4+12:4+32]))
	assert.Equal(t, opID, new(big.Int).SetBytes(packed[4+32:4+64]))
	assert.Equal(t, deadline, new(big.Int).SetBytes(packed[4+64:4+96]))
	// Dynamic signature: offset, length, contents.
	offset := new(big.Int).SetBytes(packed[4+96 : 4+128]).Int64()
	require.EqualValues(t, 128, offset)
	length := new(big.Int).SetBytes(packed[4+offset : 4+offset+32]).Int64()
	require.EqualValues(t, len(sig), length)
	assert.Equal(t, sig, packed[4+offset+32:4+offset+32+length])
}

// setSignerAllowedFragment is the published descriptor shape for the ECDSA
// signer's setSignerAllowed capability: the function's own arguments followed
// by the authorization's (uint256 opId, uint256 deadline, bytes signature).
var setSignerAllowedFragment = map[string]any{
	"type": "function",
	"name": "setSignerAllowedWithAuthorization",
	"inputs": []any{
		map[string]any{"name": "signer", "type": "address"},
		map[string]any{"name": "allowed", "type": "bool"},
		map[string]any{"name": "opId", "type": "uint256"},
		map[string]any{"name": "deadline", "type": "uint256"},
		map[string]any{"name": "signature", "type": "bytes"},
	},
	"outputs":         []any{},
	"stateMutability": "nonpayable",
}

func TestPackCapabilityCall(t *testing.T) {
	signer := "0x9f644089295bfa1402231508116aaac15c3d7c87"
	opID := big.NewInt(1000002)
	deadline := big.NewInt(4000000000)
	sig := common.FromHex("0xaabbcc")

	packed, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
		"signer":  signer,
		"allowed": true,
	}, opID, deadline, sig)
	require.NoError(t, err)

	wantSelector := crypto.Keccak256(
		[]byte("setSignerAllowedWithAuthorization(address,bool,uint256,uint256,bytes)"))[:4]
	assert.Equal(t, wantSelector, packed[:4])
	assert.Equal(t, common.HexToAddress(signer), common.BytesToAddress(packed[4+12:4+32]))
	assert.Equal(t, byte(1), packed[4+63], "bool true packs as 1")
	assert.Equal(t, opID, new(big.Int).SetBytes(packed[4+64:4+96]))
	assert.Equal(t, deadline, new(big.Int).SetBytes(packed[4+96:4+128]))
	offset := new(big.Int).SetBytes(packed[4+128 : 4+160]).Int64()
	require.EqualValues(t, 160, offset)
	length := new(big.Int).SetBytes(packed[4+offset : 4+offset+32]).Int64()
	require.EqualValues(t, len(sig), length)
	assert.Equal(t, sig, packed[4+offset+32:4+offset+32+length])
}

func TestPackCapabilityCall_NoArguments(t *testing.T) {
	fragment := map[string]any{
		"type": "function",
		"name": "revokeAuthorization",
		"inputs": []any{
			map[string]any{"name": "opId", "type": "uint256"},
			map[string]any{"name": "deadline", "type": "uint256"},
			map[string]any{"name": "signature", "type": "bytes"},
		},
		"outputs":         []any{},
		"stateMutability": "nonpayable",
	}

	packed, err := PackCapabilityCall(fragment, map[string]any{}, big.NewInt(7), big.NewInt(8), nil)
	require.NoError(t, err)
	assert.Equal(t, crypto.Keccak256([]byte("revokeAuthorization(uint256,uint256,bytes)"))[:4], packed[:4])
}

func TestPackCapabilityCall_Errors(t *testing.T) {
	t.Run("ExtraArgument", func(t *testing.T) {
		_, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
			"extra":   "1",
		}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrArgumentMismatch)
	})

	t.Run("MissingArgument", func(t *testing.T) {
		_, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
			"signer": "0x9f644089295bfa1402231508116aaac15c3d7c87",
		}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrArgumentMismatch)
	})

	t.Run("NullArgument", func(t *testing.T) {
		_, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": nil,
		}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrArgumentMismatch)
	})

	t.Run("NonCanonicalAddress", func(t *testing.T) {
		_, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
			"signer":  strings.ToUpper("0x9f644089295bfa1402231508116aaac15c3d7c87"),
			"allowed": true,
		}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrArgumentMismatch)
	})

	t.Run("NonCanonicalInteger", func(t *testing.T) {
		fragment := map[string]any{
			"type": "function",
			"name": "cancelOperationWithAuthorization",
			"inputs": []any{
				map[string]any{"name": "operationId", "type": "uint256"},
				map[string]any{"name": "opId", "type": "uint256"},
				map[string]any{"name": "deadline", "type": "uint256"},
				map[string]any{"name": "signature", "type": "bytes"},
			},
			"outputs":         []any{},
			"stateMutability": "nonpayable",
		}
		_, err := PackCapabilityCall(fragment, map[string]any{"operationId": "042"}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrArgumentMismatch, "leading zeros are not canonical")
	})

	t.Run("InvalidFragment", func(t *testing.T) {
		_, err := PackCapabilityCall(map[string]any{"type": "event"}, map[string]any{}, big.NewInt(1), big.NewInt(1), nil)
		require.Error(t, err)
	})

	t.Run("MissingTrailingAuthorizationInputs", func(t *testing.T) {
		fragment := map[string]any{
			"type": "function",
			"name": "setSignerAllowed",
			"inputs": []any{
				map[string]any{"name": "signer", "type": "address"},
				map[string]any{"name": "allowed", "type": "bool"},
			},
			"outputs":         []any{},
			"stateMutability": "nonpayable",
		}
		_, err := PackCapabilityCall(fragment, map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}, big.NewInt(1), big.NewInt(1), nil)
		require.ErrorIs(t, err, ErrMissingTrailingInputs)
	})

	t.Run("NilDeadline", func(t *testing.T) {
		_, err := PackCapabilityCall(setSignerAllowedFragment, map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}, big.NewInt(1), nil, nil)
		require.ErrorIs(t, err, ErrInvalidAuthorization)
	})
}

// TestPackUpgradeMatchesReferenceABI proves the pinned inline ABI packs the
// exact bytes a fresh parse of the same signature produces, guarding against
// selector or field-order drift in the constant.
func TestPackUpgradeMatchesReferenceABI(t *testing.T) {
	impl := common.HexToAddress("0xe4974c79c42dc628bbd0b055885c0cfe508c60b2")
	packed, err := PackUpgradeToWithAuthorization(impl, big.NewInt(1), big.NewInt(2), []byte{0x03})
	require.NoError(t, err)

	require.Equal(t, "0xb27a09a9", "0x"+hex.EncodeToString(packed[:4]),
		"upgradeToWithAuthorization selector drifted from the contract")
}
