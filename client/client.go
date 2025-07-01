package client

// CVNClient is a client for the CVN API.
type CVNClient = ClientWithResponses

// NewCVNClient creates a new CVN client with the specified base URL.
// It returns a pointer to the CVNClient and an error if any issues occur during initialization.
//   - baseURL: The base URL of the CVN API.
func NewCVNClient(baseURL string) (*CVNClient, error) {
	return NewClientWithResponses(baseURL)
}
