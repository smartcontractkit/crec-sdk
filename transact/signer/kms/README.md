# KMS Signer

AWS KMS-based signer implementation for Ethereum ECDSA signatures.

## Overview

Signs Ethereum transactions using ECDSA keys stored in AWS KMS. Private keys never leave AWS hardware security modules.

## Prerequisites

1. **AWS KMS Key** with secp256k1 curve (`ECC_SECG_P256K1`)
2. **AWS Credentials** configured
3. **KMS Permissions**: `kms:Sign` and `kms:GetPublicKey`

### Creating a KMS Key

```bash
aws kms create-key \
  --key-usage SIGN_VERIFY \
  --key-spec ECC_SECG_P256K1 \
  --description "Ethereum signing key"
```

## Usage

```go
import "github.com/smartcontractkit/crec-sdk/transact/signer/kms"

// Create signer with KMS key ID/ARN/alias
signer, err := kms.NewSigner(ctx, "arn:aws:kms:us-west-2:123456789012:key/...")

// Sign a 32-byte hash
signature, err := signer.Sign(ctx, hash)
```

Custom AWS configuration:

```go
signer, err := kms.NewSignerWithConfig(awsConfig, keyID)
```

## Integration Testing

To run integration tests against real AWS KMS:

1. **Set up AWS CLI profile**:

   ```bash
   aws sso login
   ```

2. **Export environment variables**:

   ```bash
   export AWS_PROFILE=<your-profile>
   export KMS_KEY_ARN=<your-key-arn>
   ```

3. **Run integration tests**:
   ```bash
   go test -v ./... -run ^TestKMSSignerIntegration
   ```
