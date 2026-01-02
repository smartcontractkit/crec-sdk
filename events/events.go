package events

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

const (
	ocrReportPayloadOffset = 109 // Offset of the report payload (event hash) in the OCR report
)

var (
	// Validation errors
	ErrChannelIDRequired  = errors.New("channel_id is required")
	ErrOptionsRequired    = errors.New("options is required")
	ErrCRECClientRequired = errors.New("CRECClient is required")

	// API operation errors
	ErrChannelNotFound = errors.New("channel not found")
	ErrPollEvents      = errors.New("failed to poll events")
	ErrSearchEvents    = errors.New("failed to search events")
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
	ErrBadRequest           = errors.New("invalid request parameters")
	ErrOCRReportTooShort    = errors.New("OCR report is too short")

	// Configuration errors
	ErrVerificationNotConfigured = errors.New("event verification not configured: no valid signers")
)

// Options holds the configuration options for the CREC events client.
// It includes options for logging and event retrieval.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system (required).
//   - MinRequiredSignatures: Minimum number of valid signatures required to verify an event.
//   - ValidSigners: List of valid signer addresses (as hex strings).
type Options struct {
	Logger                *slog.Logger
	CRECClient            *apiClient.ClientWithResponses
	MinRequiredSignatures int
	ValidSigners          []string
}

// Client provides operations for polling and verifying events from CREC.
type Client struct {
	crecClient            *apiClient.ClientWithResponses
	logger                *slog.Logger
	minRequiredSignatures int
	validSigners          []string
}

// NewClient creates a new CREC events client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CREC client or options are nil, it returns an error.
//   - opts: Options for configuring the CREC events client, see Options for details.
func NewClient(opts *Options) (*Client, error) {
	if opts == nil {
		return nil, ErrOptionsRequired
	}
	if opts.CRECClient == nil {
		return nil, ErrCRECClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Debug("Creating CREC events client")

	return &Client{
		crecClient:            opts.CRECClient,
		logger:                logger,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          opts.ValidSigners,
	}, nil
}

// Poll retrieves events from the CREC service for a specific channel.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - channelID: The UUID of the channel to retrieve events from.
//   - params: parameters for filtering events, see apiClient.GetChannelsChannelIdEventsParams for details.
func (c *Client) Poll(ctx context.Context, channelID uuid.UUID, params *apiClient.GetChannelsChannelIdEventsParams) ([]apiClient.Event, bool, error) {
	c.logger.Debug("Polling events by channel",
		"channel_id", channelID.String(),
		"filter", params)

	if channelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}

	resp, err := c.crecClient.GetChannelsChannelIdEventsWithResponse(ctx, channelID, params)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrGetEvents, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found", "channel_id", channelID.String())
		return nil, false, fmt.Errorf("%w (status code %d)", ErrChannelNotFound, resp.StatusCode())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to get events - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrPollEvents, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrPollEvents, ErrNilResponseBody)
	}

	c.logger.Debug("Events polled successfully",
		"count", len(resp.JSON200.Events),
		"has_more", resp.JSON200.HasMore)

	return resp.JSON200.Events, resp.JSON200.HasMore, nil
}

// SearchEvents queries and searches historical events from a channel with filtering capabilities.
// Use this method for historical queries and searches. For real-time polling, use PollEvents.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - channelID: The UUID of the channel to search events from.
//   - params: Parameters for filtering events, see client.GetChannelsChannelIdEventsSearchParams for details.
func (c *Client) SearchEvents(ctx context.Context, channelID uuid.UUID, params *apiClient.GetChannelsChannelIdEventsSearchParams) ([]apiClient.Event, bool, error) {
	c.logger.Debug("Searching historical events by channel",
		"channel_id", channelID.String(),
		"filter", params)

	if channelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}

	resp, err := c.crecClient.GetChannelsChannelIdEventsSearchWithResponse(ctx, channelID, params)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSearchEvents, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found",
			"channel_id", channelID.String())
		return nil, false, fmt.Errorf("%w (status code %d)", ErrChannelNotFound, resp.StatusCode())
	}

	if resp.StatusCode() == 400 {
		var errorMsg string
		if resp.JSON400 != nil && resp.JSON400.Message != "" {
			errorMsg = resp.JSON400.Message
		} else {
			errorMsg = "Invalid request parameters"
		}
		c.logger.Error("Failed to search events - bad request",
			"status_code", resp.StatusCode(),
			"message", errorMsg,
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w: %s (status code %d)", ErrSearchEvents, ErrBadRequest, errorMsg, resp.StatusCode())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Failed to search events - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrSearchEvents, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSearchEvents, ErrNilResponseBody)
	}

	c.logger.Debug("Events searched successfully",
		"count", len(resp.JSON200.Events),
		"has_more", resp.JSON200.HasMore)

	return resp.JSON200.Events, resp.JSON200.HasMore, nil
}

