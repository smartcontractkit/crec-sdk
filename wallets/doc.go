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
// Wallet creation takes an optional type-specific Configuration object. The expected shape depends on
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
// allowed signers. It may be omitted when no type-specific options are needed.
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
//
// # Wallet Type Version Catalog
//
// The catalog publishes the implementation versions of each wallet type, with
// changelogs, support states, capability descriptors and per-network
// deployments. Consumers use it to discover upgrade targets and the
// administrative capabilities a wallet's recorded version accepts:
//
//	versions, hasMore, err := client.Wallets.ListTypeVersions(ctx, "ecdsa", wallets.ListTypeVersionsInput{
//	    ChainSelector: &wallet.ChainSelector, // resolve addresses for the wallet's network
//	})
//
//	version, err := client.Wallets.GetTypeVersion(ctx, "ecdsa", "2025.06.0")
//
// # Upgrading Wallets
//
// An upgrade is authorized by the wallet owner's signature over the pinned
// UpgradeAuthorization EIP-712 struct and relayed by CREC. The two-step shape
// keeps what the owner inspects — the version label beside the resolved
// implementation address — exactly what they sign:
//
//	target, err := client.Wallets.ResolveUpgradeTarget(ctx, walletID, "2025.09.0")
//	// show target.Version and target.ImplementationAddress for owner inspection
//
//	upgrade, err := client.Wallets.UpgradeWallet(ctx, walletID, target, wallets.UpgradeOptions{
//	    OwnerSigner: ownerSigner, // signs the authorization
//	    Deadline:    time.Now().Add(time.Hour),
//	})
//	// upgrade.Outcome is terminal: succeeded, failed, or expired
//
// The low-level [Client.CreateUpgrade], [Client.ListUpgrades] and
// [Client.GetUpgrade] methods expose the raw endpoints for custom flows;
// [Client.AwaitUpgrade] polls an upgrade to a terminal outcome.
//
// # Administrative Operations
//
// The capabilities a wallet's recorded version publishes (signer management,
// operation cancellation, policy engine attachment) are exercised the same
// way: the owner signs the capability's pinned EIP-712 struct, and the SDK
// relays the packed call as a signed self-call operation:
//
//	adminOp, err := client.Wallets.SubmitAdminOperation(ctx, walletID, "setSignerAllowed",
//	    map[string]any{"signer": "0x123...", "allowed": true},
//	    wallets.AdminOperationOptions{
//	        OwnerSigner: ownerSigner,
//	        Deadline:    time.Now().Add(time.Hour),
//	    })
//	// adminOp.Outcome is terminal: succeeded, failed, or expired
//
// Arguments must be canonically encoded exactly as submitted: integers as
// decimal strings, addresses and byte strings as lowercase 0x hex, booleans
// and strings as JSON values. The courier re-packs the call and requires it
// to equal the transaction's calldata byte for byte.
package wallets
