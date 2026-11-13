package wallets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

// upgradeToLatest upgrades the test wallet to the capability-publishing
// version, the precondition for administrative operations.
func upgradeToLatest(t *testing.T, env *mockTestEnv, wallet *apiClient.Wallet) {
	t.Helper()
	ctx := context.Background()
	target, err := env.client.ResolveUpgradeTarget(ctx, wallet.WalletId, "2025.09.0")
	require.NoError(t, err)
	_, err = env.client.UpgradeWallet(ctx, wallet.WalletId, target, UpgradeOptions{
		OwnerSigner:  env.signer,
		Deadline:     time.Now().Add(time.Hour),
		PollInterval: 10 * time.Millisecond,
		PollTimeout:  5 * time.Second,
	})
	require.NoError(t, err)
}

func TestClient_CreateAdminOperation(t *testing.T) {
	t.Run("SuccessPendingThenAwaited", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		descriptor := setSignerAllowedDescriptor(t, env.client)
		arguments := map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}
		input := env.buildAdminOperationInput(t, wallet, descriptor, "setSignerAllowed", arguments, 402, 401)
		adminOp, err := env.client.CreateAdminOperation(context.Background(), wallet.WalletId, input)
		require.NoError(t, err)
		require.NotNil(t, adminOp)
		assert.Equal(t, apiClient.WalletUpgradeOutcomePending, adminOp.Outcome)
		assert.Equal(t, "setSignerAllowed", adminOp.Capability)
		assert.Equal(t, arguments["signer"], adminOp.Arguments["signer"])

		settled, err := env.client.AwaitAdminOperation(context.Background(), wallet.WalletId, adminOp.AdminOperationId,
			10*time.Millisecond, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WalletUpgradeOutcomeSucceeded, settled.Outcome)
	})

	t.Run("CapabilityNotPublished", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		descriptor := setSignerAllowedDescriptor(t, env.client)
		arguments := map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}
		input := env.buildAdminOperationInput(t, wallet, descriptor, "attachPolicyEngine", arguments, 404, 403)

		_, err := env.client.CreateAdminOperation(context.Background(), wallet.WalletId, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletCapabilityNotSupported)
	})

	t.Run("PayloadMismatchArguments", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		descriptor := setSignerAllowedDescriptor(t, env.client)
		arguments := map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}
		input := env.buildAdminOperationInput(t, wallet, descriptor, "setSignerAllowed", arguments, 406, 405)
		delete(input.Arguments, "allowed")

		_, err := env.client.CreateAdminOperation(context.Background(), wallet.WalletId, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletAdminOperationPayloadMismatch)
	})

	t.Run("WalletNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		descriptor := setSignerAllowedDescriptor(t, env.client)
		arguments := map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": true,
		}
		input := env.buildAdminOperationInput(t, wallet, descriptor, "setSignerAllowed", arguments, 408, 407)

		_, err := env.client.CreateAdminOperation(context.Background(), uuid.New(), input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletNotFound)
	})

	t.Run("EmptyCapability", func(t *testing.T) {
		env := setupMockEnv(t)
		_, err := env.client.CreateAdminOperation(context.Background(), uuid.Nil, CreateAdminOperationInput{})
		require.ErrorIs(t, err, ErrWalletIDRequired)

		_, err = env.client.CreateAdminOperation(context.Background(), uuid.New(), CreateAdminOperationInput{})
		require.ErrorIs(t, err, ErrCapabilityRequired)
	})
}

func TestClient_GetAdminOperation(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)

		_, err := env.client.GetAdminOperation(context.Background(), wallet.WalletId, uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletAdminOperationNotFound)
	})

	t.Run("NilIDs", func(t *testing.T) {
		env := setupMockEnv(t)
		_, err := env.client.GetAdminOperation(context.Background(), uuid.Nil, uuid.New())
		require.ErrorIs(t, err, ErrWalletIDRequired)
		_, err = env.client.GetAdminOperation(context.Background(), uuid.New(), uuid.Nil)
		require.ErrorIs(t, err, ErrAdminOperationIDRequired)
	})
}

func TestClient_ListAdminOperations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		descriptor := setSignerAllowedDescriptor(t, env.client)
		input := env.buildAdminOperationInput(t, wallet, descriptor, "setSignerAllowed", map[string]any{
			"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
			"allowed": false,
		}, 410, 409)
		_, err := env.client.CreateAdminOperation(context.Background(), wallet.WalletId, input)
		require.NoError(t, err)

		adminOps, hasMore, err := env.client.ListAdminOperations(context.Background(), wallet.WalletId, ListAdminOperationsInput{})
		require.NoError(t, err)
		assert.False(t, hasMore)
		require.Len(t, adminOps, 1)
		assert.Equal(t, "setSignerAllowed", adminOps[0].Capability)
	})

	t.Run("WalletNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		_, _, err := env.client.ListAdminOperations(context.Background(), uuid.New(), ListAdminOperationsInput{})
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletNotFound)
	})

	t.Run("InvalidOffset", func(t *testing.T) {
		env := setupMockEnv(t)
		negative := int64(-1)
		_, _, err := env.client.ListAdminOperations(context.Background(), uuid.New(), ListAdminOperationsInput{Offset: &negative})
		require.ErrorIs(t, err, ErrInvalidOffset)
	})
}

func TestClient_SubmitAdminOperation(t *testing.T) {
	t.Run("EndToEnd", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		upgradeToLatest(t, env, wallet)

		adminOp, err := env.client.SubmitAdminOperation(context.Background(), wallet.WalletId, "setSignerAllowed",
			map[string]any{
				"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
				"allowed": true,
			},
			AdminOperationOptions{
				OwnerSigner:  env.signer,
				Deadline:     time.Now().Add(time.Hour),
				PollInterval: 10 * time.Millisecond,
				PollTimeout:  5 * time.Second,
			})
		require.NoError(t, err)
		require.NotNil(t, adminOp)
		assert.Equal(t, apiClient.WalletUpgradeOutcomeSucceeded, adminOp.Outcome)
		assert.Equal(t, "setSignerAllowed", adminOp.Capability)
	})

	t.Run("CapabilityNotPublishedClientSide", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)

		// The wallet's recorded 2025.06.0 publishes no capabilities.
		_, err := env.client.SubmitAdminOperation(context.Background(), wallet.WalletId, "setSignerAllowed",
			map[string]any{
				"signer":  "0x9f644089295bfa1402231508116aaac15c3d7c87",
				"allowed": true,
			},
			AdminOperationOptions{
				OwnerSigner: env.signer,
				Deadline:    time.Now().Add(time.Hour),
			})
		require.ErrorIs(t, err, ErrCapabilityNotPublished)
	})
}
