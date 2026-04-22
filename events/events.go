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

	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"
)

const (
	ocrReportPayloadOffset = 109 // Offset of the report payload (event hash) in the OCR report
	// CreMainlineTenantID is the CRE mainline tenant ID used as default for workflow owner address derivation.
	CreMainlineTenantID = "1"
)

var (
	// ErrChannelIDRequired is returned when the channel ID is nil or missing.
	ErrChannelIDRequired = errors.New("channel_id is required")
	// ErrOptionsRequired is returned when the options parameter is nil.
	ErrOptionsRequired = errors.New("options is required")
	// ErrCRECClientRequired is returned when the CREC client is nil in options.
	ErrCRECClientRequired = errors.New("CRECClient is required")

	// ErrChannelNotFound is returned when the channel does not exist (404 response).
	ErrChannelNotFound = errors.New("channel not found")
	// ErrPollEvents is returned when polling events fails.
	ErrPollEvents = errors.New("failed to poll events")
	// ErrSearchEvents is returned when searching events fails.
	ErrSearchEvents = errors.New("failed to search events")
	// ErrGetEvents is returned when fetching events fails.
	ErrGetEvents = errors.New("failed to get events")
	// ErrVerifyEvent is returned when event verification fails.
	ErrVerifyEvent = errors.New("failed to verify event")
	// ErrDecodeEvent is returned when decoding an event fails.
	ErrDecodeEvent = errors.New("failed to decode event")
	// ErrDecodeVerifiableEvent is returned when decoding the verifiable event payload fails.
	ErrDecodeVerifiableEvent = errors.New("failed to decode verifiable event")

	// ErrParseSignature is returned when parsing a signature from hex fails.
	ErrParseSignature = errors.New("failed to parse signature")
	// ErrRecoverPubKeyFromSignature is returned when recovering the public key from a signature fails.
	ErrRecoverPubKeyFromSignature = errors.New("failed to recover public key from signature")
	// ErrParseOCRReport is returned when parsing the OCR report from hex fails.
	ErrParseOCRReport = errors.New("failed to parse OCR report")
	// ErrParseOCRContext is returned when parsing the OCR context from hex fails.
	ErrParseOCRContext = errors.New("failed to parse OCR context")
	// ErrParseEventPayload is returned when parsing the event payload fails.
	ErrParseEventPayload = errors.New("failed to parse event payload")
	// ErrOnlyWatcherEventsSupported is returned when verifying a non-watcher event type.
	ErrOnlyWatcherEventsSupported = errors.New("only watcher events are supported for event verification")
	// ErrOnlyOperationStatusSupported is returned when verifying a non-operation-status event type.
	ErrOnlyOperationStatusSupported = errors.New("only operation status events are supported for operation status verification")
	// ErrMarshalEventPayload is returned when marshaling the event payload to JSON fails.
	ErrMarshalEventPayload = errors.New("failed to marshal event payload")
	// ErrMarshalEventToJSON is returned when marshaling the event to JSON fails.
	ErrMarshalEventToJSON = errors.New("failed to marshal event to JSON")

	// ErrInvalidEventHash is returned when the event hash verification fails.
	ErrInvalidEventHash = errors.New("event hash verification failed")

	// ErrNoOCRProofs is returned when the event has no OCR proofs.
	ErrNoOCRProofs = errors.New("no OCR proofs found")
	// ErrMultipleOCRProofs is returned when the event has more than one OCR proof.
	ErrMultipleOCRProofs = errors.New("multiple OCR proofs found but should be 1")

	// ErrUnexpectedStatusCode is returned when the API returns an unexpected HTTP status code.
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	// ErrNilResponse is returned when the API response is nil.
	ErrNilResponse     = errors.New("unexpected nil response")
	// ErrNilResponseBody is returned when the API response body is nil.
	ErrNilResponseBody = errors.New("unexpected nil response body")
	// ErrBadRequest is returned when the request parameters are invalid.
	ErrBadRequest = errors.New("invalid request parameters")
	// ErrOCRReportTooShort is returned when the OCR report is shorter than the minimum required length.
	ErrOCRReportTooShort = errors.New("OCR report is too short")

	// ErrVerificationNotConfigured is returned when the client has no valid signers configured.
	ErrVerificationNotConfigured = errors.New("event verification not configured: no valid signers")

	// ErrInvalidMinRequiredSignatures is returned when MinRequiredSignatures <= 0 but valid signers are provided.
	ErrInvalidMinRequiredSignatures = errors.New("MinRequiredSignatures must be greater than zero when valid signers are provided")
	// ErrInvalidSignerAddress is returned when a configured signer address is not a valid hex string.
	ErrInvalidSignerAddress = errors.New("invalid signer address")
	// ErrDuplicateSigner is returned when a valid signer address is provided multiple times.
	ErrDuplicateSigner = errors.New("duplicate valid signer configured")
	// ErrMinSignersExceedsUnique is returned when MinRequiredSignatures is greater than the number of unique valid signers.
	ErrMinSignersExceedsUnique = errors.New("MinRequiredSignatures exceeds the number of unique valid signers")

	// ErrDeriveWorkflowOwner is returned when deriving the workflow owner address from org ID fails.
	ErrDeriveWorkflowOwner = errors.New("failed to derive workflow owner from org ID")

	// ErrOrgIDRequired is returned when org ID is required for verification but not set.
	ErrOrgIDRequired = errors.New("org ID required for verification (set in client options or pass as parameter)")
	// ErrWorkflowOwnerRequired is returned when workflow owner is required for verification but not set.
	ErrWorkflowOwnerRequired = errors.New("workflow owner required for verification (set in client options or pass as parameter)")
	// ErrOrgIDOrWorkflowOwnerReq is returned when neither org ID nor workflow owner is configured for verification.
	ErrOrgIDOrWorkflowOwnerReq = errors.New("org ID or workflow owner required for verification (set in client options or pass as parameter)")

	// ErrNilWatcherEventPayload is returned when computing a watcher event hash with a nil payload.
	ErrNilWatcherEventPayload = errors.New("event payload is nil")
	// ErrVerifiableEventRequired is returned when VerifiableEvent is empty for watcher event hash.
	ErrVerifiableEventRequired = errors.New("verifiable event is required")
	// ErrNilVerifiablePayload is returned when a verifiable payload pointer is nil (hash or decode).
	ErrNilVerifiablePayload = errors.New("payload is nil")

	// ErrDecodeNilEvent is returned when Decode is called with a nil event.
	ErrDecodeNilEvent = errors.New("event is nil")
	// ErrDecodeNilEventID is returned when an event has no event ID.
	ErrDecodeNilEventID = errors.New("event ID is nil")
	// ErrDecodeNilEventProofs is returned when an event has no proofs.
	ErrDecodeNilEventProofs = errors.New("event proofs are nil")

	// ErrDecodeVerifiableEmpty is returned when a watcher payload has an empty verifiable event string.
	ErrDecodeVerifiableEmpty = errors.New("verifiable event is empty")
	// ErrDecodeVerifiableNilOrEmpty is returned when an operation status payload has no verifiable event.
	ErrDecodeVerifiableNilOrEmpty = errors.New("verifiable event is nil or empty")
	// ErrDecodeVerifiableInvalidBase64 is returned when verifiable event base64 decoding fails.
	ErrDecodeVerifiableInvalidBase64 = errors.New("invalid base64")
	// ErrDecodeVerifiableInvalidJSON is returned when verifiable event JSON unmarshaling fails.
	ErrDecodeVerifiableInvalidJSON = errors.New("invalid JSON")

	// ErrInvalidOCRSignatureLength is returned when a signature is not 65 bytes.
	ErrInvalidOCRSignatureLength = errors.New("signature length must be 65 bytes")
	// ErrInvalidOCRSignatureRecovery is returned when the secp256k1 recovery byte is not in {0,1,27,28}.
	ErrInvalidOCRSignatureRecovery = errors.New("invalid recovery byte")
)

