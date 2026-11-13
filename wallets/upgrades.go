package wallets

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

// Upgrade and administrative operation errors.
var (
	// Validation errors
	ErrVersionRequired               = errors.New("implementation version is required")
	ErrTargetVersionRequired         = errors.New("target version is required")
	ErrImplementationAddressRequired = errors.New("implementation address is required")
	ErrInvalidImplementationAddress  = errors.New("implementation address must be a valid hex address")
	ErrUpgradeIDRequired             = errors.New("upgrade ID is required")
	ErrAdminOperationIDRequired      = errors.New("admin operation ID is required")
	ErrCapabilityRequired            = errors.New("capability is required")
	ErrAuthorizationDeadline         = errors.New("authorization deadline must be positive")
	ErrAuthorizationSignature        = errors.New("authorization signature is required")
	ErrOperationDeadlineRequired     = errors.New("operation deadline must be positive")
	ErrOperationDeadlineMismatch     = errors.New("operation deadline must equal the authorization deadline")
	ErrOperationIDCollision          = errors.New("authorization wallet_operation_id must differ from the operation's")
	ErrSingleTransactionRequired     = errors.New("an administrative operation carries exactly one transaction")
	ErrArgumentsRequired             = errors.New("arguments are required")

	// API operation errors
	ErrCreateUpgrade        = errors.New("failed to create wallet upgrade")
	ErrListUpgrades         = errors.New("failed to list wallet upgrades")
	ErrGetUpgrade           = errors.New("failed to get wallet upgrade")
	ErrCreateAdminOperation = errors.New("failed to create wallet admin operation")
	ErrListAdminOperations  = errors.New("failed to list wallet admin operations")
	ErrGetAdminOperation    = errors.New("failed to get wallet admin operation")
)

// CreateUpgradeInput defines the parameters for submitting a client-signed
// wallet upgrade. Build it with the transact/eip712 and transact/adminop
// helpers, or with [Client.UpgradeWallet] which composes them.
//   - TargetVersion: The published version label the wallet upgrades to.
//   - ImplementationAddress: The catalog's address for the target version on
//     the wallet's network. It is what the authorization commits to.
//   - Authorization: The owner's signature over the pinned upgrade EIP-712
//     struct (see eip712.UpgradeAuthorizationType).
//   - Operation: The relaying operation: exactly one zero-value self-call on
//     the wallet whose calldata is the packed upgradeToWithAuthorization call.
type CreateUpgradeInput struct {
	TargetVersion         string
	ImplementationAddress string
	Authorization         apiClient.AdminAuthorization
	Operation             apiClient.CreateOperation
}

// validateRelayConstraints checks the submission-shape rules courier enforces:
// a signed operation of exactly one transaction, whose deadline is the
// authorization's, with distinct wallet_operation_id namespaces.
func validateRelayConstraints(
	authorization apiClient.AdminAuthorization,
	operation apiClient.CreateOperation,
) error {
	if authorization.Deadline <= 0 {
		return ErrAuthorizationDeadline
	}
	if authorization.Signature == "" {
		return ErrAuthorizationSignature
	}
	if operation.Deadline <= 0 {
		return ErrOperationDeadlineRequired
	}
	if operation.Deadline != authorization.Deadline {
		return ErrOperationDeadlineMismatch
	}
	if authorization.WalletOperationId == operation.WalletOperationId {
		return ErrOperationIDCollision
	}
	if operation.Signature == nil || *operation.Signature == "" {
		return fmt.Errorf("%w: the operation must be signed", ErrAuthorizationSignature)
	}
	if len(operation.Transactions) != 1 {
		return ErrSingleTransactionRequired
	}
	return nil
}

