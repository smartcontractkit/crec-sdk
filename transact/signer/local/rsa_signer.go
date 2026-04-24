package local

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"math/big"

	signerPkg "github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signerPkg.Signer = &RSASigner{}

// RSASigner is an in-memory RSA PKCS#1 v1.5 signer for development and testing.
type RSASigner struct {
	privateKey *rsa.PrivateKey
}

// NewRSASigner creates an RSASigner from an existing *rsa.PrivateKey.
// Returns an error if privateKey is nil.
func NewRSASigner(privateKey *rsa.PrivateKey) (*RSASigner, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("privateKey must not be nil")
	}
	return &RSASigner{privateKey: privateKey}, nil
}

// GenerateRSAKey generates a new RSA private key of the given bit size.
// Typical values are 2048 or 4096. Returns an error if bits < 1024.
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// Sign signs a pre-hashed 32-byte message using PKCS#1 v1.5 with SHA-256.
// Signatures are deterministic.
func (s *RSASigner) Sign(_ context.Context, hash []byte) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	return rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash)
}

// PublicKey returns the RSA public key.
func (s *RSASigner) PublicKey() *rsa.PublicKey {
	return &s.privateKey.PublicKey
}

// GetRSAModulus returns the hex-encoded modulus N of the public key.
func (s *RSASigner) GetRSAModulus() string {
	return hex.EncodeToString(s.privateKey.PublicKey.N.Bytes())
}

// GetRSAPublicExponent returns the hex-encoded public exponent E.
// For the standard exponent 65537, returns "010001".
func (s *RSASigner) GetRSAPublicExponent() string {
	return hex.EncodeToString(big.NewInt(int64(s.privateKey.PublicKey.E)).Bytes())
}
