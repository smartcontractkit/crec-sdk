package ccip

import (
	"errors"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"

	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/erc20"
	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/routerclient"
)

// CcipServiceOptions defines the options for creating a new CVN CCIP service.
//   - Logger: Optional logger instance.
//   - CcipRouterAddress: A string representing the address of the CCIP router contract.
//   - AccountAddress: A string representing the address of the account performing the CCIP operations.
//   - IncludeTokenApprovals: A boolean indicating whether to include token approvals prior to the CCIP send operation.
type CcipServiceOptions struct {
	Logger                *zerolog.Logger
	CcipRouterAddress     string
	AccountAddress        string
	IncludeTokenApprovals bool
}

type CcipService struct {
	logger                *zerolog.Logger
	ccipRouterAddress     string
	accountAddress        string
	includeTokenApprovals bool
}

// NewCcipService creates a new CVN CCIP service with the provided options.
// Returns a pointer to the CcipService and an error if any issues occur during initialization.
//   - opts: Options for configuring the CVN CCIP service, see CcipServiceOptions for details.
func NewCcipService(opts *CcipServiceOptions) (*CcipService, error) {
	if opts == nil {
		return nil, errors.New("options must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN CCIP service")

	return &CcipService{
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
func (s *CcipService) PrepareCcipSendOperation(
	destinationChainSelector uint64, message routerclient.ClientEVM2AnyMessage,
) (*transactTypes.Operation, error) {

	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get ERC20 ABI")
		return nil, err
	}

	routerClientAbi, err := routerclient.RouterclientMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get RouterClient ABI")
		return nil, err
	}

	ccipSendCalldata, err := routerClientAbi.Pack("ccipSend", destinationChainSelector, message)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for ccipSend")
		return nil, err
	}

	ccipRouter := common.HexToAddress(s.ccipRouterAddress)
	account := common.HexToAddress(s.accountAddress)

	var transactions []*transactTypes.Transaction

	if s.includeTokenApprovals && len(message.TokenAmounts) > 0 {
		// If the message contains token amounts, we need to approve tokens to the CCIP router
		for _, destTokenAmount := range message.TokenAmounts {
			approveCalldata, err := erc20Abi.Pack("approve", ccipRouter, destTokenAmount.Amount)
			if err != nil {
				s.logger.Error().Err(err).Msg("failed to pack calldata for token approve")
				return nil, err
			}

			transactions = append(
				transactions, &transactTypes.Transaction{
					To:    &destTokenAmount.Token,
					Value: big.NewInt(0),
					Data:  approveCalldata,
				},
			)
		}
	}

	transactions = append(
		transactions, &transactTypes.Transaction{
			To:    &ccipRouter,
			Value: big.NewInt(0),
			Data:  ccipSendCalldata,
		},
	)

	return &transactTypes.Operation{
		ID:           big.NewInt(time.Now().Unix()),
		Account:      &account,
		Transactions: transactions,
	}, nil
}
