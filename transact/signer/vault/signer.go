package vault

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

// KeyType represents the type of cryptographic key to create
type KeyType string

const (
	KeyTypeRSA2048   KeyType = "rsa-2048"
	KeyTypeRSA4096   KeyType = "rsa-4096"
	KeyTypeECDSAP256 KeyType = "ecdsa-p256"
	KeyTypeECDSAP384 KeyType = "ecdsa-p384"
	KeyTypeECDSAP521 KeyType = "ecdsa-p521"
)

type Signer struct {
	client  *vault.Client
	keyName string
	mount   string
}

// Option is a functional option for configuring the Signer
type Option func(*Signer)

// WithClient sets a custom Vault client (useful for testing)
func WithClient(client *vault.Client) Option {
	return func(s *Signer) {
		s.client = client
	}
}

func NewSigner(vaultUrl, token, mountPath, key string, opts ...Option) (*Signer, error) {
	if vaultUrl == "" || token == "" || mountPath == "" || key == "" {
		return nil, fmt.Errorf("vaultUrl, token, mountPath, and key must be set")
	}

	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, err
	}

	client.SetAddress(vaultUrl)
	client.SetToken(token)

	s := &Signer{
		client:  client,
		keyName: key,
		mount:   mountPath, // usually "transit"
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	// base64 encore the payload to sign
	b64 := base64.StdEncoding.EncodeToString(hash)

	// call vault client to sign payload
	resp, err := s.client.Logical().WriteWithContext(
		ctx,
		fmt.Sprintf("%s/sign/%s", s.mount, s.keyName), map[string]any{
			"input":                b64,
			"prehashed":            true,
			"marshaling_algorithm": "asn1",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("vault sign failed: %w", err)
	}

	// pull the signature from the response
	sig, ok := resp.Data["signature"].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected signature format: %+v", resp.Data)
	}

	// split the response into its three parts to strip the "vault:v1:" prefix
	parts := strings.SplitN(sig, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid vault signature format: %s", sig)
	}

	// return the last part as bytes
	decoded, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w (signature: %s)", err, parts[2])
	}
	return decoded, nil
}

// Public retrieves the public key from Vault for this signing key
func (s *Signer) Public() (interface{}, error) {
	// Get key information from Vault
	resp, err := s.client.Logical().Read(fmt.Sprintf("%s/keys/%s", s.mount, s.keyName))
	if err != nil {
		return nil, fmt.Errorf("failed to read key from vault: %w", err)
	}

	if resp == nil || resp.Data == nil {
		return nil, fmt.Errorf("key not found in vault: %s", s.keyName)
	}

	// Get the keys map which contains version information
	keys, ok := resp.Data["keys"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys format in vault response")
	}

	// Get the latest version (version "1" for new keys)
	var publicKeyPEM string
	for _, keyInfo := range keys {
		if keyInfoMap, ok := keyInfo.(map[string]interface{}); ok {
			if pubKey, exists := keyInfoMap["public_key"]; exists {
				if pubKeyStr, ok := pubKey.(string); ok {
					publicKeyPEM = pubKeyStr
					break
				}
			}
		}
	}

	if publicKeyPEM == "" {
		return nil, fmt.Errorf("public key not found in vault response")
	}

	// Parse the PEM-encoded public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse the public key based on the key type
	keyType, ok := resp.Data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("key type not found in vault response")
	}

	switch {
	case strings.HasPrefix(keyType, "rsa"):
		// Parse RSA public key
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("key is not an RSA public key")
		}
		return rsaPubKey, nil

	case strings.HasPrefix(keyType, "ecdsa"):
		// Parse ECDSA public key
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
		}
		ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("key is not an ECDSA public key")
		}
		return ecdsaPubKey, nil

	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GetRSAModulus returns the hex-encoded modulus of the RSA public key
func (s *Signer) GetRSAModulus() (string, error) {
	pubKey, err := s.Public()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("key is not an RSA key, got %T", pubKey)
	}

	return hex.EncodeToString(rsaPubKey.N.Bytes()), nil
}

// KeyCreationResult contains information about a newly created key
type KeyCreationResult struct {
	KeyName   string
	KeyType   KeyType
	PublicKey interface{}
	Modulus   string // For RSA keys only, hex-encoded modulus
}

// CreateKey creates a new cryptographic key in Vault Transit secrets engine
func (s *Signer) CreateKey(keyName string, keyType KeyType) (*KeyCreationResult, error) {
	// Create the key in Vault
	_, err := s.client.Logical().Write(
		fmt.Sprintf("%s/keys/%s", s.mount, keyName), map[string]interface{}{
			"type": string(keyType),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create key in vault: %w", err)
	}

	// Create a temporary signer to get the public key
	tempSigner := &Signer{
		client:  s.client,
		keyName: keyName,
		mount:   s.mount,
	}

	// Get the public key
	pubKey, err := tempSigner.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created public key: %w", err)
	}

	result := &KeyCreationResult{
		KeyName:   keyName,
		KeyType:   keyType,
		PublicKey: pubKey,
	}

	// Extract modulus for RSA keys
	if rsaPubKey, ok := pubKey.(*rsa.PublicKey); ok {
		result.Modulus = hex.EncodeToString(rsaPubKey.N.Bytes())
	}

	return result, nil
}

// CreateKeyInVault is a convenience function to create a key without needing an existing signer instance
func CreateKeyInVault(vaultUrl, token, mountPath, keyName string, keyType KeyType) (*KeyCreationResult, error) {
	// Create a temporary signer for key creation
	tempSigner, err := NewSigner(vaultUrl, token, mountPath, "dummy")
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return tempSigner.CreateKey(keyName, keyType)
}
