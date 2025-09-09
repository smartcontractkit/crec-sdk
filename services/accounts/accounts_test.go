package accounts

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants generated using cast commands for verification:
//
// Function selector:
// cast sig "createAccount(address,bytes32,address,address,bytes)"
const CREATE_ACCOUNT_FUNCTION_SELECTOR = "02ce00fc"

// Account ID hashes:
// cast keccak "test-ecdsa-account"
const EXPECTED_ECDSA_ACCOUNT_ID_HASH = "87e94958a78a68727b59e81038de1d5d6dcba9f3cce5a0844f209a4505d42543"

// cast keccak "test-rsa-account"
const EXPECTED_RSA_ACCOUNT_ID_HASH = "69ede2b7ba60fc58b6ff4de63ab8147ddea47009c14a43c65d4ba8611fcedcc3"

// ECDSA address array encoding:
// cast abi-encode "f(address[])" "[0x1234567890123456789012345678901234567890,0xabcdefabcdefabcdefabcdefabcdefabcdefabcd]"
const EXPECTED_ECDSA_INIT_CONFIG_DATA = "0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020000000000000000000000001234567890123456789012345678901234567890000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"

// Complete function call encodings:
//
//	cast abi-encode "createAccount(address,bytes32,address,address,bytes)" \
//	  "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc" \
//	  "0x87e94958a78a68727b59e81038de1d5d6dcba9f3cce5a0844f209a4505d42543" \
//	  "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" \
//	  "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4" \
//	  "0x${EXPECTED_ECDSA_INIT_CONFIG_DATA}"
const EXPECTED_ECDSA_FULL_CALLDATA = "02ce00fc000000000000000000000000ce2152bfcd0995f56a07dcbfef2bc85d404d65bc87e94958a78a68727b59e81038de1d5d6dcba9f3cce5a0844f209a4505d425430000000000000000000000009a9f2ccfde556a7e9ff0848998aa4a0cfd8863ae000000000000000000000000742d35cc6634c0532925a3b8d100d3f01f14bfe400000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000020000000000000000000000001234567890123456789012345678901234567890000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"

// RSA full calldata with detailed breakdown:
//
// Parameters for createAccount(address,bytes32,address,address,bytes):
// 1. implAddress: 0xd123456789012345678901234567890123456789
// 2. accountIdBytes32: cast keccak "test-rsa-account" = 0x69ede2b7ba60fc58b6ff4de63ab8147ddea47009c14a43c65d4ba8611fcedcc3
// 3. keystoneForwarder: 0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE
// 4. accountOwner: 0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4
// 5. initializationConfigData: RSAKey[] encoded as tuple[]
//
// RSA Keys used:
// Key 1: E=0x010001, N=0x00b3c1b86f8a5f4a8d2e7c9b1a5e3f8d6c4a9b2e5f1a7c8d3b6e9f2a5c8d1b4e7f9 (67 chars - odd, needs padding)
// Key 2: E=0x010001, N=0x00a1b2c3d4e5f6789012345678901234567890abcdef123456789abcdef012345678 (68 chars - even)
//
// Note: Cast requires even-length hex strings. For RSA encoding verification:
//
// Step 1 - Encode individual RSA keys as (bytes,bytes) tuples:
// cast abi-encode "f((bytes,bytes))" "(0x010001,0x0b3c1b86f8a5f4a8d2e7c9b1a5e3f8d6c4a9b2e5f1a7c8d3b6e9f2a5c8d1b4e7f9)"
// cast abi-encode "f((bytes,bytes))" "(0x010001,0x00a1b2c3d4e5f6789012345678901234567890abcdef123456789abcdef012345678)"
//
// Step 2 - Encode the array of RSA key tuples:
// cast abi-encode "f((bytes,bytes)[])" "[(0x010001,0x0b3c1b86f8a5f4a8d2e7c9b1a5e3f8d6c4a9b2e5f1a7c8d3b6e9f2a5c8d1b4e7f9),(0x010001,0x00a1b2c3d4e5f6789012345678901234567890abcdef123456789abcdef012345678)]"
//
// Step 3 - Final createAccount call (using the result from Step 2 as the last parameter):
//
//	cast abi-encode "createAccount(address,bytes32,address,address,bytes)" \
//	  "0xd123456789012345678901234567890123456789" \
//	  "0x69ede2b7ba60fc58b6ff4de63ab8147ddea47009c14a43c65d4ba8611fcedcc3" \
//	  "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" \
//	  "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4" \
//	  "0x${RSA_INIT_CONFIG_DATA_FROM_STEP_2}"
const EXPECTED_RSA_FULL_CALLDATA = "02ce00fc000000000000000000000000d12345678901234567890123456789012345678969ede2b7ba60fc58b6ff4de63ab8147ddea47009c14a43c65d4ba8611fcedcc30000000000000000000000009a9f2ccfde556a7e9ff0848998aa4a0cfd8863ae000000000000000000000000742d35cc6634c0532925a3b8d100d3f01f14bfe400000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000240000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000301000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000022000b3c1b86f8a5f4a8d2e7c9b1a5e3f8d6c4a9b2e5f1a7c8d3b6e9f2a5c8d1b4e7f90000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000030100010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002200a1b2c3d4e5f6789012345678901234567890abcdef123456789abcdef012345678000000000000000000000000000000000000000000000000000000000000"

