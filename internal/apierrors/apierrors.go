package apierrors

import (
	"encoding/json"
	"strings"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// NotFoundKind classifies a NOT_FOUND API response by resource type.
type NotFoundKind int

const (
	NotFoundUnknown NotFoundKind = iota
	NotFoundChannel
	NotFoundWallet
	NotFoundOperation
	NotFoundWatcher
	NotFoundQuery
)

type errorEnvelope struct {
	Message string                          `json:"message"`
	Error   string                          `json:"error"`
	Code    *apiClient.ApplicationErrorCode `json:"code"`
}

// Message returns the best available human-readable error text from an API response.
func Message(apiErr *apiClient.ApplicationError, body []byte) string {
	if apiErr != nil && strings.TrimSpace(apiErr.Message) != "" {
		return strings.TrimSpace(apiErr.Message)
	}
	if len(body) == 0 {
		return ""
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil {
		if msg := strings.TrimSpace(env.Message); msg != "" {
			return msg
		}
		if msg := strings.TrimSpace(env.Error); msg != "" {
			return msg
		}
	}
	return trimmed
}

// NotFoundKindFromResponse resolves the not-found resource type from ApplicationError.code.
func NotFoundKindFromResponse(apiErr *apiClient.ApplicationError, body []byte) NotFoundKind {
	kind, _ := notFoundKindFromCode(codeFromResponse(apiErr, body))
	return kind
}

func codeFromResponse(apiErr *apiClient.ApplicationError, body []byte) *apiClient.ApplicationErrorCode {
	if apiErr != nil && apiErr.Code != nil {
		return apiErr.Code
	}
	if len(body) == 0 {
		return nil
	}
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Code != nil {
		return env.Code
	}
	return nil
}

func notFoundKindFromCode(code *apiClient.ApplicationErrorCode) (NotFoundKind, bool) {
	if code == nil {
		return NotFoundUnknown, false
	}
	switch *code {
	case apiClient.ApplicationErrorCodeChannelNotFound:
		return NotFoundChannel, true
	case apiClient.ApplicationErrorCodeWalletNotFound:
		return NotFoundWallet, true
	case apiClient.ApplicationErrorCodeOperationNotFound:
		return NotFoundOperation, true
	case apiClient.ApplicationErrorCodeWatcherNotFound:
		return NotFoundWatcher, true
	case apiClient.ApplicationErrorCodeQueryNotFound:
		return NotFoundQuery, true
	default:
		return NotFoundUnknown, false
	}
}
