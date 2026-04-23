package fireblocks

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}
var _ signer.TypedDataSigner = &Signer{}

const (
	// DefaultBaseURL is the default Fireblocks API endpoint
	DefaultBaseURL = "https://api.fireblocks.io"

	// DefaultPollingInterval is the default interval for polling operation status
	DefaultPollingInterval = 500 * time.Millisecond

	// DefaultTimeout is the default timeout for waiting for an operation to complete
	DefaultTimeout = 60 * time.Second
)

// Sentinel errors for Fireblocks signer configuration and typed-data encoding.
var (
	ErrAPIKeyRequired         = errors.New("apiKey cannot be empty")
	ErrPrivateKeyPEMRequired  = errors.New("privateKeyPEM cannot be empty")
	ErrVaultAccountIDRequired = errors.New("vaultAccountID cannot be empty")
	ErrAssetIDRequired        = errors.New("assetID cannot be empty")
	ErrEnvFireblocksAPIKey    = errors.New("FIREBLOCKS_API_KEY environment variable not set")
	ErrEnvFireblocksAPISecret = errors.New("FIREBLOCKS_API_SECRET environment variable not set")
	ErrEnvFireblocksVaultAcct = errors.New("FIREBLOCKS_VAULT_ACCOUNT_ID environment variable not set")
	ErrEnvFireblocksAssetID   = errors.New("FIREBLOCKS_ASSET_ID environment variable not set")

	ErrFailedParsePEMBlock     = errors.New("failed to parse PEM block")
	ErrTrailingGarbageAfterPEM = errors.New("trailing garbage bytes after PEM block")
	ErrPrivateKeyNotRSA        = errors.New("private key is not RSA")
	ErrFailedParsePrivateKey   = errors.New("failed to parse private key")
	ErrTypedDataNil            = errors.New("typedData cannot be nil")

	ErrNegativeUnsignedTypedValue  = errors.New("negative value for unsigned type")
	ErrFloat64PrecisionLoss        = errors.New("float64 precision loss")
	ErrParseTypedDataIntegerString = errors.New("failed to parse string as integer")

	ErrFireblocksOperationTerminal = errors.New("Fireblocks operation ended with terminal status")

	ErrCreateSigningOperationFailed      = errors.New("create signing operation failed")
	ErrCreateTypedMessageOperationFailed = errors.New("create typed message operation failed")

	ErrGetVaultAccountNonOK = errors.New("get vault account request failed")
)

// OperationStatus represents the status of a Fireblocks signing operation.
type OperationStatus string

const (
	StatusSubmitted            OperationStatus = "SUBMITTED"
	StatusPendingSignature     OperationStatus = "PENDING_SIGNATURE"
	StatusPendingAuthorization OperationStatus = "PENDING_AUTHORIZATION"
	StatusQueued               OperationStatus = "QUEUED"
	StatusPendingScreening     OperationStatus = "PENDING_SCREENING"
	StatusPending3rdParty      OperationStatus = "PENDING_3RD_PARTY"
	StatusBroadcasting         OperationStatus = "BROADCASTING"
	StatusConfirming           OperationStatus = "CONFIRMING"
	StatusCompleted            OperationStatus = "COMPLETED"
	StatusCancelled            OperationStatus = "CANCELLED"
	StatusRejected             OperationStatus = "REJECTED"
	StatusFailed               OperationStatus = "FAILED"
	StatusBlocked              OperationStatus = "BLOCKED"
)

// HTTPClient is a narrow interface for HTTP operations needed by the signer
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Signer implements the signer.Signer and signer.TypedDataSigner interfaces using Fireblocks custody.
type Signer struct {
	client          HTTPClient
	apiKey          string
	privateKey      *rsa.PrivateKey
	baseURL         string
	vaultAccountID  string
	assetID         string
	pollingInterval time.Duration
	timeout         time.Duration
}

// Option is a functional option for configuring the Signer
type Option func(*Signer)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(c HTTPClient) Option {
	return func(s *Signer) {
		s.client = c
	}
}

// WithBaseURL sets a custom base URL for the Fireblocks API
func WithBaseURL(url string) Option {
	return func(s *Signer) {
		s.baseURL = url
	}
}

