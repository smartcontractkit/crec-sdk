package local

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestNewRSASigner_NilKey(t *testing.T) {
	_, err := NewRSASigner(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "privateKey must not be nil")
}

func TestNewRSASigner_NilN(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	key.PublicKey.N = nil
	_, err = NewRSASigner(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil modulus")
}

func TestNewRSASigner_NilD(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	key.D = nil
	_, err = NewRSASigner(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil private exponent")
}

func TestNewRSASigner_InvalidE(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	key.PublicKey.E = 0
	_, err = NewRSASigner(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid public exponent")
}

func TestNewRSASigner_KeyTooSmall(t *testing.T) {
	// Bypass GenerateRSAKey's guard by using stdlib directly.
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	_, err = NewRSASigner(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least 2048 bits")
}

func TestRSASigner_Sign_Basic(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	// Use keccak256 hash (matches EIP-712 usage)
	hash := ethcrypto.Keccak256([]byte("test message"))

	sig, err := signer.Sign(context.Background(), hash)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	err = rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hash, sig)
	require.NoError(t, err)
}

func TestRSASigner_Sign_Deterministic(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	hash := ethcrypto.Keccak256([]byte("test message"))

	sig1, err := signer.Sign(context.Background(), hash)
	require.NoError(t, err)

	sig2, err := signer.Sign(context.Background(), hash)
	require.NoError(t, err)

	// Signatures must be identical (deterministic)
	require.Equal(t, sig1, sig2)

	err = rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hash, sig1)
	require.NoError(t, err)
}

func TestRSASigner_Sign_InvalidHashLength(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	_, err = signer.Sign(context.Background(), []byte("short"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "hash must be 32 bytes")
}

func TestRSASigner_GetRSAModulus(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	modulus, err := signer.GetRSAModulus()
	require.NoError(t, err)
	expected := hex.EncodeToString(key.PublicKey.N.Bytes())
	require.Equal(t, expected, modulus)

	// Verify it's valid hex
	_, err = hex.DecodeString(modulus)
	require.NoError(t, err)
}

func TestRSASigner_GetRSAModulus_AfterDestroy(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	signer, err := NewRSASigner(key)
	require.NoError(t, err)
	signer.Destroy()
	_, err = signer.GetRSAModulus()
	require.ErrorContains(t, err, "destroyed")
}

func TestRSASigner_GetRSAPublicExponent(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	exponent, err := signer.GetRSAPublicExponent()
	require.NoError(t, err)

	// Standard RSA exponent 65537 = 0x010001
	expected := hex.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())
	require.Equal(t, expected, exponent)
	require.Equal(t, "010001", exponent)
}

func TestRSASigner_GetRSAPublicExponent_AfterDestroy(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	signer, err := NewRSASigner(key)
	require.NoError(t, err)
	signer.Destroy()
	_, err = signer.GetRSAPublicExponent()
	require.ErrorContains(t, err, "destroyed")
}

func TestRSASigner_ImplementsSigner(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)
	require.NotNil(t, signer)
}

func TestGenerateRSAKey_ValidBits(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	require.Equal(t, 2048, key.N.BitLen())
}

func TestGenerateRSAKey_InvalidBits(t *testing.T) {
	// SDK enforces a minimum of 2048 bits regardless of runtime behaviour.
	_, err := GenerateRSAKey(1024)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least 2048 bits")
}

func TestRSASigner_PublicKey(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	pub, err := signer.PublicKey()
	require.NoError(t, err)

	// Value equality: N and E must match the original key.
	require.Equal(t, key.PublicKey.N, pub.N)
	require.Equal(t, key.PublicKey.E, pub.E)

	// Defensive copy: mutating the returned key must not affect the signer.
	pub.N.SetInt64(0)
	pub2, err := signer.PublicKey()
	require.NoError(t, err)
	require.NotEqual(t, int64(0), pub2.N.Int64(), "signer internal key was mutated through returned PublicKey")
}

func TestRSASigner_PublicKey_AfterDestroy(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	signer, err := NewRSASigner(key)
	require.NoError(t, err)
	signer.Destroy()
	_, err = signer.PublicKey()
	require.ErrorContains(t, err, "destroyed")
}

func TestRSASigner_RSAPublicKey(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	info, err := signer.RSAPublicKey()
	require.NoError(t, err)

	// E and N must be non-empty and match the individual helpers.
	expectedN, err := signer.GetRSAModulus()
	require.NoError(t, err)
	expectedE, err := signer.GetRSAPublicExponent()
	require.NoError(t, err)

	require.Equal(t, expectedN, info.N)
	require.Equal(t, expectedE, info.E)
	require.Equal(t, "010001", info.E) // standard exponent 65537
}

func TestRSASigner_RSAPublicKey_AfterDestroy(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	signer, err := NewRSASigner(key)
	require.NoError(t, err)
	signer.Destroy()
	_, err = signer.RSAPublicKey()
	require.ErrorContains(t, err, "destroyed")
}

func TestRSASigner_ImplementsRSAPublicKeyExporter(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)
	s, err := NewRSASigner(key)
	require.NoError(t, err)
	// The compile-time var _ signerPkg.RSAPublicKeyExporter = &RSASigner{} in
	// rsa_signer.go already enforces this; confirm the method is callable here.
	_, err = s.RSAPublicKey()
	require.NoError(t, err)
}

