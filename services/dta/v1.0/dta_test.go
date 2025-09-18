package dta

import (
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
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
				DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
				DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
				AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
			},
			expectError: false,
		},
		{
			name: "empty addresses",
			opts: &ServiceOptions{
				DTARequestManagementAddress: "",
				DTARequestSettlementAddress: "",
				AccountAddress:              "",
			},
			expectError: false, // Service creation should succeed, addresses are just converted to zero address
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				service, err := NewService(tt.opts)

				if tt.expectError {
					require.Error(t, err)
					require.Nil(t, service)
				} else {
					require.NoError(t, err)
					require.NotNil(t, service)
					require.NotNil(t, service.logger)
				}
			},
		)
	}
}

func TestPrepareRequestSubscriptionOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	amount := big.NewInt(1000000)

	operation, err := service.PrepareRequestSubscriptionOperation(fundAdminAddr, fundTokenId, amount)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRequestRedemptionOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	shares := big.NewInt(500000)

	operation, err := service.PrepareRequestRedemptionOperation(fundAdminAddr, fundTokenId, shares)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRequestSubscriptionWithTokenApprovalOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	amount := big.NewInt(1000000)
	paymentTokenAddr := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")

	operation, err := service.PrepareRequestSubscriptionWithTokenApprovalOperation(
		fundAdminAddr, fundTokenId, amount, paymentTokenAddr,
	)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 2) // Should have approve + subscription transactions

	// Verify first transaction (approve)
	approveTx := operation.Transactions[0]
	require.Equal(t, paymentTokenAddr, approveTx.To)
	require.Equal(t, big.NewInt(0), approveTx.Value)
	require.NotEmpty(t, approveTx.Data)

	// Verify second transaction (subscription)
	subscriptionTx := operation.Transactions[1]
	require.Equal(t, service.dtaRequestManagementAddress, subscriptionTx.To)
	require.Equal(t, big.NewInt(0), subscriptionTx.Value)
	require.NotEmpty(t, subscriptionTx.Data)
}

func TestPrepareRegisterDistributorOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	distributorWalletAddr := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")

	operation, err := service.PrepareRegisterDistributorOperation(distributorWalletAddr)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRegisterFundTokenOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)
	fundTokenIdStr := "Test Token"
	fundTokenIdBytes := []byte(fundTokenIdStr)
	fundTokenIdHash := sha3.NewLegacyKeccak256().Sum(fundTokenIdBytes)
	fundTokenId := [32]byte{}
	copy(fundTokenId[:], fundTokenIdHash[:])
	tokenData := FundTokenData{
		FundTokenAddr:                 common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
		NavAddr:                       common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
		TokenChainSelector:            1234567890,
		DtaRequestSettlementAddr:      common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		TimezoneOffsetSecs:            big.NewInt(-18000), // -5 hours in seconds
		NavFeedDecimals:               18,
		PurchaseTokenRoundingDecimals: 18,
		PurchaseTokenDecimals:         18,
		FundRoundingDecimals:          18,
		FundTokenDecimals:             18,
		RequestsPerDay:                10,
		NavTTL:                        big.NewInt(0),
		PaymentInfo: DTAPaymentInfo{
			OffChainPaymentCurrency: 1, // USD
			PaymentTokenSourceAddr:  common.HexToAddress("0xA0b86a33E6241e2a4C8Ca3a3b4e4F1234567890"),
			PaymentTokenDestAddr:    common.HexToAddress("0xB0b86a33E6241e2a4C8Ca3a3b4e4F1234567890"),
		},
	}

	operation, err := service.PrepareRegisterFundTokenOperation(fundTokenId, tokenData)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareAllowDisallowDistributorForTokenOperations(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	distributorAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	// Test allow operation
	allowOp, err := service.PrepareAllowDistributorForTokenOperation(fundTokenId, distributorAddr)
	require.NoError(t, err)
	require.NotNil(t, allowOp)
	require.Len(t, allowOp.Transactions, 1)
	require.Equal(t, service.dtaRequestManagementAddress, allowOp.Transactions[0].To)

	// Test disallow operation
	disallowOp, err := service.PrepareDisallowDistributorForTokenOperation(fundTokenId, distributorAddr)
	require.NoError(t, err)
	require.NotNil(t, disallowOp)
	require.Len(t, disallowOp.Transactions, 1)
	require.Equal(t, service.dtaRequestManagementAddress, disallowOp.Transactions[0].To)
}

