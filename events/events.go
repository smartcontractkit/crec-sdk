package events

import (
	"context"
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
	ErrParseSignature               = errors.New("failed to parse signature")
	ErrRecoverPubKeyFromSignature   = errors.New("failed to recover public key from signature")
	ErrParseOCRReport               = errors.New("failed to parse OCR report")
	ErrParseOCRContext              = errors.New("failed to parse OCR context")
	ErrParseEventPayload            = errors.New("failed to parse event payload")
	ErrOnlyWatcherEventsSupported   = errors.New("only watcher events are supported for event verification")
	ErrOnlyOperationStatusSupported = errors.New("only operation status events are supported for operation status verification")
	ErrMarshalEventPayload          = errors.New("failed to marshal event payload")
	ErrMarshalEventToJSON           = errors.New("failed to marshal event to JSON")

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
func (c *Client) Poll(
	ctx context.Context, channelID uuid.UUID, params *apiClient.GetChannelsChannelIdEventsParams,
) ([]apiClient.Event, bool, error) {
	c.logger.Debug(
		"Polling events by channel",
		"channel_id", channelID.String(),
		"filter", params,
	)

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
		c.logger.Error(
			"Failed to get events - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body),
		)
		return nil, false, fmt.Errorf(
			"%w: %w (status code %d)", ErrPollEvents, ErrUnexpectedStatusCode, resp.StatusCode(),
		)
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrPollEvents, ErrNilResponseBody)
	}

	c.logger.Debug(
		"Events polled successfully",
		"count", len(resp.JSON200.Events),
		"has_more", resp.JSON200.HasMore,
	)

	return resp.JSON200.Events, resp.JSON200.HasMore, nil
}

// SearchEvents queries and searches historical events from a channel with filtering capabilities.
// Use this method for historical queries and searches. For real-time polling, use PollEvents.
//   - ctx: Context for the request, used for cancellation and timeouts.
//   - channelID: The UUID of the channel to search events from.
//   - params: Parameters for filtering events, see client.GetChannelsChannelIdEventsSearchParams for details.
func (c *Client) SearchEvents(
	ctx context.Context, channelID uuid.UUID, params *apiClient.GetChannelsChannelIdEventsSearchParams,
) ([]apiClient.Event, bool, error) {
	c.logger.Debug(
		"Searching historical events by channel",
		"channel_id", channelID.String(),
		"filter", params,
	)

	if channelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}

	resp, err := c.crecClient.GetChannelsChannelIdEventsSearchWithResponse(ctx, channelID, params)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSearchEvents, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn(
			"Channel not found",
			"channel_id", channelID.String(),
		)
		return nil, false, fmt.Errorf("%w (status code %d)", ErrChannelNotFound, resp.StatusCode())
	}

	if resp.StatusCode() == 400 {
		var errorMsg string
		if resp.JSON400 != nil && resp.JSON400.Message != "" {
			errorMsg = resp.JSON400.Message
		} else {
			errorMsg = "Invalid request parameters"
		}
		c.logger.Error(
			"Failed to search events - bad request",
			"status_code", resp.StatusCode(),
			"message", errorMsg,
			"body", string(resp.Body),
		)
		return nil, false, fmt.Errorf(
			"%w: %w: %s (status code %d)", ErrSearchEvents, ErrBadRequest, errorMsg, resp.StatusCode(),
		)
	}

	if resp.StatusCode() != 200 {
		c.logger.Error(
			"Failed to search events - unexpected status code",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body),
		)
		return nil, false, fmt.Errorf(
			"%w: %w (status code %d)", ErrSearchEvents, ErrUnexpectedStatusCode, resp.StatusCode(),
		)
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSearchEvents, ErrNilResponseBody)
	}

	c.logger.Debug(
		"Events searched successfully",
		"count", len(resp.JSON200.Events),
		"has_more", resp.JSON200.HasMore,
	)

	return resp.JSON200.Events, resp.JSON200.HasMore, nil
}

