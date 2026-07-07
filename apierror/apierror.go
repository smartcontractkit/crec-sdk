// Package apierror defines cross-cutting error values that the CREC SDK returns
// when the API responds with a recognizable application-level error.
package apierror

import (
	"errors"
	"fmt"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// ErrOrganizationNotFound is returned when the CREC API reports that the
// authenticated organization is not onboarded in CRE Connect (HTTP 401 with an
// ApplicationError of type ORGANIZATION_NOT_FOUND).
var ErrOrganizationNotFound = errors.New("organization not found")

// Canonical not-found sentinels for HTTP 404 responses. The API disambiguates
// which resource was missing via ApplicationError.code; domain packages alias
// these so consumers can match with errors.Is regardless of the calling package.
var (
	// ErrChannelNotFound is returned when the channel does not exist.
	ErrChannelNotFound = errors.New("channel not found")
	// ErrWalletNotFound is returned when the referenced wallet does not exist.
	ErrWalletNotFound = errors.New("wallet not found")
	// ErrOperationNotFound is returned when the operation does not exist.
	ErrOperationNotFound = errors.New("operation not found")
	// ErrWatcherNotFound is returned when the watcher does not exist.
	ErrWatcherNotFound = errors.New("watcher not found")
	// ErrQueryNotFound is returned when the query does not exist.
	ErrQueryNotFound = errors.New("query not found")
)

// ErrUnexpectedStatusCode is returned when the API responds with an HTTP status
// the SDK does not handle explicitly.
var ErrUnexpectedStatusCode = errors.New("unexpected status code")

// ErrNilResponse is returned when the API response is nil.
var ErrNilResponse = errors.New("unexpected nil response")

// ErrNilResponseBody is returned when the API response body is nil.
var ErrNilResponseBody = errors.New("unexpected nil response body")

// FromApplicationError maps a known ApplicationError to its canonical SDK
// sentinel error, or returns nil when the error has no dedicated mapping.
func FromApplicationError(appErr *apiClient.ApplicationError) error {
	if appErr == nil {
		return nil
	}

	switch appErr.Type {
	case apiClient.ORGANIZATIONNOTFOUND:
		return ErrOrganizationNotFound
	default:
		return nil
	}
}

// Wrap returns an error wrapping opErr with the canonical sentinel for appErr's
// type. If appErr has no dedicated mapping, it falls back to a generic
// unexpected-status error wrapping [ErrUnexpectedStatusCode] with statusCode
// for diagnostics.
func Wrap(appErr *apiClient.ApplicationError, opErr error, statusCode int) error {
	if mapped := FromApplicationError(appErr); mapped != nil {
		return fmt.Errorf("%w: %w", opErr, mapped)
	}
	return fmt.Errorf("%w: %w (status code %d)", opErr, ErrUnexpectedStatusCode, statusCode)
}

// NotFound maps a 404 ApplicationError to its canonical not-found sentinel based
// on ApplicationError.code, or returns nil when the code is missing or
// unrecognized (forward-compatible for codes added after this SDK release).
func NotFound(appErr *apiClient.ApplicationError) error {
	if appErr == nil || appErr.Code == nil {
		return nil
	}

	switch *appErr.Code {
	case apiClient.ApplicationErrorCodeChannelNotFound:
		return ErrChannelNotFound
	case apiClient.ApplicationErrorCodeWalletNotFound:
		return ErrWalletNotFound
	case apiClient.ApplicationErrorCodeOperationNotFound:
		return ErrOperationNotFound
	case apiClient.ApplicationErrorCodeWatcherNotFound:
		return ErrWatcherNotFound
	case apiClient.ApplicationErrorCodeQueryNotFound:
		return ErrQueryNotFound
	default:
		return nil
	}
}

// WrapNotFound resolves a 404 ApplicationError to its canonical not-found
// sentinel and wraps it with opErr when ApplicationError.code is recognized.
// When the code is missing or unknown, it returns only opErr so callers are not
// given a wrong typed sentinel. detail is appended when the server message is
// absent (for example channel and resource IDs). Use this only for endpoints
// whose 404 can mean more than one thing; for single-cause endpoints wrap the
// sole sentinel directly.
func WrapNotFound(appErr *apiClient.ApplicationError, opErr error, detail string) error {
	msg := notFoundMessage(appErr, detail)

	if mapped := NotFound(appErr); mapped != nil {
		if msg != "" {
			return fmt.Errorf("%w: %w: %s", opErr, mapped, msg)
		}
		return fmt.Errorf("%w: %w", opErr, mapped)
	}
	if msg != "" {
		return fmt.Errorf("%w: %s", opErr, msg)
	}
	return opErr
}

func notFoundMessage(appErr *apiClient.ApplicationError, detail string) string {
	if appErr != nil && appErr.Message != "" {
		return appErr.Message
	}
	return detail
}
