package types

import (
	"errors"
	"math/big"

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

func (op *Operation) TypedData(chainId uint64) (*apitypes.TypedData, error) {
	domain := SignatureVerifyingAccountEIP712Domain(chainId, op.Account)
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

type EIP712Domain struct {
	Name              string         `json:"name"`
	Version           string         `json:"version"`
	ChainId           *big.Int       `json:"chainId"`
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
		ChainId:           math.NewHexOrDecimal256(d.ChainId.Int64()),
		VerifyingContract: d.VerifyingContract.String(),
	}
	return domain
}

func SignatureVerifyingAccountEIP712Domain(chainId uint64, account common.Address) *EIP712Domain {
	return &EIP712Domain{
		Name:              EIP712DomainName,
		Version:           EIP712DomainVersion,
		ChainId:           new(big.Int).SetUint64(chainId),
		VerifyingContract: account,
	}
}
