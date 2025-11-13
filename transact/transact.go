package transact

import (
	"context"
	"strconv"

	// "encoding/json" // Commented out - not used after migration
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	chainselectors "github.com/smartcontractkit/chain-selectors"
	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/client"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

// ClientOptions defines the options for creating a new CREC transact client used to send operations to the CREC system.
// It includes a logger for logging messages and a chain ID for the blockchain network.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system, nil for no direct CREC interaction.
type ClientOptions struct {
	Logger     *zerolog.Logger
	CRECClient *client.CRECClient
}

type Client struct {
	logger     *zerolog.Logger
	crecClient *client.CRECClient
}

// NewClient creates a new CREC transact client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC transact client, see ClientOptions for details.
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("ClientOptions is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC transact client")

	return &Client{
		logger:     logger,
		crecClient: opts.CRECClient,
	}, nil
}

// HashOperation computes the EIP-712 digest of the given operation.
//   - op: The operation to hash.
//   - chainId: The chain ID of the blockchain network in which the operation is being executed.
func (t *Client) HashOperation(op *types.Operation, chainSelector string) (common.Hash, error) {
	// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
	chainSelectorUint, err := strconv.ParseUint(chainSelector, 10, 64)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to parse chain selector: %w", err)
	}
	chainFamily, err := chainselectors.GetSelectorFamily(chainSelectorUint)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain family: %w", err)
	}
	if chainFamily != chainselectors.FamilyEVM {
		return common.Hash{}, fmt.Errorf("chain family %s is not supported", chainFamily)
	}
	chainId, err := chainselectors.GetChainIDFromSelector(chainSelectorUint)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to compute EIP-712 digest of the given operation: %w", err)
	}

	typedData, err := op.TypedData(chainId)
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
//
// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
func (t *Client) SignOperation(
	ctx context.Context,
	op *types.Operation,
	signer signer.Signer,
	chainSelector string,
) (common.Hash, []byte, error) {
	hash, err := t.HashOperation(op, chainSelector)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to hash operation: %w", err)
	}
	sig, err := signer.Sign(ctx, hash.Bytes())
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to sign operation: %w", err)
	}
	t.logger.Debug().
		Str("chain_selector", chainSelector).
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

// SendSignedOperation sends a signed operation to the CREC system.
//   - ctx: The context for the request.
//   - op: The operation to send, which must be signed.
//   - signature: The signature of the operation, to be verified by the onchain smart account.
//   - chainId: The chain ID of the blockchain network in which the operation is being executed.
func (t *Client) SendSignedOperation(
	ctx context.Context,
	op *types.Operation,
	signature []byte,
	chainId string,
) (*apiClient.Operation, error) {
	if t.crecClient == nil {
		return nil, errors.New("no CRECClient provided, cannot send signed operations")
	}

	t.logger.Debug().
		Str("chain_id", chainId).
		Str("operation_id", op.ID.String()).
		Str("signature", common.Bytes2Hex(signature)).
		Msg("Sending signed operation")

	var transactions []apiClient.TransactionRequest
	for _, tx := range op.Transactions {
		transactions = append(
			transactions, apiClient.TransactionRequest{
				To:    tx.To.String(),
				Value: tx.Value.String(),
				Data:  "0x" + common.Bytes2Hex(tx.Data),
			},
		)
	}

	// COMMENTED OUT: This method needs to be updated to work with the new channels-based API
	// TODO: Update to use /channels/{channel_id}/operations endpoint and fix CreateOperation fields
	return nil, fmt.Errorf("SendSignedOperation is temporarily disabled - needs migration to channels-based API")

	// var requestData = apiClient.CreateOperation{
	// 	WalletOperationId: op.ID.String(),
	// 	ChainId:            chainId,
	// 	Address:     op.Account.String(),
	// 	Transactions:       transactions,
	// 	Signature:          "0x" + common.Bytes2Hex(signature),
	// }
	//
	// if t.logger.GetLevel() <= zerolog.TraceLevel {
	// 	data, err := json.MarshalIndent(requestData, "", "  ")
	// 	if err != nil {
	// 		t.logger.Err(err).Msg("Failed to marshal request data to JSON")
	// 	} else {
	// 		t.logger.Trace().
	// 			Str("request_data", string(data)).
	// 			Msg("Request data for SendSignedOperation")
	// 	}
	// }
	//
	// resp, err := t.crecClient.PostChannelsChannelIdOperationsWithResponse(ctx, channelId, requestData)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to send signed operation: %w", err)
	// }
	//
	// responseState := resp.HTTPResponse.StatusCode
	// t.logger.Debug().
	// 	Int("status", responseState).
	// 	Msg("SendSignedOperation result")
	//
	// if responseState != 201 {
	// 	return nil, fmt.Errorf(
	// 		"failed to send signed operation, non-201 response received: %s", resp.HTTPResponse.Status,
	// 	)
	// }
	//
	// t.logger.Trace().Str("raw_response", string(resp.Body)).Msg("OperationResponse JSON")
	//
	// return resp.JSON201, nil
}

