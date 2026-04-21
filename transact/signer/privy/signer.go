package privy

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/smartcontractkit/crec-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

const (
	// DefaultBaseURL is the default Privy API endpoint
	DefaultBaseURL = "https://api.privy.io"
)

// HTTPClient is a narrow interface for HTTP operations needed by the signer
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Signer implements signing operations using Privy's wallet API
type Signer struct {
	client    HTTPClient
	appID     string
	appSecret string
	baseURL   string
	walletID  string
}

// Option is a functional option for configuring the Signer
type Option func(*Signer)

// WithHTTPClient sets a custom HTTP client (useful for testing)
func WithHTTPClient(c HTTPClient) Option {
	return func(s *Signer) {
		s.client = c
	}
}

// WithBaseURL sets a custom base URL for the Privy API
func WithBaseURL(url string) Option {
	return func(s *Signer) {
		s.baseURL = url
	}
}

// RPCRequest represents an RPC request to Privy
type RPCRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// RPCResponse represents the response from Privy's RPC API
type RPCResponse struct {
	Method string `json:"method"`
	Data   struct {
		Signature string `json:"signature"`
		Encoding  string `json:"encoding"`
	} `json:"data"`
}

// WalletResponse represents the response from Privy's wallet info API
type WalletResponse struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	ChainType    string `json:"chain_type"`
	WalletIndex  int    `json:"wallet_index"`
	PublicKey    string `json:"public_key"`
	WalletClient string `json:"wallet_client"`
	Imported     bool   `json:"imported"`
}

// NewSigner creates a new Privy signer with explicit parameters
func NewSigner(appID, appSecret, walletID string, opts ...Option) (*Signer, error) {
	if appID == "" {
		return nil, fmt.Errorf("appID cannot be empty")
	}
	if appSecret == "" {
		return nil, fmt.Errorf("appSecret cannot be empty")
	}
	if walletID == "" {
		return nil, fmt.Errorf("walletID cannot be empty")
	}

	s := &Signer{
		client:    &http.Client{},
		appID:     appID,
		appSecret: appSecret,
		baseURL:   DefaultBaseURL,
		walletID:  walletID,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// NewSignerFromEnv creates a new Privy signer using environment variables
// PRIVY_APP_ID, PRIVY_APP_SECRET, PRIVY_WALLET_ID, and optionally PRIVY_BASE_URL
func NewSignerFromEnv(opts ...Option) (*Signer, error) {
	appID := os.Getenv("PRIVY_APP_ID")
	appSecret := os.Getenv("PRIVY_APP_SECRET")
	walletID := os.Getenv("PRIVY_WALLET_ID")
	baseURL := os.Getenv("PRIVY_BASE_URL")

	if appID == "" {
		return nil, fmt.Errorf("PRIVY_APP_ID environment variable not set")
	}
	if appSecret == "" {
		return nil, fmt.Errorf("PRIVY_APP_SECRET environment variable not set")
	}
	if walletID == "" {
		return nil, fmt.Errorf("PRIVY_WALLET_ID environment variable not set")
	}

	allOpts := opts
	if baseURL != "" {
		allOpts = append([]Option{WithBaseURL(baseURL)}, opts...)
	}

	return NewSigner(appID, appSecret, walletID, allOpts...)
}

// NewSignerWithCustomClient creates a Privy signer with a custom HTTP client.
// Deprecated: Use NewSigner with WithHTTPClient and WithBaseURL options instead.

	// Sign signs the pre-hashed message using Privy's secp256k1_sign RPC.
	// Returns the raw signature bytes.
	func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
		hashHex := "0x" + hex.EncodeToString(hash)
	
		rpcReq := RPCRequest{
			Method: "secp256k1_sign",
			Params: map[string]interface{}{
				"message":  hashHex,
				"encoding": "hex",
			},
		}

	jsonData, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/wallets/%s/rpc", s.baseURL, s.walletID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	s.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute RPC request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	sigBytes, err := hex.DecodeString(strings.TrimPrefix(rpcResp.Data.Signature, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	return sigBytes, nil
}

// GetWalletAddress retrieves the Ethereum address for this wallet from Privy.
func (s *Signer) GetWalletAddress(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/v1/wallets/%s", s.baseURL, s.walletID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	s.setAuthHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute get wallet request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get wallet request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var wallet WalletResponse
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return "", fmt.Errorf("failed to decode get wallet response: %w", err)
	}

	return wallet.Address, nil
}

// setAuthHeaders sets the required authentication headers for Privy API
func (s *Signer) setAuthHeaders(req *http.Request) {
	// Basic Auth with app ID as username and app secret as password
	req.SetBasicAuth(s.appID, s.appSecret)
	req.Header.Set("privy-app-id", s.appID)
}
