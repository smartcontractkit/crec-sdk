package server

import (
	"context"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/transact/adminop"
	"github.com/smartcontractkit/crec-sdk/transact/eip712"
	"github.com/smartcontractkit/crec-sdk/transact/signer/local"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// testOwnerKey is a throwaway ECDSA key for mock-server upgrade and admin
// operation flows. Never used outside tests.
const testOwnerKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type mockTestClient struct {
	t       *testing.T
	client  *apiClient.ClientWithResponses
	handler *eip712.Handler
	signer  *local.Signer
	address common.Address
}

func newMockTestClient(t *testing.T, s *MockServer) *mockTestClient {
	t.Helper()
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
	key, err := crypto.HexToECDSA(testOwnerKeyHex)
	if err != nil {
		t.Fatalf("owner key: %v", err)
	}
	handler, err := eip712.NewHandler(nil)
	if err != nil {
		t.Fatalf("eip712 handler: %v", err)
	}
	return &mockTestClient{
		t:       t,
		client:  c,
		handler: handler,
		signer:  local.NewSigner(key),
		address: crypto.PubkeyToAddress(key.PublicKey),
	}
}

func (m *mockTestClient) createWallet(t *testing.T, chainSelector string) *apiClient.Wallet {
	t.Helper()
	cfg := apiClient.WalletConfiguration{"allowed_signers": []string{m.address.Hex()}}
	wallet, err := m.client.CreateWalletWithResponse(context.Background(), apiClient.CreateWallet{
		WalletOwnerAddress: m.address.Hex(),
		ChainSelector:      chainSelector,
		Name:               "Upgradeable Wallet " + uuid.NewString()[:8],
		WalletType:         apiClient.WalletTypeECDSA,
		Configuration:      &cfg,
	})
	if err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}
	if wallet.StatusCode() != http.StatusCreated || wallet.JSON201 == nil {
		t.Fatalf("unexpected wallet creation response: %d", wallet.StatusCode())
	}
	return wallet.JSON201
}

