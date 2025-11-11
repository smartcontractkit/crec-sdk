package operations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-sdk/client"
)

// Sentinel errors
var (
	// Service initialization errors
	ErrServiceOptionsRequired = errors.New("ServiceOptions is required")
	ErrCRECClientRequired     = errors.New("CRECClient is required")

	// Validation errors
	ErrChannelIDRequired             = errors.New("channel_id is required")
	ErrChainSelectorRequired         = errors.New("chain_selector is required")
	ErrAddressRequired               = errors.New("address is required")
	ErrWalletOperationIDRequired     = errors.New("wallet_operation_id is required")
	ErrAtLeastOneTransactionRequired = errors.New("at least one transaction is required")
	ErrSignatureRequired             = errors.New("signature is required")

	// Not found errors
	ErrChannelNotFound   = errors.New("channel not found")
	ErrOperationNotFound = errors.New("operation not found")

	// API operation errors
	ErrCreateOperation = errors.New("failed to create operation")
	ErrGetOperation    = errors.New("failed to get operation")
	ErrListOperations  = errors.New("failed to list operations")

	// Response errors
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrNilResponseBody      = errors.New("unexpected nil response body")
)

const (
	ServiceName      = "operations"
	MaxLogBodyLength = 200
)

// ServiceOptions defines the options for creating a new CREC Operations service.
//   - Logger: Optional logger instance.
//   - CRECClient: The CREC API client instance.
type ServiceOptions struct {
	Logger     *zerolog.Logger
	CRECClient *client.CRECClient
}

// Service provides operations for managing CREC operations.
// Operations represent transaction execution requests that are sent through channels.
// Each operation contains one or more transactions to be executed atomically on-chain.
type Service struct {
	logger     *zerolog.Logger
	crecClient *client.CRECClient
}

// NewService creates a new CREC Operations service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Operations service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, ErrServiceOptionsRequired
	}

	if opts.CRECClient == nil {
		return nil, ErrCRECClientRequired
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC Operations service")

	return &Service{
		logger:     logger,
		crecClient: opts.CRECClient,
	}, nil
}

// TransactionRequest represents a single transaction in an operation.
//   - To: The target contract address.
//   - Value: The amount of native currency to send (as string).
//   - Data: The encoded calldata for the transaction.
type TransactionRequest struct {
	To    string
	Value string
	Data  string
}

// CreateOperationInput defines the input parameters for creating a new operation.
//   - ChannelID: The UUID of the channel where the operation will be created.
//   - ChainSelector: The chain selector to identify the chain where the operation will be executed.
//   - Address: The account address performing the operation.
//   - WalletOperationID: Unique identifier for the wallet operation.
//   - Transactions: List of transactions to execute (at least one required).
//   - Signature: EIP-712 signature of the operation.
type CreateOperationInput struct {
	ChannelID         uuid.UUID
	ChainSelector     uint64
	Address           string
	WalletOperationID string
	Transactions      []TransactionRequest
	Signature         string
}

// CreateOperation creates a new operation in the specified channel.
// The operation will contain one or more transactions to be executed atomically.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The operation creation parameters.
//
// Returns the operation ID or an error if the operation fails.
func (s *Service) CreateOperation(ctx context.Context, input CreateOperationInput) (*uuid.UUID, error) {
	s.logger.Debug().
		Str("channel_id", input.ChannelID.String()).
		Str("wallet_operation_id", input.WalletOperationID).
		Uint64("chain_selector", input.ChainSelector).
		Str("address", input.Address).
		Int("num_transactions", len(input.Transactions)).
		Msg("Creating operation")

	// Validate input
	if input.ChannelID == uuid.Nil {
		return nil, ErrChannelIDRequired
	}
	if input.ChainSelector == 0 {
		return nil, ErrChainSelectorRequired
	}
	if input.Address == "" {
		return nil, ErrAddressRequired
	}
	if input.WalletOperationID == "" {
		return nil, ErrWalletOperationIDRequired
	}
	if len(input.Transactions) == 0 {
		return nil, ErrAtLeastOneTransactionRequired
	}
	if input.Signature == "" {
		return nil, ErrSignatureRequired
	}

	// Convert transactions
	transactions := make([]apiClient.TransactionRequest, 0, len(input.Transactions))
	for _, tx := range input.Transactions {
		transactions = append(transactions, apiClient.TransactionRequest{
			To:    tx.To,
			Value: tx.Value,
			Data:  tx.Data,
		})
	}

	createOperationReq := apiClient.CreateOperation{
		ChainSelector:     input.ChainSelector,
		Address:           input.Address,
		WalletOperationId: input.WalletOperationID,
		Transactions:      transactions,
		Signature:         input.Signature,
	}

	resp, err := s.crecClient.PostChannelsChannelIdOperationsWithResponse(ctx, input.ChannelID, createOperationReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create operation")
		return nil, fmt.Errorf("%w: %w", ErrCreateOperation, err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", input.ChannelID.String()).
			Msg("Channel not found")
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	}

	if resp.StatusCode() != 201 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Unexpected status code when creating operation")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateOperation, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperation, ErrNilResponseBody)
	}

	operationID := resp.JSON201.OperationId

	s.logger.Info().
		Str("operation_id", operationID.String()).
		Str("channel_id", input.ChannelID.String()).
		Str("wallet_operation_id", input.WalletOperationID).
		Msg("Operation created successfully")

	return &operationID, nil
}