func TestPrepareEnableDisableFundTokenOperations(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}

	// Test enable operation
	enableOp, err := service.PrepareEnableFundTokenOperation(fundTokenId)
	require.NoError(t, err)
	require.NotNil(t, enableOp)
	require.Len(t, enableOp.Transactions, 1)
	require.Equal(t, service.dtaRequestManagementAddress, enableOp.Transactions[0].To)

	// Test disable operation
	disableOp, err := service.PrepareDisableFundTokenOperation(fundTokenId)
	require.NoError(t, err)
	require.NotNil(t, disableOp)
	require.Len(t, disableOp.Transactions, 1)
	require.Equal(t, service.dtaRequestManagementAddress, disableOp.Transactions[0].To)
}

func TestPrepareProcessDistributorRequestOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	requestId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}

	operation, err := service.PrepareProcessDistributorRequestOperation(requestId)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareCancelDistributorRequestOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	requestId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}

	operation, err := service.PrepareCancelDistributorRequestOperation(requestId)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRegisterFundAdminOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareRegisterFundAdminOperation(fundAdminAddr)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareVerifyDistributorWalletOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	distributorAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareVerifyDistributorWalletOperation(distributorAddr)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareForceAllowDistributorForTokenOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	distributorAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareForceAllowDistributorForTokenOperation(fundTokenId, distributorAddr)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestManagementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestToJson(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
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
		t.Run(
			tc.name, func(t *testing.T) {
				event := &apiClient.Event{
					VerifiableEvent: tc.input,
				}

				result, err := service.toJson(event)

				if tc.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.Equal(t, []byte(tc.expected), result)
				}
			},
		)
	}
}

// DTARequestSettlement Operation Tests

func TestPrepareAllowDTAOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	dtaAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	dtaChainSelector := uint64(1234567890)
	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	fundTokenAddr := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
	burnType := TokenBurnTypeBurn

	operation, err := service.PrepareAllowDTAOperation(dtaAddr, dtaChainSelector, fundTokenId, fundTokenAddr, burnType)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareDisallowDTAOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	dtaAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	dtaChainSelector := uint64(1234567890)
	fundTokenId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}

	operation, err := service.PrepareDisallowDTAOperation(dtaAddr, dtaChainSelector, fundTokenId)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareWithdrawTokensOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	token := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
	recipient := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
	amount := big.NewInt(1000000)

	operation, err := service.PrepareWithdrawTokensOperation(token, recipient, amount)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareTransferDTARequestSettlementOwnershipOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	newOwner := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")

	operation, err := service.PrepareTransferDTARequestSettlementOwnershipOperation(newOwner)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareRenounceDTARequestSettlementOwnershipOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	operation, err := service.PrepareRenounceDTARequestSettlementOwnershipOperation()
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestPrepareCompleteRequestProcessingOperation(t *testing.T) {
	service, err := NewService(
		&ServiceOptions{
			DTARequestManagementAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			DTARequestSettlementAddress: "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			AccountAddress:              "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
		},
	)
	require.NoError(t, err)

	requestId := [32]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
		31, 32,
	}
	success := true
	errorData := []byte("test error data")

	operation, err := service.PrepareCompleteRequestProcessingOperation(requestId, success, errorData)
	require.NoError(t, err)
	require.NotNil(t, operation)

	// Verify operation structure
	require.NotNil(t, operation.ID)
	require.Equal(t, service.accountAddress, operation.Account)
	require.Len(t, operation.Transactions, 1)

	// Verify transaction structure
	tx := operation.Transactions[0]
	require.Equal(t, service.dtaRequestSettlementAddress, tx.To)
	require.Equal(t, big.NewInt(0), tx.Value)
	require.NotEmpty(t, tx.Data)
}

func TestTokenBurnType(t *testing.T) {
	// Test TokenBurnType constants
	require.Equal(t, TokenBurnType(0), TokenBurnTypeNone)
	require.Equal(t, TokenBurnType(1), TokenBurnTypeBurn)
	require.Equal(t, TokenBurnType(2), TokenBurnTypeTransfer)
}
