// Package retry contains internal helpers used by SDK sub-clients that poll
// the CREC API. It is not a stable surface and may change without notice.
package retry

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// statusCodePattern extracts an HTTP status code from a wrapped error string
// produced by the generated client (e.g. "... status code: 503 ...").
var statusCodePattern = regexp.MustCompile(`status code:?\s*(\d{3})`)

// transientNetworkIndicators are lowercase substrings that flag a transport
// problem (transient by nature) in errors that do not carry an HTTP status.
var transientNetworkIndicators = []string{
	"connection refused",
	"connection reset",
	"timeout",
	"temporary failure",
	"eof",
	"broken pipe",
	"no such host",
	"network is unreachable",
}

// IsTransient reports whether an error from a CREC API call is worth retrying.
// It treats context cancellation/deadline as permanent and otherwise checks for
// retriable HTTP status codes (429, 5xx) and well-known network failure
// indicators.
//
// Callers should fast-path their own validation/sentinel errors before calling
// this helper.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	if matches := statusCodePattern.FindStringSubmatch(errMsg); len(matches) > 1 {
		if statusCode, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			return IsTransientStatusCode(statusCode)
		}
	}

	for _, indicator := range transientNetworkIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}

	return false
}

// IsTransientStatusCode reports whether an HTTP status code represents a
// transient failure that should be retried.
func IsTransientStatusCode(statusCode int) bool {
	switch {
	case statusCode == 429: // Too Many Requests (rate limiting)
		return true
	case statusCode >= 500 && statusCode < 600: // 5xx Server Errors
		return true
	default:
		return false
	}
}
