package transact

import (
	"context"
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
//   - ChainId: A string representing the chain ID of the blockchain network.
type ClientOptions struct {
	Logger  *zerolog.Logger
	ChainId string
}

type Client struct {
	cvnClient *client.ClientWithResponses
	logger    *zerolog.Logger
	chainId   string
}

// NewClient creates a new CVN transact client with the provided CVN client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - cvnClient: A valid CVN client instance.
//   - opts: Options for configuring the CVN transact client, see ClientOptions for details.
func NewClient(cvnClient *client.ClientWithResponses, opts *ClientOptions) (*Client, error) {
	if cvnClient == nil {
		return nil, errors.New("a valid CVN client must be provided")
	}
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
		cvnClient: cvnClient,
		chainId:   opts.ChainId,
	}, nil
}

// HashOperation computes the EIP-712 digest of the given operation.
//   - op: The operation to hash.
func (t *Client) HashOperation(op *types.Operation) ([]byte, error) {
	chainIdInt, err := strconv.ParseUint(t.chainId, 10, 64)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse chain ID")
		return nil, err
	}
	typedData, err := op.TypedData(chainIdInt)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create typed data for operation")
		return nil, err
	}
	hash, _, err := apitypes.TypedDataAndHash(*typedData)
	return hash, err
}

// SignOperation signs the given operation using the provided signer, returning the signature.
//   - op: The operation to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
func (t *Client) SignOperation(
	op *types.Operation,
	signer signer.Signer,
) ([]byte, error) {
	hash, err := t.HashOperation(op)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to hash operation for signing")
		return nil, err
	}
	sig, err := signer.Sign(hash)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to sign operation")
		return nil, err
	}
	t.logger.Debug().
		Str("operation_id", op.ID.String()).
		Str("hash", common.Bytes2Hex(hash)).
		Str("signature", common.Bytes2Hex(sig)).
		Msg("Signed Operation")
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
		Account:            op.Account.String(),
		Transactions:       transactions,
		Signature:          "0x" + common.Bytes2Hex(signature),
	}

	resp, err := t.cvnClient.PostOperationsWithResponse(ctx, requestData)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to send signed operation")
		return err
	}

	t.logger.Info().
		Int("status", resp.StatusCode()).
		Msg("SendSignedOperation result")

	return nil
}
