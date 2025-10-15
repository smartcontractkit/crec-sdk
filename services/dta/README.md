# DTA Service

This DTA (Digital Transfer Agent) service is a Go SDK for composing on-chain operations against DTA contracts and for decoding DTA CREC verifiable events into strongly-typed Go structures.

This README focuses on three things:

- Decoding: how to turn a base64 verifiable event payload into a usable VerifiableEvent and a typed ConcreteEvent
- Prepare operations: how to build transactions for OpenMarketplace and DTAWallet contracts
- Events: what events are supported, their Go types, and how to use them in your app

## Table of Contents

- [What you get](#what-you-get)
- [Decoding](#decoding)
- [Prepare operations](#prepare-operations)
  - [Example](#example)
- [Events and usage patterns](#events-and-usage-patterns)
- [Dev Notes](#dev-notes)

## What you get

- A single Decode function that base64-decodes and unmarshals CREC events into VerifiableEvent and a typed ConcreteEvent
- A set of Prepare\*Operation builders that return a transact/types.Operation containing one or more transactions ready to be signed/sent by your transaction client

## Decoding

CREC emits verifiable events as base64-encoded JSON. Use Decode to obtain a VerifiableEvent with a populated ConcreteEvent.

Key types (simplified):

- VerifiableEvent: top-level envelope with CreatedAt, Event, Metadata, Transaction, and ConcreteEvent
- WorkflowEvent.Attributes: a map[string]Attribute where Attribute has Key, OnChain, Value, Visibility
- EventName(): resolves the event name using attributes["event_type"].Value, or falls back to the outer Event.Name. Unknown names return EventUnknown
- ConcreteEvent: an interface that will be set to the matching typed struct, e.g., *SubscriptionRequested, *DTASettlementOpened, etc.

Decode usage:

```go
ve, err := dta.Decode(ctx, crecEvent)
if err != nil { /* handle */ }

log.Printf("Name=%s Hash=%s", ve.EventName(), ve.Transaction.Hash)

// Attributes helpers
attrs := ve.Metadata.WorkflowEvent.Attributes
if attrs.Has("request_id") {
  v, _ := attrs.Get("request_id")
  log.Printf("request_id=%s", v)
}

// Access the typed event via ConcreteEvent
switch ve.EventName() {
case dta.EventSubscriptionRequested:
  e := ve.ConcreteEvent.(*dta.SubscriptionRequested)
  log.Printf("Amount=%s TokenID=%s", e.Amount.String(), e.FundTokenId.Hex())
case dta.EventDTASettlementOpened:
  e := ve.ConcreteEvent.(*dta.DTASettlementOpened)
  log.Printf("RequestType=%d Shares=%s Currency=%d", e.RequestType, e.Shares, e.Currency)
}
```

Decoding and type mapping rules:

- The custom UnmarshalJSON on VerifiableEvent creates the concrete struct based on EventName and maps attribute values into strongly-typed fields
- Numeric big values are parsed using big.Int SetString base 10
- Small integers use strconv.ParseUint with explicit bit sizes
- Addresses and hashes are parsed with go-ethereum common.HexToAddress/HexToHash

## Prepare operations

All builders return (\*transact/types.Operation, error). Operation includes:

- ID: a timestamp-based big.Int
- Account: the sender (ServiceOptions.AccountAddress)
- Transactions: 1 or more low-level contract calls with To, Value, Data

### Example

```go
package main

import (
  "context"
  "log"
  "math/big"

  "github.com/ethereum/go-ethereum/common"
  "github.com/smartcontractkit/crec-sdk/services/dta"
)

func main() {
  // Create service
  svc, err := dta.NewService(&dta.ServiceOptions{
    DTAOpenMarketplaceAddress: "0x...OpenMarketplace",
    DTAWalletAddress:          "0x...Wallet",
    AccountAddress:            "0x...YourEOA",
  })
  if err != nil { log.Fatal(err) }

  // Build a subscription with automatic ERC20 approval
  fundAdmin := common.HexToAddress("0xFUNDADMIN...")
  var fundTokenId [32]byte // fill with your 32-byte ID
  amount := big.NewInt(1_000_000)
  usdc := common.HexToAddress("0xUSDC...")

  op, err := svc.PrepareRequestSubscriptionWithTokenApprovalOperation(fundAdmin, fundTokenId, amount, usdc)
  if err != nil { log.Fatal(err) }

  log.Printf("Operation %s with %d txs ready for signing", op.ID.String(), len(op.Transactions))
}
```

Signing and sending:

- This package only builds operations. Use your transaction client (see transact/\* in this repo) to sign and send Operation.Transactions

## Events and usage patterns

Supported event names (see events.go)

Each maps to a Go struct with correctly typed fields. Example dispatch:

```go
ve, _ := dta.Decode(ctx, ev)
switch ve.EventName() {
case dta.EventRedemptionRequested:
  e := ve.ConcreteEvent.(*dta.RedemptionRequested)
  // use e.FundTokenId, e.Shares, e.CreatedAt
case dta.EventDistributorRequestProcessed:
  e := ve.ConcreteEvent.(*dta.DistributorRequestProcessed)
  // use e.Status, e.Shares, e.Error
}
```

Working with attributes directly is simplified with provided methods `Has`,`Get`,`Require`,`Default`

```go
attrs := ve.Metadata.WorkflowEvent.Attributes
if created, ok := attrs.Get("created_at"); ok {
  // created is string form; convert as needed
}
reqID := attrs.Default("request_id", "")
if reqID == "" { /* handle missing */ }
```

## Dev Notes

- If you introduce new event types, add them to events.go and update Decode’s UnmarshalJSON mapping accordingly
