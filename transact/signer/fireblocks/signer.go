package fireblocks

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

const (
	// DefaultBaseURL is the default Fireblocks API endpoint
	DefaultBaseURL = "https://api.fireblocks.io"

	// DefaultPollingInterval is the default interval for polling transaction status
	DefaultPollingInterval = 500 * time.Millisecond

	// DefaultTimeout is the default timeout for waiting for a transaction to complete
	DefaultTimeout = 60 * time.Second
)

// TransactionStatus represents the status of a Fireblocks transaction
type TransactionStatus string

const (
	StatusSubmitted            TransactionStatus = "SUBMITTED"
	StatusPendingSignature     TransactionStatus = "PENDING_SIGNATURE"
	StatusPendingAuthorization TransactionStatus = "PENDING_AUTHORIZATION"
	StatusQueued               TransactionStatus = "QUEUED"
	StatusPendingScreening     TransactionStatus = "PENDING_SCREENING"
	StatusPending3rdParty      TransactionStatus = "PENDING_3RD_PARTY"
	StatusBroadcasting         TransactionStatus = "BROADCASTING"
	StatusConfirming           TransactionStatus = "CONFIRMING"
	StatusCompleted            TransactionStatus = "COMPLETED"
	StatusCancelled            TransactionStatus = "CANCELLED"
	StatusRejected             TransactionStatus = "REJECTED"
	StatusFailed               TransactionStatus = "FAILED"
	StatusBlocked              TransactionStatus = "BLOCKED"
)

// HTTPClient is a narrow interface for HTTP operations needed by the signer
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Signer implements signing operations using Fireblocks' custody infrastructure
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

// WithPollingInterval sets the polling interval for transaction status
func WithPollingInterval(d time.Duration) Option {
	return func(s *Signer) {
		s.pollingInterval = d
	}
}

// WithTimeout sets the timeout for waiting for transaction completion
func WithTimeout(d time.Duration) Option {
	return func(s *Signer) {
		s.timeout = d
	}
}

