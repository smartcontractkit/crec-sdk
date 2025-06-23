package verifier

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/internal/mockdata"
)

func TestVerifyEvent(t *testing.T) {
	event, err := mockdata.LoadMockEvent("valid_event.json")
	require.NoError(t, err)

	v, err := NewEventVerifier(
		&EventVerifierOptions{
			ValidSigners: []string{
				"0xFD528f7bd7a6eB8d6605BF944d122da7665C69A1",
				"0x4f4ddd274635D014C4584118a0fdD6cf89B25d3b",
			},
			MinRequiredSignatures: 2,
		},
	)
	require.NoError(t, err)

	verified, err := v.Verify(event)
	require.NoError(t, err)
	require.True(t, verified)
}
