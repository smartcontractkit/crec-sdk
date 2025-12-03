<div align="center">
  <img src="assets/chainlink-logo.svg" alt="Chainlink" width="300" height="130"/>
</div>

# CRE Connect SDK

Build the next generation of verifiable applications with secure, blockchain-agnostic event processing and transaction execution — powered by the Chainlink Runtime Environment and Verifiable Network.

## What problem does CRE Connect SDK solve?

Building reliable blockchain applications requires handling:

**Event Verification Challenges:**

- Ensuring events from the blockchain are authentic and haven't been tampered with
- Decoding complex event data from various smart contracts
- Managing multiple signature verification schemes for trust

**Transaction Execution Complexities:**

- Batching multiple transactions atomically without gas estimation headaches
- Supporting various signature algorithms beyond traditional ECDSA
- Abstracting away account management while maintaining security

## How CRE Connect SDK solves it

CRE Connect SDK is a client library for the Chainlink Runtime Environment (CRE), designed to facilitate the development of applications that interact with onchain data and services.

- **Receiving verifiable events** from the blockchain with high assurance of authenticity
- **Sending operations** to the blockchain using account abstraction with gas sponsorship

## Key Features

- **🔐 Cryptographically Secure**: Multi-signature verification ensures event authenticity
- **⛽ Account Abstraction**: Batch transactions, gas sponsorship, and multiple signature support
- **🛠️ Developer-Friendly**: Rich helper services for common blockchain operations
- **🧱 Modular Design**: Use individual components or combine them for complex use cases

## Quick Start

### Prerequisites

- **Go 1.24 or higher** - Check with `go version`
- **Basic Go and blockchain knowledge**

### Installation

```bash
go get github.com/smartcontractkit/crec-sdk
```

### Initialize the Client

```go
import "github.com/smartcontractkit/crec-sdk"

client, err := crec.NewClient(
    "https://api.crec.chainlink.com",
    "your-api-key",
    crec.WithEventVerification(3, []string{
        "0x5db070ceabcf97e45d96b4f951a1df050ddb5559",
        "0xadebb9657c04692275973230b06adfabacc899bc",
        "0xc868bbb5d93e97b9d780fc93811a00ca7c016751",
    }),
)
```

The unified client provides access to all sub-clients:

```go
client.Channels   // Channel CRUD operations
client.Events     // Event polling and verification
client.Transact   // Operation signing and sending
client.Watchers   // Watcher CRUD operations
```

### Using Individual Sub-Clients

If you only need a subset of the SDK's functionality, create individual sub-clients using `NewAPIClient`:

```go
import (
    "github.com/smartcontractkit/crec-sdk"
    "github.com/smartcontractkit/crec-sdk/channels"
    "github.com/smartcontractkit/crec-sdk/watchers"
)

// Create an authenticated API client
api, err := crec.NewAPIClient(
    "https://api.crec.chainlink.com",
    "your-api-key",
)

// Create only the sub-clients you need
channelsClient, _ := channels.NewClient(&channels.Options{APIClient: api})
watchersClient, _ := watchers.NewClient(&watchers.Options{
    APIClient:    api,
    PollInterval: 5 * time.Second,
})
```

## Core Workflows

### 🔍 Receiving and Verifying Events

```mermaid
graph LR
    A[Smart Contract Event] --> B(DON Consensus) --> C(Event Signing) --> D{CREC API} --> E[SDK: Events Client] --> F[Verify Signatures] --> G[Your Logic]
```

**Poll and verify events:**

```go
// Poll events from a channel
events, hasMore, _ := client.Events.Poll(ctx, channelID, nil)

for _, event := range events {
    // CRITICAL: Always verify before processing
    verified, _ := client.Events.Verify(&event)
    if verified {
        var decoded map[string]interface{}
        client.Events.Decode(&event, &decoded)
        processEvent(decoded)
    }
}
```

### ⚡ Sending Signed Operations (Gas-less)

```mermaid
graph LR
    A[Build Operation] --> B[Sign] --> C[SDK: Transact Client] --> D{CREC API} --> E[Pays Gas & Relays] --> F[Smart Account Executes]
```

**Build and execute an operation:**

```go
import (
    "github.com/smartcontractkit/crec-sdk/transact/signer/local"
    "github.com/smartcontractkit/crec-sdk/transact/types"
)

// Build the operation
operation := &types.Operation{
    ID:      big.NewInt(time.Now().Unix()),
    Account: accountAddress,
    Transactions: []types.Transaction{
        {To: target, Value: big.NewInt(0), Data: calldata},
    },
}

// Create signer and execute
signer := local.NewSigner(privateKey)
result, _ := client.Transact.ExecuteOperation(ctx, channelID, signer, operation, chainSelector)
```

## Documentation

### API Reference

For complete API documentation, use Go's built-in documentation tools:

```bash
go doc github.com/smartcontractkit/crec-sdk
go doc github.com/smartcontractkit/crec-sdk/events
go doc github.com/smartcontractkit/crec-sdk/transact
go doc github.com/smartcontractkit/crec-sdk/channels
go doc github.com/smartcontractkit/crec-sdk/watchers
```

Or run a local documentation server:

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http :8080
# Navigate to http://localhost:8080/github.com/smartcontractkit/crec-sdk
```

### Complete Example

See the [crec-example-payment-processor](https://github.com/smartcontractkit/crec-example-payment-processor) repository for a full working application.

## Extensions

Protocol-specific helpers for common Chainlink systems:

- [crec-sdk-ext-ccip](https://github.com/smartcontractkit/crec-sdk-ext-ccip) - Cross-Chain Interoperability Protocol
- [crec-sdk-ext-dvp](https://github.com/smartcontractkit/crec-sdk-ext-dvp) - Delivery versus Payment
- [crec-sdk-ext-dta](https://github.com/smartcontractkit/crec-sdk-ext-dta) - Digital Token Assets

## Glossary

| Term                    | Description                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------|
| **CREC**                | CRE Connect - decentralized network providing cryptographic proof of event authenticity |
| **DON**                 | Decentralized Oracle Network - independent nodes that reach consensus and sign events   |
| **Verifiable Event**    | Blockchain event with cryptographic signatures from DON members                         |
| **Account Abstraction** | Transaction model with atomic execution, gas sponsorship, and flexible signing          |
| **Operation**           | Bundle of transactions executed atomically by a smart account                           |
| **EIP-712**             | Ethereum standard for typed data signing, used for operation signatures                 |
| **CCIP**                | Cross-Chain Interoperability Protocol for cross-chain transfers                         |
