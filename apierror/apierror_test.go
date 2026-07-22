package apierror_test

import (
	"errors"
	"net/http"
	"testing"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/crec-sdk/apierror"
)

func TestApierror_FromApplicationError(t *testing.T) {
	tests := []struct {
		name    string
		appErr  *apiClient.ApplicationError
		wantErr error
	}{
		{
			name:    "nil application error returns nil",
			appErr:  nil,
			wantErr: nil,
		},
		{
			name:    "organization not found maps to sentinel",
			appErr:  &apiClient.ApplicationError{Type: apiClient.ORGANIZATIONNOTFOUND, Message: "organization not found"},
			wantErr: apierror.ErrOrganizationNotFound,
		},
		{
			name:    "unknown future type degrades to nil",
			appErr:  &apiClient.ApplicationError{Type: "SOME_FUTURE_TYPE", Message: "new"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apierror.FromApplicationError(tt.appErr)

			if tt.wantErr == nil {
				assert.NoError(t, got)
				return
			}
			assert.ErrorIs(t, got, tt.wantErr)
		})
	}
}

func TestApierror_NotFound(t *testing.T) {
	channelCode := apiClient.ApplicationErrorCodeChannelNotFound
	walletCode := apiClient.ApplicationErrorCodeWalletNotFound
	operationCode := apiClient.ApplicationErrorCodeOperationNotFound
	watcherCode := apiClient.ApplicationErrorCodeWatcherNotFound
	queryCode := apiClient.ApplicationErrorCodeQueryNotFound
	futureCode := apiClient.ApplicationErrorCode("SOME_FUTURE_NOT_FOUND")

	tests := []struct {
		name    string
		appErr  *apiClient.ApplicationError
		wantErr error
	}{
		{name: "nil application error returns nil", appErr: nil, wantErr: nil},
		{name: "nil code returns nil", appErr: &apiClient.ApplicationError{Type: apiClient.NOTFOUND}, wantErr: nil},
		{name: "channel", appErr: &apiClient.ApplicationError{Code: &channelCode}, wantErr: apierror.ErrChannelNotFound},
		{name: "wallet", appErr: &apiClient.ApplicationError{Code: &walletCode}, wantErr: apierror.ErrWalletNotFound},
		{name: "operation", appErr: &apiClient.ApplicationError{Code: &operationCode}, wantErr: apierror.ErrOperationNotFound},
		{name: "watcher", appErr: &apiClient.ApplicationError{Code: &watcherCode}, wantErr: apierror.ErrWatcherNotFound},
		{name: "query", appErr: &apiClient.ApplicationError{Code: &queryCode}, wantErr: apierror.ErrQueryNotFound},
		{name: "unknown future code degrades to nil", appErr: &apiClient.ApplicationError{Code: &futureCode}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apierror.NotFound(tt.appErr)

			if tt.wantErr == nil {
				assert.NoError(t, got)
				return
			}
			assert.ErrorIs(t, got, tt.wantErr)
		})
	}
}

func TestApierror_WrapNotFound(t *testing.T) {
	opErr := errors.New("failed to get watcher")
	channelCode := apiClient.ApplicationErrorCodeChannelNotFound
	futureCode := apiClient.ApplicationErrorCode("SOME_FUTURE_NOT_FOUND")
	detail := "channel ID abc, watcher ID def"

	t.Run("recognized code wraps opErr with typed sentinel and server message", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &channelCode, Message: "channel with ID abc not found"}
		err := apierror.WrapNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.NotErrorIs(t, err, apierror.ErrWatcherNotFound)
		assert.Contains(t, err.Error(), "channel with ID abc not found")
		assert.NotContains(t, err.Error(), detail)
	})

	t.Run("recognized code uses detail when server message is empty", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &channelCode}
		err := apierror.WrapNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.Contains(t, err.Error(), detail)
	})

	t.Run("missing code returns only opErr with server message", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Type: apiClient.NOTFOUND, Message: "not found"}
		err := apierror.WrapNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrWatcherNotFound)
		assert.NotErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.Contains(t, err.Error(), "not found")
		assert.NotContains(t, err.Error(), detail)
	})

	t.Run("missing code uses detail when server message is empty", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Type: apiClient.NOTFOUND}
		err := apierror.WrapNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrWatcherNotFound)
		assert.Contains(t, err.Error(), detail)
	})

	t.Run("unknown future code returns only opErr with detail", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &futureCode}
		err := apierror.WrapNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrWatcherNotFound)
		assert.NotErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.Contains(t, err.Error(), detail)
	})

	t.Run("nil application error uses detail without panicking", func(t *testing.T) {
		err := apierror.WrapNotFound(nil, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrWatcherNotFound)
		assert.Contains(t, err.Error(), detail)
	})
}

