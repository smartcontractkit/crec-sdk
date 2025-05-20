package transact

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value"`
	Data  []byte         `json:"data"`
}

type Operation struct {
	ID           string         `json:"id"`
	Sender       common.Address `json:"sender"`
	Transactions []Transaction  `json:"transactions"`
	Signature    []byte         `json:"signature"`
}

func SendOperation(
	operation Operation,
) error {
	panic("not implemented")
}