// CreateUpgrade submits a client-signed wallet upgrade. The upgrade settles
// asynchronously: poll with [Client.GetUpgrade] or wait with
// [Client.AwaitUpgrade] until its outcome leaves pending.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet to upgrade.
//   - input: The upgrade parameters.
//
// Returns the created upgrade (outcome pending) or an error.
func (c *Client) CreateUpgrade(
	ctx context.Context,
	walletID uuid.UUID,
	input CreateUpgradeInput,
) (*apiClient.WalletUpgrade, error) {
	c.logger.Debug("Creating wallet upgrade",
		"wallet_id", walletID.String(),
		"target_version", input.TargetVersion)

	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if input.TargetVersion == "" {
		return nil, ErrTargetVersionRequired
	}
	if input.ImplementationAddress == "" {
		return nil, ErrImplementationAddressRequired
	}
	if !common.IsHexAddress(input.ImplementationAddress) {
		return nil, ErrInvalidImplementationAddress
	}
	if err := validateRelayConstraints(input.Authorization, input.Operation); err != nil {
		return nil, err
	}

	req := apiClient.CreateWalletUpgrade{
		TargetVersion:         input.TargetVersion,
		ImplementationAddress: input.ImplementationAddress,
		Authorization:         input.Authorization,
		Operation:             input.Operation,
	}

	resp, err := c.apiClient.CreateWalletUpgradeWithResponse(ctx, walletID, req)
	if err != nil {
		c.logger.Error("Failed to create wallet upgrade", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateUpgrade, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateUpgrade, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusAccepted:
		if resp.JSON202 == nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateUpgrade, apierror.ErrNilResponseBody)
		}
		c.logger.Info("Wallet upgrade submitted",
			"wallet_id", walletID.String(),
			"upgrade_id", resp.JSON202.UpgradeId.String(),
			"target_version", resp.JSON202.ToVersion)
		return resp.JSON202, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "creating wallet upgrade", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, apierror.WrapNotFound(resp.JSON404, ErrCreateUpgrade, "wallet ID "+walletID.String())
	case http.StatusConflict:
		c.logger.Warn("Conflict when creating wallet upgrade",
			"wallet_id", walletID.String(),
			"code", apierror.ConflictCode(resp.JSON409))
		return nil, apierror.WrapConflict(resp.JSON409, ErrCreateUpgrade, "wallet ID "+walletID.String())
	case http.StatusBadRequest:
		c.logger.Warn("Validation error when creating wallet upgrade",
			"wallet_id", walletID.String(),
			"code", apierror.ValidationCode(resp.JSON400))
		return nil, apierror.WrapValidation(resp.JSON400, ErrCreateUpgrade)
	default:
		return nil, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrCreateUpgrade, "creating wallet upgrade", resp.Body, c.logger)
	}
}

// ListUpgradesInput defines the filters for listing a wallet's implementation
// version history.
//   - Source: Optional; filters by writer (crec_upgrade or operator_correction).
//   - Outcome: Optional; filters by outcome (pending, succeeded, failed, expired).
//   - Limit: Maximum number of upgrades to return per page.
//   - Offset: Number of upgrades to skip for pagination (default: 0).
type ListUpgradesInput struct {
	Source  *apiClient.WalletVersionChangeSource
	Outcome *apiClient.WalletUpgradeOutcome
	Limit   *int
	Offset  *int64
}

// ListUpgrades retrieves a wallet's implementation version history: upgrades
// and operator corrections, distinguished by source.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - input: The list filters and pagination.
//
// Returns the upgrades and a boolean indicating if there are more results.
func (c *Client) ListUpgrades(
	ctx context.Context,
	walletID uuid.UUID,
	input ListUpgradesInput,
) ([]apiClient.WalletUpgrade, bool, error) {
	c.logger.Debug("Listing wallet upgrades", "wallet_id", walletID.String(), "filters", input)

	if walletID == uuid.Nil {
		return nil, false, ErrWalletIDRequired
	}
	if input.Limit != nil && *input.Limit < 1 {
		return nil, false, ErrInvalidLimit
	}
	if input.Offset != nil && *input.Offset < 0 {
		return nil, false, ErrInvalidOffset
	}

	params := apiClient.ListWalletUpgradesParams{
		Source:  input.Source,
		Outcome: input.Outcome,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}

	resp, err := c.apiClient.ListWalletUpgradesWithResponse(ctx, walletID, &params)
	if err != nil {
		c.logger.Error("Failed to list wallet upgrades", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListUpgrades, err)
	}

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListUpgrades, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, false, fmt.Errorf("%w: %w", ErrListUpgrades, apierror.ErrNilResponseBody)
		}
		return resp.JSON200.Data, resp.JSON200.HasMore, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "listing wallet upgrades", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, false, apierror.WrapNotFound(resp.JSON404, ErrListUpgrades, "wallet ID "+walletID.String())
	default:
		return nil, false, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrListUpgrades, "listing wallet upgrades", resp.Body, c.logger)
	}
}

// GetUpgrade retrieves one upgrade of a wallet, with its outcome and, on
// failure, the verbatim onchain revert reason.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - upgradeID: The UUID of the upgrade.
//
// Returns the upgrade or an error if it does not exist.
func (c *Client) GetUpgrade(
	ctx context.Context,
	walletID, upgradeID uuid.UUID,
) (*apiClient.WalletUpgrade, error) {
	c.logger.Debug("Getting wallet upgrade",
		"wallet_id", walletID.String(),
		"upgrade_id", upgradeID.String())

	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if upgradeID == uuid.Nil {
		return nil, ErrUpgradeIDRequired
	}

	resp, err := c.apiClient.GetWalletUpgradeWithResponse(ctx, walletID, upgradeID)
	if err != nil {
		c.logger.Error("Failed to get wallet upgrade", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetUpgrade, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetUpgrade, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("%w: %w", ErrGetUpgrade, apierror.ErrNilResponseBody)
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "getting wallet upgrade", nil),
			"wallet_id", walletID.String(),
			"upgrade_id", upgradeID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, apierror.WrapNotFound(resp.JSON404, ErrGetUpgrade, "upgrade ID "+upgradeID.String())
	default:
		return nil, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrGetUpgrade, "getting wallet upgrade", resp.Body, c.logger)
	}
}
