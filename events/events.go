package events

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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

	// watcherEventPayloadDiscriminator is the discriminator value that the generated API client
	// uses for WatcherEventPayload types. The generated code's FromWatcherEventPayload() method
	// sets the Type field to this value (the Go type name, not the JSON type field value).
	watcherEventPayloadDiscriminator = "WatcherEventPayload"
)

var (
	// Validation errors
	ErrChannelIDRequired     = errors.New("channel_id is required")
	ErrClientOptionsRequired = errors.New("ClientOptions is required")
	ErrCRECClientRequired    = errors.New("a valid CRECClient must be provided")

	// API operation errors
	ErrChannelNotFound = errors.New("channel not found")
	ErrListEvents      = errors.New("failed to list events")
	ErrGetEvents       = errors.New("failed to get events")
	ErrVerifyEvent     = errors.New("failed to verify event")
	ErrDecodeEvent     = errors.New("failed to decode event")

	// Parsing errors
	ErrParseSignature             = errors.New("failed to parse signature")
	ErrRecoverPubKeyFromSignature = errors.New("failed to recover public key from signature")
	ErrParseOCRReport             = errors.New("failed to parse OCR report")
	ErrParseOCRContext            = errors.New("failed to parse OCR context")
	ErrParseEventPayload          = errors.New("failed to parse event payload")
	ErrOnlyWatcherEventsSupported = errors.New("only watcher events are supported for event verification")
	ErrMarshalEventPayload        = errors.New("failed to marshal event payload")
	ErrMarshalEventToJSON         = errors.New("failed to marshal event to JSON")

	// Verification errors
	ErrInvalidEventHash = errors.New("event hash verification failed")

	// OCR Proof errors
	ErrNoOCRProofs       = errors.New("no OCR proofs found")
	ErrMultipleOCRProofs = errors.New("multiple OCR proofs found but should be 1")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
	ErrOCRReportTooShort    = errors.New("OCR report is too short")
)

// ClientOptions holds the configuration options for the CREC events client.
// It includes options for logging and event retrieval.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system.
//   - EventsAfter: Unix timestamp to start retrieving events from. Defaults to current time if not set.
type ClientOptions struct {
	Logger                *zerolog.Logger
	CRECClient            *client.CRECClient
	EventsAfter           int64
	MinRequiredSignatures int
	ValidSigners          []string
}

type Client struct {
	crecClient            *client.CRECClient
	logger                *zerolog.Logger
	eventsAfter           int64
	minRequiredSignatures int
	validSigners          []string
	lastReadTimestamp     int64
	lastReadEventId       uuid.UUID
}

// NewClient creates a new CREC events client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CREC client or options are nil, it returns an error.
//   - opts: Options for configuring the CREC events client, see ClientOptions for details.
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, ErrClientOptionsRequired
	}
	if opts.CRECClient == nil {
		return nil, ErrCRECClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC events client")

	eventsAfter := opts.EventsAfter
	if eventsAfter == 0 {
		eventsAfter = time.Now().Unix()
	}

	return &Client{
		crecClient:            opts.CRECClient,
		logger:                logger,
		eventsAfter:           eventsAfter,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          opts.ValidSigners,
		lastReadTimestamp:     0,
		lastReadEventId:       uuid.Nil,
	}, nil
}

// ListEvents retrieves events from the CREC service for a specific channel.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - channelID: The UUID of the channel to retrieve events from.
//   - params: parameters for filtering events, see client.GetChannelsChannelIdEventsParams for details.
func (c *Client) ListEvents(ctx context.Context, channelID uuid.UUID, params *apiClient.GetChannelsChannelIdEventsParams) (*[]apiClient.Event, error) {
	c.logger.Debug().
		Str("channel_id", channelID.String()).
		Interface("filter", params).
		Msg("Listing events by channel")

	if channelID == uuid.Nil {
		return nil, ErrChannelIDRequired
	}

	resp, err := c.crecClient.GetChannelsChannelIdEventsWithResponse(ctx, channelID, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetEvents, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn().
			Str("channel_id", channelID.String()).
			Msg("Channel not found")
		return nil, fmt.Errorf("%w (status code %d)", ErrChannelNotFound, resp.StatusCode())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", string(resp.Body)).
			Msg("Failed to get events - unexpected status code")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrListEvents, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrListEvents, ErrNilResponseBody)
	}

	c.logger.Debug().
		Int("count", len(*resp.JSON200)).
		Msg("Events listed successfully")

	return resp.JSON200, nil
}

