package watchers

import (
	"context"
	"fmt"
	"os"
	"strings"
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

	// Validation error messages
	errChannelIDRequired = "channel_id cannot be empty"
	errWatcherIDRequired = "watcher_id cannot be empty"
	errNameRequired      = "name is required"
	errDomainRequired    = "domain is required"
	errAddressRequired   = "address is required"
	errEventsRequired    = "events list cannot be empty"
	errABIRequired       = "abi cannot be empty"
)

// ServiceOptions defines the options for creating a new CREC Watchers service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
//   - PollInterval: Optional polling interval for WaitForActive. Defaults to 2 seconds if not set.
type ServiceOptions struct {
	Logger       *zerolog.Logger
	CRECClient   *client.CRECClient
	PollInterval time.Duration
}

// Service provides operations for managing CREC watchers.
// Watchers monitor blockchain events on specific smart contracts and trigger workflows.
type Service struct {
	logger       *zerolog.Logger
	crecClient   *client.CRECClient
	pollInterval time.Duration
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

	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	logger.Debug().
		Dur("poll_interval", pollInterval).
		Msg("Creating CREC Watchers service")

	return &Service{
		logger:       logger,
		crecClient:   opts.CRECClient,
		pollInterval: pollInterval,
	}, nil
}

// CreateWatcherWithDomain creates a new watcher with a domain (dvp, dta, test_consumer)
func (s *Service) CreateWatcherWithDomain(ctx context.Context, channelID uuid.UUID, input CreateWatcherWithDomainInput) (*apiClient.Watcher, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("domain", input.Domain).
		Str("address", input.Address).
		Msg("Creating watcher with domain")

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if input.ChainSelector == 0 {
		return nil, fmt.Errorf("chain_selector is required")
	}
	if input.Address == "" {
		return nil, fmt.Errorf(errAddressRequired)
	}
	if input.Domain == "" {
		return nil, fmt.Errorf(errDomainRequired)
	}
	if len(input.Events) == 0 {
		return nil, fmt.Errorf(errEventsRequired)
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

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if input.ChainSelector == 0 {
		return nil, fmt.Errorf("chain_selector is required")
	}
	if input.Address == "" {
		return nil, fmt.Errorf(errAddressRequired)
	}
	if len(input.Events) == 0 {
		return nil, fmt.Errorf(errEventsRequired)
	}
	if len(input.ABI) == 0 {
		return nil, fmt.Errorf(errABIRequired)
	}

	abiList := make([]apiClient.EventABI, len(input.ABI))
	for i, abi := range input.ABI {
		inputs := make([]apiClient.EventABIInput, len(abi.Inputs))
		for j, abiInput := range abi.Inputs {
			inputs[j] = apiClient.EventABIInput{
				Indexed:      abiInput.Indexed,
				InternalType: abiInput.InternalType,
				Name:         abiInput.Name,
				Type:         abiInput.Type,
			}
		}
		abiList[i] = apiClient.EventABI{
			Anonymous: abi.Anonymous,
			Inputs:    inputs,
			Name:      abi.Name,
			Type:      apiClient.EventABITypeEvent,
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

	if err := validateChannelID(channelID); err != nil {
		return nil, err
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

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf(errWatcherIDRequired)
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

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf(errWatcherIDRequired)
	}
	if input.Name == "" {
		return nil, fmt.Errorf(errNameRequired)
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

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, fmt.Errorf(errWatcherIDRequired)
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(s.pollInterval)
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

			switch WatcherStatus(watcher.Status) {
			case StatusActive:
				s.logger.Info().Msg("Watcher is now active")
				return watcher, nil
			case StatusFailed:
				s.logger.Error().Msg("Watcher deployment failed")
				return nil, fmt.Errorf("watcher deployment failed")
			case StatusPending:
				// Continue waiting
				s.logger.Debug().Msg("Watcher still pending, continuing to wait")
				continue
			case StatusDeleting:
				s.logger.Error().Msg("Watcher is being deleted")
				return nil, fmt.Errorf("watcher is being deleted and cannot become active")
			case StatusDeleted:
				s.logger.Error().Msg("Watcher has been deleted")
				return nil, fmt.Errorf("watcher has been deleted and cannot become active")
			default:
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Unexpected watcher status while waiting for active")
				return nil, fmt.Errorf("unexpected watcher status while waiting for active: %s", watcher.Status)
			}
		}
	}
}

// DeleteWatcher deletes a watcher from a channel
func (s *Service) DeleteWatcher(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID) error {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("watcher_id", watcherID.String()).
		Msg("Deleting watcher")

	if err := validateChannelID(channelID); err != nil {
		return err
	}
	if watcherID == uuid.Nil {
		return fmt.Errorf(errWatcherIDRequired)
	}

	resp, err := s.crecClient.DeleteChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete watcher")
		return fmt.Errorf("failed to delete watcher: %w", err)
	}

	// Accept both 202 (Accepted - async deletion) and 204 (No Content - sync deletion)
	if resp.StatusCode() != 202 && resp.StatusCode() != 204 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to delete watcher - unexpected status code")
		return fmt.Errorf("failed to delete watcher: status code %d", resp.StatusCode())
	}

	if resp.StatusCode() == 202 {
		s.logger.Info().
			Str("watcher_id", watcherID.String()).
			Msg("Watcher deletion initiated (async)")
	} else {
		s.logger.Info().
			Str("watcher_id", watcherID.String()).
			Msg("Watcher deleted successfully (sync)")
	}

	return nil
}

