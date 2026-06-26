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
