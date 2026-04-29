package vault

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func TestSigner_Sign_Integration(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create RSA key for signing
	keyName := "test-rsa-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
			"type": "rsa-2048",
		},
	)
	require.NoError(t, err)

	// Create our TransitSigner
	signer, err := NewSigner(
		vaultURL,
		"myroot",
		"transit",
		keyName,
	)
	require.NoError(t, err)

	// Test data to sign
	testData := []byte("hello world")
	hash := sha256.Sum256(testData)

	// Test signing
	signature, err := signer.Sign(context.Background(), hash[:])
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// Verify the signature was created
	require.Greater(t, len(signature), 100, "RSA signature should be reasonably long")

	// Get the public key from Vault
	pubKeyInterface, err := signer.Public()
	require.NoError(t, err)
	require.NotNil(t, pubKeyInterface)

	// Verify it's an RSA public key
	rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	require.True(t, ok, "Public key should be an RSA key")
	require.NotNil(t, rsaPubKey)

	// Verify the signature using the public key
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, hash[:], signature, nil)
	require.NoError(t, err, "Signature should be valid")

	// Test that we can sign the same data multiple times and get different signatures
	// (RSA with PKCS#1 v1.5 padding should produce deterministic signatures, but JWS might add randomness)
	signature2, err := signer.Sign(context.Background(), hash[:])
	require.NoError(t, err)
	require.NotEmpty(t, signature2)

	// Verify the second signature as well
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, hash[:], signature2, nil)
	require.NoError(t, err, "Second signature should also be valid")

	t.Logf("First signature length: %d", len(signature))
	t.Logf("Second signature length: %d", len(signature2))
}

func TestSigner_Sign_WithECDSA(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create ECDSA key for signing (more similar to Ethereum usage, but still not secp256k1)
	keyName := "test-ecdsa-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
			"type": "ecdsa-p256",
		},
	)
	require.NoError(t, err)

	// Create our Signer
	signer, err := NewSigner(
		vaultURL,
		"myroot",
		"transit",
		keyName,
	)
	require.NoError(t, err)

	// Test data to sign (similar to what would be used in Ethereum)
	testData := []byte("hello world")
	hash := sha256.Sum256(testData)

	// Test signing
	signature, err := signer.Sign(context.Background(), hash[:])
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// ECDSA signatures should be around 64-70 bytes depending on encoding
	require.Greater(t, len(signature), 50, "ECDSA signature should be reasonably sized")
	require.Less(t, len(signature), 200, "ECDSA signature shouldn't be too large")

	// Get the public key from Vault
	pubKeyInterface, err := signer.Public()
	require.NoError(t, err)
	require.NotNil(t, pubKeyInterface)

	// Verify it's an ECDSA public key
	ecdsaPubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	require.True(t, ok, "Public key should be an ECDSA key")
	require.NotNil(t, ecdsaPubKey)

	// Verify the signature using the public key
	valid := ecdsa.VerifyASN1(ecdsaPubKey, hash[:], signature)
	require.True(t, valid, "ECDSA signature should be valid")

	t.Logf("ECDSA signature length: %d", len(signature))
}