// buildUpgradeSubmission signs and packs a valid upgrade request body.
func (m *mockTestClient) buildUpgradeSubmission(
	t *testing.T,
	wallet *apiClient.Wallet,
	targetVersion, implementationAddress string,
	authOpID, operationOpID int64,
) apiClient.CreateWalletUpgrade {
	t.Helper()
	ctx := context.Background()
	deadline := big.NewInt(time.Now().Unix() + 3600)
	walletAddress := common.HexToAddress(wallet.Address)

	auth := &eip712.AdminAuthorization{
		EIP712Type: eip712.UpgradeAuthorizationType,
		Arguments:  map[string]any{"newImplementation": implementationAddress},
		OpID:       big.NewInt(authOpID),
		Deadline:   deadline,
	}
	_, authSig, err := m.handler.SignAdminAuthorization(ctx, auth, m.signer, walletAddress, wallet.ChainSelector)
	if err != nil {
		t.Fatalf("SignAdminAuthorization: %v", err)
	}
	calldata, err := adminop.PackUpgradeToWithAuthorization(
		common.HexToAddress(implementationAddress), auth.OpID, deadline, authSig)
	if err != nil {
		t.Fatalf("PackUpgradeToWithAuthorization: %v", err)
	}
	operation := &transactTypes.Operation{
		ID:       big.NewInt(operationOpID),
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := m.handler.SignOperation(ctx, operation, m.signer, wallet.ChainSelector)
	if err != nil {
		t.Fatalf("SignOperation: %v", err)
	}
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	return apiClient.CreateWalletUpgrade{
		TargetVersion:         targetVersion,
		ImplementationAddress: implementationAddress,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: auth.OpID.String(),
			Deadline:          deadline.Int64(),
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadline.Int64(),
			Signature:         &opSigHex,
			WalletOperationId: big.NewInt(operationOpID).String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	}
}

// buildAdminOpSubmission signs and packs a valid admin operation request body
// for the given capability descriptor.
func (m *mockTestClient) buildAdminOpSubmission(
	t *testing.T,
	wallet *apiClient.Wallet,
	descriptor apiClient.WalletCapabilityDescriptor,
	capability string,
	arguments map[string]any,
	authOpID, operationOpID int64,
) apiClient.CreateWalletAdminOperation {
	t.Helper()
	ctx := context.Background()
	deadline := big.NewInt(time.Now().Unix() + 3600)
	walletAddress := common.HexToAddress(wallet.Address)

	auth := &eip712.AdminAuthorization{
		EIP712Type: descriptor.Eip712Type,
		Arguments:  arguments,
		OpID:       big.NewInt(authOpID),
		Deadline:   deadline,
	}
	_, authSig, err := m.handler.SignAdminAuthorization(ctx, auth, m.signer, walletAddress, wallet.ChainSelector)
	if err != nil {
		t.Fatalf("SignAdminAuthorization: %v", err)
	}
	calldata, err := adminop.PackCapabilityCall(descriptor.AbiFragment, arguments, auth.OpID, deadline, authSig)
	if err != nil {
		t.Fatalf("PackCapabilityCall: %v", err)
	}
	operation := &transactTypes.Operation{
		ID:       big.NewInt(operationOpID),
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := m.handler.SignOperation(ctx, operation, m.signer, wallet.ChainSelector)
	if err != nil {
		t.Fatalf("SignOperation: %v", err)
	}
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	return apiClient.CreateWalletAdminOperation{
		Capability: capability,
		Arguments:  arguments,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: auth.OpID.String(),
			Deadline:          deadline.Int64(),
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadline.Int64(),
			Signature:         &opSigHex,
			WalletOperationId: big.NewInt(operationOpID).String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	}
}

// awaitMockUpgrade polls the upgrade until its outcome is terminal,
// synchronizing on the public read instead of sleeping a fixed duration.
func awaitMockUpgrade(t *testing.T, c *apiClient.ClientWithResponses, walletId, upgradeId uuid.UUID) *apiClient.WalletUpgrade {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := c.GetWalletUpgradeWithResponse(context.Background(), walletId, upgradeId)
		if err != nil {
			t.Fatalf("GetWalletUpgrade: %v", err)
		}
		if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
			resp.JSON200.Outcome != apiClient.WalletUpgradeOutcomePending {
			return resp.JSON200
		}
		if time.Now().After(deadline) {
			t.Fatalf("upgrade %s did not reach a terminal outcome", upgradeId)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// awaitMockAdminOperation polls the administrative operation until its outcome
// is terminal.
func awaitMockAdminOperation(t *testing.T, c *apiClient.ClientWithResponses, walletId, adminOperationId uuid.UUID) *apiClient.WalletAdminOperation {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := c.GetWalletAdminOperationWithResponse(context.Background(), walletId, adminOperationId)
		if err != nil {
			t.Fatalf("GetWalletAdminOperation: %v", err)
		}
		if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
			resp.JSON200.Outcome != apiClient.WalletUpgradeOutcomePending {
			return resp.JSON200
		}
		if time.Now().After(deadline) {
			t.Fatalf("admin operation %s did not reach a terminal outcome", adminOperationId)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestMockServer_WalletVersions_Upgrades_AdminOperations(t *testing.T) {
	s := NewMockServer()
	defer s.Close()
	m := newMockTestClient(t, s)
	ctx := context.Background()

	// Wallet type catalog: two seeded versions per type
	versions, err := m.client.ListWalletTypeVersionsWithResponse(ctx, apiClient.WalletTypeECDSA, nil)
	if err != nil {
		t.Fatalf("ListWalletTypeVersions: %v", err)
	}
	if versions.StatusCode() != http.StatusOK || versions.JSON200 == nil || len(versions.JSON200.Data) != 2 {
		t.Fatalf("unexpected wallet type versions response: %+v", versions.JSON200)
	}

	version, err := m.client.GetWalletTypeVersionWithResponse(ctx, apiClient.WalletTypeECDSA, mockDefaultWalletTypeVersion)
	if err != nil {
		t.Fatalf("GetWalletTypeVersion: %v", err)
	}
	if version.StatusCode() != http.StatusOK || version.JSON200 == nil {
		t.Fatalf("unexpected wallet type version response")
	}
	var sawDefault bool
	for _, d := range version.JSON200.Deployments {
		if d.ChainSelector == mockDefaultChainSelector && d.IsDefault {
			sawDefault = true
		}
	}
	if !sawDefault {
		t.Fatalf("default version missing its default deployment on %s: %+v", mockDefaultChainSelector, version.JSON200.Deployments)
	}

	// Defaults are per network: a chain with no default deployment yields no version
	noDefault, err := m.client.CreateWalletWithResponse(ctx, apiClient.CreateWallet{
		WalletOwnerAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		ChainSelector:      "1338",
		Name:               "No Default Wallet",
		WalletType:         apiClient.WalletTypeECDSA,
		Configuration:      &apiClient.WalletConfiguration{},
	})
	if err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}
	if noDefault.StatusCode() != http.StatusCreated || noDefault.JSON201 == nil || noDefault.JSON201.ImplementationVersion != nil {
		t.Fatalf("wallet on a chain without a default got a version: %+v", noDefault.JSON201)
	}

	// Wallet creation picks up the type's default implementation version on its
	// network. The mock's sepolia deployment lets the signing path resolve a
	// real chain ID.
	wallet := m.createWallet(t, mockSepoliaChainSelector)
	if wallet.ImplementationVersion == nil || *wallet.ImplementationVersion != mockDefaultWalletTypeVersion {
		t.Fatalf("wallet not created on the default version: %+v", wallet.ImplementationVersion)
	}
	walletId := wallet.WalletId

	// Upgrade: signed, packed and accepted in flight, then settles asynchronously
	upgradeReq := m.buildUpgradeSubmission(t, wallet, mockNextWalletTypeVersion, mockNextImplementationAddress, 2, 1)
	upgrade, err := m.client.CreateWalletUpgradeWithResponse(ctx, walletId, upgradeReq)
	if err != nil {
		t.Fatalf("CreateWalletUpgrade: %v", err)
	}
	if upgrade.StatusCode() != http.StatusAccepted || upgrade.JSON202 == nil {
		t.Fatalf("unexpected upgrade creation response: %d body %s", upgrade.StatusCode(), string(upgrade.Body))
	}
	if upgrade.JSON202.Outcome != apiClient.WalletUpgradeOutcomePending {
		t.Fatalf("upgrade should start pending, got %q", upgrade.JSON202.Outcome)
	}
	if upgrade.JSON202.FromVersion == nil || *upgrade.JSON202.FromVersion != mockDefaultWalletTypeVersion {
		t.Fatalf("unexpected upgrade from_version: %+v", upgrade.JSON202.FromVersion)
	}

	// A second upgrade while one is in flight conflicts
	racing, err := m.client.CreateWalletUpgradeWithResponse(ctx, walletId,
		m.buildUpgradeSubmission(t, wallet, mockNextWalletTypeVersion, mockNextImplementationAddress, 4, 3))
	if err != nil {
		t.Fatalf("CreateWalletUpgrade racing: %v", err)
	}
	if racing.StatusCode() != http.StatusConflict || racing.JSON409 == nil ||
		racing.JSON409.Code == nil || *racing.JSON409.Code != apiClient.ApplicationErrorCodeWalletUpgradeInProgress {
		t.Fatalf("expected 409 upgrade in progress, got %d %+v", racing.StatusCode(), racing.JSON409)
	}

	upgrades, err := m.client.ListWalletUpgradesWithResponse(ctx, walletId, nil)
	if err != nil {
		t.Fatalf("ListWalletUpgrades: %v", err)
	}
	if upgrades.StatusCode() != http.StatusOK || upgrades.JSON200 == nil || len(upgrades.JSON200.Data) != 1 {
		t.Fatalf("unexpected upgrades list response")
	}

	settled := awaitMockUpgrade(t, m.client, walletId, upgrade.JSON202.UpgradeId)
	if settled.Outcome != apiClient.WalletUpgradeOutcomeSucceeded {
		t.Fatalf("upgrade should have succeeded, got %+v", settled)
	}
	// The wallet moved onto the target version
	after, err := m.client.GetWalletWithResponse(ctx, walletId)
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}
	if after.JSON200 == nil || after.JSON200.Status != apiClient.WalletStatusDeployed {
		t.Fatalf("wallet should be back to deployed, got %+v", after.JSON200)
	}
	if after.JSON200.ImplementationVersion == nil || *after.JSON200.ImplementationVersion != mockNextWalletTypeVersion {
		t.Fatalf("wallet version not updated: %+v", after.JSON200.ImplementationVersion)
	}

	// Admin operation against a capability the new version publishes
	upgraded, err := m.client.GetWalletTypeVersionWithResponse(ctx, apiClient.WalletTypeECDSA, mockNextWalletTypeVersion)
	if err != nil || upgraded.JSON200 == nil {
		t.Fatalf("GetWalletTypeVersion upgraded: %v", err)
	}
	var setSignerAllowed *apiClient.WalletCapabilityDescriptor
	for i := range upgraded.JSON200.Capabilities {
		if upgraded.JSON200.Capabilities[i].Name == "setSignerAllowed" {
			setSignerAllowed = &upgraded.JSON200.Capabilities[i]
		}
	}
	if setSignerAllowed == nil {
		t.Fatalf("upgraded version publishes no setSignerAllowed capability")
	}

	arguments := map[string]any{
		"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
		"allowed": true,
	}
	adminOp, err := m.client.CreateWalletAdminOperationWithResponse(ctx, walletId,
		m.buildAdminOpSubmission(t, wallet, *setSignerAllowed, "setSignerAllowed", arguments, 6, 5))
	if err != nil {
		t.Fatalf("CreateWalletAdminOperation: %v", err)
	}
	if adminOp.StatusCode() != http.StatusAccepted || adminOp.JSON202 == nil {
		t.Fatalf("unexpected admin operation creation response: %d body %s", adminOp.StatusCode(), string(adminOp.Body))
	}
	if adminOp.JSON202.Outcome != apiClient.WalletUpgradeOutcomePending {
		t.Fatalf("admin operation should start pending, got %q", adminOp.JSON202.Outcome)
	}

	settledOp := awaitMockAdminOperation(t, m.client, walletId, adminOp.JSON202.AdminOperationId)
	if settledOp.Outcome != apiClient.WalletUpgradeOutcomeSucceeded {
		t.Fatalf("admin operation should have succeeded, got %+v", settledOp)
	}

	adminOps, err := m.client.ListWalletAdminOperationsWithResponse(ctx, walletId, nil)
	if err != nil {
		t.Fatalf("ListWalletAdminOperations: %v", err)
	}
	if adminOps.StatusCode() != http.StatusOK || adminOps.JSON200 == nil || len(adminOps.JSON200.Data) != 1 {
		t.Fatalf("unexpected admin operations list response")
	}
}

func TestMockServer_UpgradeRejections(t *testing.T) {
	s := NewMockServer()
	defer s.Close()
	m := newMockTestClient(t, s)
	ctx := context.Background()

	wallet := m.createWallet(t, mockSepoliaChainSelector)
	walletId := wallet.WalletId

	cases := []struct {
		name       string
		mutate     func(req *apiClient.CreateWalletUpgrade)
		wantStatus int
		wantCode   apiClient.ApplicationErrorCode
	}{
		{
			name:       "UnknownTargetVersion",
			mutate:     func(req *apiClient.CreateWalletUpgrade) { req.TargetVersion = "2999.01.0" },
			wantStatus: http.StatusNotFound,
			wantCode:   apiClient.ApplicationErrorCodeWalletTypeVersionNotFound,
		},
		{
			name: "TargetAlreadyRecorded",
			mutate: func(req *apiClient.CreateWalletUpgrade) {
				req.TargetVersion = mockDefaultWalletTypeVersion
				req.ImplementationAddress = mockDefaultImplementationAddress
			},
			wantStatus: http.StatusConflict,
			wantCode:   apiClient.ApplicationErrorCodeWalletUpgradeTargetAlreadyRecorded,
		},
		{
			name: "PayloadMismatchAddress",
			mutate: func(req *apiClient.CreateWalletUpgrade) {
				req.ImplementationAddress = "0x0000000000000000000000000000000000000bad"
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   apiClient.ApplicationErrorCodeWalletUpgradePayloadMismatch,
		},
		{
			name: "PayloadMismatchCalldata",
			mutate: func(req *apiClient.CreateWalletUpgrade) {
				req.Operation.Transactions[0].Data = "0xdeadbeef"
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   apiClient.ApplicationErrorCodeWalletUpgradePayloadMismatch,
		},
		{
			name: "TwoTransactions",
			mutate: func(req *apiClient.CreateWalletUpgrade) {
				req.Operation.Transactions = append(req.Operation.Transactions, req.Operation.Transactions[0])
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "",
		},
		{
			name: "CollidingOperationIDs",
			mutate: func(req *apiClient.CreateWalletUpgrade) {
				req.Authorization.WalletOperationId = req.Operation.WalletOperationId
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := m.buildUpgradeSubmission(t, wallet, mockNextWalletTypeVersion, mockNextImplementationAddress,
				time.Now().UnixNano()+100, time.Now().UnixNano()+200)
			tc.mutate(&req)
			resp, err := m.client.CreateWalletUpgradeWithResponse(ctx, walletId, req)
			if err != nil {
				t.Fatalf("CreateWalletUpgrade: %v", err)
			}
			if resp.StatusCode() != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, resp.StatusCode(), string(resp.Body))
			}
			if tc.wantCode != "" {
				var appErr *apiClient.ApplicationError
				switch tc.wantStatus {
				case http.StatusBadRequest:
					appErr = resp.JSON400
				case http.StatusConflict:
					appErr = resp.JSON409
				case http.StatusNotFound:
					appErr = resp.JSON404
				}
				if appErr == nil || appErr.Code == nil || *appErr.Code != tc.wantCode {
					t.Fatalf("expected code %s, got %+v", tc.wantCode, appErr)
				}
			}
		})
	}

	t.Run("WalletNotFound", func(t *testing.T) {
		req := m.buildUpgradeSubmission(t, wallet, mockNextWalletTypeVersion, mockNextImplementationAddress, 1002, 1001)
		resp, err := m.client.CreateWalletUpgradeWithResponse(ctx, uuid.New(), req)
		if err != nil {
			t.Fatalf("CreateWalletUpgrade: %v", err)
		}
		if resp.StatusCode() != http.StatusNotFound || resp.JSON404 == nil ||
			resp.JSON404.Code == nil || *resp.JSON404.Code != apiClient.ApplicationErrorCodeWalletNotFound {
			t.Fatalf("expected 404 wallet not found, got %d %+v", resp.StatusCode(), resp.JSON404)
		}
	})

	t.Run("UpgradeNotFound", func(t *testing.T) {
		resp, err := m.client.GetWalletUpgradeWithResponse(ctx, walletId, uuid.New())
		if err != nil {
			t.Fatalf("GetWalletUpgrade: %v", err)
		}
		if resp.StatusCode() != http.StatusNotFound || resp.JSON404 == nil ||
			resp.JSON404.Code == nil || *resp.JSON404.Code != apiClient.ApplicationErrorCodeWalletUpgradeNotFound {
			t.Fatalf("expected 404 upgrade not found, got %d %+v", resp.StatusCode(), resp.JSON404)
		}
	})
}

func TestMockServer_AdminOperationRejections(t *testing.T) {
	s := NewMockServer()
	defer s.Close()
	m := newMockTestClient(t, s)
	ctx := context.Background()

	// A wallet on the default version publishes no capabilities.
	wallet := m.createWallet(t, mockSepoliaChainSelector)
	walletId := wallet.WalletId

	// Upgrade it to the version that publishes capabilities.
	upgradeReq := m.buildUpgradeSubmission(t, wallet, mockNextWalletTypeVersion, mockNextImplementationAddress, 12, 11)
	upgrade, err := m.client.CreateWalletUpgradeWithResponse(ctx, walletId, upgradeReq)
	if err != nil {
		t.Fatalf("CreateWalletUpgrade: %v", err)
	}
	if upgrade.StatusCode() != http.StatusAccepted {
		t.Fatalf("upgrade rejected: %s", string(upgrade.Body))
	}
	awaitMockUpgrade(t, m.client, walletId, upgrade.JSON202.UpgradeId)

	upgraded, err := m.client.GetWalletTypeVersionWithResponse(ctx, apiClient.WalletTypeECDSA, mockNextWalletTypeVersion)
	if err != nil || upgraded.JSON200 == nil {
		t.Fatalf("GetWalletTypeVersion: %v", err)
	}
	var descriptor apiClient.WalletCapabilityDescriptor
	for _, c := range upgraded.JSON200.Capabilities {
		if c.Name == "setSignerAllowed" {
			descriptor = c
		}
	}
	validArgs := func() map[string]any {
		return map[string]any{"signer": "0x9f644089295bfa1402231508116aaac15c3d7c87", "allowed": true}
	}

	t.Run("CapabilityNotPublished", func(t *testing.T) {
		// attachPolicyEngine is a protected-wallet capability; the ecdsa
		// version does not publish it. Build with its true descriptor shape so
		// the submission itself is well formed.
		attachFragment := map[string]any{
			"type": "function",
			"name": "attachPolicyEngineWithAuthorization",
			"inputs": []any{
				map[string]any{"name": "policyEngine", "type": "address"},
				map[string]any{"name": "opId", "type": "uint256"},
				map[string]any{"name": "deadline", "type": "uint256"},
				map[string]any{"name": "signature", "type": "bytes"},
			},
			"outputs":         []any{},
			"stateMutability": "nonpayable",
		}
		attachDescriptor := apiClient.WalletCapabilityDescriptor{
			Name:        "attachPolicyEngine",
			AbiFragment: attachFragment,
			Eip712Type:  "AttachPolicyEngineAuthorization(address policyEngine,uint256 opId,uint256 deadline)",
		}
		req := m.buildAdminOpSubmission(t, wallet, attachDescriptor, "attachPolicyEngine", map[string]any{
			"policyEngine": "0x8c4f38de2a76735e4ae96a05a980ab46000f1c09",
		}, 14, 13)
		resp, err := m.client.CreateWalletAdminOperationWithResponse(ctx, walletId, req)
		if err != nil {
			t.Fatalf("CreateWalletAdminOperation: %v", err)
		}
		if resp.StatusCode() != http.StatusConflict || resp.JSON409 == nil ||
			resp.JSON409.Code == nil || *resp.JSON409.Code != apiClient.ApplicationErrorCodeWalletCapabilityNotSupported {
			t.Fatalf("expected 409 capability not supported, got %d %+v", resp.StatusCode(), resp.JSON409)
		}
	})

	t.Run("ArgumentMismatch", func(t *testing.T) {
		req := m.buildAdminOpSubmission(t, wallet, descriptor, "setSignerAllowed", validArgs(), 16, 15)
		// The submitted arguments no longer name exactly the capability's
		// parameters: the packed call could not be the one the owner signed.
		delete(req.Arguments, "allowed")
		resp, err := m.client.CreateWalletAdminOperationWithResponse(ctx, walletId, req)
		if err != nil {
			t.Fatalf("CreateWalletAdminOperation: %v", err)
		}
		if resp.StatusCode() != http.StatusBadRequest || resp.JSON400 == nil ||
			resp.JSON400.Code == nil || *resp.JSON400.Code != apiClient.ApplicationErrorCodeWalletAdminOperationPayloadMismatch {
			t.Fatalf("expected 400 payload mismatch, got %d %+v", resp.StatusCode(), resp.JSON400)
		}
	})

	t.Run("CalldataMismatch", func(t *testing.T) {
		req := m.buildAdminOpSubmission(t, wallet, descriptor, "setSignerAllowed", validArgs(), 18, 17)
		req.Operation.Transactions[0].Data = "0xdeadbeef"
		resp, err := m.client.CreateWalletAdminOperationWithResponse(ctx, walletId, req)
		if err != nil {
			t.Fatalf("CreateWalletAdminOperation: %v", err)
		}
		if resp.StatusCode() != http.StatusBadRequest || resp.JSON400 == nil ||
			resp.JSON400.Code == nil || *resp.JSON400.Code != apiClient.ApplicationErrorCodeWalletAdminOperationPayloadMismatch {
			t.Fatalf("expected 400 payload mismatch, got %d %+v", resp.StatusCode(), resp.JSON400)
		}
	})

	t.Run("AdminOperationNotFound", func(t *testing.T) {
		resp, err := m.client.GetWalletAdminOperationWithResponse(ctx, walletId, uuid.New())
		if err != nil {
			t.Fatalf("GetWalletAdminOperation: %v", err)
		}
		if resp.StatusCode() != http.StatusNotFound || resp.JSON404 == nil ||
			resp.JSON404.Code == nil || *resp.JSON404.Code != apiClient.ApplicationErrorCodeWalletAdminOperationNotFound {
			t.Fatalf("expected 404 admin operation not found, got %d %+v", resp.StatusCode(), resp.JSON404)
		}
	})
}
