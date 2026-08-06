// Package wallets provides operations for managing Smart Wallets in the CREC platform.
//
// Smart Wallets are blockchain wallets that can be used to sign transactions and interact
// with smart contracts. They are associated with specific blockchain networks through chain
// selectors and carry type-specific configuration.
//
// # Usage
//
// Smart Wallets are typically accessed through the main SDK client:
//
//	client, _ := crec.NewClient(baseURL, apiKey)
//
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:               "production-wallet",
//	    ChainSelector:      "5009297550715157269",
//	    WalletOwnerAddress: "0x1234...",
//	    WalletType:         "ecdsa",
//	    Configuration:      apiClient.WalletConfiguration{"allowed_signers": []string{"0x..."}},
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
// Wallet creation takes a type-specific Configuration object. The expected shape depends on
// the wallet type and is validated by the server at creation time. The most common key is
// "allowed_signers", which for ECDSA wallets is a list of Ethereum addresses and for RSA
// wallets is a list of RSA public keys (each a map with hex-encoded exponent "e" and modulus
// "n").
//
// Create an ECDSA wallet:
//
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:               "my-wallet",
//	    ChainSelector:      "5009297550715157269",
//	    WalletOwnerAddress: "0xabcdef...",
//	    WalletType:         "ecdsa",
//	    Configuration: apiClient.WalletConfiguration{
//	        "allowed_signers": []string{"0x123...", "0x456..."},
//	    },
//	})
//
// Create an RSA wallet:
//
//	wallet, err := client.Wallets.Create(ctx, CreateInput{
//	    Name:               "my-wallet",
//	    ChainSelector:      "5009297550715157269",
//	    WalletOwnerAddress: "0xabcdef...",
//	    WalletType:         "rsa",
//	    Configuration: apiClient.WalletConfiguration{
//	        "allowed_signers": []map[string]string{
//	            {"e": "0x010001", "n": "0x00c458..."},
//	        },
//	    },
//	})
//
// The Configuration object is the canonical way to supply type-specific options such as
// allowed signers. The deprecated allowed_ecdsa_signers/allowed_rsa_signers top-level fields
// are no longer exposed by the SDK; callers must use Configuration instead.
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
