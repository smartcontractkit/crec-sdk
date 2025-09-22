package client

import (
	"context"
	"fmt"
	"net/http"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

// ClientOptions defines the options for creating a new CVN client used to interact with the CVN API.
//   - BaseURL: The base URL of the CVN API.
//   - ApiKey: The API key for authenticating with the CVN API.
//   - HttpClient: The custom HTTP client to use for making requests. If nil, the default HTTP client is used.
type ClientOptions struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client // Optional custom HTTP client
}

// CVNClient is a client for the CVN API.
type CVNClient = apiClient.ClientWithResponses

// NewCVNClient creates a new CVN client with the given options.
//   - opts: Options for configuring the CVN client, see ClientOptions for details.
func NewCVNClient(opts *ClientOptions) (*CVNClient, error) {
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