// WithPollingInterval sets the polling interval for operation status
func WithPollingInterval(d time.Duration) Option {
	return func(s *Signer) {
		s.pollingInterval = d
	}
}

// WithTimeout sets the timeout for waiting for operation completion
func WithTimeout(d time.Duration) Option {
	return func(s *Signer) {
		s.timeout = d
	}
}

// NewSigner creates a new Fireblocks signer with explicit parameters
func NewSigner(apiKey, privateKeyPEM, vaultAccountID, assetID string, opts ...Option) (*Signer, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyRequired
	}
	if privateKeyPEM == "" {
		return nil, ErrPrivateKeyPEMRequired
	}
	if vaultAccountID == "" {
		return nil, ErrVaultAccountIDRequired
	}
	if assetID == "" {
		return nil, ErrAssetIDRequired
	}

	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedParsePrivateKey, err)
	}

	s := &Signer{
		client:          &http.Client{},
		apiKey:          apiKey,
		privateKey:      privateKey,
		baseURL:         DefaultBaseURL,
		vaultAccountID:  vaultAccountID,
		assetID:         assetID,
		pollingInterval: DefaultPollingInterval,
		timeout:         DefaultTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// NewSignerFromEnv creates a new Fireblocks signer using environment variables:
// FIREBLOCKS_API_KEY, FIREBLOCKS_API_SECRET (path to PEM file or PEM content),
// FIREBLOCKS_VAULT_ACCOUNT_ID, FIREBLOCKS_ASSET_ID, and optionally FIREBLOCKS_BASE_URL
func NewSignerFromEnv(opts ...Option) (*Signer, error) {
	apiKey := os.Getenv("FIREBLOCKS_API_KEY")
	apiSecretEnv := os.Getenv("FIREBLOCKS_API_SECRET")
	vaultAccountID := os.Getenv("FIREBLOCKS_VAULT_ACCOUNT_ID")
	assetID := os.Getenv("FIREBLOCKS_ASSET_ID")
	baseURL := os.Getenv("FIREBLOCKS_BASE_URL")

	if apiKey == "" {
		return nil, ErrEnvFireblocksAPIKey
	}
	if apiSecretEnv == "" {
		return nil, ErrEnvFireblocksAPISecret
	}
	if vaultAccountID == "" {
		return nil, ErrEnvFireblocksVaultAcct
	}
	if assetID == "" {
		return nil, ErrEnvFireblocksAssetID
	}

	var privateKeyPEM string
	if strings.HasPrefix(apiSecretEnv, "-----BEGIN") {
		privateKeyPEM = apiSecretEnv
	} else {
		// Treat as file path
		data, err := os.ReadFile(apiSecretEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		privateKeyPEM = string(data)
	}

	allOpts := opts
	if baseURL != "" {
		allOpts = append([]Option{WithBaseURL(baseURL)}, opts...)
	}

	return NewSigner(apiKey, privateKeyPEM, vaultAccountID, assetID, allOpts...)
}

// Sign signs the pre-hashed message using Fireblocks custody.
// Creates a RAW signing operation and polls until completion.
func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	// Create a raw signing operation
	opID, err := s.createSigningOperation(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing operation: %w", err)
	}

	// Wait for the operation to complete
	op, err := s.waitForOperation(ctx, opID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for operation: %w", err)
	}

	// Extract the signature from the operation response
	sig, err := s.extractSignature(op, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to extract signature: %w", err)
	}

	return sig, nil
}

// SignTypedData signs EIP-712 typed data using Fireblocks' TYPED_MESSAGE operation.
// Fireblocks can see the full typed data for policy enforcement before signing.
func (s *Signer) SignTypedData(ctx context.Context, typedData *signer.TypedData) ([]byte, error) {
	if typedData == nil {
		return nil, ErrTypedDataNil
	}

	// Create the typed message signing operation
	opID, err := s.createTypedMessageOperation(ctx, typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to create typed message operation: %w", err)
	}

	// Wait for the operation to complete
	op, err := s.waitForOperation(ctx, opID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for operation: %w", err)
	}

	// Compute the EIP-712 hash for signature extraction
	hash, err := HashTypedData(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash typed data: %w", err)
	}

	// Extract the signature from the operation response
	sig, err := s.extractSignature(op, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to extract signature: %w", err)
	}

	return sig, nil
}

// GetVaultAccountAddress retrieves the Ethereum address for the configured vault account.
func (s *Signer) GetVaultAccountAddress(ctx context.Context) (string, error) {
	path := fmt.Sprintf("/v1/vault/accounts/%s/%s", s.vaultAccountID, s.assetID)

	resp, err := s.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get vault account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("%w: status %d: %s", ErrGetVaultAccountNonOK, resp.StatusCode, string(body))
	}

	var result struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Address, nil
}

