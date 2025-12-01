// Package types provides data structures for CREC transact operations.
//
// This package defines the core types used for building and signing
// operations in the account abstraction transaction model.
//
// # Operation Structure
//
// An [Operation] represents a batch of transactions to be executed atomically:
//
//	operation := &types.Operation{
//	    ID:      big.NewInt(1),
//	    Account: common.HexToAddress("0x..."),
//	    Transactions: []types.Transaction{
//	        {
//	            To:    common.HexToAddress("0x..."),
//	            Value: big.NewInt(1000000000000000000), // 1 ETH
//	            Data:  []byte{},
//	        },
//	    },
//	}
//
// # EIP-712 Typed Data
//
// Operations are signed using EIP-712 typed data for secure, human-readable
// signatures. Generate typed data for signing:
//
//	typedData, err := operation.TypedData(chainID)
//	hash, _ := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
//
// # Transaction Type
//
// A [Transaction] represents a single call within an operation:
//
//	tx := types.Transaction{
//	    To:    common.HexToAddress("0x..."),  // Target contract
//	    Value: big.NewInt(0),                  // ETH value to send
//	    Data:  calldata,                       // Encoded function call
//	}
//
// # EIP-712 Domain
//
// The [EIP712Domain] provides domain separation for signatures:
//
//	domain := types.SignatureVerifyingAccountEIP712Domain(chainID, accountAddress)
//
// Domain parameters:
//   - Name: "SignatureVerifyingAccount"
//   - Version: "1"
//   - ChainId: Target blockchain chain ID
//   - VerifyingContract: The account contract address
package types