// GetOperation retrieves a specific operation by its ID within a channel.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel containing the operation.
//   - operationID: The UUID of the operation to retrieve.
//
// Returns the operation or an error if the operation fails or is not found.
func (s *Service) GetOperation(ctx context.Context, channelID uuid.UUID, operationID uuid.UUID) (*apiClient.Operation, error) {
	s.logger.Debug().
		Str("channel_id", channelID.String()).
		Str("operation_id", operationID.String()).
		Msg("Getting operation")

	resp, err := s.crecClient.GetChannelsChannelIdOperationsOperationIdWithResponse(ctx, channelID, operationID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get operation")
		return nil, fmt.Errorf("%w: %w", ErrGetOperation, err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", channelID.String()).
			Str("operation_id", operationID.String()).
			Msg("Operation not found")
		return nil, fmt.Errorf("%w: operation ID %s in channel %s", ErrOperationNotFound, operationID.String(), channelID.String())
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Unexpected status code when getting operation")
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetOperation, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetOperation, ErrNilResponseBody)
	}

	s.logger.Debug().
		Str("operation_id", resp.JSON200.OperationId.String()).
		Str("status", resp.JSON200.Status).
		Msg("Operation retrieved successfully")

	return resp.JSON200, nil
}

// ListOperationsInput defines the input parameters for listing operations.
//   - ChannelID: The UUID of the channel to list operations from.
//   - Status: Optional filter for operation status.
//   - ChainSelector: Optional filter for chain selector.
//   - Address: Optional filter for account address.
//   - WalletID: Optional filter for wallet ID.
//   - Limit: Maximum number of operations to return (1-100, default: 20).
//   - Offset: Number of operations to skip for pagination (default: 0).
type ListOperationsInput struct {
	ChannelID     uuid.UUID
	Status        *string
	ChainSelector *uint64
	Address       *string
	WalletID      *uuid.UUID
	Limit         *int
	Offset        *int
}

// ListOperations retrieves a list of operations for a channel.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The list parameters including filters and pagination.
//
// Returns a list of operations and a boolean indicating if there are more results.
func (s *Service) ListOperations(ctx context.Context, input ListOperationsInput) ([]apiClient.Operation, bool, error) {
	s.logger.Debug().
		Str("channel_id", input.ChannelID.String()).
		Interface("filters", input).
		Msg("Listing operations")

	if input.ChannelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}

	params := apiClient.GetChannelsChannelIdOperationsParams{
		Status:        input.Status,
		ChainSelector: convertUint64PtrToStringPtr(input.ChainSelector),
		Address:       input.Address,
		WalletId:      input.WalletID,
		Limit:         input.Limit,
		Offset:        input.Offset,
	}

	resp, err := s.crecClient.GetChannelsChannelIdOperationsWithResponse(ctx, input.ChannelID, &params)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list operations")
		return nil, false, fmt.Errorf("%w: %w", ErrListOperations, err)
	}

	if resp.StatusCode() == 404 {
		s.logger.Warn().
			Str("channel_id", input.ChannelID.String()).
			Msg("Channel not found")
		return nil, false, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	}

	if resp.StatusCode() != 200 {
		s.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("body", truncateBody(resp.Body)).
			Msg("Unexpected status code when listing operations")
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListOperations, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListOperations, ErrNilResponseBody)
	}

	s.logger.Debug().
		Int("count", len(resp.JSON200.Data)).
		Bool("has_more", resp.JSON200.HasMore).
		Msg("Operations listed successfully")

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}

// truncateBody truncates a response body to MaxLogBodyLength for logging purposes.
func truncateBody(body []byte) string {
	bodyStr := string(body)
	if len(bodyStr) <= MaxLogBodyLength {
		return bodyStr
	}
	return bodyStr[:MaxLogBodyLength] + "... (truncated)"
}

// convertUint64PtrToStringPtr converts a pointer to uint64 to a pointer to string
// for API compatibility. Returns nil if the input is nil.
func convertUint64PtrToStringPtr(val *uint64) *string {
	if val == nil {
		return nil
	}
	str := strconv.FormatUint(*val, 10)
	return &str
}
