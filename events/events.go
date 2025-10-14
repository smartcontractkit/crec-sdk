package events

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/client"
)

const (
	ocrReportPayloadOffset = 109 // Offset of the report payload (event hash) in the OCR report
)

// ClientOptions holds the configuration options for the CREc events client.
// It includes options for logging, event retrieval, and signature verification.
//   - Logger: Optional logger instance.
//   - CREcClient: A client instance for interacting with the CREc system.
//   - EventsAfter: Unix timestamp to start retrieving events from. Defaults to current time if not set.
//   - MinRequiredSignatures: Minimum number of valid signatures required to verify an event.
//   - ValidSigners: List of signer addresses that are authorized verified event signers.
type ClientOptions struct {
	Logger                *zerolog.Logger
	CREcClient            *client.CREcClient
	EventsAfter           int64
	MinRequiredSignatures int
	ValidSigners          []string
}

type Client struct {
	crecClient            *client.CREcClient
	logger                *zerolog.Logger
	eventsAfter           int64
	minRequiredSignatures int
	validSigners          []string
	lastReadTimestamp     int64
	lastReadEventId       uuid.UUID
}

// NewClient creates a new CREc events client with the provided CREc client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CREc client or options are nil, it returns an error.
//   - crecClient: A valid CREc client instance.
//   - opts: Options for configuring the CREc events client, see ClientOptions for details.
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("ClientOptions is required")
	}
	if opts.CREcClient == nil {
		return nil, fmt.Errorf("a valid CREcClient must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREc events client")

	eventsAfter := opts.EventsAfter
	if eventsAfter == 0 {
		eventsAfter = time.Now().Unix()
	}

	return &Client{
		crecClient:            opts.CREcClient,
		logger:                logger,
		eventsAfter:           eventsAfter,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          opts.ValidSigners,

		lastReadTimestamp: 0,
		lastReadEventId:   uuid.Nil,
	}, nil
}

// GetEvents retrieves events from the CREc service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - params: parameters for filtering events, see client.GetEventsParams for details.
func (c *Client) GetEvents(ctx context.Context, params *apiClient.GetEventsParams) ([]apiClient.Event, error) {
	c.logger.Trace().Msg("Getting events from CREc")

	resp, err := c.crecClient.GetEventsWithResponse(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("failed to get events from CREc: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get events from CREc, unexpected status code: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("invalid events response from CREc")
	}

	return resp.JSON200.Data, nil
}

// Reset resets the internal state of the CREc events client.
// It clears the last read timestamp and event ID, allowing the client to start reading events from scratch.
func (c *Client) Reset() {
	c.logger.Debug().Msg("Resetting event reader state")
	c.lastReadTimestamp = 0
	c.lastReadEventId = uuid.Nil
}

// Verify verifies the authenticity of a given event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
func (c *Client) Verify(event *apiClient.Event) (bool, error) {
	c.logger.Trace().
		Str("event_service", event.Service).
		Str("event_name", event.Name).
		Str("ocr_report", event.OcrReport).
		Str("ocr_context", event.OcrContext).
		Msg("Verifying event")

	ocrReport, err := common.ParseHexOrString(event.OcrReport)
	if err != nil {
		return false, fmt.Errorf("failed to parse OCR report: %w", err)
	}
	ocrContext, err := common.ParseHexOrString(event.OcrContext)
	if err != nil {
		return false, fmt.Errorf("failed to parse OCR context: %w", err)
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		return false, fmt.Errorf("OCR report is too short")
	}

	// compute the event hash from the event data
	eventHash := c.EventHash(event)

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash)
	if !eventHashValid {
		return false, fmt.Errorf("failed to verify event hash")
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
			return false, fmt.Errorf("failed to parse signature: %w", err)
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(reportHash.Bytes(), sigBytes)
		if err != nil {
			return false, fmt.Errorf("failed to recover public key from signature")
		}

		signer := crypto.PubkeyToAddress(*pubKey)
		c.logger.Trace().
			Str("signer", signer.Hex()).
			Str("signature", sig).
			Msg("Recovered signer from signature")

		if availableSigners[signer] {
			c.logger.Debug().Str("signer", signer.Hex()).Msg("Signature verified successfully")
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
func (c *Client) Decode(event *apiClient.Event, payload any) error {
	jsonBytes, err := c.ToJson(event)
	if err != nil {
		return fmt.Errorf("failed to convert verifiable event to JSON: %w", err)
	}
	return json.Unmarshal(jsonBytes, payload)
}

// ToJson converts a verifiable event into its JSON representation.
//   - event: The event to convert.
func (c *Client) ToJson(event *apiClient.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decode base64 payload: %w", err)
	}
	return decodedStr, nil
}

// EventHash computes the "EventHash" of an event used for verification.
func (c *Client) EventHash(event *apiClient.Event) common.Hash {
	return crypto.Keccak256Hash([]byte(event.Service + "." + event.Name + "." + event.VerifiableEvent))
}

// CreateListener creates a new listener for events in the CREc service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listener: The listener to create. See client.CreateListener for details on required fields.
func (c *Client) CreateListener(ctx context.Context, listener *apiClient.CreateListener) (*apiClient.Listener, error) {
	c.logger.Debug().Msg("Creating listener on CREc")

	resp, err := c.crecClient.PostListenersWithResponse(ctx, *listener)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	if resp.StatusCode() != 201 {
		return nil, fmt.Errorf("failed to create listener, unexpected status code: %s", resp.Status())
	}

	return resp.JSON201, nil
}

// GetListener retrieves a listener by its ID from the CREc service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listenerId: The UUID of the listener to retrieve.
func (c *Client) GetListener(ctx context.Context, listenerId uuid.UUID) (*apiClient.Listener, error) {
	c.logger.Trace().Msg("Getting listener")

	resp, err := c.crecClient.GetListenersListenerIdWithResponse(ctx, listenerId)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get listener")
		return nil, err
	}

	if resp.StatusCode() == 404 {
		return nil, nil
	} else if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get listener, unexpected status code: %s", resp.Status())
	}

	return resp.JSON200, nil
}

// GetListeners retrieves a list of listeners from the CREc service.
//   - ctx: Context for the request, used for cancellation and timeouts.
func (c *Client) GetListeners(ctx context.Context, params *apiClient.GetListenersParams) ([]apiClient.Listener, error) {
	c.logger.Debug().Msg("Getting listeners from CREc")

	resp, err := c.crecClient.GetListenersWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get listeners from CREc: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get listeners from CREc, unexpected status code: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("invalid listeners response from CREc")
	}

	return resp.JSON200.Data, nil
}

// DeleteListener deletes a listener by its ID from the CREc service.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - listenerId: The UUID of the listener to delete.
func (c *Client) DeleteListener(ctx context.Context, listenerId uuid.UUID) error {
	c.logger.Debug().
		Str("listener_id", listenerId.String()).
		Msg("Deleting listener")

	resp, err := c.crecClient.DeleteListenersListenerIdWithResponse(ctx, listenerId)
	if err != nil {
		return fmt.Errorf("failed to delete listener: %w", err)
	}

	if resp.StatusCode() != 202 {
		return fmt.Errorf("failed to delete listener, unexpected status code: %s", resp.Status())
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
	c.logger.Trace().
		Str("local_event_hash", eventHash.String()).
		Str("report_event_hash", reportEventHash.String()).
		Msg("Comparing event hashes")

	return eventHash == reportEventHash
}
