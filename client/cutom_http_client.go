package client

import (
	"net/http"

	"github.com/moul/http2curl"
	"github.com/rs/zerolog"
	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

type CustomHttpClient struct {
	apiClient.HttpRequestDoer
	Client *http.Client
	Logger zerolog.Logger
}

func NewCustomHttpClient(logger zerolog.Logger) *CustomHttpClient {
	return &CustomHttpClient{Logger: logger, Client: &http.Client{}}
}

func (l *CustomHttpClient) Do(req *http.Request) (*http.Response, error) {
	if curl, err := http2curl.GetCurlCommand(req); err == nil {
		l.Logger.Debug().
			Str("curl", curl.String()).
			Msg("Outgoing HTTP request dump")
	}

	return l.Client.Do(req)
}
