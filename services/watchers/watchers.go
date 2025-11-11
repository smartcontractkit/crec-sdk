package watchers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

var (
	// ErrWatcherNotFound is returned when a watcher is not found (404 response)
	ErrWatcherNotFound = errors.New("watcher not found")

	// Validation errors
	ErrChannelIDRequired      = errors.New("channel_id cannot be empty")
	ErrWatcherIDRequired      = errors.New("watcher_id cannot be empty")
	ErrNameRequired           = errors.New("name is required")
	ErrDomainRequired         = errors.New("domain is required")
	ErrAddressRequired        = errors.New("address is required")
	ErrEventsRequired         = errors.New("events list cannot be empty")
	ErrABIRequired            = errors.New("abi cannot be empty")
	ErrServiceOptionsRequired = errors.New("ServiceOptions is required")
	ErrCRECClientRequired     = errors.New("CRECClient is required")
	ErrChainSelectorRequired  = errors.New("chain_selector is required")
	ErrInvalidABIType         = errors.New("invalid ABI type, only 'event' is currently supported")
	ErrEventNotInABI          = errors.New("event not found in provided ABI")

	// Timeout errors
	ErrWaitForActiveTimeout  = errors.New("timeout waiting for watcher to become active")
	ErrWaitForDeletedTimeout = errors.New("timeout waiting for watcher to be deleted")

	// Watcher state errors
	ErrWatcherDeploymentFailed = errors.New("watcher deployment failed")
	ErrWatcherIsDeleting       = errors.New("watcher is being deleted and cannot become active")
	ErrWatcherAlreadyDeleted   = errors.New("watcher has been deleted and cannot become active")
	ErrWatcherDeletionFailed   = errors.New("watcher deletion failed")

	// API response errors
	ErrEmptyResponse    = errors.New("unexpected empty response from API")
	ErrUnexpectedStatus = errors.New("unexpected watcher status")

	// API operation errors (base errors for wrapping)
	ErrCreateWatcherRequest = errors.New("failed to create watcher request")
	ErrCreateWatcherDomain  = errors.New("failed to create watcher with domain")
	ErrCreateWatcherABI     = errors.New("failed to create watcher with ABI")
	ErrFindWatchers         = errors.New("failed to find watchers")
	ErrFindWatcher          = errors.New("failed to find watcher")
	ErrUpdateWatcher        = errors.New("failed to update watcher")
	ErrDeleteWatcher        = errors.New("failed to delete watcher")
	ErrCheckWatcherStatus   = errors.New("failed to check watcher status")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

const (
	MaxLogBodyLength = 200
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

// ServiceOptions defines the options for creating a new CREC Watchers service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
//   - PollInterval: Optional polling interval for WaitForActive. Defaults to 2 seconds if not set.
//   - EventualConsistencyWindow: Optional duration to tolerate 404 errors after creation due to eventual consistency. Defaults to 2 seconds if not set.
type ServiceOptions struct {
	Logger                    *zerolog.Logger
	CRECClient                *client.CRECClient
	PollInterval              time.Duration
	EventualConsistencyWindow time.Duration
}

// Service provides operations for managing CREC watchers.
// Watchers monitor blockchain events on specific smart contracts and trigger workflows.
type Service struct {
	logger                    *zerolog.Logger
	crecClient                *client.CRECClient
	pollInterval              time.Duration
	eventualConsistencyWindow time.Duration
}

// NewService creates a new CREC Watchers service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Watchers service, see ServiceOptions for details.
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

	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	eventualConsistencyWindow := opts.EventualConsistencyWindow
	if eventualConsistencyWindow == 0 {
		eventualConsistencyWindow = 2 * time.Second
	}

	logger.Debug().
		Dur("poll_interval", pollInterval).
		Dur("eventual_consistency_window", eventualConsistencyWindow).
		Msg("Creating CREC Watchers service")

	return &Service{
		logger:                    logger,
		crecClient:                opts.CRECClient,
		pollInterval:              pollInterval,
		eventualConsistencyWindow: eventualConsistencyWindow,
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
		return nil, ErrChainSelectorRequired
	}
	if input.Address == "" {
		return nil, ErrAddressRequired
	}
	if input.Domain == "" {
		return nil, ErrDomainRequired
	}
	if len(input.Events) == 0 {
		return nil, ErrEventsRequired
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
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherRequest, err)
	}

	resp, err := s.crecClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher with domain")
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherDomain, err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to create watcher with domain - unexpected status code")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWatcherDomain, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, ErrEmptyResponse
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
		return nil, ErrChainSelectorRequired
	}
	if input.Address == "" {
		return nil, ErrAddressRequired
	}
	if len(input.Events) == 0 {
		return nil, ErrEventsRequired
	}
	if len(input.ABI) == 0 {
		return nil, ErrABIRequired
	}

	// Validate that all ABI entries are of type "event"
	for i, abi := range input.ABI {
		if abi.Type != "event" {
			s.logger.Error().
				Int("abi_index", i).
				Str("abi_type", abi.Type).
				Str("abi_name", abi.Name).
				Msg("Invalid ABI type")
			return nil, fmt.Errorf("%w: got '%s' for event '%s' at index %d", ErrInvalidABIType, abi.Type, abi.Name, i)
		}
	}

	// Validate that all requested events exist in the provided ABI
	abiEventMap := make(map[string]bool)
	for _, abi := range input.ABI {
		abiEventMap[abi.Name] = true
	}

	for _, eventName := range input.Events {
		if !abiEventMap[eventName] {
			s.logger.Error().
				Str("event_name", eventName).
				Int("abi_count", len(input.ABI)).
				Msg("Requested event not found in provided ABI")
			return nil, fmt.Errorf("%w: event '%s' not found in ABI definitions", ErrEventNotInABI, eventName)
		}
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
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherRequest, err)
	}

	resp, err := s.crecClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create watcher with ABI")
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherABI, err)
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to create watcher with ABI - unexpected status code")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWatcherABI, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, ErrEmptyResponse
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
		return nil, fmt.Errorf("%w: %w", ErrFindWatchers, err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to find watchers - unexpected status code")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrFindWatchers, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
	}

	s.logger.Debug().
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
		return nil, ErrWatcherIDRequired
	}

	resp, err := s.crecClient.GetChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to find watcher")
		return nil, fmt.Errorf("%w: %w", ErrFindWatcher, err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to find watcher - unexpected status code")

		if resp.StatusCode() == 404 {
			return nil, fmt.Errorf("%w (status code %d)", ErrWatcherNotFound, resp.StatusCode())
		}

		return nil, fmt.Errorf("%w: %w (status code %d)", ErrFindWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
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
		return nil, ErrWatcherIDRequired
	}
	if input.Name == "" {
		return nil, ErrNameRequired
	}

	updateReq := apiClient.UpdateWatcher{
		Name: input.Name,
	}

	resp, err := s.crecClient.PatchChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID, updateReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to update watcher")
		return nil, fmt.Errorf("%w: %w", ErrUpdateWatcher, err)
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to update watcher - unexpected status code")

		if resp.StatusCode() == 404 {
			return nil, fmt.Errorf("%w (status code %d)", ErrWatcherNotFound, resp.StatusCode())
		}

		return nil, fmt.Errorf("%w: %w (status code %d)", ErrUpdateWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
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
		Dur("eventual_consistency_window", s.eventualConsistencyWindow).
		Msg("Waiting for watcher to become active")

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, ErrWatcherIDRequired
	}

	startTime := time.Now()
	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.logger.Error().Msg("Timeout waiting for watcher to become active")
			return nil, ErrWaitForActiveTimeout
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			watcher, err := s.FindWatcherByID(ctx, channelID, watcherID)
			if err != nil {
				// During the eventual consistency window, tolerate 404 errors
				// This handles the case where a watcher was just created but isn't immediately visible
				if errors.Is(err, ErrWatcherNotFound) {
					elapsedTime := time.Since(startTime)
					if elapsedTime < s.eventualConsistencyWindow {
						s.logger.Debug().
							Dur("elapsed", elapsedTime).
							Dur("window", s.eventualConsistencyWindow).
							Msg("Watcher not found yet, retrying (eventual consistency)")
						continue
					}
					// After the window, 404 is a permanent error
					s.logger.Error().
						Dur("elapsed", elapsedTime).
						Msg("Watcher not found after eventual consistency window")
					return nil, fmt.Errorf("%w: %w", ErrCheckWatcherStatus, err)
				}

				// Check if error is transient (network issues, 5xx, etc.)
				if isTransientError(err) {
					s.logger.Warn().
						Err(err).
						Msg("Transient error while checking watcher status, will retry")
					continue
				}
				// Permanent error - fail immediately
				return nil, fmt.Errorf("%w: %w", ErrCheckWatcherStatus, err)
			}

			switch WatcherStatus(watcher.Status) {
			case StatusActive:
				s.logger.Info().Msg("Watcher is now active")
				return watcher, nil
			case StatusFailed:
				s.logger.Error().Msg("Watcher deployment failed")
				return nil, ErrWatcherDeploymentFailed
			case StatusPending:
				// Continue waiting
				s.logger.Debug().Msg("Watcher still pending, continuing to wait")
				continue
			case StatusDeleting:
				s.logger.Error().Msg("Watcher is being deleted")
				return nil, ErrWatcherIsDeleting
			case StatusDeleted:
				s.logger.Error().Msg("Watcher has been deleted")
				return nil, ErrWatcherAlreadyDeleted
			default:
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Unexpected watcher status while waiting for active")
				return nil, fmt.Errorf("%w: %s", ErrUnexpectedStatus, watcher.Status)
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
		return ErrWatcherIDRequired
	}

	resp, err := s.crecClient.DeleteChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to delete watcher")
		return fmt.Errorf("%w: %w", ErrDeleteWatcher, err)
	}

	// Accept both 202 (Accepted - async deletion) and 204 (No Content - sync deletion)
	if resp.StatusCode() != 202 && resp.StatusCode() != 204 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Failed to delete watcher - unexpected status code")

		if resp.StatusCode() == 404 {
			return fmt.Errorf("%w (status code %d)", ErrWatcherNotFound, resp.StatusCode())
		}

		return fmt.Errorf("%w: %w (status code %d)", ErrDeleteWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
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
		return ErrWatcherIDRequired
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			s.logger.Error().Msg("Timeout waiting for watcher to be deleted")
			return ErrWaitForDeletedTimeout
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			watcher, err := s.FindWatcherByID(ctx, channelID, watcherID)
			if err != nil {
				// If watcher is not found (404), it's been deleted
				if errors.Is(err, ErrWatcherNotFound) {
					s.logger.Debug().Msg("Watcher has been deleted (404 not found)")
					return nil
				}
				// Check if error is transient (network issues, 5xx, etc.)
				if isTransientError(err) {
					s.logger.Warn().
						Err(err).
						Msg("Transient error while checking watcher deletion status, will retry")
					continue
				}
				// Permanent error - fail immediately
				return fmt.Errorf("%w: %w", ErrCheckWatcherStatus, err)
			}

			switch WatcherStatus(watcher.Status) {
			case StatusDeleted:
				s.logger.Debug().Msg("Watcher is now deleted (status confirmed)")
				return nil
			case StatusDeleting:
				s.logger.Debug().Msg("Watcher is being deleted, continuing to wait")
				continue
			case StatusActive, StatusPending, StatusFailed:
				// If the watcher is in any other valid state, it means deletion was rolled back or failed
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Watcher deletion appears to have failed")
				return fmt.Errorf("%w, watcher is in %s state", ErrWatcherDeletionFailed, watcher.Status)
			default:
				s.logger.Error().
					Str("status", watcher.Status).
					Msg("Unexpected watcher status while waiting for deletion")
				return fmt.Errorf("%w while waiting for deletion: %s", ErrUnexpectedStatus, watcher.Status)
			}
		}
	}
}

