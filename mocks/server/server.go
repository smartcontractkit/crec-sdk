package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/smartcontractkit/crec-api-go/stdserver"
)

type MockServer struct {
	TestServer *httptest.Server

	// simple in-memory stores for the mock
	wallets    []stdserver.Wallet
	channels   []stdserver.Channel
	operations []stdserver.Operation
	watchers   []stdserver.Watcher
}

var _ stdserver.ServerInterface = (*MockServer)(nil)

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

func (s *MockServer) Close() {
	if s.TestServer != nil {
		s.TestServer.Close()
	}
}

func (s *MockServer) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		stdserver.HealthCheck{
			Status: "ok",
		},
	)
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
	}
	s.channels = append(s.channels, channel)

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
		offset = *params.Offset
	}

	filteredChannels := s.channels
	if params.Name != nil && *params.Name != "" {
		filtered := []stdserver.Channel{}
		for _, ch := range s.channels {
			if ch.Name == *params.Name {
				filtered = append(filtered, ch)
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

func (s *MockServer) DeleteChannelsChannelId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	for _, ch := range s.channels {
		if ch.ChannelId == channelId {
			// Just mark as accepted - in a real implementation would set deleted_at
			w.WriteHeader(http.StatusAccepted)
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
		Status:            "pending",
		ChainSelector:     request.ChainSelector,
		Address:           request.Address,
		WalletOperationId: request.WalletOperationId,
		Transactions:      request.Transactions,
		Signature:         request.Signature,
		CreatedAt:         now,
	}

	// Store the operation so it can be retrieved later
	s.operations = append(s.operations, operation)

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
		offset = *params.Offset
	}

	// Filter operations by optional query params
	// Note: Operations don't have a ChannelId field in the schema, so we return all operations
	filteredOps := []stdserver.Operation{}
	for _, op := range s.operations {
		// Filter by status if provided
		if params.Status != nil && op.Status != *params.Status {
			continue
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
		offset = *params.Offset
	}

	// Filter by channel ID and query parameters
	filteredWatchers := []stdserver.Watcher{}
	for _, watcher := range s.watchers {
		// First, filter by channelId - only return watchers from this channel
		if watcher.ChannelId != channelId {
			continue
		}

		// Apply optional query filters
		if params.Status != nil && watcher.Status != *params.Status {
			continue
		}
		if params.Address != nil && watcher.Address != *params.Address {
			continue
		}
		if params.Domain != nil {
			if watcher.Domain == nil || *watcher.Domain != *params.Domain {
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
	data := []stdserver.Watcher{}
	if offset < len(filteredWatchers) {
		data = filteredWatchers[offset:end]
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
		Status:    "pending", // Start as pending
		CreatedAt: now,
		Events:    []string{},
	}

	// Handle the union type - try to unmarshal as domain first, then ABI
	if domain, err := request.AsCreateWatcherWithDomain(); err == nil {
		watcher.Name = domain.Name
		watcher.Domain = &domain.Domain
		watcher.Address = domain.Address
		watcher.ChainSelector = domain.ChainSelector
		watcher.Events = domain.Events
	} else if abiReq, err := request.AsCreateWatcherWithABI(); err == nil {
		watcher.Name = abiReq.Name
		watcher.Address = abiReq.Address
		watcher.ChainSelector = abiReq.ChainSelector
		watcher.Events = abiReq.Events
		watcher.Abi = &abiReq.Abi
	} else {
		http.Error(w, "invalid watcher request", http.StatusBadRequest)
		return
	}

	s.watchers = append(s.watchers, watcher)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(watcher)
}

func (s *MockServer) GetChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
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

	// Find and update the watcher
	for i, watcher := range s.watchers {
		if watcher.WatcherId == watcherId {
			// Validate that the watcher belongs to the requested channel
			if watcher.ChannelId != channelId {
				http.Error(w, "watcher not found", http.StatusNotFound)
				return
			}

			// Update name (always set since it's a required field in UpdateWatcher)
			s.watchers[i].Name = &request.Name

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(s.watchers[i])
			return
		}
	}
	http.Error(w, "watcher not found", http.StatusNotFound)
}

func (s *MockServer) DeleteChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
	for i, watcher := range s.watchers {
		if watcher.WatcherId == watcherId {
			// Validate that the watcher belongs to the requested channel
			if watcher.ChannelId != channelId {
				http.Error(w, "watcher not found", http.StatusNotFound)
				return
			}

			// Mark as deleting to simulate async deletion
			s.watchers[i].Status = "deleting"

			// Schedule automatic transition to "deleted" after a brief delay
			// Tests can also manually advance the state using helper methods
			scheduleStatusTransition(
				watcherId,
				"deleting",
				"deleted",
				50*time.Millisecond,
				func(id uuid.UUID, from, to string) bool {
					return s.updateWatcherStatusConditional(id, from, to)
				},
			)

			w.WriteHeader(http.StatusAccepted)
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
	wallet := stdserver.Wallet{
		WalletId:      id,
		Address:       in.WalletOwnerAddress,
		ChainSelector: in.ChainSelector,
		Name:          &in.Name,
	}
	s.wallets = append(s.wallets, wallet)

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
		offset = *params.Offset
	}

	end := offset + limit
	if end > len(s.wallets) {
		end = len(s.wallets)
	}
	data := []stdserver.Wallet{}
	if offset < len(s.wallets) {
		data = s.wallets[offset:end]
	}
	hasMore := end < len(s.wallets)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.WalletList{Data: data, HasMore: hasMore})
}

func (s *MockServer) GetWalletsWalletId(w http.ResponseWriter, r *http.Request, walletId openapiTypes.UUID) {
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

	// Find and update the wallet
	for i, wallet := range s.wallets {
		if wallet.WalletId == walletId {
			s.wallets[i].Name = &request.Name
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

// scheduleStatusTransition is a generic helper to simulate async state transitions.
// It schedules a status change after a delay, checking the current status before updating.
func scheduleStatusTransition(id uuid.UUID, fromStatus, toStatus string, delay time.Duration, updateFn func(uuid.UUID, string, string) bool) {
	go func() {
		time.Sleep(delay)
		updateFn(id, fromStatus, toStatus)
	}()
}

// updateWatcherStatusConditional updates a watcher's status only if it matches the expected current status.
// Returns true if the update was successful.
func (s *MockServer) updateWatcherStatusConditional(watcherID uuid.UUID, fromStatus, toStatus string) bool {
	for i, w := range s.watchers {
		if w.WatcherId == watcherID && w.Status == fromStatus {
			s.watchers[i].Status = toStatus
			return true
		}
	}
	return false
}
