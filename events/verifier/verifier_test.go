package verifier

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/events/types"
)

func TestVerifyEvent(t *testing.T) {
	// Generate a new private key
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create a payload
	payload := []byte("test-payload")

	// Hash the payload
	hash := crypto.Keccak256Hash(payload)

	// Sign the hash
	sig, err := crypto.Sign(hash.Bytes(), privKey)
	require.NoError(t, err)

	// Encode signature as hex string
	sigHex := hex.EncodeToString(sig)

	// Prepare VerifiableEvent
	event := &types.VerifiableEvent{
		Type:       "TestEvent",
		Payload:    string(payload),
		Signatures: []string{sigHex},
	}

	signer := crypto.PubkeyToAddress(privKey.PublicKey)

	v := NewEventVerifier(
		&EventVerifierOptions{
			ValidSigners:          []string{signer.Hex()},
			MinRequiredSignatures: 1,
		},
	)

	verified, err := v.Verify(event)
	require.NoError(t, err)
	require.True(t, verified)
}
