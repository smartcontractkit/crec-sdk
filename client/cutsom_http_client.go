package client

import (
	"net/http"

	"github.com/moul/http2curl"
	"github.com/rs/zerolog"
	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

type CustomHTTPClient struct {
	Client *http.Client
	Logger zerolog.Logger
}

var _ apiClient.HttpRequestDoer = (*CustomHTTPClient)(nil)

func NewCustomHTTPClient(logger zerolog.Logger) *CustomHTTPClient {
	return &CustomHTTPClient{Logger: logger, Client: &http.Client{}}
}

func (l *CustomHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if curl, err := http2curl.GetCurlCommand(req); err == nil {
		l.Logger.Debug().
			Str("curl", curl.String()).
			Msg("Outgoing HTTP request dump")
	}

	return l.Client.Do(req)
}
