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

// minRSAKeyBits is the minimum RSA key size accepted by the SDK.
// Keys smaller than this are rejected as cryptographically inadequate.
const minRSAKeyBits = 2048

var _ signerPkg.Signer = &RSASigner{}

// RSASigner is an in-memory RSA PKCS#1 v1.5 signer for development and testing.
type RSASigner struct {
	privateKey *rsa.PrivateKey
}

// NewRSASigner creates an RSASigner from an existing *rsa.PrivateKey.
// Returns an error if the key is nil, structurally incomplete, too small
// (< minRSAKeyBits bits), or fails internal consistency checks.
func NewRSASigner(privateKey *rsa.PrivateKey) (*RSASigner, error) {
	if err := validateRSAKey(privateKey); err != nil {
		return nil, err
	}
	return &RSASigner{privateKey: privateKey}, nil
}

// GenerateRSAKey generates a new RSA private key of the given bit size.
// Typical values are 2048 or 4096. Returns an error if bits < minRSAKeyBits.
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error) {
	if bits < minRSAKeyBits {
		return nil, fmt.Errorf("RSA key size must be at least %d bits, got %d", minRSAKeyBits, bits)
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// validateRSAKey checks that key is non-nil, structurally complete, meets the
// minimum key-size requirement, and passes the crypto/rsa internal consistency check.
func validateRSAKey(key *rsa.PrivateKey) error {
	if key == nil {
		return fmt.Errorf("privateKey must not be nil")
	}
	if key.PublicKey.N == nil {
		return fmt.Errorf("privateKey has nil modulus (N)")
	}
	if key.D == nil {
		return fmt.Errorf("privateKey has nil private exponent (D)")
	}
	if key.PublicKey.E <= 0 {
		return fmt.Errorf("privateKey has invalid public exponent (E=%d)", key.PublicKey.E)
	}
	if key.PublicKey.N.BitLen() < minRSAKeyBits {
		return fmt.Errorf("RSA key size must be at least %d bits, got %d", minRSAKeyBits, key.PublicKey.N.BitLen())
	}
	if err := key.Validate(); err != nil {
		return fmt.Errorf("privateKey failed consistency check: %w", err)
	}
	return nil
}

// Sign signs a pre-hashed 32-byte message using PKCS#1 v1.5 with SHA-256.
// Signatures are deterministic.
func (s *RSASigner) Sign(_ context.Context, hash []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("signer has been destroyed")
	}
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	return rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash)
}

// PublicKey returns a copy of the RSA public key.
// A defensive copy is returned so that callers cannot mutate the signer's
// internal key state (e.g. by modifying the N big.Int in place).
func (s *RSASigner) PublicKey() *rsa.PublicKey {
	return &rsa.PublicKey{
		N: new(big.Int).Set(s.privateKey.PublicKey.N),
		E: s.privateKey.PublicKey.E,
	}
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

// Destroy performs a best-effort in-memory wipe of the private key material:
// it zeroes D, each prime, and the CRT precomputed values, then nils the
// private key reference so future Sign calls return an error.
//
// Note: Go's garbage collector and runtime do not guarantee that key material
// is fully erased from memory; copies may exist in other locations due to GC
// movement or past assignments. Destroy reduces exposure but is not a
// cryptographic-strength zeroisation.
func (s *RSASigner) Destroy() {
	if s.privateKey == nil {
		return
	}

	if s.privateKey.D != nil {
		s.privateKey.D.SetInt64(0)
	}
	for _, p := range s.privateKey.Primes {
		if p != nil {
			p.SetInt64(0)
		}
	}
	if s.privateKey.Precomputed.Dp != nil {
		s.privateKey.Precomputed.Dp.SetInt64(0)
	}
	if s.privateKey.Precomputed.Dq != nil {
		s.privateKey.Precomputed.Dq.SetInt64(0)
	}
	if s.privateKey.Precomputed.Qinv != nil {
		s.privateKey.Precomputed.Qinv.SetInt64(0)
	}

	s.privateKey = nil
}
