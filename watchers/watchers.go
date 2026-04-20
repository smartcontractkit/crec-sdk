package watchers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

const watcherNameMinRunes = 4

// statusCodePattern is a compiled regex for extracting HTTP status codes from error messages.
// Compiled at package level to avoid recompilation on every call to isTransientError.
var statusCodePattern = regexp.MustCompile(`status code:?\s*(\d{3})`)

var (
	// ErrWatcherNotFound is returned when a watcher is not found (404 response).
	ErrWatcherNotFound = errors.New("watcher not found")

	// ErrChannelIDRequired is returned when the channel ID is nil or empty.
	ErrChannelIDRequired = errors.New("channel_id cannot be empty")
	// ErrWatcherIDRequired is returned when the watcher ID is nil or empty.
	ErrWatcherIDRequired = errors.New("watcher_id cannot be empty")
	// ErrNameRequired is returned when the watcher name is empty or whitespace-only.
	ErrNameRequired = errors.New("name cannot be an empty string")
	// ErrWatcherNameTooShort is returned when the watcher name is shorter than the API minimum (4 characters, after trim).
	ErrWatcherNameTooShort = errors.New("watcher name must be at least 4 characters")
	// ErrServiceRequired is returned when the service is required but empty.
	ErrServiceRequired = errors.New("service is required")
	// ErrAddressRequired is returned when the contract address is required but empty.
	ErrAddressRequired = errors.New("address is required")
	// ErrEventsRequired is returned when the events list is empty.
	ErrEventsRequired = errors.New("events list cannot be empty")
	// ErrABIRequired is returned when the ABI is required but empty.
	ErrABIRequired = errors.New("abi cannot be empty")
	// ErrOptionsRequired is returned when the options parameter is nil.
	ErrOptionsRequired = errors.New("Options is required")
	// ErrAPIClientRequired is returned when the API client is nil in options.
	ErrAPIClientRequired = errors.New("APIClient is required")
	// ErrChainSelectorRequired is returned when the chain selector is empty or zero.
	ErrChainSelectorRequired = errors.New("chain_selector is required")
	// ErrInvalidABIType is returned when an ABI entry is not of type "event".
	ErrInvalidABIType = errors.New("invalid ABI type, only 'event' is currently supported")
	// ErrEventNotInABI is returned when a requested event is not in the provided ABI.
	ErrEventNotInABI = errors.New("event not found in provided ABI")

	// ErrWaitForActiveTimeout is returned when the watcher does not become active before the timeout.
	ErrWaitForActiveTimeout = errors.New("timeout waiting for watcher to become active")
	// ErrWaitForArchivedTimeout is returned when the watcher is not archived before the timeout.
	ErrWaitForArchivedTimeout = errors.New("timeout waiting for watcher to be archived")

	// ErrWatcherDeploymentFailed is returned when the watcher deployment fails.
	ErrWatcherDeploymentFailed = errors.New("watcher deployment failed")
	// ErrWatcherIsArchiving is returned when the watcher is being archived and cannot become active.
	ErrWatcherIsArchiving = errors.New("watcher is being archived and cannot become active")
	// ErrWatcherAlreadyArchived is returned when the watcher has already been archived.
	ErrWatcherAlreadyArchived = errors.New("watcher has been archived and cannot become active")
	// ErrWatcherArchiveFailed is returned when the watcher archive operation fails.
	ErrWatcherArchiveFailed = errors.New("watcher archive failed")

	// ErrEmptyResponse is returned when the API returns an unexpected empty response.
	ErrEmptyResponse = errors.New("unexpected empty response from API")
	// ErrUnexpectedStatus is returned when the watcher has an unexpected status value.
	ErrUnexpectedStatus = errors.New("unexpected watcher status")

	// ErrCreateWatcherRequest is returned when the create watcher request cannot be built.
	ErrCreateWatcherRequest = errors.New("failed to create watcher request")
	// ErrCreateWatcherService is returned when creating a watcher with service fails.
	ErrCreateWatcherService = errors.New("failed to create watcher with service")
	// ErrCreateWatcherABI is returned when creating a watcher with ABI fails.
	ErrCreateWatcherABI = errors.New("failed to create watcher with ABI")
	// ErrListWatchers is returned when listing watchers fails.
	ErrListWatchers = errors.New("failed to list watchers")
	// ErrGetWatcher is returned when fetching a watcher fails.
	ErrGetWatcher = errors.New("failed to get watcher")
	// ErrUpdateWatcher is returned when updating a watcher fails.
	ErrUpdateWatcher = errors.New("failed to update watcher")
	// ErrArchiveWatcher is returned when archiving a watcher fails.
	ErrArchiveWatcher = errors.New("failed to archive watcher")
	// ErrCheckWatcherStatus is returned when checking watcher status fails.
	ErrCheckWatcherStatus = errors.New("failed to check watcher status")
	// ErrUnexpectedStatusCode is returned when the API returns an unexpected HTTP status code.
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)

