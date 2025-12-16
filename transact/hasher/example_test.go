package hasher_test

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/crec-sdk/transact/hasher"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

// This example demonstrates using the hasher client independently
// for offline signing without requiring any network dependencies.
func Example() {
	// Create a hasher client without any API dependencies
	hasherClient, err := hasher.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create a test operation
	operation := &types.Operation{
		ID:      big.NewInt(12345),
		Account: common.HexToAddress("0x1234567890123456789012345678901234567890"),
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
	hash, err := hasherClient.HashOperation(operation, chainSelector)
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
	opHash, signature, err := hasherClient.SignOperation(context.Background(), operation, signer, chainSelector)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Signed operation: %s\n", opHash.Hex())
	fmt.Printf("Signature length: %d bytes\n", len(signature))
}