// createSigningOperation creates a raw message signing operation in Fireblocks
func (s *Signer) createSigningOperation(ctx context.Context, hash []byte) (string, error) {
	hashHex := hex.EncodeToString(hash)

	payload := map[string]any{
		"operation": "RAW",
		"assetId":   s.assetID,
		"source": map[string]any{
			"type": "VAULT_ACCOUNT",
			"id":   s.vaultAccountID,
		},
		"note": "CREC SDK raw message signing",
		"extraParameters": map[string]any{
			"rawMessageData": map[string]any{
				"messages": []map[string]any{
					{
						"content": hashHex,
					},
				},
			},
		},
	}

	resp, err := s.doRequest(ctx, "POST", "/v1/transactions", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("%w: status %d: %s", ErrCreateSigningOperationFailed, resp.StatusCode, string(body))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

// createTypedMessageOperation creates an EIP-712 typed message signing operation in Fireblocks
func (s *Signer) createTypedMessageOperation(ctx context.Context, typedData *signer.TypedData) (string, error) {
	// Convert domain to Fireblocks format
	domain := map[string]any{}
	if typedData.Domain.Name != "" {
		domain["name"] = typedData.Domain.Name
	}
	if typedData.Domain.Version != "" {
		domain["version"] = typedData.Domain.Version
	}
	if typedData.Domain.ChainID != 0 {
		domain["chainId"] = typedData.Domain.ChainID
	}
	if typedData.Domain.VerifyingContract != "" {
		domain["verifyingContract"] = typedData.Domain.VerifyingContract
	}
	if typedData.Domain.Salt != "" {
		domain["salt"] = typedData.Domain.Salt
	}

	// Convert types to Fireblocks format (array of {name, type} objects)
	types := make(map[string][]map[string]string)
	for typeName, fields := range typedData.Types {
		typeFields := make([]map[string]string, len(fields))
		for i, field := range fields {
			typeFields[i] = map[string]string{
				"name": field.Name,
				"type": field.Type,
			}
		}
		types[typeName] = typeFields
	}

	payload := map[string]any{
		"operation": "TYPED_MESSAGE",
		"assetId":   s.assetID,
		"source": map[string]any{
			"type": "VAULT_ACCOUNT",
			"id":   s.vaultAccountID,
		},
		"note": "CREC SDK EIP-712 typed message signing",
		"extraParameters": map[string]any{
			"typedMessageData": map[string]any{
				"types":       types,
				"primaryType": typedData.PrimaryType,
				"domain":      domain,
				"message":     typedData.Message,
			},
		},
	}

	resp, err := s.doRequest(ctx, "POST", "/v1/transactions", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("%w: status %d: %s", ErrCreateTypedMessageOperationFailed, resp.StatusCode, string(body))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

// HashTypedData computes the EIP-712 hash of typed data for signature verification.
// This is exported for testing purposes.
func HashTypedData(typedData *signer.TypedData) ([]byte, error) {
	// Build the EIP712Domain type if not present
	types := make(map[string][]signer.TypedDataField)
	for k, v := range typedData.Types {
		types[k] = v
	}

	// Add EIP712Domain if not present
	if _, ok := types["EIP712Domain"]; !ok {
		var domainFields []signer.TypedDataField
		if typedData.Domain.Name != "" {
			domainFields = append(domainFields, signer.TypedDataField{Name: "name", Type: "string"})
		}
		if typedData.Domain.Version != "" {
			domainFields = append(domainFields, signer.TypedDataField{Name: "version", Type: "string"})
		}
		if typedData.Domain.ChainID != 0 {
			domainFields = append(domainFields, signer.TypedDataField{Name: "chainId", Type: "uint256"})
		}
		if typedData.Domain.VerifyingContract != "" {
			domainFields = append(domainFields, signer.TypedDataField{Name: "verifyingContract", Type: "address"})
		}
		if typedData.Domain.Salt != "" {
			domainFields = append(domainFields, signer.TypedDataField{Name: "salt", Type: "bytes32"})
		}
		types["EIP712Domain"] = domainFields
	}

	// Encode the domain separator
	domainData := make(map[string]any)
	if typedData.Domain.Name != "" {
		domainData["name"] = typedData.Domain.Name
	}
	if typedData.Domain.Version != "" {
		domainData["version"] = typedData.Domain.Version
	}
	if typedData.Domain.ChainID != 0 {
		domainData["chainId"] = big.NewInt(typedData.Domain.ChainID)
	}
	if typedData.Domain.VerifyingContract != "" {
		domainData["verifyingContract"] = typedData.Domain.VerifyingContract
	}
	if typedData.Domain.Salt != "" {
		domainData["salt"] = typedData.Domain.Salt
	}

	domainSeparator, err := hashStruct("EIP712Domain", domainData, types)
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain: %w", err)
	}

	// Encode the message
	messageHash, err := hashStruct(typedData.PrimaryType, typedData.Message, types)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message: %w", err)
	}

	// Combine: keccak256("\x19\x01" || domainSeparator || messageHash)
	rawData := make([]byte, 0, 2+32+32)
	rawData = append(rawData, 0x19, 0x01)
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, messageHash...)

	return ethcrypto.Keccak256(rawData), nil
}

// hashStruct computes the hash of a struct according to EIP-712
func hashStruct(typeName string, data map[string]any, types map[string][]signer.TypedDataField) ([]byte, error) {
	typeHash := hashType(typeName, types)
	encodedData, err := encodeData(typeName, data, types)
	if err != nil {
		return nil, err
	}

	combined := make([]byte, 0, len(typeHash)+len(encodedData))
	combined = append(combined, typeHash...)
	combined = append(combined, encodedData...)

	return ethcrypto.Keccak256(combined), nil
}

// hashType computes the type hash according to EIP-712
func hashType(typeName string, types map[string][]signer.TypedDataField) []byte {
	typeString := encodeType(typeName, types)
	return ethcrypto.Keccak256([]byte(typeString))
}

// encodeType creates the type encoding string for EIP-712
func encodeType(typeName string, types map[string][]signer.TypedDataField) string {
	fields := types[typeName]
	var parts []string
	for _, field := range fields {
		parts = append(parts, field.Type+" "+field.Name)
	}

	result := typeName + "(" + strings.Join(parts, ",") + ")"

	// Find and append referenced types (sorted alphabetically)
	deps := findTypeDependencies(typeName, types, make(map[string]bool))
	delete(deps, typeName)

	var sortedDeps []string
	for dep := range deps {
		sortedDeps = append(sortedDeps, dep)
	}
	// Simple sort
	for i := 0; i < len(sortedDeps); i++ {
		for j := i + 1; j < len(sortedDeps); j++ {
			if sortedDeps[i] > sortedDeps[j] {
				sortedDeps[i], sortedDeps[j] = sortedDeps[j], sortedDeps[i]
			}
		}
	}

	for _, dep := range sortedDeps {
		depFields := types[dep]
		var depParts []string
		for _, field := range depFields {
			depParts = append(depParts, field.Type+" "+field.Name)
		}
		result += dep + "(" + strings.Join(depParts, ",") + ")"
	}

	return result
}

// findTypeDependencies recursively finds all type dependencies
func findTypeDependencies(typeName string, types map[string][]signer.TypedDataField, visited map[string]bool) map[string]bool {
	if visited[typeName] {
		return visited
	}
	visited[typeName] = true

	fields, ok := types[typeName]
	if !ok {
		return visited
	}

	for _, field := range fields {
		// Check if the type is a custom type (not a primitive)
		baseType := strings.TrimSuffix(field.Type, "[]")
		if _, isCustom := types[baseType]; isCustom {
			findTypeDependencies(baseType, types, visited)
		}
	}

	return visited
}

// encodeData encodes the data according to EIP-712
func encodeData(typeName string, data map[string]any, types map[string][]signer.TypedDataField) ([]byte, error) {
	fields := types[typeName]
	var encoded []byte

	for _, field := range fields {
		value, ok := data[field.Name]
		if !ok {
			// Use zero value
			encoded = append(encoded, make([]byte, 32)...)
			continue
		}

		encodedValue, err := encodeValue(field.Type, value, types)
		if err != nil {
			return nil, fmt.Errorf("failed to encode field %s: %w", field.Name, err)
		}
		encoded = append(encoded, encodedValue...)
	}

	return encoded, nil
}

// encodeValue encodes a single value according to EIP-712
func encodeValue(typeName string, value any, types map[string][]signer.TypedDataField) ([]byte, error) {
	// Handle arrays
	if strings.HasSuffix(typeName, "[]") {
		baseType := strings.TrimSuffix(typeName, "[]")
		arr, ok := value.([]any)
		if !ok {
			return nil, fmt.Errorf("expected array for type %s", typeName)
		}
		var hashes []byte
		for _, item := range arr {
			encoded, err := encodeValue(baseType, item, types)
			if err != nil {
				return nil, err
			}
			hashes = append(hashes, encoded...)
		}
		return ethcrypto.Keccak256(hashes), nil
	}

	// Handle custom struct types
	if _, isCustom := types[typeName]; isCustom {
		mapValue, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map for struct type %s", typeName)
		}
		return hashStruct(typeName, mapValue, types)
	}

	// Handle primitive types
	return encodePrimitive(typeName, value)
}

