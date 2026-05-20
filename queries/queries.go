package queries

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/models"
)

const (
	defaultPollInterval = 2 * time.Second
	zeroAddress        = "0x0000000000000000000000000000000000000000"
)

var statusCodePattern = regexp.MustCompile(`status code:?\s*(\d{3})`)

var (
	// Client initialization errors.
	ErrOptionsRequired   = errors.New("options is required")
	ErrAPIClientRequired = errors.New("APIClient is required")

	// Input validation errors.
	ErrChannelIDRequired        = errors.New("channel_id is required")
	ErrQueryIDRequired          = errors.New("query_id is required")
	ErrIdempotencyKeyRequired   = errors.New("idempotency_key is required")
	ErrQueryKindRequired        = errors.New("query_kind is required")
	ErrUnsupportedQueryKind     = errors.New("unsupported query_kind")
	ErrChainSelectorRequired    = errors.New("chain_selector is required")
	ErrContractAddressRequired  = errors.New("contract_address is required")
	ErrInvalidContractAddress   = errors.New("contract_address must be a valid hex address")
	ErrInvalidFromAddress       = errors.New("from_address must be a valid hex address")
	ErrCallDataRequired         = errors.New("call_data is required")
	ErrInvalidCallData          = errors.New("call_data must be 0x-prefixed even-length hex bytes")
	ErrBlockSelectionRequired   = errors.New("block_selection is required")
	ErrInvalidBlockSelection    = errors.New("invalid block_selection")
	ErrInvalidCallOption        = errors.New("invalid call option")
	ErrInvalidLimit             = errors.New("limit must be between 1 and 100")
	ErrInvalidOffset            = errors.New("offset cannot be negative")
	ErrQueryRequired            = errors.New("query is required")
	ErrVerifiableResultRequired = errors.New("verifiable_result is required")
	ErrInvalidHexBytes          = errors.New("invalid 0x-prefixed even-length hex bytes")

	// API/resource errors.
	ErrChannelNotFound      = errors.New("channel not found")
	ErrQueryNotFound        = errors.New("query not found")
	ErrIdempotencyConflict  = errors.New("idempotency conflict")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrCreateQuery          = errors.New("failed to create query")
	ErrGetQuery             = errors.New("failed to get query")
	ErrListQueries          = errors.New("failed to list queries")
	ErrWaitQuery            = errors.New("failed waiting for query")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponse          = errors.New("unexpected nil response")
	ErrNilResponseBody      = errors.New("unexpected nil response body")

	// Decoding/result errors.
	ErrDecodeVerifiableResult       = errors.New("failed to decode verifiable_result")
	ErrInvalidVerifiableResultBase64 = errors.New("invalid verifiable_result base64")
	ErrInvalidVerifiableResultJSON   = errors.New("invalid verifiable_result JSON")
	ErrBuildCallContractResult       = errors.New("failed to build call contract result")

	// ABI helper errors.
	ErrABIRequired             = errors.New("abi fragment is required")
	ErrParseABI                = errors.New("failed to parse abi fragment")
	ErrABIFunctionNameRequired = errors.New("function_name is required")
	ErrABIFunctionNotFound     = errors.New("function not found in abi")
	ErrABIArgumentCount        = errors.New("abi argument count mismatch")
	ErrABIArgumentType         = errors.New("invalid abi argument type")
	ErrDecodeABIOutput         = errors.New("failed to decode abi output")
)

// BlockSelection is the generated API union for query block selectors.
type BlockSelection = apiClient.QueryBlockSelection

// OCRProof is the generated API OCR proof type.
type OCRProof = apiClient.OCRProof

// Options defines options for creating a CREC Queries client.
type Options struct {
	Logger       *slog.Logger
	APIClient    *apiClient.ClientWithResponses
	PollInterval time.Duration
}

// Client provides channel-scoped chain query operations.
type Client struct {
	logger       *slog.Logger
	apiClient    *apiClient.ClientWithResponses
	pollInterval time.Duration
}

// NewClient creates a new Queries client.
func NewClient(opts *Options) (*Client, error) {
	if opts == nil {
		return nil, ErrOptionsRequired
	}
	if opts.APIClient == nil {
		return nil, ErrAPIClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	pollInterval := opts.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}

	logger.Debug("Creating CREC Queries client", "poll_interval", pollInterval)

	return &Client{
		logger:       logger,
		apiClient:    opts.APIClient,
		pollInterval: pollInterval,
	}, nil
}