// validateChannelID validates that channel ID is not empty
func validateChannelID(channelID uuid.UUID) error {
	if channelID == uuid.Nil {
		return ErrChannelIDRequired
	}
	return nil
}

// truncateBody truncates a response body to MaxLogBodyLength characters to avoid
// logging sensitive data or creating overly verbose logs.
func truncateBody(body []byte) string {
	bodyStr := string(body)
	if len(bodyStr) <= MaxLogBodyLength {
		return bodyStr
	}
	return bodyStr[:MaxLogBodyLength] + "... (truncated)"
}

// isTransientError determines if an error is transient and should be retried during polling.
// It checks for known transient HTTP status codes and network errors.
func isTransientError(err error) bool {
	// Validation errors are permanent
	if errors.Is(err, ErrChannelIDRequired) ||
		errors.Is(err, ErrWatcherIDRequired) ||
		errors.Is(err, ErrNameRequired) ||
		errors.Is(err, ErrDomainRequired) ||
		errors.Is(err, ErrAddressRequired) ||
		errors.Is(err, ErrEventsRequired) ||
		errors.Is(err, ErrABIRequired) {
		return false
	}

	// Context cancellation errors are permanent
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	errMsg := err.Error()

	// First, try to extract HTTP status code from error message
	statusCodePattern := regexp.MustCompile(`status code:?\s*(\d{3})`)
	if matches := statusCodePattern.FindStringSubmatch(errMsg); len(matches) > 1 {
		if statusCode, err := strconv.Atoi(matches[1]); err == nil {
			// Check if status code indicates a transient error
			if isTransientStatusCode(statusCode) {
				return true
			}
			// 4xx errors (except 429) are typically permanent
			if statusCode >= 400 && statusCode < 500 {
				return false
			}
		}
	}

	// Fallback: Check for network error indicators in the message
	// These typically don't have HTTP status codes
	transientIndicators := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"EOF",
		"broken pipe",
		"no such host",
		"network is unreachable",
	}

	for _, indicator := range transientIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}

	// By default, treat unknown errors as permanent to fail fast
	return false
}

// isTransientStatusCode determines if an HTTP status code represents a transient error
// that should be retried.
func isTransientStatusCode(statusCode int) bool {
	switch {
	case statusCode == 429: // Too Many Requests (rate limiting)
		return true
	case statusCode >= 500 && statusCode < 600: // 5xx Server Errors
		return true
	default:
		return false
	}
}

// convertUint64PtrToStringPtr converts a pointer to uint64 to a pointer to string
func convertUint64PtrToStringPtr(val *uint64) *string {
	if val == nil {
		return nil
	}
	str := strconv.FormatUint(*val, 10)
	return &str
}
