package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestTypedData_NilDeadlineError(t *testing.T) {
	op := &Operation{
		ID:      big.NewInt(1),
		Account: common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Transactions: []Transaction{
			{
				To:    common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f"),
				Value: big.NewInt(0),
				Data:  []byte{},
			},
		},
	}
	td, err := op.TypedData("31337")
	require.Error(t, err)
	require.Nil(t, td)
	require.Contains(t, err.Error(), "deadline is required")
}

func TestTypedData_NoTransactionsError(t *testing.T) {
	op := &Operation{
		ID:           big.NewInt(1),
		Account:      common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Deadline:     big.NewInt(0),
		Transactions: nil, // no txs
	}
	td, err := op.TypedData("31337")
	require.Error(t, err)
	require.Nil(t, td)
}

func TestTypedData_DomainAndMessage(t *testing.T) {
	acc := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	op := &Operation{
		ID:       big.NewInt(42),
		Account:  acc,
		Deadline: big.NewInt(0),
		Transactions: []Transaction{
			{
				To:    common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f"),
				Value: big.NewInt(0),
				Data:  []byte{},
			},
		},
	}
	td, err := op.TypedData("31337")
	require.NoError(t, err)
	require.NotNil(t, td)

	// Primary type and presence of types
	require.Equal(t, "Operation", td.PrimaryType)
	require.Contains(t, td.Types, "Operation")
	require.Contains(t, td.Types, "Transaction")
	require.Contains(t, td.Types, "EIP712Domain")

	// Domain checks
	require.Equal(t, EIP712DomainName, td.Domain.Name)
	require.Equal(t, EIP712DomainVersion, td.Domain.Version)
	require.Equal(t, acc.Hex(), td.Domain.VerifyingContract)
}

func TestSmartAccountEIP712Domain(t *testing.T) {
	acc := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	d := SmartAccountEIP712Domain(1337, acc)
	require.Equal(t, EIP712DomainName, d.Name)
	require.Equal(t, EIP712DomainVersion, d.Version)
	require.Equal(t, acc, d.VerifyingContract)
	require.Equal(t, int64(1337), d.ChainId)
}
