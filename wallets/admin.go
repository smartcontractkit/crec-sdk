package wallets

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/transact/adminop"
	"github.com/smartcontractkit/crec-sdk/transact/eip712"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// High-level administrative flow errors.
var (
	ErrOwnerSignerRequired   = errors.New("owner signer is required")
	ErrUpgradeTargetRequired = errors.New("upgrade target is required")
	ErrTargetTypeMismatch    = errors.New("upgrade target belongs to a different wallet type")
	ErrWalletAddressMissing  = errors.New("wallet has no address")
	ErrDeadlineRequired      = errors.New("deadline is required and must be in the future")
	ErrUpgradeFailed         = errors.New("wallet upgrade failed")
	ErrUpgradeExpired        = errors.New("wallet upgrade expired")
	ErrAdminOperationFailed  = errors.New("wallet admin operation failed")
	ErrAdminOperationExpired = errors.New("wallet admin operation expired")
	ErrNoVersionRecorded     = errors.New("wallet has no recorded implementation version")
	ErrAwaitTimeout          = errors.New("timed out waiting for a terminal outcome")
)

// UpgradeTarget is the resolved, inspectable fact an upgrade commits to: the
// version label beside the implementation address the owner signs. Build it
// with [Client.ResolveUpgradeTarget], inspect it, then pass it to
// [Client.UpgradeWallet]; what the owner inspects is what they sign.
type UpgradeTarget struct {
	// Version is the target implementation version label.
	Version string
	// ImplementationAddress is the version's implementation address on the
	// wallet's network, the value the authorization commits to.
	ImplementationAddress string
	// CurrentVersion is the version the wallet is recorded on, empty where
	// none was recorded.
	CurrentVersion string
	// IsDefaultOnNetwork reports whether the target is the wallet type's
	// default on the wallet's network.
	IsDefaultOnNetwork bool
}

// ResolveUpgradeTarget resolves the catalog facts for upgrading a wallet to
// targetVersion: the implementation address on the wallet's network, whether
// it is the default there, and the version the wallet is recorded on. Render
// the version label and address for owner inspection before signing.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - targetVersion: The published version label to resolve.
func (c *Client) ResolveUpgradeTarget(
	ctx context.Context,
	walletID uuid.UUID,
	targetVersion string,
) (*UpgradeTarget, error) {
	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if targetVersion == "" {
		return nil, ErrTargetVersionRequired
	}

	wallet, err := c.Get(ctx, walletID)
	if err != nil {
		return nil, err
	}

	version, err := c.GetTypeVersion(ctx, wallet.WalletType, targetVersion)
	if err != nil {
		return nil, err
	}

	target := &UpgradeTarget{Version: version.ImplementationVersion}
	if wallet.ImplementationVersion != nil {
		target.CurrentVersion = *wallet.ImplementationVersion
	}
	for _, d := range version.Deployments {
		if d.ChainSelector == wallet.ChainSelector {
			target.ImplementationAddress = d.ImplementationAddress
			target.IsDefaultOnNetwork = d.IsDefault
			break
		}
	}
	if target.ImplementationAddress == "" {
		return nil, fmt.Errorf("%w: version %s is not deployed on chain %s",
			ErrNoDeploymentOnNetwork, targetVersion, wallet.ChainSelector)
	}
	return target, nil
}

// ErrNoDeploymentOnNetwork is returned when the resolved version has no
// deployment on the wallet's network.
var ErrNoDeploymentOnNetwork = errors.New("version not deployed on the wallet's network")

// UpgradeOptions configures [Client.UpgradeWallet].
//   - OwnerSigner: Signs the upgrade authorization. Required; the wallet's
//     owner identity (ECDSA key).
//   - OperationSigner: Signs the relaying operation. Optional; defaults to
//     OwnerSigner. Required separately when the owner is not an allowed
//     operation signer of the wallet.
//   - Deadline: Expiry shared by the authorization and the operation. Required;
//     must be in the future. A stalled upgrade expires back to operational.
//   - AuthorizationOpID: The authorization's identifier in the wallet's
//     admin-authorization namespace. Optional; a random uint256 is generated.
//   - OperationOpID: The relaying operation's identifier in the
//     executed-operation namespace. Optional; a random uint256 is generated.
//     Must differ from AuthorizationOpID.
//   - PollInterval: How often [Client.UpgradeWallet] polls the upgrade.
//     Optional; defaults to 1s.
//   - PollTimeout: How long [Client.UpgradeWallet] waits for a terminal
//     outcome. Optional; defaults to 5 minutes.
type UpgradeOptions struct {
	OwnerSigner       signer.Signer
	OperationSigner   signer.Signer
	Deadline          time.Time
	AuthorizationOpID *big.Int
	OperationOpID     *big.Int
	PollInterval      time.Duration
	PollTimeout       time.Duration
}

