// Package channels provides operations for managing channels in the CREC platform.
//
// Channels are logical groupings that organize watchers, events, and operations.
// Think of channels as "workspaces" or "projects" that group related blockchain
// monitoring and transaction activities together.
//
// # Usage
//
// Channels are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	channel, err := client.Channels.Create(ctx, CreateInput{
//	    Name: "production-settlements",
//	})
//
// For advanced use cases, create the client directly:
//
//	channelsClient, err := channels.NewClient(&channels.Options{
//	    APIClient: apiClient,
//	    Logger:    &logger,
//	})
//
// # Creating Channels
//
// Channel names must be unique within your account:
//
//	channel, err := client.Channels.Create(ctx, CreateInput{
//	    Name: "production-dvp-settlements",
//	})
//	fmt.Printf("Created: %s (ID: %s)\n", channel.Name, channel.ChannelId)
//
// # Listing Channels
//
// Use [Client.List] with optional filtering:
//
//	// List all channels
//	channels, hasMore, err := client.Channels.List(ctx, ListInput{})
//
//	// Filter by name
//	filterName := "production"
//	channels, hasMore, err := client.Channels.List(ctx, ListInput{
//	    Name:  &filterName,
//	    Limit: ptr(10),
//	})
//
// # Getting and Deleting
//
// Retrieve a specific channel by ID:
//
//	channel, err := client.Channels.Get(ctx, channelID)
//
// Delete a channel (soft delete):
//
//	err := client.Channels.Delete(ctx, channelID)
//
// # Integration with Other Clients
//
// Channels scope watchers, events, and operations:
//
//	// Create a channel
//	channel, _ := client.Channels.Create(ctx, CreateInput{Name: "my-channel"})
//
//	// Create watchers in the channel
//	watcher, _ := client.Watchers.CreateWithDomain(ctx, channel.ChannelId, ...)
//
//	// Poll events from the channel
//	events, _, _ := client.Events.Poll(ctx, channel.ChannelId, nil)
//
//	// Execute operations in the channel
//	result, _ := client.Transact.ExecuteOperation(ctx, channel.ChannelId, ...)
//
// # Error Handling
//
// All errors can be inspected with errors.Is:
//
//	if errors.Is(err, ErrChannelNotFound) {
//	    // Handle 404
//	}
//	if errors.Is(err, ErrChannelNameRequired) {
//	    // Handle validation error
//	}
package channels
