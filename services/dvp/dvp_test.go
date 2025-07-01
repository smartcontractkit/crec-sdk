package dvp

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/contract"
)

func TestHashSettlement(t *testing.T) {

	dvpService, err := NewDvpService(
		&DvpServiceOptions{
			DvpCoordinatorAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
			AccountAddress:        "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
		},
	)
	if err != nil {
		t.Fatalf("failed to create DvP service: %v", err)
	}

	settlement := &contract.Settlement{
		SettlementId: big.NewInt(1751404299),
		PartyInfo: contract.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
			BuyerDestinationAddress:  common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
			SellerSourceAddress:      common.HexToAddress("0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc"),
			SellerDestinationAddress: common.HexToAddress("0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: contract.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
			AssetTokenDestinationAddress:   common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
			PaymentCurrency:                147,
			PaymentTokenAmount:             big.NewInt(1000000),
			AssetTokenAmount:               big.NewInt(1000000000000000000),
			PaymentTokenType:               0,
			AssetTokenType:                 1,
		},
		DeliveryInfo: contract.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1234567890),
			PaymentDestinationChainSelector: uint64(1234567890),
			AssetSourceChainSelector:        uint64(1234567890),
			AssetDestinationChainSelector:   uint64(1234567890),
		},
		SecretHash:           common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(1751490699),
		CcipCallbackGasLimit: 0,
		Data:                 []byte(""),
	}

	hash, err := dvpService.HashSettlement(settlement)
	if err != nil {
		t.Fatalf("failed to hash settlement: %v", err)
	}

	require.Equal(t, common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec"), hash)
}
