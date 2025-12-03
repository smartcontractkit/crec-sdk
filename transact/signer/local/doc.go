// Package local provides a signer using local ECDSA private keys.
//
// The local signer manages private keys in memory and is suitable for
// development and testing scenarios where you control the private key directly.
//
// # Usage
//
// Generate or load a private key and create the signer:
//
//	privateKey, err := crypto.GenerateKey()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	signer := local.NewSigner(privateKey)
//
// Load from hex string:
//
//	privateKey, err := crypto.HexToECDSA("your-private-key-hex")
//	signer := local.NewSigner(privateKey)
//
// # Signing
//
// Sign a 32-byte hash:
//
//	hash := crypto.Keccak256([]byte("message"))
//	signature, err := signer.Sign(ctx, hash)
//
// # Features
//
//   - Fast signing operations
//   - Ethereum-compatible signature format (65 bytes with recovery ID)
//   - Uses secp256k1 curve
//
// # Use Cases
//
//   - Development and testing
//   - Single-node deployments
//   - Scenarios where key management is handled externally
//
// # Security Note
//
// This signer keeps the private key in memory. For production deployments
// requiring HSM-backed key storage, consider using the vault or kms signers.
package local
