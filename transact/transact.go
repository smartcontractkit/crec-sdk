package transact

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
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
		return nil, errors.New("options must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN transact client")

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
		t.logger.Error().Err(err).Msg("Failed to parse chain ID")
		return common.Hash{}, err
	}
	typedData, err := op.TypedData(chainIdInt)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create typed data for operation")
		return common.Hash{}, err
	}
	hashBytes, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to compute operation hash")
		return common.Hash{}, err
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
		t.logger.Error().Err(err).Msg("Failed to hash operation for signing")
		return common.Hash{}, nil, err
	}
	sig, err := signer.Sign(ctx, hash.Bytes())
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to sign operation")
		return common.Hash{}, nil, err
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
		t.logger.Error().Err(err).Msg("Failed to sign operation")
		return nil, err
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
		return errors.New("no CVNClient provided, cannot send signed operations")
	}

	t.logger.Info().
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
		AccountAddress:     op.Account.String(), // FIX: field name must match API
		Transactions:       transactions,
		Signature:          "0x" + common.Bytes2Hex(signature),
	}

	if t.logger.GetLevel() <= zerolog.DebugLevel {
		data, err := json.MarshalIndent(requestData, "", "  ")
		if err != nil {
			t.logger.Err(err).Msg("Failed to marshal request data to JSON")
		} else {
			t.logger.Debug().
				Str("request_data", string(data)).
				Msg("Request data for SendSignedOperation")
		}
	}

	resp, err := t.cvnClient.PostOperationsWithResponse(ctx, requestData)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to send signed operation")
		return err
	}

	responseState := resp.HTTPResponse.StatusCode
	t.logger.Info().
		Int("status", responseState).
		Msg("SendSignedOperation result")

	if responseState != 201 {
		bodyBytes := resp.Body
		t.logger.Error().
			Int("status", responseState).
			Str("body", string(bodyBytes)).
			Msg("Failed to send signed operation, non-201 response")

		return errors.New("failed to send signed operation, non-201 response")
	}
	return nil
}