// UpgradeWallet builds, signs, submits and awaits a wallet upgrade: it signs
// the pinned upgrade authorization with the owner identity, relays it as a
// signed self-call operation on the wallet, submits it, and polls until the
// outcome is terminal. On a failed or expired outcome it returns the upgrade
// alongside [ErrUpgradeFailed] (with the verbatim revert reason) or
// [ErrUpgradeExpired].
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet to upgrade.
//   - target: The resolved upgrade target, from [Client.ResolveUpgradeTarget].
//   - opts: Signing, deadline and polling options.
func (c *Client) UpgradeWallet(
	ctx context.Context,
	walletID uuid.UUID,
	target *UpgradeTarget,
	opts UpgradeOptions,
) (*apiClient.WalletUpgrade, error) {
	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if target == nil {
		return nil, ErrUpgradeTargetRequired
	}
	if opts.OwnerSigner == nil {
		return nil, ErrOwnerSignerRequired
	}
	now := time.Now()
	if opts.Deadline.IsZero() || !opts.Deadline.After(now) {
		return nil, ErrDeadlineRequired
	}
	opSigner := opts.OperationSigner
	if opSigner == nil {
		opSigner = opts.OwnerSigner
	}

	wallet, err := c.Get(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Address == "" {
		return nil, ErrWalletAddressMissing
	}
	walletAddress := common.HexToAddress(wallet.Address)

	authOpID := opts.AuthorizationOpID
	if authOpID == nil {
		if authOpID, err = randomUint256(); err != nil {
			return nil, err
		}
	}
	operationOpID := opts.OperationOpID
	if operationOpID == nil {
		if operationOpID, err = randomUint256(); err != nil {
			return nil, err
		}
	}
	deadlineUnix := opts.Deadline.Unix()
	deadline := big.NewInt(deadlineUnix)

	handler, err := eip712.NewHandler(&eip712.Options{Logger: c.logger})
	if err != nil {
		return nil, err
	}

	// Owner authorization over the pinned upgrade struct.
	authorization := &eip712.AdminAuthorization{
		EIP712Type: eip712.UpgradeAuthorizationType,
		Arguments: map[string]any{
			"newImplementation": target.ImplementationAddress,
		},
		OpID:     authOpID,
		Deadline: deadline,
	}
	_, authSig, err := handler.SignAdminAuthorization(ctx, authorization, opts.OwnerSigner, walletAddress, wallet.ChainSelector)
	if err != nil {
		return nil, err
	}

	// The relaying operation: one zero-value self-call packing the upgrade.
	calldata, err := adminop.PackUpgradeToWithAuthorization(
		common.HexToAddress(target.ImplementationAddress), authOpID, deadline, authSig)
	if err != nil {
		return nil, err
	}
	operation := &transactTypes.Operation{
		ID:       operationOpID,
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := handler.SignOperation(ctx, operation, opSigner, wallet.ChainSelector)
	if err != nil {
		return nil, err
	}
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	upgrade, err := c.CreateUpgrade(ctx, walletID, CreateUpgradeInput{
		TargetVersion:         target.Version,
		ImplementationAddress: target.ImplementationAddress,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: authOpID.String(),
			Deadline:          deadlineUnix,
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadlineUnix,
			Signature:         &opSigHex,
			WalletOperationId: operationOpID.String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	})
	if err != nil {
		return nil, err
	}

	return c.AwaitUpgrade(ctx, walletID, upgrade.UpgradeId, opts.pollInterval(), opts.pollTimeout())
}

// AwaitUpgrade polls the upgrade until its outcome is terminal (succeeded,
// failed or expired) or the timeout elapses. On failed or expired it returns
// the upgrade alongside [ErrUpgradeFailed] (with the verbatim revert reason)
// or [ErrUpgradeExpired].
func (c *Client) AwaitUpgrade(
	ctx context.Context,
	walletID, upgradeID uuid.UUID,
	pollInterval, pollTimeout time.Duration,
) (*apiClient.WalletUpgrade, error) {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	if pollTimeout <= 0 {
		pollTimeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	for {
		upgrade, err := c.GetUpgrade(ctx, walletID, upgradeID)
		if err != nil {
			return nil, err
		}
		switch upgrade.Outcome {
		case apiClient.WalletUpgradeOutcomeSucceeded:
			return upgrade, nil
		case apiClient.WalletUpgradeOutcomeFailed:
			reason := ""
			if upgrade.FailureReason != nil {
				reason = *upgrade.FailureReason
			}
			return upgrade, fmt.Errorf("%w: %s", ErrUpgradeFailed, reason)
		case apiClient.WalletUpgradeOutcomeExpired:
			return upgrade, ErrUpgradeExpired
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", ErrAwaitTimeout, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

// AdminOperationOptions configures [Client.SubmitAdminOperation].
//   - OwnerSigner: Signs the administrative authorization. Required; the
//     wallet's owner identity (ECDSA key).
//   - OperationSigner: Signs the relaying operation. Optional; defaults to
//     OwnerSigner.
//   - Deadline: Expiry shared by the authorization and the operation.
//     Required; must be in the future.
//   - AuthorizationOpID: The authorization's identifier in the wallet's
//     admin-authorization namespace. Optional; a random uint256 is generated.
//   - OperationOpID: The relaying operation's identifier. Optional; a random
//     uint256 is generated. Must differ from AuthorizationOpID.
//   - PollInterval / PollTimeout: How submission polling behaves; default 1s
//     and 5 minutes.
type AdminOperationOptions struct {
	OwnerSigner       signer.Signer
	OperationSigner   signer.Signer
	Deadline          time.Time
	AuthorizationOpID *big.Int
	OperationOpID     *big.Int
	PollInterval      time.Duration
	PollTimeout       time.Duration
}

// SubmitAdminOperation builds, signs, submits and awaits an administrative
// wallet operation for a capability the wallet's recorded version publishes:
// it resolves the capability from the catalog, signs its pinned EIP-712 struct
// with the owner identity, relays the packed call as a signed self-call
// operation, submits it, and polls until the outcome is terminal. Arguments
// must be canonically encoded exactly as submitted (integers as decimal
// strings, addresses and byte strings lowercase 0x hex, booleans and strings
// as JSON values).
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - capability: The capability name the wallet's recorded version publishes.
//   - arguments: Values keyed by parameter name, canonically encoded.
//   - opts: Signing, deadline and polling options.
func (c *Client) SubmitAdminOperation(
	ctx context.Context,
	walletID uuid.UUID,
	capability string,
	arguments map[string]any,
	opts AdminOperationOptions,
) (*apiClient.WalletAdminOperation, error) {
	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if capability == "" {
		return nil, ErrCapabilityRequired
	}
	if opts.OwnerSigner == nil {
		return nil, ErrOwnerSignerRequired
	}
	now := time.Now()
	if opts.Deadline.IsZero() || !opts.Deadline.After(now) {
		return nil, ErrDeadlineRequired
	}
	opSigner := opts.OperationSigner
	if opSigner == nil {
		opSigner = opts.OwnerSigner
	}

	wallet, err := c.Get(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if wallet.Address == "" {
		return nil, ErrWalletAddressMissing
	}
	walletAddress := common.HexToAddress(wallet.Address)
	if wallet.ImplementationVersion == nil || *wallet.ImplementationVersion == "" {
		return nil, ErrNoVersionRecorded
	}

	// The wallet's recorded version publishes the capabilities it accepts.
	version, err := c.GetTypeVersion(ctx, wallet.WalletType, *wallet.ImplementationVersion)
	if err != nil {
		return nil, err
	}
	var descriptor *apiClient.WalletCapabilityDescriptor
	for i := range version.Capabilities {
		if version.Capabilities[i].Name == capability {
			descriptor = &version.Capabilities[i]
			break
		}
	}
	if descriptor == nil {
		return nil, fmt.Errorf("%w: version %s of %s does not publish %q",
			ErrCapabilityNotPublished, *wallet.ImplementationVersion, wallet.WalletType, capability)
	}

	authOpID := opts.AuthorizationOpID
	if authOpID == nil {
		if authOpID, err = randomUint256(); err != nil {
			return nil, err
		}
	}
	operationOpID := opts.OperationOpID
	if operationOpID == nil {
		if operationOpID, err = randomUint256(); err != nil {
			return nil, err
		}
	}
	deadlineUnix := opts.Deadline.Unix()
	deadline := big.NewInt(deadlineUnix)

	handler, err := eip712.NewHandler(&eip712.Options{Logger: c.logger})
	if err != nil {
		return nil, err
	}

	// Owner authorization over the capability's pinned struct.
	authorization := &eip712.AdminAuthorization{
		EIP712Type: descriptor.Eip712Type,
		Arguments:  arguments,
		OpID:       authOpID,
		Deadline:   deadline,
	}
	_, authSig, err := handler.SignAdminAuthorization(ctx, authorization, opts.OwnerSigner, walletAddress, wallet.ChainSelector)
	if err != nil {
		return nil, err
	}

	// The relaying operation: one zero-value self-call packing the capability.
	calldata, err := adminop.PackCapabilityCall(descriptor.AbiFragment, arguments, authOpID, deadline, authSig)
	if err != nil {
		return nil, err
	}
	operation := &transactTypes.Operation{
		ID:       operationOpID,
		Account:  walletAddress,
		Deadline: deadline,
		Transactions: []transactTypes.Transaction{{
			To:    walletAddress,
			Value: big.NewInt(0),
			Data:  calldata,
		}},
	}
	_, opSig, err := handler.SignOperation(ctx, operation, opSigner, wallet.ChainSelector)
	if err != nil {
		return nil, err
	}
	opSigHex := "0x" + common.Bytes2Hex(opSig)

	adminOp, err := c.CreateAdminOperation(ctx, walletID, CreateAdminOperationInput{
		Capability: capability,
		Arguments:  arguments,
		Authorization: apiClient.AdminAuthorization{
			WalletOperationId: authOpID.String(),
			Deadline:          deadlineUnix,
			Signature:         "0x" + common.Bytes2Hex(authSig),
		},
		Operation: apiClient.CreateOperation{
			Address:           wallet.Address,
			ChainSelector:     wallet.ChainSelector,
			Deadline:          deadlineUnix,
			Signature:         &opSigHex,
			WalletOperationId: operationOpID.String(),
			Transactions: []apiClient.TransactionRequest{{
				To:    wallet.Address,
				Value: "0",
				Data:  "0x" + common.Bytes2Hex(calldata),
			}},
		},
	})
	if err != nil {
		return nil, err
	}

	return c.AwaitAdminOperation(ctx, walletID, adminOp.AdminOperationId, opts.pollInterval(), opts.pollTimeout())
}

// AwaitAdminOperation polls the administrative operation until its outcome is
// terminal (succeeded, failed or expired) or the timeout elapses. On failed or
// expired it returns the operation alongside [ErrAdminOperationFailed] (with
// the verbatim revert reason) or [ErrAdminOperationExpired].
func (c *Client) AwaitAdminOperation(
	ctx context.Context,
	walletID, adminOperationID uuid.UUID,
	pollInterval, pollTimeout time.Duration,
) (*apiClient.WalletAdminOperation, error) {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	if pollTimeout <= 0 {
		pollTimeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	for {
		adminOp, err := c.GetAdminOperation(ctx, walletID, adminOperationID)
		if err != nil {
			return nil, err
		}
		switch adminOp.Outcome {
		case apiClient.WalletUpgradeOutcomeSucceeded:
			return adminOp, nil
		case apiClient.WalletUpgradeOutcomeFailed:
			reason := ""
			if adminOp.FailureReason != nil {
				reason = *adminOp.FailureReason
			}
			return adminOp, fmt.Errorf("%w: %s", ErrAdminOperationFailed, reason)
		case apiClient.WalletUpgradeOutcomeExpired:
			return adminOp, ErrAdminOperationExpired
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", ErrAwaitTimeout, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

// ErrCapabilityNotPublished is returned when the wallet's recorded version does
// not publish the named capability in the catalog.
var ErrCapabilityNotPublished = errors.New("capability not published by the wallet's recorded version")

// randomUint256 draws a uniformly random uint256 for caller-chosen operation
// namespaces. Constructing an authorization never requires reading chain state.
func randomUint256() (*big.Int, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to draw random operation id: %w", err)
	}
	return new(big.Int).SetBytes(b), nil
}

func (o UpgradeOptions) pollInterval() time.Duration {
	if o.PollInterval <= 0 {
		return time.Second
	}
	return o.PollInterval
}

func (o UpgradeOptions) pollTimeout() time.Duration {
	if o.PollTimeout <= 0 {
		return 5 * time.Minute
	}
	return o.PollTimeout
}

func (o AdminOperationOptions) pollInterval() time.Duration {
	if o.PollInterval <= 0 {
		return time.Second
	}
	return o.PollInterval
}

func (o AdminOperationOptions) pollTimeout() time.Duration {
	if o.PollTimeout <= 0 {
		return 5 * time.Minute
	}
	return o.PollTimeout
}
