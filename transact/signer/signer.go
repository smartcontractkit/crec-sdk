package signer

import "context"

// Signer signs pre-hashed messages (e.g., keccak256 digests)
type Signer interface {
	Sign(ctx context.Context, hash []byte) ([]byte, error)
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

// TypedDataDomain contains the EIP-712 domain separator parameters
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
