package identityvalidatorpolicy

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/identityvalidatorpolicy/gen/events"
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
				IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
				AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
			},
			expectError: false,
		},
		{
			name: "empty addresses",
			opts: &ServiceOptions{
				IdentityValidatorPolicyAddress: "",
				AccountAddress:                 "",
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

func TestPrepareAddCredentialRequirementOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	input := CredentialRequirementInput{
		RequirementId: [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		CredentialTypeIds: [][32]byte{
			{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		MinValidations: big.NewInt(1),
		Invert:         false,
	}

	operation, err := service.PrepareAddCredentialRequirementOperation(input)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareAddCredentialSourceOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	input := CredentialSourceInput{
		CredentialTypeId:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		IdentityRegistry:   common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
		CredentialRegistry: common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
		DataValidator:      common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
	}

	operation, err := service.PrepareAddCredentialSourceOperation(input)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRemoveCredentialRequirementOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	requirementId := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	operation, err := service.PrepareRemoveCredentialRequirementOperation(requirementId)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRemoveCredentialSourceOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	credentialTypeId := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	identityRegistry := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
	credentialRegistry := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareRemoveCredentialSourceOperation(credentialTypeId, identityRegistry, credentialRegistry)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareInitializeWithCredentialsOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	credentialSourceInputs := []CredentialSourceInput{
		{
			CredentialTypeId:   [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			IdentityRegistry:   common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
			CredentialRegistry: common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
			DataValidator:      common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
	}

	credentialRequirementInputs := []CredentialRequirementInput{
		{
			RequirementId: [32]byte{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			CredentialTypeIds: [][32]byte{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			MinValidations: big.NewInt(1),
			Invert:         false,
		},
	}

	operation, err := service.PrepareInitializeWithCredentialsOperation(credentialSourceInputs, credentialRequirementInputs)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareInitializeWithPolicyEngineOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	policyEngine := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
	initialOwner := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	configParams := []byte("test config parameters")

	operation, err := service.PrepareInitializeWithPolicyEngineOperation(policyEngine, initialOwner, configParams)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareOnInstallOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	selector := [4]byte{0x12, 0x34, 0x56, 0x78}

	operation, err := service.PrepareOnInstallOperation(selector)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareOnUninstallOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	selector := [4]byte{0x87, 0x65, 0x43, 0x21}

	operation, err := service.PrepareOnUninstallOperation(selector)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPreparePostRunOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	sender := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
	target := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	selector := [4]byte{0xab, 0xcd, 0xef, 0x12}
	parameters := [][]byte{
		[]byte("param1"),
		[]byte("param2"),
	}
	context := []byte("test context")

	operation, err := service.PreparePostRunOperation(sender, target, selector, parameters, context)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareTransferOwnershipOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
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
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRenounceOwnershipOperation(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
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
	require.Equal(t, service.identityValidatorPolicyAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestDecodeCredentialRequirementAdded(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.IdentityValidatorPolicyEvents{
		CredentialRequirementAdded: &events.IdentityValidatorPolicyEvent{
			Event: &events.CredentialRequirementAddedEvent{
				RequirementId:     "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				CredentialTypeIds: []string{"0x2010ffeeddccbbaa99887766554433221100ffeeddccbbaa9988776655443322"},
				MinValidations:    "1",
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

	decodedEvent, err := service.DecodeCredentialRequirementAdded(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestDecodeCredentialSourceAdded(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.IdentityValidatorPolicyEvents{
		CredentialSourceAdded: &events.IdentityValidatorPolicyEvent{
			Event: &events.CredentialSourceAddedEvent{
				CredentialTypeId:   "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				IdentityRegistry:   "0xA5F12FDA3e8B7209a3019141F105e5DB43445B86",
				CredentialRegistry: "0xeb457346d2218f7f77aa23ac6d9e394b505dd621",
				DataValidator:      "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
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

	decodedEvent, err := service.DecodeCredentialSourceAdded(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestDecodeCredentialRequirementRemoved(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.IdentityValidatorPolicyEvents{
		CredentialRequirementRemoved: &events.IdentityValidatorPolicyEvent{
			Event: &events.CredentialRequirementRemovedEvent{
				RequirementId:     "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				CredentialTypeIds: []string{"0x2010ffeeddccbbaa99887766554433221100ffeeddccbbaa9988776655443322"},
				MinValidations:    "1",
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

	decodedEvent, err := service.DecodeCredentialRequirementRemoved(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestDecodeCredentialSourceRemoved(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
	})
	require.NoError(t, err)

	// Create mock event data
	mockEvent := &events.IdentityValidatorPolicyEvents{
		CredentialSourceRemoved: &events.IdentityValidatorPolicyEvent{
			Event: &events.CredentialSourceRemovedEvent{
				CredentialTypeId:   "0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
				IdentityRegistry:   "0xA5F12FDA3e8B7209a3019141F105e5DB43445B86",
				CredentialRegistry: "0xeb457346d2218f7f77aa23ac6d9e394b505dd621",
				DataValidator:      "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
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

	decodedEvent, err := service.DecodeCredentialSourceRemoved(event)
	require.NoError(t, err)
	require.NotNil(t, decodedEvent)
}

func TestToJson(t *testing.T) {
	service, err := NewService(&ServiceOptions{
		IdentityValidatorPolicyAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:                 "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
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
