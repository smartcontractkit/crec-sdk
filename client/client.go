package client

type CVNClient = ClientWithResponses

func NewCVNClient(baseURL string) (*CVNClient, error) {
	return NewClientWithResponses(baseURL)
}