// Reset resets the internal state of the CREC events service.
// It clears the last read timestamp and event ID, allowing the service to start reading events from scratch.
func (c *Client) Reset() {
	c.logger.Debug().Msg("Resetting event reader state")
	c.lastReadTimestamp = 0
	c.lastReadEventId = uuid.Nil
}

// Verify verifies the authenticity of a given event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
func (c *Client) Verify(event *apiClient.Event) (bool, error) {
	ocrProof, err := getOCRProof(event)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// Check the payload discriminator type to ensure it's a watcher event
	discriminator, err := event.Payload.Discriminator()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrParseEventPayload, err)
	}
	if discriminator != watcherEventPayloadDiscriminator {
		return false, fmt.Errorf("%w (expected: %s, got: %s)", ErrOnlyWatcherEventsSupported, watcherEventPayloadDiscriminator, discriminator)
	}

	eventPayload, err := event.Payload.AsWatcherEventPayload()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrParseEventPayload, err)
	}

	ocrReport, err := common.ParseHexOrString(ocrProof.OcrReport)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrParseOCRReport, err)
	}
	ocrContext, err := common.ParseHexOrString(ocrProof.OcrContext)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrParseOCRContext, err)
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		return false, ErrOCRReportTooShort
	}

	c.logger.Trace().
		Str("event_address", eventPayload.Address).
		Str("event_watcher_id", eventPayload.WatcherId).
		Str("event_domain", *eventPayload.Event.Domain).
		Str("event_name", eventPayload.Event.EventName).
		Str("ocr_report", ocrProof.OcrReport).
		Str("ocr_context", ocrProof.OcrContext).
		Msg("Verifying event")

	// compute the event hash from the event data
	eventHash, err := c.EventHash(&eventPayload)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash)
	if !eventHashValid {
		return false, ErrInvalidEventHash
	}

	// generate the report hash matching the DON signing format
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)
	for _, signer := range c.validSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range ocrProof.Signatures {
		sigBytes, err := common.ParseHexOrString(sig)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrParseSignature, err)
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(reportHash.Bytes(), sigBytes)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrRecoverPubKeyFromSignature, err)
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

	if validSigCount >= c.minRequiredSignatures {
		c.logger.Debug().
			Int("valid_signatures", validSigCount).
			Int("required_signatures", c.minRequiredSignatures).
			Msg("Event verified successfully")
		return true, nil
	}
	c.logger.Warn().
		Int("valid_signatures", validSigCount).
		Int("required_signatures", c.minRequiredSignatures).
		Msg("Not enough valid signatures")
	return false, nil
}

// Decode decodes a verifiable event into a specified payload structure.
//   - event: The event to decode.
//   - payload: A pointer to the structure where the decoded event will be stored.
func (c *Client) Decode(event *apiClient.Event, payload any) error {
	// Marshal the event data to JSON
	jsonBytes, err := c.ToJson(*event)
	if err != nil {
		return fmt.Errorf("%w: %w: %w", ErrDecodeEvent, ErrMarshalEventToJSON, err)
	}
	return json.Unmarshal(jsonBytes, payload)
}

// ToJson converts a verifiable event into its JSON representation.
//   - event: The event to convert.
func (c *Client) ToJson(event apiClient.Event) ([]byte, error) {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalEventPayload, err)
	}
	return jsonBytes, nil
}

// EventHash computes the "EventHash" of an event used for verification.
func (c *Client) EventHash(event *apiClient.WatcherEventPayload) (common.Hash, error) {
	dataBytes, err := json.Marshal(event.Event.Data)
	if err != nil {
		return common.Hash{}, err
	}
	dataStr := base64.StdEncoding.EncodeToString(dataBytes)
	return crypto.Keccak256Hash([]byte(*event.Event.Domain + "." + event.Event.EventName + "." + dataStr)), nil
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

func getOCRProof(event *apiClient.Event) (apiClient.OCRProof, error) {
	var ocrProofs []apiClient.OCRProof
	for _, proofItem := range event.Headers.Proofs {
		ocrProof, err := proofItem.AsOCRProof()
		if err != nil {
			// Skip non-OCRProof types instead of failing
			continue
		}
		ocrProofs = append(ocrProofs, ocrProof)
	}
	switch {
	case len(ocrProofs) == 0:
		return apiClient.OCRProof{}, ErrNoOCRProofs
	case len(ocrProofs) > 1:
		return apiClient.OCRProof{}, ErrMultipleOCRProofs
	default:
		return ocrProofs[0], nil
	}
}
