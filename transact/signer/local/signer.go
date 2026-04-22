package local

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewSigner(privateKey *ecdsa.PrivateKey) *Signer {
	return &Signer{
		privateKey: privateKey,
	}
}

func (s *Signer) Sign(_ context.Context, hash []byte) ([]byte, error) {
	sig, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, err
	}
	if sig[64] <= 1 {
		sig[64] += 27 // Adjust the signature to be in the ethereum form
	}
	return sig, nil
}

// Destroy attempts to zeroize the private key material in memory.
// Note that Go's garbage collector and escape analysis make it hard to guarantee
// complete zeroization, but this clears the immediate reference.
func (s *Signer) Destroy() {
	if s.privateKey != nil && s.privateKey.D != nil {
		words := s.privateKey.D.Bits()
		for i := range words {
			words[i] = 0
		}
	}
	s.privateKey = nil
}
