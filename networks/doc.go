// Package networks provides operations for listing available CREC networks.
//
// Networks represent blockchain networks supported by the CREC platform (e.g., EVM chains).
// Use this client to discover which networks you can use for watchers, wallets, and operations.
//
// # Usage
//
// Networks are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//	networks, hasMore, err := client.Networks.List(ctx)
//
// For advanced use cases, create the client directly:
//
//	networksClient, err := networks.NewClient(&networks.Options{
//	    APIClient: apiClient,
//	    Logger:    &logger,
//	})
package networks
