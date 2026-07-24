package wallets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

const (
	// MaxWalletNameLength is the maximum allowed length for a wallet name in characters.
	MaxWalletNameLength = 255
)

// Sentinel errors
var (
	// ErrWalletNotFound is returned when a wallet is not found (404 response)
	ErrWalletNotFound = apierror.ErrWalletNotFound

	// Client initialization errors
	ErrOptionsRequired   = errors.New("options is required")
	ErrAPIClientRequired = errors.New("APIClient is required")

	// Validation errors
	ErrNameRequired               = errors.New("wallet name is required")
	ErrNameTooLong                = errors.New("wallet name too long")
	ErrChainSelectorRequired      = errors.New("chain selector is required")
	ErrWalletOwnerAddressRequired = errors.New("wallet owner address is required")
	ErrInvalidWalletOwnerAddress  = errors.New("wallet owner address must be a valid hex address")
	ErrWalletTypeRequired         = errors.New("wallet type is required")
	ErrWalletIDRequired           = errors.New("wallet ID is required")
	ErrStatusChannelIDZero        = errors.New("status channel ID cannot be the zero UUID")
	ErrConfigurationRequired      = errors.New("configuration is required")
	ErrInvalidLimit               = errors.New("limit must be positive")
	ErrInvalidOffset              = errors.New("offset cannot be negative")
	ErrInvalidOwnerAddress        = errors.New("owner address must be a valid hex address")

	// API operation errors
	ErrCreateWallet  = errors.New("failed to create wallet")
	ErrGetWallet     = errors.New("failed to get wallet")
	ErrListWallets   = errors.New("failed to list wallets")
	ErrUpdateWallet  = errors.New("failed to update wallet")
	ErrArchiveWallet = errors.New("failed to archive wallet")
)

// Options defines the options for creating a new CREC Wallets client.
//   - Logger: Optional logger instance.
//   - APIClient: The CREC API client instance.
type Options struct {
	Logger    *slog.Logger
	APIClient *apiClient.ClientWithResponses
}

// Client provides operations for managing CREC wallets.
// Wallets are blockchain wallets that can be used to sign transactions
// and interact with smart contracts.
type Client struct {
	logger    *slog.Logger
	apiClient *apiClient.ClientWithResponses
}