// LatestBlockSelection returns an explicit latest block selector.
func LatestBlockSelection() BlockSelection {
	var selection BlockSelection
	_ = selection.FromLatestBlockSelection(apiClient.LatestBlockSelection{})
	return selection
}

// Latest is a short alias for LatestBlockSelection.
func Latest() BlockSelection {
	return LatestBlockSelection()
}

// FinalizedBlockSelection returns an explicit finalized block selector.
func FinalizedBlockSelection() BlockSelection {
	var selection BlockSelection
	_ = selection.FromFinalizedBlockSelection(apiClient.FinalizedBlockSelection{})
	return selection
}

// Finalized is a short alias for FinalizedBlockSelection.
func Finalized() BlockSelection {
	return FinalizedBlockSelection()
}

// BlockNumberBlockSelection returns an explicit block_number selector.
func BlockNumberBlockSelection(blockNumber uint64) BlockSelection {
	selection, _ := BlockNumberBlockSelectionString(strconv.FormatUint(blockNumber, 10))
	return selection
}

// BlockNumber is a short alias for BlockNumberBlockSelection.
func BlockNumber(blockNumber uint64) BlockSelection {
	return BlockNumberBlockSelection(blockNumber)
}

// BlockNumberBlockSelectionString returns an explicit block_number selector from a decimal uint64 string.
func BlockNumberBlockSelectionString(blockNumber string) (BlockSelection, error) {
	if blockNumber == "" {
		return BlockSelection{}, ErrInvalidBlockSelection
	}
	if _, err := strconv.ParseUint(blockNumber, 10, 64); err != nil {
		return BlockSelection{}, fmt.Errorf("%w: %w", ErrInvalidBlockSelection, err)
	}

	var selection BlockSelection
	if err := selection.FromBlockNumberBlockSelection(apiClient.BlockNumberBlockSelection{
		BlockNumber: blockNumber,
	}); err != nil {
		return BlockSelection{}, fmt.Errorf("%w: %w", ErrInvalidBlockSelection, err)
	}
	return selection, nil
}

// CreateInput defines input for creating a chain query directly.
type CreateInput struct {
	ChannelID      uuid.UUID
	IdempotencyKey string
	QueryKind      apiClient.QueryKind
	ChainSelector  string
	Params         apiClient.EVMCallQueryParams
	Metadata       map[string]interface{}
}

// ListInput defines filters and pagination for listing channel queries.
type ListInput struct {
	ChannelID uuid.UUID
	Status    *[]apiClient.QueryStatus
	Limit     *int
	Offset    *int64
}

// CallContractInput defines a raw EVM call query request.
type CallContractInput struct {
	ChannelID       uuid.UUID
	ChainSelector   string
	ContractAddress string
	CallData        any
	BlockSelection  BlockSelection
	IdempotencyKey  string
	FromAddress     *string
	Metadata        map[string]interface{}
}

// CallOption customizes CallContract and CallContractWithABI requests.
type CallOption func(*CallContractInput)

// WithFromAddress sets the EVM call sender. If omitted, the SDK sends the zero address.
func WithFromAddress(fromAddress string) CallOption {
	return func(input *CallContractInput) {
		input.FromAddress = &fromAddress
	}
}

// WithMetadata attaches customer metadata to the create-query request.
func WithMetadata(metadata map[string]interface{}) CallOption {
	return func(input *CallContractInput) {
		input.Metadata = metadata
	}
}

// ResolvedBlock contains concrete block metadata decoded from a terminal verifiable_result.
type ResolvedBlock struct {
	BlockNumber    string
	BlockHash      string
	BlockTimestamp int64
}

// QueryError contains terminal query execution error details.
type QueryError struct {
	Code             string
	Message          string
	RawRevertData    []byte
	RawRevertDataHex string
}

// CallContractResult is the SDK structured result for raw EVM calls.
type CallContractResult struct {
	QueryID          string
	Status           string
	ChainSelector    string
	Target           string
	RawReturnData    []byte
	VerifiableQuery  []byte
	VerifiableResult string
	EventHash        string
	Proof            OCRProof
	Block            *ResolvedBlock
	Error            *QueryError
	Query            *apiClient.Query
}

