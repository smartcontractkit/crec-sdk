package ccip

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/routerclient"
)

func TestPrepareCcipSendOperation_NoApprovals(t *testing.T) {
	s, err := NewService(&ServiceOptions{
		CcipRouterAddress:     "0x1111111111111111111111111111111111111111",
		AccountAddress:        "0x2222222222222222222222222222222222222222",
		IncludeTokenApprovals: false,
	})
	require.NoError(t, err)

	// Minimal EVM2AnyMessage
	msg := &routerclient.ClientEVM2AnyMessage{
		Receiver:     []byte{0x01, 0x02},
		Data:         []byte("hello"),
		TokenAmounts: []routerclient.ClientEVMTokenAmount{}, // no tokens => no approvals
		FeeToken:     common.Address{},
		ExtraArgs:    []byte{},
	}

	op, err := s.PrepareCcipSendOperation(1234, msg)
	require.NoError(t, err)
	require.NotNil(t, op)

	// Only one transaction expected: ccipSend to router
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x1111111111111111111111111111111111111111"), op.Transactions[0].To)
	require.NotNil(t, op.ID)
}

func TestPrepareCcipSendOperation_WithApprovals(t *testing.T) {
	s, err := NewService(&ServiceOptions{
		CcipRouterAddress:     "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		AccountAddress:        "0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		IncludeTokenApprovals: true,
	})
	require.NoError(t, err)

	ta := []routerclient.ClientEVMTokenAmount{
		{Token: common.HexToAddress("0xCcCcCcCcCcCcCcCcCcCcCcCcCcCcCcCcCcCcCcCc"), Amount: big.NewInt(10)},
		{Token: common.HexToAddress("0xDdDdDdDdDdDdDdDdDdDdDdDdDdDdDdDdDdDdDdDd"), Amount: big.NewInt(20)},
	}
	msg := &routerclient.ClientEVM2AnyMessage{
		Receiver:     []byte{0xaa},
		Data:         []byte("payload"),
		TokenAmounts: ta,
		FeeToken:     common.Address{},
		ExtraArgs:    []byte{},
	}

	op, err := s.PrepareCcipSendOperation(777, msg)
	require.NoError(t, err)
	require.NotNil(t, op)

	// Expect one approve per token + final ccipSend
	require.Len(t, op.Transactions, len(ta)+1)

	// First approvals go to each token address, final tx goes to router
	require.Equal(t, ta[0].Token, op.Transactions[0].To)
	require.Equal(t, ta[1].Token, op.Transactions[1].To)
	require.Equal(t, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), op.Transactions[2].To)
}