// Options holds the configuration options for the CREC events client.
// It includes options for logging and event retrieval.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system (required).
//   - MinRequiredSignatures: Minimum number of valid signatures required to verify an event.
//   - ValidSigners: List of valid signer addresses (as hex strings).
//   - OrgID: Optional default organization ID for [Client.Verify] and [Client.VerifyOperationStatus].
//     When set, those methods can be called without passing org ID. For multi-org use, omit and use
//     [Client.VerifyWithOrgID] or [Client.VerifyOperationStatusWithOrgID] with explicit org ID.
//   - WorkflowOwner: Optional default workflow owner address for [Client.Verify] and
//     [Client.VerifyOperationStatus]. When set (and OrgID is not), those methods use it.
//     For explicit workflow owner per event, use [Client.VerifyWithWorkflowOwner] or
//     [Client.VerifyOperationStatusWithWorkflowOwner].
//   - CRETenantID: CRE tenant ID referring to different environments of CRE, used for workflow owner address derivation. Defaults to CreMainlineTenantID ("1") if not provided.
type Options struct {
	Logger                *slog.Logger
	CRECClient            *apiClient.ClientWithResponses
	MinRequiredSignatures int
	ValidSigners          []string
	OrgID                 string
	WorkflowOwner         string
	CRETenantID           string
}

