package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/smartcontractkit/cvn-sdk/mocks/events"
	"github.com/smartcontractkit/cvn-sdk/mocks/server/api"
)

type MockServer struct {
	TestServer *httptest.Server
	eventsFile string
}

var _ api.ServerInterface = (*MockServer)(nil)

func NewMockServer() *MockServer {
	s := &MockServer{
		eventsFile: "event_list.json",
	}
	r := http.NewServeMux()
	h := api.HandlerFromMux(s, r)

	s.TestServer = httptest.NewServer(h)
	return s
}

func (s *MockServer) SetEventsResponse(filename string) {
	s.eventsFile = filename
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
		api.HealthCheck{
			Status: "ok",
		},
	)
}

func (s *MockServer) PostListeners(w http.ResponseWriter, r *http.Request) {
	var createListenerReq api.CreateListener
	err := json.NewDecoder(r.Body).Decode(&createListenerReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	listenerId, err := uuid.NewUUID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	listener := api.Listener{
		ListenerId: listenerId,
		Service:    createListenerReq.Service,
		Name:       createListenerReq.Name,
		ChainId:    createListenerReq.ChainId,
		Address:    createListenerReq.Address,
		CreatedAt:  time.Now().Unix(),
		Status:     "active",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(listener)
}

func (s *MockServer) GetListeners(w http.ResponseWriter, r *http.Request, params api.GetListenersParams) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(
		api.ListenerList{
			Data:    []api.Listener{},
			HasMore: false,
		},
	)
}

func (s *MockServer) GetListenersListenerId(w http.ResponseWriter, r *http.Request, listenerId openapi_types.UUID) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		api.Listener{
			ListenerId: listenerId,
			Service:    "dvp",
			Name:       "SettlementAccepted",
			ChainId:    "1337",
			Address:    "0x1234567890abcdef1234567890abcdef12345678",
			Status:     "active",
		},
	)
}

func (s *MockServer) DeleteListenersListenerId(w http.ResponseWriter, r *http.Request, listenerId openapi_types.UUID) {
	w.WriteHeader(http.StatusAccepted)
}

func (s *MockServer) PostEvents(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *MockServer) GetEvents(w http.ResponseWriter, r *http.Request, params api.GetEventsParams) {
	var eventResponse api.EventList
	err := events.LoadJson(s.eventsFile, &eventResponse)
	if err != nil {
		log.Fatalf("Failed to load event data: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(eventResponse)
}

func (s *MockServer) PostOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := `{
	"account": "",
	"account_id": "df72db52-fa74-4dd8-8c53-a9742be70caa",
	"account_operation_id": "1755036418",
	"chain_id": "",
	"created_at": 1755036421,
	"operation_id": "a47760fc-6eae-456f-84b9-e6e349d13281",
	"signature": "0xe227e4533ea6198d58d167971f13733239447b0a298416355a226acb07cc2e5c3efff04c41b32a1a2fefdbf5aba4f15312c527be249d19756850e61b269fcecd1c",
	"status": "pending",
	"transactions": [
		{
			"data": "0x40c10f19000000000000000000000000f32efccd3087563c226ef51c7e659888de78c59c0000000000000000000000000000000000000000000000000000000000087a23",
			"to": "0x79009066202906e3d1644029d0B6e441370fD865",
			"value": "0"
		}
	]
}`

	w.Write([]byte(response))
}

func (s *MockServer) GetOperations(w http.ResponseWriter, r *http.Request, params api.GetOperationsParams) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(
		api.OperationList{
			Data:    []api.Operation{},
			HasMore: false,
		},
	)
}

func (s *MockServer) GetOperationsOperationId(
	w http.ResponseWriter, r *http.Request, operationId openapi_types.UUID,
) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.Operation{})
}

func (s *MockServer) PostOperationStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
