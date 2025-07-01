package dvp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"

	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/contract"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/events"
)

// DvpServiceOptions defines the options for creating a new CVN DvP service.
//   - Logger: Optional logger instance.
//   - DvpCoordinatorAddress: A string representing the address of the DvP coordinator contract.
//   - AccountAddress: A string representing the address of the account performing the DvP operations.
type DvpServiceOptions struct {
	Logger                *zerolog.Logger
	DvpCoordinatorAddress string
	AccountAddress        string
}

type DvpService struct {
	logger                *zerolog.Logger
	dvpCoordinatorAddress string
	accountAddress        string
}

// NewDvpService creates a new CVN DvP service with the provided options.
// Returns a pointer to the DvpService and an error if any issues occur during initialization.
//   - opts: Options for configuring the CVN DvP service, see DvpServiceOptions for details.
func NewDvpService(opts *DvpServiceOptions) (*DvpService, error) {
	if opts == nil {
		return nil, errors.New("options must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN DvP service")

	return &DvpService{
		logger:                logger,
		dvpCoordinatorAddress: opts.DvpCoordinatorAddress,
		accountAddress:        opts.AccountAddress,
	}, nil
}

// DecodeSettlementOpened decodes a DvP settlement opened event from the provided CVN event.
func (s *DvpService) DecodeSettlementOpened(event *client.Event) (
	*events.DvpSettlementOpened, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementOpened
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

// DecodeSettlementAccepted decodes a DvP settlement accepted event from the provided CVN event.
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

// DecodeSettlementClosing decodes a DvP settlement closing event from the provided CVN event.
func (s *DvpService) DecodeSettlementClosing(event *client.Event) (
	*events.DvpSettlementClosing, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementClosing
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

// DecodeSettlementSettled decodes a DvP settlement settled event from the provided CVN event.
func (s *DvpService) DecodeSettlementSettled(event *client.Event) (
	*events.DvpSettlementSettled, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementSettled
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

// DecodeSettlementCanceling decodes a DvP settlement canceling event from the provided CVN event.
func (s *DvpService) DecodeSettlementCanceling(event *client.Event) (
	*events.DvpSettlementCanceling, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementCanceling
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

// DecodeSettlementCanceled decodes a DvP settlement canceled event from the provided CVN event.
func (s *DvpService) DecodeSettlementCanceled(event *client.Event) (
	*events.DvpSettlementCanceled, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpSettlementCanceled
	err = json.Unmarshal(jsonBytes, &event)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

func (s *DvpService) PrepareAcceptSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {

	abi, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abi.Pack("acceptSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for acceptSettlement")
		return nil, err
	}

	dvpCoordinator := common.HexToAddress(s.dvpCoordinatorAddress)
	account := common.HexToAddress(s.accountAddress)

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: &account,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &dvpCoordinator,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

func (s *DvpService) PrepareExecuteSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {

	abi, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abi.Pack("executeSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for executeSettlement")
		return nil, err
	}

	dvpCoordinator := common.HexToAddress(s.dvpCoordinatorAddress)
	account := common.HexToAddress(s.accountAddress)

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: &account,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &dvpCoordinator,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

func (s *DvpService) HashSettlement(settlement *contract.Settlement) (common.Hash, error) {
	uint256Ty, err := abi.NewType("uint256", "", nil)
	uint64Ty, err := abi.NewType("uint64", "", nil)
	uint48Ty, err := abi.NewType("uint48", "", nil)
	uint8Ty, err := abi.NewType("uint8", "", nil)
	addressTy, err := abi.NewType("address", "", nil)
	bytes32Ty, err := abi.NewType("bytes32", "", nil)
	bytesTy, err := abi.NewType("bytes", "", nil)

	partyInfoArgs := abi.Arguments{
		{Type: addressTy}, // buyerSourceAddress
		{Type: addressTy}, // buyerDestinationAddress
		{Type: addressTy}, // sellerSourceAddress
		{Type: addressTy}, // sellerDestinationAddress
		{Type: addressTy}, // executorAddress
	}
	partyInfoData, err := partyInfoArgs.Pack(
		settlement.PartyInfo.BuyerSourceAddress,
		settlement.PartyInfo.BuyerDestinationAddress,
		settlement.PartyInfo.SellerSourceAddress,
		settlement.PartyInfo.SellerDestinationAddress,
		settlement.PartyInfo.ExecutorAddress,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack party info data")
		return common.Hash{}, err
	}

	tokenInfoArgs := abi.Arguments{
		{Type: uint256Ty}, // paymentTokenAmount
		{Type: uint256Ty}, // assetTokenAmount
		{Type: addressTy}, // paymentTokenSourceAddress
		{Type: addressTy}, // paymentTokenDestinationAddress
		{Type: addressTy}, // assetTokenSourceAddress
		{Type: addressTy}, // assetTokenDestinationAddress
		{Type: uint8Ty},   // paymentTokenType
		{Type: uint8Ty},   // assetTokenType
	}
	tokenInfoData, err := tokenInfoArgs.Pack(
		settlement.TokenInfo.PaymentTokenAmount,
		settlement.TokenInfo.AssetTokenAmount,
		settlement.TokenInfo.PaymentTokenSourceAddress,
		settlement.TokenInfo.PaymentTokenDestinationAddress,
		settlement.TokenInfo.AssetTokenSourceAddress,
		settlement.TokenInfo.AssetTokenDestinationAddress,
		settlement.TokenInfo.PaymentTokenType,
		settlement.TokenInfo.AssetTokenType,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack token info data")
		return common.Hash{}, err
	}

	deliveryInfoArgs := abi.Arguments{
		{Type: uint64Ty}, // paymentSourceChainSelector
		{Type: uint64Ty}, // paymentDestinationChainSelector
		{Type: uint64Ty}, // assetSourceChainSelector
		{Type: uint64Ty}, // assetDestinationChainSelector
	}
	deliveryInfoData, err := deliveryInfoArgs.Pack(
		settlement.DeliveryInfo.PaymentSourceChainSelector,
		settlement.DeliveryInfo.PaymentDestinationChainSelector,
		settlement.DeliveryInfo.AssetSourceChainSelector,
		settlement.DeliveryInfo.AssetDestinationChainSelector,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack delivery info data")
		return common.Hash{}, err
	}

	partyInfoHash := crypto.Keccak256Hash(partyInfoData)
	tokenInfoHash := crypto.Keccak256Hash(tokenInfoData)
	deliveryDataHash := crypto.Keccak256Hash(deliveryInfoData)

	settlementInfoArgs := abi.Arguments{
		{Type: bytes32Ty}, // keccak256("CCIP_DVP_COORDINATOR_V1_SETTLEMENT")
		{Type: uint256Ty}, // settlementId
		{Type: bytes32Ty}, // partyInfoHash
		{Type: bytes32Ty}, // tokenInfoHash
		{Type: bytes32Ty}, // deliveryDataHash
		{Type: bytes32Ty}, // secretHash
		{Type: uint48Ty},  // expiration
		{Type: bytesTy},   // data
	}
	settlementInfoData, err := settlementInfoArgs.Pack(
		crypto.Keccak256Hash([]byte("CCIP_DVP_COORDINATOR_V1_SETTLEMENT")),
		settlement.SettlementId,
		partyInfoHash,
		tokenInfoHash,
		deliveryDataHash,
		settlement.SecretHash,
		settlement.Expiration,
		settlement.Data,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack settlement info data")
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(settlementInfoData), nil
}

// toJson decodes an encoded VerifiableEvent from a CVN event into a JSON byte slice.
func (s *DvpService) toJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}
