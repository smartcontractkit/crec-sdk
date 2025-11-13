package channels

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

const (
	MaxChannelNameLength = 255
)

// Sentinel errors
var (
	// ErrChannelNotFound is returned when a channel is not found (404 response)
	ErrChannelNotFound = errors.New("channel not found")

	// Service initialization errors
	ErrServiceOptionsRequired = errors.New("ServiceOptions is required")
	ErrCRECClientRequired     = errors.New("CRECClient is required")

	// Validation errors
	ErrChannelNameRequired = errors.New("channel name is required")
	ErrChannelNameTooLong  = errors.New("channel name too long")

	// API operation errors
	ErrCreateChannel = errors.New("failed to create channel")
	ErrGetChannel    = errors.New("failed to get channel")
	ErrListChannels  = errors.New("failed to list channels")
	ErrDeleteChannel = errors.New("failed to delete channel")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
)

// ServiceOptions defines the options for creating a new CREC Channels service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
type ServiceOptions struct {
	Logger     *zerolog.Logger
	CRECClient *client.CRECClient
}

// Service provides operations for managing CREC channels.
// Channels are logical groupings that allow you to organize your watchers,
// events, and operations.
type Service struct {
	logger     *zerolog.Logger
	crecClient *client.CRECClient
}

// NewService creates a new CREC Channels service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Channels service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, ErrServiceOptionsRequired
	}

	if opts.CRECClient == nil {
		return nil, ErrCRECClientRequired
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
//   - Name: The name of the channel. Must be unique.
type CreateChannelInput struct {
	Name string
}

// CreateChannel creates a new channel in the CREC backend.
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
		return nil, ErrChannelNameRequired
	}

	if len(input.Name) > MaxChannelNameLength {
		return nil, fmt.Errorf("%w: cannot exceed %d characters", ErrChannelNameTooLong, MaxChannelNameLength)
	}

	createChannelReq := apiClient.CreateChannel{
		Name: input.Name,
	}

	resp, err := s.crecClient.PostChannelsWithResponse(ctx, createChannelReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create channel")
		return nil, fmt.Errorf("%w: %w", ErrCreateChannel, err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when creating channel")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateChannel, ErrNilResponseBody)
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
		return nil, fmt.Errorf("%w: %w", ErrGetChannel, err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", channelID.String()).
			Msg("Channel not found")
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, channelID.String())
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when getting channel")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChannel, ErrNilResponseBody)
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

// ListChannels retrieves a list of channels.
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
		return nil, false, fmt.Errorf("%w: %w", ErrListChannels, err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when listing channels")
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListChannels, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListChannels, ErrNilResponseBody)
	}

	s.logger.Debug().
		Int("count", len(resp.JSON200.Data)).
		Bool("has_more", resp.JSON200.HasMore).
		Msg("Channels listed successfully")

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}

// DeleteChannel deletes a channel.
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
		return fmt.Errorf("%w: %w", ErrDeleteChannel, err)
	}

	if resp.StatusCode() != 202 && resp.StatusCode() != 204 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Unexpected status code when deleting channel")

		if resp.StatusCode() == 404 {
			s.logger.Warn().
				Str("channel_id", channelID.String()).
				Msg("Channel not found")
			return fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, channelID.String())
		}

		return fmt.Errorf("%w: %w (status code %d)", ErrDeleteChannel, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.StatusCode() == 202 {
		s.logger.Info().
			Str("channel_id", channelID.String()).
			Msg("Channel deletion initiated (async)")
	} else {
		s.logger.Info().
			Str("channel_id", channelID.String()).
			Msg("Channel deleted successfully (sync)")
	}

	return nil
}
