package hasher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	chainselectors "github.com/smartcontractkit/chain-selectors"

	"github.com/smartcontractkit/crec-sdk/transact/signer"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

// Sentinel errors
var (
	// Validation errors
	ErrOperationRequired = errors.New("operation is required")
	ErrSignerRequired    = errors.New("signer is required")

	// Chain/Signing errors
	ErrParseChainSelector     = errors.New("failed to parse chain selector")
	ErrGetChainFamily         = errors.New("failed to get chain family")
	ErrUnsupportedChainFamily = errors.New("chain family is not supported")
	ErrGetChainID             = errors.New("failed to get chain ID from selector")
	ErrCreateTypedData        = errors.New("failed to create typed data for operation")
	ErrComputeOperationHash   = errors.New("failed to compute operation hash")
	ErrHashOperation          = errors.New("failed to hash operation")
	ErrSignOperation          = errors.New("failed to sign operation")
)

// Options defines the options for creating a new hasher client.
// The hasher client is used to compute EIP-712 hashes and signatures for operations
// without requiring network connectivity.
//   - Logger: Optional logger instance for logging messages.
type Options struct {
	Logger *slog.Logger
}

// Client provides operations for hashing and signing CREC operations.
// It handles EIP-712 typed data generation and signing without requiring
// network access or API client dependencies.
type Client struct {
	logger *slog.Logger
}

// NewClient creates a new hasher client with the provided options.
// The hasher client can compute operation hashes and signatures independently
// of network operations.
//   - opts: Options for configuring the hasher client. Can be nil for defaults.
func NewClient(opts *Options) (*Client, error) {
	var logger *slog.Logger
	if opts != nil && opts.Logger != nil {
		logger = opts.Logger
	} else {
		logger = slog.Default()
	}

	logger.Debug("Creating CREC hasher client")

	return &Client{
		logger: logger,
	}, nil
}

// HashOperation computes the EIP-712 digest of the given operation.
//   - op: The operation to hash.
//   - chainSelector: chainSelector of the blockchain network in which the operation is being executed.
//
// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
func (c *Client) HashOperation(op *types.Operation, chainSelector string) (common.Hash, error) {
	if op == nil {
		return common.Hash{}, ErrOperationRequired
	}

	chainSelectorUint, err := strconv.ParseUint(chainSelector, 10, 64)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrParseChainSelector, err)
	}
	chainFamily, err := chainselectors.GetSelectorFamily(chainSelectorUint)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrGetChainFamily, err)
	}
	if chainFamily != chainselectors.FamilyEVM {
		return common.Hash{}, fmt.Errorf("%w: %s", ErrUnsupportedChainFamily, chainFamily)
	}
	chainId, err := chainselectors.GetChainIDFromSelector(chainSelectorUint)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrGetChainID, err)
	}

	typedData, err := op.TypedData(chainId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrCreateTypedData, err)
	}
	hashBytes, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrComputeOperationHash, err)
	}
	hash := common.BytesToHash(hashBytes)
	return hash, nil
}

// SignOperation signs the given operation using the provided signer, returning the operation hash and the signature
// over the hash.
//   - ctx: The context for the request.
//   - op: The operation to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
//   - chainSelector: chainSelector of the blockchain network in which the operation is being executed.
//
// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
func (c *Client) SignOperation(
	ctx context.Context,
	op *types.Operation,
	signer signer.Signer,
	chainSelector string,
) (common.Hash, []byte, error) {
	if signer == nil {
		return common.Hash{}, nil, ErrSignerRequired
	}

	hash, err := c.HashOperation(op, chainSelector)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("%w: %w", ErrHashOperation, err)
	}
	sig, err := signer.Sign(ctx, hash.Bytes())
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("%w: %w", ErrSignOperation, err)
	}
	c.logger.Debug("Signed Operation",
		"chain_selector", chainSelector,
		"operation_id", op.ID.String(),
		"hash", hash.Hex(),
		"signature", common.Bytes2Hex(sig))
	return hash, sig, nil
}

// SignOperationHash signs the given operation hash using the provided signer, returning the signature.
//   - ctx: The context for the request.
//   - opHash: The operation hash to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
func (c *Client) SignOperationHash(
	ctx context.Context,
	opHash common.Hash,
	signer signer.Signer,
) ([]byte, error) {
	if signer == nil {
		return nil, ErrSignerRequired
	}

	sig, err := signer.Sign(ctx, opHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSignOperation, err)
	}
	c.logger.Debug("Signed Operation hash",
		"hash", opHash.Hex(),
		"signature", common.Bytes2Hex(sig))
	return sig, nil
}

// GetChainIDFromSelector is a utility function that extracts the chain ID from a chain selector string.
// This is useful for applications that need the numeric chain ID for other purposes.
//   - chainSelector: The chain selector string to parse.
//
// Returns the chain ID as a big.Int or an error if the selector is invalid or unsupported.
func GetChainIDFromSelector(chainSelector string) (*big.Int, error) {
	chainSelectorUint, err := strconv.ParseUint(chainSelector, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseChainSelector, err)
	}
	chainFamily, err := chainselectors.GetSelectorFamily(chainSelectorUint)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChainFamily, err)
	}
	if chainFamily != chainselectors.FamilyEVM {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedChainFamily, chainFamily)
	}
	chainIdStr, err := chainselectors.GetChainIDFromSelector(chainSelectorUint)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChainID, err)
	}
	chainId := new(big.Int)
	chainId, ok := chainId.SetString(chainIdStr, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse chain ID: %s", chainIdStr)
	}
	return chainId, nil
}
