// Package privy provides a signer using Privy's wallet-as-a-service platform.
//
// The Privy signer integrates with Privy's managed wallet infrastructure,
// using the personal_sign method for Ethereum-compatible message signing.
//
// # Usage with Environment Variables (Recommended)
//
// Set environment variables:
//
//	PRIVY_APP_ID      - Your Privy app ID (required)
//	PRIVY_APP_SECRET  - Your Privy app secret (required)
//	PRIVY_WALLET_ID   - Your Privy wallet ID (required)
//	PRIVY_BASE_URL    - Privy API base URL (optional, defaults to https://api.privy.io)
//
// Create a signer:
//
//	signer, err := privy.NewSignerFromEnv()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Usage with Explicit Parameters
//
// Create a signer with explicit parameters:
//
//	signer, err := privy.NewSigner(
//	    "your-app-id",
//	    "your-app-secret",
//	    "your-wallet-id",
//	    privy.WithBaseURL("https://custom-api.privy.io"), // optional
//	)
//
// # Signing
//
// Sign a 32-byte hash:
//
//	hash := crypto.Keccak256([]byte("message"))
//	signature, err := signer.Sign(ctx, hash)
//
// # Wallet Address
//
// Retrieve the wallet's Ethereum address:
//
//	address, err := signer.GetWalletAddress(ctx)
//
// # Testing
//
// For unit testing, inject a custom HTTP client:
//
//	signer, err := privy.NewSigner(
//	    appID, appSecret, walletID,
//	    privy.WithHTTPClient(mockHTTPClient),
//	    privy.WithBaseURL("https://api.privy.io"),
//	)
//
// # Use Cases
//
//   - Customer-facing applications
//   - Wallet-as-a-service integrations
//   - Managed wallet infrastructure
package privy
