package eip712_test

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/crec-sdk/transact/eip712"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

// This example demonstrates using the EIP-712 handler independently
// for offline signing without requiring any network dependencies.
func Example() {
	// Create an EIP-712 handler without any API dependencies
	handler, err := eip712.NewHandler(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create a test operation
	operation := &types.Operation{
		ID:       big.NewInt(12345),
		Account:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Deadline: big.NewInt(0),
		Transactions: []types.Transaction{
			{
				To:    common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				Value: big.NewInt(0),
				Data:  []byte{0x01, 0x02, 0x03},
			},
		},
	}

	// Base Sepolia chain selector
	chainSelector := "10344971235874465080"

	// Hash the operation
	hash, err := handler.HashOperation(operation, chainSelector)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Operation hash: %s\n", hash.Hex())

	// Create a signer (in production, use a secure key management solution)
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatal(err)
	}
	signer := local.NewSigner(privateKey)

	// Sign the operation
	opHash, signature, err := handler.SignOperation(context.Background(), operation, signer, chainSelector)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Signed operation: %s\n", opHash.Hex())
	fmt.Printf("Signature length: %d bytes\n", len(signature))
}
