package wallets

import (
	"crypto/rand"
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

const (
	ServiceName = "wallets"
)

// ServiceOptions defines the options for creating a new CREC Accounts service.
//   - Logger: Optional logger instance.
//   - OperationExecutionWallet: A previously deployed smart wallet address that will execute operations to invoke the factory and deploy wallets.
//   - KeystoneForwarderAddress: A string representing the address of the Keystone forwarder contract.
//   - AccountFactoryAddress: A string representing the address of the wallet factory contract.
//   - ECDSASignatureVerifyingWalletImplAddress: A string representing the address of the ECDSA signature verifying wallet implementation contract.
//   - RSASignatureVerifyingWalletImplAddress: A string representing the address of the RSA signature verifying wallet implementation contract.
type ServiceOptions struct {
	Logger                                   *zerolog.Logger
	OperationExecutionWallet                 string
	KeystoneForwarderAddress                 string
	AccountFactoryAddress                    string
	ECDSASignatureVerifyingWalletImplAddress string
	RSASignatureVerifyingWalletImplAddress   string
}

// Service provides operations for managing CREC wallet creation and deployment.
type Service struct {
	logger                                   *zerolog.Logger
	operationExecutionWallet                 common.Address
	keystoneForwarderAddress                 common.Address
	walletFactoryAddress                     common.Address
	ecdsaSignatureVerifyingWalletImplAddress common.Address
	rsaSignatureVerifyingWalletImplAddress   common.Address
}

// NewService creates a new CREC Wallets service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC Wallets service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, fmt.Errorf("ServiceOptions is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC Wallets service")

	return &Service{
		logger:                                   logger,
		operationExecutionWallet:                 common.HexToAddress(opts.OperationExecutionWallet),
		keystoneForwarderAddress:                 common.HexToAddress(opts.KeystoneForwarderAddress),
		walletFactoryAddress:                     common.HexToAddress(opts.AccountFactoryAddress),
		ecdsaSignatureVerifyingWalletImplAddress: common.HexToAddress(opts.ECDSASignatureVerifyingWalletImplAddress),
		rsaSignatureVerifyingWalletImplAddress:   common.HexToAddress(opts.RSASignatureVerifyingWalletImplAddress),
	}, nil
}

// PrepareDeployNewECDSAWalletOperation prepares an operation to deploy a new ECDSA signature verifying wallet.
// The operation includes the wallet creation transaction with properly encoded allowed signers.
//   - walletOwnerAddress: The address of the wallet owner as a hex string.
//   - allowedSigners: A slice of allowed signer addresses as hex strings.
//   - walletId: A unique identifier for the wallet.
//
// Returns the operation and the predicted address where the wallet will be deployed.
func (s *Service) PrepareDeployNewECDSAWalletOperation(
	walletOwnerAddress string, allowedSigners []string, walletId string,
) (*transactTypes.Operation, common.Address, error) {
	s.logger.Trace().
		Str("operationExecutionWallet", s.operationExecutionWallet.Hex()).
		Str("walletOwnerAddress", walletOwnerAddress).
		Str("allowedSigners", fmt.Sprintf("%v", allowedSigners)).
		Str("walletId", walletId).
		Msg("Preparing deploy new ECDSA wallet operation")

	signerAddresses := make([]common.Address, len(allowedSigners))
	for i, signerStr := range allowedSigners {
		signerAddresses[i] = common.HexToAddress(signerStr)
	}

	addressArrayType, err := abi.NewType("address[]", "", nil)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to create address array ABI type: %w", err)
	}

	initializationConfigData, err := abi.Arguments{{Type: addressArrayType}}.Pack(signerAddresses)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to encode allowed signers: %w", err)
	}

	return s.createWalletOperation(
		s.operationExecutionWallet, s.ecdsaSignatureVerifyingWalletImplAddress, walletOwnerAddress, walletId,
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

// PrepareDeployNewRSAWalletOperation prepares an operation to deploy a new RSA signature verifying wallet.
// The operation includes the wallet creation transaction with properly encoded RSA public keys.
//   - walletOwnerAddress: The address of the wallet owner as a hex string.
//   - allowedSigners: A slice of RSA public keys for allowed signers.
//   - walletId: A unique identifier for the wallet.
//
// Returns the operation and the predicted address where the wallet will be deployed.
func (s *Service) PrepareDeployNewRSAWalletOperation(
	walletOwnerAddress string, allowedSigners []RSAKey, walletId string,
) (*transactTypes.Operation, common.Address, error) {
	s.logger.Trace().
		Str("operationExecutionWallet", s.operationExecutionWallet.Hex()).
		Str("walletOwnerAddress", walletOwnerAddress).
		Str("allowedSigners", fmt.Sprintf("%v", allowedSigners)).
		Str("walletId", walletId).
		Msg("Preparing deploy new RSA wallet operation")

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
		return nil, common.Address{}, fmt.Errorf("failed to create RSAKey array ABI type: %w", err)
	}

	initializationConfigData, err := abi.Arguments{{Type: rsaKeyType}}.Pack(abiKeys)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to encode RSA keys: %w", err)
	}

	return s.createWalletOperation(
		s.operationExecutionWallet, s.rsaSignatureVerifyingWalletImplAddress, walletOwnerAddress, walletId,
		initializationConfigData,
	)
}

