package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/smartcontractkit/crec-api-go/stdserver"
)

type MockServer struct {
	TestServer *httptest.Server

	// simple in-memory stores for the mock
	accounts   []stdserver.Account
	channels   []stdserver.Channel
	operations []stdserver.Operation
}

var _ stdserver.ServerInterface = (*MockServer)(nil)

func NewMockServer() *MockServer {
	s := &MockServer{
		accounts:   make([]stdserver.Account, 0),
		channels:   make([]stdserver.Channel, 0),
		operations: make([]stdserver.Operation, 0),
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

	filteredOps := []stdserver.Operation{}
	for _, op := range s.operations {
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
// EVENTS & WATCHERS - COMMENTED OUT (not part of channels/operations)
// ============================================================================

func (s *MockServer) PostEvents(w http.ResponseWriter, r *http.Request) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdEvents(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdEventsParams) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdWatchers(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, params stdserver.GetChannelsChannelIdWatchersParams) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) PostChannelsChannelIdWatchers(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) GetChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) DeleteChannelsChannelIdWatchersWatcherId(w http.ResponseWriter, r *http.Request, channelId openapiTypes.UUID, watcherId openapiTypes.UUID) {
	// Commented out - not part of channels/operations service
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *MockServer) PostAccounts(w http.ResponseWriter, r *http.Request) {
	var in stdserver.CreateAccount
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uuid.New()
	acc := stdserver.Account{
		AccountId: id,
		Address:   strings.ToLower(in.Address),
		ChainId:   in.ChainId,
		Name:      in.Name,
	}
	s.accounts = append(s.accounts, acc)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(acc)
}

func (s *MockServer) GetAccounts(w http.ResponseWriter, r *http.Request, params stdserver.GetAccountsParams) {
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
	if end > len(s.accounts) {
		end = len(s.accounts)
	}
	data := []stdserver.Account{}
	if offset < len(s.accounts) {
		data = s.accounts[offset:end]
	}
	hasMore := end < len(s.accounts)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stdserver.AccountList{Data: data, HasMore: hasMore})
}

func (s *MockServer) GetAccountsAccountId(w http.ResponseWriter, r *http.Request, accountId openapiTypes.UUID) {
	for _, a := range s.accounts {
		if a.AccountId == accountId {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(a)
			return
		}
	}
	http.Error(w, "account not found", http.StatusNotFound)
}

func (s *MockServer) PatchAccountsAccountId(w http.ResponseWriter, r *http.Request, accountId openapiTypes.UUID) {
	var request stdserver.UpdateAccount
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find and update the account
	for i, a := range s.accounts {
		if a.AccountId == accountId {
			s.accounts[i].Name = &request.Name
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(s.accounts[i])
			return
		}
	}
	http.Error(w, "account not found", http.StatusNotFound)
}
