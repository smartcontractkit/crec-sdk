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

type ServiceOptions struct {
	Logger                                    *zerolog.Logger
	KeystoneForwarderAddress                  string
	AccountFactoryAddress                     string
	ECDSASignatureVerifyingAccountImplAddress string
	RSASignatureVerifyingAccountImplAddress   string
}

type Service struct {
	logger                                    *zerolog.Logger
	keystoneForwarderAddress                  common.Address
	accountFactoryAddress                     common.Address
	ecdsaSignatureVerifyingAccountImplAddress common.Address
	rsaSignatureVerifyingAccountImplAddress   common.Address
}

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

func (s *Service) PrepareDeployNewECDSAAccountOperation(accountOwnerAddress string, allowedSigners []string, accountId string) (*transactTypes.Operation, error) {
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

	abiEncoder, err := accounts.AccountsMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get account factory ABI: %w", err)
	}

	accountIdBytes32 := crypto.Keccak256Hash([]byte(accountId))

	calldata, err := abiEncoder.Pack("createAccount",
		s.ecdsaSignatureVerifyingAccountImplAddress,
		accountIdBytes32,
		s.keystoneForwarderAddress,
		accountOwnerAddress,
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

func (s *Service) PrepareDeployNewRSAAccountOperation(accountOwnerAddress string, allowedSigners []string, accountId string) (*transactTypes.Operation, error) {

}

func newSecureRandomID() *big.Int {
	n, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	return n
}
