package dta

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// All events from DTAOpenMarketplaceU.abi.json

type EventName string

const (
	EventDistributorRegistered        EventName = "DistributorRegistered"
	EventDistributorRequestCanceled   EventName = "DistributorRequestCanceled"
	EventDistributorRequestProcessed  EventName = "DistributorRequestProcessed"
	EventDistributorRequestProcessing EventName = "DistributorRequestProcessing"
	EventFundAdminRegistered          EventName = "FundAdminRegistered"
	EventFundTokenAllowlistUpdated    EventName = "FundTokenAllowlistUpdated"
	EventFundTokenRegistered          EventName = "FundTokenRegistered"
	EventInitialized                  EventName = "Initialized"
	EventInvalidDTAWallet             EventName = "InvalidDTAWallet"
	EventMessageFailed                EventName = "MessageFailed"
	EventNativeFundsRecovered         EventName = "NativeFundsRecovered"
	EventOwnershipTransferred         EventName = "OwnershipTransferred"
	EventRedemptionRequested          EventName = "RedemptionRequested"
	EventSubscriptionRequested        EventName = "SubscriptionRequested"
)

var allEvents = map[string]EventName{
	string(EventDistributorRegistered):        EventDistributorRegistered,
	string(EventDistributorRequestCanceled):   EventDistributorRequestCanceled,
	string(EventDistributorRequestProcessed):  EventDistributorRequestProcessed,
	string(EventDistributorRequestProcessing): EventDistributorRequestProcessing,
	string(EventFundAdminRegistered):          EventFundAdminRegistered,
	string(EventFundTokenAllowlistUpdated):    EventFundTokenAllowlistUpdated,
	string(EventFundTokenRegistered):          EventFundTokenRegistered,
	string(EventInitialized):                  EventInitialized,
	string(EventInvalidDTAWallet):             EventInvalidDTAWallet,
	string(EventMessageFailed):                EventMessageFailed,
	string(EventNativeFundsRecovered):         EventNativeFundsRecovered,
	string(EventOwnershipTransferred):         EventOwnershipTransferred,
	string(EventRedemptionRequested):          EventRedemptionRequested,
	string(EventSubscriptionRequested):        EventSubscriptionRequested,
}

func parseEvent(value string) (EventName, bool) {
	ev, ok := allEvents[value]
	return ev, ok
}

func (ev EventName) String() string {
	return string(ev)
}

// DistributorRegistered (address distributorAddr)
type DistributorRegistered struct {
	DistributorAddr common.Address
}

// DistributorRequestCanceled (bytes32 fundTokenId, address distributorAddr, bytes32 requestId)
type DistributorRequestCanceled struct {
	FundTokenId     [32]byte
	DistributorAddr common.Address
	RequestId       [32]byte
}

// DistributorRequestProcessed (bytes32 requestId, uint256 shares, uint8 status, bytes error)
type DistributorRequestProcessed struct {
	RequestId [32]byte
	Shares    *big.Int
	Status    uint8
	Error     []byte
}

// DistributorRequestProcessing (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 shares, uint256 amount)
type DistributorRequestProcessing struct {
	FundTokenId     [32]byte
	DistributorAddr common.Address
	RequestId       [32]byte
	Shares          *big.Int
	Amount          *big.Int
}

// FundAdminRegistered (address fundAdminAddr)
type FundAdminRegistered struct {
	FundAdminAddr common.Address
}

// FundTokenAllowlistUpdated (address fundAdminAddr, bytes32 fundTokenId, address distributorAddr, bool allowed)
type FundTokenAllowlistUpdated struct {
	FundAdminAddr   common.Address
	FundTokenId     [32]byte
	DistributorAddr common.Address
	Allowed         bool
}

// FundTokenRegistered (address fundAdminAddr, bytes32 fundTokenId, address fundTokenAddr, address navAddr, uint64 tokenChainSelector)
type FundTokenRegistered struct {
	FundAdminAddr      common.Address
	FundTokenId        [32]byte
	FundTokenAddr      common.Address
	NavAddr            common.Address
	TokenChainSelector uint64
}

// Initialized (uint64 version)
type Initialized struct {
	Version uint64
}

// InvalidDTAWallet (address fundAdminAddr, bytes32 fundTokenId, bytes32 requestId, uint64 actualChainSelector, address actualDTAAdminWalletAddr)
type InvalidDTAWallet struct {
	FundAdminAddr            common.Address
	FundTokenId              [32]byte
	RequestId                [32]byte
	ActualChainSelector      uint64
	ActualDTAAdminWalletAddr common.Address
}

// MessageFailed (bytes32 messageId, bytes reason)
type MessageFailed struct {
	MessageId [32]byte
	Reason    []byte
}

// NativeFundsRecovered (address to, uint256 amount)
type NativeFundsRecovered struct {
	To     common.Address
	Amount *big.Int
}

// OwnershipTransferred (address previousOwner, address newOwner)
type OwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
}

// RedemptionRequested (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 shares, uint40 createdAt)
type RedemptionRequested struct {
	FundTokenId     [32]byte
	DistributorAddr common.Address
	RequestId       [32]byte
	Shares          *big.Int
	CreatedAt       uint64 // uint40 in Solidity -> uint64 in Go
}

// SubscriptionRequested (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 amount, uint40 createdAt)
type SubscriptionRequested struct {
	FundTokenId     [32]byte
	DistributorAddr common.Address
	RequestId       [32]byte
	Amount          *big.Int
	CreatedAt       uint64 // uint40 in Solidity -> uint64 in Go
}
