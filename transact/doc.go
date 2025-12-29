// Package transact provides operations for signing and sending blockchain operations.
//
// The transact client handles the full operation lifecycle: preparing transactions,
// signing with EIP-712 typed data, and submitting to the CREC network for
// gas-sponsored execution.
//
// # Architecture
//
// The transact package is composed of two main components:
//
//  1. EIP-712 Handler (transact/eip712) - Handles EIP-712 hashing and signing operations.
//     This client has no network dependencies and can be used independently for
//     offline signing workflows.
//
//  2. Transact Client - Handles API operations for submitting signed operations
//     to the CREC network. This client embeds the EIP-712 handler and delegates
//     hash/sign operations to it.
//
// # Usage
//
// Transact is typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	// Create a signer
//	signer, _ := local.NewSigner(privateKey)
//
//	// Execute an operation
//	result, err := client.Transact.ExecuteOperation(ctx, channelID, signer, operation, chainSelector)
//
// For advanced use cases, create the client directly:
//
//	transactClient, err := transact.NewClient(&transact.Options{
//	    CRECClient: apiClient,
//	    Logger:     &logger,
//	})
//
// For offline signing workflows without network dependencies:
//
//	handler, err := eip712.NewHandler(&eip712.Options{
//	    Logger: logger,
//	})
//	hash, signature, err := handler.SignOperation(ctx, operation, signer, chainSelector)
//
// # Building Operations
//
// An Operation bundles one or more transactions for atomic execution:
//
//	operation := &types.Operation{
//	    ID:      big.NewInt(time.Now().Unix()), // Unique ID to prevent replay
//	    Account: executorAccount,                // Smart account address
//	    Transactions: []types.Transaction{
//	        {
//	            To:    targetContract,
//	            Value: big.NewInt(0),
//	            Data:  encodedCalldata,
//	        },
//	    },
//	}
//
// # Signing Operations
//
// Use [Client.SignOperation] to compute the EIP-712 hash and sign:
//
//	hash, signature, err := client.Transact.SignOperation(ctx, operation, signer, chainSelector)
//
// Or compute just the hash with [Client.HashOperation]:
//
//	hash, err := client.Transact.HashOperation(operation, chainSelector)

// # Sending Operations
//
// Use [Client.ExecuteOperation] to sign and send in one step:
//
//	result, err := client.Transact.ExecuteOperation(ctx, channelID, signer, operation, chainSelector)
//	fmt.Printf("Operation: %s, Status: %s\n", result.OperationId, result.Status)
//
// Or send a pre-signed operation with [Client.SendSignedOperation]:
//
//	result, err := client.Transact.SendSignedOperation(ctx, channelID, operation, signature, chainSelector)
//
// # Signers
//
// The package supports multiple signer implementations:
//
//	// Local private key
//	import "github.com/smartcontractkit/crec-sdk/transact/signer/local"
//	signer, _ := local.NewSigner(privateKey)
//
//	// AWS KMS
//	import "github.com/smartcontractkit/crec-sdk/transact/signer/kms"
//	signer, _ := kms.NewSigner(ctx, kms.Options{KeyID: "...", Region: "us-east-1"})
//
//	// HashiCorp Vault
//	import "github.com/smartcontractkit/crec-sdk/transact/signer/vault"
//	signer, _ := vault.NewSigner(vault.Options{Address: "...", Token: "...", KeyName: "..."})
//
//	// Privy
//	import "github.com/smartcontractkit/crec-sdk/transact/signer/privy"
//	signer, _ := privy.NewSigner(privy.Options{AppID: "...", AppSecret: "...", WalletID: "..."})
//
// # Operation Management
//
// List and retrieve operations:
//
//	// List operations
//	ops, hasMore, err := client.Transact.ListOperations(ctx, ListOperationsInput{
//	    ChannelID: channelID,
//	})
//
//	// Get a specific operation
//	op, err := client.Transact.GetOperation(ctx, channelID, operationID)
//
// # Operation Lifecycle
//
// Operations progress through these states:
//   - pending: Created, waiting to be sent
//   - sent: Submitted to the blockchain
//   - confirmed: Successfully executed on-chain
//   - failed: Execution failed
//
// # Error Handling
//
// All errors can be inspected with errors.Is:
//
//	if errors.Is(err, ErrChannelNotFound) {
//	    // Handle missing channel
//	}
//	if errors.Is(err, ErrOperationNotFound) {
//	    // Handle missing operation
//	}
//	if errors.Is(err, ErrSignOperation) {
//	    // Handle signing failure
//	}
package transact
