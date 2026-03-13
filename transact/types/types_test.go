package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestTypedData_NoTransactionsError(t *testing.T) {
	op := &Operation{
		ID:           big.NewInt(1),
		Account:      common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Transactions: nil, // no txs
	}
	td, err := op.TypedData("31337")
	require.Error(t, err)
	require.Nil(t, td)
}

func TestTypedData_DomainAndMessage(t *testing.T) {
	acc := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	op := &Operation{
		ID:      big.NewInt(42),
		Account: acc,
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

func TestSignatureVerifyingAccountEIP712Domain(t *testing.T) {
	acc := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	d := SignatureVerifyingAccountEIP712Domain(1337, acc)
	require.Equal(t, EIP712DomainName, d.Name)
	require.Equal(t, EIP712DomainVersion, d.Version)
	require.Equal(t, acc, d.VerifyingContract)
	require.Equal(t, int64(1337), d.ChainId)
}

func TestNewEIP712Domain(t *testing.T) {
	acc := common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	d := NewEIP712Domain("CustomDomain", 42161, acc)
	require.Equal(t, "CustomDomain", d.Name)
	require.Equal(t, EIP712DomainVersion, d.Version)
	require.Equal(t, acc, d.VerifyingContract)
	require.Equal(t, int64(42161), d.ChainId)
}

func TestTypedDataWithCustomDomainName(t *testing.T) {
	acc := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	op := &Operation{
		ID:      big.NewInt(42),
		Account: acc,
		Transactions: []Transaction{
			{
				To:    common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f"),
				Value: big.NewInt(0),
				Data:  []byte{},
			},
		},
	}

	t.Run("custom domain name is used", func(t *testing.T) {
		td, err := op.TypedData("31337", "CustomDomain")
		require.NoError(t, err)
		require.NotNil(t, td)
		require.Equal(t, "CustomDomain", td.Domain.Name)
	})

	t.Run("omitted domain name uses default", func(t *testing.T) {
		td, err := op.TypedData("31337")
		require.NoError(t, err)
		require.NotNil(t, td)
		require.Equal(t, EIP712DomainName, td.Domain.Name)
	})

	t.Run("empty string domain name uses default", func(t *testing.T) {
		td, err := op.TypedData("31337", "")
		require.NoError(t, err)
		require.NotNil(t, td)
		require.Equal(t, EIP712DomainName, td.Domain.Name)
	})
}
