package networks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// Sentinel errors
var (
	// Client initialization errors
	ErrOptionsRequired   = errors.New("options is required")
	ErrAPIClientRequired = errors.New("APIClient is required")

	// API operation errors
	ErrListNetworks = errors.New("failed to list networks")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
)

// Options defines the options for creating a new CREC Networks client.
//   - Logger: Optional logger instance.
//   - APIClient: The CREC API client instance.
type Options struct {
	Logger    *slog.Logger
	APIClient *apiClient.ClientWithResponses
}

// Client provides operations for listing available CREC networks.
type Client struct {
	logger    *slog.Logger
	apiClient *apiClient.ClientWithResponses
}

// NewClient creates a new CREC Networks client with the provided options.
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

	return &Client{
		logger:    logger,
		apiClient: opts.APIClient,
	}, nil
}

// List retrieves the list of available networks supported by the CREC platform.
//
// Parameters:
//   - ctx: The context for the request.
//
// Returns the list of networks, a boolean indicating if there are more results (for future pagination),
// and an error if the operation fails.
func (c *Client) List(ctx context.Context) ([]apiClient.Network, bool, error) {
	c.logger.Debug("Listing networks")

	resp, err := c.apiClient.GetNetworksWithResponse(ctx)
	if err != nil {
		c.logger.Error("Failed to list networks", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListNetworks, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when listing networks",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListNetworks, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListNetworks, ErrNilResponseBody)
	}

	c.logger.Debug("Networks listed successfully", "count", len(resp.JSON200.Data))

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}
