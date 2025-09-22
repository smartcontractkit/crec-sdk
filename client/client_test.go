package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCVNClient_HealthCheck_AddsAPIKeyHeader(t *testing.T) {
	const apiKey = "MY-DEV-KEY"

	// Minimal test server that validates Api-Key header and returns a HC payload
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
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
			},
		),
	)
	defer ts.Close()

	c, err := NewCVNClient(
		&ClientOptions{
			BaseURL: ts.URL,
			APIKey:  apiKey,
		},
	)
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

type roundtripLogger struct {
}

func (r roundtripLogger) RoundTrip(request *http.Request) (*http.Response, error) {
	fmt.Println("request", request.URL.String())
	res, err := http.DefaultTransport.RoundTrip(request)
	var bufReader bytes.Buffer
	io.Copy(&bufReader, res.Body)
	fmt.Println("response", bufReader.String())
	// res.Body is already closed. you need to make a copy again to pass it for the code
	res.Body = io.NopCloser(bytes.NewReader(bufReader.Bytes()))
	return res, err
}

func TestNewCVNClient_HealthCheck_AddsAPIKeyHeader_CustomHTTPClient(t *testing.T) {
	const apiKey = "MY-DEV-KEY"

	// Minimal test server that validates Api-Key header and returns a HC payload
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
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
			},
		),
	)
	defer ts.Close()

	httpClient := &http.Client{Transport: roundtripLogger{}}

	c, err := NewCVNClient(
		&ClientOptions{
			BaseURL:    ts.URL,
			APIKey:     apiKey,
			HTTPClient: httpClient,
		},
	)
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
