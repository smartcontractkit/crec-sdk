package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/smartcontractkit/crec-api-go/stdserver"
)

// MockServer is an in-memory HTTP server that implements the CREC API for testing.
// It provides channels, watchers, wallets, and operations endpoints with no persistence.
type MockServer struct {
	TestServer *httptest.Server

	// mu protects concurrent access to the in-memory stores
	mu sync.RWMutex

	// simple in-memory stores for the mock
	wallets    []stdserver.Wallet
	channels   []stdserver.Channel
	operations []stdserver.Operation
	watchers   []stdserver.Watcher
}

var _ stdserver.ServerInterface = (*MockServer)(nil)

// NewMockServer creates and starts a new MockServer with an in-memory store.
// Call Close when done to shut down the HTTP server.
func NewMockServer() *MockServer {
	s := &MockServer{
		wallets:    make([]stdserver.Wallet, 0),
		channels:   make([]stdserver.Channel, 0),
		operations: make([]stdserver.Operation, 0),
		watchers:   make([]stdserver.Watcher, 0),
	}
	r := http.NewServeMux()
	h := stdserver.HandlerFromMux(s, r)

	s.TestServer = httptest.NewServer(h)
	return s
}

// Close shuts down the mock HTTP server and releases resources.
func (s *MockServer) Close() {
	if s.TestServer != nil {
		s.TestServer.Close()
	}
}

// GetHealthCheck handles GET /health and returns a healthy status.
func (s *MockServer) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		stdserver.HealthCheck{
			Status: "ok",
		},
	)
}

// GetNetworks handles GET /networks and returns an empty network list.
func (s *MockServer) GetNetworks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.NetworkList{
		Data:    []stdserver.Network{},
		HasMore: false,
	})
}

// ============================================================================
// CHANNELS ENDPOINTS
// ============================================================================

func (s *MockServer) PostChannels(w http.ResponseWriter, r *http.Request) {
	var request stdserver.CreateChannel
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	channelId := uuid.New()
	now := time.Now().Unix()
	channel := stdserver.Channel{
		ChannelId: channelId,
		Name:      request.Name,
		CreatedAt: now,
		Status:    stdserver.ChannelStatusActive,
	}

	s.mu.Lock()
	s.channels = append(s.channels, channel)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(channel)
}

func (s *MockServer) GetChannels(w http.ResponseWriter, r *http.Request, params stdserver.GetChannelsParams) {
	limit := 20
	offset := 0
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int(*params.Offset)
	}

	s.mu.RLock()
	channels := make([]stdserver.Channel, len(s.channels))
	copy(channels, s.channels)
	s.mu.RUnlock()

	filteredChannels := channels
	if params.Name != nil && *params.Name != "" {
		filtered := []stdserver.Channel{}
		for _, ch := range channels {
			if ch.Name == *params.Name {
				filtered = append(filtered, ch)
			}
		}
		filteredChannels = filtered
	}
	if params.Status != nil {
		filtered := []stdserver.Channel{}
		for _, ch := range filteredChannels {
			for _, st := range *params.Status {
				if ch.Status == st {
					filtered = append(filtered, ch)
					break
				}
			}
		}
		filteredChannels = filtered
	}

	end := offset + limit
	if end > len(filteredChannels) {
		end = len(filteredChannels)
	}
	data := []stdserver.Channel{}
	if offset < len(filteredChannels) {
		data = filteredChannels[offset:end]
	}
	hasMore := end < len(filteredChannels)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.ChannelList{Data: data, HasMore: hasMore})
}

func (s *MockServer) GetChannelsChannelId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.channels {
		if ch.ChannelId == channelId {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ch)
			return
		}
	}
	http.Error(w, "channel not found", http.StatusNotFound)
}

func (s *MockServer) PatchChannelsChannelId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	var request stdserver.PatchChannel
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and update the channel
	for i, ch := range s.channels {
		if ch.ChannelId == channelId {
			if request.Name != nil {
				s.channels[i].Name = *request.Name
			}
			if request.Description != nil {
				s.channels[i].Description = request.Description
			}
			if request.Status != nil {
				s.channels[i].Status = *request.Status
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(s.channels[i])
			return
		}
	}
	http.Error(w, "channel not found", http.StatusNotFound)
}

// ============================================================================
// OPERATIONS ENDPOINTS (under channels)
// ============================================================================

func (s *MockServer) PostChannelsChannelIdOperations(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	var request stdserver.CreateOperation
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	operationId := uuid.New()
	now := time.Now().Unix()

	operation := stdserver.Operation{
		OperationId:       operationId,
		Status:            stdserver.OperationStatusAccepted,
		ChainSelector:     request.ChainSelector,
		Address:           request.Address,
		WalletOperationId: request.WalletOperationId,
		Deadline:          request.Deadline,
		Transactions:      request.Transactions,
		Signature:         request.Signature,
		CreatedAt:         now,
	}

	// Store the operation so it can be retrieved later
	s.mu.Lock()
	s.operations = append(s.operations, operation)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(stdserver.OperationResponse{
		OperationId: operationId,
	})
}