// WaitForDeleted waits for a watcher to be fully deleted.
// The method polls the watcher status until it reaches "deleted" state or the timeout is reached.
func (s *Service) WaitForDeleted(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, maxWaitTime time.Duration) error {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("watcher_id", watcherID.String()).
		Dur("max_wait_time", maxWaitTime).
		Msg("Waiting for watcher to be deleted")

	if err := validateChannelID(channelID); err != nil {
		return err
	}
	if watcherID == uuid.Nil {
		return fmt.Errorf(errWatcherIDRequired)
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.logger.Error().Msg("Timeout waiting for watcher to be deleted")
			return fmt.Errorf("timeout waiting for watcher to be deleted")
		case <-ticker.C:
			watcher, err := s.FindWatcherByID(ctx, channelID, watcherID)
			if err != nil {
				// If watcher is not found (404), it's been deleted
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
					s.logger.Info().Msg("Watcher has been deleted")
					return nil
				}
				return fmt.Errorf("failed to check watcher status: %w", err)
			}

			switch WatcherStatus(watcher.Status) {
			case StatusDeleted:
				s.logger.Info().Msg("Watcher is now deleted")
				return nil
			case StatusDeleting:
				s.logger.Debug().Msg("Watcher is being deleted, continuing to wait")
				continue
			case StatusActive, StatusPending, StatusFailed:
				// If the watcher is in any other valid state, it means deletion was rolled back or failed
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Watcher deletion appears to have failed")
				return fmt.Errorf("watcher deletion failed, watcher is in %s state", watcher.Status)
			default:
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Unexpected watcher status while waiting for deletion")
				return fmt.Errorf("unexpected watcher status while waiting for deletion: %s", watcher.Status)
			}
		}
	}
}

// validateChannelID validates that channel ID is not empty
func validateChannelID(channelID uuid.UUID) error {
	if channelID == uuid.Nil {
		return fmt.Errorf(errChannelIDRequired)
	}
	return nil
}

// convertUint64PtrToStringPtr converts a pointer to uint64 to a pointer to string
func convertUint64PtrToStringPtr(val *uint64) *string {
	if val == nil {
		return nil
	}
	str := fmt.Sprintf("%d", *val)
	return &str
}
