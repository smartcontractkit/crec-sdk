package wallets

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

// CreateAdminOperationInput defines the parameters for submitting a signed
// administrative wallet operation. Build it with the transact/eip712 and
// transact/adminop helpers, or with [Client.SubmitAdminOperation] which
// composes them.
//   - Capability: A capability name the wallet's recorded version publishes.
//   - Arguments: Values keyed by parameter name, canonically encoded (integers
//     as decimal strings, addresses and byte strings as lowercase 0x hex,
//     booleans and strings as JSON values).
//   - Authorization: The owner's signature over the capability's pinned
//     EIP-712 struct (the version's eip712_type).
//   - Operation: The relaying operation: exactly one zero-value self-call on
//     the wallet whose calldata is the packed capability call.
type CreateAdminOperationInput struct {
	Capability    string
	Arguments     map[string]any
	Authorization apiClient.AdminAuthorization
	Operation     apiClient.CreateOperation
}

// CreateAdminOperation submits a signed administrative wallet operation. The
// operation settles asynchronously: poll with [Client.GetAdminOperation] or
// wait with [Client.AwaitAdminOperation] until its outcome leaves pending.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - input: The administrative operation parameters.
//
// Returns the created administrative operation (outcome pending) or an error.
func (c *Client) CreateAdminOperation(
	ctx context.Context,
	walletID uuid.UUID,
	input CreateAdminOperationInput,
) (*apiClient.WalletAdminOperation, error) {
	c.logger.Debug("Creating wallet admin operation",
		"wallet_id", walletID.String(),
		"capability", input.Capability)

	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if input.Capability == "" {
		return nil, ErrCapabilityRequired
	}
	if input.Arguments == nil {
		return nil, ErrArgumentsRequired
	}
	if err := validateRelayConstraints(input.Authorization, input.Operation); err != nil {
		return nil, err
	}

	req := apiClient.CreateWalletAdminOperation{
		Capability:    input.Capability,
		Arguments:     input.Arguments,
		Authorization: input.Authorization,
		Operation:     input.Operation,
	}

	resp, err := c.apiClient.CreateWalletAdminOperationWithResponse(ctx, walletID, req)
	if err != nil {
		c.logger.Error("Failed to create wallet admin operation", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateAdminOperation, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdminOperation, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusAccepted:
		if resp.JSON202 == nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateAdminOperation, apierror.ErrNilResponseBody)
		}
		c.logger.Info("Wallet admin operation submitted",
			"wallet_id", walletID.String(),
			"admin_operation_id", resp.JSON202.AdminOperationId.String(),
			"capability", resp.JSON202.Capability)
		return resp.JSON202, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "creating wallet admin operation", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, apierror.WrapNotFound(resp.JSON404, ErrCreateAdminOperation, "wallet ID "+walletID.String())
	case http.StatusConflict:
		c.logger.Warn("Conflict when creating wallet admin operation",
			"wallet_id", walletID.String(),
			"code", apierror.ConflictCode(resp.JSON409))
		return nil, apierror.WrapConflict(resp.JSON409, ErrCreateAdminOperation, "wallet ID "+walletID.String())
	case http.StatusBadRequest:
		c.logger.Warn("Validation error when creating wallet admin operation",
			"wallet_id", walletID.String(),
			"code", apierror.ValidationCode(resp.JSON400))
		return nil, apierror.WrapValidation(resp.JSON400, ErrCreateAdminOperation)
	default:
		return nil, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrCreateAdminOperation, "creating wallet admin operation", resp.Body, c.logger)
	}
}

// ListAdminOperationsInput defines the pagination for listing a wallet's
// administrative operations.
//   - Limit: Maximum number of operations to return per page.
//   - Offset: Number of operations to skip for pagination (default: 0).
type ListAdminOperationsInput struct {
	Limit  *int
	Offset *int64
}

// ListAdminOperations retrieves a wallet's administrative operations with
// their capability, arguments, authorizing key and outcome.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - input: The pagination parameters.
//
// Returns the operations and a boolean indicating if there are more results.
func (c *Client) ListAdminOperations(
	ctx context.Context,
	walletID uuid.UUID,
	input ListAdminOperationsInput,
) ([]apiClient.WalletAdminOperation, bool, error) {
	c.logger.Debug("Listing wallet admin operations", "wallet_id", walletID.String(), "filters", input)

	if walletID == uuid.Nil {
		return nil, false, ErrWalletIDRequired
	}
	if input.Limit != nil && *input.Limit < 1 {
		return nil, false, ErrInvalidLimit
	}
	if input.Offset != nil && *input.Offset < 0 {
		return nil, false, ErrInvalidOffset
	}

	params := apiClient.ListWalletAdminOperationsParams{
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	resp, err := c.apiClient.ListWalletAdminOperationsWithResponse(ctx, walletID, &params)
	if err != nil {
		c.logger.Error("Failed to list wallet admin operations", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListAdminOperations, err)
	}

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListAdminOperations, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, false, fmt.Errorf("%w: %w", ErrListAdminOperations, apierror.ErrNilResponseBody)
		}
		return resp.JSON200.Data, resp.JSON200.HasMore, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "listing wallet admin operations", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, false, apierror.WrapNotFound(resp.JSON404, ErrListAdminOperations, "wallet ID "+walletID.String())
	default:
		return nil, false, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrListAdminOperations, "listing wallet admin operations", resp.Body, c.logger)
	}
}

// GetAdminOperation retrieves one administrative operation of a wallet, with
// its outcome and, on failure, the verbatim onchain revert reason.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet.
//   - adminOperationID: The UUID of the administrative operation.
//
// Returns the operation or an error if it does not exist.
func (c *Client) GetAdminOperation(
	ctx context.Context,
	walletID, adminOperationID uuid.UUID,
) (*apiClient.WalletAdminOperation, error) {
	c.logger.Debug("Getting wallet admin operation",
		"wallet_id", walletID.String(),
		"admin_operation_id", adminOperationID.String())

	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}
	if adminOperationID == uuid.Nil {
		return nil, ErrAdminOperationIDRequired
	}

	resp, err := c.apiClient.GetWalletAdminOperationWithResponse(ctx, walletID, adminOperationID)
	if err != nil {
		c.logger.Error("Failed to get wallet admin operation", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetAdminOperation, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetAdminOperation, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("%w: %w", ErrGetAdminOperation, apierror.ErrNilResponseBody)
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "getting wallet admin operation", nil),
			"wallet_id", walletID.String(),
			"admin_operation_id", adminOperationID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, apierror.WrapNotFound(resp.JSON404, ErrGetAdminOperation, "admin operation ID "+adminOperationID.String())
	default:
		return nil, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrGetAdminOperation, "getting wallet admin operation", resp.Body, c.logger)
	}
}
