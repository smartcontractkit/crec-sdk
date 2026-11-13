package wallets_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/mocks/server"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	"github.com/smartcontractkit/crec-sdk/wallets"
)

// Example_upgradeWalletEndToEnd demonstrates the full wallet upgrade flow
// against the SDK's in-memory mock server: discover versions, resolve and
// inspect the upgrade target, sign with the owner key, submit, and poll the
// upgrade to a terminal outcome.
//
// In production, point the client at the CREC API and use your custody
// provider's signer (KMS, HSM, Vault) instead of a local key.
func Example_upgradeWalletEndToEnd() {
	ctx := context.Background()

	// Mock CREC API for the example; in production use the real base URL.
	mock := server.NewMockServer()
	defer mock.Close()

	crecAPIClient, err := apiClient.NewClientWithResponses(mock.TestServer.URL)
	if err != nil {
		log.Fatal(err)
	}
	client, err := wallets.NewClient(&wallets.Options{APIClient: crecAPIClient})
	if err != nil {
		log.Fatal(err)
	}

	// The wallet owner's key. In production, use your custody signer.
	ownerKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		log.Fatal(err)
	}
	ownerSigner := local.NewSigner(ownerKey)
	ownerAddress := crypto.PubkeyToAddress(ownerKey.PublicKey).Hex()

	// Create a wallet on Ethereum Sepolia; it deploys at the catalog's default
	// version for the network.
	wallet, err := client.Create(ctx, wallets.CreateInput{
		Name:               "example-wallet",
		ChainSelector:      "16015286601757825753", // Ethereum Sepolia
		WalletOwnerAddress: ownerAddress,
		WalletType:         apiClient.WalletTypeECDSA,
		Configuration: apiClient.WalletConfiguration{
			"allowed_signers": []string{ownerAddress},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wallet deployed at version %s\n", *wallet.ImplementationVersion)

	// Discover the published versions for the wallet's type, filtered to the
	// wallet's chain so deployments carry the addresses to sign.
	versions, _, err := client.ListTypeVersions(ctx, wallet.WalletType, wallets.ListTypeVersionsInput{
		ChainSelector: &wallet.ChainSelector,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range versions {
		fmt.Printf("published version: %s (default on network: %v)\n",
			v.ImplementationVersion, v.Deployments[0].IsDefault)
	}

	// Resolve the upgrade target and render label and address for owner
	// inspection: what the owner inspects is what they sign.
	target, err := client.ResolveUpgradeTarget(ctx, wallet.WalletId, "2025.09.0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("upgrading %s -> %s at %s\n",
		target.CurrentVersion, target.Version, target.ImplementationAddress)

	// Sign, submit, and poll the upgrade to a terminal outcome.
	upgrade, err := client.UpgradeWallet(ctx, wallet.WalletId, target, wallets.UpgradeOptions{
		OwnerSigner:  ownerSigner,
		Deadline:     time.Now().Add(time.Hour),
		PollInterval: 10 * time.Millisecond,
		PollTimeout:  5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("upgrade outcome: %s\n", upgrade.Outcome)

	// The wallet is recorded on the target version, which publishes the
	// administrative capabilities its implementation supports.
	adminOp, err := client.SubmitAdminOperation(ctx, wallet.WalletId, "setSignerAllowed",
		map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		},
		wallets.AdminOperationOptions{
			OwnerSigner:  ownerSigner,
			Deadline:     time.Now().Add(time.Hour),
			PollInterval: 10 * time.Millisecond,
			PollTimeout:  5 * time.Second,
		})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("admin operation %s outcome: %s\n", adminOp.Capability, adminOp.Outcome)

	// Output:
	// wallet deployed at version 2025.06.0
	// published version: 2025.06.0 (default on network: true)
	// published version: 2025.09.0 (default on network: false)
	// upgrading 2025.06.0 -> 2025.09.0 at 0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0
	// upgrade outcome: succeeded
	// admin operation setSignerAllowed outcome: succeeded
}