func (s *MockServer) GetChannelsChannelIdOperations(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdOperationsParams) {
	limit := 20
	offset := 0
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int(*params.Offset)
	}

	s.mu.RLock()
	operations := make([]stdserver.Operation, len(s.operations))
	copy(operations, s.operations)
	s.mu.RUnlock()

	// Filter operations by optional query params
	// Note: Operations don't have a ChannelId field in the schema, so we return all operations
	filteredOps := []stdserver.Operation{}
	for _, op := range operations {
		// Filter by status if provided
		if params.Status != nil {
			found := false
			for _, st := range *params.Status {
				if op.Status == st {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by chainSelector if provided
		if params.ChainSelector != nil {
			if op.ChainSelector != *params.ChainSelector {
				continue
			}
		}

		// Filter by address if provided
		if params.Address != nil && op.Address != *params.Address {
			continue
		}

		filteredOps = append(filteredOps, op)
	}

	end := offset + limit
	if end > len(filteredOps) {
		end = len(filteredOps)
	}
	data := []stdserver.Operation{}
	if offset < len(filteredOps) {
		data = filteredOps[offset:end]
	}
	hasMore := end < len(filteredOps)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.OperationList{Data: data, HasMore: hasMore})
}

func (s *MockServer) GetChannelsChannelIdOperationsOperationId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, operationId openapiTypes.UUID) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, op := range s.operations {
		if op.OperationId == operationId {
			// Note: Operations don't have a ChannelId field in the schema
			// This is a simplified mock that returns the operation if it exists, regardless of channel
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(op)
			return
		}
	}
	http.Error(w, "operation not found", http.StatusNotFound)
}

func (s *MockServer) PostOperationStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

// ============================================================================
// EVENTS ENDPOINTS
// ============================================================================

