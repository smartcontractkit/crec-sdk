package apierrors

import (
	"fmt"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// NotFoundResponse holds 404 fields shared by oapi-codegen response types.
type NotFoundResponse struct {
	JSON404 *apiClient.ApplicationError
	Body    []byte
}

// NotFoundMapping configures how ApplicationError.code maps to package sentinel errors.
type NotFoundMapping struct {
	Channel   error
	Wallet    error
	Operation error
	Watcher   error
	Query     error
	Empty     func() error
	Unknown   func(msg string) error
}

// MapNotFound builds a typed not-found error from an API 404 response.
func MapNotFound(resp NotFoundResponse, mapping NotFoundMapping) error {
	msg := Message(resp.JSON404, resp.Body)
	if msg == "" {
		if mapping.Empty != nil {
			return mapping.Empty()
		}
		return mapping.unknown(msg)
	}

	switch NotFoundKindFromResponse(resp.JSON404, resp.Body) {
	case NotFoundChannel:
		if mapping.Channel != nil {
			return fmt.Errorf("%w: %s", mapping.Channel, msg)
		}
	case NotFoundWallet:
		if mapping.Wallet != nil {
			return fmt.Errorf("%w: %s", mapping.Wallet, msg)
		}
	case NotFoundOperation:
		if mapping.Operation != nil {
			return fmt.Errorf("%w: %s", mapping.Operation, msg)
		}
	case NotFoundWatcher:
		if mapping.Watcher != nil {
			return fmt.Errorf("%w: %s", mapping.Watcher, msg)
		}
	case NotFoundQuery:
		if mapping.Query != nil {
			return fmt.Errorf("%w: %s", mapping.Query, msg)
		}
	case NotFoundUnknown:
	}

	return mapping.unknown(msg)
}

func (m NotFoundMapping) unknown(msg string) error {
	if m.Unknown != nil {
		return m.Unknown(msg)
	}
	if msg == "" {
		return fmt.Errorf("not found")
	}
	return fmt.Errorf("not found: %s", msg)
}