// NewClient creates a new CREC Wallets client with the provided options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Wallets client, see Options for details.
func NewClient(opts *Options) (*Client, error) {
	if opts == nil {
		return nil, ErrOptionsRequired
	}

	if opts.APIClient == nil {
		return nil, ErrAPIClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Debug("Creating CREC Wallets client")

	return &Client{
		logger:    logger,
		apiClient: opts.APIClient,
	}, nil
}

// CreateInput defines the input parameters for creating a new wallet.
//   - Name: The name of the wallet.
//   - ChainSelector: The chain selector identifying the blockchain network.
//   - WalletOwnerAddress: The wallet contract owner address (42-character hex string starting with 0x).
//   - WalletType: The type of the wallet (e.g., "ecdsa").
//   - Configuration: Wallet type-specific configuration (required by the CREC API).
//   - Description: Optional description of the wallet.
//   - StatusChannelId: Optional unique identifier for the channel where the wallet status will be published
type CreateInput struct {
	Name               string
	ChainSelector      string
	WalletOwnerAddress string
	WalletType         apiClient.WalletType
	Configuration      apiClient.WalletConfiguration
	Description        *string
	StatusChannelId    *uuid.UUID `json:"status_channel_id,omitempty"`
}

// Create creates a new wallet in the CREC backend.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The wallet creation parameters.
//
// Returns the created wallet or an error if the operation fails.
func (c *Client) Create(ctx context.Context, input CreateInput) (*apiClient.Wallet, error) {
	c.logger.Debug("Creating wallet", "name", input.Name, "chain_selector", input.ChainSelector)

	if input.Name == "" {
		return nil, ErrNameRequired
	}

	if len(input.Name) > MaxWalletNameLength {
		return nil, fmt.Errorf("%w: cannot exceed %d characters", ErrNameTooLong, MaxWalletNameLength)
	}

	if input.ChainSelector == "" {
		return nil, ErrChainSelectorRequired
	}

	if input.WalletOwnerAddress == "" {
		return nil, ErrWalletOwnerAddressRequired
	}

	if !common.IsHexAddress(input.WalletOwnerAddress) {
		return nil, ErrInvalidWalletOwnerAddress
	}

	if input.WalletType == "" {
		return nil, ErrWalletTypeRequired
	}

	if input.StatusChannelId != nil && *input.StatusChannelId == uuid.Nil {
		return nil, ErrStatusChannelIDZero
	}

	if len(input.Configuration) == 0 {
		return nil, fmt.Errorf("%w: wallet configuration is required", ErrConfigurationRequired)
	}

	var apiStatusChannelID *openapitypes.UUID
	if input.StatusChannelId != nil {
		v := openapitypes.UUID(*input.StatusChannelId)
		apiStatusChannelID = &v
	}

	createWalletReq := apiClient.CreateWallet{
		Name:               input.Name,
		ChainSelector:      input.ChainSelector,
		WalletOwnerAddress: input.WalletOwnerAddress,
		WalletType:         input.WalletType,
		Configuration:      input.Configuration,
		Description:        input.Description,
		StatusChannelId:    apiStatusChannelID,
	}

	resp, err := c.apiClient.CreateWalletWithResponse(ctx, createWalletReq)
	if err != nil {
		c.logger.Error("Failed to create wallet", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWallet, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateWallet, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateWallet, apierror.ErrNilResponseBody)
		}
		c.logger.Info("Wallet created successfully",
			"wallet_id", resp.JSON201.WalletId.String(),
			"name", resp.JSON201.Name,
			"address", resp.JSON201.Address,
			"chain_selector", resp.JSON201.ChainSelector)
		return resp.JSON201, nil
	case http.StatusConflict:
		c.logger.Warn("Conflict when creating wallet",
			"name", input.Name,
			"code", apierror.ConflictCode(resp.JSON409))
		return nil, apierror.WrapConflict(resp.JSON409, ErrCreateWallet, "name "+input.Name)
	case http.StatusUnauthorized:
		c.logger.Error("Unauthorized when creating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, apierror.Wrap(resp.JSON401, ErrCreateWallet, resp.StatusCode())
	default:
		c.logger.Error("Unexpected status code when creating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWallet, apierror.ErrUnexpectedStatusCode, resp.StatusCode())
	}
}

// Get retrieves a specific wallet by its ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet to retrieve.
//
// Returns the wallet or an error if the operation fails or the wallet is not found.
func (c *Client) Get(ctx context.Context, walletID uuid.UUID) (*apiClient.Wallet, error) {
	c.logger.Debug("Getting wallet", "wallet_id", walletID.String())

	if walletID == uuid.Nil {
		return nil, ErrWalletIDRequired
	}

	resp, err := c.apiClient.GetWalletWithResponse(ctx, walletID)
	if err != nil {
		c.logger.Error("Failed to get wallet", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetWallet, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetWallet, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("%w: %w", ErrGetWallet, apierror.ErrNilResponseBody)
		}
		c.logger.Debug("Wallet retrieved successfully",
			"wallet_id", resp.JSON200.WalletId.String(),
			"name", resp.JSON200.Name,
			"address", resp.JSON200.Address)
		return resp.JSON200, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "getting wallet", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, fmt.Errorf("%w: wallet ID %s", ErrWalletNotFound, walletID.String())
	case http.StatusUnauthorized:
		c.logger.Error("Unauthorized when getting wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, apierror.Wrap(resp.JSON401, ErrGetWallet, resp.StatusCode())
	default:
		c.logger.Error("Unexpected status code when getting wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetWallet, apierror.ErrUnexpectedStatusCode, resp.StatusCode())
	}
}

// ListInput defines the input parameters for listing wallets.
//   - Name: Optional filter to search wallets by name (case-insensitive partial match).
//   - ChainSelector: Optional filter to search wallets by chain selector.
//   - Owner: Optional filter to search wallets by owner address (42-character hex string starting with 0x).
//   - Address: Optional filter to search wallets by wallet address (42-character hex string starting with 0x).
//   - Type: Optional filter to search wallets by type (e.g., "ecdsa", "rsa").
//   - Status: Optional filter to search wallets by status (e.g., "deployed", "deploying", "failed", "pending", "deleted").
//   - Limit: Maximum number of wallets to return per page.
//   - Offset: Number of wallets to skip for pagination (default: 0).
type ListInput struct {
	Name          *string
	ChainSelector *string
	Owner         *string
	Address       *string
	Type          *apiClient.WalletType
	Status        *[]apiClient.WalletStatus
	Limit         *int
	Offset        *int64
}

// List retrieves a list of wallets.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The list parameters including filters and pagination.
//
// Returns a list of wallets and a boolean indicating if there are more results.
func (c *Client) List(ctx context.Context, input ListInput) ([]apiClient.Wallet, bool, error) {
	c.logger.Debug("Listing wallets", "filters", input)

	// Validate limit if provided
	if input.Limit != nil && *input.Limit < 1 {
		return nil, false, ErrInvalidLimit
	}

	// Validate offset if provided
	if input.Offset != nil && *input.Offset < 0 {
		return nil, false, ErrInvalidOffset
	}

	// Validate owner address if provided
	if input.Owner != nil && !common.IsHexAddress(*input.Owner) {
		return nil, false, fmt.Errorf("%w: %s", ErrInvalidOwnerAddress, *input.Owner)
	}

	// Validate wallet address if provided
	if input.Address != nil && !common.IsHexAddress(*input.Address) {
		return nil, false, fmt.Errorf("%w: %s", ErrInvalidOwnerAddress, *input.Address)
	}

	params := apiClient.ListWalletsParams{
		Name:          input.Name,
		ChainSelector: input.ChainSelector,
		Owner:         input.Owner,
		Address:       input.Address,
		Type:          input.Type,
		Status:        input.Status,
		Limit:         input.Limit,
		Offset:        input.Offset,
	}

	resp, err := c.apiClient.ListWalletsWithResponse(ctx, &params)
	if err != nil {
		c.logger.Error("Failed to list wallets", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListWallets, err)
	}

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListWallets, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, false, fmt.Errorf("%w: %w", ErrListWallets, apierror.ErrNilResponseBody)
		}
		c.logger.Debug("Wallets listed successfully",
			"count", len(resp.JSON200.Data),
			"has_more", resp.JSON200.HasMore)
		return resp.JSON200.Data, resp.JSON200.HasMore, nil
	case http.StatusUnauthorized:
		c.logger.Error("Unauthorized when listing wallets",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, apierror.Wrap(resp.JSON401, ErrListWallets, resp.StatusCode())
	default:
		c.logger.Error("Unexpected status code when listing wallets",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListWallets, apierror.ErrUnexpectedStatusCode, resp.StatusCode())
	}
}

// UpdateInput defines the input parameters for updating a wallet.
//   - Name: The new name for the wallet.
type UpdateInput struct {
	Name string
}

// Update updates a wallet's information.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet to update.
//   - input: The wallet update parameters.
//
// Returns an error if the operation fails or the wallet is not found.
func (c *Client) Update(ctx context.Context, walletID uuid.UUID, input UpdateInput) error {
	c.logger.Debug("Updating wallet", "wallet_id", walletID.String(), "name", input.Name)

	if walletID == uuid.Nil {
		return ErrWalletIDRequired
	}

	if input.Name == "" {
		return ErrNameRequired
	}

	if len(input.Name) > MaxWalletNameLength {
		return fmt.Errorf("%w: cannot exceed %d characters", ErrNameTooLong, MaxWalletNameLength)
	}

	updateWalletReq := apiClient.UpdateWallet{
		Name: &input.Name,
	}

	resp, err := c.apiClient.UpdateWalletWithResponse(ctx, walletID, updateWalletReq)
	if err != nil {
		c.logger.Error("Failed to update wallet", "error", err)
		return fmt.Errorf("%w: %w", ErrUpdateWallet, err)
	}

	if resp == nil {
		return fmt.Errorf("%w: %w", ErrUpdateWallet, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		c.logger.Info("Wallet updated successfully", "wallet_id", walletID.String())
		return nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "updating wallet", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return fmt.Errorf("%w: wallet ID %s", ErrWalletNotFound, walletID.String())
	case http.StatusConflict:
		c.logger.Warn("Conflict when updating wallet",
			"wallet_id", walletID.String(),
			"code", apierror.ConflictCode(resp.JSON409))
		return apierror.WrapConflict(resp.JSON409, ErrUpdateWallet, "wallet ID "+walletID.String())
	case http.StatusUnauthorized:
		c.logger.Error("Unauthorized when updating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return apierror.Wrap(resp.JSON401, ErrUpdateWallet, resp.StatusCode())
	default:
		c.logger.Error("Unexpected status code when updating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return fmt.Errorf("%w: %w (status code %d)", ErrUpdateWallet, apierror.ErrUnexpectedStatusCode, resp.StatusCode())
	}
}

// Archive archives a wallet by transitioning it to archived status via PATCH.
// Archiving is synchronous and returns the wallet in "archived" status.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletID: The UUID of the wallet to archive.
//
// Returns an error if the operation fails or the wallet is not found.
func (c *Client) Archive(ctx context.Context, walletID uuid.UUID) error {
	c.logger.Debug("Archiving wallet", "wallet_id", walletID.String())

	if walletID == uuid.Nil {
		return ErrWalletIDRequired
	}

	archiveStatus := apiClient.WalletStatusArchived
	archiveReq := apiClient.UpdateWallet{
		Status: &archiveStatus,
	}

	resp, err := c.apiClient.UpdateWalletWithResponse(ctx, walletID, archiveReq)
	if err != nil {
		c.logger.Error("Failed to archive wallet", "error", err)
		return fmt.Errorf("%w: %w", ErrArchiveWallet, err)
	}

	if resp == nil {
		return fmt.Errorf("%w: %w", ErrArchiveWallet, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		c.logger.Info("Wallet archived successfully", "wallet_id", walletID.String())
		return nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "archiving wallet", apierror.ErrWalletNotFound),
			"wallet_id", walletID.String(),
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return fmt.Errorf("%w: wallet ID %s", ErrWalletNotFound, walletID.String())
	case http.StatusUnauthorized:
		c.logger.Error("Unauthorized when archiving wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return apierror.Wrap(resp.JSON401, ErrArchiveWallet, resp.StatusCode())
	default:
		c.logger.Error("Unexpected status code when archiving wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return fmt.Errorf("%w: %w (status code %d)", ErrArchiveWallet, apierror.ErrUnexpectedStatusCode, resp.StatusCode())
	}
}