func TestRSASigner_Destroy(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	signer.Destroy()

	// Sign must fail after Destroy.
	hash := make([]byte, 32)
	_, err = signer.Sign(context.Background(), hash)
	require.Error(t, err)
}

func TestRSASigner_Destroy_ClearsKeyMaterial(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	// Hold references to the big.Int fields before handing the key to the signer
	// so we can observe that they were zeroed.
	d := key.D
	p0 := key.Primes[0]

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	signer.Destroy()

	require.Equal(t, int64(0), d.Int64(), "private exponent D was not zeroed")
	require.Equal(t, int64(0), p0.Int64(), "prime P was not zeroed")
}

func TestRSASigner_Destroy_OverwritesBackingWords(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	dWords := key.D.Bits()
	pWords := key.Primes[0].Bits()
	require.NotEmpty(t, dWords)
	require.NotEmpty(t, pWords)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	signer.Destroy()

	for _, w := range dWords {
		require.Equal(t, big.Word(0), w, "D backing words were not overwritten")
	}
	for _, w := range pWords {
		require.Equal(t, big.Word(0), w, "prime backing words were not overwritten")
	}
}

func TestRSASigner_Destroy_ClearsMultiPrimeCRTValues(t *testing.T) {
	key, err := rsa.GenerateMultiPrimeKey(rand.Reader, 3, 2048)
	require.NoError(t, err)
	key.Precompute()
	require.NotEmpty(t, key.Precomputed.CRTValues)

	coeffWords := key.Precomputed.CRTValues[0].Coeff.Bits()
	expWords := key.Precomputed.CRTValues[0].Exp.Bits()
	rWords := key.Precomputed.CRTValues[0].R.Bits()
	require.NotEmpty(t, coeffWords)
	require.NotEmpty(t, expWords)
	require.NotEmpty(t, rWords)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	signer.Destroy()

	for _, w := range coeffWords {
		require.Equal(t, big.Word(0), w, "CRT coeff backing words were not overwritten")
	}
	for _, w := range expWords {
		require.Equal(t, big.Word(0), w, "CRT exp backing words were not overwritten")
	}
	for _, w := range rWords {
		require.Equal(t, big.Word(0), w, "CRT R backing words were not overwritten")
	}
}

func TestRSASigner_Destroy_Idempotent(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	// Calling Destroy twice must not panic.
	signer.Destroy()
	signer.Destroy()
}
