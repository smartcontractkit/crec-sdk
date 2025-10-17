package ccip

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/crec-api-go/services/ccip/gen/routerclient"

	"github.com/smartcontractkit/crec-sdk/interfaces/erc20"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// ServiceOptions defines the options for creating a new CREC CCIP service.
//   - Logger: Optional logger instance.
//   - CcipRouterAddress: A string representing the address of the CCIP router contract.
//   - AccountAddress: A string representing the address of the account performing the CCIP operations.
//   - IncludeTokenApprovals: A boolean indicating whether to include token approvals prior to the CCIP send operation.
type ServiceOptions struct {
	Logger                *zerolog.Logger
	CcipRouterAddress     string
	AccountAddress        string
	IncludeTokenApprovals bool
}

type Service struct {
	logger                *zerolog.Logger
	ccipRouterAddress     string
	accountAddress        string
	includeTokenApprovals bool
}

// NewService creates a new CREC CCIP service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC CCIP service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, fmt.Errorf("ServiceOptions is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC CCIP service")

	return &Service{
		logger:                logger,
		ccipRouterAddress:     opts.CcipRouterAddress,
		accountAddress:        opts.AccountAddress,
		includeTokenApprovals: opts.IncludeTokenApprovals,
	}, nil
}

// PrepareCcipSendOperation prepares a CCIP send operation with the given destination chain selector and message.
// It constructs the necessary transactions, including token approvals if required.
//   - destinationChainSelector: The selector for the destination chain.
//   - message: The message to be sent via CCIP, containing token amounts and other data.
func (s *Service) PrepareCcipSendOperation(
	destinationChainSelector uint64, message *routerclient.ClientEVM2AnyMessage,
) (*transactTypes.Operation, error) {
	s.logger.Trace().
		Msg("Preparing ccipSend operation")

	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get ERC20 ABI: %w", err)
	}

	routerClientAbi, err := routerclient.RouterclientMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get RouterClient ABI: %w", err)
	}

	ccipSendCalldata, err := routerClientAbi.Pack("ccipSend", destinationChainSelector, message)
	if err != nil {
		return nil, fmt.Errorf("failed to pack calldata for ccipSend: %w", err)
	}

	ccipRouter := common.HexToAddress(s.ccipRouterAddress)
	account := common.HexToAddress(s.accountAddress)

	var transactions []transactTypes.Transaction

	if s.includeTokenApprovals && len(message.TokenAmounts) > 0 {
		// If the message contains token amounts, we need to approve tokens to the CCIP router
		for _, destTokenAmount := range message.TokenAmounts {
			approveCalldata, err := erc20Abi.Pack("approve", ccipRouter, destTokenAmount.Amount)
			if err != nil {
				return nil, fmt.Errorf("failed to pack calldata for token approve: %w", err)
			}

			transactions = append(
				transactions, transactTypes.Transaction{
					To:    destTokenAmount.Token,
					Value: big.NewInt(0),
					Data:  approveCalldata,
				},
			)
		}
	}

	transactions = append(
		transactions, transactTypes.Transaction{
			To:    ccipRouter,
			Value: big.NewInt(0),
			Data:  ccipSendCalldata,
		},
	)

	return &transactTypes.Operation{
		ID:           big.NewInt(time.Now().Unix()),
		Account:      account,
		Transactions: transactions,
	}, nil
}
