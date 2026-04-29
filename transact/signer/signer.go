package signer

import "context"

// Signer signs pre-hashed messages (e.g., keccak256 digests)
type Signer interface {
	Sign(ctx context.Context, hash []byte) ([]byte, error)
}

// RSAPublicKeyInfo contains the public components of an RSA key in the
// string encoding used by the CREC platform: hex-encoded exponent and modulus.
// These fields are used wherever the platform needs to identify or register an
// RSA public key — for example, when configuring allowed signers on a Smart
// Wallet, or when verifying key ownership out-of-band.
//   - E is the hex-encoded public exponent (e.g. "010001" for 65537).
//   - N is the hex-encoded modulus.
type RSAPublicKeyInfo struct {
	E string
	N string
}

// RSAPublicKeyExporter is an optional interface implemented by RSA-capable
// signers. It provides a signer-implementation-agnostic way to obtain the
// public components of an RSA key in the string encoding expected by the CREC
// platform, regardless of whether the key lives in memory, Vault, or another
// backend.
//
// Not all signers implement this interface; callers should type-assert before use:
//
//	if exporter, ok := s.(signer.RSAPublicKeyExporter); ok {
//	    info, err := exporter.RSAPublicKey()
//	    // use info.E and info.N
//	}
type RSAPublicKeyExporter interface {
	RSAPublicKey() (RSAPublicKeyInfo, error)
}

// TypedDataSigner signs EIP-712 typed structured data.
// This interface allows custody providers to see the full typed data
// for policy enforcement before signing.
type TypedDataSigner interface {
	SignTypedData(ctx context.Context, typedData *TypedData) ([]byte, error)
}

// TypedData represents EIP-712 typed structured data
type TypedData struct {
	// Types contains type definitions for the structured data
	Types map[string][]TypedDataField `json:"types"`
	// PrimaryType is the primary type being signed
	PrimaryType string `json:"primaryType"`
	// Domain contains the signing domain separator parameters
	Domain TypedDataDomain `json:"domain"`
	// Message contains the structured data to be signed
	Message map[string]any `json:"message"`
}

// TypedDataField represents a single field in a struct type
type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypedDataDomain contains the EIP-712 domain separator parameters for signing.
type TypedDataDomain struct {
	// Name is the user-readable name of the signing domain
	Name string `json:"name,omitempty"`
	// Version is the current major version of the signing domain
	Version string `json:"version,omitempty"`
	// ChainID is the EIP-155 chain ID
	ChainID int64 `json:"chainId,omitempty"`
	// VerifyingContract is the address of the contract that will verify the signature
	VerifyingContract string `json:"verifyingContract,omitempty"`
	// Salt is a disambiguating salt for the protocol
	Salt string `json:"salt,omitempty"`
}
