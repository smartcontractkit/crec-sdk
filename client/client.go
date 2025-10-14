package client

import (
	"context"
	"fmt"
	"net/http"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// ClientOptions defines the options for creating a new CREc client used to interact with the CREc API.
//   - BaseURL: The base URL of the CREc API.
//   - ApiKey: The API key for authenticating with the CREc API.
//   - HttpClient: The custom HTTP client to use for making requests. If nil, the default HTTP client is used.
type ClientOptions struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client // Optional custom HTTP client
}

// CREcClient is a client for the CREc API.
type CREcClient = apiClient.ClientWithResponses

// NewCREcClient creates a new CREc client with the given options.
//   - opts: Options for configuring the CREc client, see ClientOptions for details.
func NewCREcClient(opts *ClientOptions) (*CREcClient, error) {
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