// encodePrimitive encodes primitive EIP-712 types
func encodePrimitive(typeName string, value any) ([]byte, error) {
	result := make([]byte, 32)

	switch {
	case typeName == "string":
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		return ethcrypto.Keccak256([]byte(str)), nil

	case typeName == "bytes":
		var data []byte
		switch v := value.(type) {
		case string:
			data = []byte(strings.TrimPrefix(v, "0x"))
			decoded, err := hex.DecodeString(string(data))
			if err != nil {
				data = []byte(v)
			} else {
				data = decoded
			}
		case []byte:
			data = v
		default:
			return nil, fmt.Errorf("expected bytes, got %T", value)
		}
		return ethcrypto.Keccak256(data), nil

	case typeName == "bool":
		b, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, got %T", value)
		}
		if b {
			result[31] = 1
		}
		return result, nil

	case typeName == "address":
		var addr string
		switch v := value.(type) {
		case string:
			addr = v
		default:
			return nil, fmt.Errorf("expected address string, got %T", value)
		}
		addrBytes, err := hex.DecodeString(strings.TrimPrefix(addr, "0x"))
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
		copy(result[32-len(addrBytes):], addrBytes)
		return result, nil

	case strings.HasPrefix(typeName, "uint") || strings.HasPrefix(typeName, "int"):
		isUint := strings.HasPrefix(typeName, "uint")
		prefixLen := 3
		if isUint {
			prefixLen = 4
		}

		bitWidth := 256
		if len(typeName) > prefixLen {
			var err error
			bitWidth, err = strconv.Atoi(typeName[prefixLen:])
			if err != nil || bitWidth <= 0 || bitWidth > 256 || bitWidth%8 != 0 {
				return nil, fmt.Errorf("invalid integer type: %s", typeName)
			}
		}

		var n *big.Int
		switch v := value.(type) {
		case float64:
			// float64 can only represent integers exactly up to 2^53 - 1 (9007199254740991).
			// Values beyond this may have already lost precision during parsing/unmarshaling.
			if v > 9007199254740991 || v < -9007199254740991 {
				return nil, fmt.Errorf("%w for value %v", ErrFloat64PrecisionLoss, v)
			}
			if math.Trunc(v) != v {
				return nil, fmt.Errorf("float64 value %v is not an integer", v)
			}
			n = big.NewInt(int64(v))
		case int:
			n = big.NewInt(int64(v))
		case int64:
			n = big.NewInt(v)
		case string:
			n = new(big.Int)
			var ok bool
			if strings.HasPrefix(v, "0x") {
				_, ok = n.SetString(v[2:], 16)
			} else {
				_, ok = n.SetString(v, 10)
			}
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrParseTypedDataIntegerString, v)
			}
		case *big.Int:
			n = new(big.Int).Set(v)
		default:
			return nil, fmt.Errorf("expected number for %s, got %T", typeName, value)
		}

		if isUint && n.Sign() < 0 {
			return nil, fmt.Errorf("%w: %s", ErrNegativeUnsignedTypedValue, typeName)
		}

		if isUint {
			max := new(big.Int).Lsh(big.NewInt(1), uint(bitWidth))
			max.Sub(max, big.NewInt(1))
			if n.Cmp(max) > 0 {
				return nil, fmt.Errorf("value exceeds bit width for type %s", typeName)
			}
		} else {
			max := new(big.Int).Lsh(big.NewInt(1), uint(bitWidth-1))
			min := new(big.Int).Neg(max)
			max.Sub(max, big.NewInt(1))
			if n.Cmp(max) > 0 || n.Cmp(min) < 0 {
				return nil, fmt.Errorf("value exceeds bit width for type %s", typeName)
			}
		}

		var bytes []byte
		if n.Sign() < 0 {
			// Convert to two's complement for negative signed integers (EIP-712 spec)
			mask := new(big.Int).Lsh(big.NewInt(1), 256)
			twoComp := new(big.Int).Add(n, mask)
			bytes = twoComp.Bytes()
		} else {
			bytes = n.Bytes()
		}

		copy(result[32-len(bytes):], bytes)
		return result, nil

	case strings.HasPrefix(typeName, "bytes"):
		// Fixed size bytes (bytes1 to bytes32)
		var data []byte
		switch v := value.(type) {
		case string:
			decoded, err := hex.DecodeString(strings.TrimPrefix(v, "0x"))
			if err != nil {
				return nil, fmt.Errorf("invalid hex for %s: %w", typeName, err)
			}
			data = decoded
		case []byte:
			data = v
		default:
			return nil, fmt.Errorf("expected bytes for %s, got %T", typeName, value)
		}
		copy(result, data)
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", typeName)
	}
}