// Create creates an asynchronous chain query under a channel.
func (c *Client) Create(ctx context.Context, input CreateInput) (*apiClient.QueryAcceptedResponse, error) {
	if err := validateCreateInput(&input); err != nil {
		return nil, err
	}

	var metadata *map[string]interface{}
	if input.Metadata != nil {
		m := input.Metadata
		metadata = &m
	}

	req := apiClient.CreateQuery{
		IdempotencyKey: input.IdempotencyKey,
		QueryKind:      input.QueryKind,
		ChainSelector:  apiClient.ChainSelector(input.ChainSelector),
		Params:         input.Params,
		Metadata:       metadata,
	}

	resp, err := c.apiClient.CreateQueryWithResponse(ctx, input.ChannelID, req)
	if err != nil {
		c.logger.Error("Failed to create query", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateQuery, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateQuery, ErrNilResponse)
	}

	switch resp.StatusCode() {
	case 202:
		if resp.JSON202 == nil {
			return nil, fmt.Errorf("%w: %w", ErrCreateQuery, ErrNilResponseBody)
		}
		c.logger.Info("Query accepted",
			"query_id", resp.JSON202.QueryId.String(),
			"channel_id", input.ChannelID.String(),
			"status", resp.JSON202.Status)
		return resp.JSON202, nil
	case 404:
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	case 409:
		return nil, fmt.Errorf("%w: %w", ErrCreateQuery, ErrIdempotencyConflict)
	case 429:
		return nil, fmt.Errorf("%w: %w", ErrCreateQuery, ErrRateLimitExceeded)
	default:
		c.logger.Error("Unexpected status code when creating query",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateQuery, ErrUnexpectedStatusCode, resp.StatusCode())
	}
}

// CreateEVMCall creates an evm_call query without waiting for terminal status.
func (c *Client) CreateEVMCall(ctx context.Context, input CallContractInput, options ...any) (*apiClient.QueryAcceptedResponse, error) {
	if err := applyCallOptions(&input, options...); err != nil {
		return nil, err
	}
	params, err := buildEVMCallQueryParams(input)
	if err != nil {
		return nil, err
	}
	return c.Create(ctx, CreateInput{
		ChannelID:      input.ChannelID,
		IdempotencyKey: input.IdempotencyKey,
		QueryKind:      apiClient.QueryKindEVMCall,
		ChainSelector:  input.ChainSelector,
		Params:         params,
		Metadata:       input.Metadata,
	})
}

// Get retrieves a specific query resource by channel ID and query ID.
func (c *Client) Get(ctx context.Context, channelID uuid.UUID, queryID uuid.UUID) (*apiClient.Query, error) {
	if channelID == uuid.Nil {
		return nil, ErrChannelIDRequired
	}
	if queryID == uuid.Nil {
		return nil, ErrQueryIDRequired
	}

	resp, err := c.apiClient.GetQuery(ctx, channelID, queryID)
	if err != nil {
		c.logger.Error("Failed to get query", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetQuery, err)
	}
	body, err := readRawResponseBody(resp, ErrGetQuery)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var query apiClient.Query
		if err := unmarshalJSONBody(body, &query, ErrGetQuery); err != nil {
			return nil, err
		}
		return &query, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: query ID %s in channel %s", ErrQueryNotFound, queryID.String(), channelID.String())
	default:
		c.logger.Error("Unexpected status code when getting query",
			"status_code", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetQuery, ErrUnexpectedStatusCode, resp.StatusCode)
	}
}