func (s *MockServer) PostEvents(w http.ResponseWriter, r *http.Request) {
	// Not implemented - events are typically created via watchers
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdEvents(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdEventsParams) {
	// Not implemented - events service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdEventsSearch(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdEventsSearchParams) {
	// Not implemented - events service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdEventsSearchEventId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, eventId openapiTypes.UUID) {
	// Not implemented - events service
	w.WriteHeader(http.StatusNotImplemented)
}

// ============================================================================
// WATCHERS ENDPOINTS (under channels)
// ============================================================================

func (s *MockServer) GetChannelsChannelIdWatchers(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdWatchersParams) {
	limit := 20
	offset := 0
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int(*params.Offset)
	}

	s.mu.RLock()
	watchers := make([]stdserver.Watcher, len(s.watchers))
	copy(watchers, s.watchers)
	s.mu.RUnlock()

	// Filter by channel ID and query parameters
	filteredWatchers := []stdserver.Watcher{}
	for _, watcher := range watchers {
		// First, filter by channelId - only return watchers from this channel
		if watcher.ChannelId != channelId {
			continue
		}

		// Apply optional query filters
		if params.Status != nil {
			found := false
			for _, st := range *params.Status {
				if watcher.Status == st {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if params.Address != nil && watcher.Address != *params.Address {
			continue
		}
		if params.Service != nil {
			if watcher.Service == nil {
				continue
			}
			found := false
			for _, svc := range *params.Service {
				if *watcher.Service == svc {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if params.Name != nil {
			if watcher.Name == nil || *watcher.Name != *params.Name {
				continue
			}
		}
		if params.ChainSelector != nil {
			if watcher.ChainSelector != *params.ChainSelector {
				continue
			}
		}
		if params.EventName != nil {
			// Check if the watcher monitors this event
			eventFound := false
			for _, event := range watcher.Events {
				if event == *params.EventName {
					eventFound = true
					break
				}
			}
			if !eventFound {
				continue
			}
		}

		filteredWatchers = append(filteredWatchers, watcher)
	}

	end := offset + limit
	if end > len(filteredWatchers) {
		end = len(filteredWatchers)
	}
	data := []stdserver.WatcherSummary{}
	if offset < len(filteredWatchers) {
		watchersSlice := filteredWatchers[offset:end]
		for _, watcher := range watchersSlice {
			summary := stdserver.WatcherSummary{
				WatcherId:     watcher.WatcherId,
				Name:          watcher.Name,
				ChainSelector: watcher.ChainSelector,
				Address:       watcher.Address,
				Status:        watcher.Status,
				ChannelId:     watcher.ChannelId,
				CreatedAt:     watcher.CreatedAt,
				DonFamily:     watcher.DonFamily,
				WorkflowId:    watcher.WorkflowId,
			}
			data = append(data, summary)
		}
	}
	hasMore := end < len(filteredWatchers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.WatcherList{Data: data, HasMore: hasMore})
}

func (s *MockServer) PostChannelsChannelIdWatchers(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	var request stdserver.CreateWatcher
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	watcherId := uuid.New()
	now := time.Now().Unix()

	// Create watcher based on the union type
	watcher := stdserver.Watcher{
		WatcherId: watcherId,
		ChannelId: channelId,
		Address:   "",
		Status:    stdserver.WatcherStatusPending, // Start as pending
		CreatedAt: now,
		Events:    []string{},
	}

	// Handle the union type - try to unmarshal as service first, then ABI
	if svcReq, err := request.AsCreateWatcherWithService(); err == nil {
		watcher.Name = &svcReq.Name
		serviceVal := svcReq.Service
		watcher.Service = &serviceVal
		watcher.Address = svcReq.Address
		watcher.ChainSelector = svcReq.ChainSelector
		watcher.Events = svcReq.Events
	} else if abiReq, err := request.AsCreateWatcherWithABI(); err == nil {
		watcher.Name = &abiReq.Name
		watcher.Address = abiReq.Address
		watcher.ChainSelector = abiReq.ChainSelector
		watcher.Events = abiReq.Events
		watcher.Abi = &abiReq.Abi
	} else {
		http.Error(w, "invalid watcher request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.watchers = append(s.watchers, watcher)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(watcher)
}

func (s *MockServer) GetChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, watcher := range s.watchers {
		if watcher.WatcherId == watcherId {
			// Validate that the watcher belongs to the requested channel
			if watcher.ChannelId != channelId {
				http.Error(w, "watcher not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(watcher)
			return
		}
	}
	http.Error(w, "watcher not found", http.StatusNotFound)
}

func (s *MockServer) PatchChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
	var request stdserver.UpdateWatcher
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, watcher := range s.watchers {
		if watcher.WatcherId == watcherId {
			if watcher.ChannelId != channelId {
				http.Error(w, "watcher not found", http.StatusNotFound)
				return
			}

			if request.Name != nil {
				s.watchers[i].Name = request.Name
			}

			if request.Status != nil && *request.Status == stdserver.WatcherStatusArchived {
				s.watchers[i].Status = stdserver.WatcherStatusArchiving
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(s.watchers[i])

				scheduleWatcherArchive(
					watcherId,
					50*time.Millisecond,
					func(id uuid.UUID) bool {
						return s.archiveWatcher(id)
					},
				)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(s.watchers[i])
			return
		}
	}
	http.Error(w, "watcher not found", http.StatusNotFound)
}

func (s *MockServer) PostWallets(w http.ResponseWriter, r *http.Request) {
	var in stdserver.CreateWallet
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uuid.New()
	now := time.Now().Unix()
	wallet := stdserver.Wallet{
		WalletId:      id,
		Address:       in.WalletOwnerAddress,
		ChainSelector: in.ChainSelector,
		Name:          in.Name,
		CreatedAt:     &now,
		Status:        stdserver.WalletStatusDeployed,
	}

	s.mu.Lock()
	s.wallets = append(s.wallets, wallet)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(wallet)
}

func (s *MockServer) GetWallets(w http.ResponseWriter, r *http.Request, params stdserver.GetWalletsParams) {
	// simple pagination on our in-memory slice
	limit := 50
	offset := 0
	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = int(*params.Offset)
	}

	s.mu.RLock()
	wallets := make([]stdserver.Wallet, len(s.wallets))
	copy(wallets, s.wallets)
	s.mu.RUnlock()

	end := offset + limit
	if end > len(wallets) {
		end = len(wallets)
	}
	data := []stdserver.Wallet{}
	if offset < len(wallets) {
		data = wallets[offset:end]
	}
	hasMore := end < len(wallets)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.WalletList{Data: data, HasMore: hasMore})
}

func (s *MockServer) GetWalletsWalletId(w http.ResponseWriter, r *http.Request, walletId openapiTypes.UUID) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, wallet := range s.wallets {
		if wallet.WalletId == walletId {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(wallet)
			return
		}
	}
	http.Error(w, "wallet not found", http.StatusNotFound)
}

func (s *MockServer) PatchWalletsWalletId(w http.ResponseWriter, r *http.Request, walletId openapiTypes.UUID) {
	var request stdserver.UpdateWallet
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, wallet := range s.wallets {
		if wallet.WalletId == walletId {
			if request.Name != nil {
				s.wallets[i].Name = *request.Name
			}
			if request.Status != nil {
				s.wallets[i].Status = *request.Status
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(s.wallets[i])
			return
		}
	}
	http.Error(w, "wallet not found", http.StatusNotFound)
}

// ============================================================================
// HELPER METHODS (INTERNAL)
// ============================================================================

// scheduleWatcherArchive schedules the transition of a watcher from "archiving" to "archived" after a delay.
func scheduleWatcherArchive(id uuid.UUID, delay time.Duration, archiveFn func(uuid.UUID) bool) {
	go func() {
		time.Sleep(delay)
		archiveFn(id)
	}()
}

// archiveWatcher transitions a watcher from "archiving" to "archived" status.
func (s *MockServer) archiveWatcher(watcherID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, w := range s.watchers {
		if w.WatcherId == watcherID && w.Status == stdserver.WatcherStatusArchiving {
			s.watchers[i].Status = stdserver.WatcherStatusArchived
			return true
		}
	}
	return false
}
