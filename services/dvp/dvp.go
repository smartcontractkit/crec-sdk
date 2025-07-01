package dvp

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"

	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/contract"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/events"
)

type DvpServiceOptions struct {
	Logger                *zerolog.Logger
	DvpCoordinatorAddress string
	AccountAddress        string
}

type DvpService struct {
	logger  *zerolog.Logger
	options *DvpServiceOptions
}

func NewDvpService(opts *DvpServiceOptions) *DvpService {
	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN DvP service")

	return &DvpService{
		logger:  logger,
		options: opts,
	}
}

func (s *DvpService) DecodeSettlementAccepted(event *client.Event) (
	*events.DvpSettlementAccepted, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementAccepted
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

func (s *DvpService) toJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (s *DvpService) PrepareExecuteSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {

	abi, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	executeSettlementCalldata, err := abi.Pack("executeSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for executeSettlement")
		return nil, err
	}

	dvpCoordinator := common.HexToAddress(s.options.DvpCoordinatorAddress)
	account := common.HexToAddress(s.options.AccountAddress)

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: &account,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &dvpCoordinator,
				Value: big.NewInt(0),
				Data:  executeSettlementCalldata,
			},
		},
	}, nil
}
