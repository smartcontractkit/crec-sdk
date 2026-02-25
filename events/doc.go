// Package events provides operations for polling and verifying events from CREC.
//
// Events are blockchain occurrences detected by watchers and signed by the
// Decentralized Oracle Network (DON). This package enables fetching events
// from channels, verifying their cryptographic signatures, and decoding
// the event data into structured Go types.
//
// # Usage
//
// Events are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(
//	    baseURL,
//	    apiKey,
//	    crec.WithEventVerification(3, []string{
//	        "0x5db070ceabcf97e45d96b4f951a1df050ddb5559",
//	        "0xadebb9657c04692275973230b06adfabacc899bc",
//	        "0xc868bbb5d93e97b9d780fc93811a00ca7c016751",
//	    }),
//	)
//
//	events, hasMore, err := client.Events.Poll(ctx, channelID, nil)
//
// For advanced use cases, create the client directly:
//
//	eventsClient, err := events.NewClient(&events.Options{
//	    CRECClient:            apiClient,
//	    MinRequiredSignatures: 3,
//	    ValidSigners:          []string{"0x...", "0x...", "0x..."},
//	})
//
// # Polling Events
//
// Use [Client.Poll] to fetch events from a channel:
//
//	// Poll with default parameters
//	events, hasMore, err := client.Events.Poll(ctx, channelID, nil)
//
//	// Poll with pagination
//	params := &apiClient.GetChannelsChannelIdEventsParams{
//	    Limit: ptr(100),
//	}
//	events, hasMore, err := client.Events.Poll(ctx, channelID, params)
//
// # Verifying Events
//
// CRITICAL: Always verify events before processing. Verification ensures
// the event was signed by enough trusted DON members and matches the expected workflow.
//
// ## Verifying Watcher Events
//
// Use [Client.Verify] to verify watcher events using your organization ID.
// The workflow owner address is derived automatically from the org ID:
//
//	for _, event := range events {
//	    verified, err := client.Events.Verify(&event, orgID)
//	    if err != nil {
//	        // Handle verification error
//	        continue
//	    }
//	    if !verified {
//	        // Not enough valid signatures or workflow mismatch, skip this event
//	        continue
//	    }
//
//	    // Event is verified, safe to process
//	    processEvent(event)
//	}
//
// If you already have the workflow owner address, use [Client.VerifyWithWorkflowOwner]:
//
//	verified, err := client.Events.VerifyWithWorkflowOwner(&event, workflowOwnerAddress)
//
// ## Verifying Operation Status Events
//
// Use [Client.VerifyOperationStatus] to verify operation status events using your org ID:
//
//	verified, err := client.Events.VerifyOperationStatus(&event, orgID)
//	if err != nil {
//	    // Handle verification error
//	}
//	if !verified {
//	    // Not enough valid signatures or workflow mismatch
//	}
//
// Or with a known workflow owner address via [Client.VerifyOperationStatusWithWorkflowOwner]:
//
//	verified, err := client.Events.VerifyOperationStatusWithWorkflowOwner(&event, workflowOwnerAddress)
//
// ## Deriving Workflow Owner from Org ID
//
// Use [WorkflowOwnerFromOrgID] to derive the workflow owner Ethereum address
// from an org ID without performing verification:
//
//	ownerAddress, err := events.WorkflowOwnerFromOrgID(orgID)
//
// # Decoding Events
//
// Use [Client.Decode] to unpack event data into a Go struct:
//
//	var decodedEvent MyEventStruct
//	err := client.Events.Decode(&event, &decodedEvent)
//
// Or decode to a map for dynamic handling:
//
//	var data map[string]interface{}
//	err := client.Events.Decode(&event, &data)
//
// # Decoding Verifiable Events
//
// The event payload contains a base64-encoded VerifiableEvent field with rich event metadata.
// Use [Client.DecodeVerifiableEvent] to decode it into a models.VerifiableEvent struct:
//
//	import "github.com/smartcontractkit/crec-api-go/models"
//
//	// Get the watcher event payload
//	watcherPayload, err := event.Payload.AsWatcherEventPayload()
//	if err != nil {
//	    // Handle error
//	}
//
//	// Decode the verifiable event
//	verifiableEvent, err := client.Events.DecodeVerifiableEvent(&watcherPayload)
//	if err != nil {
//	    // Handle error
//	}
//
//	// Access event metadata
//	fmt.Printf("Event: %s at %v\n", verifiableEvent.Name, verifiableEvent.Timestamp)
//	if verifiableEvent.ChainFamily != nil {
//	    fmt.Printf("Chain Family: %s\n", *verifiableEvent.ChainFamily)
//	}
//
//	// Access chain event details (for EVM chains)
//	if verifiableEvent.ChainEvent != nil {
//	    evmEvent, err := verifiableEvent.ChainEvent.AsEVMEvent()
//	    if err == nil {
//	        fmt.Printf("Contract: %s, Block: %d\n", evmEvent.Address, evmEvent.BlockNumber)
//	        fmt.Printf("Tx Hash: %s\n", evmEvent.TxHash)
//	        // Access event parameters
//	        if evmEvent.Params != nil {
//	            for key, value := range *evmEvent.Params {
//	                fmt.Printf("  %s: %v\n", key, value)
//	            }
//	        }
//	    }
//	}
//
// For operation status events, use [Client.DecodeOperationStatusVerifiableEvent]:
//
//	opStatusPayload, _ := event.Payload.AsOperationStatusPayload()
//	verifiableEvent, err := client.Events.DecodeOperationStatusVerifiableEvent(&opStatusPayload)
//
// # JSON Serialization
//
// Convert events to JSON for logging or storage:
//
//	jsonBytes, err := client.Events.ToJSON(&event)
//
// # Event Types
//
// Events can be:
//   - Watcher events: Blockchain events captured by watchers (verify with [Client.Verify])
//   - Operation status events: Status updates for operations (verify with [Client.VerifyOperationStatus])
//   - Application events: Other status updates for watchers, etc.
//
// Both watcher events and operation status events support cryptographic verification.
//
// # Error Handling
//
// All errors can be inspected with errors.Is:
//
//	if errors.Is(err, ErrChannelNotFound) {
//	    // Handle missing channel
//	}
//	if errors.Is(err, ErrInvalidEventHash) {
//	    // Event data was tampered with
//	}
//	if errors.Is(err, ErrOnlyWatcherEventsSupported) {
//	    // Tried to verify a non-watcher event with Verify
//	}
//	if errors.Is(err, ErrOnlyOperationStatusSupported) {
//	    // Tried to verify a non-operation-status event with VerifyOperationStatus
//	}
//	if errors.Is(err, ErrDecodeVerifiableEvent) {
//	    // Failed to decode base64 verifiable event (invalid base64 or JSON)
//	}
//	if errors.Is(err, ErrDeriveWorkflowOwner) {
//	    // Failed to derive workflow owner address from org ID
//	}
package events