func TestSigner_Public(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Test with RSA key
	rsaKeyName := "test-rsa-public-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", rsaKeyName), map[string]interface{}{
			"type": "rsa-2048",
		},
	)
	require.NoError(t, err)

	// Create signer for RSA key
	rsaSigner, err := NewSigner(vaultURL, "myroot", "transit", rsaKeyName)
	require.NoError(t, err)

	// Test getting RSA public key
	rsaPubKeyInterface, err := rsaSigner.Public()
	require.NoError(t, err)
	require.NotNil(t, rsaPubKeyInterface)

	// Verify it's an RSA public key
	rsaPubKey, ok := rsaPubKeyInterface.(*rsa.PublicKey)
	require.True(t, ok, "Should return RSA public key")
	require.NotNil(t, rsaPubKey)
	require.Equal(t, 2048, rsaPubKey.Size()*8, "Should be RSA-2048 key")

	// Test with ECDSA key
	ecdsaKeyName := "test-ecdsa-public-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", ecdsaKeyName), map[string]interface{}{
			"type": "ecdsa-p256",
		},
	)
	require.NoError(t, err)

	// Create signer for ECDSA key
	ecdsaSigner, err := NewSigner(vaultURL, "myroot", "transit", ecdsaKeyName)
	require.NoError(t, err)

	// Test getting ECDSA public key
	ecdsaPubKeyInterface, err := ecdsaSigner.Public()
	require.NoError(t, err)
	require.NotNil(t, ecdsaPubKeyInterface)

	// Verify it's an ECDSA public key
	ecdsaPubKey, ok := ecdsaPubKeyInterface.(*ecdsa.PublicKey)
	require.True(t, ok, "Should return ECDSA public key")
	require.NotNil(t, ecdsaPubKey)
	require.Equal(t, "P-256", ecdsaPubKey.Curve.Params().Name, "Should be P-256 curve")

	// Test key rotation to ensure we get the latest key version
	// Rotate the ECDSA key
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s/rotate", ecdsaKeyName), map[string]interface{}{},
	)
	require.NoError(t, err)

	// Fetch the public key again. It should fetch the newly rotated key (version 2)
	rotatedEcdsaPubKeyInterface, err := ecdsaSigner.Public()
	require.NoError(t, err)
	require.NotNil(t, rotatedEcdsaPubKeyInterface)

	rotatedEcdsaPubKey, ok := rotatedEcdsaPubKeyInterface.(*ecdsa.PublicKey)
	require.True(t, ok, "Should return ECDSA public key")
	require.NotNil(t, rotatedEcdsaPubKey)

	// Ensure the public key has changed
	require.False(t, ecdsaPubKey.Equal(rotatedEcdsaPubKey), "Public key should change after rotation")

	t.Logf("RSA public key size: %d bits", rsaPubKey.Size()*8)
	t.Logf("ECDSA curve: %s", ecdsaPubKey.Curve.Params().Name)
}

func TestSigner_CreateKey(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create a signer (we need this to create keys)
	signer, err := NewSigner(vaultURL, "myroot", "transit", "dummy")
	require.NoError(t, err)

	// Test creating RSA-2048 key
	rsaKeyName := "test-created-rsa-key"
	rsaResult, err := signer.CreateKey(rsaKeyName, KeyTypeRSA2048)
	require.NoError(t, err)
	require.NotNil(t, rsaResult)
	require.Equal(t, rsaKeyName, rsaResult.KeyName)
	require.Equal(t, KeyTypeRSA2048, rsaResult.KeyType)
	require.NotNil(t, rsaResult.PublicKey)
	require.NotEmpty(t, rsaResult.Modulus)

	// Verify it's an RSA public key
	rsaPubKey, ok := rsaResult.PublicKey.(*rsa.PublicKey)
	require.True(t, ok, "Should be RSA public key")
	require.Equal(t, 2048, rsaPubKey.Size()*8, "Should be RSA-2048")

	// Verify the modulus matches
	expectedModulus := hex.EncodeToString(rsaPubKey.N.Bytes())
	require.Equal(t, expectedModulus, rsaResult.Modulus)

	t.Logf("Created RSA key: %s", rsaKeyName)
	t.Logf("RSA modulus: %s", rsaResult.Modulus[:32]+"...") // Show first 32 chars

	// Test creating ECDSA P-256 key
	ecdsaKeyName := "test-created-ecdsa-key"
	ecdsaResult, err := signer.CreateKey(ecdsaKeyName, KeyTypeECDSAP256)
	require.NoError(t, err)
	require.NotNil(t, ecdsaResult)
	require.Equal(t, ecdsaKeyName, ecdsaResult.KeyName)
	require.Equal(t, KeyTypeECDSAP256, ecdsaResult.KeyType)
	require.NotNil(t, ecdsaResult.PublicKey)
	require.Empty(t, ecdsaResult.Modulus) // ECDSA keys don't have modulus

	// Verify it's an ECDSA public key
	ecdsaPubKey, ok := ecdsaResult.PublicKey.(*ecdsa.PublicKey)
	require.True(t, ok, "Should be ECDSA public key")
	require.Equal(t, "P-256", ecdsaPubKey.Curve.Params().Name)

	t.Logf("Created ECDSA key: %s", ecdsaKeyName)

	// Test creating RSA-4096 key
	rsa4096KeyName := "test-created-rsa4096-key"
	rsa4096Result, err := signer.CreateKey(rsa4096KeyName, KeyTypeRSA4096)
	require.NoError(t, err)
	require.NotNil(t, rsa4096Result)
	require.Equal(t, KeyTypeRSA4096, rsa4096Result.KeyType)
	require.NotEmpty(t, rsa4096Result.Modulus)

	// Verify it's RSA-4096
	rsa4096PubKey, ok := rsa4096Result.PublicKey.(*rsa.PublicKey)
	require.True(t, ok, "Should be RSA public key")
	require.Equal(t, 4096, rsa4096PubKey.Size()*8, "Should be RSA-4096")

	t.Logf("Created RSA-4096 key: %s", rsa4096KeyName)
	t.Logf("RSA-4096 modulus length: %d chars", len(rsa4096Result.Modulus))
}