// Verify verifies the authenticity of a given event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
//   - workflowOwner: The expected workflow owner address (Ethereum address) that deployed the workflow. This is used to verify the event originated from a workflow owned by the expected address.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) Verify(event *apiClient.Event, workflowOwner string) (bool, error) {
	ocrProof, payload, err := c.prepareVerification(event)
	if err != nil {
		return false, err
	}

	// Check the event type in headers to ensure it's a watcher event
	if event.Headers.Type != apiClient.EventHeadersTypeWatcherEvent {
		return false, ErrOnlyWatcherEventsSupported
	}

	// Check the payload type to ensure it's a watcher event (defense in depth)
	eventPayload, err := payload.AsWatcherEventPayload()
	if err != nil {
		return false, ErrOnlyWatcherEventsSupported
	}

	ocrReport, ocrContext, err := c.parseOCRProofData(ocrProof)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	c.logger.Debug(
		"Verifying event",
		"event_hash", eventPayload.EventHash,
		"event_watcher_id", eventPayload.WatcherId,
		"ocr_report", ocrProof.OcrReport,
		"ocr_context", ocrProof.OcrContext,
	)

	// compute the event hash from the event data
	eventHash, err := c.EventHash(&eventPayload)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash, workflowOwner)
	if !eventHashValid {
		return false, ErrInvalidEventHash
	}

	// verify signatures
	verified, err := c.verifySignatures(ocrProof, ocrReport, ocrContext)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	if verified {
		c.logger.Debug("Event verified successfully")
	}
	return verified, nil
}

// VerifyOperationStatus verifies the authenticity of an operation status event.
// It checks whether the event was signed by at least a minimum number of authorized signers.
//   - event: The event to verify.
//   - workflowOwner: The expected workflow owner address (Ethereum address) that deployed the workflow. This is used to verify the event originated from a workflow owned by the expected address.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyOperationStatus(event *apiClient.Event, workflowOwner string) (bool, error) {
	ocrProof, payload, err := c.prepareVerification(event)
	if err != nil {
		return false, err
	}

	// Check the event type in headers to ensure it's an operation status event
	if event.Headers.Type != apiClient.EventHeadersTypeOperationStatus {
		return false, ErrOnlyOperationStatusSupported
	}

	// Check the payload type to ensure it's an operation status event (defense in depth)
	operationStatusPayload, err := payload.AsOperationStatusPayload()
	if err != nil {
		return false, ErrOnlyOperationStatusSupported
	}

	if operationStatusPayload.VerifiableEvent == nil || *operationStatusPayload.VerifiableEvent == "" {
		return false, fmt.Errorf("%w: verifiable event is required for operation status verification", ErrVerifyEvent)
	}

	ocrReport, ocrContext, err := c.parseOCRProofData(ocrProof)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	c.logger.Debug(
		"Verifying operation status event",
		"operation_id", operationStatusPayload.OperationId.String(),
		"wallet_operation_id", operationStatusPayload.WalletOperationId,
		"status", operationStatusPayload.Status,
		"ocr_report", ocrProof.OcrReport,
		"ocr_context", ocrProof.OcrContext,
	)

	// compute the event hash from the operation status payload
	eventHash, err := c.OperationStatusHash(&operationStatusPayload)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	// ensure locally computed event hash matches the one in the report
	eventHashValid := c.verifyEventHash(ocrReport, eventHash, workflowOwner)
	if !eventHashValid {
		return false, ErrInvalidEventHash
	}

	// verify signatures
	verified, err := c.verifySignatures(ocrProof, ocrReport, ocrContext)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	if verified {
		c.logger.Debug("Operation status event verified successfully")
	}
	return verified, nil
}

// VerifyOCRSignatures verifies that the OCR proof contains valid DON signatures.
// This is a lower-level verification that checks only signatures, without
// requiring a full Event structure or validating event hash/workflow owner.
//
// Returns true if at least minRequiredSignatures valid signatures are found.
func (c *Client) VerifyOCRSignatures(ocrReport, ocrContext string, signatures []string) (bool, error) {
	if len(c.validSigners) == 0 {
		return false, ErrVerificationNotConfigured
	}

	ocrProof := apiClient.OCRProof{
		OcrReport:  ocrReport,
		OcrContext: ocrContext,
		Signatures: signatures,
	}

	reportBytes, contextBytes, err := c.parseOCRProofData(ocrProof)
	if err != nil {
		return false, err
	}

	return c.verifySignatures(ocrProof, reportBytes, contextBytes)
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
	return crypto.Keccak256Hash([]byte(event.VerifiableEvent)), nil
}