func TestService_PrepareDeployNewECDSAAccountOperation(t *testing.T) {
	accountOwnerAddress := "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4"
	allowedSigners := []string{
		"0x1234567890123456789012345678901234567890",
		"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
	}
	accountId := "test-ecdsa-account"

	service, err := NewService(&ServiceOptions{
		Logger:                   &zerolog.Logger{},
		KeystoneForwarderAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountFactoryAddress:    "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
		ECDSASignatureVerifyingAccountImplAddress: "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		RSASignatureVerifyingAccountImplAddress:   "0xd123456789012345678901234567890123456789",
	})
	require.NoError(t, err)

	operation, err := service.PrepareDeployNewECDSAAccountOperation(accountOwnerAddress, allowedSigners, accountId)

	require.NoError(t, err)
	require.NotNil(t, operation)

	assert.NotNil(t, operation.ID)
	assert.Equal(t, service.accountFactoryAddress, operation.Account)
	assert.Len(t, operation.Transactions, 1)

	tx := operation.Transactions[0]
	assert.Equal(t, service.accountFactoryAddress, tx.To)
	assert.Equal(t, "0", tx.Value.String())
	assert.NotEmpty(t, tx.Data)

	actualFunctionSelector := hex.EncodeToString(tx.Data[:4])
	assert.Equal(t, CREATE_ACCOUNT_FUNCTION_SELECTOR, actualFunctionSelector)

	actualFullCalldata := hex.EncodeToString(tx.Data)
	assert.Equal(t, EXPECTED_ECDSA_FULL_CALLDATA, actualFullCalldata, "Full calldata should match expected encoding")

	// Verify account ID hash
	actualAccountIdHash := hex.EncodeToString(crypto.Keccak256([]byte(accountId)))
	assert.Equal(t, EXPECTED_ECDSA_ACCOUNT_ID_HASH, actualAccountIdHash, "Account ID hash should match expected")

	t.Logf("Function selector: 0x%s", actualFunctionSelector)
	t.Logf("Full calldata: 0x%s", actualFullCalldata)

	// Log the components for manual verification
	t.Logf("ECDSA Implementation Address: %s", service.ecdsaSignatureVerifyingAccountImplAddress.Hex())
	t.Logf("Account ID (keccak256): 0x%s", actualAccountIdHash)
	t.Logf("Keystone Forwarder: %s", service.keystoneForwarderAddress.Hex())
	t.Logf("Account Owner: %s", accountOwnerAddress)
}

func TestService_PrepareDeployNewRSAAccountOperation(t *testing.T) {
	// Test data
	accountOwnerAddress := "0x742d35Cc6634C0532925a3b8D100d3F01F14bFE4"
	allowedSigners := []RSAKey{
		{
			E: "0x010001",                                                                   // Standard RSA exponent (65537)
			N: "0x" + "00b3c1b86f8a5f4a8d2e7c9b1a5e3f8d6c4a9b2e5f1a7c8d3b6e9f2a5c8d1b4e7f9", // Example 256-bit modulus (padded)
		},
		{
			E: "0x010001",
			N: "0x" + "00a1b2c3d4e5f6789012345678901234567890abcdef123456789abcdef012345678",
		},
	}
	accountId := "test-rsa-account"

	// Create service
	service, err := NewService(&ServiceOptions{
		Logger:                   &zerolog.Logger{},
		KeystoneForwarderAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountFactoryAddress:    "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
		ECDSASignatureVerifyingAccountImplAddress: "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		RSASignatureVerifyingAccountImplAddress:   "0xd123456789012345678901234567890123456789",
	})
	require.NoError(t, err)

	operation, err := service.PrepareDeployNewRSAAccountOperation(accountOwnerAddress, allowedSigners, accountId)

	require.NoError(t, err)
	require.NotNil(t, operation)

	assert.NotNil(t, operation.ID)
	assert.Equal(t, service.accountFactoryAddress, operation.Account)
	assert.Len(t, operation.Transactions, 1)

	tx := operation.Transactions[0]
	assert.Equal(t, service.accountFactoryAddress, tx.To)
	assert.Equal(t, "0", tx.Value.String())
	assert.NotEmpty(t, tx.Data)

	actualFunctionSelector := hex.EncodeToString(tx.Data[:4])
	assert.Equal(t, CREATE_ACCOUNT_FUNCTION_SELECTOR, actualFunctionSelector)

	actualFullCalldata := hex.EncodeToString(tx.Data)
	assert.Equal(t, EXPECTED_RSA_FULL_CALLDATA, actualFullCalldata, "Full calldata should match expected encoding")

	actualAccountIdHash := hex.EncodeToString(crypto.Keccak256([]byte(accountId)))
	assert.Equal(t, EXPECTED_RSA_ACCOUNT_ID_HASH, actualAccountIdHash, "Account ID hash should match expected")

	t.Logf("Function selector: 0x%s", actualFunctionSelector)
	t.Logf("Full calldata: 0x%s", actualFullCalldata)

	t.Logf("RSA Implementation Address: %s", service.rsaSignatureVerifyingAccountImplAddress.Hex())
	t.Logf("Account ID (keccak256): 0x%s", actualAccountIdHash)
	t.Logf("Keystone Forwarder: %s", service.keystoneForwarderAddress.Hex())
	t.Logf("Account Owner: %s", accountOwnerAddress)
}