// List retrieves query records for a channel.
func (c *Client) List(ctx context.Context, input ListInput) ([]apiClient.Query, bool, error) {
	if input.ChannelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}
	if input.Limit != nil && (*input.Limit < 1 || *input.Limit > 100) {
		return nil, false, ErrInvalidLimit
	}
	if input.Offset != nil && *input.Offset < 0 {
		return nil, false, ErrInvalidOffset
	}

	params := apiClient.ListQueriesParams{
		Status: input.Status,
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	resp, err := c.apiClient.ListQueries(ctx, input.ChannelID, &params)
	if err != nil {
		c.logger.Error("Failed to list queries", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListQueries, err)
	}
	body, err := readRawResponseBody(resp, ErrListQueries)
	if err != nil {
		return nil, false, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var queryList apiClient.QueryList
		if err := unmarshalJSONBody(body, &queryList, ErrListQueries); err != nil {
			return nil, false, err
		}
		return queryList.Data, queryList.HasMore, nil
	case http.StatusNotFound:
		return nil, false, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	default:
		c.logger.Error("Unexpected status code when listing queries",
			"status_code", resp.StatusCode,
			"body", string(body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListQueries, ErrUnexpectedStatusCode, resp.StatusCode)
	}
}

// Wait polls a query until it reaches a terminal status.
func (c *Client) Wait(ctx context.Context, channelID uuid.UUID, queryID uuid.UUID) (*apiClient.Query, error) {
	if channelID == uuid.Nil {
		return nil, ErrChannelIDRequired
	}
	if queryID == uuid.Nil {
		return nil, ErrQueryIDRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pollInterval := c.pollInterval
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		query, err := c.Get(ctx, channelID, queryID)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			if isTransientQueryError(err) {
				c.logger.Warn("Transient error while waiting for query, will retry", "error", err)
				goto waitForNextPoll
			}
			return nil, fmt.Errorf("%w: %w", ErrWaitQuery, err)
		}
		if IsTerminalStatus(query.Status) {
			return query, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
		continue

	waitForNextPoll:
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// CallContract creates an evm_call query, waits for terminal status, and decodes the verifiable_result.
func (c *Client) CallContract(
	ctx context.Context,
	channelID uuid.UUID,
	chainSelector string,
	contractAddress string,
	callData any,
	blockSelection BlockSelection,
	idempotencyKey string,
	options ...any,
) (*CallContractResult, error) {
	input := CallContractInput{
		ChannelID:       channelID,
		ChainSelector:   chainSelector,
		ContractAddress: contractAddress,
		CallData:        callData,
		BlockSelection:  blockSelection,
		IdempotencyKey:  idempotencyKey,
	}

	accepted, err := c.CreateEVMCall(ctx, input, options...)
	if err != nil {
		return nil, err
	}

	query, err := c.Wait(ctx, channelID, accepted.QueryId)
	if err != nil {
		return nil, err
	}

	result, err := ResultFromQuery(query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBuildCallContractResult, err)
	}
	return result, nil
}

// ResultFromQuery builds a CallContractResult from a query resource.
func ResultFromQuery(query *apiClient.Query) (*CallContractResult, error) {
	if query == nil {
		return nil, ErrQueryRequired
	}

	result := &CallContractResult{
		QueryID:       query.QueryId.String(),
		Status:        string(query.Status),
		ChainSelector: string(query.ChainSelector),
		Target:        queryTarget(query),
		Query:         query,
	}

	if query.EventHash != nil {
		result.EventHash = *query.EventHash
	}
	if query.Proof != nil {
		result.Proof = *query.Proof
	}

	if query.Status == apiClient.QueryStatusCompleted && (query.VerifiableResult == nil || *query.VerifiableResult == "") {
		return nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableResult, ErrVerifiableResultRequired)
	}

	if query.VerifiableResult != nil && *query.VerifiableResult != "" {
		decodedBytes, verifiableEvent, err := DecodeVerifiableResultBytes(*query.VerifiableResult)
		if err != nil {
			return nil, err
		}

		result.VerifiableResult = *query.VerifiableResult
		result.VerifiableQuery = decodedBytes
		result.ChainSelector = verifiableEvent.ChainSelector
		if result.Target == "" {
			result.Target = displayTargetFromVerifiableEvent(verifiableEvent)
		}

		if verifiableEvent.Data.Result != nil {
			rawReturnData, err := hexToBytesStrict(verifiableEvent.Data.Result.RawReturnData)
			if err != nil {
				return nil, fmt.Errorf("%w: raw_return_data: %w", ErrDecodeVerifiableResult, err)
			}
			result.RawReturnData = rawReturnData
		}

		if verifiableEvent.Data.BlockSelection.Resolved != nil {
			result.Block = &ResolvedBlock{
				BlockNumber:    verifiableEvent.Data.BlockSelection.Resolved.BlockNumber,
				BlockHash:      verifiableEvent.Data.BlockSelection.Resolved.BlockHash,
				BlockTimestamp: verifiableEvent.Data.BlockSelection.Resolved.BlockTimestamp,
			}
		}

		if verifiableEvent.Data.Error != nil {
			queryErr := &QueryError{
				Code:    string(verifiableEvent.Data.Error.Code),
				Message: verifiableEvent.Data.Error.Message,
			}
			if verifiableEvent.Data.Error.RawRevertData != nil {
				queryErr.RawRevertDataHex = *verifiableEvent.Data.Error.RawRevertData
				rawRevertData, err := hexToBytesStrict(*verifiableEvent.Data.Error.RawRevertData)
				if err != nil {
					return nil, fmt.Errorf("%w: raw_revert_data: %w", ErrDecodeVerifiableResult, err)
				}
				queryErr.RawRevertData = rawRevertData
			}
			result.Error = queryErr
		}
	}

	if result.Error == nil {
		switch query.Status {
		case apiClient.QueryStatusFailed:
			result.Error = &QueryError{Code: "CRE_WORKFLOW_FAILED"}
		case apiClient.QueryStatusExpired:
			result.Error = &QueryError{Code: "QUERY_EXPIRED", Message: "query expired before terminal callback"}
		}
	}

	return result, nil
}

// ResultFromQuery builds a CallContractResult from a query resource.
func (c *Client) ResultFromQuery(query *apiClient.Query) (*CallContractResult, error) {
	return ResultFromQuery(query)
}

// DecodeVerifiableResult decodes a base64-encoded chain query verifiable_result.
func DecodeVerifiableResult(verifiableResult string) (*models.ChainQueryVerifiableEvent, error) {
	_, event, err := DecodeVerifiableResultBytes(verifiableResult)
	return event, err
}

// DecodeVerifiableResult decodes a base64-encoded chain query verifiable_result.
func (c *Client) DecodeVerifiableResult(verifiableResult string) (*models.ChainQueryVerifiableEvent, error) {
	return DecodeVerifiableResult(verifiableResult)
}

// DecodeVerifiableResultBytes decodes verifiable_result and returns both decoded JSON bytes and typed data.
func DecodeVerifiableResultBytes(verifiableResult string) ([]byte, *models.ChainQueryVerifiableEvent, error) {
	if verifiableResult == "" {
		return nil, nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableResult, ErrVerifiableResultRequired)
	}

	decoded, err := base64.StdEncoding.DecodeString(verifiableResult)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableResult, fmt.Errorf("%w: %w", ErrInvalidVerifiableResultBase64, err))
	}

	var event models.ChainQueryVerifiableEvent
	if err := json.Unmarshal(decoded, &event); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDecodeVerifiableResult, fmt.Errorf("%w: %w", ErrInvalidVerifiableResultJSON, err))
	}

	return decoded, &event, nil
}