// OperationStatusHash computes the "EventHash" of an OperationStatusPayload used for verification.
// The hash is computed using the pattern: eventName + "." + base64VerifiableEvent
// Note: VerifiableEvent must be present and non-empty (should be validated by caller).
func (c *Client) OperationStatusHash(payload *apiClient.OperationStatusPayload) (common.Hash, error) {
	if payload.VerifiableEvent == nil || *payload.VerifiableEvent == "" {
		return common.Hash{}, fmt.Errorf("%w: verifiable event is required for operation status verification", ErrVerifyEvent)
	}
	payloadToSign := *payload.VerifiableEvent
	eventHash := crypto.Keccak256Hash([]byte(payloadToSign))

	return eventHash, nil
}

// prepareVerification performs the initial verification setup steps:
// - Validates that verification is configured (validSigners are set)
// - Extracts the OCR proof from the event
// - Extracts the payload from the event
// Returns the OCR proof and payload, or an error if any step fails.
func (c *Client) prepareVerification(event *apiClient.Event) (apiClient.OCRProof, *apiClient.Event_Payload, error) {
	if len(c.validSigners) == 0 {
		return apiClient.OCRProof{}, nil, ErrVerificationNotConfigured
	}

	ocrProof, err := getOCRProof(event)
	if err != nil {
		return apiClient.OCRProof{}, nil, fmt.Errorf("%w: %w", ErrVerifyEvent, err)
	}

	return ocrProof, &event.Payload, nil
}

// parseOCRProofData parses the OCR proof and returns the OCR report and context bytes.
// It also validates that the OCR report has the minimum required length.
func (c *Client) parseOCRProofData(ocrProof apiClient.OCRProof) ([]byte, []byte, error) {
	ocrReport, err := common.ParseHexOrString(ocrProof.OcrReport)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParseOCRReport, err)
	}
	ocrContext, err := common.ParseHexOrString(ocrProof.OcrContext)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParseOCRContext, err)
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		return nil, nil, ErrOCRReportTooShort
	}

	return ocrReport, ocrContext, nil
}

// verifySignatures verifies that the OCR proof signatures are valid and signed by authorized signers.
// It returns true if at least minRequiredSignatures valid signatures are found, false otherwise.
func (c *Client) verifySignatures(ocrProof apiClient.OCRProof, ocrReport, ocrContext []byte) (bool, error) {
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
		c.logger.Debug(
			"Recovered signer from signature",
			"signer", signer.Hex(),
			"signature", sig,
		)

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
		c.logger.Debug(
			"Signatures verified successfully",
			"valid_signatures", validSigCount,
			"required_signatures", c.minRequiredSignatures,
		)
		return true, nil
	}
	c.logger.Warn(
		"Not enough valid signatures",
		"valid_signatures", validSigCount,
		"required_signatures", c.minRequiredSignatures,
	)
	return false, nil
}

func (c *Client) verifyEventHash(ocrReport []byte, eventHash common.Hash, workflowOwner string) bool {

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

	reportWorkflowOwner := common.BytesToAddress(ocrReport[87:107])
	expectedOwner := common.HexToAddress(workflowOwner)
	if reportWorkflowOwner != expectedOwner {
		c.logger.Warn(
			"Workflow owner mismatch",
			"report_workflow_owner", reportWorkflowOwner.Hex(),
			"expected_workflow_owner", expectedOwner.Hex(),
		)
		return false
	}

	reportEventHash := common.BytesToHash(ocrReport[ocrReportPayloadOffset:])
	c.logger.Debug(
		"Comparing event hashes",
		"local_event_hash", eventHash.String(),
		"report_event_hash", reportEventHash.String(),
	)

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
