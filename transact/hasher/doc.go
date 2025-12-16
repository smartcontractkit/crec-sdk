// Package hasher provides functionality for computing EIP-712 hashes and signatures
// for CREC operations without requiring network connectivity or API client dependencies.
//
// This package is useful when you need to:
//   - Generate operation hashes for verification
//   - Sign operations offline
//   - Implement custom signing workflows
//
// # Usage
//
// Create a hasher client without network dependencies:
//
//	hasher, err := hasher.NewClient(&hasher.Options{
//	    Logger: logger, // optional
//	})
//
//	// Hash an operation
//	hash, err := hasher.HashOperation(operation, chainSelector)
//
//	// Sign an operation
//	hash, signature, err := hasher.SignOperation(ctx, operation, signer, chainSelector)
//
// The hasher client can be used standalone or embedded in higher-level clients
// that need signing capabilities.
package hasher
