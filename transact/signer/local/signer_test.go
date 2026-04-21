package local

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestLocalSigner_Sign_EthereumVAdjusted(t *testing.T) {
	// Deterministic dev key (publicly known)
	priv, err := crypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	s := NewSigner(priv)

	// 32-byte hash (as Ethereum expects)
	hash := crypto.Keccak256([]byte("hello world"))
	sig, err := s.Sign(context.Background(), hash)
	require.NoError(t, err)
	require.Len(t, sig, 65)

	// Expect v in {27,28}
	require.True(t, sig[64] == 27 || sig[64] == 28, "unexpected v byte: %d", sig[64])

	// Cross-check against crypto.Sign (which returns v in {0,1})
	raw, err := crypto.Sign(hash, priv)
	require.NoError(t, err)
	if raw[64] <= 1 {
		raw[64] += 27
	}
	require.Equal(t, raw, sig)
}

func TestLocalSigner_Destroy(t *testing.T) {
	priv, err := crypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	s := NewSigner(priv)

	// Call Destroy to zero out memory
	s.Destroy()

	hash := crypto.Keccak256([]byte("hello world"))
	
	// Ensure that signing panics or fails after destruction
	// Depending on implementation, D.Bytes might panic if D is nil or zeroized.
	defer func() {
		if r := recover(); r != nil {
			// Panic caught, meaning the key was properly zeroized/nil'd.
			t.Logf("Panic caught after Destroy: %v", r)
		}
	}()
	
	_, err = s.Sign(context.Background(), hash)
	if err == nil {
		t.Error("expected error or panic when signing with destroyed key")
	}
}
