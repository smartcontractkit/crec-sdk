package dvp

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	eventTypes "github.com/smartcontractkit/cvn-sdk/events/types"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/contract"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/events"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transactor/types"
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

func (s *DvpService) DecodeSettlementAccepted(verifiableEvent *eventTypes.VerifiableEvent) (
	*events.DvpSettlementAccepted, error,
) {
	var event events.DvpSettlementAccepted
	jsonBytes, err := s.toJson(verifiableEvent)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *DvpService) toJson(verifiableEvent *eventTypes.VerifiableEvent) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(verifiableEvent.Payload)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (s *DvpService) PrepareExecuteSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {

	abi, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get DVP ABI")
		return nil, err
	}

	executeSettlementCalldata, err := abi.Pack("executeSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for executeSettlement")
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
