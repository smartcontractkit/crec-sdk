package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func TestMockServer_Health_Events_Listeners_Wallets(t *testing.T) {
	s := NewMockServer()
	defer s.Close()

	// Health via raw HTTP
	resp, err := http.Get(s.TestServer.URL + "/health-check")
	if err != nil {
		t.Fatalf("health-check request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health-check status: %d", resp.StatusCode)
	}

	// Use the generated CREC client against the mock server
	c, err := apiClient.NewClientWithResponses(
		s.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey test-key")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}

	_ = uuid.New() // Keep uuid import for future use

	// Wallets: create, list, get by id
	testWalletName := "Test Wallet"
	walletCfg := apiClient.WalletConfiguration{"allowed_signers": []string{}}
	wallet, err := c.CreateWalletWithResponse(
		context.Background(), apiClient.CreateWallet{
			WalletOwnerAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			ChainSelector:      "1337",
			Name:               testWalletName,
			WalletType:         apiClient.WalletTypeECDSA,
			Configuration:      &walletCfg,
		},
	)
	if err != nil {
		t.Fatalf("PostWallets: %v", err)
	}
	if wallet.StatusCode() != http.StatusCreated || wallet.JSON201 == nil {
		t.Fatalf("unexpected wallet creation response")
	}
	list, err := c.ListWalletsWithResponse(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetWallets: %v", err)
	}
	if list.StatusCode() != http.StatusOK || list.JSON200 == nil {
		t.Fatalf("unexpected wallets list response")
	}
	one, err := c.GetWalletWithResponse(context.Background(), wallet.JSON201.WalletId)
	if err != nil {
		t.Fatalf("GetWalletsWalletId: %v", err)
	}
	if one.StatusCode() != http.StatusOK || one.JSON200 == nil {
		t.Fatalf("unexpected wallet get response")
	}

	// Test updating wallet name
	updatedName := "Updated Wallet Name"
	updateRequest := apiClient.UpdateWallet{
		Name: &updatedName,
	}
	updateResp, err := c.UpdateWalletWithResponse(context.Background(), wallet.JSON201.WalletId, updateRequest)
	if err != nil {
		t.Fatalf("PatchWalletsWalletId: %v", err)
	}
	if updateResp.StatusCode() != http.StatusOK || updateResp.JSON200 == nil {
		t.Fatalf("unexpected wallet update response")
	}
	if updateResp.JSON200.Name != "Updated Wallet Name" {
		t.Fatalf("wallet name not updated correctly")
	}
}

func TestMockServer_WalletVersions_Upgrades_AdminOperations(t *testing.T) {
	s := NewMockServer()
	defer s.Close()

	c, err := apiClient.NewClientWithResponses(
		s.TestServer.URL,
		apiClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Apikey test-key")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}
	ctx := context.Background()

	// Wallet type catalog: list and get the seeded default version
	versions, err := c.ListWalletTypeVersionsWithResponse(ctx, apiClient.WalletTypeECDSA, nil)
	if err != nil {
		t.Fatalf("ListWalletTypeVersions: %v", err)
	}
	if versions.StatusCode() != http.StatusOK || versions.JSON200 == nil || len(versions.JSON200.Data) == 0 {
		t.Fatalf("unexpected wallet type versions response")
	}

	version, err := c.GetWalletTypeVersionWithResponse(ctx, apiClient.WalletTypeECDSA, mockDefaultWalletTypeVersion)
	if err != nil {
		t.Fatalf("GetWalletTypeVersion: %v", err)
	}
	if version.StatusCode() != http.StatusOK || version.JSON200 == nil || !version.JSON200.IsDefault {
		t.Fatalf("unexpected wallet type version response")
	}

	// Wallet creation picks up the type's default implementation version
	walletCfg := apiClient.WalletConfiguration{"allowed_signers": []string{}}
	wallet, err := c.CreateWalletWithResponse(ctx, apiClient.CreateWallet{
		WalletOwnerAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		ChainSelector:      "1337",
		Name:               "Upgradeable Wallet",
		WalletType:         apiClient.WalletTypeECDSA,
		Configuration:      &walletCfg,
	})
	if err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}
	if wallet.StatusCode() != http.StatusCreated || wallet.JSON201 == nil {
		t.Fatalf("unexpected wallet creation response")
	}
	if wallet.JSON201.ImplementationVersion == nil || *wallet.JSON201.ImplementationVersion != mockDefaultWalletTypeVersion {
		t.Fatalf("wallet not created on the default version: %+v", wallet.JSON201.ImplementationVersion)
	}
	walletId := wallet.JSON201.WalletId

	newVersion := "2025.07.0"
	signedOperation := apiClient.CreateOperation{
		ChainSelector:     "1337",
		Address:           wallet.JSON201.Address,
		Deadline:          1735689600,
		WalletOperationId: "1",
		Transactions:      []apiClient.TransactionRequest{},
	}
	authorization := apiClient.AdminAuthorization{
		Deadline:          1735689600,
		Signature:         "0x1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
		WalletOperationId: "2",
	}

	// Upgrade: accepted in flight, then settles asynchronously
	upgrade, err := c.CreateWalletUpgradeWithResponse(ctx, walletId, apiClient.CreateWalletUpgrade{
		TargetVersion:          newVersion,
		ImplementationAddress:  "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		Authorization:          authorization,
		Operation:              signedOperation,
	})
	if err != nil {
		t.Fatalf("CreateWalletUpgrade: %v", err)
	}
	if upgrade.StatusCode() != http.StatusAccepted || upgrade.JSON202 == nil {
		t.Fatalf("unexpected upgrade creation response")
	}
	if upgrade.JSON202.Outcome != apiClient.WalletUpgradeOutcomePending {
		t.Fatalf("upgrade should start pending, got %q", upgrade.JSON202.Outcome)
	}
	if upgrade.JSON202.FromVersion == nil || *upgrade.JSON202.FromVersion != mockDefaultWalletTypeVersion {
		t.Fatalf("unexpected upgrade from_version: %+v", upgrade.JSON202.FromVersion)
	}

	upgrades, err := c.ListWalletUpgradesWithResponse(ctx, walletId, nil)
	if err != nil {
		t.Fatalf("ListWalletUpgrades: %v", err)
	}
	if upgrades.StatusCode() != http.StatusOK || upgrades.JSON200 == nil || len(upgrades.JSON200.Data) != 1 {
		t.Fatalf("unexpected upgrades list response")
	}

	time.Sleep(150 * time.Millisecond)

	settled, err := c.GetWalletUpgradeWithResponse(ctx, walletId, upgrade.JSON202.UpgradeId)
	if err != nil {
		t.Fatalf("GetWalletUpgrade: %v", err)
	}
	if settled.StatusCode() != http.StatusOK || settled.JSON200 == nil || settled.JSON200.Outcome != apiClient.WalletUpgradeOutcomeSucceeded {
		t.Fatalf("upgrade should have succeeded, got %+v", settled.JSON200)
	}

	// The wallet moved onto the target version
	after, err := c.GetWalletWithResponse(ctx, walletId)
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if after.JSON200 == nil || after.JSON200.Status != apiClient.WalletStatusDeployed {
		t.Fatalf("wallet should be back to deployed, got %+v", after.JSON200)
	}
	if after.JSON200.ImplementationVersion == nil || *after.JSON200.ImplementationVersion != newVersion {
		t.Fatalf("wallet version not updated: %+v", after.JSON200.ImplementationVersion)
	}

	// Admin operation: accepted in flight, then settles asynchronously
	adminOp, err := c.CreateWalletAdminOperationWithResponse(ctx, walletId, apiClient.CreateWalletAdminOperation{
		Capability:    "rotate_signers",
		Arguments:     map[string]interface{}{"signer": "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"},
		Authorization: authorization,
		Operation:     signedOperation,
	})
	if err != nil {
		t.Fatalf("CreateWalletAdminOperation: %v", err)
	}
	if adminOp.StatusCode() != http.StatusAccepted || adminOp.JSON202 == nil {
		t.Fatalf("unexpected admin operation creation response")
	}
	if adminOp.JSON202.Outcome != apiClient.WalletUpgradeOutcomePending {
		t.Fatalf("admin operation should start pending, got %q", adminOp.JSON202.Outcome)
	}

	time.Sleep(150 * time.Millisecond)

	settledOp, err := c.GetWalletAdminOperationWithResponse(ctx, walletId, adminOp.JSON202.AdminOperationId)
	if err != nil {
		t.Fatalf("GetWalletAdminOperation: %v", err)
	}
	if settledOp.StatusCode() != http.StatusOK || settledOp.JSON200 == nil || settledOp.JSON200.Outcome != apiClient.WalletUpgradeOutcomeSucceeded {
		t.Fatalf("admin operation should have succeeded, got %+v", settledOp.JSON200)
	}

	adminOps, err := c.ListWalletAdminOperationsWithResponse(ctx, walletId, nil)
	if err != nil {
		t.Fatalf("ListWalletAdminOperations: %v", err)
	}
	if adminOps.StatusCode() != http.StatusOK || adminOps.JSON200 == nil || len(adminOps.JSON200.Data) != 1 {
		t.Fatalf("unexpected admin operations list response")
	}

	// Not-found cases
	missing, err := c.GetWalletUpgradeWithResponse(ctx, walletId, uuid.New())
	if err != nil {
		t.Fatalf("GetWalletUpgrade missing: %v", err)
	}
	if missing.StatusCode() != http.StatusNotFound {
		t.Fatalf("expected 404 for missing upgrade, got %d", missing.StatusCode())
	}
}
