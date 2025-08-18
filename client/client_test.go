package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCVNClient_HealthCheck_AddsAPIKeyHeader(t *testing.T) {
	const apiKey = "MY-DEV-KEY"

	// Minimal test server that validates Api-Key header and returns a HC payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health-check" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Api-Key"); got != apiKey {
			http.Error(w, "missing or wrong Api-Key header", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer ts.Close()

	c, err := NewCVNClient(ts.URL, apiKey)
	if err != nil {
		t.Fatalf("NewCVNClient: %v", err)
	}

	resp, err := c.GetHealthCheckWithResponse(context.Background())
	if err != nil {
		t.Fatalf("GetHealthCheckWithResponse: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("health-check status: %d", resp.StatusCode())
	}
	if resp.JSON200 == nil || resp.JSON200.Status != "ok" {
		t.Fatalf("unexpected payload: %+v", resp.JSON200)
	}
}
