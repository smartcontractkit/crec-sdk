package dvp

import (
	"encoding/base64"
	"errors"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/interfaces/holdmanager"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"

	"github.com/smartcontractkit/cvn-sdk/interfaces/erc20"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/contract"
	"github.com/smartcontractkit/cvn-sdk/services/dvp/gen/events"
)

const (
	TokenTypeNone = iota
	TokenTypeERC20
	TokenTypeERC3643
)

const (
	ServiceName         = "dvp"
	SettlementOpened    = "SettlementOpened"
	SettlementAccepted  = "SettlementAccepted"
	SettlementCanceling = "SettlementCanceling"
	SettlementCanceled  = "SettlementCanceled"
	SettlementClosing   = "SettlementClosing"
	SettlementSettled   = "SettlementSettled"
)

// ServiceOptions defines the options for creating a new CVN DvP service.
//   - Logger: Optional logger instance.
//   - DvpCoordinatorAddress: A string representing the address of the DvP coordinator contract.
//   - AccountAddress: A string representing the address of the account performing the DvP operations.
type ServiceOptions struct {
	Logger                *zerolog.Logger
	DvpCoordinatorAddress string
	AccountAddress        string
}

type Service struct {
	logger                *zerolog.Logger
	dvpCoordinatorAddress common.Address
	accountAddress        common.Address
}

// NewService creates a new CVN DvP service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CVN DvP service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, errors.New("options must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN DvP service")

	return &Service{
		logger:                logger,
		dvpCoordinatorAddress: common.HexToAddress(opts.DvpCoordinatorAddress),
		accountAddress:        common.HexToAddress(opts.AccountAddress),
	}, nil
}

// DecodeDvpEvent decodes a DvP settlement event from the provided CVN event.
func (s *Service) DecodeDvpEvent(event *client.Event) (
	*events.DvpEvent, error,
) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var dvpEvent events.DvpEvent
	err = dvpEvent.UnmarshalJSON(jsonBytes)
	if err != nil {
		return nil, err
	}

	return &dvpEvent, nil
}

