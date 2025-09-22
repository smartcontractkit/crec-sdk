package client

import (
	"net/http"

	"github.com/moul/http2curl"
	"github.com/rs/zerolog"
	apiClient "github.com/smartcontractkit/cvn-api-go/client"
)

type HTTPClientWithCURLLogger struct {
	Client *http.Client
	Logger *zerolog.Logger
}

var _ apiClient.HttpRequestDoer = (*HTTPClientWithCURLLogger)(nil)

func NewHTTPClientWithCURLLogger(logger *zerolog.Logger, httpClient *http.Client) *HTTPClientWithCURLLogger {
	if httpClient == nil {
		return &HTTPClientWithCURLLogger{Logger: logger, Client: &http.Client{}}
	} else {
		return &HTTPClientWithCURLLogger{Logger: logger, Client: httpClient}
	}
}

func (l *HTTPClientWithCURLLogger) Do(req *http.Request) (*http.Response, error) {
	l.logCurl(req)
	return l.Client.Do(req)
}

func (l *HTTPClientWithCURLLogger) logCurl(req *http.Request) {
	// Clone to avoid mutating original request
	clone := req.Clone(req.Context())

	// Redact sensitive headers
	sensitiveHeaders := []string{"Api-Key", "Authorization", "Cookie"}
	for _, h := range sensitiveHeaders {
		clone.Header.Del(h)
	}

	// Generate curl command and log at debug level
	if curl, err := http2curl.GetCurlCommand(clone); err == nil {
		l.Logger.Debug().
			Str("curl", curl.String()).
			Msg("Outgoing HTTP request dump")
	}
}