// createWalletOperation is a helper function that creates a transact operation for wallet deployment.
// It encodes the createAccount function call with the provided parameters and predicts the wallet address.
//   - operationExecutionWallet: A previously deployed smart wallet address that will execute the operation to invoke the factory and deploy the wallet.
//   - implAddress: The implementation contract address for the wallet type.
//   - walletOwnerAddress: The address of the wallet owner as a hex string.
//   - walletId: A unique identifier for the wallet.
//   - initializationConfigData: ABI-encoded initialization data for the wallet.
//
// Returns the operation and the predicted wallet address.
func (s *Service) createWalletOperation(
	operationExecutionWallet common.Address, implAddress common.Address, walletOwnerAddress string, walletId string,
	initializationConfigData []byte,
) (*transactTypes.Operation, common.Address, error) {
	abiEncoder, err := accounts.AccountsMetaData.GetAbi()
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to get wallet factory ABI: %w", err)

	}
	walletIdBytes32 := crypto.Keccak256Hash([]byte(walletId))

	// Predict the wallet address using the same logic as the contract
	salt := GetSalt(operationExecutionWallet, walletIdBytes32)
	predictedAddress := PredictWalletAddress(implAddress, salt, s.walletFactoryAddress)

	calldata, err := abiEncoder.Pack(
		// TODO: Update this to the correct function selector when renaming to createWallet
		"createAccount",
		implAddress,
		walletIdBytes32,
		s.keystoneForwarderAddress,
		common.HexToAddress(walletOwnerAddress),
		initializationConfigData,
	)

	if err != nil {
		// TODO: Update this to the correct function selector when renaming to createWallet
		return nil, common.Address{}, fmt.Errorf("failed to pack createAccount calldata: %w", err)
	}

	return &transactTypes.Operation{
		ID:      newSecureRandomID(),
		Account: operationExecutionWallet,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.walletFactoryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, predictedAddress, nil
}

// GetSalt generates a deterministic salt for wallet deployment.
// This matches the contract's GetSalt function: keccak256(abi.encodePacked(sender, uniqueAccountId))
func GetSalt(sender common.Address, uniqueAccountID [32]byte) [32]byte {
	return crypto.Keccak256Hash(append(sender.Bytes(), uniqueAccountID[:]...))
}

// PredictWalletAddress predicts the deterministic address where an wallet would be deployed.
// This replicates the behavior of Clones.predictDeterministicAddress from OpenZeppelin.
func PredictWalletAddress(implementation common.Address, salt [32]byte, factory common.Address) common.Address {
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
