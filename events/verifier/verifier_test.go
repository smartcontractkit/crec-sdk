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
				"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			},
			MinRequiredSignatures: 1,
		},
	)
	require.NoError(t, err)

	verified, err := v.Verify(event)
	require.NoError(t, err)
	require.True(t, verified)
}
