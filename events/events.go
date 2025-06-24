package events

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
)

const (
	ocrReportPayloadOffset = 109 // Offset of the payload (event hash) in the OCR report
)

type ClientOptions struct {
	Logger                *zerolog.Logger
	EventsAfter           int64
	ValidSigners          []string
	MinRequiredSignatures int
}

type Client struct {
	cvnClient             *client.ClientWithResponses
	logger                *zerolog.Logger
	eventsAfter           int64
	minRequiredSignatures int
	validSigners          []string
	lastReadTimestamp     int64
	lastReadEventId       uuid.UUID
}

func NewClient(cvnClient *client.ClientWithResponses, opts *ClientOptions) (*Client, error) {
	if cvnClient == nil {
		return nil, errors.New("EventsClient requires a valid CVN client")
	}
	if opts == nil {
		return nil, errors.New("EventsClient requires a valid options struct")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN event reader")

	eventsAfter := opts.EventsAfter
	if eventsAfter == 0 {
		eventsAfter = time.Now().Unix()
	}

	return &Client{
		cvnClient:             cvnClient,
		logger:                logger,
		eventsAfter:           eventsAfter,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          opts.ValidSigners,

		lastReadTimestamp: 0,
		lastReadEventId:   uuid.Nil,
	}, nil
}

func (r *Client) Read(ctx context.Context) (*[]client.Event, error) {
	r.logger.Debug().Msg("Reading events from CVN")

	eventsAfter := r.lastReadTimestamp
	if eventsAfter == 0 {
		eventsAfter = r.eventsAfter
	}

	params := &client.GetEventsParams{
		CreatedGt: &eventsAfter,
	}
	if r.lastReadEventId != uuid.Nil {
		params.StartingAfter = &r.lastReadEventId
	}

	if r.logger.GetLevel() <= zerolog.DebugLevel {
		reqParams, err := json.Marshal(params)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to marshal request parameters")
			return nil, err
		}
		r.logger.Debug().
			RawJSON("params", reqParams).
			Msg("Reading events from CVN")
	}

	resp, err := r.cvnClient.GetEventsWithResponse(ctx, params)

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to read events from CVN")
		return nil, err
	}

	if resp.StatusCode() != 200 {
		r.logger.Error().Int("status", resp.StatusCode()).Msg("Failed to read events from CVN, unexpected status code")
		return nil, nil
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		r.logger.Warn().Msg("No events found in response from CVN")
		return nil, nil
	}

	var eventList = *resp.JSON200.Data
	if len(eventList) > 0 {
		r.lastReadTimestamp = *eventList[len(eventList)-1].CreatedAt
		r.lastReadEventId = *eventList[len(eventList)-1].EventId
	}

	return resp.JSON200.Data, nil
}

func (r *Client) Reset() {
	r.logger.Info().Msg("Resetting event reader state")
	r.lastReadTimestamp = 0
	r.lastReadEventId = uuid.Nil
}

func (r *Client) Verify(event *client.Event) (bool, error) {
	r.logger.Debug().
		Str("event_service", *event.Service).
		Str("event_name", *event.Name).
		Str("ocr_report", *event.OcrReport).
		Str("ocr_context", *event.OcrContext).
		Msg("Verifying event")

	ocrReport, err := common.ParseHexOrString(*event.OcrReport)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to parse report")
		return false, err
	}
	ocrContext, err := common.ParseHexOrString(*event.OcrContext)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to parse report context")
		return false, err
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		r.logger.Error().Msg("Report is too short")
		return false, err
	}

	// compute the event hash from the event data
	eventHash := r.EventHash(event)

	// ensure locally computed event hash matches the one in the report
	eventHashValid := r.VerifyEventHash(ocrReport, eventHash)
	if !eventHashValid {
		r.logger.Error().Err(err).Msg("Failed to verify event hash")
		return false, err
	}

	// generate the report hash matching the DON signing format
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)
	for _, signer := range r.validSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range *event.Signatures {
		sigBytes, err := common.ParseHexOrString(sig)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to parse signature")
			return false, err
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(reportHash.Bytes(), sigBytes)
		if err != nil {
			r.logger.Error().Err(err).Msg("Failed to recover public key from signature")
			return false, err
		}

		signer := crypto.PubkeyToAddress(*pubKey)
		r.logger.Debug().
			Str("signer", signer.Hex()).
			Str("signature", sig).
			Msg("Recovered signer from signature")

		if availableSigners[signer] {
			r.logger.Info().Str("signer", signer.Hex()).Msg("Signature verified successfully")
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}

		// If we have enough valid signatures, we can stop
		if validSigCount >= r.minRequiredSignatures {
			break
		}
	}
	r.logger.Debug().Int("valid_signatures", validSigCount).Msg("Finished signature checking")

	return validSigCount >= r.minRequiredSignatures, nil
}

func (r *Client) Decode(event *client.Event, payload any) error {
	jsonBytes, err := r.ToJson(event)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to convert verifiable event to JSON")
		return err
	}
	return json.Unmarshal(jsonBytes, payload)
}

func (r *Client) ToJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(*event.VerifiableEvent)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (r *Client) EventHash(event *client.Event) common.Hash {
	return crypto.Keccak256Hash([]byte(*event.Service + "." + *event.Name + "." + *event.VerifiableEvent))
}

func (r *Client) VerifyEventHash(ocrReport []byte, eventHash common.Hash) bool {

	// OCR report layout:
	// version                offset   0, size  1
	// workflow_execution_id  offset   1, size 32
	// timestamp              offset  33, size  4
	// don_id                 offset  37, size  4
	// don_config_version,    offset  41, size  4
	// workflow_cid           offset  45, size 32
	// workflow_name          offset  77, size 10
	// workflow_owner         offset  87, size 20
	// report_id              offset 107, size  2
	// payload			      offset 109, size  ... <- event hash

	reportEventHash := common.BytesToHash(ocrReport[ocrReportPayloadOffset:])
	r.logger.Debug().
		Str("local_event_hash", eventHash.String()).
		Str("report_event_hash", reportEventHash.String()).
		Msg("Comparing event hashes")

	return eventHash == reportEventHash
}
