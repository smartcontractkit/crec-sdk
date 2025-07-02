# Chainlink Runtime Environment Client Library (CRELib)

CRELib is a client library for the Chainlink Runtime Environment (CRE), designed to facilitate the
development of services that interact with onchain data and services.

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

## Verifiable Events

Receiving events consists of several phases:
- Reading the event from the Chainlink Verifiable Network
- Verifying the event's authenticity and integrity using digital signatures
- Decoding the verified event into a structured format that can be used by the application

## Operations

Sending operations allows interacting with onchain smart contracts using a flexible account abstraction model.
Operations can contain multiple transactions which will be executed atomically by an onchain smart account. Various
smart accounts are available to support a number of signature algorithms, such as ECDSA and RSA.

Using the helper services allows for the easy formation of the onchain transaction payloads, but it is also possible
to create custom transactions with the application performing its own contract calldata encoding.

### Signing

The signing of operations is performed by various implementations of the `Signer` interface. Currently, the library
supports signing using a local ECDSA private key, but additional signing methods will be added in the future.

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

### CCIP Service

The CCIP (Cross Chain Interoperability Protocol) service allows for the transfer of tokens and sending of messages
between different blockchains. The CCIP service can optionally include the token approvals for the tokens attached
to the CCIP message sending operation.



