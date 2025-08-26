
## Privy Signer

The Privy signer integrates with Privy's wallet-as-a-service platform to provide secure signing operations using managed wallets. It uses Privy's [`personal_sign`](https://docs.privy.io/api-reference/wallets/ethereum/personal-sign) method for Ethereum-compatible message signing.

## Usage

### Using Environment Variables (Recommended)

```go
package main

import (
    "context"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/smartcontractkit/cvn-sdk/transact/signer/privy"
)

// Environment Variables
// - `PRIVY_APP_ID` - Your Privy app ID (required)
// - `PRIVY_APP_SECRET` - Your Privy app secret (required)
// - `PRIVY_WALLET_ID` - Your Privy wallet ID (required)
// - `PRIVY_BASE_URL` - Privy API base URL (optional, defaults to https://api.privy.io)

func main() {
    // Create signer from environment variables
    signer, err := privy.NewSignerFromEnv()
    if err != nil {
        panic(err)
    }

    // Other options:
    // Create signer with explicit parameters
    // signer, err := privy.NewSigner(
    //     "your-app-id",
    //     "your-app-secret", 
    //     "your-wallet-id",
    //     "https://api.privy.io",
    // )
    // if err != nil {
    //     panic(err)
    // }
    
    // Sign data
    ctx := context.Background()
    hash := crypto.Keccak256([]byte("hello world"))
    signature, err := signer.Sign(ctx, hash)
    if err != nil {
        panic(err)
    }
    
    // Use signature...
}