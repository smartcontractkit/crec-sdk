package client

import (
	"context"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

// CVNClient is a client for the CVN API.
type CVNClient = apiClient.ClientWithResponses

// NewCVNClient creates a new CVN client with the specified base URL.
// It returns a pointer to the CVNClient and an error if any issues occur during initialization.
//   - baseURL: The base URL of the CVN API.
func NewCVNClient(baseURL string, apiKey string) (*CVNClient, error) {
	apiKeyHeaderEditor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Api-Key", apiKey)
		return nil
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	customHttpClient := NewHTTPClientWithCURLLogger(logger)

	return apiClient.NewClientWithResponses(baseURL,
		apiClient.WithRequestEditorFn(apiKeyHeaderEditor),
		apiClient.WithHTTPClient(customHttpClient),
	)
}