// Verify verifies the authenticity of a given event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
//   - workflowId: The expected workflow CID (Content Identifier) that generated the event. This is the identifier of the workflow that should have generated this event.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) Verify(event *apiClient.Event, workflowId string) (bool, error) {
	if len(c.validSigners) == 0 {
		return false, ErrVerificationNotConfigured
	}
	ocrProof, err := getOCRProof(event)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// Check the payload type to ensure it's a watcher event
	payloadValue, err := event.Payload.ValueByDiscriminator()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrParseEventPayload, err)
	}

	eventPayload, ok := payloadValue.(apiClient.WatcherEventPayload)
	if !ok {
		return false, ErrOnlyWatcherEventsSupported
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
	c.logger.Debug("Verifying event",
		"event_address", eventPayload.Address,
		"event_watcher_id", eventPayload.WatcherId,
		"event_name", eventPayload.Name,
		"ocr_report", ocrProof.OcrReport,
		"ocr_context", ocrProof.OcrContext)

	// compute the event hash from the event data
	eventHash, err := c.EventHash(&eventPayload)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash, workflowId)
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
		c.logger.Debug("Recovered signer from signature",
			"signer", signer.Hex(),
			"signature", sig)

		if availableSigners[signer] {
			c.logger.Debug("Signature verified successfully", "signer", signer.Hex())
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}

		// If we have enough valid signatures, we can stop
		if validSigCount >= c.minRequiredSignatures {
			break
		}
	}

	if validSigCount >= c.minRequiredSignatures {
		c.logger.Debug("Event verified successfully",
			"valid_signatures", validSigCount,
			"required_signatures", c.minRequiredSignatures)
		return true, nil
	}
	c.logger.Warn("Not enough valid signatures",
		"valid_signatures", validSigCount,
		"required_signatures", c.minRequiredSignatures)
	return false, nil
}

// Decode decodes a verifiable event into a specified payload structure.
//   - event: The event to decode.
//   - payload: A pointer to the structure where the decoded event will be stored.
func (c *Client) Decode(event *apiClient.Event, payload any) error {
	// Marshal the event data to JSON
	jsonBytes, err := c.ToJSON(*event)
	if err != nil {
		return fmt.Errorf("%w: %w: %w", ErrDecodeEvent, ErrMarshalEventToJSON, err)
	}
	return json.Unmarshal(jsonBytes, payload)
}

// ToJSON converts a verifiable event into its JSON representation.
//   - event: The event to convert.
func (c *Client) ToJSON(event apiClient.Event) ([]byte, error) {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalEventPayload, err)
	}
	return jsonBytes, nil
}

// EventHash computes the "EventHash" of an event used for verification.
func (c *Client) EventHash(event *apiClient.WatcherEventPayload) (common.Hash, error) {
	dataBytes, err := json.Marshal(event.VerifiableEvent)
	if err != nil {
		return common.Hash{}, err
	}
	dataStr := base64.StdEncoding.EncodeToString(dataBytes)
	return crypto.Keccak256Hash([]byte(event.Domain + "." + event.Name + "." + dataStr)), nil
}

func (c *Client) verifyEventHash(ocrReport []byte, eventHash common.Hash, workflowId string) bool {

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

	reportWorkflowCid := common.BytesToHash(ocrReport[45:77])
	if reportWorkflowCid.String() != workflowId {
		c.logger.Warn("Workflow CID mismatch",
			"report_workflow_cid", reportWorkflowCid.String(),
			"workflow_id", workflowId)
		return false
	}

	reportEventHash := common.BytesToHash(ocrReport[ocrReportPayloadOffset:])
	c.logger.Debug("Comparing event hashes",
		"local_event_hash", eventHash.String(),
		"report_event_hash", reportEventHash.String())

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