// PrepareProposeSettlementOperation prepares a DvP propose settlement operation. It assumes a token approval
// has already been issued for the asset token.
//   - settlement: The settlement details to be included in the operation.
func (s *Service) PrepareProposeSettlementOperation(settlement *contract.Settlement) (
	*transactTypes.Operation, error,
) {
	settlementHash, err := s.HashSettlement(settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash settlement")
		return nil, err
	}

	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("proposeSettlement", settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for proposeSettlement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      settlementHash.Big(),
		Account: &s.accountAddress,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &s.dvpCoordinatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareProposeSettlementWithTokenApprovalOperation prepares a DvP propose settlement operation, including
// a token approval.
//   - settlement: The settlement details to be included in the operation.
func (s *Service) PrepareProposeSettlementWithTokenApprovalOperation(settlement *contract.Settlement) (
	*transactTypes.Operation, error,
) {
	settlementHash, err := s.HashSettlement(settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash settlement")
		return nil, err
	}

	approveTransaction, err := s.prepareTokenApproveTransaction(
		&settlement.TokenInfo.AssetTokenSourceAddress, settlement.TokenInfo.AssetTokenAmount,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to prepare approve transaction")
		return nil, err
	}

	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("proposeSettlement", settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for proposeSettlement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      settlementHash.Big(),
		Account: &s.accountAddress,
		Transactions: []*transactTypes.Transaction{
			approveTransaction,
			{
				To:    &s.dvpCoordinatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareProposeSettlementWithTokenHoldOperation prepares a DvP propose settlement operation, including
// a issuing a token hold for the asset token.
//   - settlement: The settlement details to be included in the operation.
//   - holdManagerAddress: The address of the hold manager contract to be used for the token hold.
func (s *Service) PrepareProposeSettlementWithTokenHoldOperation(
	settlement *contract.Settlement, holdManagerAddress *common.Address,
) (
	*transactTypes.Operation, error,
) {
	if settlement.TokenInfo.AssetTokenType != TokenTypeERC3643 {
		s.logger.Error().Msg("Token hold is only supported for ERC3643 asset tokens")
		return nil, errors.New("token hold is only supported for ERC3643 asset tokens")
	}

	settlementHash, err := s.HashSettlement(settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash settlement")
		return nil, err
	}

	holdTransaction, err := s.prepareTokenHoldTransaction(
		holdManagerAddress,
		settlementHash,
		&settlement.TokenInfo.AssetTokenSourceAddress,
		&settlement.PartyInfo.SellerSourceAddress,
		settlement.Expiration,
		settlement.TokenInfo.AssetTokenAmount,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to prepare hold transaction")
		return nil, err
	}

	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("proposeSettlement", settlement)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for proposeSettlement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      settlementHash.Big(),
		Account: &s.accountAddress,
		Transactions: []*transactTypes.Transaction{
			holdTransaction,
			{
				To:    &s.dvpCoordinatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareAcceptSettlementOperation prepares a DvP accept settlement operation.
// It constructs the necessary transaction to accept a settlement based on its hash.
//   - settlementHash: The hash of the settlement to be accepted.
func (s *Service) PrepareAcceptSettlementOperation(settlementHash *common.Hash) (*transactTypes.Operation, error) {

	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("acceptSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for acceptSettlement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      settlementHash.Big(),
		Account: &s.accountAddress,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &s.dvpCoordinatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareExecuteSettlementOperation prepares a DvP execute settlement operation.
// It constructs the necessary transaction to execute a settlement based on its hash.
//   - settlementHash: The hash of the settlement to be executed.
func (s *Service) PrepareExecuteSettlementOperation(settlementHash *common.Hash) (*transactTypes.Operation, error) {

	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DVP ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("executeSettlement", settlementHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for executeSettlement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      settlementHash.Big(),
		Account: &s.accountAddress,
		Transactions: []*transactTypes.Transaction{
			{
				To:    &s.dvpCoordinatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// HashSettlement computes the hash of a DvP settlement.
//   - settlement: The settlement to be hashed.
func (s *Service) HashSettlement(settlement *contract.Settlement) (*common.Hash, error) {
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
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	hash := crypto.Keccak256Hash(settlementInfoData)

	return &hash, nil
}

// toJson decodes an encoded VerifiableEvent from a CVN event into a JSON byte slice.
func (s *Service) toJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (s *Service) prepareTokenApproveTransaction(
	tokenAddress *common.Address, tokenAmount *big.Int,
) (*transactTypes.Transaction, error) {
	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get ERC20 ABI")
		return nil, err
	}

	calldata, err := erc20Abi.Pack("approve", s.dvpCoordinatorAddress, tokenAmount)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for token approve")
		return nil, err
	}

	return &transactTypes.Transaction{
		To:    tokenAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}, nil
}

func (s *Service) prepareTokenHoldTransaction(
	holdManagerAddress *common.Address, holdId *common.Hash, tokenAddress *common.Address, sender *common.Address,
	expiresAt *big.Int, tokenAmount *big.Int,
) (*transactTypes.Transaction, error) {
	holdmanagerAbi, err := holdmanager.HoldmanagerMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get HoldManager ABI")
		return nil, err
	}

	hold := holdmanager.IHoldManagerHold{
		Token:     *tokenAddress,
		Sender:    *sender,
		Recipient: s.dvpCoordinatorAddress,
		Executor:  s.dvpCoordinatorAddress,
		ExpiresAt: expiresAt,
		Value:     tokenAmount,
	}

	calldata, err := holdmanagerAbi.Pack("createHold", *holdId, hold)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for token hold")
		return nil, err
	}

	return &transactTypes.Transaction{
		To:    holdManagerAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}, nil
}
