package identityvalidator

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/identityvalidator/gen/events"
	"github.com/smartcontractkit/cvn-sdk/services/ace/ccid/identityvalidator/gen/identityvalidator"
	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"
)

// ServiceOptions defines the options for creating a new ACE CCID Identity Validator service.
//   - Logger: Optional logger instance.
//   - IdentityValidatorAddress: A string representing the address of the Identity Validator contract.
//   - AccountAddress: A string representing the account address for operations.
type ServiceOptions struct {
	Logger                   *zerolog.Logger
	IdentityValidatorAddress string
	AccountAddress           string
}

// Service represents the ACE CCID Identity Validator service.
type Service struct {
	logger                   zerolog.Logger
	identityValidatorAddress common.Address
	accountAddress           common.Address
}

// CredentialRequirementInput represents the input structure for credential requirements.
type CredentialRequirementInput struct {
	RequirementId     [32]byte
	CredentialTypeIds [][32]byte
	MinValidations    *big.Int
	Invert            bool
}

// CredentialSourceInput represents the input structure for credential sources.
type CredentialSourceInput struct {
	CredentialTypeId   [32]byte
	IdentityRegistry   common.Address
	CredentialRegistry common.Address
	DataValidator      common.Address
}

// NewService creates a new ACE CCID Identity Validator service instance.
func NewService(opts *ServiceOptions) (*Service, error) {
	var logger zerolog.Logger
	if opts.Logger != nil {
		logger = *opts.Logger
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &Service{
		logger:                   logger,
		identityValidatorAddress: common.HexToAddress(opts.IdentityValidatorAddress),
		accountAddress:           common.HexToAddress(opts.AccountAddress),
	}, nil
}

// Event decoding methods

// DecodeCredentialRequirementAdded decodes a credential requirement added event from the provided CVN event.
func (s *Service) DecodeCredentialRequirementAdded(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var validatorEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &validatorEvent)
	if err != nil {
		return nil, err
	}

	return &validatorEvent, nil
}

// DecodeCredentialRequirementRemoved decodes a credential requirement removed event from the provided CVN event.
func (s *Service) DecodeCredentialRequirementRemoved(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var validatorEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &validatorEvent)
	if err != nil {
		return nil, err
	}

	return &validatorEvent, nil
}

// DecodeCredentialSourceAdded decodes a credential source added event from the provided CVN event.
func (s *Service) DecodeCredentialSourceAdded(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var validatorEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &validatorEvent)
	if err != nil {
		return nil, err
	}

	return &validatorEvent, nil
}

// DecodeCredentialSourceRemoved decodes a credential source removed event from the provided CVN event.
func (s *Service) DecodeCredentialSourceRemoved(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var validatorEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &validatorEvent)
	if err != nil {
		return nil, err
	}

	return &validatorEvent, nil
}

// DecodeInitialized decodes an initialized event from the provided CVN event.
func (s *Service) DecodeInitialized(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var initEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &initEvent)
	if err != nil {
		return nil, err
	}

	return &initEvent, nil
}

// DecodeOwnershipTransferred decodes an ownership transferred event from the provided CVN event.
func (s *Service) DecodeOwnershipTransferred(event *client.Event) (*events.IdentityValidatorEvents, error) {
	jsonBytes, err := s.toJson(event)
	if err != nil {
		return nil, err
	}

	var ownershipEvent events.IdentityValidatorEvents
	err = json.Unmarshal(jsonBytes, &ownershipEvent)
	if err != nil {
		return nil, err
	}

	return &ownershipEvent, nil
}

// Operation preparation methods

// PrepareAddCredentialRequirementOperation prepares a credential requirement addition operation.
// It constructs the necessary transaction to add a credential requirement.
//   - input: The credential requirement input data.
func (s *Service) PrepareAddCredentialRequirementOperation(input CredentialRequirementInput) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
		return nil, err
	}

	// Convert to the struct expected by the ABI
	abiInput := struct {
		RequirementId     [32]byte
		CredentialTypeIds [][32]byte
		MinValidations    *big.Int
		Invert            bool
	}{
		RequirementId:     input.RequirementId,
		CredentialTypeIds: input.CredentialTypeIds,
		MinValidations:    input.MinValidations,
		Invert:            input.Invert,
	}

	calldata, err := abiEncoder.Pack("addCredentialRequirement", abiInput)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for addCredentialRequirement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareAddCredentialSourceOperation prepares a credential source addition operation.
