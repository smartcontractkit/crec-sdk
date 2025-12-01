// Package kms provides a signer using AWS Key Management Service.
//
// The KMS signer uses ECDSA keys stored in AWS KMS for signing operations.
// Private keys never leave AWS hardware security modules.
//
// # Prerequisites
//
//  1. AWS KMS Key with secp256k1 curve (ECC_SECG_P256K1)
//  2. AWS credentials configured (environment, IAM role, or config file)
//  3. KMS permissions: kms:Sign and kms:GetPublicKey
//
// Create a KMS key:
//
//	aws kms create-key \
//	    --key-usage SIGN_VERIFY \
//	    --key-spec ECC_SECG_P256K1 \
//	    --description "Ethereum signing key"
//
// # Usage
//
// Create a signer with the key ARN, key ID, or alias:
//
//	signer, err := kms.NewSigner(ctx, "arn:aws:kms:us-west-2:123456789012:key/...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Sign a 32-byte hash:
//
//	signature, err := signer.Sign(ctx, hash)
//
// # Custom AWS Configuration
//
// Use custom AWS configuration:
//
//	cfg, _ := config.LoadDefaultConfig(ctx,
//	    config.WithRegion("us-east-1"),
//	)
//	signer, err := kms.NewSignerWithConfig(cfg, keyID)
//
// # Testing
//
// For unit testing, inject a mock KMS client:
//
//	mockClient := &mocks.KMSClient{}
//	signer, err := kms.NewSignerWithClient(mockClient, keyID)
//
// # Features
//
//   - HSM-backed key storage
//   - AWS IAM integration
//   - CloudTrail audit logging
//   - Ethereum-compatible signatures (secp256k1)
//
// # Use Cases
//
//   - AWS-based infrastructure
//   - Serverless deployments (Lambda)
//   - Managed key lifecycle
package kms
