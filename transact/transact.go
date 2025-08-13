package transact

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/transact/signer"
	"github.com/smartcontractkit/cvn-sdk/transact/types"
)

// ClientOptions defines the options for creating a new CVN transact client used to send operations to the CVN system.
// It includes a logger for logging messages and a chain ID for the blockchain network.
//   - Logger: Optional logger instance.
//   - CVNClient: A client instance for interacting with the CVN system, nil for no direct CVN interaction.
//   - ChainId: A string representing the chain ID of the blockchain network.
type ClientOptions struct {
	Logger    *zerolog.Logger
	CVNClient *client.ClientWithResponses
	ChainId   string
}

type Client struct {
	logger    *zerolog.Logger
	cvnClient *client.ClientWithResponses
	chainId   string
}

// NewClient creates a new CVN transact client with the provided CVN client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CVN transact client, see ClientOptions for details.
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("ClientOptions is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CVN transact client")

	return &Client{
		logger:    logger,
		cvnClient: opts.CVNClient,
		chainId:   opts.ChainId,
	}, nil
}

// HashOperation computes the EIP-712 digest of the given operation.
//   - op: The operation to hash.
func (t *Client) HashOperation(op *types.Operation) (common.Hash, error) {
	chainIdInt, err := strconv.ParseUint(t.chainId, 10, 64)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	typedData, err := op.TypedData(chainIdInt)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create typed data for operation: %w", err)
	}
	hashBytes, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to compute operation hash: %w", err)
	}
	hash := common.BytesToHash(hashBytes)
	return hash, err
}

// SignOperation signs the given operation using the provided signer, returning the operation hash and the signature
// over the hash.
//   - ctx: The context for the request.
//   - op: The operation to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
func (t *Client) SignOperation(
	ctx context.Context,
	op *types.Operation,
	signer signer.Signer,
) (common.Hash, []byte, error) {
	hash, err := t.HashOperation(op)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to hash operation: %w", err)
	}
	sig, err := signer.Sign(ctx, hash.Bytes())
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to sign operation: %w", err)
	}
	t.logger.Debug().
		Str("operation_id", op.ID.String()).
		Str("hash", hash.Hex()).
		Str("signature", common.Bytes2Hex(sig)).
		Msg("Signed Operation")
	return hash, sig, nil
}

// SignOperationHash signs the given operation hash using the provided signer, returning the signature.
//   - ctx: The context for the request.
//   - opHash: The operation hash to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
func (t *Client) SignOperationHash(
	ctx context.Context,
	opHash common.Hash,
	signer signer.Signer,
) ([]byte, error) {
	sig, err := signer.Sign(ctx, opHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign operation: %w", err)
	}
	t.logger.Debug().
		Str("hash", opHash.Hex()).
		Str("signature", common.Bytes2Hex(sig)).
		Msg("Signed Operation hash")
	return sig, nil
}

// SendSignedOperation sends a signed operation to the CVN system.
//   - ctx: The context for the request.
//   - op: The operation to send, which must be signed.
//   - signature: The signature of the operation, to be verified by the onchain smart account.
func (t *Client) SendSignedOperation(
	ctx context.Context,
	op *types.Operation,
	signature []byte,
) error {
	if t.cvnClient == nil {
		return fmt.Errorf("no CVNClient provided, cannot send signed operations")
	}

	t.logger.Debug().
		Str("chain_id", t.chainId).
		Str("operation_id", op.ID.String()).
		Str("signature", common.Bytes2Hex(signature)).
		Msg("Sending signed operation")

	var transactions []client.TransactionRequest
	for _, tx := range op.Transactions {
		transactions = append(
			transactions, client.TransactionRequest{
				To:    tx.To.String(),
				Value: tx.Value.String(),
				Data:  "0x" + common.Bytes2Hex(tx.Data),
			},
		)
	}

	var requestData = client.CreateOperation{
		AccountOperationId: op.ID.String(),
		ChainId:            t.chainId,
		Account:            op.Account.String(),
		Transactions:       transactions,
		Signature:          "0x" + common.Bytes2Hex(signature),
	}

	if t.logger.GetLevel() <= zerolog.TraceLevel {
		data, err := json.MarshalIndent(requestData, "", "  ")
		if err != nil {
			t.logger.Err(err).Msg("Failed to marshal request data to JSON")
		} else {
			t.logger.Trace().
				Str("request_data", string(data)).
				Msg("Request data for SendSignedOperation")
		}
	}

	resp, err := t.cvnClient.PostOperationsWithResponse(ctx, requestData)
	if err != nil {
		return fmt.Errorf("failed to send signed operation: %w", err)
	}

	responseState := resp.HTTPResponse.StatusCode
	t.logger.Debug().
		Int("status", responseState).
		Msg("SendSignedOperation result")

	if responseState != 201 {
		return fmt.Errorf("failed to send signed operation, non-201 response received: %s", resp.HTTPResponse.Status)
	}
	return nil
}

// GetOperation retrieves an operation by its ID from the CVN service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - operationId: The UUID of the operation to retrieve.
func (t *Client) GetOperation(ctx context.Context, operationId uuid.UUID) (*client.Operation, error) {
	t.logger.Trace().Msg("Getting operation")

	resp, err := t.cvnClient.GetOperationsOperationIdWithResponse(ctx, operationId)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation id %v: %w", operationId, err)
	}

	if resp.StatusCode() == 404 {
		return nil, nil
	} else if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get operation, unexpected status code: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// GetOperations retrieves a list of operations from the CVN service.
//   - ctx: Context for the request, used for cancellation and timeouts.
func (t *Client) GetOperations(ctx context.Context, params *client.GetOperationsParams) ([]client.Operation, error) {
	t.logger.Trace().Msg("Getting operations from CVN")

	resp, err := t.cvnClient.GetOperationsWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get operations from CVN: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get operations from CVN, unexpected status code: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("invalid operations response from CVN")
	}

	return resp.JSON200.Data, nil
}
