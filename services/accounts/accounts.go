package accounts

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/crec-api-go/services/accounts/gen/accounts"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// Sentinel errors
var (
	// Service initialization errors
	ErrServiceOptionsRequired = errors.New("ServiceOptions is required")

	// ABI encoding errors
	ErrCreateAddressArrayType = errors.New("failed to create address array ABI type")
	ErrCreateRSAKeyArrayType  = errors.New("failed to create RSA key array ABI type")
	ErrEncodeSignerAddresses  = errors.New("failed to encode signer addresses")
	ErrEncodeRSAKeys          = errors.New("failed to encode RSA keys")
	ErrGetAccountABI          = errors.New("failed to get account ABI")
	ErrPackAccountData        = errors.New("failed to pack account data")

	// Operation preparation errors
	ErrPrepareDeployECDSA = errors.New("failed to prepare deploy ECDSA account operation")
	ErrPrepareDeployRSA   = errors.New("failed to prepare deploy RSA account operation")
	ErrGenerateAccountID  = errors.New("failed to generate account ID")
)

// ServiceOptions defines the options for creating a new CREC Accounts service.
//   - Logger: Optional logger instance.
//   - OperationExecutionAccount: A previously deployed smart wallet address that will execute operations to invoke the factory and deploy accounts.
//   - KeystoneForwarderAddress: A string representing the address of the Keystone forwarder contract.
//   - AccountFactoryAddress: A string representing the address of the account factory contract.
//   - ECDSASignatureVerifyingAccountImplAddress: A string representing the address of the ECDSA signature verifying account implementation contract.
//   - RSASignatureVerifyingAccountImplAddress: A string representing the address of the RSA signature verifying account implementation contract.
type ServiceOptions struct {
	Logger                                    *zerolog.Logger
	OperationExecutionAccount                 string
	KeystoneForwarderAddress                  string
	AccountFactoryAddress                     string
	ECDSASignatureVerifyingAccountImplAddress string
	RSASignatureVerifyingAccountImplAddress   string
}

// Service provides operations for managing CREC account creation and deployment.
type Service struct {
	logger                                    *zerolog.Logger
	operationExecutionAccount                 common.Address
	keystoneForwarderAddress                  common.Address
	accountFactoryAddress                     common.Address
	ecdsaSignatureVerifyingAccountImplAddress common.Address
	rsaSignatureVerifyingAccountImplAddress   common.Address
}

// NewService creates a new CREC Accounts service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Accounts service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, ErrServiceOptionsRequired
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC Accounts service")

	return &Service{
		logger:                                    logger,
		operationExecutionAccount:                 common.HexToAddress(opts.OperationExecutionAccount),
		keystoneForwarderAddress:                  common.HexToAddress(opts.KeystoneForwarderAddress),
		accountFactoryAddress:                     common.HexToAddress(opts.AccountFactoryAddress),
		ecdsaSignatureVerifyingAccountImplAddress: common.HexToAddress(opts.ECDSASignatureVerifyingAccountImplAddress),
		rsaSignatureVerifyingAccountImplAddress:   common.HexToAddress(opts.RSASignatureVerifyingAccountImplAddress),
	}, nil
}

// PrepareDeployNewECDSAAccountOperation prepares an operation to deploy a new ECDSA signature verifying account.
// The operation includes the account creation transaction with properly encoded allowed signers.
//   - accountOwnerAddress: The address of the account owner as a hex string.
//   - allowedSigners: A slice of allowed signer addresses as hex strings.
//   - accountId: A unique identifier for the account.
//
// Returns the operation and the predicted address where the account will be deployed.
func (s *Service) PrepareDeployNewECDSAAccountOperation(
	accountOwnerAddress string, allowedSigners []string, accountId string,
) (*transactTypes.Operation, common.Address, error) {
	s.logger.Trace().
		Str("operationExecutionAccount", s.operationExecutionAccount.Hex()).
		Str("accountOwnerAddress", accountOwnerAddress).
		Str("allowedSigners", fmt.Sprintf("%v", allowedSigners)).
		Str("accountId", accountId).
		Msg("Preparing deploy new ECDSA account operation")

	signerAddresses := make([]common.Address, len(allowedSigners))
	for i, signerStr := range allowedSigners {
		signerAddresses[i] = common.HexToAddress(signerStr)
	}

	addressArrayType, err := abi.NewType("address[]", "", nil)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrCreateAddressArrayType, err)
	}

	initializationConfigData, err := abi.Arguments{{Type: addressArrayType}}.Pack(signerAddresses)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrEncodeSignerAddresses, err)
	}

	return s.createAccountOperation(
		s.operationExecutionAccount, s.ecdsaSignatureVerifyingAccountImplAddress, accountOwnerAddress, accountId,
		initializationConfigData,
	)
}

// RSAKey represents an RSA public key with exponent and modulus as hex strings.
//   - E: The public exponent as a hex string.
//   - N: The modulus as a hex string.
type RSAKey struct {
	E string
	N string
}

// abiRSAKey is an internal struct used for ABI encoding RSA keys as (bytes,bytes) tuples.
type abiRSAKey struct {
	E []byte
	N []byte
}