// EventABIInput describes a single input parameter of an event in the ABI.
type EventABIInput struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

// EventABI describes an event definition in Solidity ABI format.
type EventABI struct {
	Anonymous bool            `json:"anonymous"`
	Inputs    []EventABIInput `json:"inputs"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
}

// CreateWithServiceInput defines the input for creating a watcher using a predefined service (e.g., dvp, dta).
type CreateWithServiceInput struct {
	Name          string                 `json:"name"`
	ChainSelector string                 `json:"chain_selector"`
	Service       string                 `json:"service"`
	Events        []string               `json:"events"`
	Address       string                 `json:"address"`
	ServiceConfig map[string]interface{} `json:"service_config,omitempty"`
}

// CreateWithABIInput defines the input for creating a watcher with a custom event ABI.
type CreateWithABIInput struct {
	Name          string     `json:"name"`
	ChainSelector string     `json:"chain_selector"`
	Address       string     `json:"address"`
	Events        []string   `json:"events"`
	ABI           []EventABI `json:"abi"`
}

// UpdateInput defines the input for updating a watcher.
type UpdateInput struct {
	Name string `json:"name"`
}

// ListFilters defines optional filters for listing watchers (pagination, status, chain, etc.).
type ListFilters struct {
	Limit         *int                       `url:"limit,omitempty"`
	Offset        *int64                     `url:"offset,omitempty"`
	Name          *string                    `url:"name,omitempty"`
	Status        *[]apiClient.WatcherStatus `url:"status,omitempty"`
	ChainSelector *string                    `url:"chain_selector,omitempty"`
	Address       *string                    `url:"address,omitempty"`
	Service       *[]string                  `url:"service,omitempty"`
	EventName     *string                    `url:"event_name,omitempty"`
}

// Options defines the options for creating a new CREC Watchers client.
//   - Logger: Optional logger instance.
//   - APIClient: The CREC API client instance.
//   - PollInterval: Optional polling interval for WaitForActive. Defaults to 2 seconds if not set.
//   - EventualConsistencyWindow: Optional duration to tolerate 404 errors after creation due to eventual consistency. Defaults to 2 seconds if not set.
type Options struct {
	Logger                    *slog.Logger
	APIClient                 *apiClient.ClientWithResponses
	PollInterval              time.Duration
	EventualConsistencyWindow time.Duration
}

// Client provides operations for managing CREC watchers.
// Watchers monitor blockchain events on specific smart contracts and trigger workflows.
type Client struct {
	logger                    *slog.Logger
	apiClient                 *apiClient.ClientWithResponses
	pollInterval              time.Duration
	eventualConsistencyWindow time.Duration
}

// NewClient creates a new CREC Watchers client with the provided options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Watchers client, see Options for details.
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

	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	eventualConsistencyWindow := opts.EventualConsistencyWindow
	if eventualConsistencyWindow == 0 {
		eventualConsistencyWindow = 2 * time.Second
	}

	logger.Debug("Creating CREC Watchers client",
		"poll_interval", pollInterval,
		"eventual_consistency_window", eventualConsistencyWindow)

	return &Client{
		logger:                    logger,
		apiClient:                 opts.APIClient,
		pollInterval:              pollInterval,
		eventualConsistencyWindow: eventualConsistencyWindow,
	}, nil
}

// CreateWithService creates a new watcher using a predefined service (e.g., dvp, dta, test_consumer).
func (c *Client) CreateWithService(ctx context.Context, channelID uuid.UUID, input CreateWithServiceInput) (*apiClient.Watcher, error) {
	c.logger.Debug("Creating watcher with service",
		"channel_id", channelID.String(),
		"service", input.Service)

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if input.ChainSelector == "" || input.ChainSelector == "0" {
		return nil, ErrChainSelectorRequired
	}
	if input.Address == "" {
		return nil, ErrAddressRequired
	}
	if input.Service == "" {
		return nil, ErrServiceRequired
	}
	if len(input.Events) == 0 {
		return nil, ErrEventsRequired
	}

	normalizedName, err := normalizeWatcherCreateName(input.Name)
	if err != nil {
		return nil, err
	}
	createWatcherWithService := apiClient.CreateWatcherWithService{
		Name:          normalizedName,
		ChainSelector: input.ChainSelector,
		Address:       input.Address,
		Service:       input.Service,
		Events:        input.Events,
	}
	if input.ServiceConfig != nil {
		createWatcherWithService.ServiceConfig = &input.ServiceConfig
	}

	var createWatcherReq apiClient.CreateWatcher
	if err := createWatcherReq.FromCreateWatcherWithService(createWatcherWithService); err != nil {
		c.logger.Error("Failed to create watcher request", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherRequest, err)
	}

	resp, err := c.apiClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		c.logger.Error("Failed to create watcher with service", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherService, err)
	}

	if resp.StatusCode() != 201 {
		c.logger.Error("Failed to create watcher with service - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWatcherService, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, ErrEmptyResponse
	}

	c.logger.Info("Watcher created successfully", "watcher_id", resp.JSON201.WatcherId.String())

	return resp.JSON201, nil
}

// CreateWithABI creates a new watcher with a custom event ABI.
func (c *Client) CreateWithABI(ctx context.Context, channelID uuid.UUID, input CreateWithABIInput) (*apiClient.Watcher, error) {
	c.logger.Debug("Creating watcher with ABI",
		"channel_id", channelID.String(),
		"address", input.Address,
		"abi_count", len(input.ABI))

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if input.ChainSelector == "" || input.ChainSelector == "0" {
		return nil, ErrChainSelectorRequired
	}
	if input.Address == "" {
		return nil, ErrAddressRequired
	}
	if len(input.Events) == 0 {
		return nil, ErrEventsRequired
	}
	normalizedName, err := normalizeWatcherCreateName(input.Name)
	if err != nil {
		return nil, err
	}
	if len(input.ABI) == 0 {
		return nil, ErrABIRequired
	}

	// Validate that all ABI entries are of type "event"
	for i, abi := range input.ABI {
		if abi.Type != "event" {
			c.logger.Error("Invalid ABI type",
				"abi_index", i,
				"abi_type", abi.Type,
				"abi_name", abi.Name)
			return nil, fmt.Errorf("%w: got '%s' for event '%s' at index %d", ErrInvalidABIType, abi.Type, abi.Name, i)
		}
	}

	// Validate that all requested events exist in the provided ABI
	abiEventMap := make(map[string]bool, len(input.ABI))
	for _, abi := range input.ABI {
		abiEventMap[abi.Name] = true
	}

	for _, eventName := range input.Events {
		if !abiEventMap[eventName] {
			c.logger.Error("Requested event not found in provided ABI",
				"event_name", eventName,
				"abi_count", len(input.ABI))
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
		Name:          normalizedName,
		ChainSelector: input.ChainSelector,
		Address:       input.Address,
		Events:        input.Events,
		Abi:           abiList,
	}

	var createWatcherReq apiClient.CreateWatcher
	if err := createWatcherReq.FromCreateWatcherWithABI(createWatcherWithABI); err != nil {
		c.logger.Error("Failed to create watcher request", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherRequest, err)
	}

	resp, err := c.apiClient.PostChannelsChannelIdWatchersWithResponse(ctx, channelID, createWatcherReq)
	if err != nil {
		c.logger.Error("Failed to create watcher with ABI", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateWatcherABI, err)
	}

	if resp.StatusCode() != 201 {
		c.logger.Error("Failed to create watcher with ABI - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateWatcherABI, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, ErrEmptyResponse
	}

	c.logger.Info("Watcher created successfully", "watcher_id", resp.JSON201.WatcherId.String())

	return resp.JSON201, nil
}

// List retrieves watchers for a specific channel with optional filters.
func (c *Client) List(ctx context.Context, channelID uuid.UUID, filters ListFilters) (*apiClient.WatcherList, error) {
	c.logger.Debug("Listing watchers", "channel_id", channelID.String())

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}

	params := &apiClient.GetChannelsChannelIdWatchersParams{
		Limit:         filters.Limit,
		Offset:        filters.Offset,
		Name:          filters.Name,
		ChainSelector: filters.ChainSelector,
		Address:       filters.Address,
		EventName:     filters.EventName,
		Service:       filters.Service,
		Status:        filters.Status,
	}

	resp, err := c.apiClient.GetChannelsChannelIdWatchersWithResponse(ctx, channelID, params)
	if err != nil {
		c.logger.Error("Failed to list watchers", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrListWatchers, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to list watchers - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrListWatchers, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
	}

	c.logger.Debug("Watchers listed successfully", "count", len(resp.JSON200.Data))

	return resp.JSON200, nil
}

// Get retrieves a specific watcher by channel ID and watcher ID.
func (c *Client) Get(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID) (*apiClient.Watcher, error) {
	c.logger.Debug("Getting watcher",
		"channel_id", channelID.String(),
		"watcher_id", watcherID.String())

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, ErrWatcherIDRequired
	}

	resp, err := c.apiClient.GetChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID)
	if err != nil {
		c.logger.Error("Failed to get watcher", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetWatcher, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to get watcher - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))

		if resp.StatusCode() == 404 {
			return nil, fmt.Errorf("%w: watcher ID %s", ErrWatcherNotFound, watcherID.String())
		}

		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
	}

	return resp.JSON200, nil
}

// Update updates a watcher's name.
func (c *Client) Update(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, input UpdateInput) (*apiClient.Watcher, error) {
	c.logger.Debug("Updating watcher",
		"channel_id", channelID.String(),
		"watcher_id", watcherID.String(),
		"name", input.Name)

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
		Name: &input.Name,
	}

	resp, err := c.apiClient.PatchChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID, updateReq)
	if err != nil {
		c.logger.Error("Failed to update watcher", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrUpdateWatcher, err)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to update watcher - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))

		if resp.StatusCode() == 404 {
			return nil, fmt.Errorf("%w: watcher ID %s", ErrWatcherNotFound, watcherID.String())
		}

		return nil, fmt.Errorf("%w: %w (status code %d)", ErrUpdateWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, ErrEmptyResponse
	}

	c.logger.Info("Watcher updated successfully", "watcher_id", watcherID.String())

	return resp.JSON200, nil
}

// WaitForActive polls a watcher until it reaches active status, fails, or times out.
// Returns ErrWaitForActiveTimeout if maxWaitTime is exceeded before the watcher becomes active.
func (c *Client) WaitForActive(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, maxWaitTime time.Duration) (*apiClient.Watcher, error) {
	c.logger.Debug("Waiting for watcher to become active",
		"channel_id", channelID.String(),
		"watcher_id", watcherID.String(),
		"max_wait_time", maxWaitTime,
		"eventual_consistency_window", c.eventualConsistencyWindow)

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, ErrWatcherIDRequired
	}

	startTime := time.Now()
	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			c.logger.Error("Timeout waiting for watcher to become active")
			return nil, ErrWaitForActiveTimeout
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			watcher, err := c.Get(ctx, channelID, watcherID)
			if err != nil {
				// During the eventual consistency window, tolerate 404 errors
				// This handles the case where a watcher was just created but isn't immediately visible
				if errors.Is(err, ErrWatcherNotFound) {
					elapsedTime := time.Since(startTime)
					if elapsedTime < c.eventualConsistencyWindow {
						c.logger.Debug("Watcher not found yet, retrying (eventual consistency)",
							"elapsed", elapsedTime,
							"window", c.eventualConsistencyWindow)
						continue
					}
					c.logger.Error("Watcher not found after consistency window",
						"elapsed", elapsedTime)
					return nil, ErrWatcherNotFound
				}

				// Check if error is transient (network issues, 5xx, etc.)
				if isTransientError(err) {
					c.logger.Warn("Transient error while checking watcher status, will retry", "error", err)
					continue
				}
				// Permanent error - fail immediately
				return nil, fmt.Errorf("%w: %w", ErrCheckWatcherStatus, err)
			}

			switch watcher.Status {
			case apiClient.WatcherStatusActive:
				c.logger.Info("Watcher is now active")
				return watcher, nil
			case apiClient.WatcherStatusFailed:
				c.logger.Error("Watcher deployment failed")
				return nil, ErrWatcherDeploymentFailed
			case apiClient.WatcherStatusPending:
				c.logger.Debug("Watcher still pending, continuing to wait")
				continue
			case apiClient.WatcherStatusArchiving:
				c.logger.Error("Watcher is being archived")
				return nil, ErrWatcherIsArchiving
			case apiClient.WatcherStatusArchived:
				c.logger.Error("Watcher has been archived")
				return nil, ErrWatcherAlreadyArchived
			default:
				c.logger.Error("Unexpected watcher status while waiting for active", "status", watcher.Status)
				return nil, fmt.Errorf("%w: %s", ErrUnexpectedStatus, watcher.Status)
			}
		}
	}
}

// Archive archives a watcher by transitioning it to archived status via PATCH.
// Archiving is asynchronous: the PATCH returns 202 with the watcher in "archiving" status;
// use WaitForArchived to poll until the watcher reaches "archived" or "archive_failed".
func (c *Client) Archive(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID) (*apiClient.Watcher, error) {
	c.logger.Debug("Archiving watcher",
		"channel_id", channelID.String(),
		"watcher_id", watcherID.String())

	if err := validateChannelID(channelID); err != nil {
		return nil, err
	}
	if watcherID == uuid.Nil {
		return nil, ErrWatcherIDRequired
	}

	archiveStatus := apiClient.WatcherStatusArchived
	updateReq := apiClient.UpdateWatcher{
		Status: &archiveStatus,
	}

	resp, err := c.apiClient.PatchChannelsChannelIdWatchersWatcherIdWithResponse(ctx, channelID, watcherID, updateReq)
	if err != nil {
		c.logger.Error("Failed to archive watcher", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrArchiveWatcher, err)
	}

	if resp.StatusCode() == 404 {
		return nil, fmt.Errorf("%w: watcher ID %s", ErrWatcherNotFound, watcherID.String())
	}

	if resp.StatusCode() == 202 {
		if resp.JSON202 == nil {
			return nil, fmt.Errorf("%w: %w", ErrArchiveWatcher, ErrEmptyResponse)
		}
		c.logger.Info("Watcher archive initiated (async)", "watcher_id", watcherID.String())
		return resp.JSON202, nil
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to archive watcher - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrArchiveWatcher, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrArchiveWatcher, ErrEmptyResponse)
	}

	c.logger.Info("Watcher archived successfully", "watcher_id", watcherID.String())
	return resp.JSON200, nil
}

// WaitForArchived polls the watcher until it reaches "archived" status or the timeout is reached.
// Returns ErrWaitForArchivedTimeout if maxWaitTime is exceeded before the watcher is archived.
func (c *Client) WaitForArchived(ctx context.Context, channelID uuid.UUID, watcherID uuid.UUID, maxWaitTime time.Duration) error {
	c.logger.Debug("Waiting for watcher to be archived",
		"channel_id", channelID.String(),
		"watcher_id", watcherID.String(),
		"max_wait_time", maxWaitTime)

	if err := validateChannelID(channelID); err != nil {
		return err
	}
	if watcherID == uuid.Nil {
		return ErrWatcherIDRequired
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			c.logger.Error("Timeout waiting for watcher to be archived")
			return ErrWaitForArchivedTimeout
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			watcher, err := c.Get(ctx, channelID, watcherID)
			if err != nil {
				if isTransientError(err) {
					c.logger.Warn("Transient error while checking watcher archive status, will retry", "error", err)
					continue
				}
				return fmt.Errorf("%w: %w", ErrCheckWatcherStatus, err)
			}

			switch watcher.Status {
			case apiClient.WatcherStatusArchived:
				c.logger.Debug("Watcher is now archived")
				return nil
			case apiClient.WatcherStatusArchiving:
				c.logger.Debug("Watcher is being archived, continuing to wait")
				continue
			case apiClient.WatcherStatusActive, apiClient.WatcherStatusPending, apiClient.WatcherStatusFailed:
				c.logger.Error("Watcher archive appears to have failed", "status", watcher.Status)
				return fmt.Errorf("%w, watcher is in %s state", ErrWatcherArchiveFailed, watcher.Status)
			default:
				c.logger.Error("Unexpected watcher status while waiting for archive", "status", watcher.Status)
				return fmt.Errorf("%w while waiting for archive: %s", ErrUnexpectedStatus, watcher.Status)
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

func normalizeWatcherCreateName(name string) (string, error) {
	s := strings.TrimSpace(name)
	if s == "" {
		return "", ErrNameRequired
	}
	if utf8.RuneCountInString(s) < watcherNameMinRunes {
		return "", ErrWatcherNameTooShort
	}
	return s, nil
}

// isTransientError determines if an error is transient and should be retried during polling.
// It checks for known transient HTTP status codes and network errors.
func isTransientError(err error) bool {
	// Validation errors are permanent
	if errors.Is(err, ErrChannelIDRequired) ||
		errors.Is(err, ErrWatcherIDRequired) ||
		errors.Is(err, ErrNameRequired) ||
		errors.Is(err, ErrWatcherNameTooShort) ||
		errors.Is(err, ErrServiceRequired) ||
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