// Client provides operations for polling and verifying events from CREC.
type Client struct {
	crecClient            *apiClient.ClientWithResponses
	logger                *slog.Logger
	minRequiredSignatures int
	validSigners          []string
	orgID                 string
	workflowOwner         string
	creTenantID           string
}

// NewClient creates a new CREC events client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CREC client or options are nil, it returns an error.
//   - opts: Options for configuring the CREC events client, see Options for details.
func NewClient(opts *Options) (*Client, error) {
	if opts == nil {
		return nil, ErrOptionsRequired
	}
	if len(opts.ValidSigners) > 0 && opts.MinRequiredSignatures <= 0 {
		return nil, ErrInvalidMinRequiredSignatures
	}
	if opts.CRECClient == nil {
		return nil, ErrCRECClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Debug("Creating CREC events client")

	seenSigners := make(map[string]bool)
	for _, signer := range opts.ValidSigners {
		if !common.IsHexAddress(signer) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidSignerAddress, signer)
		}
		addr := common.HexToAddress(signer).Hex()
		if seenSigners[addr] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateSigner, signer)
		}
		seenSigners[addr] = true
	}

	if len(seenSigners) > 0 && opts.MinRequiredSignatures > len(seenSigners) {
		return nil, fmt.Errorf("%w: requested %d but only have %d", ErrMinSignersExceedsUnique, opts.MinRequiredSignatures, len(seenSigners))
	}

	creTenantID := opts.CRETenantID
	if creTenantID == "" {
		creTenantID = CreMainlineTenantID
	}

	return &Client{
		crecClient:            opts.CRECClient,
		logger:                logger,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          opts.ValidSigners,
		orgID:                 opts.OrgID,
		workflowOwner:         opts.WorkflowOwner,
		creTenantID:           creTenantID,
	}, nil
}

// WorkflowOwnerFromOrgID derives the workflow owner Ethereum address from an
// organization ID. It uses the CRE canonical CREATE2-style address derivation
// with the client's configured CRE tenant ID.
func (c *Client) WorkflowOwnerFromOrgID(orgID string) (string, error) {
	addrBytes, err := workflows.GenerateWorkflowOwnerAddress(c.creTenantID, orgID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDeriveWorkflowOwner, err)
	}
	return common.BytesToAddress(addrBytes).Hex(), nil
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

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrGetEvents, ErrNilResponse)
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

	if resp == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSearchEvents, ErrNilResponse)
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

// Verify verifies the authenticity of a given watcher event using the default org ID or
// workflow owner configured on the client. Prefers OrgID when set (derives workflow owner);
// otherwise uses WorkflowOwner when set. Returns [ErrOrgIDOrWorkflowOwnerReq] if neither is configured.
//   - event: The event to verify.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) Verify(event *apiClient.Event) (bool, error) {
	if c.orgID != "" {
		return c.VerifyWithOrgID(event, c.orgID)
	}
	if c.workflowOwner != "" {
		return c.VerifyWithWorkflowOwner(event, c.workflowOwner)
	}
	return false, ErrOrgIDOrWorkflowOwnerReq
}

