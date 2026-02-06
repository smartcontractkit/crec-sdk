package channels

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

const (
	MaxChannelNameLength = 255
)

// Sentinel errors
var (
	// ErrChannelNotFound is returned when a channel is not found (404 response)
	ErrChannelNotFound = errors.New("channel not found")

	// Client initialization errors
	ErrOptionsRequired   = errors.New("options is required")
	ErrAPIClientRequired = errors.New("APIClient is required")

	// Validation errors
	ErrChannelNameRequired = errors.New("channel name is required")
	ErrChannelNameTooLong  = errors.New("channel name too long")

	// API operation errors
	ErrCreateChannel = errors.New("failed to create channel")
	ErrGetChannel    = errors.New("failed to get channel")
	ErrListChannels  = errors.New("failed to list channels")
	ErrUpdateChannel = errors.New("failed to update channel")
	ErrDeleteChannel = errors.New("failed to delete channel")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
)

// Options defines the options for creating a new CREC Channels client.
//   - Logger: Optional logger instance.
//   - APIClient: The CREC API client instance.
type Options struct {
	Logger    *slog.Logger
	APIClient *apiClient.ClientWithResponses
}

// Client provides operations for managing CREC channels.
// Channels are logical groupings that allow you to organize your watchers,
// events, and operations.
type Client struct {
	logger    *slog.Logger
	apiClient *apiClient.ClientWithResponses
}

// NewClient creates a new CREC Channels client with the provided options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Channels client, see Options for details.
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

	logger.Debug("Creating CREC Channels client")

	return &Client{
		logger:    logger,
		apiClient: opts.APIClient,
	}, nil
}

// CreateInput defines the input parameters for creating a new channel.
//   - Name: The name of the channel. Must be unique.
//   - Description: Optional description of the channel.
type CreateInput struct {
	Name        string
	Description *string
}

// Create creates a new channel in the CREC backend.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The channel creation parameters.
//
// Returns the created channel or an error if the operation fails.
func (c *Client) Create(ctx context.Context, input CreateInput) (*apiClient.Channel, error) {
	c.logger.Debug("Creating channel", "name", input.Name)

	if input.Name == "" {
		return nil, ErrChannelNameRequired
	}

	if len(input.Name) > MaxChannelNameLength {
		return nil, fmt.Errorf("%w: cannot exceed %d characters", ErrChannelNameTooLong, MaxChannelNameLength)
	}

	createChannelReq := apiClient.CreateChannel{
		Name:        input.Name,
		Description: input.Description,
	}

	resp, err := c.apiClient.PostChannelsWithResponse(ctx, createChannelReq)
	if err != nil {
		c.logger.Error("Failed to create channel", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateChannel, err)
	}

	if resp.StatusCode() != 201 {
		c.logger.Error("Unexpected status code when creating channel",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateChannel, ErrNilResponseBody)
	}

	c.logger.Info("Channel created successfully",
		"channel_id", resp.JSON201.ChannelId.String(),
		"name", resp.JSON201.Name)

	return resp.JSON201, nil
}

