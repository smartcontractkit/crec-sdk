// Package eip712 provides functionality for computing EIP-712 hashes and signatures
// for CREC operations without requiring network connectivity or API client dependencies.
//
// This package is useful when you need to:
//   - Generate operation hashes for verification
//   - Sign operations offline
//   - Implement custom signing workflows
//
// # Usage
//
// Create an EIP-712 handler without network dependencies:
//
//	handler, err := eip712.NewHandler(&eip712.Options{
//	    Logger: logger, // optional
//	})
//
//	// Hash an operation
//	hash, err := handler.HashOperation(operation, chainSelector)
//
//	// Sign an operation
//	hash, signature, err := handler.SignOperation(ctx, operation, signer, chainSelector)
//
// The handler can be used standalone or embedded in higher-level clients
// that need signing capabilities.
package eip712