// VerifyWithOrgID verifies the authenticity of a given watcher event using an explicit org ID.
// It derives the expected workflow owner address from the org ID and delegates to the workflow
// owner verification. Use this for multi-org scenarios where a single client verifies events
// from different organizations.
//   - event: The event to verify.
//   - orgID: The organization ID used to derive the workflow owner address.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyWithOrgID(event *apiClient.Event, orgID string) (bool, error) {
	workflowOwner, err := c.WorkflowOwnerFromOrgID(orgID)
	if err != nil {
		return false, err
	}
	return c.VerifyWithWorkflowOwner(event, workflowOwner)
}

// VerifyWithWorkflowOwner verifies the authenticity of a given watcher event using an explicit
// workflow owner address. Use this for multi-org or when you have the address per event.
// For client-default verification, use [Client.Verify].
//   - event: The event to verify.
//   - workflowOwner: The workflow owner address (Ethereum address) that deployed the workflow.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyWithWorkflowOwner(event *apiClient.Event, workflowOwner string) (bool, error) {
	if workflowOwner == "" {
		return false, ErrWorkflowOwnerRequired
	}
	ocrProof, payload, err := c.prepareVerification(event)
	if err != nil {
		return false, err
	}

	// Check the event type in headers to ensure it's a watcher event
	if event.Headers.Type != apiClient.EventTypeWatcherEvent {
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

// VerifyOperationStatus verifies the authenticity of an operation status event using the default
// org ID or workflow owner configured on the client. Prefers OrgID when set; otherwise uses
// WorkflowOwner. Returns [ErrOrgIDOrWorkflowOwnerReq] if neither is configured.
//   - event: The event to verify.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyOperationStatus(event *apiClient.Event) (bool, error) {
	if c.orgID != "" {
		return c.VerifyOperationStatusWithOrgID(event, c.orgID)
	}
	if c.workflowOwner != "" {
		return c.VerifyOperationStatusWithWorkflowOwner(event, c.workflowOwner)
	}
	return false, ErrOrgIDOrWorkflowOwnerReq
}

// VerifyOperationStatusWithOrgID verifies the authenticity of an operation status event using
// an explicit org ID. It derives the expected workflow owner address and delegates to
// [Client.VerifyOperationStatusWithWorkflowOwner]. Use this for multi-org scenarios.
//   - event: The event to verify.
//   - orgID: The organization ID used to derive the workflow owner address.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyOperationStatusWithOrgID(event *apiClient.Event, orgID string) (bool, error) {
	workflowOwner, err := c.WorkflowOwnerFromOrgID(orgID)
	if err != nil {
		return false, err
	}
	return c.VerifyOperationStatusWithWorkflowOwner(event, workflowOwner)
}

// VerifyOperationStatusWithWorkflowOwner verifies the authenticity of an operation status event
// using an explicit workflow owner address. Use this for multi-org or per-event workflow owner.
// For client-default verification, use [Client.VerifyOperationStatus].
//   - event: The event to verify.
//   - workflowOwner: The workflow owner address (Ethereum address) that deployed the workflow.
//
// Returns true if the event is valid and signed by enough authorized signers, false otherwise.
func (c *Client) VerifyOperationStatusWithWorkflowOwner(event *apiClient.Event, workflowOwner string) (bool, error) {
	if workflowOwner == "" {
		return false, ErrWorkflowOwnerRequired
	}
	ocrProof, payload, err := c.prepareVerification(event)
	if err != nil {
		return false, err
	}

	// Check the event type in headers to ensure it's an operation status event
	if event.Headers.Type != apiClient.EventTypeOperationStatus {
		return false, ErrOnlyOperationStatusSupported
	}

	// Check the payload type to ensure it's an operation status event (defense in depth)
	operationStatusPayload, err := payload.AsOperationStatusPayload()
	if err != nil {
		return false, ErrOnlyOperationStatusSupported
	}

	if operationStatusPayload.VerifiableEvent == nil || *operationStatusPayload.VerifiableEvent == "" {
		return false, fmt.Errorf("%w: %w", ErrVerifyEvent, ErrVerifiableEventRequired)
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
	if event == nil {
		return fmt.Errorf("%w: %w", ErrDecodeEvent, ErrDecodeNilEvent)
	}
	if event.EventId == nil || *event.EventId == uuid.Nil {
		return fmt.Errorf("%w: %w", ErrDecodeEvent, ErrDecodeNilEventID)
	}
	if event.Headers.Proofs == nil {
		return fmt.Errorf("%w: %w", ErrDecodeEvent, ErrDecodeNilEventProofs)
	}

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
	if event.EventId == nil || *event.EventId == uuid.Nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalEventPayload, ErrDecodeNilEventID)
	}
	if event.Headers.Proofs == nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalEventPayload, ErrDecodeNilEventProofs)
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMarshalEventPayload, err)
	}
	return jsonBytes, nil
}

