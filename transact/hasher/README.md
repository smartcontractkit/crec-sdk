# Hasher Package

The `hasher` package provides standalone EIP-712 hashing and signing functionality for CREC operations without requiring network connectivity or API client dependencies.

## Purpose

This package was created to decouple cryptographic operations (hashing and signing) from network operations (sending to the CREC API). This allows applications to:

- Generate operation hashes for verification without initializing a full client
- Sign operations offline in secure environments
- Implement custom signing workflows independent of the CREC API

## Architecture

The hasher client is embedded within the main `transact.Client` but can also be used standalone. When you call `HashOperation` or `SignOperation` on a `transact.Client`, it delegates to the embedded hasher.

```
┌─────────────────────┐
│  Transact Client    │
│  ┌───────────────┐  │
│  │ Hasher Client │  │  ← Can be used standalone
│  └───────────────┘  │
│  + API operations   │
└─────────────────────┘
```

## Usage

### Standalone Usage (No Network Dependencies)

```go
import (
    "context"
    "github.com/smartcontractkit/crec-sdk/transact/hasher"
    "github.com/smartcontractkit/crec-sdk/transact/signer/local"
)

// Create hasher client without any API dependencies
hasher, err := hasher.NewClient(&hasher.Options{
    Logger: logger, // optional
})

// Hash an operation
hash, err := hasher.HashOperation(operation, chainSelector)

// Sign an operation
hash, signature, err := hasher.SignOperation(ctx, operation, signer, chainSelector)

// Sign a pre-computed hash
signature, err := hasher.SignOperationHash(ctx, hash, signer)
```

### Via Transact Client

```go
import "github.com/smartcontractkit/crec-sdk/transact"

transactClient, err := transact.NewClient(&transact.Options{
    CRECClient: apiClient,
})

// Access the embedded hasher directly
hash, err := transactClient.Hasher.HashOperation(operation, chainSelector)

// Or use the delegating methods
hash, signature, err := transactClient.SignOperation(ctx, operation, signer, chainSelector)
```

## API

### Types

- `Client` - The main hasher client
- `Options` - Configuration options (optional logger)

### Methods

- `HashOperation(op *types.Operation, chainSelector string) (common.Hash, error)` - Computes EIP-712 hash
- `SignOperation(ctx context.Context, op *types.Operation, signer signer.Signer, chainSelector string) (common.Hash, []byte, error)` - Hashes and signs an operation
- `SignOperationHash(ctx context.Context, opHash common.Hash, signer signer.Signer) ([]byte, error)` - Signs a pre-computed hash

### Utility Functions

- `GetChainIDFromSelector(chainSelector string) (*big.Int, error)` - Extracts chain ID from a chain selector

## Examples

See `example_test.go` for a complete working example.

## Error Handling

The hasher package defines its own error types:

- `ErrOperationRequired` - Operation parameter is nil
- `ErrSignerRequired` - Signer parameter is nil
- `ErrParseChainSelector` - Invalid chain selector format
- `ErrGetChainFamily` - Failed to determine chain family
- `ErrUnsupportedChainFamily` - Chain family is not EVM
- `ErrGetChainID` - Failed to get chain ID
- `ErrCreateTypedData` - Failed to create EIP-712 typed data
- `ErrComputeOperationHash` - Failed to compute hash
- `ErrHashOperation` - Failed to hash operation
- `ErrSignOperation` - Failed to sign operation

## Testing

Run tests with:

```bash
go test ./transact/hasher/...
```
