package credentialregistry

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/credentialregistry/gen/events"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		opts        *ServiceOptions
		expectError bool
	}{
		{
			name: "valid service options",
			opts: &ServiceOptions{
				CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
				AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
			},
			expectError: false,
		},
		{
			name: "empty addresses",
			opts: &ServiceOptions{
				CredentialRegistryAddress: "",
				AccountAddress:            "",
			},
			expectError: false, // Service creation should succeed, addresses are just converted to zero address
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.opts)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
				require.NotNil(t, service.logger)
			}
		})
	}
}

func TestPrepareRegisterCredentialOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	ccid := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	credentialTypeId := [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	expiresAt := uint64(1735689600) // 2025-01-01
	credentialData := []byte("test credential data")
	context := []byte("test context")

	operation, err := service.PrepareRegisterCredentialOperation(ccid, credentialTypeId, expiresAt, credentialData, context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRegisterCredentialsOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	ccid := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	credentialTypeIds := [][32]byte{
		{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32},
	}
	expiresAt := uint64(1735689600) // 2025-01-01
	credentialDatas := [][]byte{
		[]byte("first credential data"),
		[]byte("second credential data"),
	}
	context := []byte("batch registration context")

	operation, err := service.PrepareRegisterCredentialsOperation(ccid, credentialTypeIds, expiresAt, credentialDatas, context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRemoveCredentialOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	ccid := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	credentialTypeId := [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	context := []byte("removal context")

	operation, err := service.PrepareRemoveCredentialOperation(ccid, credentialTypeId, context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRenewCredentialOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	ccid := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	credentialTypeId := [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	expiresAt := uint64(1767225600) // 2026-01-01
	context := []byte("renewal context")

	operation, err := service.PrepareRenewCredentialOperation(ccid, credentialTypeId, expiresAt, context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareAttachPolicyEngineOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	policyEngine := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")

	operation, err := service.PrepareAttachPolicyEngineOperation(policyEngine)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareSetContextOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	context := []byte("new context data")

	operation, err := service.PrepareSetContextOperation(context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareClearContextOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	operation, err := service.PrepareClearContextOperation()
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareInitializeOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	policyEngine := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")

	operation, err := service.PrepareInitializeOperation(policyEngine)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareTransferOwnershipOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	newOwner := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareTransferOwnershipOperation(newOwner)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRenounceOwnershipOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	operation, err := service.PrepareRenounceOwnershipOperation()
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.credentialRegistryAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestDecodeCredentialRegistered(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.CredentialRegistryEvents{
		CredentialRegistered: &events.CredentialRegistryEvent{
			Event: &events.CredentialRegisteredEvent{
				Ccid:             "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				CredentialTypeId: "0x2010ffeeddccbbaa99887766554433221100ffeeddccbbaa9988776655443322",
				ExpiresAt:        "1735689600",
				CredentialData:   "0x746573742063726564656e7469616c2064617461",
			},
			CreatedAt: &time.Time{},
		},
	}

	jsonData, err := json.Marshal(mockEvent)
	require.NoError(t, err)

	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	event := &client.Event{
		VerifiableEvent: encodedData,
	}

	decodedEvent, err := service.DecodeCredentialRegistered(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestDecodeCredentialRemoved(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.CredentialRegistryEvents{
		CredentialRemoved: &events.CredentialRegistryEvent{
			Event: &events.CredentialRemovedEvent{
				Ccid:             "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				CredentialTypeId: "0x2010ffeeddccbbaa99887766554433221100ffeeddccbbaa9988776655443322",
			},
			CreatedAt: &time.Time{},
		},
	}

	jsonData, err := json.Marshal(mockEvent)
	require.NoError(t, err)

	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	event := &client.Event{
		VerifiableEvent: encodedData,
	}

	decodedEvent, err := service.DecodeCredentialRemoved(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestDecodeCredentialRenewed(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.CredentialRegistryEvents{
		CredentialRenewed: &events.CredentialRegistryEvent{
			Event: &events.CredentialRenewedEvent{
				Ccid:             "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				CredentialTypeId: "0x2010ffeeddccbbaa99887766554433221100ffeeddccbbaa9988776655443322",
				ExpiresAt:        "1767225600",
			},
			CreatedAt: &time.Time{},
		},
	}

	jsonData, err := json.Marshal(mockEvent)
	require.NoError(t, err)

	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	event := &client.Event{
		VerifiableEvent: encodedData,
	}

	decodedEvent, err := service.DecodeCredentialRenewed(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestToJson(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		CredentialRegistryAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid base64 JSON",
			input:       base64.StdEncoding.EncodeToString([]byte(`{"test": "data"}`)),
			expected:    `{"test": "data"}`,
			expectError: false,
		},
		{
			name:        "empty base64",
			input:       base64.StdEncoding.EncodeToString([]byte("")),
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid base64",
			input:       "invalid_base64_string_with_invalid_characters!!!",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &client.Event{
				VerifiableEvent: tc.input,
			}

			result, err := service.toJson(event)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, []byte(tc.expected), result)
			}
		})
	}
}