var chainQueryTargetFieldPriority = []string{
	"TargetTxHash",
	"TargetBlockNumber",
	"TargetLogFilter",
	"TargetFilterName",
	"TargetContract",
	"TargetAccount",
	"TargetEmitterContract",
	"ContractAddress",
	"TxHash",
	"BlockNumber",
	"LogFilter",
	"FilterName",
	"Account",
	"EmitterContract",
	"Address",
}

var chainQueryTargetMapKeyPriority = []string{
	"targetTxHash",
	"targetBlockNumber",
	"targetLogFilter",
	"targetFilterName",
	"targetContract",
	"targetAccount",
	"targetEmitterContract",
	"contractAddress",
	"txHash",
	"blockNumber",
	"logFilter",
	"filterName",
	"account",
	"emitterContract",
	"address",
}

func queryTarget(query *apiClient.Query) string {
	if query == nil {
		return ""
	}
	return targetStringFromNamedField(reflect.ValueOf(query), "Target")
}

func displayTargetFromVerifiableEvent(event *models.ChainQueryVerifiableEvent) string {
	if event == nil {
		return ""
	}
	return targetStringFromTargetValue(reflect.ValueOf(event.Data.Target))
}

func targetStringFromNamedField(value reflect.Value, fieldName string) string {
	value = unwrapTargetValue(value)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return ""
	}
	return targetStringFromValue(value.FieldByName(fieldName))
}

func targetStringFromTargetValue(value reflect.Value) string {
	value = unwrapTargetValue(value)
	if !value.IsValid() {
		return ""
	}

	switch value.Kind() {
	case reflect.Struct:
		for _, fieldName := range chainQueryTargetFieldPriority {
			if target := targetStringFromValue(value.FieldByName(fieldName)); target != "" {
				return target
			}
		}
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return ""
		}
		for _, keyName := range chainQueryTargetMapKeyPriority {
			key := reflect.ValueOf(keyName)
			if key.Type().ConvertibleTo(value.Type().Key()) {
				key = key.Convert(value.Type().Key())
			} else {
				continue
			}
			if target := targetStringFromValue(value.MapIndex(key)); target != "" {
				return target
			}
		}
	}

	return ""
}

func unwrapTargetValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func targetStringFromValue(value reflect.Value) string {
	value = unwrapTargetValue(value)
	if !value.IsValid() {
		return ""
	}

	switch value.Kind() {
	case reflect.String:
		return strings.TrimSpace(value.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10)
	}

	if value.CanInterface() {
		if stringer, ok := value.Interface().(fmt.Stringer); ok {
			return strings.TrimSpace(stringer.String())
		}
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			bytes := make([]byte, value.Len())
			for i := range bytes {
				bytes[i] = byte(value.Index(i).Uint())
			}
			if len(bytes) == 0 {
				return ""
			}
			return "0x" + hex.EncodeToString(bytes)
		}
	}

	if value.CanInterface() {
		encoded, err := json.Marshal(value.Interface())
		if err == nil {
			target := strings.TrimSpace(string(encoded))
			if target != "" && target != "null" && target != "{}" && target != "[]" {
				return target
			}
		}
	}

	return ""
}

// IsTerminalStatus reports whether a query status is terminal.
func IsTerminalStatus(status apiClient.QueryStatus) bool {
	switch status {
	case apiClient.QueryStatusCompleted, apiClient.QueryStatusFailed, apiClient.QueryStatusExpired:
		return true
	default:
		return false
	}
}

func validateCreateInput(input *CreateInput) error {
	if input.ChannelID == uuid.Nil {
		return ErrChannelIDRequired
	}
	if input.IdempotencyKey == "" {
		return ErrIdempotencyKeyRequired
	}
	if input.QueryKind == "" {
		return ErrQueryKindRequired
	}
	if input.QueryKind != apiClient.QueryKindEVMCall {
		return fmt.Errorf("%w: %s", ErrUnsupportedQueryKind, input.QueryKind)
	}
	if input.ChainSelector == "" || input.ChainSelector == "0" {
		return ErrChainSelectorRequired
	}

	contractAddress, err := normalizeAddress(input.Params.ContractAddress, ErrContractAddressRequired, ErrInvalidContractAddress)
	if err != nil {
		return err
	}
	input.Params.ContractAddress = apiClient.EthereumAddress(contractAddress)

	callData, err := normalizeHexBytes(input.Params.CallData, ErrCallDataRequired, ErrInvalidCallData)
	if err != nil {
		return err
	}
	input.Params.CallData = callData

	if err := validateBlockSelection(input.Params.BlockSelection); err != nil {
		return err
	}

	if input.Params.FromAddress != nil {
		from, err := normalizeAddress(*input.Params.FromAddress, ErrInvalidFromAddress, ErrInvalidFromAddress)
		if err != nil {
			return err
		}
		fromAddress := apiClient.EthereumAddress(from)
		input.Params.FromAddress = &fromAddress
	}

	return nil
}

func buildEVMCallQueryParams(input CallContractInput) (apiClient.EVMCallQueryParams, error) {
	if input.ChannelID == uuid.Nil {
		return apiClient.EVMCallQueryParams{}, ErrChannelIDRequired
	}
	if input.ChainSelector == "" || input.ChainSelector == "0" {
		return apiClient.EVMCallQueryParams{}, ErrChainSelectorRequired
	}
	if input.IdempotencyKey == "" {
		return apiClient.EVMCallQueryParams{}, ErrIdempotencyKeyRequired
	}

	contractAddress, err := normalizeAddress(input.ContractAddress, ErrContractAddressRequired, ErrInvalidContractAddress)
	if err != nil {
		return apiClient.EVMCallQueryParams{}, err
	}

	callData, err := normalizeHexBytes(input.CallData, ErrCallDataRequired, ErrInvalidCallData)
	if err != nil {
		return apiClient.EVMCallQueryParams{}, err
	}

	if err := validateBlockSelection(input.BlockSelection); err != nil {
		return apiClient.EVMCallQueryParams{}, err
	}

	fromAddress := zeroAddress
	if input.FromAddress != nil {
		fromAddress = *input.FromAddress
	}
	normalizedFrom, err := normalizeAddress(fromAddress, ErrInvalidFromAddress, ErrInvalidFromAddress)
	if err != nil {
		return apiClient.EVMCallQueryParams{}, err
	}
	from := apiClient.EthereumAddress(normalizedFrom)

	return apiClient.EVMCallQueryParams{
		ContractAddress: apiClient.EthereumAddress(contractAddress),
		CallData:        callData,
		BlockSelection:  input.BlockSelection,
		FromAddress:     &from,
	}, nil
}

