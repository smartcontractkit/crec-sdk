package dta

import (
	"encoding/base64"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/services/dta/gen/dtaopenmarketplace"
	"github.com/smartcontractkit/crec-api-go/services/dta/gen/dtawallet"

	"github.com/smartcontractkit/crec-sdk/interfaces/erc20"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// ServiceOptions defines the options for creating a new CREC DTA service.
//   - Logger: Optional logger instance.
//   - DTAOpenMarketplaceAddress: A string representing the address of the DTA OpenMarketplace contract.
//   - DTAWalletAddress: A string representing the address of the DTA Wallet contract.
//   - AccountAddress: A string representing the address of the account performing the DTA operations.
type ServiceOptions struct {
	Logger                    *zerolog.Logger
	DTAOpenMarketplaceAddress string
	DTAWalletAddress          string
	AccountAddress            string
}

type Service struct {
	logger                    *zerolog.Logger
	dtaOpenMarketplaceAddress common.Address
	dtaWalletAddress          common.Address
	accountAddress            common.Address
}

// NewService creates a new CREC DTA service with the provided options.
// Returns a pointer to the Service and an error if any issues occur during initialization.
//   - opts: Options for configuring the CREC DTA service, see ServiceOptions for details.
func NewService(opts *ServiceOptions) (*Service, error) {
	var logger *zerolog.Logger
	if opts.Logger != nil {
		logger = opts.Logger
	} else {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	logger.Info().Msg("Creating CREC DTA service")

	return &Service{
		logger:                    logger,
		dtaOpenMarketplaceAddress: common.HexToAddress(opts.DTAOpenMarketplaceAddress),
		dtaWalletAddress:          common.HexToAddress(opts.DTAWalletAddress),
		accountAddress:            common.HexToAddress(opts.AccountAddress),
	}, nil
}

// Operation preparation methods

// PrepareRequestSubscriptionOperation prepares a DTA request subscription operation.
// It constructs the necessary transaction to request a subscription for a fund token.
//   - fundAdminAddr: The address of the fund admin.
//   - fundTokenId: The ID of the fund token.
//   - amount: The amount to subscribe.
func (s *Service) PrepareRequestSubscriptionOperation(
	fundAdminAddr common.Address,
	fundTokenId [32]byte,
	amount *big.Int,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("requestSubscription", fundAdminAddr, fundTokenId, amount)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for requestSubscription")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRequestRedemptionOperation prepares a DTA request redemption operation.
// It constructs the necessary transaction to request a redemption for fund shares.
//   - fundAdminAddr: The address of the fund admin.
//   - fundTokenId: The ID of the fund token.
//   - shares: The number of shares to redeem.
func (s *Service) PrepareRequestRedemptionOperation(
	fundAdminAddr common.Address,
	fundTokenId [32]byte,
	shares *big.Int,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("requestRedemption", fundAdminAddr, fundTokenId, shares)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for requestRedemption")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRequestSubscriptionWithTokenApprovalOperation prepares a DTA request subscription operation with token approval.
// It constructs both the token approval transaction and the subscription request transaction.
//   - fundAdminAddr: The address of the fund admin.
//   - fundTokenId: The ID of the fund token.
//   - amount: The amount to subscribe.
//   - paymentTokenAddress: The address of the payment token to approve.
func (s *Service) PrepareRequestSubscriptionWithTokenApprovalOperation(
	fundAdminAddr common.Address,
	fundTokenId [32]byte,
	amount *big.Int,
	paymentTokenAddress common.Address,
) (*transactTypes.Operation, error) {
	// Prepare the token approval transaction
	approveTransaction, err := s.prepareTokenApproveTransaction(&paymentTokenAddress, amount)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to prepare token approve transaction")
		return nil, err
	}

	// Prepare the subscription request transaction
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("requestSubscription", fundAdminAddr, fundTokenId, amount)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for requestSubscription")
		return nil, err
	}

	subscriptionTransaction := &transactTypes.Transaction{
		To:    s.dtaOpenMarketplaceAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			*approveTransaction,
			*subscriptionTransaction,
		},
	}, nil
}

// PrepareProcessDistributorRequestOperation prepares a DTA process distributor request operation.
// It constructs the necessary transaction to process a distributor request.
//   - requestId: The ID of the request to process.
func (s *Service) PrepareProcessDistributorRequestOperation(requestId [32]byte) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("processDistributorRequest", requestId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for processDistributorRequest")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareCancelDistributorRequestOperation prepares a DTA cancel distributor request operation.
// It constructs the necessary transaction to cancel a distributor request.
//   - requestId: The ID of the request to cancel.
func (s *Service) PrepareCancelDistributorRequestOperation(requestId [32]byte) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("cancelDistributorRequest", requestId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for cancelDistributorRequest")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRegisterDistributorOperation prepares a DTA register distributor operation.
// It constructs the necessary transaction to register a new distributor.
//   - distributorAddr: The address of the distributor to register.
//   - distributorWalletAddr: The wallet address of the distributor.
func (s *Service) PrepareRegisterDistributorOperation(
	distributorAddr common.Address,
	distributorWalletAddr common.Address,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerDistributor", distributorAddr, distributorWalletAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerDistributor")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareUpdateDistributorOperation prepares a DTA update distributor operation.
// It constructs the necessary transaction to update a distributor.
//   - distributorWalletAddr: The updated wallet address of the distributor.
func (s *Service) PrepareUpdateDistributorOperation(
	distributorWalletAddr common.Address,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("updateDistributor", distributorWalletAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for updateDistributor")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRegisterFundAdminOperation prepares a DTA register fund admin operation.
// It constructs the necessary transaction to register a new fund admin.
//   - fundAdminAddr: The address of the fund admin to register.
func (s *Service) PrepareRegisterFundAdminOperation(fundAdminAddr common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("registerFundAdmin", fundAdminAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerFundAdmin")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// FundTokenData represents the data structure for fund token registration.
// This struct mirrors the IFundTokenRegistry.FundTokenData from the contract.
type FundTokenData struct {
	FundTokenAddr                 common.Address
	NavFeedDecimals               uint8
	PurchaseTokenRoundingDecimals uint8
	PurchaseTokenDecimals         uint8
	FundRoundingDecimals          uint8
	FundTokenDecimals             uint8
	RequestsPerDay                uint8
	NavAddr                       common.Address
	TokenChainSelector            uint64
	DtaWalletAddr                 common.Address
	TimezoneOffsetSecs            *big.Int // int24 in contract
	NavTTL                        *big.Int // uint24 in contract
	PaymentInfo                   DTAPaymentInfo
}

// DTAPaymentInfo represents the payment information for DTA operations.
type DTAPaymentInfo struct {
	OffChainPaymentCurrency uint8 // Currency enum
	PaymentTokenSourceAddr  common.Address
	PaymentTokenDestAddr    common.Address
}

// PrepareRegisterFundTokenOperation prepares a DTA register fund token operation.
// It constructs the necessary transaction to register a new fund token with its metadata.
//   - fundTokenId: The ID of the fund token to register.
//   - tokenData: The fund token data containing all metadata.
func (s *Service) PrepareRegisterFundTokenOperation(
	fundTokenId [32]byte,
	tokenData FundTokenData,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	// Convert our struct to the contract's expected tuple format
	contractTokenData := struct {
		FundTokenAddr                 common.Address
		NavFeedDecimals               uint8
		PurchaseTokenRoundingDecimals uint8
		PurchaseTokenDecimals         uint8
		FundRoundingDecimals          uint8
		FundTokenDecimals             uint8
		RequestsPerDay                uint8
		NavAddr                       common.Address
		TokenChainSelector            uint64
		DtaWalletAddr                 common.Address
		TimezoneOffsetSecs            *big.Int
		NavTTL                        *big.Int
		PaymentInfo                   struct {
			OffChainPaymentCurrency uint8
			PaymentTokenSourceAddr  common.Address
			PaymentTokenDestAddr    common.Address
		}
	}{
		FundTokenAddr:                 tokenData.FundTokenAddr,
		NavFeedDecimals:               tokenData.NavFeedDecimals,
		PurchaseTokenRoundingDecimals: tokenData.PurchaseTokenRoundingDecimals,
		PurchaseTokenDecimals:         tokenData.PurchaseTokenDecimals,
		FundRoundingDecimals:          tokenData.FundRoundingDecimals,
		FundTokenDecimals:             tokenData.FundTokenDecimals,
		RequestsPerDay:                tokenData.RequestsPerDay,
		NavAddr:                       tokenData.NavAddr,
		TokenChainSelector:            tokenData.TokenChainSelector,
		DtaWalletAddr:                 tokenData.DtaWalletAddr,
		TimezoneOffsetSecs:            tokenData.TimezoneOffsetSecs,
		NavTTL:                        tokenData.NavTTL,
		PaymentInfo: struct {
			OffChainPaymentCurrency uint8
			PaymentTokenSourceAddr  common.Address
			PaymentTokenDestAddr    common.Address
		}{
			OffChainPaymentCurrency: tokenData.PaymentInfo.OffChainPaymentCurrency,
			PaymentTokenSourceAddr:  tokenData.PaymentInfo.PaymentTokenSourceAddr,
			PaymentTokenDestAddr:    tokenData.PaymentInfo.PaymentTokenDestAddr,
		},
	}

	calldata, err := abiEncoder.Pack("registerFundToken", fundTokenId, contractTokenData)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for registerFundToken")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareAllowDistributorForTokenOperation prepares a DTA allow distributor for token operation.
// It constructs the necessary transaction to allow a distributor for a specific token.
//   - fundTokenId: The ID of the fund token.
//   - distributorAddr: The address of the distributor to allow.
func (s *Service) PrepareAllowDistributorForTokenOperation(
	fundTokenId [32]byte,
	distributorAddr common.Address,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("allowDistributorForToken", fundTokenId, distributorAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for allowDistributorForToken")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareDisallowDistributorForTokenOperation prepares a DTA disallow distributor for token operation.
// It constructs the necessary transaction to disallow a distributor for a specific token.
//   - fundTokenId: The ID of the fund token.
//   - distributorAddr: The address of the distributor to disallow.
func (s *Service) PrepareDisallowDistributorForTokenOperation(
	fundTokenId [32]byte,
	distributorAddr common.Address,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("disallowDistributorForToken", fundTokenId, distributorAddr)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for disallowDistributorForToken")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareEnableFundTokenOperation prepares a DTA enable fund token operation.
// It constructs the necessary transaction to enable a fund token.
//   - fundTokenId: The ID of the fund token to enable.
func (s *Service) PrepareEnableFundTokenOperation(fundTokenId [32]byte) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("enableFundToken", fundTokenId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for enableFundToken")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareDisableFundTokenOperation prepares a DTA disable fund token operation.
// It constructs the necessary transaction to disable a fund token.
//   - fundTokenId: The ID of the fund token to disable.
func (s *Service) PrepareDisableFundTokenOperation(fundTokenId [32]byte) (*transactTypes.Operation, error) {
	abiEncoder, err := dtaopenmarketplace.DtaopenmarketplaceMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA OpenMarketplace ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("disableFundToken", fundTokenId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for disableFundToken")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaOpenMarketplaceAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// toJson decodes an encoded VerifiableEvent from a CREC event into a JSON byte slice.
func (s *Service) toJson(event *apiClient.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

// prepareTokenApproveTransaction prepares a token approval transaction.
// It constructs the necessary transaction to approve tokens for the DTA OpenMarketplace contract.
//   - tokenAddress: The address of the token to approve.
//   - tokenAmount: The amount of tokens to approve.
func (s *Service) prepareTokenApproveTransaction(
	tokenAddress *common.Address, tokenAmount *big.Int,
) (*transactTypes.Transaction, error) {
	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get ERC20 ABI")
		return nil, err
	}

	calldata, err := erc20Abi.Pack("approve", s.dtaOpenMarketplaceAddress, tokenAmount)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to pack calldata for token approve")
		return nil, err
	}

	return &transactTypes.Transaction{
		To:    *tokenAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}, nil
}

// DTAWallet Operations - These target the DTAWallet contract

// TokenBurnType represents the burn type for DTA operations
type TokenBurnType uint8

const (
	TokenBurnTypeNone TokenBurnType = iota
	TokenBurnTypeBurn
	TokenBurnTypeTransfer
)

// PrepareAllowDTAOperation prepares a DTA wallet allow DTA operation.
// It constructs the necessary transaction to allow a DTA address to interact with a fund token.
//   - dtaAddr: The DTA contract address to allow.
//   - dtaChainSelector: The chain selector for the DTA contract.
//   - fundTokenId: The ID of the fund token.
//   - fundTokenAddr: The address of the fund token contract.
//   - burnType: The type of token burn/transfer behavior.
func (s *Service) PrepareAllowDTAOperation(
	dtaAddr common.Address,
	dtaChainSelector uint64,
	fundTokenId [32]byte,
	fundTokenAddr common.Address,
	burnType TokenBurnType,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("allowDTA", dtaAddr, dtaChainSelector, fundTokenId, fundTokenAddr, uint8(burnType))
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for allowDTA")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareDisallowDTAOperation prepares a DTA wallet disallow DTA operation.
// It constructs the necessary transaction to disallow a DTA address from interacting with a fund token.
//   - dtaAddr: The DTA contract address to disallow.
//   - dtaChainSelector: The chain selector for the DTA contract.
//   - fundTokenId: The ID of the fund token.
func (s *Service) PrepareDisallowDTAOperation(
	dtaAddr common.Address,
	dtaChainSelector uint64,
	fundTokenId [32]byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("disallowDTA", dtaAddr, dtaChainSelector, fundTokenId)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for disallowDTA")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareWithdrawTokensOperation prepares a DTA wallet withdraw tokens operation.
// It constructs the necessary transaction to withdraw tokens from the DTA wallet.
//   - token: The address of the token to withdraw.
//   - recipient: The address to receive the withdrawn tokens.
//   - amount: The amount of tokens to withdraw.
func (s *Service) PrepareWithdrawTokensOperation(
	token common.Address,
	recipient common.Address,
	amount *big.Int,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("withdrawTokens", token, recipient, amount)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for withdrawTokens")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareTransferWalletOwnershipOperation prepares a DTA wallet transfer ownership operation.
// It constructs the necessary transaction to transfer ownership of the DTA wallet.
//   - newOwner: The address of the new owner.
func (s *Service) PrepareTransferWalletOwnershipOperation(newOwner common.Address) (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
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
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareRenounceWalletOwnershipOperation prepares a DTA wallet renounce ownership operation.
// It constructs the necessary transaction to renounce ownership of the DTA wallet.
func (s *Service) PrepareRenounceWalletOwnershipOperation() (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
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
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}

// PrepareCompleteRequestProcessingOperation prepares a DTA wallet complete request processing operation.
// It constructs the necessary transaction to complete the processing of a request in the DTA wallet.
//   - requestId: The ID of the request to complete.
//   - success: Whether the request processing was successful.
//   - errorData: Error data if the request failed (can be empty for successful requests).
func (s *Service) PrepareCompleteRequestProcessingOperation(
	requestId [32]byte,
	success bool,
	errorData []byte,
) (*transactTypes.Operation, error) {
	abiEncoder, err := dtawallet.DtawalletMetaData.GetAbi()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get DTA Wallet ABI")
		return nil, err
	}

	calldata, err := abiEncoder.Pack("completeRequestProcessing", requestId, success, errorData)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to pack calldata for completeRequestProcessing")
		return nil, err
	}

	return &transactTypes.Operation{
		ID:      big.NewInt(time.Now().Unix()),
		Account: s.accountAddress,
		Transactions: []transactTypes.Transaction{
			{
				To:    s.dtaWalletAddress,
				Value: big.NewInt(0),
				Data:  calldata,
			},
		},
	}, nil
}
