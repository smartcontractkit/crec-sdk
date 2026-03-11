package transact

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk/transact/eip712"
	"github.com/smartcontractkit/crec-sdk/transact/signer"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

var (
	// ErrOptionsRequired is returned when the options parameter is nil.
	ErrOptionsRequired = errors.New("options is required")
	// ErrCRECClientRequired is returned when the CREC client is nil in options.
	ErrCRECClientRequired = errors.New("CRECClient is required")

	// ErrChannelIDRequired is returned when the channel ID is nil.
	ErrChannelIDRequired = errors.New("channel_id is required")
	// ErrChainSelectorRequired is returned when the chain selector is empty or zero.
	ErrChainSelectorRequired = errors.New("chain_selector is required")
	// ErrAddressRequired is returned when the address is empty.
	ErrAddressRequired = errors.New("address is required")
	// ErrWalletOperationIDRequired is returned when the wallet operation ID is empty.
	ErrWalletOperationIDRequired = errors.New("wallet_operation_id is required")
	// ErrAtLeastOneTransactionRequired is returned when the transactions list is empty.
	ErrAtLeastOneTransactionRequired = errors.New("at least one transaction is required")
	// ErrSignatureRequired is returned when the signature is empty.
	ErrSignatureRequired = errors.New("signature is required")

	// ErrChannelNotFound is returned when the channel does not exist (404 response).
	ErrChannelNotFound = errors.New("channel not found")
	// ErrOperationNotFound is returned when the operation does not exist (404 response).
	ErrOperationNotFound = errors.New("operation not found")

	// ErrCreateOperation is returned when creating an operation fails.
	ErrCreateOperation = errors.New("failed to create operation")
	// ErrGetOperation is returned when fetching an operation fails.
	ErrGetOperation = errors.New("failed to get operation")
	// ErrListOperations is returned when listing operations fails.
	ErrListOperations = errors.New("failed to list operations")
	// ErrSendOperation is returned when sending an operation fails.
	ErrSendOperation = errors.New("failed to send operation")

	// ErrUnexpectedStatusCode is returned when the API returns an unexpected HTTP status code.
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	// ErrNilResponseBody is returned when the API response body is nil.
	ErrNilResponseBody = errors.New("unexpected nil response body")
)

// Options defines the options for creating a new CREC transact client used to send operations to the CREC system.
// It includes a logger for logging messages and a chain ID for the blockchain network.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system (required).
type Options struct {
	Logger     *slog.Logger
	CRECClient *apiClient.ClientWithResponses
}

// Client provides operations for creating, signing, and sending CREC operations.
// It embeds an EIP712Handler for hashing and signing operations.
type Client struct {
	logger     *slog.Logger
	crecClient *apiClient.ClientWithResponses
	// EIP712Handler provides hashing and signing operations for CREC operations.
	EIP712Handler *eip712.Handler
}

// NewClient creates a new CREC transact client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC transact client, see Options for details.
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

	logger.Debug("Creating CREC transact client")

	eip712Handler, err := eip712.NewHandler(&eip712.Options{
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create EIP-712 handler: %w", err)
	}

	return &Client{
		logger:        logger,
		crecClient:    opts.CRECClient,
		EIP712Handler: eip712Handler,
	}, nil
}

// HashOperation computes the EIP-712 digest of the given operation.
// This method delegates to the embedded EIP712Handler.
//   - op: The operation to hash.
//   - chainSelector: chainSelector of the blockchain network in which the operation is being executed.
//
// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
func (c *Client) HashOperation(op *types.Operation, chainSelector string) (common.Hash, error) {
	return c.EIP712Handler.HashOperation(op, chainSelector)
}

// SignOperation signs the given operation using the provided signer, returning the operation hash and the signature
// over the hash. This method delegates to the embedded EIP712Handler.
//   - ctx: The context for the request.
//   - op: The operation to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
//   - chainSelector: chainSelector of the blockchain network in which the operation is being executed.
//
// Fetches chainID corresponding to the chain selector from smartcontractkit/chain-selectors package.
func (c *Client) SignOperation(
	ctx context.Context,
	op *types.Operation,
	signer signer.Signer,
	chainSelector string,
) (common.Hash, []byte, error) {
	return c.EIP712Handler.SignOperation(ctx, op, signer, chainSelector)
}