// It constructs the necessary transaction to add a credential source.
//   - input: The credential source input data.
func (s *Service) PrepareAddCredentialSourceOperation(input CredentialSourceInput) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
		return nil, err
	}

	// Convert to the struct expected by the ABI
	abiInput := struct {
		CredentialTypeId   [32]byte
		IdentityRegistry   common.Address
		CredentialRegistry common.Address
		DataValidator      common.Address
	}{
		CredentialTypeId:   input.CredentialTypeId,
		IdentityRegistry:   input.IdentityRegistry,
		CredentialRegistry: input.CredentialRegistry,
		DataValidator:      input.DataValidator,
	}

	calldata, err := abiEncoder.Pack("addCredentialSource", abiInput)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for addCredentialSource")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRemoveCredentialRequirementOperation prepares a credential requirement removal operation.
// It constructs the necessary transaction to remove a credential requirement.
//   - requirementId: The requirement ID to remove.
func (s *Service) PrepareRemoveCredentialRequirementOperation(requirementId [32]byte) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("removeCredentialRequirement", requirementId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for removeCredentialRequirement")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRemoveCredentialSourceOperation prepares a credential source removal operation.
// It constructs the necessary transaction to remove a credential source.
//   - credentialTypeId: The credential type ID.
//   - identityRegistry: The identity registry address.
//   - credentialRegistry: The credential registry address.
func (s *Service) PrepareRemoveCredentialSourceOperation(
	credentialTypeId [32]byte,
	identityRegistry common.Address,
	credentialRegistry common.Address,
) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("removeCredentialSource", credentialTypeId, identityRegistry, credentialRegistry)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for removeCredentialSource")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareInitializeOperation prepares an initialization operation.
// It constructs the necessary transaction to initialize the validator with credential sources and requirements.
//   - credentialSourceInputs: Array of credential source inputs.
//   - credentialRequirementInputs: Array of credential requirement inputs.
func (s *Service) PrepareInitializeOperation(
	credentialSourceInputs []CredentialSourceInput,
	credentialRequirementInputs []CredentialRequirementInput,
) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
		return nil, err
	}

	// Convert to the structs expected by the ABI
	abiSourceInputs := make([]struct {
		CredentialTypeId   [32]byte
		IdentityRegistry   common.Address
		CredentialRegistry common.Address
		DataValidator      common.Address
	}, len(credentialSourceInputs))

	for i, input := range credentialSourceInputs {
		abiSourceInputs[i] = struct {
			CredentialTypeId   [32]byte
			IdentityRegistry   common.Address
			CredentialRegistry common.Address
			DataValidator      common.Address
		}{
			CredentialTypeId:   input.CredentialTypeId,
			IdentityRegistry:   input.IdentityRegistry,
			CredentialRegistry: input.CredentialRegistry,
			DataValidator:      input.DataValidator,
		}
	}

	abiRequirementInputs := make([]struct {
		RequirementId     [32]byte
		CredentialTypeIds [][32]byte
		MinValidations    *big.Int
		Invert            bool
	}, len(credentialRequirementInputs))

	for i, input := range credentialRequirementInputs {
		abiRequirementInputs[i] = struct {
			RequirementId     [32]byte
			CredentialTypeIds [][32]byte
			MinValidations    *big.Int
			Invert            bool
		}{
			RequirementId:     input.RequirementId,
			CredentialTypeIds: input.CredentialTypeIds,
			MinValidations:    input.MinValidations,
			Invert:            input.Invert,
		}
	}

	calldata, err := abiEncoder.Pack("initialize", abiSourceInputs, abiRequirementInputs)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for initialize")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareTransferOwnershipOperation prepares an ownership transfer operation.
// It constructs the necessary transaction to transfer ownership of the validator.
//   - newOwner: The address of the new owner.
func (s *Service) PrepareTransferOwnershipOperation(newOwner common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
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
				To:    s.identityValidatorAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRenounceOwnershipOperation prepares an ownership renouncement operation.
// It constructs the necessary transaction to renounce ownership of the validator.
func (s *Service) PrepareRenounceOwnershipOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := identityvalidator.IdentityvalidatorMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get Identity Validator ABI")
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
				To:    s.identityValidatorAddress,
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
