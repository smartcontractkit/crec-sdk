package client

import (
	"net/http"
	"regexp"

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
	// Generate curl command
	if curl, err := http2curl.GetCurlCommand(req); err == nil {
		// Redact sensitive headers
		filteredCURL := filterCURL(curl.String())

		l.Logger.Debug().
			Str("curl", filteredCURL).
			Msg("Outgoing HTTP request dump")
	}
}

func filterCURL(curl string) string {
	sensitiveHeaders := []string{"Api-Key", "Authorization", "Cookie"}

	for _, h := range sensitiveHeaders {
		// Regex pattern to match: -H 'Header-Name: some-value'
		re := regexp.MustCompile(`(?i)(-H\s+'` + h + `: )[^']+'`)
		// Replace actual value with masked version
		curl = re.ReplaceAllString(curl, `${1}REDACTED'`)
	}

	return curl
}
