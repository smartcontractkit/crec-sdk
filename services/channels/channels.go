package channels

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

const (
	ServiceName = "channels"
)

// ServiceOptions defines the options for creating a new CREC Channels service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
type ServiceOptions struct {
	Logger     *zerolog.Logger
	CRECClient *client.CRECClient
}

// Service provides operations for managing CREC channels.
// Channels are logical groupings that allow customers to organize their watchers,
// events, and operations. Each channel is scoped to an organization.
type Service struct {
	logger     *zerolog.Logger
	crecClient *client.CRECClient
}

// NewService creates a new CREC Channels service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Channels service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, fmt.Errorf("ServiceOptions is required")
	}

	if opts.CRECClient == nil {
		return nil, fmt.Errorf("CRECClient is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC Channels service")

	return &Service{
		logger:     logger,
		crecClient: opts.CRECClient,
	}, nil
}

// CreateChannelInput defines the input parameters for creating a new channel.
//   - Name: The name of the channel. Must be unique within the organization.
type CreateChannelInput struct {
	Name string
}

// CreateChannel creates a new channel in the CREC backend.
// The channel will be scoped to the organization associated with the API key.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The channel creation parameters.
//
// Returns the created channel or an error if the operation fails.
func (s *Service) CreateChannel(ctx context.Context, input CreateChannelInput) (*apiClient.Channel, error) {
	s.logger.Debug().
		Str("name", input.Name).
		Msg("Creating channel")

	if input.Name == "" {
		return nil, fmt.Errorf("channel name is required")
	}

	createChannelReq := apiClient.CreateChannel{
		Name: input.Name,
	}

	resp, err := s.crecClient.PostChannelsWithResponse(ctx, createChannelReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create channel")
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when creating channel")
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body))
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected nil response body")
	}

	s.logger.Info().
		Str("channel_id", resp.JSON201.ChannelId.String()).
		Str("name", resp.JSON201.Name).
		Msg("Channel created successfully")

	return resp.JSON201, nil
}

// GetChannel retrieves a specific channel by its ID.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to retrieve.
//
// Returns the channel or an error if the operation fails or the channel is not found.
func (s *Service) GetChannel(ctx context.Context, channelID uuid.UUID) (*apiClient.Channel, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Msg("Getting channel")

	resp, err := s.crecClient.GetChannelsChannelIdWithResponse(ctx, channelID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get channel")
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", channelID.String()).
			Msg("Channel not found")
		return nil, fmt.Errorf("channel not found: %s", channelID.String())
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when getting channel")
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body))
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected nil response body")
	}

	s.logger.Debug().
		Str("channel_id", resp.JSON200.ChannelId.String()).
		Str("name", resp.JSON200.Name).
		Msg("Channel retrieved successfully")

	return resp.JSON200, nil
}

// ListChannelsInput defines the input parameters for listing channels.
//   - Name: Optional filter to search channels by name.
//   - Limit: Maximum number of channels to return (1-50, default: 20).
//   - Offset: Number of channels to skip for pagination (default: 0).
type ListChannelsInput struct {
	Name   *string
	Limit  *int
	Offset *int
}

// ListChannels retrieves a list of channels for the organization.
// Results are scoped to the organization associated with the API key.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The list parameters including filters and pagination.
//
// Returns a list of channels and a boolean indicating if there are more results.
func (s *Service) ListChannels(ctx context.Context, input ListChannelsInput) ([]apiClient.Channel, bool, error) {
	s.logger.Debug().
		Interface("filters", input).
		Msg("Listing channels")

	params := apiClient.GetChannelsParams{
		Name:   input.Name,
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	resp, err := s.crecClient.GetChannelsWithResponse(ctx, &params)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list channels")
		return nil, false, fmt.Errorf("failed to list channels: %w", err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when listing channels")
		return nil, false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body))
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("unexpected nil response body")
	}

	s.logger.Debug().
		Int("count", len(resp.JSON200.Data)).
		Bool("has_more", resp.JSON200.HasMore).
		Msg("Channels listed successfully")

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}

// DeleteChannel performs a logical delete of a channel.
// The channel will be marked as deleted but not physically removed from the database.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to delete.
//
// Returns an error if the operation fails or the channel is not found.
func (s *Service) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Msg("Deleting channel")

	resp, err := s.crecClient.DeleteChannelsChannelIdWithResponse(ctx, channelID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete channel")
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", channelID.String()).
			Msg("Channel not found")
		return fmt.Errorf("channel not found: %s", channelID.String())
	}

	if resp.StatusCode() != 202 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when deleting channel")
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body))
	}

	s.logger.Info().
		Str("channel_id", channelID.String()).
		Msg("Channel deleted successfully")

	return nil
}
