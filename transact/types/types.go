package types

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	EIP712DomainName    = `SignatureVerifyingAccount`
	EIP712DomainVersion = `1`
)

type Transaction struct {
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value,string"`
	Data  hexutil.Bytes  `json:"data"`
}

func (tx *Transaction) EIP712Type() string {
	return "Transaction"
}

func (tx *Transaction) EIP712Types() []apitypes.Type {
	return []apitypes.Type{
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "data", Type: "bytes"},
	}
}

func (tx *Transaction) EIP712Message() apitypes.TypedDataMessage {
	return apitypes.TypedDataMessage{
		"to":    tx.To.String(),
		"value": tx.Value.String(),
		"data":  tx.Data,
	}
}

type Operation struct {
	ID           *big.Int       `json:"id"`
	Account      common.Address `json:"account"`
	Transactions []Transaction  `json:"transactions"`
}

func (op *Operation) EIP712Type() string {
	return "Operation"
}

func (op *Operation) EIP712Types() []apitypes.Type {
	return []apitypes.Type{
		{Name: "id", Type: "uint256"},
		{Name: "account", Type: "address"},
		{Name: "transactions", Type: "Transaction[]"},
	}
}

func (op *Operation) EIP712Message() apitypes.TypedDataMessage {
	var txns []apitypes.TypedDataMessage
	for _, tx := range op.Transactions {
		txns = append(txns, tx.EIP712Message())
	}
	return apitypes.TypedDataMessage{
		"id":           op.ID.String(),
		"account":      op.Account.Hex(),
		"transactions": txns,
	}
}

// Creates the EIP-712 typed data for the operation to be hashed and
// signed.
//
// ---------------------
// ChainID Constraint
// ---------------------
//
// ChainId is parsed as int64 because go-ethereum's apitypes.TypedDataDomain
// uses math.HexOrDecimal256 type for ChainID, whose constructor accepts
// only int64.
func (op *Operation) TypedData(chainId string) (*apitypes.TypedData, error) {
	chainIdInt, err := strconv.ParseInt(chainId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	domain := SignatureVerifyingAccountEIP712Domain(chainIdInt, op.Account)
	message := op.EIP712Message()

	if len(op.Transactions) == 0 {
		return nil, errors.New("no transactions")
	}

	return &apitypes.TypedData{
		Types: apitypes.Types{
			op.EIP712Type():                 op.EIP712Types(),
			op.Transactions[0].EIP712Type(): op.Transactions[0].EIP712Types(), // get 712 type from first transaction
			domain.Type():                   domain.Types(),
		},
		PrimaryType: op.EIP712Type(),
		Domain:      domain.TypedData(),
		Message:     message,
	}, nil
}

// EIP712Domain represents the EIP-712 domain for the Signature Verifying Account.
// It contains chainID and version metadata of the Signature Verifying Account
// contract, which are used for generating the EIP-712 typed data hash and for signing.
//
// ---------------------
// ChainID Constraint
// ---------------------
//
// ChainId is int64 because go-ethereum's apitypes.TypedDataDomain uses uses
// math.HexOrDecimal256 type for ChainID, whose constructor accepts only int64.
type EIP712Domain struct {
	Name              string         `json:"name"`
	Version           string         `json:"version"`
	ChainId           int64          `json:"chainId"`
	VerifyingContract common.Address `json:"verifyingContract"`
}

func (d *EIP712Domain) Type() string {
	return "EIP712Domain"
}

func (d *EIP712Domain) Types() []apitypes.Type {
	types := []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}
	return types
}

func (d *EIP712Domain) TypedData() apitypes.TypedDataDomain {
	domain := apitypes.TypedDataDomain{
		Name:              d.Name,
		Version:           d.Version,
		ChainId:           math.NewHexOrDecimal256(d.ChainId),
		VerifyingContract: d.VerifyingContract.String(),
	}
	return domain
}

func SignatureVerifyingAccountEIP712Domain(chainId int64, account common.Address) *EIP712Domain {
	return &EIP712Domain{
		Name:              EIP712DomainName,
		Version:           EIP712DomainVersion,
		ChainId:           chainId,
		VerifyingContract: account,
	}
}