// OperationResponse represents a Fireblocks signing operation response.
type OperationResponse struct {
	ID             string          `json:"id"`
	Status         OperationStatus `json:"status"`
	SignedMessages []SignedMessage `json:"signedMessages"`
}

// SignedMessage represents a signed message from a Fireblocks operation.
type SignedMessage struct {
	Content        string    `json:"content"`
	Algorithm      string    `json:"algorithm"`
	DerivationPath []int     `json:"derivationPath"`
	Signature      Signature `json:"signature"`
	PublicKey      string    `json:"publicKey"`
}

// Signature represents an ECDSA signature (r, s, v) from Fireblocks.
type Signature struct {
	R       string `json:"r"`
	S       string `json:"s"`
	V       int    `json:"v"`
	FullSig string `json:"fullSig"`
}

// waitForOperation polls for signing operation completion
func (s *Signer) waitForOperation(ctx context.Context, opID string) (*OperationResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for operation %s: %w", opID, timeoutCtx.Err())
		case <-ticker.C:
			op, err := s.getOperation(timeoutCtx, opID)
			if err != nil {
				return nil, err
			}

			switch op.Status {
			case StatusCompleted:
				return op, nil
			case StatusCancelled, StatusRejected, StatusFailed, StatusBlocked:
				return nil, fmt.Errorf("%w: operation %s ended with status %s", ErrFireblocksOperationTerminal, opID, op.Status)
			}
		}
	}
}

