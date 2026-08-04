// Package wallets provides operations for managing Smart Wallets in the CREC platform.
//
// Smart Wallets are blockchain wallets that can be used to sign transactions and interact
// with smart contracts. They are associated with specific blockchain networks through chain
// selectors and carry type-specific allowed signer lists.
//
// # Usage
//
// Smart Wallets are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	ecdsaSigners := apiClient.ECDSASignersList{"0x..."}
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:                "production-wallet",
//	    ChainSelector:       "5009297550715157269",
//	    WalletOwnerAddress:  "0x1234...",
//	    WalletType:          "ecdsa",
//	    AllowedEcdsaSigners: &ecdsaSigners,
//	})
//
// For advanced use cases, create the client directly:
//
//	walletsClient, err := wallets.NewClient(&wallets.Options{
//	    APIClient: apiClient,
//	    Logger:    &logger,
//	})
//
// # Creating Smart Wallets
//
// The CREC API accepts explicit allowed-signer lists on wallet creation. ECDSA wallets
// take a list of Ethereum addresses; RSA wallets take a list of RSA public keys (each
// defined by a hex-encoded exponent E and modulus N).
//
// Create an ECDSA wallet:
//
//	ecdsaSigners := apiClient.ECDSASignersList{"0x123...", "0x456..."}
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:                "my-wallet",
//	    ChainSelector:       "5009297550715157269",
//	    WalletOwnerAddress:  "0xabcdef...",
//	    WalletType:          "ecdsa",
//	    AllowedEcdsaSigners: &ecdsaSigners,
//	})
//
// Create an RSA wallet:
//
//	rsaSigners := apiClient.RSASignersList{
//	    {E: "0x010001", N: "0x00c458..."},
//	}
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:              "my-wallet",
//	    ChainSelector:     "5009297550715157269",
//	    WalletOwnerAddress: "0xabcdef...",
//	    WalletType:        "rsa",
//	    AllowedRsaSigners: &rsaSigners,
//	})
//
// # Listing Smart Wallets
//
// Use [Client.List] with optional filtering:
//
//	// List all Smart Wallets
//	wallets, hasMore, err := client.Wallets.List(ctx, ListInput{})
//
//	// Filter by name and chain selector
//	filterName := "production"
//	filterChain := "5009297550715157269"
//	wallets, hasMore, err := client.Wallets.List(ctx, ListInput{
//	    Name:          &filterName,
//	    ChainSelector: &filterChain,
//	    Limit:         ptr(10),
//	})
//
// # Getting, Updating, and Archiving Smart Wallets
//
// Retrieve a specific Smart Wallet by ID:
//
//	wallet, err := client.Wallets.Get(ctx, walletID)
//
// Update a Smart Wallet's name:
//
//	err := client.Wallets.Update(ctx, walletID, UpdateInput{
//	    Name: "updated-wallet-name",
//	})
//
// Archive a Smart Wallet (sets status to "archived"):
//
//	wallet, err := client.Wallets.Archive(ctx, walletID)
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
