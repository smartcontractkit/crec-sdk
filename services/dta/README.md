# DTA Service

The DTA (Digital Transfer Agent) Service provides a comprehensive Go SDK for interacting with the DTA smart contracts. This service enables distributors to manage fund token requests and fund administrators to administer their marketplace and process subscription/redemption requests.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Service Configuration](#service-configuration)
- [Event Decoding](#event-decoding)
- [Operation Preparation](#operation-preparation)
- [Data Structures](#data-structures)
- [Testing](#testing)
- [Examples](#examples)
- [Contract Integration](#contract-integration)

## Overview

The DTA Service integrates with two main smart contracts:
- **DTAOpenMarketplaceU**: Handles fund token registration, distributor management, and subscription/redemption requests
- **DTAWalletU**: Manages settlement operations and token transfers

This service provides:
- ✅ **Unified Event Decoder** for all contract events
- ✅ **18 Operation Builders** for all major contract functions
- ✅ **Token Approval Integration** for seamless ERC20 payment token handling
- ✅ **Comprehensive Testing** with full coverage
- ✅ **Type-Safe Operations** with proper error handling

## Architecture

```
DTA Service
├── Event Decoding (Unified)
│   └── Single decoder returns a VerifiableEvent for all event types
└── Operation Preparation (19 operations)
    ├── Distributor Operations (6)
    │   ├── PrepareRegisterDistributorOperation
    │   ├── PrepareUpdateDistributorOperation
    │   ├── PrepareRequestSubscriptionOperation
    │   ├── PrepareRequestRedemptionOperation
    │   ├── PrepareRequestSubscriptionWithTokenApprovalOperation
    │   └── PrepareCancelDistributorRequestOperation
    ├── OpenMarketplace Admin Operations (7)
    │   ├── PrepareRegisterFundAdminOperation
    │   ├── PrepareRegisterFundTokenOperation
    │   ├── PrepareAllowDistributorForTokenOperation
    │   ├── PrepareDisallowDistributorForTokenOperation
    │   ├── PrepareProcessDistributorRequestOperation
    │   ├── PrepareEnableFundTokenOperation
    │   └── PrepareDisableFundTokenOperation
    └── DTAWallet Operations (6)
        ├── PrepareAllowDTAOperation
        ├── PrepareDisallowDTAOperation
        ├── PrepareWithdrawTokensOperation
        ├── PrepareCompleteRequestProcessingOperation
        ├── PrepareTransferWalletOwnershipOperation
        └── PrepareRenounceWalletOwnershipOperation
```

## Installation

```bash
go get github.com/smartcontractkit/cvn-sdk
```

## Quick Start

```go
package main

import (
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/smartcontractkit/cvn-sdk/services/dta"
)

func main() {
    // Create DTA service
    dtaService, err := dta.NewService(&dta.ServiceOptions{
        DTAOpenMarketplaceAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
        DTAWalletAddress:          "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
        AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Request a subscription with automatic token approval
    fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
    fundTokenId := [32]byte{1, 2, 3, /* ... */ 32}
    amount := big.NewInt(1000000) // 1M tokens (adjust for decimals)
    paymentToken := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
    
    operation, err := dtaService.PrepareRequestSubscriptionWithTokenApprovalOperation(
        fundAdminAddr, fundTokenId, amount, paymentToken,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Operation is ready to be signed and executed
    log.Printf("Operation ID: %s", operation.ID.String())
    log.Printf("Transactions: %d", len(operation.Transactions))
}
```

## Service Configuration

### ServiceOptions

```go
type ServiceOptions struct {
    Logger                      *zerolog.Logger  // Optional: custom logger
    DTAOpenMarketplaceAddress   string          // Required: marketplace contract address
    DTAWalletAddress            string          // Required: wallet contract address  
    AccountAddress              string          // Required: account performing operations
}
```

### Example Configuration

```go
// Basic configuration
opts := &dta.ServiceOptions{
    DTAOpenMarketplaceAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
    DTAWalletAddress:          "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
    AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
}

// With custom logger
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
opts.Logger = &logger

service, err := dta.NewService(opts)
```

## Event Decoding

The DTA service uses a unified event decoding model. All CVN events are base64-encoded verifiable events that decode into a single VerifiableEvent struct.

Example:

```go
// cvnEvent is client.Event from the CVN client
ve, err := dta.Decode(context.Background(), cvnEvent)
if err != nil {
    log.Fatal(err)
}

// Access core fields
log.Printf("Created At: %s", ve.CreatedAt)
log.Printf("Event Name (parsed): %s", ve.EventName())
log.Printf("Event Type: %s", ve.Event.Type)
log.Printf("Event Address: %s", ve.Event.Address)
log.Printf("Request ID: %s", ve.Event.RequestId)
log.Printf("Topic Hash: %s", ve.Event.TopicHash)

// Workflow attributes (key/value map)
if attr, ok := ve.Metadata.WorkflowEvent.Attributes["some_key"]; ok {
    log.Printf("Attribute %s = %s", attr.Key, attr.Value)
}
```

Full example using DecodeAndParseAttributes:

```go
package main

import (
    "context"
    "log"

    "github.com/smartcontractkit/cvn-sdk/client"
    "github.com/smartcontractkit/cvn-sdk/services/dta"
)

func handleEvent(ctx context.Context, cvnEvent client.Event) {
    // Decode the base64 event, get the typed name, the unified VerifiableEvent, and parsed attributes
    name, ve, attrs, err := dta.DecodeAndParseAttributes(ctx, cvnEvent)
    if err != nil {
        log.Printf("failed to decode/parse event: %v", err)
        return
    }

    // Use core fields from the unified VerifiableEvent
    log.Printf("Event: name=%s type=%s address=%s requestId=%s at=%s",
        name, ve.Event.Type, ve.Event.Address, ve.Event.RequestId, ve.CreatedAt)

    // Switch on the resolved event name and type-assert the parsed attrs
    switch name {
    case dta.EventSubscriptionRequested:
        e := attrs.(dta.SubscriptionRequested)
        log.Printf("SubscriptionRequested: fundTokenId=%x distributor=%s requestId=%x amount=%s createdAt=%d",
            e.FundTokenId, e.DistributorAddr, e.RequestId, e.Amount.String(), e.CreatedAt)
        // further processing

    case dta.EventRedemptionRequested:
        e := attrs.(dta.RedemptionRequested)
        log.Printf("RedemptionRequested: fundTokenId=%x distributor=%s requestId=%x shares=%s createdAt=%d",
            e.FundTokenId, e.DistributorAddr, e.RequestId, e.Shares.String(), e.CreatedAt)

    case dta.EventFundTokenRegistered:
        e := attrs.(dta.FundTokenRegistered)
        log.Printf("FundTokenRegistered: fundAdmin=%s fundTokenId=%x tokenAddr=%s navAddr=%s chainSelector=%d",
            e.FundAdminAddr, e.FundTokenId, e.FundTokenAddr, e.NavAddr, e.TokenChainSelector)

    // ... handle other event names as needed
    default:
        log.Printf("Unhandled event: %s (raw attributes available: %d keys)", name, len(ve.Metadata.WorkflowEvent.Attributes))
    }
}
```

Notes:
- VerifiableEvent is a single struct covering metadata, transaction, and workflow attributes for every event type.
- Use ve.EventName() to get a event name parsed from the "event_type" attribute of the WorkflowEvent struct.
- DecodeAndParseAttributes returns (EventName, VerifiableEvent, any), where the third value is a concrete, typed struct corresponding to the event (e.g., SubscriptionRequested). You can type-assert it based on the returned EventName.

## Operation Preparation

### User Operations

#### Request Subscription

```go
// Simple subscription request
operation, err := dtaService.PrepareRequestSubscriptionOperation(
    fundAdminAddr,  // Fund admin managing the token
    fundTokenId,    // 32-byte fund token identifier
    amount,         // Amount to subscribe (big.Int)
)

// With automatic token approval (recommended)
operation, err := dtaService.PrepareRequestSubscriptionWithTokenApprovalOperation(
    fundAdminAddr,    // Fund admin managing the token
    fundTokenId,      // 32-byte fund token identifier  
    amount,           // Amount to subscribe (big.Int)
    paymentTokenAddr, // ERC20 token address for payment
)
```

#### Request Redemption

```go
operation, err := dtaService.PrepareRequestRedemptionOperation(
    fundAdminAddr, // Fund admin managing the token
    fundTokenId,   // 32-byte fund token identifier
    shares,        // Number of shares to redeem (big.Int)
)
```

### Request Management Operations

#### Process Request

```go
// Process a pending distributor request
operation, err := dtaService.PrepareProcessDistributorRequestOperation(
    requestId, // 32-byte request identifier
)
```

#### Cancel Request

```go
// Cancel a pending distributor request
operation, err := dtaService.PrepareCancelDistributorRequestOperation(
    requestId, // 32-byte request identifier
)
```

### OpenMarketplace Admin Operations

#### Register Entities

```go
// Register a new distributor
operation, err := dtaService.PrepareRegisterDistributorOperation(
    distributorAddr,       // Distributor's address
    distributorWalletAddr, // Distributor's wallet address
)

// Update a distributor
operation, err := dtaService.PrepareUpdateDistributorOperation(
    distributorWalletAddr, // Distributor's new wallet address
)

// Register a new fund admin
operation, err := dtaService.PrepareRegisterFundAdminOperation(
    fundAdminAddr, // Fund admin's address
)
```

#### Token Management

```go
// Register a new fund token with complete metadata
fundTokenIdStr := "Test Token"
fundTokenIdBytes := []byte(fundTokenIdStr)
fundTokenIdHash := sha3.NewLegacyKeccak256().Sum(fundTokenIdBytes)
fundTokenId := [32]byte{}
copy(fundTokenId[:], fundTokenIdHash[:])
tokenData := dta.FundTokenData{
    FundTokenAddr:                 common.HexToAddress("0xToken..."),
    NavFeedDecimals:               18,
    PurchaseTokenRoundingDecimals: 2,
    PurchaseTokenDecimals:         6,
    FundRoundingDecimals:          3,
    FundTokenDecimals:             18,
    RequestsPerDay:                10,
    NavAddr:                       common.HexToAddress("0xNAV..."),  
    TokenChainSelector:            1234567890,
    DtaWalletAddr:                 common.HexToAddress("0xWallet..."),
    TimezoneOffsetSecs:            big.NewInt(-18000), // -5 hours in seconds
    NavTTL:                        0,
    PaymentInfo: dta.DTAPaymentInfo{
        OffChainPaymentCurrency:    1, // USD
        PaymentTokenSourceAddr:     common.HexToAddress("0xPayment..."),
        PaymentSourceChainSelector: 1234567890,
        PaymentTokenDestAddr:       common.HexToAddress("0xDest..."),
    },
}

operation, err := dtaService.PrepareRegisterFundTokenOperation(fundTokenId, tokenData)
```

```go
// Allow distributor for specific token
operation, err := dtaService.PrepareAllowDistributorForTokenOperation(
    fundTokenId,    // 32-byte fund token identifier
    distributorAddr, // Distributor to allow
)

// Disallow distributor for specific token
operation, err := dtaService.PrepareDisallowDistributorForTokenOperation(
    fundTokenId,    // 32-byte fund token identifier
    distributorAddr, // Distributor to disallow
)

// Enable/disable fund tokens
enableOp, err := dtaService.PrepareEnableFundTokenOperation(fundTokenId)
disableOp, err := dtaService.PrepareDisableFundTokenOperation(fundTokenId)
```

### DTAWallet Operations

#### DTA Access Control

```go
// Allow a DTA contract to interact with fund tokens
operation, err := dtaService.PrepareAllowDTAOperation(
    dtaAddr,          // DTA contract address
    dtaChainSelector, // Chain selector for DTA
    fundTokenId,      // Fund token ID
    fundTokenAddr,    // Fund token contract address
    dta.TokenBurnTypeBurn, // Burn type (None, Burn, Transfer)
)

// Disallow a DTA contract
operation, err := dtaService.PrepareDisallowDTAOperation(
    dtaAddr,          // DTA contract address
    dtaChainSelector, // Chain selector for DTA
    fundTokenId,      // Fund token ID
)
```

#### Wallet Management

```go
// Withdraw tokens from the DTA wallet
operation, err := dtaService.PrepareWithdrawTokensOperation(
    tokenAddr,     // Token contract address
    recipientAddr, // Recipient address
    amount,        // Amount to withdraw (big.Int)
)

// Complete request processing
operation, err := dtaService.PrepareCompleteRequestProcessingOperation(
    requestId,   // Request ID to complete ([32]byte)
    success,     // Whether processing was successful (bool)
    errorData,   // Error data if failed ([]byte)
)

// Transfer wallet ownership
operation, err := dtaService.PrepareTransferWalletOwnershipOperation(
    newOwnerAddr, // New owner address
)

// Renounce wallet ownership
operation, err := dtaService.PrepareRenounceWalletOwnershipOperation()
```

## Data Structures

### FundTokenData

Complete metadata structure for fund token registration:

```go
type FundTokenData struct {
    FundTokenAddr                 common.Address // Token contract address
    NavFeedDecimals               uint8          // Decimals for NAV feed
    PurchaseTokenRoundingDecimals uint8          // Decimals for rounding purchase token
    PurchaseTokenDecimals         uint8          // Decimals for purchase token
    FundRoundingDecimals          uint8          // Decimals for rounding fund token
    FundTokenDecimals             uint8          // Decimals for fund token
    RequestsPerDay                uint8          // Daily request limit
    NavAddr                       common.Address // NAV oracle address
    TokenChainSelector            uint64         // Chain selector for token
    DtaWalletAddr                 common.Address // Associated DTA wallet
    TimezoneOffsetSecs            *big.Int       // Timezone offset for requests
    NavTTL                        *big.Int       // Seconds from time the NAV is set until it expires
    PaymentInfo                   DTAPaymentInfo // Payment configuration
}
```

### DTAPaymentInfo

Payment configuration for DTA operations:

```go
type DTAPaymentInfo struct {
    OffChainPaymentCurrency    uint8          // Currency enum (1=USD, etc.)
    PaymentTokenSourceAddr     common.Address // Source payment token
    PaymentTokenDestAddr       common.Address // Destination payment token
}
```

## Testing

### Running Tests

```bash
# Run all DTA tests
go test ./services/dta

# Run with verbose output
go test -v ./services/dta

# Run specific test
go test -v ./services/dta -run TestPrepareRequestSubscriptionOperation

# Run with coverage
go test -cover ./services/dta
```

### Test Coverage

The test suite includes:
- ✅ **Service initialization** with various configurations
- ✅ **All operation preparation methods** with proper validation
- ✅ **Event decoding functionality** with mock data
- ✅ **Error handling** for invalid inputs
- ✅ **Transaction structure validation** for all operations
- ✅ **Multi-transaction operations** (like token approval + subscription)

### Example Test

```go
func TestPrepareRequestSubscriptionOperation(t *testing.T) {
    service, err := dta.NewService(&dta.ServiceOptions{
        DTAOpenMarketplaceAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
        DTAWalletAddress:          "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
        AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
    })
    require.NoError(t, err)

    operation, err := service.PrepareRequestSubscriptionOperation(
        common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
        [32]byte{1, 2, 3, /* ... */ 32},
        big.NewInt(1000000),
    )
    require.NoError(t, err)
    require.NotNil(t, operation)
    require.Len(t, operation.Transactions, 1)
}
```

## Examples

### Complete Subscription Flow

```go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/smartcontractkit/cvn-sdk/services/dta"
    "github.com/smartcontractkit/cvn-sdk/transact"
)

func subscriptionExample() {
    // 1. Create DTA service
    dtaService, err := dta.NewService(&dta.ServiceOptions{
        DTAOpenMarketplaceAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
        DTAWalletAddress:          "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
        AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Prepare subscription with token approval
    fundAdminAddr := common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621")
    fundTokenId := [32]byte{1, 2, 3, /* ... fund token ID ... */ 32}
    subscriptionAmount := big.NewInt(1000000) // 1M tokens
    paymentToken := common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86")
    
    operation, err := dtaService.PrepareRequestSubscriptionWithTokenApprovalOperation(
        fundAdminAddr, fundTokenId, subscriptionAmount, paymentToken,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. Sign and execute operation (using transact client)
    transactClient := transact.NewClient(/* ... transact options ... */)
    
    result, err := transactClient.SignOperation(context.Background(), operation)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Operation executed successfully: %s", result.OperationID)
}
```

### Event Monitoring

```go
func monitorDTAEvents(cvnClient *client.Client, dtaService *dta.Service) {
    // Listen for CVN events
    eventChan := make(chan client.Event)

    go func() {
        for event := range eventChan {
            ve, err := dta.Decode(context.Background(), *event)
            if err != nil {
                log.Printf("failed to decode event: %v", err)
                continue
            }

            // Use unified struct
            log.Printf("Event: name=%s type=%s address=%s requestId=%s at=%s",
                ve.EventName(), ve.Event.Type, ve.Event.Address, ve.Event.RequestId, ve.CreatedAt)

            // Optionally dispatch by name
            switch ve.EventName() {
            case dta.EventSubscriptionRequested:
                // handle subscription
            case dta.EventRedemptionRequested:
                // handle redemption
            // ... handle other event names as needed
            }
        }
    }()

    // Start event monitoring
    cvnClient.SubscribeToEvents(eventChan)
}
```

### Admin Operations

```go
func adminOperationsExample() {
    dtaService, _ := dta.NewService(&dta.ServiceOptions{
        DTAOpenMarketplaceAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
        DTAWalletAddress:          "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
        AccountAddress:            "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc", // Admin account
    })
    
    // Register a new fund token
    fundTokenId := [32]byte{/* unique token ID */}
    tokenData := dta.FundTokenData{
        FundTokenAddr:                 common.HexToAddress("0xTokenContract..."),
        NavFeedDecimals:               18,
        PurchaseTokenRoundingDecimals: 2,  // USD
        PurchaseTokenDecimals:         6,  // USDC
        FundTokenRoundingDecimals:     3,
        FundTokenDecimals:             18,
        RequestsPerDay:                1,
        NavAddr:                       common.HexToAddress("0xNAVOracle..."),
        TokenChainSelector:            1,
        DtaWalletAddr:                 common.HexToAddress("0xDTAWallet..."),
        TimezoneOffsetSecs:            big.NewInt(0), // UTC
        NavTTL:                        0, // manual processing
        PaymentInfo: dta.DTAPaymentInfo{
            OffChainPaymentCurrency:    147, // USD
            PaymentTokenSourceAddr:     common.HexToAddress("0xUSDC..."),
            PaymentTokenDestAddr:       common.HexToAddress("0xUSDC..."),
        },
    }
    
    registerOp, err := dtaService.PrepareRegisterFundTokenOperation(fundTokenId, tokenData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register a distributor for this token
    distributorAddr := common.HexToAddress("0xDistributor...")
    
    allowOp, err := dtaService.PrepareAllowDistributorForTokenOperation(fundTokenId, distributorAddr)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Admin operations prepared: Register=%s, Allow=%s", 
        registerOp.ID.String(), allowOp.ID.String())
}
```

## Contract Integration

### OpenMarketplace Functions

The service integrates with these key OpenMarketplace functions:
- `requestSubscription(address,bytes32,uint256)` → Returns request ID
- `requestRedemption(address,bytes32,uint256)` → Returns request ID  
- `processDistributorRequest(bytes32)` → Processes pending request
- `cancelDistributorRequest(bytes32)` → Cancels pending request
- `registerDistributor(address,address)` → Registers new distributor
- `updateDistributor(address)` → Updates distributor
- `registerFundAdmin(address)` → Registers new fund admin
- `registerFundToken(bytes32,tuple)` → Registers fund token with metadata
- `allowDistributorForToken(bytes32,address)` → Allows distributor for token
- `disallowDistributorForToken(bytes32,address)` → Disallows distributor
- `enableFundToken(bytes32)` → Enables fund token
- `disableFundToken(bytes32)` → Disables fund token

### Wallet Functions

The service also integrates with DTA Wallet functions for settlement operations.

### Event Monitoring

All events are decoded via the unified decoder into a VerifiableEvent, providing comprehensive monitoring capabilities for DTA operations.

---

## Support

For issues and questions:
- **Documentation**: Check this README and inline code documentation
- **Examples**: See the `examples/` directory in the repository
- **Testing**: Run the comprehensive test suite
- **Issues**: Open GitHub issues for bugs or feature requests

The DTA Service provides a complete, production-ready SDK for all DTA operations within the CVN ecosystem.
