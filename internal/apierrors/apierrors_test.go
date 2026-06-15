package apierrors

import (
	"errors"
	"fmt"
	"testing"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Run("from ApplicationError", func(t *testing.T) {
		msg := Message(&apiClient.ApplicationError{Message: "channel with ID x not found"}, nil)
		assert.Equal(t, "channel with ID x not found", msg)
	})

	t.Run("from message field in body", func(t *testing.T) {
		body := []byte(`{"message":"wallet with address 0xabc not found","type":"NOT_FOUND","code":"WALLET_NOT_FOUND"}`)
		assert.Equal(t, "wallet with address 0xabc not found", Message(nil, body))
	})

	t.Run("from error field in body", func(t *testing.T) {
		body := []byte(`{"error":"resource not found: wallet with address 0xabc not found","type":"NOT_FOUND"}`)
		assert.Equal(t, "resource not found: wallet with address 0xabc not found", Message(nil, body))
	})

	t.Run("prefers ApplicationError over body", func(t *testing.T) {
		body := []byte(`{"message":"ignored"}`)
		msg := Message(&apiClient.ApplicationError{Message: "from struct"}, body)
		assert.Equal(t, "from struct", msg)
	})

	t.Run("ignores null body", func(t *testing.T) {
		assert.Equal(t, "", Message(nil, []byte("null")))
	})
}

func TestNotFoundKindFromResponse(t *testing.T) {
	walletCode := apiClient.ApplicationErrorCodeWalletNotFound
	channelCode := apiClient.ApplicationErrorCodeChannelNotFound

	tests := []struct {
		name   string
		apiErr *apiClient.ApplicationError
		body   []byte
		want   NotFoundKind
	}{
		{
			name:   "wallet code from ApplicationError",
			apiErr: &apiClient.ApplicationError{Code: &walletCode, Message: "wallet missing"},
			want:   NotFoundWallet,
		},
		{
			name: "channel code from body",
			body: []byte(`{"type":"NOT_FOUND","code":"CHANNEL_NOT_FOUND","message":"channel missing"}`),
			want: NotFoundChannel,
		},
		{
			name: "unknown when code absent",
			apiErr: &apiClient.ApplicationError{
				Message: "operation with ID abc not found in channel def",
			},
			want: NotFoundUnknown,
		},
		{
			name:   "code takes precedence over unrelated message",
			apiErr: &apiClient.ApplicationError{Code: &channelCode, Message: "wallet with address 0xabc not found"},
			want:   NotFoundChannel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NotFoundKindFromResponse(tt.apiErr, tt.body))
		})
	}
}

func TestMapNotFound(t *testing.T) {
	walletCode := apiClient.ApplicationErrorCodeWalletNotFound

	channelErr := errors.New("channel not found")
	walletErr := errors.New("wallet not found")
	fallbackErr := errors.New("fallback")

	t.Run("maps wallet code", func(t *testing.T) {
		err := MapNotFound(
			NotFoundResponse{
				JSON404: &apiClient.ApplicationError{Code: &walletCode, Message: "wallet missing"},
			},
			NotFoundMapping{
				Wallet: walletErr,
				Unknown: func(msg string) error {
					return fmt.Errorf("%w: %s", fallbackErr, msg)
				},
			},
		)
		require.ErrorIs(t, err, walletErr)
		require.Contains(t, err.Error(), "wallet missing")
	})

	t.Run("empty message uses Empty", func(t *testing.T) {
		err := MapNotFound(
			NotFoundResponse{},
			NotFoundMapping{Empty: func() error { return channelErr }},
		)
		require.ErrorIs(t, err, channelErr)
	})

	t.Run("unknown code uses Unknown", func(t *testing.T) {
		err := MapNotFound(
			NotFoundResponse{JSON404: &apiClient.ApplicationError{Message: "something"}},
			NotFoundMapping{
				Unknown: func(msg string) error { return fmt.Errorf("%w: %s", fallbackErr, msg) },
			},
		)
		require.ErrorIs(t, err, fallbackErr)
	})
}
