// Package wallets provides operations for managing wallets in the CREC platform.
//
// Wallets are blockchain wallets that can be used to sign transactions and interact
// with smart contracts. They can be configured with allowed signers and are associated
// with specific blockchain networks through chain selectors.
//
// # Usage
//
// Wallets are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:               "production-wallet",
//	    ChainSelector:      "ethereum-mainnet",
//	    WalletOwnerAddress: "0x1234...",
//	    WalletType:         "ecdsa",
//	})
//
// For advanced use cases, create the client directly:
//
//	walletsClient, err := wallets.NewClient(&wallets.Options{
//	    APIClient: apiClient,
//	    Logger:    &logger,
//	})
//
// # Creating Wallets
//
// Create a new wallet with required configuration:
//
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:               "my-wallet",
//	    ChainSelector:      "ethereum-sepolia",
//	    WalletOwnerAddress: "0xabcdef...",
//	    WalletType:         "ecdsa",
//	    AllowedEcdsaSigners: &[]string{"0x123...", "0x456..."},
//	})
//	fmt.Printf("Created: %s (ID: %s, Address: %s)\n",
//	    *wallet.Name, wallet.WalletId, wallet.Address)
//
// # Listing Wallets
//
// Use [Client.List] with optional filtering:
//
//	// List all wallets
//	wallets, hasMore, err := client.Wallets.List(ctx, ListInput{})
//
//	// Filter by name and chain selector
//	filterName := "production"
//	filterChain := "ethereum-mainnet"
//	wallets, hasMore, err := client.Wallets.List(ctx, ListInput{
//	    Name:          &filterName,
//	    ChainSelector: &filterChain,
//	    Limit:         ptr(10),
//	})
//
// # Getting and Updating Wallets
//
// Retrieve a specific wallet by ID:
//
//	wallet, err := client.Wallets.Get(ctx, walletID)
//
// Update a wallet's name:
//
//	err := client.Wallets.Update(ctx, walletID, UpdateInput{
//	    Name: "updated-wallet-name",
//	})
//
// # Error Handling
//
// All operations return typed errors that can be checked:
//
//	if errors.Is(err, wallets.ErrWalletNotFound) {
//	    // Handle not found case
//	}
//
//	if errors.Is(err, wallets.ErrNameRequired) {
//	    // Handle validation error
//	}
package wallets