func TestApierror_WrapChannelNotFound(t *testing.T) {
	opErr := errors.New("failed to list watchers")
	detail := "channel ID abc"

	t.Run("wraps opErr with channel sentinel and server message", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Message: "channel with ID abc not found"}
		err := apierror.WrapChannelNotFound(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.Contains(t, err.Error(), "channel with ID abc not found")
	})

	t.Run("uses detail when server message is empty", func(t *testing.T) {
		err := apierror.WrapChannelNotFound(nil, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrChannelNotFound)
		assert.Contains(t, err.Error(), detail)
	})
}

func TestApierror_NotFoundWarnMessage(t *testing.T) {
	channelCode := apiClient.ApplicationErrorCodeChannelNotFound

	t.Run("uses code when present", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &channelCode}
		msg := apierror.NotFoundWarnMessage(appErr, "getting operation", nil)
		assert.Equal(t, "channel not found when getting operation", msg)
	})

	t.Run("uses fallback for single-cause endpoints without code", func(t *testing.T) {
		msg := apierror.NotFoundWarnMessage(nil, "listing watchers", apierror.ErrChannelNotFound)
		assert.Equal(t, "channel not found when listing watchers", msg)
	})

	t.Run("generic when no code and no fallback", func(t *testing.T) {
		msg := apierror.NotFoundWarnMessage(nil, "creating operation", nil)
		assert.Equal(t, "Resource not found when creating operation", msg)
	})
}

func TestApierror_NotFoundCode(t *testing.T) {
	channelCode := apiClient.ApplicationErrorCodeChannelNotFound
	assert.Equal(t, "", apierror.NotFoundCode(nil))
	assert.Equal(t, "CHANNEL_NOT_FOUND", apierror.NotFoundCode(&apiClient.ApplicationError{Code: &channelCode}))
}

func TestApierror_Wrap(t *testing.T) {
	opErr := errors.New("operation failed")

	t.Run("mapped sentinel wraps opErr", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Type: apiClient.ORGANIZATIONNOTFOUND, Message: "organization not found"}
		err := apierror.Wrap(appErr, opErr, http.StatusUnauthorized)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrOrganizationNotFound)
		assert.NotErrorIs(t, err, apierror.ErrUnexpectedStatusCode)
	})

	t.Run("unmapped type falls back to unexpected-status error", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Type: "SOME_FUTURE_TYPE", Message: "new"}
		err := apierror.Wrap(appErr, opErr, http.StatusUnauthorized)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrUnexpectedStatusCode)
		assert.NotErrorIs(t, err, apierror.ErrOrganizationNotFound)
	})

	t.Run("nil application error falls back to unexpected-status error", func(t *testing.T) {
		err := apierror.Wrap(nil, opErr, http.StatusUnauthorized)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrUnexpectedStatusCode)
	})
}

