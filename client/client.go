package client

import (
	"context"
	"fmt"
	"net/http"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// ClientOptions defines the options for creating a new CREC client used to interact with the CREC API.
//   - BaseURL: The base URL of the CREC API.
//   - ApiKey: The API key for authenticating with the CREC API.
//   - HttpClient: The custom HTTP client to use for making requests. If nil, the default HTTP client is used.
type ClientOptions struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client // Optional custom HTTP client
}

// CRECClient is a client for the CREC API.
type CRECClient = apiClient.ClientWithResponses

// NewCRECClient creates a new CREC client with the given options.
//   - opts: Options for configuring the CREC client, see ClientOptions for details.
func NewCRECClient(opts *ClientOptions) (*CRECClient, error) {
	if opts == nil {
		return nil, fmt.Errorf("ClientOptions is required")
	}
	apiKeyHeaderEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", opts.APIKey)
		return nil
	}

	if opts.HTTPClient == nil {
		return apiClient.NewClientWithResponses(
			opts.BaseURL,
			apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
		)
	} else {
		return apiClient.NewClientWithResponses(
			opts.BaseURL,
			apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
			apiClient.WithHTTPClient(opts.HTTPClient),
		)
	}
}
