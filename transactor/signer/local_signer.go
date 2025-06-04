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
	return crypto.Sign(hash, s.privateKey)
}
