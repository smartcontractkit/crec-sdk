package client

import (
	"log"
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
	if l.Logger.Debug().Enabled() {

		curl, _ := http2curl.GetCurlCommand(req)
		log.Println(curl.String())
		/*if curlCmd, err := DumpRequestAsCurl(req); err == nil {
			l.Logger.Debug().Msg(curlCmd)
		}*/
	}

	return l.Client.Do(req)
}

/*
func DumpRequestAsCurl(req *http.Request) (string, error) {
	var curlCmd strings.Builder

	// Start with method + URL
	fmt.Fprintf(&curlCmd, "curl -X %s '%s'", req.Method, req.URL.String())

	// Headers
	for name, values := range req.Header {
		for _, v := range values {
			fmt.Fprintf(&curlCmd, " -H '%s: %s'", name, v)
		}
	}

	// Body (if any, and rewindable)
	if req.Body != nil {
		// Copy body so we don’t consume the original
		var buf bytes.Buffer
		bodyCopy := io.TeeReader(req.Body, &buf)

		b, err := io.ReadAll(bodyCopy)
		if err != nil {
			return "", err
		}

		// restore body for actual request sending
		req.Body = io.NopCloser(&buf)

		if len(b) > 0 {
			fmt.Fprintf(&curlCmd, " --data '%s'", string(b))
		}
	}

	return curlCmd.String(), nil
}
*/
