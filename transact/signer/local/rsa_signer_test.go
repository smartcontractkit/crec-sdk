package local

import (
	"context"
	"crypto"
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

	modulus := signer.GetRSAModulus()
	expected := hex.EncodeToString(key.PublicKey.N.Bytes())
	require.Equal(t, expected, modulus)

	// Verify it's valid hex
	_, err = hex.DecodeString(modulus)
	require.NoError(t, err)
}

func TestRSASigner_GetRSAPublicExponent(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	exponent := signer.GetRSAPublicExponent()

	// Standard RSA exponent 65537 = 0x010001
	expected := hex.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())
	require.Equal(t, expected, exponent)
	require.Equal(t, "010001", exponent)
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
	// Go 1.26+ rejects keys < 1024 bits
	_, err := GenerateRSAKey(512)
	require.Error(t, err)
}

func TestRSASigner_PublicKey(t *testing.T) {
	key, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	signer, err := NewRSASigner(key)
	require.NoError(t, err)

	pub := signer.PublicKey()
	require.Equal(t, &key.PublicKey, pub)
}
