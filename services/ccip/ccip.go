package ccip

import (
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	transactTypes "github.com/smartcontractkit/cvn-sdk/transactor/types"

	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/erc20"
	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/router"
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

func (s *CcipService) PrepareRouteMessageOperation(
	message router.ClientAny2EVMMessage, gasForCallExactCheck uint16, gasLimit *big.Int, receiver common.Address,
) (*transactTypes.Operation, error) {

	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get ERC20 ABI")
		return nil, err
	}

	routerAbi, err := router.RouterMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get CCIP router ABI")
		return nil, err
	}

	routeMessageCalldata, err := routerAbi.Pack("routeMessage", message, gasForCallExactCheck, gasLimit, receiver)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for routeMessage")
		return nil, err
	}

	ccipRouter := common.HexToAddress(s.options.CcipRouterAddress)
	account := common.HexToAddress(s.options.AccountAddress)

	var transactions []*transactTypes.Transaction

	if s.options.IncludeTokenApprovals && len(message.DestTokenAmounts) > 0 {
		// If the message contains token amounts, we need to approve tokens to the CCIP router
		for _, destTokenAmount := range message.DestTokenAmounts {
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
			Data:  routeMessageCalldata,
		},
	)

	return &transactTypes.Operation{
		ID:           big.NewInt(time.Now().Unix()),
		Account:      &account,
		Transactions: transactions,
	}, nil
}
