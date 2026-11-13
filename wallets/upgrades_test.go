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

func TestClient_CreateUpgrade(t *testing.T) {
	t.Run("SuccessPendingThenAwaited", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)

		input := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 102, 101)
		upgrade, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
		require.NoError(t, err)
		require.NotNil(t, upgrade)
		assert.Equal(t, apiClient.WalletUpgradeOutcomePending, upgrade.Outcome)
		assert.Equal(t, "2025.09.0", upgrade.ToVersion)
		require.NotNil(t, upgrade.FromVersion)
		assert.Equal(t, "2025.06.0", *upgrade.FromVersion)
		assert.Equal(t, apiClient.WalletVersionChangeSourceCRECUpgrade, upgrade.Source)

		settled, err := env.client.AwaitUpgrade(context.Background(), wallet.WalletId, upgrade.UpgradeId,
			10*time.Millisecond, 5*time.Second)
		require.NoError(t, err)
		assert.Equal(t, apiClient.WalletUpgradeOutcomeSucceeded, settled.Outcome)

		// The wallet is recorded on the target version.
		after, err := env.client.Get(context.Background(), wallet.WalletId)
		require.NoError(t, err)
		require.NotNil(t, after.ImplementationVersion)
		assert.Equal(t, "2025.09.0", *after.ImplementationVersion)
	})

	t.Run("NilWalletID", func(t *testing.T) {
		env := setupMockEnv(t)
		_, err := env.client.CreateUpgrade(context.Background(), uuid.Nil, CreateUpgradeInput{})
		require.ErrorIs(t, err, ErrWalletIDRequired)
	})

	t.Run("EmptyTargetVersion", func(t *testing.T) {
		env := setupMockEnv(t)
		_, err := env.client.CreateUpgrade(context.Background(), uuid.New(), CreateUpgradeInput{})
		require.ErrorIs(t, err, ErrTargetVersionRequired)
	})

	t.Run("WalletNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		input := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 202, 201)

		_, err := env.client.CreateUpgrade(context.Background(), uuid.New(), input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletNotFound)
	})

	t.Run("TargetVersionNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		input := env.buildUpgradeInput(t, wallet, "1999.01.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 204, 203)

		_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletTypeVersionNotFound)
	})

	t.Run("TargetAlreadyRecorded", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		input := env.buildUpgradeInput(t, wallet, "2025.06.0", "0x5fbdb2315678afecb367f032d93f642f64180aa3", 206, 205)

		_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletUpgradeTargetAlreadyRecorded)
	})

	t.Run("UpgradeInProgress", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)

		first := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 208, 207)
		_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, first)
		require.NoError(t, err)

		second := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 210, 209)
		_, err = env.client.CreateUpgrade(context.Background(), wallet.WalletId, second)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletUpgradeInProgress)
	})

	t.Run("PayloadMismatch", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		input := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x0000000000000000000000000000000000000bad", 212, 211)

		_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletUpgradePayloadMismatch)
	})

	t.Run("ClientSideRelayConstraints", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		base := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 214, 213)

		t.Run("CollidingOperationIDs", func(t *testing.T) {
			input := base
			input.Authorization.WalletOperationId = base.Operation.WalletOperationId
			_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
			require.ErrorIs(t, err, ErrOperationIDCollision)
		})

		t.Run("DeadlineMismatch", func(t *testing.T) {
			input := base
			input.Operation.Deadline = base.Authorization.Deadline + 1
			_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
			require.ErrorIs(t, err, ErrOperationDeadlineMismatch)
		})

		t.Run("ZeroAuthorizationDeadline", func(t *testing.T) {
			input := base
			input.Authorization.Deadline = 0
			_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
			require.ErrorIs(t, err, ErrAuthorizationDeadline)
		})

		t.Run("NoTransactions", func(t *testing.T) {
			input := base
			input.Operation.Transactions = nil
			_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
			require.ErrorIs(t, err, ErrSingleTransactionRequired)
		})
	})
}

func TestClient_ListUpgrades(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		input := env.buildUpgradeInput(t, wallet, "2025.09.0", "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", 302, 301)
		_, err := env.client.CreateUpgrade(context.Background(), wallet.WalletId, input)
		require.NoError(t, err)

		upgrades, hasMore, err := env.client.ListUpgrades(context.Background(), wallet.WalletId, ListUpgradesInput{})
		require.NoError(t, err)
		assert.False(t, hasMore)
		require.Len(t, upgrades, 1)
		assert.Equal(t, "2025.09.0", upgrades[0].ToVersion)

		source := apiClient.WalletVersionChangeSourceCRECUpgrade
		filtered, _, err := env.client.ListUpgrades(context.Background(), wallet.WalletId, ListUpgradesInput{Source: &source})
		require.NoError(t, err)
		assert.Len(t, filtered, 1)

		correction := apiClient.WalletVersionChangeSourceOperatorCorrection
		filtered, _, err = env.client.ListUpgrades(context.Background(), wallet.WalletId, ListUpgradesInput{Source: &correction})
		require.NoError(t, err)
		assert.Empty(t, filtered)
	})

	t.Run("WalletNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		_, _, err := env.client.ListUpgrades(context.Background(), uuid.New(), ListUpgradesInput{})
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletNotFound)
	})

	t.Run("InvalidLimit", func(t *testing.T) {
		env := setupMockEnv(t)
		zero := 0
		_, _, err := env.client.ListUpgrades(context.Background(), uuid.New(), ListUpgradesInput{Limit: &zero})
		require.ErrorIs(t, err, ErrInvalidLimit)
	})
}

