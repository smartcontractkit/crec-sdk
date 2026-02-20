package server

import (
	"context"
	"net/http"
	"testing"

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
	wallet, err := c.PostWalletsWithResponse(
		context.Background(), apiClient.CreateWallet{
			WalletOwnerAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			ChainSelector:      "1337",
			Name:               testWalletName,
		},
	)
	if err != nil {
		t.Fatalf("PostWallets: %v", err)
	}
	if wallet.StatusCode() != http.StatusCreated || wallet.JSON201 == nil {
		t.Fatalf("unexpected wallet creation response")
	}
	list, err := c.GetWalletsWithResponse(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetWallets: %v", err)
	}
	if list.StatusCode() != http.StatusOK || list.JSON200 == nil {
		t.Fatalf("unexpected wallets list response")
	}
	one, err := c.GetWalletsWalletIdWithResponse(context.Background(), wallet.JSON201.WalletId)
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
	updateResp, err := c.PatchWalletsWalletIdWithResponse(context.Background(), wallet.JSON201.WalletId, updateRequest)
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