// GetOperation retrieves an operation by its ID from the CREC service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - operationId: The UUID of the operation to retrieve.
//
// COMMENTED OUT: This method needs to be updated to work with the new channels-based API
// TODO: Update to use /channels/{channel_id}/operations/{operation_id} endpoint
func (t *Client) GetOperation(ctx context.Context, operationId uuid.UUID) (*apiClient.Operation, error) {
	return nil, fmt.Errorf("GetOperation is temporarily disabled - needs migration to channels-based API")
	// t.logger.Trace().Msg("Getting operation")
	//
	// resp, err := t.crecClient.GetChannelsChannelIdOperationsOperationIdWithResponse(ctx, channelId, operationId)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get operation id %v: %w", operationId, err)
	// }
	//
	// if resp.StatusCode() == 404 {
	// 	return nil, nil
	// } else if resp.StatusCode() != 200 {
	// 	return nil, fmt.Errorf("failed to get operation, unexpected status code: %s", resp.Status())
	// }
	//
	// return resp.JSON200, nil
}

// GetOperations retrieves a list of operations from the CREC service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//
// COMMENTED OUT: This method needs to be updated to work with the new channels-based API
// TODO: Update to use /channels/{channel_id}/operations endpoint
// Commenting out types because they don't exist in new API
func (t *Client) GetOperations(ctx context.Context, params interface{}) (
	[]apiClient.Operation, error,
) {
	return nil, fmt.Errorf("GetOperations is temporarily disabled - needs migration to channels-based API")
	// t.logger.Trace().Msg("Getting operations from CREC")
	//
	// resp, err := t.crecClient.GetChannelsChannelIdOperationsWithResponse(ctx, channelId, params)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get operations from CREC: %w", err)
	// }
	//
	// if resp.StatusCode() != 200 {
	// 	return nil, fmt.Errorf("failed to get operations from CREC, unexpected status code: %d", resp.StatusCode())
	// }
	//
	// if resp.JSON200 == nil {
	// 	return nil, fmt.Errorf("invalid operations response from CREC")
	// }
	//
	// return resp.JSON200.Data, nil
}

// ExecuteTransactions executes a list of transactions using the provided signer and executor account.
// It bundles the transactions into an operation and executes it.
//   - ctx: The context for the request.
//   - operationSigner: The signer to use for signing the operation.
//   - executorAccount: The account to use for executing the operation.
//   - txs: The transactions to execute.
//   - chainId: The chain ID of the blockchain network in which the transactions are being executed.
func (t *Client) ExecuteTransactions(
	ctx context.Context,
	operationSigner signer.Signer,
	executorAccount common.Address,
	txs []types.Transaction,
	chainId string,
) (*apiClient.Operation, error) {
	operation := &types.Operation{
		ID:           big.NewInt(time.Now().Unix()),
		Account:      executorAccount,
		Transactions: txs,
	}

	return t.ExecuteOperation(ctx, operationSigner, operation, chainId)
}

func (t *Client) ExecuteOperation(
	ctx context.Context, operationSigner signer.Signer, operation *types.Operation, chainId string,
) (*apiClient.Operation, error) {
	_, sig, err := t.SignOperation(ctx, operation, operationSigner, chainId)
	if err != nil {
		return nil, err
	}

	opr, err := t.SendSignedOperation(ctx, operation, sig, chainId)
	if err != nil {
		return nil, err
	}

	t.logger.Debug().
		Str("chainID", chainId).
		Str("operationID", operation.ID.String()).
		Str("account", operation.Account.Hex()).
		Msg("ExecuteOperation: operation sent successfully")
	return opr, nil
}