func TestSigner_GetRSAModulus(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create RSA key manually
	rsaKeyName := "test-modulus-rsa-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", rsaKeyName), map[string]interface{}{
			"type": "rsa-2048",
		},
	)
	require.NoError(t, err)

	// Create signer for the RSA key
	rsaSigner, err := NewSigner(vaultURL, "myroot", "transit", rsaKeyName)
	require.NoError(t, err)

	// Test getting RSA modulus
	modulus, err := rsaSigner.GetRSAModulus()
	require.NoError(t, err)
	require.NotEmpty(t, modulus)

	// Verify it's a valid hex string
	_, err = hex.DecodeString(modulus)
	require.NoError(t, err, "Modulus should be valid hex")

	// Verify the modulus matches what we get from Public()
	pubKey, err := rsaSigner.Public()
	require.NoError(t, err)
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	require.True(t, ok)
	require.NotNil(t, rsaPubKey.N)
	expectedModulus := hex.EncodeToString(rsaPubKey.N.Bytes())
	require.Equal(t, expectedModulus, modulus)

	t.Logf("RSA modulus: %s", modulus[:32]+"...")

	// Test with ECDSA key (should fail)
	ecdsaKeyName := "test-modulus-ecdsa-key"
	_, err = client.Logical().Write(
		fmt.Sprintf("transit/keys/%s", ecdsaKeyName), map[string]interface{}{
			"type": "ecdsa-p256",
		},
	)
	require.NoError(t, err)

	ecdsaSigner, err := NewSigner(vaultURL, "myroot", "transit", ecdsaKeyName)
	require.NoError(t, err)

	// Should fail to get modulus from ECDSA key
	_, err = ecdsaSigner.GetRSAModulus()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotValidRSAKey)
}

func TestSigner_GetRSAPublicExponent(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.3", vault.WithToken("myroot"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	err = client.Sys().Mount("transit", &api.MountInput{Type: "transit"})
	require.NoError(t, err)

	rsaKeyName := "test-exponent-rsa-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", rsaKeyName), map[string]interface{}{"type": "rsa-2048"})
	require.NoError(t, err)

	rsaSigner, err := NewSigner(vaultURL, "myroot", "transit", rsaKeyName)
	require.NoError(t, err)

	exponent, err := rsaSigner.GetRSAPublicExponent()
	require.NoError(t, err)
	require.NotEmpty(t, exponent)

	// Verify it decodes as valid hex and matches what Public() gives us.
	_, err = hex.DecodeString(exponent)
	require.NoError(t, err, "exponent should be valid hex")

	pubKey, err := rsaSigner.Public()
	require.NoError(t, err)
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	require.True(t, ok)
	expectedExp := hex.EncodeToString(big.NewInt(int64(rsaPubKey.E)).Bytes())
	require.Equal(t, expectedExp, exponent)
	t.Logf("RSA public exponent: %s", exponent)

	// Should fail for an ECDSA key.
	ecdsaKeyName := "test-exponent-ecdsa-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", ecdsaKeyName), map[string]interface{}{"type": "ecdsa-p256"})
	require.NoError(t, err)

	ecdsaSigner, err := NewSigner(vaultURL, "myroot", "transit", ecdsaKeyName)
	require.NoError(t, err)

	_, err = ecdsaSigner.GetRSAPublicExponent()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotValidRSAKey)
}

func TestSigner_RSAPublicKey(t *testing.T) {
	ctx := context.Background()

	vaultContainer, err := vault.Run(ctx, "hashicorp/vault:1.13.3", vault.WithToken("myroot"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)
	time.Sleep(2 * time.Second)

	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	err = client.Sys().Mount("transit", &api.MountInput{Type: "transit"})
	require.NoError(t, err)

	rsaKeyName := "test-rsapublickey-rsa-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", rsaKeyName), map[string]interface{}{"type": "rsa-2048"})
	require.NoError(t, err)

	rsaSigner, err := NewSigner(vaultURL, "myroot", "transit", rsaKeyName)
	require.NoError(t, err)

	info, err := rsaSigner.RSAPublicKey()
	require.NoError(t, err)
	require.NotEmpty(t, info.E)
	require.NotEmpty(t, info.N)

	// Values must match the individual helpers.
	expectedN, err := rsaSigner.GetRSAModulus()
	require.NoError(t, err)
	expectedE, err := rsaSigner.GetRSAPublicExponent()
	require.NoError(t, err)
	require.Equal(t, expectedN, info.N)
	require.Equal(t, expectedE, info.E)
	t.Logf("RSAPublicKey info: E=%s N=%s...", info.E, info.N[:16])
}

