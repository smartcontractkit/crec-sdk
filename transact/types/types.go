package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
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

// Operation defines model for Operation.
type OperationResponse struct {
	// Account Onchain account address performing the operation
	Account string `json:"account"`

	// AccountId Identifier of the account performing the operation
	AccountId uuid.UUID `json:"account_id"`

	// AccountOperationId Unique account operation identifier
	AccountOperationId string `json:"account_operation_id"`

	// ChainId The id that identifies the chain where the account performing the operation lives
	ChainId string `json:"chain_id"`

	// ConfirmedAt Timestamp of when the operation was confirmed
	ConfirmedAt *int64 `json:"confirmed_at,omitempty"`

	// CreatedAt Timestamp of when the operation was created
	CreatedAt int64 `json:"created_at"`

	// OperationId Unique identifier for the operation
	OperationId uuid.UUID `json:"operation_id"`

	// Signature EIP-712 signature of the operation
	Signature string `json:"signature"`

	// Status Current status of the operation
	Status string `json:"status"`

	// TransactionHash Onchain transaction hash which included the operation
	TransactionHash *string       `json:"transaction_hash,omitempty"`
	Transactions    []Transaction `json:"transactions"`
}

func UnmarshalOperationResponse(respBody []byte) (*OperationResponse, error) {
	var rawData map[string]interface{}
	if err := json.Unmarshal(respBody, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	// Process transactions to handle the value field
	if transactions, ok := rawData["transactions"].([]interface{}); ok {
		for _, txInterface := range transactions {
			if tx, ok := txInterface.(map[string]interface{}); ok {
				if valueStr, ok := tx["value"].(string); ok {
					// Convert string to number for big.Int
					if valueStr == "0" {
						tx["value"] = 0
					} else {
						if strings.HasPrefix(valueStr, "0x") {
							if val, err := strconv.ParseInt(valueStr, 0, 64); err == nil {
								tx["value"] = val
							}
						} else if val, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
							tx["value"] = val
						}
					}
				}
			}
		}
	}

	modifiedJSON, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modified data: %w", err)
	}

	var opResp OperationResponse
	if err := json.Unmarshal(modifiedJSON, &opResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to OperationResponse: %w", err)
	}

	return &opResp, nil
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
		ChainId:           big.NewInt(int64(chainId)),
		VerifyingContract: account,
	}
}