// Get retrieves a specific channel by its ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to retrieve.
//
// Returns the channel or an error if the operation fails or the channel is not found.
func (c *Client) Get(ctx context.Context, channelID uuid.UUID) (*apiClient.Channel, error) {
	c.logger.Debug("Getting channel", "channel_id", channelID.String())

	resp, err := c.apiClient.GetChannelsChannelIdWithResponse(ctx, channelID)
	if err != nil {
		c.logger.Error("Failed to get channel", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetChannel, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found", "channel_id", channelID.String())
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, channelID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when getting channel",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChannel, ErrNilResponseBody)
	}

	c.logger.Debug("Channel retrieved successfully",
		"channel_id", resp.JSON200.ChannelId.String(),
		"name", resp.JSON200.Name)

	return resp.JSON200, nil
}

// ListInput defines the input parameters for listing channels.
//   - Name: Optional filter to search channels by name.
//   - Status: Optional filter to search channels by status (e.g., active, deactivated).
//   - Limit: Maximum number of channels to return (1-50, default: 20).
//   - Offset: Number of channels to skip for pagination (default: 0).
type ListInput struct {
	Name   *string
	Status *apiClient.ChannelStatus
	Limit  *int
	Offset *int64
}

// List retrieves a list of channels.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The list parameters including filters and pagination.
//
// Returns a list of channels and a boolean indicating if there are more results.
func (c *Client) List(ctx context.Context, input ListInput) ([]apiClient.Channel, bool, error) {
	c.logger.Debug("Listing channels", "filters", input)

	params := apiClient.GetChannelsParams{
		Name:   input.Name,
		Status: input.Status,
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	resp, err := c.apiClient.GetChannelsWithResponse(ctx, &params)
	if err != nil {
		c.logger.Error("Failed to list channels", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListChannels, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when listing channels",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListChannels, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListChannels, ErrNilResponseBody)
	}

	c.logger.Debug("Channels listed successfully",
		"count", len(resp.JSON200.Data),
		"has_more", resp.JSON200.HasMore)

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}

// UpdateInput defines the input parameters for updating a channel.
//   - Name: The new name for the channel.
//   - Description: The new description for the channel.
type UpdateInput struct {
	Name        string
	Description string
}

// Update updates a channel's information.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to update.
//   - input: The channel update parameters.
//
// Returns the updated channel or an error if the operation fails or the channel is not found.
func (c *Client) Update(ctx context.Context, channelID uuid.UUID, input UpdateInput) (*apiClient.Channel, error) {
	c.logger.Debug("Updating channel", "channel_id", channelID.String(), "name", input.Name)

	if input.Name == "" {
		return nil, ErrChannelNameRequired
	}

	if len(input.Name) > MaxChannelNameLength {
		return nil, fmt.Errorf("%w: cannot exceed %d characters", ErrChannelNameTooLong, MaxChannelNameLength)
	}

	updateChannelReq := apiClient.UpdateChannel{
		Name:        input.Name,
		Description: input.Description,
	}

	resp, err := c.apiClient.PutChannelsChannelIdWithResponse(ctx, channelID, updateChannelReq)
	if err != nil {
		c.logger.Error("Failed to update channel", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrUpdateChannel, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found", "channel_id", channelID.String())
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, channelID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when updating channel",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrUpdateChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrUpdateChannel, ErrNilResponseBody)
	}

	c.logger.Info("Channel updated successfully",
		"channel_id", resp.JSON200.ChannelId.String(),
		"name", resp.JSON200.Name)

	return resp.JSON200, nil
}

// Delete deletes a channel.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to delete.
//
// Returns an error if the operation fails or the channel is not found.
func (c *Client) Delete(ctx context.Context, channelID uuid.UUID) error {
	c.logger.Debug("Deleting channel", "channel_id", channelID.String())

	resp, err := c.apiClient.DeleteChannelsChannelIdWithResponse(ctx, channelID)
	if err != nil {
		c.logger.Error("Failed to delete channel", "error", err)
		return fmt.Errorf("%w: %w", ErrDeleteChannel, err)
	}

	if resp.StatusCode() != 202 && resp.StatusCode() != 204 {
		c.logger.Error("Unexpected status code when deleting channel",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))

		if resp.StatusCode() == 404 {
			c.logger.Warn("Channel not found", "channel_id", channelID.String())
			return fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, channelID.String())
		}

		return fmt.Errorf("%w: %w (status code %d)", ErrDeleteChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.StatusCode() == 202 {
		c.logger.Info("Channel deletion initiated (async)", "channel_id", channelID.String())
	} else {
		c.logger.Info("Channel deleted successfully (sync)", "channel_id", channelID.String())
	}

	return nil
}
