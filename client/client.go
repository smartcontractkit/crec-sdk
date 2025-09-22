package client

import (
	"context"
	"net/http"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

// CVNClient is a client for the CVN API.
type CVNClient = apiClient.ClientWithResponses

// NewCVNClient creates a new CVN client with the specified base URL and API key.
// It returns a pointer to the CVNClient and an error if any issues occur during initialization.
//   - baseURL: The base URL of the CVN API.
//   - apiKey: The API key for authenticating with the CVN API.
func NewCVNClient(baseURL string, apiKey string) (*CVNClient, error) {
	return NewCVNClientWithHTTPClient(baseURL, apiKey, nil)
}

// NewCVNClientWithHTTPClient creates a new CVN client with the specified base URL and API key, using the provided HTTP client.
// It returns a pointer to the CVNClient and an error if any issues occur during initialization.
//   - baseURL: The base URL of the CVN API.
//   - apiKey: The API key for authenticating with the CVN API.
//   - httpClient: The custom HTTP client to use for making requests. If nil, the default HTTP client is used.
func NewCVNClientWithHTTPClient(baseURL, apiKey string, httpClient *http.Client) (*CVNClient, error) {
	apiKeyHeaderEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", apiKey)
		return nil
	}

	if httpClient == nil {
		return apiClient.NewClientWithResponses(
			baseURL,
			apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
		)
	} else {
		return apiClient.NewClientWithResponses(
			baseURL,
			apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
			apiClient.WithHTTPClient(httpClient),
		)
	}
}
