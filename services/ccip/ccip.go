package ccip

import (
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"

	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/erc20"
	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/routerclient"
)

type CcipServiceOptions struct {
	Logger                *zerolog.Logger
	CcipRouterAddress     string
	AccountAddress        string
	IncludeTokenApprovals bool
}

type CcipService struct {
	logger  *zerolog.Logger
	options *CcipServiceOptions
}

func NewCcipService(opts *CcipServiceOptions) *CcipService {
	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN CCIP service")

	return &CcipService{
		logger:  logger,
		options: opts,
	}
}

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

	ccipRouter := common.HexToAddress(s.options.CcipRouterAddress)
	account := common.HexToAddress(s.options.AccountAddress)

	var transactions []*transactTypes.Transaction

	if s.options.IncludeTokenApprovals && len(message.TokenAmounts) > 0 {
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
