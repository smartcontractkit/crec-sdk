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
// the event was signed by enough trusted DON members and matches the expected workflow:
//
//	// The workflowId is the CID (Content Identifier) of the workflow that should
//	// have generated this event. Obtain this from your workflow deployment configuration.
//	workflowId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
//
//	for _, event := range events {
//	    verified, err := client.Events.Verify(&event, workflowId)
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
// # JSON Serialization
//
// Convert events to JSON for logging or storage:
//
//	jsonBytes, err := client.Events.ToJSON(&event)
//
// # Event Types
//
// Events can be:
//   - Watcher events: Blockchain events captured by watchers (support verification)
//   - Application events: Status updates for operations, watchers, etc.
//
// Only watcher events support cryptographic verification via [Client.Verify].
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
//	    // Tried to verify a non-watcher event
//	}
package events