func applyCallOptions(input *CallContractInput, options ...any) error {
	for _, option := range options {
		switch opt := option.(type) {
		case nil:
			continue
		case CallOption:
			opt(input)
		case func(*CallContractInput):
			opt(input)
		case string:
			v := opt
			input.FromAddress = &v
		case *string:
			if opt != nil {
				v := *opt
				input.FromAddress = &v
			}
		case common.Address:
			v := opt.Hex()
			input.FromAddress = &v
		case *common.Address:
			if opt != nil {
				v := opt.Hex()
				input.FromAddress = &v
			}
		case map[string]interface{}:
			input.Metadata = opt
		default:
			return fmt.Errorf("%w: %T", ErrInvalidCallOption, option)
		}
	}
	return nil
}

func validateBlockSelection(selection BlockSelection) error {
	discriminator, err := selection.Discriminator()
	if err != nil || discriminator == "" {
		return ErrBlockSelectionRequired
	}

	switch discriminator {
	case "latest", "finalized":
		return nil
	case "block_number":
		blockSelection, err := selection.AsBlockNumberBlockSelection()
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidBlockSelection, err)
		}
		if blockSelection.BlockNumber == "" {
			return fmt.Errorf("%w: block_number is required", ErrInvalidBlockSelection)
		}
		if _, err := strconv.ParseUint(blockSelection.BlockNumber, 10, 64); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidBlockSelection, err)
		}
		return nil
	default:
		return fmt.Errorf("%w: unsupported type %q", ErrInvalidBlockSelection, discriminator)
	}
}

func normalizeAddress(address string, requiredErr error, invalidErr error) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", requiredErr
	}
	if !common.IsHexAddress(address) {
		return "", invalidErr
	}
	return strings.ToLower(common.HexToAddress(address).Hex()), nil
}

func normalizeHexBytes(value any, requiredErr error, invalidErr error) (string, error) {
	switch v := value.(type) {
	case []byte:
		return "0x" + hex.EncodeToString(v), nil
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return "", requiredErr
		}
		if _, err := hexToBytesStrict(v); err != nil {
			return "", invalidErr
		}
		return "0x" + strings.ToLower(strings.TrimPrefix(v, "0x")), nil
	default:
		return "", fmt.Errorf("%w: got %T", invalidErr, value)
	}
}

func hexToBytesStrict(value string) ([]byte, error) {
	if value == "" {
		return nil, ErrInvalidHexBytes
	}
	if !strings.HasPrefix(value, "0x") {
		return nil, ErrInvalidHexBytes
	}
	hexPart := strings.TrimPrefix(value, "0x")
	if len(hexPart)%2 != 0 {
		return nil, ErrInvalidHexBytes
	}
	decoded, err := hex.DecodeString(hexPart)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidHexBytes, err)
	}
	return decoded, nil
}

func readRawResponseBody(resp *http.Response, wrapErr error) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("%w: %w", wrapErr, ErrNilResponse)
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("%w: %w", wrapErr, ErrNilResponseBody)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", wrapErr, err)
	}
	return body, nil
}

func unmarshalJSONBody(body []byte, dest any, wrapErr error) error {
	if len(body) == 0 {
		return fmt.Errorf("%w: %w", wrapErr, ErrNilResponseBody)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("%w: %w", wrapErr, err)
	}
	return nil
}

func isTransientQueryError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, ErrChannelIDRequired) ||
		errors.Is(err, ErrQueryIDRequired) ||
		errors.Is(err, ErrChannelNotFound) ||
		errors.Is(err, ErrQueryNotFound) {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	if matches := statusCodePattern.FindStringSubmatch(errMsg); len(matches) > 1 {
		if statusCode, err := strconv.Atoi(matches[1]); err == nil {
			if statusCode == http.StatusTooManyRequests || (statusCode >= http.StatusInternalServerError && statusCode < 600) {
				return true
			}
		}
	}

	transientIndicators := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"eof",
		"broken pipe",
		"no such host",
		"network is unreachable",
	}
	for _, indicator := range transientIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}

	return false
}
