package ccip

import (
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	transactTypes "github.com/smartcontractkit/cvn-sdk/transactor/types"

	"github.com/smartcontractkit/cvn-sdk/services/ccip/gen/router"
)

type CcipServiceOptions struct {
	Logger            *zerolog.Logger
	CcipRouterAddress string
	AccountAddress    string
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

	logger.Info().Msg("Creating CVN DvP service")

	return &CcipService{
		logger:  logger,
		options: opts,
	}
}

func (s *CcipService) PrepareRouteMessageOperation(
	message router.ClientAny2EVMMessage, gasForCallExactCheck uint16, gasLimit *big.Int, receiver common.Address,
) (*transactTypes.Operation, error) {
	abi, err := router.RouterMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get CCIP router ABI")
		return nil, err
	}

	routeMessageCalldata, err := abi.Pack("routeMessage", message, gasForCallExactCheck, gasLimit, receiver)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for routeMessage")
		return nil, err
	}

	ccipRouter := common.HexToAddress(s.options.CcipRouterAddress)
	account := common.HexToAddress(s.options.AccountAddress)

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: &account,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &ccipRouter,
				Value: big.NewInt(0),
				Data:  routeMessageCalldata,
			},
		},
	}, nil
}
