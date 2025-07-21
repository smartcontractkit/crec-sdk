package client

import (
	"context"
	"net/http"
)

// CVNClient is a client for the CVN API.
type CVNClient = ClientWithResponses

// NewCVNClient creates a new CVN client with the specified base URL.
// It returns a pointer to the CVNClient and an error if any issues occur during initialization.
//   - baseURL: The base URL of the CVN API.
func NewCVNClient(baseURL string, apiKey string) (*CVNClient, error) {
	apiKeyHeaderEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", apiKey)
		return nil
	}
	return NewClientWithResponses(baseURL, WithRequestEditorFn(apiKeyHeaderEditor))
}