// EventHash computes the Keccak256 hash of the verifiable event string used for signature verification.
func (c *Client) EventHash(event *apiClient.WatcherEventPayload) (common.Hash, error) {
	if event == nil {
		return common.Hash{}, ErrNilWatcherEventPayload
	}
	if event.VerifiableEvent == "" {
		return common.Hash{}, ErrVerifiableEventRequired
	}
	return crypto.Keccak256Hash([]byte(event.VerifiableEvent)), nil
}

// OperationStatusHash computes the "EventHash" of an OperationStatusPayload used for verification.
// The hash is computed using the pattern: eventName + "." + base64VerifiableEvent
// Note: VerifiableEvent must be present and non-empty (should be validated by caller).
func (c *Client) OperationStatusHash(payload *apiClient.OperationStatusPayload) (common.Hash, error) {
	if payload == nil {
		return common.Hash{}, ErrNilVerifiablePayload
	}
	if payload.VerifiableEvent == nil || *payload.VerifiableEvent == "" {
		return common.Hash{}, ErrVerifiableEventRequired
	}
	payloadToSign := *payload.VerifiableEvent
	eventHash := crypto.Keccak256Hash([]byte(payloadToSign))

	return eventHash, nil
}

// DecodeVerifiableEvent decodes the base64-encoded VerifiableEvent from a WatcherEventPayload
// into a models.VerifiableEvent struct containing the full event data.
func (c *Client) DecodeVerifiableEvent(payload *apiClient.WatcherEventPayload) (*models.VerifiableEvent, error) {
	if payload == nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, ErrNilVerifiablePayload)
	}
	if payload.VerifiableEvent == "" {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, ErrDecodeVerifiableEmpty)
	}

	return c.decodeVerifiableEventString(payload.VerifiableEvent)
}

// DecodeOperationStatusVerifiableEvent decodes the base64-encoded VerifiableEvent
// from an OperationStatusPayload into a models.VerifiableEvent struct.
func (c *Client) DecodeOperationStatusVerifiableEvent(payload *apiClient.OperationStatusPayload) (*models.VerifiableEvent, error) {
	if payload == nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, ErrNilVerifiablePayload)
	}
	if payload.VerifiableEvent == nil || *payload.VerifiableEvent == "" {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, ErrDecodeVerifiableNilOrEmpty)
	}

	return c.decodeVerifiableEventString(*payload.VerifiableEvent)
}

// decodeVerifiableEventString decodes a base64-encoded verifiable event string
// into a models.VerifiableEvent struct.
func (c *Client) decodeVerifiableEventString(verifiableEventBase64 string) (*models.VerifiableEvent, error) {
	decoded, err := base64.StdEncoding.DecodeString(verifiableEventBase64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, fmt.Errorf("%w: %w", ErrDecodeVerifiableInvalidBase64, err))
	}

	var verifiableEvent models.VerifiableEvent
	if err := json.Unmarshal(decoded, &verifiableEvent); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableEvent, fmt.Errorf("%w: %w", ErrDecodeVerifiableInvalidJSON, err))
	}

	return &verifiableEvent, nil
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
		if len(sigBytes) != 65 {
			return false, fmt.Errorf("%w: %w", ErrParseSignature, ErrInvalidOCRSignatureLength)
		}

		v := sigBytes[64]
		if v != 0 && v != 1 && v != 27 && v != 28 {
			return false, fmt.Errorf("%w: %w: %d", ErrParseSignature, ErrInvalidOCRSignatureRecovery, v)
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
