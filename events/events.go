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
	ocrReportPayloadOffset = 109 // Offset of the report payload (event hash) in the OCR report
)

// ClientOptions holds the configuration options for the CVN events client.
// It includes options for logging, event retrieval, and signature verification.
//   - Logger: Optional logger instance.
//   - EventsAfter: Unix timestamp to start retrieving events from. Defaults to current time if not set.
//   - MinRequiredSignatures: Minimum number of valid signatures required to verify an event.
//   - ValidSigners: List of signer addresses that are authorized verified event signers.
type ClientOptions struct {
	Logger                *zerolog.Logger
	EventsAfter           int64
	MinRequiredSignatures int
	ValidSigners          []string
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

// NewClient creates a new CVN events client with the provided CVN client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CVN client or options are nil, it returns an error.
//   - cvnClient: A valid CVN client instance.
//   - opts: Options for configuring the CVN events client, see ClientOptions for details.
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

	logger.Info().Msg("Creating CVN events client")

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

// GetEvents retrieves events from the CVN service.
// It returns a slice of events that were created after the last read timestamp or the configured eventsAfter timestamp.
// If no events are found, it returns an empty slice.
//   - ctx: Context for the request, used for cancellation and timeouts.
func (c *Client) GetEvents(ctx context.Context) (*[]client.Event, error) {
	c.logger.Debug().Msg("Getting events from CVN")

	eventsAfter := c.lastReadTimestamp
	if eventsAfter == 0 {
		eventsAfter = c.eventsAfter
	}

	params := &client.GetEventsParams{
		CreatedGt: &eventsAfter,
	}
	if c.lastReadEventId != uuid.Nil {
		params.StartingAfter = &c.lastReadEventId
	}

	if c.logger.GetLevel() <= zerolog.DebugLevel {
		reqParams, err := json.Marshal(params)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to marshal request parameters")
			return nil, err
		}
		c.logger.Debug().
			RawJSON("params", reqParams).
			Msg("Getting events from CVN")
	}

	resp, err := c.cvnClient.GetEventsWithResponse(ctx, params)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get events from CVN")
		return nil, err
	}

	if resp.StatusCode() != 200 {
		c.logger.Error().Int("status", resp.StatusCode()).Msg("Failed to get events from CVN, unexpected status code")
		return nil, nil
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		c.logger.Warn().Msg("No events found in response from CVN")
		return nil, nil
	}

	var eventList = resp.JSON200.Data
	if len(eventList) > 0 {
		c.lastReadTimestamp = eventList[len(eventList)-1].CreatedAt
		c.lastReadEventId = eventList[len(eventList)-1].EventId
	}

	return &resp.JSON200.Data, nil
}

// Reset resets the internal state of the CVN events client.
// It clears the last read timestamp and event ID, allowing the client to start reading events from scratch.
func (c *Client) Reset() {
	c.logger.Info().Msg("Resetting event reader state")
	c.lastReadTimestamp = 0
	c.lastReadEventId = uuid.Nil
}

// Verify verifies the authenticity of a given event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
func (c *Client) Verify(event *client.Event) (bool, error) {
	c.logger.Debug().
		Str("event_service", event.Service).
		Str("event_name", event.Name).
		Str("ocr_report", event.OcrReport).
		Str("ocr_context", event.OcrContext).
		Msg("Verifying event")

	ocrReport, err := common.ParseHexOrString(event.OcrReport)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to parse OCR report")
		return false, err
	}
	ocrContext, err := common.ParseHexOrString(event.OcrContext)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to parse OCR context")
		return false, err
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		c.logger.Error().Msg("OCR report is too short")
		return false, err
	}

	// compute the event hash from the event data
	eventHash := c.EventHash(event)

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash)
	if !eventHashValid {
		c.logger.Error().Err(err).Msg("Failed to verify event hash")
		return false, err
	}

	// generate the report hash matching the DON signing format
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)
	for _, signer := range c.validSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range event.Signatures {
		sigBytes, err := common.ParseHexOrString(sig)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to parse signature")
			return false, err
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(reportHash.Bytes(), sigBytes)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to recover public key from signature")
			return false, err
		}

		signer := crypto.PubkeyToAddress(*pubKey)
		c.logger.Debug().
			Str("signer", signer.Hex()).
			Str("signature", sig).
			Msg("Recovered signer from signature")

		if availableSigners[signer] {
			c.logger.Info().Str("signer", signer.Hex()).Msg("Signature verified successfully")
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}

		// If we have enough valid signatures, we can stop
		if validSigCount >= c.minRequiredSignatures {
			break
		}
	}
	c.logger.Debug().Int("valid_signatures", validSigCount).Msg("Finished signature checking")

	return validSigCount >= c.minRequiredSignatures, nil
}