// PrepareDeployNewRSAAccountOperation prepares an operation to deploy a new RSA signature verifying account.
// The operation includes the account creation transaction with properly encoded RSA public keys.
//   - accountOwnerAddress: The address of the account owner as a hex string.
//   - allowedSigners: A slice of RSA public keys for allowed signers.
//   - accountId: A unique identifier for the account.
//
// Returns the operation and the predicted address where the account will be deployed.
func (s *Service) PrepareDeployNewRSAAccountOperation(
	accountOwnerAddress string, allowedSigners []RSAKey, accountId string,
) (*transactTypes.Operation, common.Address, error) {
	s.logger.Trace().
		Str("operationExecutionAccount", s.operationExecutionAccount.Hex()).
		Str("accountOwnerAddress", accountOwnerAddress).
		Str("allowedSigners", fmt.Sprintf("%v", allowedSigners)).
		Str("accountId", accountId).
		Msg("Preparing deploy new RSA account operation")

	abiKeys := make([]abiRSAKey, len(allowedSigners))
	for i, key := range allowedSigners {
		// Convert hex strings to bytes
		eBytes := common.FromHex(key.E)
		nBytes := common.FromHex(key.N)

		abiKeys[i] = abiRSAKey{
			E: eBytes,
			N: nBytes,
		}
	}

	rsaKeyType, err := abi.NewType(
		"tuple[]", "", []abi.ArgumentMarshaling{
			{Name: "e", Type: "bytes"},
			{Name: "n", Type: "bytes"},
		},
	)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrCreateRSAKeyArrayType, err)
	}

	initializationConfigData, err := abi.Arguments{{Type: rsaKeyType}}.Pack(abiKeys)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrEncodeRSAKeys, err)
	}

	return s.createAccountOperation(
		s.operationExecutionAccount, s.rsaSignatureVerifyingAccountImplAddress, accountOwnerAddress, accountId,
		initializationConfigData,
	)
}

// createAccountOperation is a helper function that creates a transact operation for account deployment.
// It encodes the createAccount function call with the provided parameters and predicts the account address.
//   - operationExecutionAccount: A previously deployed smart wallet address that will execute the operation to invoke the factory and deploy the account.
//   - implAddress: The implementation contract address for the account type.
//   - accountOwnerAddress: The address of the account owner as a hex string.
//   - accountId: A unique identifier for the account.
//   - initializationConfigData: ABI-encoded initialization data for the account.
//
// Returns the operation and the predicted account address.
func (s *Service) createAccountOperation(
	operationExecutionAccount common.Address, implAddress common.Address, accountOwnerAddress string, accountId string,
	initializationConfigData []byte,
) (*transactTypes.Operation, common.Address, error) {
	abiEncoder, err := accounts.AccountsMetaData.GetAbi()
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrGetAccountABI, err)
	}
	accountIdBytes32 := crypto.Keccak256Hash([]byte(accountId))

	// Predict the account address using the same logic as the contract
	salt := GetSalt(operationExecutionAccount, accountIdBytes32)
	predictedAddress := PredictAccountAddress(implAddress, salt, s.accountFactoryAddress)

	calldata, err := abiEncoder.Pack(
		"createAccount",
		implAddress,
		accountIdBytes32,
		s.keystoneForwarderAddress,
		common.HexToAddress(accountOwnerAddress),
		initializationConfigData,
	)

	if err != nil {
		return nil, common.Address{}, fmt.Errorf("%w: %w", ErrPackAccountData, err)
	}

	return &transactTypes.Operation{
		ID:      newSecureRandomID(),
		Account: operationExecutionAccount,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.accountFactoryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, predictedAddress, nil
}

// GetSalt generates a deterministic salt for account deployment.
// This matches the contract's GetSalt function: keccak256(abi.encodePacked(sender, uniqueAccountId))
func GetSalt(sender common.Address, uniqueAccountID [32]byte) [32]byte {
	return crypto.Keccak256Hash(append(sender.Bytes(), uniqueAccountID[:]...))
}

// PredictAccountAddress predicts the deterministic address where an account would be deployed.
// This replicates the behavior of Clones.predictDeterministicAddress from OpenZeppelin.
func PredictAccountAddress(implementation common.Address, salt [32]byte, factory common.Address) common.Address {
	// CREATE2 address calculation: keccak256(0xff ++ factory ++ salt ++ keccak256(bytecode))
	// For minimal proxy (EIP-1167), the bytecode is: 0x3d602d80600a3d3981f3363d3d373d3d3d363d73 + implementation + 0x5af43d82803e903d91602b57fd5bf3

	// EIP-1167 minimal proxy bytecode template
	initCode := make([]byte, 0, 55)
	initCode = append(
		initCode, []byte{
			0x3d, 0x60, 0x2d, 0x80, 0x60, 0x0a, 0x3d, 0x39, 0x81, 0xf3, 0x36, 0x3d, 0x3d, 0x37, 0x3d, 0x3d, 0x3d, 0x36,
			0x3d, 0x73,
		}...,
	)
	initCode = append(initCode, implementation.Bytes()...)
	initCode = append(
		initCode, []byte{0x5a, 0xf4, 0x3d, 0x82, 0x80, 0x3e, 0x90, 0x3d, 0x91, 0x60, 0x2b, 0x57, 0xfd, 0x5b, 0xf3}...,
	)

	bytecodeHash := crypto.Keccak256Hash(initCode)

	// CREATE2 address: keccak256(0xff ++ factory ++ salt ++ bytecodeHash)[12:]
	data := make([]byte, 0, 85)
	data = append(data, 0xff)
	data = append(data, factory.Bytes()...)
	data = append(data, salt[:]...)
	data = append(data, bytecodeHash.Bytes()...)

	addressHash := crypto.Keccak256(data)
	return common.BytesToAddress(addressHash[12:])
}

// newSecureRandomID generates a cryptographically secure random 256-bit integer for operation IDs.
func newSecureRandomID() *big.Int {
	n, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	return n
}
