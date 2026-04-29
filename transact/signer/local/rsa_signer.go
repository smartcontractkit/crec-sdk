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
var _ signerPkg.RSAPublicKeyExporter = &RSASigner{}

// RSASigner is an in-memory RSA signer for development and testing.
// It produces PKCS#1 v1.5 signatures — deterministic and compatible with
// the CREC certification test infrastructure.
//
// Note: the Vault RSA signer uses RSA-PSS, which is a different padding scheme.
// Signatures from RSASigner and the Vault signer are NOT interchangeable.
// Use RSASigner when your verifier expects PKCS#1 v1.5; use the Vault signer
// when your verifier expects PSS.
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

// Sign signs a pre-hashed 32-byte message using PKCS#1 v1.5 with SHA-256 as
// the hash identifier. Signatures are deterministic (no randomness in PKCS#1 v1.5).
// The input hash is treated as opaque pre-hashed bytes; keccak256 digests from
// EIP-712 operations are the typical input.
func (s *RSASigner) Sign(_ context.Context, hash []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("signer has been destroyed")
	}
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	return rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash)
}

// PublicKey returns a defensive copy of the RSA public key so that callers
// cannot mutate the signer's internal key state.
// Returns an error if the signer has been destroyed.
func (s *RSASigner) PublicKey() (*rsa.PublicKey, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("signer has been destroyed")
	}
	return &rsa.PublicKey{
		N: new(big.Int).Set(s.privateKey.PublicKey.N),
		E: s.privateKey.PublicKey.E,
	}, nil
}

// GetRSAModulus returns the hex-encoded modulus N of the public key.
// Returns an error if the signer has been destroyed.
func (s *RSASigner) GetRSAModulus() (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("signer has been destroyed")
	}
	return hex.EncodeToString(s.privateKey.PublicKey.N.Bytes()), nil
}

// GetRSAPublicExponent returns the hex-encoded public exponent E.
// For the standard exponent 65537, returns "010001".
// Returns an error if the signer has been destroyed.
func (s *RSASigner) GetRSAPublicExponent() (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("signer has been destroyed")
	}
	return hex.EncodeToString(big.NewInt(int64(s.privateKey.PublicKey.E)).Bytes()), nil
}

// RSAPublicKey returns the public components of this signer's RSA key in the
// hex-encoded string format used by the CREC platform.
// Implements signer.RSAPublicKeyExporter.
func (s *RSASigner) RSAPublicKey() (signerPkg.RSAPublicKeyInfo, error) {
	n, err := s.GetRSAModulus()
	if err != nil {
		return signerPkg.RSAPublicKeyInfo{}, err
	}
	e, err := s.GetRSAPublicExponent()
	if err != nil {
		return signerPkg.RSAPublicKeyInfo{}, err
	}
	return signerPkg.RSAPublicKeyInfo{E: e, N: n}, nil
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

	zeroBigInt(s.privateKey.D)
	s.privateKey.D = nil

	for i, p := range s.privateKey.Primes {
		zeroBigInt(p)
		s.privateKey.Primes[i] = nil
	}

	// CRTValues is deprecated by crypto/rsa for optimization use, but it may
	// still be populated for backward compatibility; wipe it if present.
	for i := range s.privateKey.Precomputed.CRTValues {
		zeroBigInt(s.privateKey.Precomputed.CRTValues[i].Coeff)
		zeroBigInt(s.privateKey.Precomputed.CRTValues[i].Exp)
		zeroBigInt(s.privateKey.Precomputed.CRTValues[i].R)
		s.privateKey.Precomputed.CRTValues[i].Coeff = nil
		s.privateKey.Precomputed.CRTValues[i].Exp = nil
		s.privateKey.Precomputed.CRTValues[i].R = nil
	}
	s.privateKey.Precomputed.CRTValues = nil

	zeroBigInt(s.privateKey.Precomputed.Dp)
	s.privateKey.Precomputed.Dp = nil
	zeroBigInt(s.privateKey.Precomputed.Dq)
	s.privateKey.Precomputed.Dq = nil
	zeroBigInt(s.privateKey.Precomputed.Qinv)
	s.privateKey.Precomputed.Qinv = nil

	zeroBigInt(s.privateKey.PublicKey.N)
	s.privateKey.PublicKey.N = nil

	s.privateKey = nil
}

// zeroBigInt overwrites the backing words of x before resetting it.
// This is best-effort: Go's GC and escape analysis may leave copies elsewhere.
func zeroBigInt(x *big.Int) {
	if x == nil {
		return
	}
	words := x.Bits()
	for i := range words {
		words[i] = 0
	}
	x.SetInt64(0)
}
