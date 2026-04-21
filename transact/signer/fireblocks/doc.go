// Package fireblocks provides a [signer.Signer] implementation using Fireblocks' custody infrastructure.
//
// Fireblocks is a platform that provides secure key management
// and signing capabilities. This package enables signing raw message hashes using keys
// stored in Fireblocks vault accounts.
//
// # Authentication
//
// Fireblocks uses JWT-based API authentication. You need:
//   - An API key from the Fireblocks console
//   - An RSA private key (PEM-encoded) for signing JWT tokens
//   - A vault account ID containing the signing key
//   - An asset ID (e.g., "ETH", "ETH_TEST5" for Sepolia)
//
// # Basic Usage
//
//	privateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
//	...your RSA private key...
//	-----END RSA PRIVATE KEY-----`
//
//	signer, err := fireblocks.NewSigner(
//	    "your-api-key",
//	    privateKeyPEM,
//	    "0",           // vault account ID
//	    "ETH",         // asset ID
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	hash := crypto.Keccak256([]byte("message to sign"))
//	signature, err := signer.Sign(context.Background(), hash)
//
// # Environment Variables
//
// For convenience, you can use environment variables:
//
//	export FIREBLOCKS_API_KEY="your-api-key"
//	export FIREBLOCKS_API_SECRET="/path/to/private-key.pem"  // or PEM content directly
//	export FIREBLOCKS_VAULT_ACCOUNT_ID="0"
//	export FIREBLOCKS_ASSET_ID="ETH"
//	export FIREBLOCKS_BASE_URL="https://api.fireblocks.io"   // optional
//
//	signer, err := fireblocks.NewSignerFromEnv()
//
// # Configuration Options
//
// The signer supports functional options for customization:
//
//	signer, err := fireblocks.NewSigner(
//	    apiKey, privateKeyPEM, vaultAccountID, assetID,
//	    fireblocks.WithTimeout(30 * time.Second),
//	    fireblocks.WithPollingInterval(time.Second),
//	    fireblocks.WithBaseURL("https://sandbox-api.fireblocks.io"),
//	)
//
// # Signing Flow
//
// When Sign is called (raw hash signing), the signer:
//  1. Creates a RAW signing operation in Fireblocks
//  2. Polls for operation completion
//  3. Extracts the ECDSA signature (r, s, v) from the response
//  4. Returns an Ethereum-compatible 65-byte signature
//
// # EIP-712 Typed Data Signing
//
// The signer also implements [signer.TypedDataSigner] for EIP-712 typed data:
//
//	typedData := &signer.TypedData{
//	    Types: map[string][]signer.TypedDataField{
//	        "Permit": {
//	            {Name: "owner", Type: "address"},
//	            {Name: "spender", Type: "address"},
//	            {Name: "value", Type: "uint256"},
//	            {Name: "nonce", Type: "uint256"},
//	            {Name: "deadline", Type: "uint256"},
//	        },
//	    },
//	    PrimaryType: "Permit",
//	    Domain: signer.TypedDataDomain{
//	        Name:              "MyToken",
//	        Version:           "1",
//	        ChainID:           1,
//	        VerifyingContract: "0x...",
//	    },
//	    Message: map[string]any{
//	        "owner":    "0x...",
//	        "spender":  "0x...",
//	        "value":    "1000000000000000000",
//	        "nonce":    0,
//	        "deadline": 1893456000,
//	    },
//	}
//	signature, err := signer.SignTypedData(ctx, typedData)
//
// This uses Fireblocks' TYPED_MESSAGE operation, which allows Fireblocks
// to see the full typed data structure for policy enforcement.
//
// # Security Considerations
//
//   - Store the RSA private key securely (HSM, secrets manager, etc.)
//   - Use environment variables or secure secret injection for credentials
//   - Configure appropriate Fireblocks policies for signing operations
//   - The API key and RSA key should have minimal required permissions
//   - Consider using Fireblocks' sandbox environment for testing
//
// # Error Handling
//
// The signer returns detailed errors for common failure cases:
//   - Invalid or missing credentials
//   - Network failures
//   - Operation rejections or policy violations
//   - Timeout waiting for operation completion
//
// # Integration with CREC SDK
//
// Use the Fireblocks signer with the transact client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//	signer, _ := fireblocks.NewSignerFromEnv()
//	signature, err := client.Transact.SignOperation(operation, signer)
package fireblocks
