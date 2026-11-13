package wallets

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

// Catalog operation errors.
var (
	ErrListTypeVersions = errors.New("failed to list wallet type versions")
	ErrGetTypeVersion   = errors.New("failed to get wallet type version")
)

// ListTypeVersionsInput defines the filters for listing published wallet type
// versions.
//   - ChainSelector: Optional; filters each version's deployments to the chain.
//     Versions with no deployment on that chain are still returned, with an
//     empty deployments list.
//   - Status: Optional; filters versions by support state (e.g. active).
type ListTypeVersionsInput struct {
	ChainSelector *string
	Status        *apiClient.WalletTypeVersionStatus
}

// ListTypeVersions retrieves the published implementation versions of a wallet
// type. Consumers use it to discover upgrade targets and published
// capabilities; filtering to the wallet's chain selector returns the
// implementation addresses to inspect and sign.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletType: The wallet type (e.g., "ecdsa", "protected_ecdsa").
//   - input: The list filters.
//
// Returns the versions and a boolean indicating if there are more results.
func (c *Client) ListTypeVersions(
	ctx context.Context,
	walletType apiClient.WalletType,
	input ListTypeVersionsInput,
) ([]apiClient.WalletTypeVersion, bool, error) {
	c.logger.Debug("Listing wallet type versions", "wallet_type", walletType, "filters", input)

	if walletType == "" {
		return nil, false, ErrWalletTypeRequired
	}

	params := apiClient.ListWalletTypeVersionsParams{
		ChainSelector: input.ChainSelector,
		Status:        input.Status,
	}

	resp, err := c.apiClient.ListWalletTypeVersionsWithResponse(ctx, walletType, &params)
	if err != nil {
		c.logger.Error("Failed to list wallet type versions", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListTypeVersions, err)
	}

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListTypeVersions, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, false, fmt.Errorf("%w: %w", ErrListTypeVersions, apierror.ErrNilResponseBody)
		}
		c.logger.Debug("Wallet type versions listed successfully",
			"wallet_type", walletType,
			"count", len(resp.JSON200.Data))
		return resp.JSON200.Data, resp.JSON200.HasMore, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "listing wallet type versions", nil),
			"wallet_type", walletType,
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, false, apierror.WrapNotFound(resp.JSON404, ErrListTypeVersions, "wallet type "+string(walletType))
	default:
		return nil, false, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrListTypeVersions, "listing wallet type versions", resp.Body, c.logger)
	}
}

// GetTypeVersion retrieves one published implementation version of a wallet
// type, with its deployments and capability descriptors. This is the fact
// fetch before building and signing an upgrade or administrative
// authorization.
//
// Parameters:
//   - ctx: The context for the request.
//   - walletType: The wallet type (e.g., "ecdsa", "protected_ecdsa").
//   - implementationVersion: The operator-assigned version label.
//
// Returns the version or an error if the operation fails or the version is
// not published for the type.
func (c *Client) GetTypeVersion(
	ctx context.Context,
	walletType apiClient.WalletType,
	implementationVersion string,
) (*apiClient.WalletTypeVersion, error) {
	c.logger.Debug("Getting wallet type version",
		"wallet_type", walletType,
		"implementation_version", implementationVersion)

	if walletType == "" {
		return nil, ErrWalletTypeRequired
	}
	if implementationVersion == "" {
		return nil, ErrVersionRequired
	}

	resp, err := c.apiClient.GetWalletTypeVersionWithResponse(ctx, walletType, implementationVersion)
	if err != nil {
		c.logger.Error("Failed to get wallet type version", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetTypeVersion, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetTypeVersion, apierror.ErrNilResponse)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("%w: %w", ErrGetTypeVersion, apierror.ErrNilResponseBody)
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		c.logger.Warn(
			apierror.NotFoundWarnMessage(resp.JSON404, "getting wallet type version", nil),
			"wallet_type", walletType,
			"implementation_version", implementationVersion,
			"code", apierror.NotFoundCode(resp.JSON404),
		)
		return nil, apierror.WrapNotFound(resp.JSON404, ErrGetTypeVersion,
			fmt.Sprintf("wallet type %s version %s", walletType, implementationVersion))
	default:
		return nil, apierror.HandleErrorStatus(resp.StatusCode(), resp.JSON401, resp.JSON403, ErrGetTypeVersion, "getting wallet type version", resp.Body, c.logger)
	}
}
