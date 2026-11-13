// Package eip712 provides functionality for computing EIP-712 hashes and signatures
// for CREC operations without requiring network connectivity or API client dependencies.
//
// This package is useful when you need to:
//   - Generate operation hashes for verification
//   - Sign operations offline
//   - Implement custom signing workflows
//   - Sign owner authorizations for wallet upgrades and administrative operations
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
//
// # Administrative Authorizations
//
// Wallet upgrades and administrative operations are authorized by the wallet
// owner's EIP-712 signature over a capability's pinned struct, bound to the
// wallet's own CLLSmartAccount domain. [Handler.SignAdminAuthorization] builds
// and signs that struct from its pinned type string — the catalog's
// eip712_type for administrative capabilities, [UpgradeAuthorizationType] for
// upgrades:
//
//	auth := &eip712.AdminAuthorization{
//	    EIP712Type: eip712.UpgradeAuthorizationType,
//	    Arguments:  map[string]any{"newImplementation": implementationAddress},
//	    OpID:       authOpID,   // the authorization's own wallet_operation_id
//	    Deadline:   deadline,
//	}
//	hash, signature, err := handler.SignAdminAuthorization(ctx, auth, ownerSigner, walletAddress, chainSelector)
//
// Arguments carry the capability's values canonically encoded (integers as
// decimal strings, addresses and byte strings as lowercase 0x hex, booleans
// and strings as JSON values). The struct's trailing opId and deadline fields
// are supplied by OpID and Deadline, and must not appear in Arguments.
package eip712