// SignOperationHash signs the given operation hash using the provided signer, returning the signature.
// This method delegates to the embedded EIP712Handler.
//   - ctx: The context for the request.
//   - opHash: The operation hash to sign.
//   - signer: The signer to use for signing the operation. See signer.Signer for details.
func (c *Client) SignOperationHash(
	ctx context.Context,
	opHash common.Hash,
	signer signer.Signer,
) ([]byte, error) {
	return c.EIP712Handler.SignOperationHash(ctx, opHash, signer)
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
	ChainSelector     string
	Address           string
	WalletOperationID string
	Transactions      []TransactionRequest
	Signature         string
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

// CreateOperation creates a new operation in the specified channel.
// The operation will contain one or more transactions to be executed atomically.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The operation creation parameters.
//
// Returns the operation ID or an error if the operation fails.
func (c *Client) CreateOperation(ctx context.Context, input CreateOperationInput) (*uuid.UUID, error) {
	c.logger.Debug("Creating operation",
		"channel_id", input.ChannelID.String(),
		"wallet_operation_id", input.WalletOperationID,
		"chain_selector", input.ChainSelector,
		"address", input.Address,
		"num_transactions", len(input.Transactions))

	// Validate input
	if input.ChannelID == uuid.Nil {
		return nil, ErrChannelIDRequired
	}
	if input.ChainSelector == "" || input.ChainSelector == "0" {
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

	resp, err := c.crecClient.PostChannelsChannelIdOperationsWithResponse(ctx, input.ChannelID, createOperationReq)
	if err != nil {
		c.logger.Error("Failed to create operation", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrCreateOperation, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found", "channel_id", input.ChannelID.String())
		return nil, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	}

	if resp.StatusCode() != 201 {
		c.logger.Error("Unexpected status code when creating operation",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrCreateOperation, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON201 == nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperation, ErrNilResponseBody)
	}

	operationID := resp.JSON201.OperationId

	c.logger.Info("Operation created successfully",
		"operation_id", operationID.String(),
		"channel_id", input.ChannelID.String(),
		"wallet_operation_id", input.WalletOperationID)

	return &operationID, nil
}

// SendSignedOperation sends a signed operation to the CREC system via the specified channel.
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to send the operation to.
//   - op: The operation to send, which must be signed.
//   - signature: The signature of the operation, to be verified by the onchain smart account.
//   - chainSelector: The chain selector of the blockchain network in which the operation is being executed.
func (c *Client) SendSignedOperation(
	ctx context.Context,
	channelID uuid.UUID,
	op *types.Operation,
	signature []byte,
	chainSelector string,
) (*apiClient.Operation, error) {
	if op == nil {
		return nil, errors.New("operation is required")
	}

	c.logger.Debug("Sending signed operation",
		"channel_id", channelID.String(),
		"chain_selector", chainSelector,
		"operation_id", op.ID.String(),
		"signature", common.Bytes2Hex(signature))

	var transactions []TransactionRequest
	for _, tx := range op.Transactions {
		transactions = append(
			transactions, TransactionRequest{
				To:    tx.To.String(),
				Value: tx.Value.String(),
				Data:  "0x" + common.Bytes2Hex(tx.Data),
			},
		)
	}

	input := CreateOperationInput{
		ChannelID:         channelID,
		ChainSelector:     chainSelector,
		Address:           op.Account.String(),
		WalletOperationID: op.ID.String(),
		Transactions:      transactions,
		Signature:         "0x" + common.Bytes2Hex(signature),
	}

	opID, err := c.CreateOperation(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSendOperation, err)
	}

	// Retrieve the created operation
	operation, err := c.GetOperation(ctx, channelID, *opID)
	if err != nil {
		return nil, fmt.Errorf("%w: created but failed to retrieve: %w", ErrSendOperation, err)
	}

	return operation, nil
}

// GetOperation retrieves a specific operation by its ID within a channel.
//
// Parameters:
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel containing the operation.
//   - operationID: The UUID of the operation to retrieve.
//
// Returns the operation or an error if the operation fails or is not found.
func (c *Client) GetOperation(ctx context.Context, channelID uuid.UUID, operationID uuid.UUID) (*apiClient.Operation, error) {
	c.logger.Debug("Getting operation",
		"channel_id", channelID.String(),
		"operation_id", operationID.String())

	resp, err := c.crecClient.GetChannelsChannelIdOperationsOperationIdWithResponse(ctx, channelID, operationID)
	if err != nil {
		c.logger.Error("Failed to get operation", "error", err)
		return nil, fmt.Errorf("%w: %w", ErrGetOperation, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Operation not found",
			"channel_id", channelID.String(),
			"operation_id", operationID.String())
		return nil, fmt.Errorf("%w: operation ID %s in channel %s", ErrOperationNotFound, operationID.String(), channelID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when getting operation",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, fmt.Errorf("%w: %w (status code %d)", ErrGetOperation, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("%w: %w", ErrGetOperation, ErrNilResponseBody)
	}

	c.logger.Debug("Operation retrieved successfully",
		"operation_id", resp.JSON200.OperationId.String(),
		"status", resp.JSON200.Status)

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
	Status        *[]apiClient.OperationStatus
	ChainSelector *string
	Address       *string
	WalletID      *uuid.UUID
	Limit         *int
	Offset        *int64
}

// ListOperations retrieves a list of operations for a channel.
//
// Parameters:
//   - ctx: The context for the request.
//   - input: The list parameters including filters and pagination.
//
// Returns a list of operations and a boolean indicating if there are more results.
func (c *Client) ListOperations(ctx context.Context, input ListOperationsInput) ([]apiClient.Operation, bool, error) {
	c.logger.Debug("Listing operations",
		"channel_id", input.ChannelID.String(),
		"filters", input)

	if input.ChannelID == uuid.Nil {
		return nil, false, ErrChannelIDRequired
	}

	params := apiClient.GetChannelsChannelIdOperationsParams{
		ChainSelector: input.ChainSelector,
		Address:       input.Address,
		WalletId:      input.WalletID,
		Limit:         input.Limit,
		Offset:        input.Offset,
		Status:        input.Status,
	}

	resp, err := c.crecClient.GetChannelsChannelIdOperationsWithResponse(ctx, input.ChannelID, &params)
	if err != nil {
		c.logger.Error("Failed to list operations", "error", err)
		return nil, false, fmt.Errorf("%w: %w", ErrListOperations, err)
	}

	if resp.StatusCode() == 404 {
		c.logger.Warn("Channel not found", "channel_id", input.ChannelID.String())
		return nil, false, fmt.Errorf("%w: channel ID %s", ErrChannelNotFound, input.ChannelID.String())
	}

	if resp.StatusCode() != 200 {
		c.logger.Error("Unexpected status code when listing operations",
			"status_code", resp.StatusCode(),
			"body", string(resp.Body))
		return nil, false, fmt.Errorf("%w: %w (status code %d)", ErrListOperations, ErrUnexpectedStatusCode, resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, false, fmt.Errorf("%w: %w", ErrListOperations, ErrNilResponseBody)
	}

	c.logger.Debug("Operations listed successfully",
		"count", len(resp.JSON200.Data),
		"has_more", resp.JSON200.HasMore)

	return resp.JSON200.Data, resp.JSON200.HasMore, nil
}

// ExecuteTransactions executes a list of transactions using the provided signer and executor account.
// It bundles the transactions into an operation, signs it, and sends it to the CREC system.
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to send the operation to.
//   - operationSigner: The signer to use for signing the operation.
//   - executorAccount: The account to use for executing the operation.
//   - txs: The transactions to execute.
//   - chainSelector: The chain selector of the blockchain network in which the transactions are being executed.
func (c *Client) ExecuteTransactions(
	ctx context.Context,
	channelID uuid.UUID,
	operationSigner signer.Signer,
	executorAccount common.Address,
	txs []types.Transaction,
	chainSelector string,
) (*apiClient.Operation, error) {
	operation := &types.Operation{
		ID:           big.NewInt(time.Now().Unix()),
		Account:      executorAccount,
		Transactions: txs,
	}

	return c.ExecuteOperation(ctx, channelID, operationSigner, operation, chainSelector)
}

// ExecuteOperation signs and sends an operation to the CREC system.
//   - ctx: The context for the request.
//   - channelID: The UUID of the channel to send the operation to.
//   - operationSigner: The signer to use for signing the operation.
//   - operation: The operation to execute.
//   - chainSelector: The chain selector of the blockchain network in which the operation is being executed.
func (c *Client) ExecuteOperation(
	ctx context.Context, channelID uuid.UUID, operationSigner signer.Signer, operation *types.Operation, chainSelector string,
) (*apiClient.Operation, error) {
	_, sig, err := c.SignOperation(ctx, operation, operationSigner, chainSelector)
	if err != nil {
		return nil, err
	}

	opr, err := c.SendSignedOperation(ctx, channelID, operation, sig, chainSelector)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("ExecuteOperation: operation sent successfully",
		"channel_id", channelID.String(),
		"chain_selector", chainSelector,
		"operation_id", operation.ID.String(),
		"account", operation.Account.Hex())
	return opr, nil
}
