package signer

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vault"
)

func TestTransitSigner_Sign_Integration(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

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
	err = client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})
	require.NoError(t, err)

	// Create RSA key for signing
	keyName := "test-rsa-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
		"type": "rsa-2048",
	})
	require.NoError(t, err)

	// Create our TransitSigner
	signer, err := NewTransitSigner(
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
	signature, err := signer.Sign(hash[:])
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
	signature2, err := signer.Sign(hash[:])
	require.NoError(t, err)
	require.NotEmpty(t, signature2)

	// Verify the second signature as well
	err = rsa.VerifyPSS(rsaPubKey, crypto.SHA256, hash[:], signature2, nil)
	require.NoError(t, err, "Second signature should also be valid")

	t.Logf("First signature length: %d", len(signature))
	t.Logf("Second signature length: %d", len(signature2))
}

func TestTransitSigner_Sign_WithECDSA(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

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
	err = client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})
	require.NoError(t, err)

	// Create ECDSA key for signing (more similar to Ethereum usage, but still not secp256k1)
	keyName := "test-ecdsa-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
		"type": "ecdsa-p256",
	})
	require.NoError(t, err)

	// Create our TransitSigner
	signer, err := NewTransitSigner(
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
	signature, err := signer.Sign(hash[:])
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

func TestTransitSigner_Public(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

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
	err = client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})
	require.NoError(t, err)

	// Test with RSA key
	rsaKeyName := "test-rsa-public-key"
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", rsaKeyName), map[string]interface{}{
		"type": "rsa-2048",
	})
	require.NoError(t, err)

	// Create signer for RSA key
	rsaSigner, err := NewTransitSigner(vaultURL, "myroot", "transit", rsaKeyName)
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
	_, err = client.Logical().Write(fmt.Sprintf("transit/keys/%s", ecdsaKeyName), map[string]interface{}{
		"type": "ecdsa-p256",
	})
	require.NoError(t, err)

	// Create signer for ECDSA key
	ecdsaSigner, err := NewTransitSigner(vaultURL, "myroot", "transit", ecdsaKeyName)
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

	t.Logf("RSA public key size: %d bits", rsaPubKey.Size()*8)
	t.Logf("ECDSA curve: %s", ecdsaPubKey.Curve.Params().Name)
}

func TestTransitSigner_InvalidKey(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

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
	err = client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})
	require.NoError(t, err)

	// Create our TransitSigner with a non-existent key
	signer, err := NewTransitSigner(
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
	signature, err := signer.Sign(hash[:])
	require.Error(t, err)
	require.Empty(t, signature)
	require.Contains(t, err.Error(), "vault sign failed")
}

func TestTransitSigner_InvalidToken(t *testing.T) {
	ctx := context.Background()

	// Start Vault container
	vaultContainer, err := vault.Run(ctx,
		"hashicorp/vault:1.13.3",
		vault.WithToken("myroot"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	// Get container connection info
	vaultURL, err := vaultContainer.HttpHostAddress(ctx)
	require.NoError(t, err)

	// Wait for Vault to be ready
	time.Sleep(2 * time.Second)

	// Create our TransitSigner with invalid token
	signer, err := NewTransitSigner(
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
	signature, err := signer.Sign(hash[:])
	require.Error(t, err)
	require.Empty(t, signature)
	require.Contains(t, err.Error(), "vault sign failed")
}
