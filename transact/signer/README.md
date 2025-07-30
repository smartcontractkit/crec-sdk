# CVN SDK Signers

This package provides signing interfaces for the CVN SDK, allowing you to sign operations using different key management strategies. Currently, two signers are available: `LocalSigner` for local private key management and `TransitSigner` for HashiCorp Vault Transit secrets engine integration.

## Table of Contents

- [Available Signers](#available-signers)
  - [LocalSigner](#localsigner)
  - [TransitSigner](#transitsigner)
- [Signer Interface](#signer-interface)
- [Usage Examples](#usage-examples)
- [Key Type Options](#key-type-options)
- [Testing](#testing)
- [Production Considerations](#production-considerations)

## Available Signers

### LocalSigner

The `LocalSigner` manages private keys locally in memory and is suitable for development and testing scenarios where you control the private key directly.

**Features:**
- Local ECDSA private key management
- Fast signing operations
- Ethereum-compatible signature format
- Simple setup for development

**Use Cases:**
- Development and testing
- Single-node deployments
- Scenarios where key management is handled externally

### TransitSigner

The `TransitSigner` integrates with HashiCorp Vault's Transit secrets engine, providing enterprise-grade key management with hardware security module (HSM) support.

**Features:**
- Secure key storage in Vault
- Support for HSM-backed keys
- Key rotation capabilities
- Audit logging
- Multi-tenant key isolation
- Support for RSA and ECDSA keys

**Use Cases:**
- Production deployments
- Enterprise environments
- Compliance requirements (SOC 2, FIPS 140-2)
- Multi-service key sharing
- Key rotation requirements

## Signer Interface

All signers implement the simple `Signer` interface:

```go
type Signer interface {
    Sign(hash []byte) ([]byte, error)
}
```

This allows you to swap between different signing implementations without changing your application code.

**Note**: The `TransitSigner` also provides a `Public()` method to retrieve the public key from Vault:

```go
// Additional method available on TransitSigner
func (s *TransitSigner) Public() (interface{}, error)
```

This method returns either an `*rsa.PublicKey` or `*ecdsa.PublicKey` depending on the key type configured in Vault.

## Usage Examples

### LocalSigner Example

```go
package main

import (
    "crypto/ecdsa"
    "fmt"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/smartcontractkit/cvn-sdk/transact/signer"
)

func main() {
    // Generate or load a private key
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        panic(err)
    }
    
    // Create the local signer
    localSigner := signer.NewLocalSigner(privateKey)
    
    // Sign some data
    hash := crypto.Keccak256([]byte("hello world"))
    signature, err := localSigner.Sign(hash)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Signature: %x\n", signature)
}
```

### TransitSigner Example

```go
package main

import (
    "fmt"
    "crypto/sha256"
    "github.com/smartcontractkit/cvn-sdk/transact/signer"
)

func main() {
    // Create the transit signer
    transitSigner, err := signer.NewTransitSigner(
        "https://vault.example.com:8200", // Vault URL
        "your-vault-token",               // Vault token
        "transit",                        // Mount path
        "my-signing-key",                 // Key name
    )
    if err != nil {
        panic(err)
    }
    
    // Sign some data
    data := []byte("hello world")
    hash := sha256.Sum256(data)
    signature, err := transitSigner.Sign(hash[:])
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Signature: %x\n", signature)
    
    // Optionally get the public key for verification
    pubKey, err := transitSigner.Public()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Public key retrieved: %T\n", pubKey)

    rsaPubKey, ok := pubKey.(*rsa.PublicKey)
    if !ok {
        panic(fmt.Errorf("Public key should be an RSA key"))
    }

	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, hash, signature, nil)
    if err != nil {
        panic(err)
    }
}
```

### Integration with CVN Transact Client

```go
package main

import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/smartcontractkit/cvn-sdk/client"
    "github.com/smartcontractkit/cvn-sdk/transact"
    "github.com/smartcontractkit/cvn-sdk/transact/signer"
    "github.com/smartcontractkit/cvn-sdk/transact/types"
)

func main() {
    // Create CVN client
    cvnClient, err := client.NewCVNClient("https://api.example.com", "api-key")
    if err != nil {
        panic(err)
    }
    
    // Create transact client
    transactClient, err := transact.NewClient(cvnClient, &transact.ClientOptions{
        ChainId: "1", // Ethereum mainnet
    })
    if err != nil {
        panic(err)
    }
    
    // Choose your signer (Local or Transit)
    var s signer.Signer
    
    // Option 1: Local signer
    privateKey, _ := crypto.GenerateKey()
    s = signer.NewLocalSigner(privateKey)
    
    // Option 2: Transit signer
    // s, err = signer.NewTransitSigner("vault-url", "token", "transit", "key-name")
    
    // Create an operation
    operation := &types.Operation{
        ID:      big.NewInt(1),
        Account: &common.Address{},
        Transactions: []*types.Transaction{
            {
                To:    &common.Address{},
                Value: big.NewInt(1000000000000000000), // 1 ETH
                Data:  []byte{},
            },
        },
    }
    
    // Sign the operation
    signature, err := transactClient.SignOperation(operation, s)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Operation signed: %x\n", signature)
}
```

## Key Type Options

### RSA Keys

RSA keys are well-established and widely supported but produce larger signatures.

**Vault Transit Configuration:**
```bash
# Create RSA-2048 key
vault write transit/keys/my-rsa-key type=rsa-2048

# Create RSA-4096 key (more secure, larger signatures)
vault write transit/keys/my-rsa-key type=rsa-4096
```

**Characteristics:**
- **RSA-2048**: 256-byte signatures
- **RSA-4096**: 512-byte signatures
- Deterministic signatures
- Well-established security properties

### ECDSA Keys

ECDSA keys produce smaller signatures and are commonly used in blockchain applications.

**Vault Transit Configuration:**
```bash
# Create ECDSA P-256 key
vault write transit/keys/my-ecdsa-key type=ecdsa-p256

# Create ECDSA P-384 key
vault write transit/keys/my-ecdsa-key type=ecdsa-p384
```

**Characteristics:**
- **P-256**: ~71-byte signatures
- **P-384**: ~103-byte signatures
- Non-deterministic signatures (includes randomness)

### LocalSigner Key Support

The `LocalSigner` currently supports ECDSA keys compatible with Ethereum:

```go
// Generate secp256k1 key (Ethereum standard)
privateKey, err := crypto.GenerateKey()

// Load from hex string
privateKey, err := crypto.HexToECDSA("your-private-key-hex")

// Create signer
localSigner := signer.NewLocalSigner(privateKey)
```

## Testing

### Test Setup

The test suite uses **testcontainers** to spin up real HashiCorp Vault instances, ensuring tests run against actual Vault behavior rather than mocks.

**Prerequisites:**
- Docker installed and running
- Go 1.21+ 

**Test Dependencies:**
```bash
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/testcontainers/testcontainers-go/modules/vault@latest
```

### Running Tests

**Run All Signer Tests:**
```bash
./test-vault-transit.sh
```

**Run Individual Test Suites:**
```bash
# Standalone Transit signer tests
go test ./transact/signer -v -run TestTransitSigner -timeout=60s

# Integration tests with transact client
go test ./transact -v -run TestSignOperationWithVaultTransit -timeout=60s

# Local signer tests (if any)
go test ./transact/signer -v -run TestLocalSigner
```

### Test Coverage

The test suite includes:

1. **RSA-2048 Signing Test** (`TestTransitSigner_Sign_Integration`)
   - Creates RSA-2048 key in Vault
   - Tests signing operations
   - Validates 256-byte signature length
   - Tests multiple signing operations

2. **ECDSA Signing Test** (`TestTransitSigner_Sign_WithECDSA`)
   - Creates ECDSA P-256 key in Vault
   - Tests signing operations
   - Validates ~71-byte signature length

3. **Error Handling Tests**
   - `TestTransitSigner_InvalidKey`: Tests behavior with non-existent keys
   - `TestTransitSigner_InvalidToken`: Tests authentication failures

4. **Public Key Retrieval Test** (`TestTransitSigner_Public`)
   - Tests public key retrieval for both RSA and ECDSA keys
   - Validates correct key types and properties
   - Verifies key size and curve parameters

5. **Integration Test** (`TestSignOperationWithVaultTransit`)
   - Full end-to-end workflow
   - CVN client + Transact client + Vault signer
   - Real operation signing and cryptographic verification

### Test Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Test Suite    │───▶│  TestContainers  │───▶│  Vault Docker   │
│                 │    │                  │    │   Container     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                                              │
         │              ┌─────────────────┐             │
         └─────────────▶│  Transit API    │◀────────────┘
                        │                 │
                        └─────────────────┘
```

**Test Container Configuration:**
- **Image**: `hashicorp/vault:1.13.3`
- **Root Token**: `myroot`
- **Transit Mount**: `transit/`
- **Auto-cleanup**: Containers terminated after tests

**Test Data:**
- Consistent test data across runs for reproducible results
- SHA-256 hashing of "hello world" for signature tests
- Ethereum-compatible operation structures
- Cryptographic verification of all signatures using public keys retrieved from Vault

## Production Considerations

### Security Best Practices

**Vault Configuration:**
- Use TLS/HTTPS for all Vault communication
- Implement proper authentication (not root tokens)
- Enable audit logging
- Use least-privilege policies
- Consider HSM integration for highest security

**Key Management:**
- Implement key rotation policies
- Use separate keys per environment
- Monitor key usage through Vault audit logs
- Implement key backup and disaster recovery

**Application Security:**
- Store Vault tokens securely (Kubernetes secrets, etc.)
- Implement token renewal logic
- Use short-lived tokens when possible
- Validate all signatures before processing

### Example Production Configuration

```go
// Production-ready Transit signer configuration
func NewProductionTransitSigner() (*signer.TransitSigner, error) {
    vaultConfig := vault.DefaultConfig()
    vaultConfig.Address = os.Getenv("VAULT_ADDR")
    vaultConfig.MaxRetries = 3
    vaultConfig.Timeout = 10 * time.Second
    
    // Configure TLS
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
    }
    vaultConfig.ConfigureTLS(&vault.TLSConfig{
        TLSConfig: tlsConfig,
    })
    
    return signer.NewTransitSigner(
        vaultConfig.Address,
        os.Getenv("VAULT_TOKEN"),
        "transit",
        os.Getenv("SIGNING_KEY_NAME"),
    )
}
```

This comprehensive setup ensures your signing operations are secure, performant, and observable in production environments.
