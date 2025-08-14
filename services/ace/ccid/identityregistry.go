package ccid

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/gen/events"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/gen/identityregistry"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"
)

// ServiceOptions defines the options for creating a new ACE CCID Identity Registry service.
//   - Logger: Optional logger instance.
//   - IdentityRegistryAddress: A string representing the address of the Identity Registry contract.
//   - AccountAddress: A string representing the account address for operations.
type ServiceOptions struct {
	Logger                  *zerolog.Logger
	IdentityRegistryAddress string
	AccountAddress          string
}

// Service represents the ACE CCID Identity Registry service.
type Service struct {
	logger                  zerolog.Logger
	identityRegistryAddress common.Address
	accountAddress          common.Address
}

// NewService creates a new ACE CCID Identity Registry service instance.
func NewService(opts *ServiceOptions) (*Service, error) {
	var logger zerolog.Logger
	if opts.Logger != nil {
		logger = *opts.Logger
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &Service{
		logger:                  logger,
		identityRegistryAddress: common.HexToAddress(opts.IdentityRegistryAddress),
		accountAddress:          common.HexToAddress(opts.AccountAddress),
	}, nil
}

// Event decoding methods

// DecodeIdentityRegistered decodes an identity registered event from the provided CVN event.
func (s *Service) DecodeIdentityRegistered(event *client.Event) (*events.IdentityRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var identityEvent events.IdentityRegistryEvents
	err = json.Unmarshal(jsonBytes, &identityEvent)
	if err != nil {
		return nil, err
	}

	return &identityEvent, nil
}

// DecodeIdentityRemoved decodes an identity removed event from the provided CVN event.
func (s *Service) DecodeIdentityRemoved(event *client.Event) (*events.IdentityRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var identityEvent events.IdentityRegistryEvents
	err = json.Unmarshal(jsonBytes, &identityEvent)
	if err != nil {
		return nil, err
	}

	return &identityEvent, nil
}

// DecodeInitialized decodes an initialized event from the provided CVN event.
func (s *Service) DecodeInitialized(event *client.Event) (*events.IdentityRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var initEvent events.IdentityRegistryEvents
	err = json.Unmarshal(jsonBytes, &initEvent)
	if err != nil {
		return nil, err
	}

	return &initEvent, nil
}

// DecodeOwnershipTransferred decodes an ownership transferred event from the provided CVN event.
func (s *Service) DecodeOwnershipTransferred(event *client.Event) (*events.IdentityRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var ownershipEvent events.IdentityRegistryEvents
	err = json.Unmarshal(jsonBytes, &ownershipEvent)
	if err != nil {
		return nil, err
	}

	return &ownershipEvent, nil
}

// DecodePolicyEngineAttached decodes a policy engine attached event from the provided CVN event.
func (s *Service) DecodePolicyEngineAttached(event *client.Event) (*events.IdentityRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var policyEvent events.IdentityRegistryEvents
	err = json.Unmarshal(jsonBytes, &policyEvent)
	if err != nil {
		return nil, err
	}

	return &policyEvent, nil
}

// Operation preparation methods

// PrepareRegisterIdentityOperation prepares an identity registration operation.
// It constructs the necessary transaction to register a single identity.
//   - ccid: The CCID (bytes32) of the identity to register.
//   - account: The account address to associate with the identity.
//   - context: Additional context data for the registration.
func (s *Service) PrepareRegisterIdentityOperation(
	ccid [32]byte,
	account common.Address,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerIdentity", ccid, account, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerIdentity")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRegisterIdentitiesOperation prepares a batch identity registration operation.
// It constructs the necessary transaction to register multiple identities at once.
//   - ccids: Array of CCIDs (bytes32) of the identities to register.
//   - accounts: Array of account addresses to associate with the identities.
//   - context: Additional context data for the registration.
func (s *Service) PrepareRegisterIdentitiesOperation(
	ccids [][32]byte,
	accounts []common.Address,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerIdentities", ccids, accounts, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerIdentities")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRemoveIdentityOperation prepares an identity removal operation.
// It constructs the necessary transaction to remove an identity.
//   - ccid: The CCID (bytes32) of the identity to remove.
//   - account: The account address to remove from the identity.
//   - context: Additional context data for the removal.
func (s *Service) PrepareRemoveIdentityOperation(
	ccid [32]byte,
	account common.Address,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("removeIdentity", ccid, account, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for removeIdentity")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareAttachPolicyEngineOperation prepares a policy engine attachment operation.
// It constructs the necessary transaction to attach a policy engine to the registry.
//   - policyEngine: The address of the policy engine to attach.
func (s *Service) PrepareAttachPolicyEngineOperation(policyEngine common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("attachPolicyEngine", policyEngine)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for attachPolicyEngine")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareSetContextOperation prepares a context setting operation.
// It constructs the necessary transaction to set context data in the registry.
//   - context: The context data to set.
func (s *Service) PrepareSetContextOperation(context []byte) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("setContext", context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for setContext")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareClearContextOperation prepares a context clearing operation.
// It constructs the necessary transaction to clear context data in the registry.
func (s *Service) PrepareClearContextOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("clearContext")
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for clearContext")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareInitializeOperation prepares an initialization operation.
// It constructs the necessary transaction to initialize the registry with a policy engine.
//   - policyEngine: The address of the policy engine to initialize with.
func (s *Service) PrepareInitializeOperation(policyEngine common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("initialize", policyEngine)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for initialize")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareTransferOwnershipOperation prepares an ownership transfer operation.
// It constructs the necessary transaction to transfer ownership of the registry.
//   - newOwner: The address of the new owner.
func (s *Service) PrepareTransferOwnershipOperation(newOwner common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("transferOwnership", newOwner)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for transferOwnership")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRenounceOwnershipOperation prepares an ownership renouncement operation.
// It constructs the necessary transaction to renounce ownership of the registry.
func (s *Service) PrepareRenounceOwnershipOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := identityregistry.IdentityregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("renounceOwnership")
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for renounceOwnership")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
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
