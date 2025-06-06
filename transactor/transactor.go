package transactor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/chainlink/v2/core/capabilities/webapi/webapicap"
	"github.com/smartcontractkit/chainlink/v2/core/services/gateway/api"

	"github.com/smartcontractkit/cvn-sdk/transactor/signer"
	"github.com/smartcontractkit/cvn-sdk/transactor/types"
)

type Options struct {
	Logger            *zerolog.Logger
	TriggerPrivateKey string
	GatewayURL        string
	DonID             string
	ChainID           uint64
	KeystoneForwarder string
}

type Transactor struct {
	logger  *zerolog.Logger
	options *Options
}

func NewTransactor(opts *Options) *Transactor {
	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN transactor")

	// TODO: Validate options

	return &Transactor{
		logger:  logger,
		options: opts,
	}
}

func (t *Transactor) HashOperation(op *types.Operation) ([]byte, error) {
	typedData, err := op.TypedData(t.options.ChainID)
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create typed data for operation")
		return nil, err
	}
	hash, _, err := apitypes.TypedDataAndHash(*typedData)
	return hash, err
}

func (t *Transactor) SignOperation(
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

func (t *Transactor) SendSignedOperation(
	op *types.Operation,
	signature []byte,
) error {
	t.logger.Info().
		Str("operation_id", op.ID.String()).
		Str("signature", common.Bytes2Hex(signature)).
		Msg("Sending signed operation")

	key, err := crypto.HexToECDSA(t.options.TriggerPrivateKey)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error parsing private key")
		return err
	}

	var transactions []interface{}
	for _, tx := range op.Transactions {
		transactions = append(
			transactions, map[string]interface{}{
				"to":    tx.To.String(),
				"value": tx.Value.String(),
				"data":  "0x" + common.Bytes2Hex(tx.Data),
			},
		)
	}

	var triggerParams = map[string]interface{}{
		"operation_id": op.ID.String(),
		"transactions": transactions,
		"account":      op.Account,
		"signature":    "0x" + common.Bytes2Hex(signature),
	}

	payload := webapicap.TriggerRequestPayload{
		TriggerId:      "web-api-trigger@1.0.0",
		TriggerEventId: op.ID.String(),
		Timestamp:      time.Now().Unix(),
		Topics:         []string{"writeOperation"},
		Params:         triggerParams,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error marshalling trigger")
		return err
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)
	msg := &api.Message{
		Body: api.MessageBody{
			MessageId: payload.TriggerEventId,
			Method:    "web_api_trigger",
			DonId:     t.options.DonID,
			Payload:   json.RawMessage(payloadJson),
			Sender:    addr.String(),
		},
	}
	if err = msg.Sign(key); err != nil {
		t.logger.Error().Err(err).Msg("Error signing message")
		return err
	}
	codec := api.JsonRPCCodec{}
	rawMsg, err := codec.EncodeRequest(msg)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error JSON-RPC encoding")
		return err
	}
	t.logger.Debug().
		Str("request_url", t.options.GatewayURL).
		RawJSON("request_body", rawMsg).
		Msg("Trigger request")

	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		context.Background(), "POST", t.options.GatewayURL, bytes.NewBuffer(rawMsg),
	)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error creating request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error sending request")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.logger.Error().Err(err).Msg("Error reading response")
		return err
	}
	t.logger.Debug().
		Int("response_code", resp.StatusCode).
		RawJSON("response_body", body).
		Msg("Trigger response")

	return nil
}
