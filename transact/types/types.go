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
	// EIP712DomainName is the EIP-712 domain name for all smart account contracts.
	EIP712DomainName = `CLLSmartAccount`
	// EIP712DomainVersion is the EIP-712 domain version for all smart account contracts.
	EIP712DomainVersion = `1`
)

// Transaction represents a single transaction within an operation for EIP-712 signing.
type Transaction struct {
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value,string"`
	Data  hexutil.Bytes  `json:"data"`
}

// EIP712Type returns the EIP-712 type name for Transaction.
func (tx *Transaction) EIP712Type() string {
	return "Transaction"
}

// EIP712Types returns the EIP-712 type definitions for Transaction fields.
func (tx *Transaction) EIP712Types() []apitypes.Type {
	return []apitypes.Type{
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "data", Type: "bytes"},
	}
}

// EIP712Message returns the EIP-712 message representation of the transaction.
func (tx *Transaction) EIP712Message() apitypes.TypedDataMessage {
	return apitypes.TypedDataMessage{
		"to":    tx.To.String(),
		"value": tx.Value.String(),
		"data":  tx.Data,
	}
}

// Operation represents a batch of transactions to be executed atomically by a smart account.
type Operation struct {
	ID           *big.Int       `json:"id"`
	Account      common.Address `json:"account"`
	Deadline     *big.Int       `json:"deadline"`
	Transactions []Transaction  `json:"transactions"`
}

// EIP712Type returns the EIP-712 type name for Operation.
func (op *Operation) EIP712Type() string {
	return "Operation"
}

// EIP712Types returns the EIP-712 type definitions for Operation fields.
func (op *Operation) EIP712Types() []apitypes.Type {
	return []apitypes.Type{
		{Name: "id", Type: "uint256"},
		{Name: "account", Type: "address"},
		{Name: "deadline", Type: "uint256"},
		{Name: "transactions", Type: "Transaction[]"},
	}
}

// EIP712Message returns the EIP-712 message representation of the operation.
func (op *Operation) EIP712Message() apitypes.TypedDataMessage {
	var txns []apitypes.TypedDataMessage
	for _, tx := range op.Transactions {
		txns = append(txns, tx.EIP712Message())
	}
	return apitypes.TypedDataMessage{
		"id":           op.ID.String(),
		"account":      op.Account.Hex(),
		"deadline":     op.Deadline.String(),
		"transactions": txns,
	}
}

// TypedData creates the EIP-712 typed data for the operation to be hashed and signed
// using the default CLLSmartAccount domain.
// Returns an error if chainId is invalid or the operation has no transactions.
// ChainId is parsed as int64 because go-ethereum's apitypes.TypedDataDomain uses
// math.HexOrDecimal256 for ChainID, whose constructor accepts only int64.
func (op *Operation) TypedData(chainId string) (*apitypes.TypedData, error) {
	if op.Deadline == nil {
		return nil, errors.New("deadline is required")
	}

	chainIdInt, err := strconv.ParseInt(chainId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	domain := SmartAccountEIP712Domain(chainIdInt, op.Account)
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

// EIP712Domain represents the EIP-712 domain for the smart account contract.
// It contains chainID and version metadata of the smart account contract,
// which are used for generating the EIP-712 typed data hash and for signing.
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

// Type returns the EIP-712 type name for the domain.
func (d *EIP712Domain) Type() string {
	return "EIP712Domain"
}

// Types returns the EIP-712 type definitions for the domain fields.
func (d *EIP712Domain) Types() []apitypes.Type {
	types := []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}
	return types
}

// TypedData returns the go-ethereum TypedDataDomain representation.
func (d *EIP712Domain) TypedData() apitypes.TypedDataDomain {
	domain := apitypes.TypedDataDomain{
		Name:              d.Name,
		Version:           d.Version,
		ChainId:           math.NewHexOrDecimal256(d.ChainId),
		VerifyingContract: d.VerifyingContract.String(),
	}
	return domain
}

// SmartAccountEIP712Domain creates the EIP-712 domain for the smart account contract.
func SmartAccountEIP712Domain(chainId int64, account common.Address) *EIP712Domain {
	return &EIP712Domain{
		Name:              EIP712DomainName,
		Version:           EIP712DomainVersion,
		ChainId:           chainId,
		VerifyingContract: account,
	}
}
