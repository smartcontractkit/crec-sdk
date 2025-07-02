# Chainlink Runtime Environment Client Library (CRELib)

CRELib is a client library for the Chainlink Runtime Environment (CRE), designed to facilitate the
development of applications that interact with onchain data and services.

## Overview

The CRELib integrates with the following capabilities of the Chainlink Verifiable Network:

* Receiving verifiable events from the blockchain with high assurance of the event's authenticity. Events can come
  from well known services in which the events are decoded and decorated with extensive metadata to increase the
  usefulness of the event, or events can be received from any smart contract and decoding of the event can be done
  by the application.
* Sending operations to the blockchain using an account abstraction model such that an operation can
  contain a batch of transactions, have gas sponsorship and use a wider variety of signature algorithms, such as
  RSA signatures.

The CRELib includes a number of helper services to simplify both reading events from and sending transactions to a
number of Chainlink onchain systems, such as the DvP (Delivery vs Payment) service, the CCIP
(Cross Chain Interoperability Protocol) service, DTA (Digital Transfer Agent) service, and more.

## Example

An example application using the CRELib can be found in
the [cvn-example-payment-processor](https://github.com/smartcontractkit/cvn-example-payment-processor) repository.

## Packages

The following packages are available in the CRELib:
* `client`: The main client package for interacting with the Chainlink Verifiable Network.
* `events`: Provides functionality for receiving and decoding verifiable events from the Chainlink Verifiable Network.
* `transact`: Provides functionality for sending onchain operations using the account abstraction model.
* `services/dvp`: Provides the DvP (Delivery vs Payment) service for asset and payment exchange.
* `services/ccip`: Provides the CCIP (Cross Chain Interoperability Protocol) service for cross-chain token transfers and messaging.

### Documentation

The recommended way to explore the code documentation is to use `godoc`.

Ensure godoc is installed:
```bash
go install golang.org/x/tools/cmd/godoc@latest
```

Run the godoc server
```bash
godoc -http :8080
```

And you can view the documentation in your browser:
* client: [http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/client/](http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/client/)
* events: [http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/events/](http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/events/)
* transact: [http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/transact/](http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/transact/)
* dvp service: [http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/services/dvp/](http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/services/dvp/)
* ccip service: [http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/services/ccip/](http://localhost:8080/pkg/github.com/smartcontractkit/cvn-sdk/services/ccip/)

## Verifiable Events

Receiving events consists of several phases:
- Reading the event from the Chainlink Verifiable Network
- Verifying the event's authenticity and integrity using digital signatures
- Decoding the verified event into a structured format that can be used by the application

### Example Usage

```go
import (
    "github.com/smartcontractkit/cvn-sdk/client"
    "github.com/smartcontractkit/cvn-sdk/events"
)

// create a CVN client pointed to the Chainlink Verifiable Network URL
cvnClient, _ := client.NewCVNClient(cvnURL)

// Create CVN events client
cvnEventsClient, _ := events.NewClient(
    cvnClient, 
    &events.ClientOptions{
        MinRequiredSignatures: 3,
        ValidSigners: []string{
            "0x5db070ceabcf97e45d96b4f951a1df050ddb5559",
            "0xadebb9657c04692275973230b06adfabacc899bc",
            "0xc868bbb5d93e97b9d780fc93811a00ca7c016751",
            "0x1804f720c6c42b8075d03f3ddda8bd3cf49960de",
            "0xf191da826a7757ea2e3a8a5e147ddb378d6d0efe",
        },
    },
)

// Get events from CVN
eventList, _ := cvnEventsClient.GetEvents(context.Background())

for _, event := range *eventList {
    // Verify the event's authenticity and integrity
    verified, _ := cvnEventsClient.Verify(event)
    if verified {
        // Decode the event into a structured format
        var decodedEvent map[string]interface{}
        cvnEventsClient.Decode(event, &decodedEvent)

        handle(decodedEvent) // Handle the decoded event
    } else {
        fmt.Println("Event verification failed")
    }
}
```

## Transacting

Sending onchain operations allows interacting with onchain smart contracts using a flexible account abstraction model.
Operations can contain multiple transactions which will be executed atomically by an onchain smart account. Various
smart accounts are available to support a number of signature algorithms, such as ECDSA and RSA.

Using the helper services allows for the easy formation of the onchain transaction payloads, but it is also possible
to create custom transactions with the application performing its own contract calldata encoding.

### Signing

The signing of operations is performed by various implementations of the `Signer` interface. Currently, the library
supports signing using a local ECDSA private key, but additional signing methods will be added in the future.

### Example Usage


```go
import (
    "github.com/smartcontractkit/cvn-sdk/client"
    "github.com/smartcontractkit/cvn-sdk/transact"
    "github.com/smartcontractkit/cvn-sdk/transact/signer"
)

// create a CVN client pointed to the Chainlink Verifiable Network URL
cvnClient, _ := client.NewCVNClient(cvnURL)

// Create CVN transact client
cvnTransactClient, _ = transact.NewClient(
    cvnClient,
    &transact.ClientOptions{
        ChainId: "1337",
    },
)

// Create a transaction to call a smart contract function
operation := &transactTypes.Operation{
    ID: big.NewInt(time.Now().Unix()), // unique ID for the operation to prevent replay attacks
    Account: accountAddress, // address of the smart account that will perform the operation
    Transactions: []*transactTypes.Transaction { // list of transactions to be executed atomically by the smart account
        {
            To:    target, // address of the contract to call
            Value: big.NewInt(0),
            Data:  calldata, // encoded calldata for the contract call
        },
    },
}

// Create a local signer with the private key of an address authorized to sign the operation in the smart account
operationSigner = signer.NewLocalSigner(privateKey)

// Sign the operation using the local signer
signature, _ := cvnTransactClient.SignOperation(operation, operationSigner)

// Send the signed operation to the Chainlink Verifiable Network for relaying onchain
cvnTransactClient.SendSignedOperation(context.Background(), operation, signature)
```

## Services

The helper services include a number of packages to simplfy the interaction with the following Chainlink systems:

### DvP Service

The DvP (Delivery vs Payment) service allows for the secure and trustless transfer of assets between parties,
ensuring that the transfer of assets is only completed when both parties have agreed to the settlement terms
and the payment has been made.

The CRELib DVP helper service supports the following features:

- Proposing a settlement as the seller of an asset token
- Accepting of a settlement as the buyer of an asset token
- Execution of a settlement as a designated 3rd party such as the offchain payment network

The DvP service can optionally include the token approval/hold transaction in the settlement operations.

For more details on the DvP service, see the [DVP Service README](services/dvp/README.md).

### CCIP Service

The CCIP (Cross Chain Interoperability Protocol) service allows for the transfer of tokens and sending of messages
between different blockchains. The CCIP service can optionally include the token approvals for the tokens attached
to the CCIP message sending operation.



