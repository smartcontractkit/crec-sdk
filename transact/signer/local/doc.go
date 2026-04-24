// Package local provides in-memory signers for local development and testing.
//
// # ECDSA Signing
//
// The ECDSA signer manages secp256k1 private keys in memory.
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
// Sign a 32-byte hash:
//
//	hash := crypto.Keccak256([]byte("message"))
//	signature, err := signer.Sign(ctx, hash)
//
// # RSA Signing
//
// The RSA signer manages RSA private keys in memory and produces PKCS#1 v1.5
// signatures for local development and testing.
//
// Generate or load an RSA key and create the signer:
//
//	privateKey, err := local.GenerateRSAKey(2048)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	signer, err := local.NewRSASigner(privateKey)
//
// Register the public key with a CREC Smart Wallet:
//
//	wallet, err := client.Wallets.Create(ctx, wallets.CreateInput{
//	    WalletType: "rsa",
//	    AllowedRsaSigners: &apiClient.RSASignersList{
//	        {E: signer.GetRSAPublicExponent(), N: signer.GetRSAModulus()},
//	    },
//	    ...
//	})
//
// # Use Cases
//
//   - Development and testing
//   - CI pipelines without external key infrastructure
//   - Sample code and quickstarts
//
// # Security Note
//
// Both signers keep private keys in memory. For production deployments
// requiring HSM-backed key storage, consider using the vault or kms signers.
package local
