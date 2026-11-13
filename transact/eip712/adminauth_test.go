package eip712

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adminAuthorizationVectors mirrors the JSON published by chainlink-smart-accounts
// at test-vectors/admin_authorization_vectors.json, generated from the contracts
// with script/GenerateAdminAuthorizationVectors.s.sol. Replaying them in SDK CI
// proves the SDK and the contracts agree on the authorization encoding: the
// digest each vector carries is the contract-computed one.
type adminAuthorizationVectors struct {
	Description string `json:"description"`
	Domain      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"domain"`
	Vectors []struct {
		Capability    string         `json:"capability"`
		Family        string         `json:"family"`
		EIP712Type    string         `json:"eip712_type"`
		WalletAddress string         `json:"wallet_address"`
		ChainID       string         `json:"chain_id"`
		Arguments     map[string]any `json:"arguments"`
		OpID          string         `json:"op_id"`
		Deadline      string         `json:"deadline"`
		StructHash    string         `json:"struct_hash"`
		Digest        string         `json:"digest"`
		Signature     string         `json:"signature"`
		SignerAddress string         `json:"signer_address"`
		SigningKey    string         `json:"signing_key"`
	} `json:"vectors"`
}

func TestAdminAuthorizationContractVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/admin_authorization_vectors.json")
	require.NoError(t, err)

	var vectors adminAuthorizationVectors
	require.NoError(t, json.Unmarshal(raw, &vectors))
	require.NotEmpty(t, vectors.Vectors, "test vectors must not be empty")

	// The SDK's domain constants are the contracts' CLLSmartAccount domain.
	assert.Equal(t, vectors.Domain.Name, "CLLSmartAccount")
	assert.Equal(t, vectors.Domain.Version, "1")

	for i, v := range vectors.Vectors {
		name := fmt.Sprintf("%d-%s-%s", i, v.Capability, v.WalletAddress)
		t.Run(name, func(t *testing.T) {
			opID, ok := new(big.Int).SetString(v.OpID, 10)
			require.True(t, ok, "op_id must be a decimal integer")
			deadline, ok := new(big.Int).SetString(v.Deadline, 10)
			require.True(t, ok, "deadline must be a decimal integer")
			chainID, ok := new(big.Int).SetString(v.ChainID, 10)
			require.True(t, ok, "chain_id must be a decimal integer")

			auth := &AdminAuthorization{
				EIP712Type: v.EIP712Type,
				Arguments:  v.Arguments,
				OpID:       opID,
				Deadline:   deadline,
			}
			typedData, err := auth.typedData(common.HexToAddress(v.WalletAddress), chainID)
			require.NoError(t, err)

			// Struct hash agreement.
			primaryType, fields, err := ParseEIP712Type(v.EIP712Type)
			require.NoError(t, err)
			require.Equal(t, typedData.PrimaryType, primaryType)
			structHash, err := typedData.HashStruct(primaryType, typedData.Message)
			require.NoError(t, err)
			assert.Equal(t, v.StructHash, "0x"+common.Bytes2Hex(structHash),
				"struct hash mismatch for fields %v", fields)

			// Full domain-bound digest agreement.
			digest, _, err := apitypes.TypedDataAndHash(*typedData)
			require.NoError(t, err)
			assert.Equal(t, v.Digest, "0x"+common.Bytes2Hex(digest), "digest mismatch")

			// Signing the digest with the vector key reproduces the vector
			// signature and recovers the vector signer address.
			key, err := crypto.HexToECDSA(v.SigningKey[2:])
			require.NoError(t, err)
			sig, err := crypto.Sign(digest, key)
			require.NoError(t, err)

			recovered, err := crypto.SigToPub(digest, sig)
			require.NoError(t, err)
			assert.Equal(t, v.SignerAddress, strings.ToLower(crypto.PubkeyToAddress(*recovered).Hex()),
				"recovered signer mismatch")

			// The vector signature carries the Ethereum-adjusted recovery id.
			sig[64] += 27
			assert.Equal(t, v.Signature, "0x"+common.Bytes2Hex(sig), "signature mismatch")
		})
	}
}
