package mockserver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/smartcontractkit/cvn-sdk/internal/mockdata"
	"github.com/smartcontractkit/cvn-sdk/internal/mockserver/api"
)

type MockServer struct {
	TestServer *httptest.Server
	eventsFile string
}

func NewMockServer(t *testing.T) *MockServer {
	s := &MockServer{
		eventsFile: "event_list.json",
	}
	r := http.NewServeMux()
	h := api.HandlerFromMux(s, r)

	s.TestServer = httptest.NewServer(h)

	t.Logf("Mock server started at URL: %s", s.TestServer.URL)

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
	w.WriteHeader(http.StatusOK)
}

func (s *MockServer) GetEvents(w http.ResponseWriter, r *http.Request, params api.GetEventsParams) {
	var eventResponse api.EventResponse
	err := mockdata.LoadJson(s.eventsFile, &eventResponse)
	if err != nil {
		log.Fatalf("Failed to load event data: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(eventResponse)
}

func (s *MockServer) PostEvents(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *MockServer) PostListener(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *MockServer) PostOperationSend(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *MockServer) GetOperationStatus(w http.ResponseWriter, r *http.Request, params api.GetOperationStatusParams) {
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode([]api.Operation{})

}

func (s *MockServer) PostOperationStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

func (s *MockServer) GetOperationStatusOperationId(
	w http.ResponseWriter, r *http.Request, operationId openapi_types.UUID,
) {
	_ = json.NewEncoder(w).Encode(api.Operation{})
}
