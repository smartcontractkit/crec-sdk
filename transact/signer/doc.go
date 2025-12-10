// Package signer provides signing interfaces for the CREC SDK.
//
// This package defines the [Signer] interface and provides multiple
// implementations for different key management strategies:
//
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/local] - Local private key
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/vault] - HashiCorp Vault Transit
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/kms] - AWS KMS
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/privy] - Privy wallet-as-a-service
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/fireblocks] - Fireblocks custody
//
// # Signer Interface
//
// All signers implement the [Signer] interface:
//
//	type Signer interface {
//	    Sign(ctx context.Context, hash []byte) ([]byte, error)
//	}
//
// This allows swapping between signing implementations without changing
// application code.
//
// # TypedDataSigner Interface
//
// Some signers also implement [TypedDataSigner] for EIP-712 typed data signing:
//
//	type TypedDataSigner interface {
//	    SignTypedData(ctx context.Context, typedData *TypedData) ([]byte, error)
//	}
//
// This is useful for custody providers that need to see the full typed data
// structure for policy enforcement. Currently implemented by:
//   - [github.com/smartcontractkit/crec-sdk/transact/signer/fireblocks]
//
// # Choosing a Signer
//
// LocalSigner is suitable for development and testing:
//
//	privateKey, _ := crypto.GenerateKey()
//	signer := local.NewSigner(privateKey)
//
// TransitSigner provides enterprise-grade security with Vault:
//
//	signer, _ := vault.NewSigner(vaultURL, token, "transit", "my-key")
//
// KMSSigner integrates with AWS infrastructure:
//
//	signer, _ := kms.NewSigner(ctx, "arn:aws:kms:...")
//
// PrivySigner provides wallet-as-a-service for customer-facing apps:
//
//	signer, _ := privy.NewSignerFromEnv()
//
// FireblocksSigner provides custody infrastructure:
//
//	signer, _ := fireblocks.NewSignerFromEnv()
//
// # Integration with Transact Client
//
// Use any signer with the transact client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//	signature, err := client.Transact.SignOperation(operation, signer)
//
// # Production Considerations
//
// For production deployments:
//   - Use TLS for all key management communication
//   - Implement proper authentication (not root tokens)
//   - Enable audit logging
//   - Use least-privilege policies
//   - Consider HSM integration for highest security
//   - Implement key rotation policies
package signer
