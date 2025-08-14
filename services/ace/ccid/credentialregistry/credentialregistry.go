package credentialregistry

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/credentialregistry/gen/credentialregistry"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/credentialregistry/gen/events"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"
)

// ServiceOptions defines the options for creating a new ACE CCID Credential Registry service.
//   - Logger: Optional logger instance.
//   - CredentialRegistryAddress: A string representing the address of the Credential Registry contract.
//   - AccountAddress: A string representing the account address for operations.
type ServiceOptions struct {
	Logger                    *zerolog.Logger
	CredentialRegistryAddress string
	AccountAddress            string
}

// Service represents the ACE CCID Credential Registry service.
type Service struct {
	logger                    zerolog.Logger
	credentialRegistryAddress common.Address
	accountAddress            common.Address
}

// NewService creates a new ACE CCID Credential Registry service instance.
func NewService(opts *ServiceOptions) (*Service, error) {
	var logger zerolog.Logger
	if opts.Logger != nil {
		logger = *opts.Logger
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &Service{
		logger:                    logger,
		credentialRegistryAddress: common.HexToAddress(opts.CredentialRegistryAddress),
		accountAddress:            common.HexToAddress(opts.AccountAddress),
	}, nil
}

// Event decoding methods

// DecodeCredentialRegistered decodes a credential registered event from the provided CVN event.
func (s *Service) DecodeCredentialRegistered(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var credentialEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &credentialEvent)
	if err != nil {
		return nil, err
	}

	return &credentialEvent, nil
}

// DecodeCredentialRemoved decodes a credential removed event from the provided CVN event.
func (s *Service) DecodeCredentialRemoved(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var credentialEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &credentialEvent)
	if err != nil {
		return nil, err
	}

	return &credentialEvent, nil
}

// DecodeCredentialRenewed decodes a credential renewed event from the provided CVN event.
func (s *Service) DecodeCredentialRenewed(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var credentialEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &credentialEvent)
	if err != nil {
		return nil, err
	}

	return &credentialEvent, nil
}

// DecodeInitialized decodes an initialized event from the provided CVN event.
func (s *Service) DecodeInitialized(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var initEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &initEvent)
	if err != nil {
		return nil, err
	}

	return &initEvent, nil
}

// DecodeOwnershipTransferred decodes an ownership transferred event from the provided CVN event.
func (s *Service) DecodeOwnershipTransferred(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var ownershipEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &ownershipEvent)
	if err != nil {
		return nil, err
	}

	return &ownershipEvent, nil
}

// DecodePolicyEngineAttached decodes a policy engine attached event from the provided CVN event.
func (s *Service) DecodePolicyEngineAttached(event *client.Event) (*events.CredentialRegistryEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var policyEvent events.CredentialRegistryEvents
	err = json.Unmarshal(jsonBytes, &policyEvent)
	if err != nil {
		return nil, err
	}

	return &policyEvent, nil
}

// Operation preparation methods

// PrepareRegisterCredentialOperation prepares a credential registration operation.
// It constructs the necessary transaction to register a single credential.
//   - ccid: The CCID (bytes32) of the identity.
//   - credentialTypeId: The credential type ID (bytes32).
//   - expiresAt: The expiration timestamp (uint40).
//   - credentialData: The credential data (bytes).
//   - context: Additional context data for the registration.
func (s *Service) PrepareRegisterCredentialOperation(
	ccid [32]byte,
	credentialTypeId [32]byte,
	expiresAt uint64,
	credentialData []byte,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerCredential", ccid, credentialTypeId, big.NewInt(int64(expiresAt)), credentialData, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerCredential")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.credentialRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRegisterCredentialsOperation prepares a batch credential registration operation.
// It constructs the necessary transaction to register multiple credentials at once.
//   - ccid: The CCID (bytes32) of the identity.
//   - credentialTypeIds: Array of credential type IDs (bytes32).
//   - expiresAt: The expiration timestamp (uint40) for all credentials.
//   - credentialDatas: Array of credential data (bytes).
//   - context: Additional context data for the registration.
func (s *Service) PrepareRegisterCredentialsOperation(
	ccid [32]byte,
	credentialTypeIds [][32]byte,
	expiresAt uint64,
	credentialDatas [][]byte,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerCredentials", ccid, credentialTypeIds, big.NewInt(int64(expiresAt)), credentialDatas, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerCredentials")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.credentialRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRemoveCredentialOperation prepares a credential removal operation.
// It constructs the necessary transaction to remove a credential.
//   - ccid: The CCID (bytes32) of the identity.
//   - credentialTypeId: The credential type ID (bytes32) to remove.
//   - context: Additional context data for the removal.
func (s *Service) PrepareRemoveCredentialOperation(
	ccid [32]byte,
	credentialTypeId [32]byte,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("removeCredential", ccid, credentialTypeId, context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for removeCredential")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.credentialRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRenewCredentialOperation prepares a credential renewal operation.
// It constructs the necessary transaction to renew a credential's expiration.
//   - ccid: The CCID (bytes32) of the identity.
//   - credentialTypeId: The credential type ID (bytes32) to renew.
//   - expiresAt: The new expiration timestamp (uint40).
//   - context: Additional context data for the renewal.
func (s *Service) PrepareRenewCredentialOperation(
	ccid [32]byte,
	credentialTypeId [32]byte,
	expiresAt uint64,
	context []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("renewCredential", ccid, credentialTypeId, big.NewInt(int64(expiresAt)), context)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for renewCredential")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.credentialRegistryAddress,
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
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
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
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareClearContextOperation prepares a context clearing operation.
// It constructs the necessary transaction to clear context data in the registry.
func (s *Service) PrepareClearContextOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
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
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
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
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRenounceOwnershipOperation prepares an ownership renouncement operation.
// It constructs the necessary transaction to renounce ownership of the registry.
func (s *Service) PrepareRenounceOwnershipOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := credentialregistry.CredentialregistryMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Credential Registry ABI")
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
				To:    s.credentialRegistryAddress,
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
