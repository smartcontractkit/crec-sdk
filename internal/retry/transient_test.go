package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTransient(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, IsTransient(nil))
	})

	t.Run("ContextErrorsArePermanent", func(t *testing.T) {
		assert.False(t, IsTransient(context.Canceled))
		assert.False(t, IsTransient(context.DeadlineExceeded))
		assert.False(t, IsTransient(fmt.Errorf("wrapped: %w", context.Canceled)))
	})

	t.Run("TransientStatusCodes", func(t *testing.T) {
		cases := []string{
			"failed: status code 429",
			"failed: status code: 500",
			"failed: status code 503",
			"failed: status code 599",
		}
		for _, msg := range cases {
			assert.True(t, IsTransient(errors.New(msg)), msg)
		}
	})

	t.Run("PermanentStatusCodes", func(t *testing.T) {
		cases := []string{
			"failed: status code 400",
			"failed: status code: 401",
			"failed: status code 404",
			"failed: status code 409",
		}
		for _, msg := range cases {
			assert.False(t, IsTransient(errors.New(msg)), msg)
		}
	})

	t.Run("NetworkIndicators", func(t *testing.T) {
		cases := []string{
			"dial tcp: connection refused",
			"read tcp: connection reset by peer",
			"request timeout",
			"unexpected EOF",
			"write: broken pipe",
			"lookup api.example.com: no such host",
			"network is unreachable",
			"temporary failure in name resolution",
		}
		for _, msg := range cases {
			assert.True(t, IsTransient(errors.New(msg)), msg)
		}
	})

	t.Run("UnknownErrorsArePermanent", func(t *testing.T) {
		assert.False(t, IsTransient(errors.New("some unknown error")))
	})
}

func TestIsTransientStatusCode(t *testing.T) {
	transient := []int{429, 500, 502, 503, 504, 599}
	permanent := []int{200, 301, 400, 401, 403, 404, 409, 418, 422}

	for _, code := range transient {
		assert.True(t, IsTransientStatusCode(code), "status %d should be transient", code)
	}
	for _, code := range permanent {
		assert.False(t, IsTransientStatusCode(code), "status %d should be permanent", code)
	}
}
