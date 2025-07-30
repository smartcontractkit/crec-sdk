package signer

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

type TransitSigner struct {
	client  *vault.Client
	keyName string
	mount   string
}

func NewTransitSigner(vaultUrl, token, mountPath, key string) (*TransitSigner, error) {
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultUrl)
	client.SetToken(token)

	return &TransitSigner{
		client:  client,
		keyName: key,
		mount:   mountPath, // usually "transit"
	}, nil
}

func (s *TransitSigner) Sign(hash []byte) ([]byte, error) {
	// base64 encore the payload to sign
	b64 := base64.StdEncoding.EncodeToString(hash)

	// call vault client to sign payload
	resp, err := s.client.Logical().Write(fmt.Sprintf("%s/sign/%s", s.mount, s.keyName), map[string]any{
		"input":                b64,
		"prehashed":            true,
		"marshaling_algorithm": "asn1",
	})
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
func (s *TransitSigner) Public() (interface{}, error) {
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