func TestApierror_Conflict(t *testing.T) {
	channelCode := apiClient.ApplicationErrorCodeChannelAlreadyExists
	walletCode := apiClient.ApplicationErrorCodeWalletAlreadyExists
	watcherCode := apiClient.ApplicationErrorCodeWatcherAlreadyExists
	idempotencyCode := apiClient.ApplicationErrorCodeIdempotencyKeyMismatch
	notFinalizableCode := apiClient.ApplicationErrorCodeOperationNotFinalizable
	notCancellableCode := apiClient.ApplicationErrorCodeOperationNotCancellable
	deadlineCode := apiClient.ApplicationErrorCodeOperationDeadlineElapsed
	versionCode := apiClient.ApplicationErrorCodeResourceVersionConflict
	archivedCode := apiClient.ApplicationErrorCodeWalletAlreadyArchived
	chainUnavailableCode := apiClient.ApplicationErrorCodeChainUnavailable
	futureCode := apiClient.ApplicationErrorCode("SOME_FUTURE_CONFLICT")

	tests := []struct {
		name    string
		appErr  *apiClient.ApplicationError
		wantErr error
	}{
		{name: "nil application error returns nil", appErr: nil, wantErr: nil},
		{name: "nil code returns nil", appErr: &apiClient.ApplicationError{Type: apiClient.CONFLICT}, wantErr: nil},
		{name: "channel", appErr: &apiClient.ApplicationError{Code: &channelCode}, wantErr: apierror.ErrChannelAlreadyExists},
		{name: "wallet", appErr: &apiClient.ApplicationError{Code: &walletCode}, wantErr: apierror.ErrWalletAlreadyExists},
		{name: "watcher", appErr: &apiClient.ApplicationError{Code: &watcherCode}, wantErr: apierror.ErrWatcherAlreadyExists},
		{name: "idempotency", appErr: &apiClient.ApplicationError{Code: &idempotencyCode}, wantErr: apierror.ErrIdempotencyKeyMismatch},
		{name: "not finalizable", appErr: &apiClient.ApplicationError{Code: &notFinalizableCode}, wantErr: apierror.ErrOperationNotFinalizable},
		{name: "not cancellable", appErr: &apiClient.ApplicationError{Code: &notCancellableCode}, wantErr: apierror.ErrOperationNotCancellable},
		{name: "deadline elapsed", appErr: &apiClient.ApplicationError{Code: &deadlineCode}, wantErr: apierror.ErrOperationDeadlineElapsed},
		{name: "version conflict", appErr: &apiClient.ApplicationError{Code: &versionCode}, wantErr: apierror.ErrResourceVersionConflict},
		{name: "already archived", appErr: &apiClient.ApplicationError{Code: &archivedCode}, wantErr: apierror.ErrWalletAlreadyArchived},
		{name: "chain unavailable", appErr: &apiClient.ApplicationError{Code: &chainUnavailableCode}, wantErr: apierror.ErrChainUnavailable},
		{name: "unknown future code degrades to nil", appErr: &apiClient.ApplicationError{Code: &futureCode}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apierror.Conflict(tt.appErr)

			if tt.wantErr == nil {
				assert.NoError(t, got)
				return
			}
			assert.ErrorIs(t, got, tt.wantErr)
		})
	}
}

func TestApierror_WrapConflict(t *testing.T) {
	opErr := errors.New("failed to create channel")
	channelCode := apiClient.ApplicationErrorCodeChannelAlreadyExists
	futureCode := apiClient.ApplicationErrorCode("SOME_FUTURE_CONFLICT")
	detail := "name my-channel"

	t.Run("recognized code wraps opErr with typed sentinel and server message", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &channelCode, Message: "channel already exists"}
		err := apierror.WrapConflict(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.ErrorIs(t, err, apierror.ErrChannelAlreadyExists)
		assert.NotErrorIs(t, err, apierror.ErrWalletAlreadyExists)
		assert.Contains(t, err.Error(), "channel already exists")
		assert.NotContains(t, err.Error(), detail)
	})

	t.Run("missing code returns only opErr with server message", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Type: apiClient.CONFLICT, Message: "conflict"}
		err := apierror.WrapConflict(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrChannelAlreadyExists)
		assert.Contains(t, err.Error(), "conflict")
	})

	t.Run("unknown future code returns only opErr with detail", func(t *testing.T) {
		appErr := &apiClient.ApplicationError{Code: &futureCode}
		err := apierror.WrapConflict(appErr, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.NotErrorIs(t, err, apierror.ErrChannelAlreadyExists)
		assert.Contains(t, err.Error(), detail)
	})

	t.Run("nil application error uses detail without panicking", func(t *testing.T) {
		err := apierror.WrapConflict(nil, opErr, detail)

		assert.ErrorIs(t, err, opErr)
		assert.Contains(t, err.Error(), detail)
	})
}

func TestApierror_ConflictCode(t *testing.T) {
	channelCode := apiClient.ApplicationErrorCodeChannelAlreadyExists
	assert.Equal(t, "", apierror.ConflictCode(nil))
	assert.Equal(t, "CHANNEL_ALREADY_EXISTS", apierror.ConflictCode(&apiClient.ApplicationError{Code: &channelCode}))
}
