package wallets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

const (
	MaxWalletNameLength = 255
)

// Sentinel errors
var (
	// ErrWalletNotFound is returned when a wallet is not found (404 response)
	ErrWalletNotFound = errors.New("wallet not found")

	// Client initialization errors
	ErrOptionsRequired   = errors.New("options is required")
	ErrAPIClientRequired = errors.New("APIClient is required")

	// Validation errors
	ErrNameRequired               = errors.New("wallet name is required")
	ErrNameTooLong                = errors.New("wallet name too long")
	ErrChainSelectorRequired      = errors.New("chain selector is required")
	ErrWalletOwnerAddressRequired = errors.New("wallet owner address is required")
	ErrWalletTypeRequired         = errors.New("wallet type is required")

	// API operation errors
	ErrCreateWallet = errors.New("failed to create wallet")
	ErrGetWallet    = errors.New("failed to get wallet")
	ErrListWallets  = errors.New("failed to list wallets")
	ErrUpdateWallet = errors.New("failed to update wallet")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
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
//   - AllowedEcdsaSigners: Optional list of allowed ECDSA public signing keys.
//   - AllowedRsaSigners: Optional list of allowed RSA public signing keys.
type CreateInput struct {
	Name                string
	ChainSelector       string
	WalletOwnerAddress  string
	WalletType          apiClient.CreateWalletWalletType
	AllowedEcdsaSigners *[]string
	AllowedRsaSigners   *[]struct {
		E string `json:"e"` // RSA public exponent
		N string `json:"n"` // RSA modulus
	}
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

	if input.WalletType == "" {
		return nil, ErrWalletTypeRequired
	}

	createWalletReq := apiClient.CreateWallet{
		Name:                input.Name,
		ChainSelector:       input.ChainSelector,
		WalletOwnerAddress:  input.WalletOwnerAddress,
		WalletType:          input.WalletType,
		AllowedEcdsaSigners: input.AllowedEcdsaSigners,
		AllowedRsaSigners:   input.AllowedRsaSigners,
	}

	resp, err := c.apiClient.PostWalletsWithResponse(ctx, createWalletReq)
	if err != nil {
		c.logger.Error("Failed to create wallet", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWallet, err)
	}

	if resp.StatusCode() != 201 {
		c.logger.Error("Unexpected status code when creating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWallet, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateWallet, ErrNilResponseBody)
	}

	c.logger.Info("Wallet created successfully",
		"wallet_id", resp.JSON201.WalletId.String(),
		"name", resp.JSON201.Name,
		"address", resp.JSON201.Address,
		"chain_selector", resp.JSON201.ChainSelector)

	return resp.JSON201, nil
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

	resp, err := c.apiClient.GetWalletsWalletIdWithResponse(ctx, walletID)
	if err != nil {
		c.logger.Error("Failed to get wallet", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetWallet, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Wallet not found", "wallet_id", walletID.String())
		return nil, fmt.Errorf("%w: wallet ID %s", ErrWalletNotFound, walletID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when getting wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetWallet, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetWallet, ErrNilResponseBody)
	}

	c.logger.Debug("Wallet retrieved successfully",
		"wallet_id", resp.JSON200.WalletId.String(),
		"name", resp.JSON200.Name,
		"address", resp.JSON200.Address)

	return resp.JSON200, nil
}

// ListInput defines the input parameters for listing wallets.
//   - Name: Optional filter to search wallets by name (case-insensitive partial match).
//   - ChainSelector: Optional filter to search wallets by chain selector.
//   - Limit: Maximum number of wallets to return (1-50, default: 20).
//   - Offset: Number of wallets to skip for pagination (default: 0).
type ListInput struct {
	Name          *string
	ChainSelector *string
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

	params := apiClient.GetWalletsParams{
		Name:          input.Name,
		ChainSelector: input.ChainSelector,
		Limit:         input.Limit,
		Offset:        input.Offset,
	}

	resp, err := c.apiClient.GetWalletsWithResponse(ctx, &params)
	if err != nil {
		c.logger.Error("Failed to list wallets", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListWallets, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when listing wallets",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListWallets, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListWallets, ErrNilResponseBody)
	}

	c.logger.Debug("Wallets listed successfully",
		"count", len(resp.JSON200.Data),
		"has_more", resp.JSON200.HasMore)

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
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

	if input.Name == "" {
		return ErrNameRequired
	}

	if len(input.Name) > MaxWalletNameLength {
		return fmt.Errorf("%w: cannot exceed %d characters", ErrNameTooLong, MaxWalletNameLength)
	}

	updateWalletReq := apiClient.UpdateWallet{
		Name: input.Name,
	}

	resp, err := c.apiClient.PatchWalletsWalletIdWithResponse(ctx, walletID, updateWalletReq)
	if err != nil {
		c.logger.Error("Failed to update wallet", "error", err)
		return fmt.Errorf("%w: %w", ErrUpdateWallet, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Wallet not found", "wallet_id", walletID.String())
		return fmt.Errorf("%w: wallet ID %s", ErrWalletNotFound, walletID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when updating wallet",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return fmt.Errorf("%w: %w (status code %d)", ErrUpdateWallet, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	c.logger.Info("Wallet updated successfully", "wallet_id", walletID.String())

	return nil
}
