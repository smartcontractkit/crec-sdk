# DvP Service

The DvP (Delivery vs Payment) service allows for the secure and trustless transfer of assets between parties,
ensuring that the transfer of assets is only completed when both parties have agreed to the settlement terms
and the payment has been made.

# Actors

The DvP service involves three main actors:

- **Seller**: The party proposing the settlement, typically the owner of the asset token.
- **Buyer**: The party accepting the settlement, typically the buyer of the asset token.
- **Executor**: An optional party executing the settlement, typically a representative of the offchain payment network.
  This party observes the offchain payment and is authorized to execute the settlement onchain once the payment is 
  confirmed.

# Multi-Chain Support

The DvP service supports multi-chain operations utilizing Chainlink CCIP, allowing the settlement to be proposed, 
accepted, and executed across different blockchains. This is particularly useful for scenarios where the asset token
and payment token are on different chains. A seller must propose a settlement on the chain that contains the asset
token. Similarly, a buyer must accept the settlement on the chain that contains the payment token. The executor can be
on either chain, but it is most commonly on the chain where the asset token is located.

# Payments

The DvP service supports both onchain and offchain payments. In the case of onchain payments, the payment token is
specified as part of the settlement. For offchain payments, the payment token can be set to `None`, and the
`paymentCurrency` field can be used to specify the currency of the payment. In this case, the buyer is expected
to make the payment offchain. At this point, the settlement can be excuted in one of three ways:

- The seller can directly execute the settlement.
- The buyer can execute the settlement themselves using the secret provided by the seller.
- The executor can execute the settlement, typically after observing the offchain payment.

# The Settlement Object

The settlement is represented by a `Settlement` object, which contains the following fields:

- `settlementId`: A unique identifier for the settlement, which is used to track the settlement process.
- `partyInfo` - A struct contains the address of the various actors involved in the settlement:
  - `buyerSourceAddress`: Address of buyer on the source chain where the payment originates.
  - `buyerDestinationAddress`: Address of buyer on the destination chain where the asset will be delivered.
  - `sellerSourceAddress`: Address of seller on the source chain where the asset originates.
  - `sellerDestinationAddress`: Address of seller on the destination chain where the payment will be delivered.
  - `executorAddress`: Optional address of the 3rd party designated as allowed to execute the settlement.
- `tokenInfo`: A struct contains the details of the asset and payment tokens:
  - `paymentTokenAmount`: Amount of payment token being paid by buyer to close the settlement.
  - `assetTokenAmount`: Amount of asset token being sold by seller to close the settlement.
  - `paymentTokenSourceAddress`: Address of payment token on the source chain.
  - `paymentTokenDestinationAddress`: Address of payment token on the destination chain.
  - `paymentCurrency`: The currency of the payment. Used for off-chain payment i.e. paymentTokenType is None.
  - `paymentTokenType`: The token type of the payment token.
  - `assetTokenSourceAddress`: Address of asset token being delivered on the source chain.
  - `assetTokenDestinationAddress`: Address of asset token being delivered on the destination chain.
  - `assetTokenType`: The token type of the asset token.
- `deliveryInfo`: A struct contains the details of the delivery:
  - `paymentSourceChainSelector`: CCIP chain selector of where buyer payment is originating.
  - `paymentDestinationChainSelector`: CCIP chain selector of where payment will be delivered to seller.
  - `assetSourceChainSelector`: CCIP chain selector of where seller assets are originating.
  - `assetDestinationChainSelector`: CCIP chain selector of where assets will be delivered to buyer.
- `secretHash`: A hash of a secret that can be provided from the seller to the buyer to allow the buyer to 
  execute the settlement. This would typically be used in an offchain payment scenario where the buyer provides
  the payment and then the seller provides the secret to the buyer to allow them to execute the settlement.
- `executeAfter`: A timestamp indicating when the settlement can be executed. This is used to ensure that the settlement
  is not executed before a certain time.
- `expiration`: A timestamp indicating when the settlement must be executed by. This is used to ensure that the
  settlement is not left open indefinitely.
- `ccipCallbackGasLimit`: The gas limit to supply on remote chains to allow the DvP coordinator to process
  the CCIP message.
- `data`: Additional data that can be included in the settlement, such as metadata or instructions for the
  settlement process.

# The Settlement Hash

The settlement is hashed using the `SettlementHash` function, which combines all the relevant fields of the settlement
into a single hash. This hash is used to uniquely identify the settlement. Other than the initial `proposeSettlement`
call, all other calls accept a `settlementHash` parameter to identify the settlement being acted upon.

# Token Types

The DvP service supports two types of tokens:
- **ERC-20**: Standard fungible tokens that follow the ERC20 token standard.
- **ERC-3643**: A permissioned token that is expected to have a "Hold Manager" attached, facilitating the hold and 
  release of tokens during the settlement process. This is particularly useful for ensuring that the asset token is 
  remains in the seller's wallet until the payment is settled.

# Example Usage

## Executing a Settlement as a Payment Network

In this example, we will demonstrate how to execute a settlement as a payment network using the DvP service. The
address of the payment network smart account must have been specified as the `executorAddress` in the settlement
when it was proposed by the seller. The payment network will observe the offchain payment and execute the settlement.

```go
import (
    "github.com/smartcontractkit/cvn-sdk/client"
    "github.com/smartcontractkit/cvn-sdk/event"
    "github.com/smartcontractkit/cvn-sdk/transact"
    "github.com/smartcontractkit/cvn-sdk/transact/signer"
    "github.com/smartcontractkit/cvn-sdk/services/dvp"
)

// create a CVN client pointed to the Chainlink Verifiable Network URL
cvnClient, _ := client.NewCVNClient(cvnURL)

// Create DVP service instance
dvpService, _ = dvp.NewService(
    &dvp.ServiceOptions{
        DvpCoordinatorAddress: dvpCoordinatorAddress, // address of the DvP coordinator contract
        AccountAddress: accountAddress, // the executor smart account performing the settlement execution
    },
)

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

// Create CVN transact client
cvnTransactClient, _ := transact.NewClient(
    cvnClient,
    &transact.ClientOptions{
        ChainId: "1337", // specify the chain ID for the CVN
    },
)

// Decode the CVN verifiable event into a DvP event. It is assumed that the event has been verified before this point.
dvpEvent, _ := dvpService.DecodeSettlementAccepted(event)

// Check if the event is a dvp SettlementAccepted event
if event.Service == "dvp" && event.Name == "SettlementAccepted" {
    
	// *** Perform Offchain Payment ***
	
    // Prepare an "executeSettlement" operation
    operation, _ := dvpService.PrepareExecuteSettlementOperation(dvpEvent.Event.SettlementHash)

    // Sign as a valid signer on the executor smart account
    signature, _ := cvnTransactClient.SignOperation(operation, operationSigner)

    // Create a local signer with the private key of an address authorized to sign the operation in the smart account
    operationSigner = signer.NewLocalSigner(privateKey)

    // Send the signed operation to the CVN for relaying onchain using the CVN transact client from the CRELib 
    // transact package
    cvnTransactClient.SendSignedOperation(context.Background(), operation, signature)
}
```