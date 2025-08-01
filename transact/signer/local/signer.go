package local

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
}

func NewSigner(privateKey *ecdsa.PrivateKey) *Signer {
	return &Signer{
		privateKey: privateKey,
	}
}

func (s *Signer) Sign(hash []byte) ([]byte, error) {
	sig, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, err
	}
	if sig[64] <= 1 {
		sig[64] += 27 // Adjust the signature to be in the ethereum form
	}
	return sig, nil
}
