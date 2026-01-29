// Package crec provides a unified SDK for interacting with the CREC system.
//
// # Overview
//
// The CREC SDK follows a resource-oriented design with a single top-level [Client]
// that provides access to sub-clients for each domain:
//
//   - [github.com/smartcontractkit/crec-sdk/channels] - Channel CRUD operations
//   - [github.com/smartcontractkit/crec-sdk/events] - Event polling and verification
//   - [github.com/smartcontractkit/crec-sdk/networks] - List available networks
//   - [github.com/smartcontractkit/crec-sdk/transact] - Operation signing and sending
//   - [github.com/smartcontractkit/crec-sdk/watchers] - Watcher CRUD operations
//
// # Getting Started
//
// Create a new client with your API credentials:
//
//	import "github.com/smartcontractkit/crec-sdk"
//
//	client, err := crec.NewClient(
//	    "https://api.crec.example.com",
//	    "your-api-key",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Working with Channels
//
// Channels are logical groupings for organizing watchers, events, and operations:
//
//	// Create a channel
//	channel, err := client.Channels.Create(ctx, channels.CreateInput{
//	    Name: "my-channel",
//	})
//
//	// List channels
//	channels, hasMore, err := client.Channels.List(ctx, channels.ListInput{})
//
//	// Delete a channel
//	err = client.Channels.Delete(ctx, channelID)
//
// # Working with Watchers
//
// Watchers monitor blockchain events on specific smart contracts:
//
//	// Create a watcher with a known domain
//	watcher, err := client.Watchers.CreateWithDomain(ctx, channelID, watchers.CreateWithDomainInput{
//	    ChainSelector: "16015286601757825753",
//	    Address:       "0x...",
//	    Domain:        "dvp",
//	    Events:        []string{"SettlementProposed"},
//	})
//
//	// Wait for watcher to become active
//	watcher, err = client.Watchers.WaitForActive(ctx, channelID, watcher.WatcherId, 30*time.Second)
//
// # Working with Events
//
// Poll for events from a channel:
//
//	// Poll events
//	events, hasMore, err := client.Events.Poll(ctx, channelID, &apiClient.GetChannelsChannelIdEventsParams{
//	    Limit: ptr(100),
//	})
//
//	// Verify an event's authenticity
//	valid, err := client.Events.Verify(&event)
//
// # Signing and Sending Operations
//
// The Transact client provides full operation lifecycle management:
//
//	// Create an operation from transactions
//	operation := &types.Operation{
//	    ID:           big.NewInt(time.Now().Unix()),
//	    Account:      executorAccount,
//	    Transactions: []types.Transaction{...},
//	}
//
//	// Sign and send in one step
//	result, err := client.Transact.ExecuteOperation(ctx, channelID, signer, operation, chainSelector)
//
//	// Or manually: sign, then send
//	hash, signature, err := client.Transact.SignOperation(ctx, operation, signer, chainSelector)
//	result, err := client.Transact.SendSignedOperation(ctx, channelID, operation, signature, chainSelector)
//
// # Configuration Options
//
// Use functional options to customize the client:
//
//	client, err := crec.NewClient(
//	    baseURL,
//	    apiKey,
//	    crec.WithLogger(logger),
//	    crec.WithHTTPClient(customHTTPClient),
//	    crec.WithEventVerification(2, []string{"0xSigner1", "0xSigner2", "0xSigner3"}),
//	    crec.WithWatcherPolling(5*time.Second, 10*time.Second),
//	)
//
// # Using Individual Sub-Clients
//
// If you only need a subset of the SDK's functionality, you can create individual
// sub-clients without instantiating the full [Client]. Use [NewAPIClient] to create
// an authenticated API client, then pass it to the sub-client you need:
//
//	import (
//	    "github.com/smartcontractkit/crec-sdk"
//	    "github.com/smartcontractkit/crec-sdk/channels"
//	    "github.com/smartcontractkit/crec-sdk/watchers"
//	)
//
//	// Create an authenticated API client
//	api, err := crec.NewAPIClient("https://api.crec.example.com", "your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create only the sub-clients you need
//	channelsClient, err := channels.NewClient(&channels.Options{APIClient: api})
//	watchersClient, err := watchers.NewClient(&watchers.Options{APIClient: api})
//
// This approach is useful when:
//   - You want to minimize dependencies in a specific package
//   - You're building a focused service that only uses one domain
//   - You need fine-grained control over sub-client configuration
//
// # Error Handling
//
// All errors are wrapped with context and can be inspected using errors.Is:
//
//	if errors.Is(err, channels.ErrChannelNotFound) {
//	    // Handle 404
//	}
//
//	if errors.Is(err, transact.ErrOperationNotFound) {
//	    // Handle operation not found
//	}
package crec
