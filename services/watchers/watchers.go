package watchers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

type WatcherStatus string

const (
	StatusPending  WatcherStatus = "pending"
	StatusActive   WatcherStatus = "active"
	StatusFailed   WatcherStatus = "failed"
	StatusDeleting WatcherStatus = "deleting"
	StatusDeleted  WatcherStatus = "deleted"
)

type EventABIInput struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

type EventABI struct {
	Anonymous bool            `json:"anonymous"`
	Inputs    []EventABIInput `json:"inputs"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
}

type CreateWatcherWithDomainInput struct {
	Name          *string  `json:"name,omitempty"`
	ChainSelector uint64   `json:"chain_selector"`
	Address       string   `json:"address"`
	Domain        string   `json:"domain"`
	Events        []string `json:"events"`
}

type CreateWatcherWithABIInput struct {
	Name          *string    `json:"name,omitempty"`
	ChainSelector uint64     `json:"chain_selector"`
	Address       string     `json:"address"`
	Events        []string   `json:"events"`
	ABI           []EventABI `json:"abi"`
}

type UpdateWatcherInput struct {
	Name string `json:"name"`
}

type WatcherFilters struct {
	Limit         *int           `url:"limit,omitempty"`
	Offset        *int           `url:"offset,omitempty"`
	Name          *string        `url:"name,omitempty"`
	Status        *WatcherStatus `url:"status,omitempty"`
	ChainSelector *uint64        `url:"chain_selector,omitempty"`
	Address       *string        `url:"address,omitempty"`
	Domain        *string        `url:"domain,omitempty"`
	EventName     *string        `url:"event_name,omitempty"`
}

const (
	ServiceName = "watchers"
)

// ServiceOptions defines the options for creating a new CREC Watchers service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
type ServiceOptions struct {
	Logger     *zerolog.Logger
	CRECClient *client.CRECClient
}

// Service provides operations for managing CREC watchers.
// Watchers monitor blockchain events on specific smart contracts and trigger workflows.
type Service struct {
	logger     *zerolog.Logger
	crecClient *client.CRECClient
}

// NewService creates a new CREC Watchers service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Watchers service, see ServiceOptions for details.
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

	logger.Debug().Msg("Creating CREC Watchers service")

	return &Service{
		logger:     logger,
		crecClient: opts.CRECClient,
	}, nil
}

// CreateWatcherWithDomain creates a new watcher with a domain (dvp, dta, test_consumer)
func (s *Service) CreateWatcherWithDomain(ctx context.Context, channelID uuid.UUID, input CreateWatcherWithDomainInput) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("domain", input.Domain).
		Str("address", input.Address).
		Msg("Creating watcher with domain")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}
	if input.ChainSelector == 0 {
		return nil, fmt.Errorf("chain_selector is required")
	}
	if input.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if input.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if len(input.Events) == 0 {
		return nil, fmt.Errorf("events list cannot be empty")
	}

	createWatcherWithDomain := apiClient.CreateWatcherWithDomain{
		Name:          input.Name,
		ChainSelector: input.ChainSelector,
		Address:       input.Address,
		Domain:        input.Domain,
		Events:        input.Events,
	}

	var createWatcherReq apiClient.CreateWatcher
	if err := createWatcherReq.FromCreateWatcherWithDomain(createWatcherWithDomain); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher request")
		return nil, fmt.Errorf("failed to create watcher request: %w", err)
	}

	resp, err := s.crecClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher with domain")
		return nil, fmt.Errorf("failed to create watcher with domain: %w", err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to create watcher with domain - unexpected status code")
		return nil, fmt.Errorf("failed to create watcher with domain: status code %d", resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected empty response from create watcher with domain")
	}

	s.logger.Info().
		Str("watcher_id", resp.JSON201.WatcherId.String()).
		Msg("Watcher created successfully")

	return resp.JSON201, nil
}

// CreateWatcherWithABI creates a new watcher with custom event ABI
func (s *Service) CreateWatcherWithABI(ctx context.Context, channelID uuid.UUID, input CreateWatcherWithABIInput) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("address", input.Address).
		Int("abi_count", len(input.ABI)).
		Msg("Creating watcher with ABI")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}
	if input.ChainSelector == 0 {
		return nil, fmt.Errorf("chain_selector is required")
	}
	if input.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if len(input.Events) == 0 {
		return nil, fmt.Errorf("events list cannot be empty")
	}
	if len(input.ABI) == 0 {
		return nil, fmt.Errorf("abi cannot be empty")
	}

	// Convert EventABI to apiClient.EventABI
	abiList := make([]apiClient.EventABI, len(input.ABI))
	for i, abi := range input.ABI {
		inputs := make([]apiClient.EventABIInput, len(abi.Inputs))
		for j, input := range abi.Inputs {
			inputs[j] = apiClient.EventABIInput{
				Indexed:      input.Indexed,
				InternalType: input.InternalType,
				Name:         input.Name,
				Type:         input.Type,
			}
		}
		abiList[i] = apiClient.EventABI{
			Anonymous: abi.Anonymous,
			Inputs:    inputs,
			Name:      abi.Name,
			Type:      apiClient.EventABITypeEvent, // Use the constant from the generated client
		}
	}

	createWatcherWithABI := apiClient.CreateWatcherWithABI{
		Name:          input.Name,
		ChainSelector: input.ChainSelector,
		Address:       input.Address,
		Events:        input.Events,
		Abi:           abiList,
	}

	var createWatcherReq apiClient.CreateWatcher
	if err := createWatcherReq.FromCreateWatcherWithABI(createWatcherWithABI); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher request")
		return nil, fmt.Errorf("failed to create watcher request: %w", err)
	}

	resp, err := s.crecClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher with ABI")
		return nil, fmt.Errorf("failed to create watcher with ABI: %w", err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to create watcher with ABI - unexpected status code")
		return nil, fmt.Errorf("failed to create watcher with ABI: status code %d", resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("unexpected empty response from create watcher with ABI")
	}

	s.logger.Info().
		Str("watcher_id", resp.JSON201.WatcherId.String()).
		Msg("Watcher created successfully")

	return resp.JSON201, nil
}

// FindWatchersByChannel retrieves watchers for a specific channel with optional filters
func (s *Service) FindWatchersByChannel(ctx context.Context, channelID uuid.UUID, filters WatcherFilters) (*apiClient.WatcherList, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Msg("Finding watchers by channel")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}

	params := &apiClient.GetChannelsChannelIdWatchersParams{
		Limit:         filters.Limit,
		Offset:        filters.Offset,
		Name:          filters.Name,
		ChainSelector: convertUint64PtrToStringPtr(filters.ChainSelector),
		Address:       filters.Address,
		Domain:        filters.Domain,
		EventName:     filters.EventName,
	}

	if filters.Status != nil {
		statusStr := string(*filters.Status)
		params.Status = &statusStr
	}

	resp, err := s.crecClient.GetChannelsChannelIdWatchersWithResponse(ctx, channelID, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to find watchers")
		return nil, fmt.Errorf("failed to find watchers: %w", err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to find watchers - unexpected status code")
		return nil, fmt.Errorf("failed to find watchers: status code %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response from find watchers")
	}

	s.logger.Info().
		Int("count", len(resp.JSON200.Data)).
		Msg("Watchers found")

	return resp.JSON200, nil
}

// FindWatcherByID retrieves a specific watcher by channel and watcher ID
func (s *Service) FindWatcherByID(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("watcher_id", watcherID.String()).
		Msg("Finding watcher by ID")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf("watcher_id cannot be empty")
	}

	resp, err := s.crecClient.GetChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to find watcher")
		return nil, fmt.Errorf("failed to find watcher: %w", err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to find watcher - unexpected status code")
		return nil, fmt.Errorf("failed to find watcher: status code %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response from find watcher")
	}

	return resp.JSON200, nil
}

// UpdateWatcher updates a watcher's name
func (s *Service) UpdateWatcher(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, input UpdateWatcherInput) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("watcher_id", watcherID.String()).
		Str("name", input.Name).
		Msg("Updating watcher")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf("watcher_id cannot be empty")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	updateReq := apiClient.UpdateWatcher{
		Name: input.Name,
	}

	resp, err := s.crecClient.PatchChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID, updateReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update watcher")
		return nil, fmt.Errorf("failed to update watcher: %w", err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to update watcher - unexpected status code")
		return nil, fmt.Errorf("failed to update watcher: status code %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected empty response from update watcher")
	}

	s.logger.Info().
		Str("watcher_id", watcherID.String()).
		Msg("Watcher updated successfully")

	return resp.JSON200, nil
}

// WaitForActive polls a watcher until it reaches active status or fails
func (s *Service) WaitForActive(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, maxWaitTime time.Duration) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("watcher_id", watcherID.String()).
		Dur("max_wait_time", maxWaitTime).
		Msg("Waiting for watcher to become active")

	if channelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id cannot be empty")
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf("watcher_id cannot be empty")
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.logger.Error().Msg("Timeout waiting for watcher to become active")
			return nil, fmt.Errorf("timeout waiting for watcher to become active")
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			watcher, err := s.FindWatcherByID(ctx, channelID, watcherID)
			if err != nil {
				return nil, fmt.Errorf("failed to check watcher status: %w", err)
			}

			switch watcher.Status {
			case "active":
				s.logger.Info().Msg("Watcher is now active")
				return watcher, nil
			case "failed":
				s.logger.Error().Msg("Watcher deployment failed")
				return nil, fmt.Errorf("watcher deployment failed")
			case "pending":
				// Continue waiting
				s.logger.Debug().Msg("Watcher still pending, continuing to wait")
				continue
			default:
				return nil, fmt.Errorf("unexpected watcher status: %s", watcher.Status)
			}
		}
	}
}

// convertUint64PtrToStringPtr converts a pointer to uint64 to a pointer to string
func convertUint64PtrToStringPtr(val *uint64) *string {
	if val == nil {
		return nil
	}
	str := fmt.Sprintf("%d", *val)
	return &str
}
