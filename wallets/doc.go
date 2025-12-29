// Package wallets provides operations for managing Smart Wallets in the CREC platform.
//
// Smart Wallets are blockchain wallets that can be used to sign transactions and interact
// with smart contracts. They can be configured with allowed signers and are associated
// with specific blockchain networks through chain selectors.
//
// # Usage
//
// Smart Wallets are typically accessed through the main SDK client:
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
// # Creating Smart Wallets
//
// Smart Wallets support different cryptographic algorithms for signing:
//   - ECDSA (Elliptic Curve Digital Signature Algorithm): Standard for Ethereum transactions,
//     uses elliptic curve cryptography. Specify "ecdsa" as the wallet type and provide
//     AllowedEcdsaSigners as hex-encoded public keys.
//   - RSA (Rivest-Shamir-Adleman): Alternative signing algorithm using RSA keys.
//     Specify "rsa" as the wallet type and provide AllowedRsaSigners with public
//     exponent (E) and modulus (N) values.
//
// Create a new Smart Wallet with required configuration:
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
// # Listing Smart Wallets
//
// Use [Client.List] with optional filtering:
//
//	// List all Smart Wallets
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
// # Getting and Updating Smart Wallets
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
