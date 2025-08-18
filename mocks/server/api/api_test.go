package api_test

import (
	"net/http"
	"testing"

	mockserver "github.com/smartcontractkit/cvn-sdk/mocks/server"
)

func TestStdHTTPServerHandler_HealthCheck(t *testing.T) {
	// Re-use the fully implemented mock server (it uses the generated std-http server handler)
	s := mockserver.NewMockServer()
	defer s.Close()

	resp, err := http.Get(s.TestServer.URL + "/health-check")
	if err != nil {
		t.Fatalf("GET /health-check: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}