// Decode decodes a verifiable event into a specified payload structure.
//   - event: The event to decode.
//   - payload: A pointer to the structure where the decoded event will be stored.
func (c *Client) Decode(event *client.Event, payload any) error {
	jsonBytes, err := c.ToJson(event)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to convert verifiable event to JSON")
		return err
	}
	return json.Unmarshal(jsonBytes, payload)
}

// ToJson converts a verifiable event into its JSON representation.
//   - event: The event to convert.
func (c *Client) ToJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

// EventHash computes the "EventHash" of an event used for verification.
func (c *Client) EventHash(event *client.Event) common.Hash {
	return crypto.Keccak256Hash([]byte(event.Service + "." + event.Name + "." + event.VerifiableEvent))
}

// CreateListener creates a new listener for events in the CVN service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listener: The listener to create. See client.CreateListener for details on required fields.
func (c *Client) CreateListener(ctx context.Context, listener *client.CreateListener) (*client.Listener, error) {
	c.logger.Debug().
		Str("listener_name", listener.Name).
		Str("listener_service", listener.Service).
		Str("listener_chain_id", listener.ChainId).
		Msg("Creating listener")

	resp, err := c.cvnClient.PostListenersWithResponse(ctx, *listener)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to create listener")
		return nil, err
	}

	if resp.StatusCode() != 201 {
		c.logger.Error().Err(err).Msg("Failed to create listener")
		return nil, errors.New("failed to create listener, unexpected status code: " + resp.Status())
	}

	return resp.JSON201, nil
}

// GetListener retrieves a listener by its ID from the CVN service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listenerId: The UUID of the listener to retrieve.
func (c *Client) GetListener(ctx context.Context, listenerId uuid.UUID) (*client.Listener, error) {
	c.logger.Debug().
		Str("listener_id", listenerId.String()).
		Msg("Getting listener")

	resp, err := c.cvnClient.GetListenersListenerIdWithResponse(ctx, listenerId)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get listener")
		return nil, err
	}

	if resp.StatusCode() == 404 {
		c.logger.Debug().
			Str("listener_id", listenerId.String()).
			Msg("Listener not found")
		return nil, nil
	} else if resp.StatusCode() != 200 {
		c.logger.Error().Err(err).Msg("Failed to get listener")
		return nil, errors.New("failed to get listener, unexpected status code: " + resp.Status())
	}

	return resp.JSON200, nil
}

// DeleteListener deletes a listener by its ID from the CVN service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listenerId: The UUID of the listener to delete.
func (c *Client) DeleteListener(ctx context.Context, listenerId uuid.UUID) error {
	c.logger.Debug().
		Str("listener_id", listenerId.String()).
		Msg("Deleting listener")

	resp, err := c.cvnClient.DeleteListenersListenerIdWithResponse(ctx, listenerId)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to delete listener")
		return err
	}

	if resp.StatusCode() != 202 {
		c.logger.Error().Err(err).Msg("Failed to delete listener")
		return errors.New("failed to delete listener, unexpected status code: " + resp.Status())
	}

	return nil
}

func (c *Client) verifyEventHash(ocrReport []byte, eventHash common.Hash) bool {

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
	c.logger.Debug().
		Str("local_event_hash", eventHash.String()).
		Str("report_event_hash", reportEventHash.String()).
		Msg("Comparing event hashes")

	return eventHash == reportEventHash
}
