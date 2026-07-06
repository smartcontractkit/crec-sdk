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