func TestClient_GetUpgrade(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)

		_, err := env.client.GetUpgrade(context.Background(), wallet.WalletId, uuid.New())
		require.Error(t, err)
		assert.ErrorIs(t, err, apierror.ErrWalletUpgradeNotFound)
	})

	t.Run("NilUpgradeID", func(t *testing.T) {
		env := setupMockEnv(t)
		_, err := env.client.GetUpgrade(context.Background(), uuid.New(), uuid.Nil)
		require.ErrorIs(t, err, ErrUpgradeIDRequired)
	})
}

func TestClient_ResolveUpgradeTarget_And_UpgradeWallet(t *testing.T) {
	t.Run("EndToEnd", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		ctx := context.Background()

		// Discover and inspect: label beside the resolved address.
		target, err := env.client.ResolveUpgradeTarget(ctx, wallet.WalletId, "2025.09.0")
		require.NoError(t, err)
		assert.Equal(t, "2025.09.0", target.Version)
		assert.Equal(t, "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0", target.ImplementationAddress)
		assert.Equal(t, "2025.06.0", target.CurrentVersion)
		assert.False(t, target.IsDefaultOnNetwork)

		// Sign, submit, and poll to a terminal outcome.
		upgrade, err := env.client.UpgradeWallet(ctx, wallet.WalletId, target, UpgradeOptions{
			OwnerSigner:  env.signer,
			Deadline:     time.Now().Add(time.Hour),
			PollInterval: 10 * time.Millisecond,
			PollTimeout:  5 * time.Second,
		})
		require.NoError(t, err)
		require.NotNil(t, upgrade)
		assert.Equal(t, apiClient.WalletUpgradeOutcomeSucceeded, upgrade.Outcome)

		after, err := env.client.Get(ctx, wallet.WalletId)
		require.NoError(t, err)
		require.NotNil(t, after.ImplementationVersion)
		assert.Equal(t, "2025.09.0", *after.ImplementationVersion)
	})

	t.Run("NoDeploymentOnNetwork", func(t *testing.T) {
		env := setupMockEnv(t)
		// A wallet on a chain the mock catalog does not deploy 2025.09.0 to.
		wallet, err := env.client.Create(context.Background(), CreateInput{
			Name:               "other-chain-" + uuid8(),
			ChainSelector:      "1337",
			WalletOwnerAddress: env.address.Hex(),
			WalletType:         apiClient.WalletTypeECDSA,
			Configuration:      apiClient.WalletConfiguration{},
		})
		require.NoError(t, err)
		// 2025.09.0 is deployed on 1337 too; ask for a version that is not on
		// this chain by upgrading on a chain with no deployments at all.
		other, err := env.client.Create(context.Background(), CreateInput{
			Name:               "no-deploy-" + uuid8(),
			ChainSelector:      "1338",
			WalletOwnerAddress: env.address.Hex(),
			WalletType:         apiClient.WalletTypeECDSA,
			Configuration:      apiClient.WalletConfiguration{},
		})
		require.NoError(t, err)
		_ = wallet

		_, err = env.client.ResolveUpgradeTarget(context.Background(), other.WalletId, "2025.09.0")
		require.ErrorIs(t, err, ErrNoDeploymentOnNetwork)
	})

	t.Run("TargetVersionNotFound", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		_, err := env.client.ResolveUpgradeTarget(context.Background(), wallet.WalletId, "1999.01.0")
		require.ErrorIs(t, err, apierror.ErrWalletTypeVersionNotFound)
	})

	t.Run("UpgradeWalletValidation", func(t *testing.T) {
		env := setupMockEnv(t)
		wallet := env.createWallet(t)
		target, err := env.client.ResolveUpgradeTarget(context.Background(), wallet.WalletId, "2025.09.0")
		require.NoError(t, err)

		_, err = env.client.UpgradeWallet(context.Background(), wallet.WalletId, nil, UpgradeOptions{
			OwnerSigner: env.signer,
			Deadline:    time.Now().Add(time.Hour),
		})
		require.ErrorIs(t, err, ErrUpgradeTargetRequired)

		_, err = env.client.UpgradeWallet(context.Background(), wallet.WalletId, target, UpgradeOptions{
			Deadline: time.Now().Add(time.Hour),
		})
		require.ErrorIs(t, err, ErrOwnerSignerRequired)

		_, err = env.client.UpgradeWallet(context.Background(), wallet.WalletId, target, UpgradeOptions{
			OwnerSigner: env.signer,
		})
		require.ErrorIs(t, err, ErrDeadlineRequired)
	})
}
