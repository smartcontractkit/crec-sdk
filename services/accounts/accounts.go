package accounts

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-api-go/services/accounts/gen/accounts"

	transactTypes "github.com/smartcontractkit/cvn-sdk/transact/types"
)

const (
	ServiceName = "accounts"
)

// ServiceOptions defines the options for creating a new CVN Accounts service.
//   - Logger: Optional logger instance.
//   - KeystoneForwarderAddress: A string representing the address of the Keystone forwarder contract.
//   - AccountFactoryAddress: A string representing the address of the account factory contract.
//   - ECDSASignatureVerifyingAccountImplAddress: A string representing the address of the ECDSA signature verifying account implementation contract.
//   - RSASignatureVerifyingAccountImplAddress: A string representing the address of the RSA signature verifying account implementation contract.
type ServiceOptions struct {
	Logger                                    *zerolog.Logger
	KeystoneForwarderAddress                  string
	AccountFactoryAddress                     string
	ECDSASignatureVerifyingAccountImplAddress string
	RSASignatureVerifyingAccountImplAddress   string
}

// Service provides operations for managing CVN account creation and deployment.
type Service struct {
	logger                                    *zerolog.Logger
	keystoneForwarderAddress                  common.Address
	accountFactoryAddress                     common.Address
	ecdsaSignatureVerifyingAccountImplAddress common.Address
	rsaSignatureVerifyingAccountImplAddress   common.Address
}

// NewService creates a new CVN Accounts service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CVN Accounts service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	if opts == nil {
		return nil, fmt.Errorf("ServiceOptions is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CVN Accounts service")

	return &Service{
		logger:                   logger,
		keystoneForwarderAddress: common.HexToAddress(opts.KeystoneForwarderAddress),
		accountFactoryAddress:    common.HexToAddress(opts.AccountFactoryAddress),
		ecdsaSignatureVerifyingAccountImplAddress: common.HexToAddress(opts.ECDSASignatureVerifyingAccountImplAddress),
		rsaSignatureVerifyingAccountImplAddress:   common.HexToAddress(opts.RSASignatureVerifyingAccountImplAddress),
	}, nil
}

// PrepareDeployNewECDSAAccountOperation prepares an operation to deploy a new ECDSA signature verifying account.
// The operation includes the account creation transaction with properly encoded allowed signers.
//   - accountOwnerAddress: The address of the account owner as a hex string.
//   - allowedSigners: A slice of allowed signer addresses as hex strings.
//   - accountId: A unique identifier for the account.
func (s *Service) PrepareDeployNewECDSAAccountOperation(
	accountOwnerAddress string, allowedSigners []string, accountId string,
) (*transactTypes.Operation, error) {
	s.logger.Trace().
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
		return nil, fmt.Errorf("failed to create address array ABI type: %w", err)
	}

	initializationConfigData, err := abi.Arguments{{Type: addressArrayType}}.Pack(signerAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to encode allowed signers: %w", err)
	}

	return s.createAccountOperation(
		s.ecdsaSignatureVerifyingAccountImplAddress, accountOwnerAddress, accountId, initializationConfigData,
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
func (s *Service) PrepareDeployNewRSAAccountOperation(
	accountOwnerAddress string, allowedSigners []RSAKey, accountId string,
) (*transactTypes.Operation, error) {
	s.logger.Trace().
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
		return nil, fmt.Errorf("failed to create RSAKey array ABI type: %w", err)
	}

	initializationConfigData, err := abi.Arguments{{Type: rsaKeyType}}.Pack(abiKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to encode RSA keys: %w", err)
	}

	return s.createAccountOperation(
		s.rsaSignatureVerifyingAccountImplAddress, accountOwnerAddress, accountId, initializationConfigData,
	)
}

// createAccountOperation is a helper function that creates a transact operation for account deployment.
// It encodes the createAccount function call with the provided parameters.
//   - implAddress: The implementation contract address for the account type.
//   - accountOwnerAddress: The address of the account owner as a hex string.
//   - accountId: A unique identifier for the account.
//   - initializationConfigData: ABI-encoded initialization data for the account.
func (s *Service) createAccountOperation(
	implAddress common.Address, accountOwnerAddress string, accountId string, initializationConfigData []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := accounts.AccountsMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get account factory ABI: %w", err)
	}

	accountIdBytes32 := crypto.Keccak256Hash([]byte(accountId))

	calldata, err := abiEncoder.Pack(
		"createAccount",
		implAddress,
		accountIdBytes32,
		s.keystoneForwarderAddress,
		common.HexToAddress(accountOwnerAddress),
		initializationConfigData,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to pack createAccount calldata: %w", err)
	}

	return &transactTypes.Operation{
		ID:      newSecureRandomID(),
		Account: s.accountFactoryAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.accountFactoryAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// newSecureRandomID generates a cryptographically secure random 256-bit integer for operation IDs.
func newSecureRandomID() *big.Int {
	n, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	return n
}
