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

	"github.com/smartcontractkit/cvn-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

// Signer implements signing operations using Privy's wallet API
type Signer struct {
	client    *http.Client
	appID     string
	appSecret string
	baseURL   string
	walletID  string
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
func NewSigner(appID, appSecret, walletID, baseURL string) (*Signer, error) {
	if appID == "" || appSecret == "" || walletID == "" || baseURL == "" {
		return nil, fmt.Errorf("appID, appSecret, walletID, and baseURL must be set")
	}

	return &Signer{
		client:    &http.Client{},
		appID:     appID,
		appSecret: appSecret,
		baseURL:   baseURL,
		walletID:  walletID,
	}, nil
}

// NewSignerFromEnv creates a new Privy signer using environment variables
// PRIVY_APP_ID, PRIVY_APP_SECRET, PRIVY_WALLET_ID, and optionally PRIVY_BASE_URL
func NewSignerFromEnv() (*Signer, error) {
	appID := os.Getenv("PRIVY_APP_ID")
	appSecret := os.Getenv("PRIVY_APP_SECRET")
	walletID := os.Getenv("PRIVY_WALLET_ID")
	baseURL := os.Getenv("PRIVY_BASE_URL")

	if baseURL == "" {
		baseURL = "https://api.privy.io"
	}

	if appID == "" {
		return nil, fmt.Errorf("PRIVY_APP_ID environment variable not set")
	}
	if appSecret == "" {
		return nil, fmt.Errorf("PRIVY_APP_SECRET environment variable not set")
	}
	if walletID == "" {
		return nil, fmt.Errorf("PRIVY_WALLET_ID environment variable not set")
	}

	return NewSigner(appID, appSecret, walletID, baseURL)
}

// NewSignerWithCustomClient creates a new Privy signer with a custom HTTP client (useful for testing)
func NewSignerWithCustomClient(appID, appSecret, walletID string, client *http.Client, baseURL string) (*Signer, error) {
	if appID == "" || appSecret == "" || walletID == "" || baseURL == "" {
		return nil, fmt.Errorf("appID, appSecret, walletID, and baseURL must be set")
	}

	if client == nil {
		client = &http.Client{}
	}

	return &Signer{
		client:    client,
		appID:     appID,
		appSecret: appSecret,
		baseURL:   baseURL,
		walletID:  walletID,
	}, nil
}

func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	hashHex := "0x" + hex.EncodeToString(hash)

	rpcReq := RPCRequest{
		Method: "personal_sign",
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

// GetWalletAddress retrieves the Ethereum address for this wallet from Privy
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
