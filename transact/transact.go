package transact

import (
	"context"
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/transact/signer"
	"github.com/smartcontractkit/cvn-sdk/transact/types"
)

type ClientOptions struct {
	Logger  *zerolog.Logger
	ChainID uint64
}

type Client struct {
	cvnClient *client.ClientWithResponses
	logger    *zerolog.Logger
	chainID   uint64
}

func NewClient(cvnClient *client.ClientWithResponses, opts *ClientOptions) (*Client, error) {
	if cvnClient == nil {
		return nil, errors.New("Client requires a valid CVN client")
	}
	if opts == nil {
		return nil, errors.New("Client requires a valid options struct")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN transact")

	return &Client{
		logger:    logger,
		cvnClient: cvnClient,
		chainID:   opts.ChainID,
	}, nil
}

func (t *Client) HashOperation(op *types.Operation) ([]byte, error) {
	typedData, err := op.TypedData(t.chainID)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create typed data for operation")
		return nil, err
	}
	hash, _, err := apitypes.TypedDataAndHash(*typedData)
	return hash, err
}

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

func (t *Client) SendSignedOperation(
	ctx context.Context,
	op *types.Operation,
	signature []byte,
) error {
	t.logger.Info().
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

	var requestData = client.OperationRequest{
		AccountOperationId: op.ID.String(),
		Transactions:       transactions,
		Account:            op.Account.String(),
		Signature:          "0x" + common.Bytes2Hex(signature),
	}

	resp, err := t.cvnClient.PostOperationSendWithResponse(ctx, requestData)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to send signed operation")
		return err
	}

	t.logger.Info().
		Int("status", resp.StatusCode()).
		Msg("SendSignedOperation result")

	return nil
}
