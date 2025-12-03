// Package vault provides a signer using HashiCorp Vault Transit secrets engine.
//
// The vault signer integrates with HashiCorp Vault's Transit secrets engine,
// providing enterprise-grade key management with HSM support.
//
// # Usage
//
// Create a signer connected to Vault:
//
//	signer, err := vault.NewSigner(
//	    "https://vault.example.com:8200", // Vault URL
//	    "your-vault-token",               // Vault token
//	    "transit",                         // Mount path
//	    "my-signing-key",                  // Key name
//	)
//
// Sign a hash:
//
//	hash := sha256.Sum256([]byte("message"))
//	signature, err := signer.Sign(ctx, hash[:])
//
// # Key Types
//
// The signer supports multiple key types:
//
//	vault.KeyTypeRSA2048   // RSA 2048-bit key
//	vault.KeyTypeRSA4096   // RSA 4096-bit key
//	vault.KeyTypeECDSAP256 // ECDSA P-256 curve
//	vault.KeyTypeECDSAP384 // ECDSA P-384 curve
//	vault.KeyTypeECDSAP521 // ECDSA P-521 curve
//
// # Key Creation
//
// Create keys directly in Vault:
//
//	result, err := signer.CreateKey("my-new-key", vault.KeyTypeRSA2048)
//	fmt.Println(result.KeyName, result.Modulus)
//
// Or use the convenience function:
//
//	result, err := vault.CreateKeyInVault(vaultURL, token, "transit", "key-name", vault.KeyTypeRSA4096)
//
// # Public Key Retrieval
//
// Get the public key for verification:
//
//	pubKey, err := signer.Public()
//	rsaPubKey := pubKey.(*rsa.PublicKey)
//
// For RSA keys, extract the modulus:
//
//	modulus, err := signer.GetRSAModulus() // hex-encoded
//
// # Features
//
//   - Secure key storage in Vault
//   - HSM-backed keys support
//   - Key rotation capabilities
//   - Audit logging
//   - Multi-tenant key isolation
//   - RSA and ECDSA key support
//
// # Use Cases
//
//   - Production deployments
//   - Enterprise environments
//   - Compliance requirements (SOC 2, FIPS 140-2)
//   - Multi-service key sharing
//   - Key rotation requirements
//
// # Vault Configuration
//
// Create keys using Vault CLI:
//
//	vault write transit/keys/my-key type=rsa-2048
//	vault write transit/keys/my-key type=ecdsa-p256
package vault