// getOperation retrieves a signing operation by ID
func (s *Signer) getOperation(ctx context.Context, opID string) (*OperationResponse, error) {
	path := fmt.Sprintf("/v1/transactions/%s", opID)

	resp, err := s.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("get operation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var op OperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		return nil, fmt.Errorf("failed to decode operation response: %w", err)
	}

	return &op, nil
}

var secp256k1N = ethcrypto.S256().Params().N
var secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))

// extractSignature extracts the Ethereum-compatible signature from the operation response
func (s *Signer) extractSignature(op *OperationResponse, hash []byte) ([]byte, error) {
	if len(op.SignedMessages) == 0 {
		return nil, fmt.Errorf("no signed messages in operation response")
	}

	signedMsg := op.SignedMessages[0]

	// Parse R and S from hex
	rBytes, err := hex.DecodeString(strings.TrimPrefix(signedMsg.Signature.R, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode R: %w", err)
	}
	sBytes, err := hex.DecodeString(strings.TrimPrefix(signedMsg.Signature.S, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode S: %w", err)
	}

	rBigInt := new(big.Int).SetBytes(rBytes)
	sBigInt := new(big.Int).SetBytes(sBytes)
	if rBigInt.Cmp(big.NewInt(0)) <= 0 || rBigInt.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("R value out of range [1, N-1]")
	}
	if sBigInt.Cmp(big.NewInt(0)) <= 0 || sBigInt.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("S value out of range [1, N-1]")
	}

	// Adjust S value according to Ethereum standard (EIP-2)
	sBigInt = new(big.Int).SetBytes(sBytes)
	if sBigInt.Cmp(secp256k1HalfN) > 0 {
		sBytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
	}

	// Pad R and S to 32 bytes
	rBytes = padTo32Bytes(rBytes)
	sBytes = padTo32Bytes(sBytes)

	// Parse public key
	pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(signedMsg.PublicKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// If compressed public key, decompress it
	if len(pubKeyBytes) == 33 {
		pubKey, err := ethcrypto.DecompressPubkey(pubKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress public key: %w", err)
		}
		if pubKey == nil || pubKey.X == nil || pubKey.Y == nil {
			return nil, fmt.Errorf("invalid decompressed public key")
		}
		pubKeyBytes = secp256k1.S256().Marshal(pubKey.X, pubKey.Y)
	}

	// Build the signature and recover V
	signature, err := getEthereumSignature(pubKeyBytes, hash, rBytes, sBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to construct ethereum signature: %w", err)
	}

	return signature, nil
}

// getEthereumSignature constructs an Ethereum-compatible signature with correct V value
func getEthereumSignature(expectedPublicKeyBytes []byte, hash []byte, r []byte, s []byte) ([]byte, error) {
	rsSignature := append(r, s...)
	signature := append(rsSignature, byte(0))

	recoveredPublicKeyBytes, err := ethcrypto.Ecrecover(hash, signature)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(recoveredPublicKeyBytes, expectedPublicKeyBytes) {
		signature = append(rsSignature, byte(1))
		recoveredPublicKeyBytes, err = ethcrypto.Ecrecover(hash, signature)
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(recoveredPublicKeyBytes, expectedPublicKeyBytes) {
			return nil, fmt.Errorf("cannot reconstruct public key from signature")
		}
	}

	return signature, nil
}

// padTo32Bytes pads a byte slice to 32 bytes
func padTo32Bytes(b []byte) []byte {
	b = bytes.TrimLeft(b, "\x00")
	for len(b) < 32 {
		b = append([]byte{0}, b...)
	}
	return b
}

// doRequest performs an authenticated request to the Fireblocks API
func (s *Signer) doRequest(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var body []byte
	var err error

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := s.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := s.signJWT(path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT: %w", err)
	}

	req.Header.Set("X-API-Key", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.client.Do(req)
}

// signJWT creates a signed JWT for Fireblocks API authentication
func (s *Signer) signJWT(path string, body []byte) (string, error) {
	now := time.Now()

	// Calculate body hash
	bodyHash := sha256.Sum256(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])

	// Generate nonce
	nonce := uuid.New().String()

	claims := jwt.MapClaims{
		"uri":      path,
		"nonce":    nonce,
		"iat":      now.Unix(),
		"exp":      now.Add(30 * time.Second).Unix(),
		"sub":      s.apiKey,
		"bodyHash": bodyHashHex,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// parsePrivateKey parses a PEM-encoded RSA private key
func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, ErrFailedParsePEMBlock
	}
	if len(bytes.TrimSpace(rest)) > 0 {
		return nil, ErrTrailingGarbageAfterPEM
	}

	// Try PKCS1 first
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	// Try PKCS8
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrPrivateKeyNotRSA
	}

	return rsaKey, nil
}

// GenerateTestPrivateKey generates a 2048-bit RSA private key for testing.
func GenerateTestPrivateKey() (*rsa.PrivateKey, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return key, string(pem.EncodeToMemory(pemBlock)), nil
}