// NewSigner creates a new Fireblocks signer with explicit parameters
func NewSigner(apiKey, privateKeyPEM, vaultAccountID, assetID string, opts ...Option) (*Signer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey cannot be empty")
	}
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("privateKeyPEM cannot be empty")
	}
	if vaultAccountID == "" {
		return nil, fmt.Errorf("vaultAccountID cannot be empty")
	}
	if assetID == "" {
		return nil, fmt.Errorf("assetID cannot be empty")
	}

	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
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
		return nil, fmt.Errorf("FIREBLOCKS_API_KEY environment variable not set")
	}
	if apiSecretEnv == "" {
		return nil, fmt.Errorf("FIREBLOCKS_API_SECRET environment variable not set")
	}
	if vaultAccountID == "" {
		return nil, fmt.Errorf("FIREBLOCKS_VAULT_ACCOUNT_ID environment variable not set")
	}
	if assetID == "" {
		return nil, fmt.Errorf("FIREBLOCKS_ASSET_ID environment variable not set")
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

// Sign implements the Signer interface for signing raw message hashes
func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	// Create a raw signing transaction
	txID, err := s.createRawSigningTransaction(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing transaction: %w", err)
	}

	// Wait for the transaction to complete
	tx, err := s.waitForTransaction(ctx, txID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Extract the signature from the transaction
	sig, err := s.extractSignature(tx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to extract signature: %w", err)
	}

	return sig, nil
}

// GetVaultAccountAddress retrieves the address for the configured vault account
func (s *Signer) GetVaultAccountAddress(ctx context.Context) (string, error) {
	path := fmt.Sprintf("/v1/vault/accounts/%s/%s", s.vaultAccountID, s.assetID)

	resp, err := s.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get vault account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get vault account failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Address, nil
}

// createRawSigningTransaction creates a raw message signing transaction in Fireblocks
func (s *Signer) createRawSigningTransaction(ctx context.Context, hash []byte) (string, error) {
	hashHex := hex.EncodeToString(hash)

	payload := map[string]interface{}{
		"operation": "RAW",
		"assetId":   s.assetID,
		"source": map[string]interface{}{
			"type": "VAULT_ACCOUNT",
			"id":   s.vaultAccountID,
		},
		"note": "CREC SDK raw message signing",
		"extraParameters": map[string]interface{}{
			"rawMessageData": map[string]interface{}{
				"messages": []map[string]interface{}{
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
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create transaction failed with status %d: %s", resp.StatusCode, string(body))
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

// TransactionResponse represents a Fireblocks transaction response
type TransactionResponse struct {
	ID             string            `json:"id"`
	Status         TransactionStatus `json:"status"`
	SignedMessages []SignedMessage   `json:"signedMessages"`
}

// SignedMessage represents a signed message from Fireblocks
type SignedMessage struct {
	Content       string    `json:"content"`
	Algorithm     string    `json:"algorithm"`
	DerivationPath []int    `json:"derivationPath"`
	Signature     Signature `json:"signature"`
	PublicKey     string    `json:"publicKey"`
}

// Signature represents an ECDSA signature
type Signature struct {
	R        string `json:"r"`
	S        string `json:"s"`
	V        int    `json:"v"`
	FullSig  string `json:"fullSig"`
}

// waitForTransaction polls for transaction completion
func (s *Signer) waitForTransaction(ctx context.Context, txID string) (*TransactionResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ticker := time.NewTicker(s.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction %s: %w", txID, timeoutCtx.Err())
		case <-ticker.C:
			tx, err := s.getTransaction(ctx, txID)
			if err != nil {
				return nil, err
			}

			switch tx.Status {
			case StatusCompleted:
				return tx, nil
			case StatusCancelled, StatusRejected, StatusFailed, StatusBlocked:
				return nil, fmt.Errorf("transaction %s ended with status: %s", txID, tx.Status)
			}
		}
	}
}

// getTransaction retrieves a transaction by ID
func (s *Signer) getTransaction(ctx context.Context, txID string) (*TransactionResponse, error) {
	path := fmt.Sprintf("/v1/transactions/%s", txID)

	resp, err := s.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get transaction failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tx TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, fmt.Errorf("failed to decode transaction response: %w", err)
	}

	return &tx, nil
}

var secp256k1N = ethcrypto.S256().Params().N
var secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))

// extractSignature extracts the Ethereum-compatible signature from the transaction
func (s *Signer) extractSignature(tx *TransactionResponse, hash []byte) ([]byte, error) {
	if len(tx.SignedMessages) == 0 {
		return nil, fmt.Errorf("no signed messages in transaction")
	}

	signedMsg := tx.SignedMessages[0]

	// Parse R and S from hex
	rBytes, err := hex.DecodeString(strings.TrimPrefix(signedMsg.Signature.R, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode R: %w", err)
	}
	sBytes, err := hex.DecodeString(strings.TrimPrefix(signedMsg.Signature.S, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode S: %w", err)
	}

	// Adjust S value according to Ethereum standard (EIP-2)
	sBigInt := new(big.Int).SetBytes(sBytes)
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

	if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
		signature = append(rsSignature, byte(1))
		recoveredPublicKeyBytes, err = ethcrypto.Ecrecover(hash, signature)
		if err != nil {
			return nil, err
		}

		if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
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
func (s *Signer) doRequest(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
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
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
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
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}

// GenerateTestPrivateKey generates a test RSA private key (for testing only)
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

// Verify interface compliance at compile time
var _ crypto.Signer = (*rsaSignerAdapter)(nil)

// rsaSignerAdapter is not used but demonstrates crypto.Signer compliance
type rsaSignerAdapter struct {
	key *rsa.PrivateKey
}

func (a *rsaSignerAdapter) Public() crypto.PublicKey {
	return &a.key.PublicKey
}

func (a *rsaSignerAdapter) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return a.key.Sign(rand, digest, opts)
}