func TestCreateKeyInVault(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Test the convenience function
	keyName := "test-convenience-key"
	result, err := CreateKeyInVault(vaultURL, "myroot", "transit", keyName, KeyTypeRSA2048)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, keyName, result.KeyName)
	require.Equal(t, KeyTypeRSA2048, result.KeyType)
	require.NotEmpty(t, result.Modulus)

	// Verify the key was actually created in Vault
	resp, err := client.Logical().Read(fmt.Sprintf("transit/keys/%s", keyName))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "rsa-2048", resp.Data["type"])

	t.Logf("Convenience function created key: %s", keyName)
}

func TestSigner_CreateKey_ErrorCases(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create signer
	signer, err := NewSigner(vaultURL, "myroot", "transit", "dummy")
	require.NoError(t, err)

	// Test creating key with duplicate name
	keyName := "duplicate-key"
	result1, err := signer.CreateKey(keyName, KeyTypeRSA2048)
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Try to create key with same name (should still work in Vault, but key will be overwritten)
	result2, err := signer.CreateKey(keyName, KeyTypeECDSAP256)
	require.NoError(t, err)
	require.NotNil(t, result2)
	require.Equal(t, KeyTypeECDSAP256, result2.KeyType) // Should be the new type

	// Test with invalid Vault token
	invalidSigner, err := NewSigner(vaultURL, "invalid-token", "transit", "dummy")
	require.NoError(t, err)

	_, err = invalidSigner.CreateKey("test-invalid", KeyTypeRSA2048)
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "vault")
}

func TestSigner_InvalidKey(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create Vault client for setup
	client, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	client.SetAddress(vaultURL)
	client.SetToken("myroot")

	// Enable transit secrets engine
	err = client.Sys().Mount(
		"transit", &api.MountInput{
			Type: "transit",
		},
	)
	require.NoError(t, err)

	// Create our TransitSigner with a non-existent key
	signer, err := NewSigner(
		vaultURL,
		"myroot",
		"transit",
		"non-existent-key",
	)
	require.NoError(t, err)

	// Test data to sign
	testData := []byte("hello world")
	hash := sha256.Sum256(testData)

	// Test signing should fail
	signature, err := signer.Sign(context.Background(), hash[:])
	require.Error(t, err)
	require.Empty(t, signature)
	require.ErrorIs(t, err, ErrVaultSignFailed)
}

func TestSigner_InvalidToken(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(
		ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		},
	)

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create our TransitSigner with invalid token
	signer, err := NewSigner(
		vaultURL,
		"invalid-token",
		"transit",
		"any-key",
	)
	require.NoError(t, err)

	// Test data to sign
	testData := []byte("hello world")
	hash := sha256.Sum256(testData)

	// Test signing should fail due to auth
	signature, err := signer.Sign(context.Background(), hash[:])
	require.Error(t, err)
	require.Empty(t, signature)
	require.ErrorIs(t, err, ErrVaultSignFailed)
}

func TestSigner_Public_TrailingGarbage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Valid PEM but with trailing garbage
		pubKeyPEM := "-----BEGIN PUBLIC KEY-----\n" +
			"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEP/T/B1R5A6l3rI7G9zYk3gR4k8Y4\n" +
			"qjZJ6uQxL7b+E8M2j5X6M9k1u7Y0vR4y3M1qL6p9x4b8X2A7z3E5n4a9gA==\n" +
			"-----END PUBLIC KEY-----\n" +
			"trailing garbage bytes"

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"type": "ecdsa-p256",
				"keys": map[string]interface{}{
					"1": map[string]interface{}{
						"public_key": pubKeyPEM,
					},
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	signer, err := NewSigner(server.URL, "dummy-token", "transit", "test-key")
	require.NoError(t, err)

	_, err = signer.Public()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTrailingGarbageAfterPEM)
}
