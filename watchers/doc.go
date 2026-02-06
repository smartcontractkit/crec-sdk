// Package watchers provides operations for managing blockchain event watchers.
//
// Watchers monitor specific smart contract events on blockchain networks and
// trigger workflows when those events occur. They can be configured with
// pre-defined service ABIs (dvp, dta, test_consumer) or custom event ABIs.
//
// # Usage
//
// Watchers are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	watcher, err := client.Watchers.CreateWithService(ctx, channelID, watchers.CreateWithServiceInput{
//	    ChainSelector: "16015286601757825753",
//	    Address:       "0x...",
//	    Service:       "dvp",
//	    Events:        []string{"SettlementProposed"},
//	})
//
// For advanced use cases, create the client directly:
//
//	watchersClient, err := watchers.NewClient(&watchers.Options{
//	    APIClient: apiClient,
//	    Logger:    &logger,
//	})
//
// # Creating Watchers
//
// Use [Client.CreateWithService] for known contract types (dvp, dta, test_consumer):
//
//	name := "My DVP Watcher"
//	watcher, err := client.Watchers.CreateWithService(ctx, channelID, CreateWithServiceInput{
//	    Name:          &name,
//	    ChainSelector: "16015286601757825753",
//	    Address:       "0x1234...",
//	    Service:       "dvp",
//	    Events:        []string{"OperationExecuted"},
//	})
//
// Use [Client.CreateWithABI] for custom contracts with your own event definitions:
//
//	watcher, err := client.Watchers.CreateWithABI(ctx, channelID, CreateWithABIInput{
//	    ChainSelector: "16015286601757825753",
//	    Address:       "0x1234...",
//	    Events:        []string{"Transfer"},
//	    ABI: []EventABI{{
//	        Name: "Transfer",
//	        Type: "event",
//	        Inputs: []EventABIInput{
//	            {Name: "from", Type: "address", Indexed: true},
//	            {Name: "to", Type: "address", Indexed: true},
//	            {Name: "value", Type: "uint256", Indexed: false},
//	        },
//	    }},
//	})
//
// # Watcher Lifecycle
//
// Watchers start in "pending" status and must be deployed before becoming "active".
// Use [Client.WaitForActive] to poll until ready:
//
//	activeWatcher, err := client.Watchers.WaitForActive(ctx, channelID, watcherID, 2*time.Minute)
//	if err != nil {
//	    if errors.Is(err, ErrWatcherDeploymentFailed) {
//	        // Handle deployment failure
//	    }
//	}
//
// The watcher lifecycle states are:
//   - pending: Created, deployment in progress
//   - active: Deployed and monitoring events
//   - failed: Deployment failed (terminal)
//   - deleting: Deletion in progress (watcher returns 404 Not Found once fully deleted)
//
// # Listing and Filtering
//
// Use [Client.List] with [ListFilters] to query watchers:
//
//	statusFilter := []apiClient.WatcherStatus{apiClient.WatcherStatusActive}
//	result, err := client.Watchers.List(ctx, channelID, ListFilters{
//	    Status: &statusFilter,
//	    Limit:  ptr(10),
//	})
//	for _, w := range result.Data {
//	    fmt.Printf("Watcher: %s\n", w.WatcherId)
//	}
//
// # Error Handling
//
// All errors can be inspected with errors.Is:
//
//	if errors.Is(err, ErrWatcherNotFound) {
//	    // Handle 404
//	}
//	if errors.Is(err, ErrWatcherDeploymentFailed) {
//	    // Handle deployment failure
//	}
//	if errors.Is(err, ErrWaitForActiveTimeout) {
//	    // Handle timeout
//	}
//
// The client automatically retries transient errors (5xx, 429, network issues)
// during polling operations like [Client.WaitForActive] and [Client.WaitForDeleted].
package watchers
