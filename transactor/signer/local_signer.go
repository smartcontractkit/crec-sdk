package signer

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

type LocalSigner struct {
	privateKey *ecdsa.PrivateKey
}

func NewLocalSigner(privateKey *ecdsa.PrivateKey) *LocalSigner {
	return &LocalSigner{
		privateKey: privateKey,
	}
}

func (s *LocalSigner) Sign(hash []byte) ([]byte, error) {
	sig, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, err
	}
	if sig[64] <= 1 {
		sig[64] += 27 // Adjust the signature to be in the ethereum form
	}
	return sig, nil
}
