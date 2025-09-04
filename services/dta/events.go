package dta

import (
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

type EventName string

const (
	// DTAOpenMarketplaceU events
	EventDistributorRegistered        EventName = "DistributorRegistered"
	EventDistributorRequestCanceled   EventName = "DistributorRequestCanceled"
	EventDistributorRequestProcessed  EventName = "DistributorRequestProcessed"
	EventDistributorRequestProcessing EventName = "DistributorRequestProcessing"
	EventFundAdminRegistered          EventName = "FundAdminRegistered"
	EventFundTokenAllowlistUpdated    EventName = "FundTokenAllowlistUpdated"
	EventFundTokenRegistered          EventName = "FundTokenRegistered"
	EventSubscriptionRequested        EventName = "SubscriptionRequested"
	EventRedemptionRequested          EventName = "RedemptionRequested"
	// DTAOpenMarketplaceU (currently not used by DTAO)
	EventInitialized          EventName = "Initialized"
	EventInvalidDTAWallet     EventName = "InvalidDTAWallet"
	EventMessageFailed        EventName = "MessageFailed"
	EventNativeFundsRecovered EventName = "NativeFundsRecovered"
	EventOwnershipTransferred EventName = "OwnershipTransferred"

	// NAV events
	EventAnswerUpdated EventName = "AnswerUpdated"

	// DTAWalletU events (currently not used by DTAO)
	EventCCIPMessageRecvFailed           EventName = "CCIPMessageRecvFailed"
	EventDTAAdded                        EventName = "DTAAdded"
	EventDTARemoved                      EventName = "DTARemoved"
	EventDTASettlementClosed             EventName = "DTASettlementClosed"
	EventDTASettlementOpened             EventName = "DTASettlementOpened"
	EventEmptyRequestType                EventName = "EmptyRequestType"
	EventInsufficientPaymentTokenBalance EventName = "InsufficientPaymentTokenBalance"
	EventSettlementFailed                EventName = "SettlementFailed"
	EventTokenWithdrawn                  EventName = "TokenWithdrawn"
	EventUnauthorizedSenderDTA           EventName = "UnauthorizedSenderDTA"

	Unknown EventName = "Unknown"
)

var allEvents = map[string]EventName{
	string(EventDistributorRegistered):           EventDistributorRegistered,
	string(EventDistributorRequestCanceled):      EventDistributorRequestCanceled,
	string(EventDistributorRequestProcessed):     EventDistributorRequestProcessed,
	string(EventDistributorRequestProcessing):    EventDistributorRequestProcessing,
	string(EventFundAdminRegistered):             EventFundAdminRegistered,
	string(EventFundTokenAllowlistUpdated):       EventFundTokenAllowlistUpdated,
	string(EventFundTokenRegistered):             EventFundTokenRegistered,
	string(EventInitialized):                     EventInitialized,
	string(EventInvalidDTAWallet):                EventInvalidDTAWallet,
	string(EventMessageFailed):                   EventMessageFailed,
	string(EventNativeFundsRecovered):            EventNativeFundsRecovered,
	string(EventOwnershipTransferred):            EventOwnershipTransferred,
	string(EventRedemptionRequested):             EventRedemptionRequested,
	string(EventSubscriptionRequested):           EventSubscriptionRequested,
	string(EventAnswerUpdated):                   EventAnswerUpdated,
	string(EventCCIPMessageRecvFailed):           EventCCIPMessageRecvFailed,
	string(EventDTAAdded):                        EventDTAAdded,
	string(EventDTARemoved):                      EventDTARemoved,
	string(EventDTASettlementClosed):             EventDTASettlementClosed,
	string(EventDTASettlementOpened):             EventDTASettlementOpened,
	string(EventEmptyRequestType):                EventEmptyRequestType,
	string(EventInsufficientPaymentTokenBalance): EventInsufficientPaymentTokenBalance,
	string(EventSettlementFailed):                EventSettlementFailed,
	string(EventTokenWithdrawn):                  EventTokenWithdrawn,
	string(EventUnauthorizedSenderDTA):           EventUnauthorizedSenderDTA,
}

// eventTypeRegistry maps event names to their corresponding Go struct types for event data deserialization.
var eventTypeRegistry = map[EventName]reflect.Type{
	EventDistributorRegistered:           reflect.TypeOf(DistributorRegistered{}),
	EventDistributorRequestCanceled:      reflect.TypeOf(DistributorRequestCanceled{}),
	EventDistributorRequestProcessed:     reflect.TypeOf(DistributorRequestProcessed{}),
	EventDistributorRequestProcessing:    reflect.TypeOf(DistributorRequestProcessing{}),
	EventFundAdminRegistered:             reflect.TypeOf(FundAdminRegistered{}),
	EventFundTokenAllowlistUpdated:       reflect.TypeOf(FundTokenAllowlistUpdated{}),
	EventFundTokenRegistered:             reflect.TypeOf(FundTokenRegistered{}),
	EventInitialized:                     reflect.TypeOf(Initialized{}),
	EventInvalidDTAWallet:                reflect.TypeOf(InvalidDTAWallet{}),
	EventMessageFailed:                   reflect.TypeOf(MessageFailed{}),
	EventNativeFundsRecovered:            reflect.TypeOf(NativeFundsRecovered{}),
	EventOwnershipTransferred:            reflect.TypeOf(OwnershipTransferred{}),
	EventRedemptionRequested:             reflect.TypeOf(RedemptionRequested{}),
	EventSubscriptionRequested:           reflect.TypeOf(SubscriptionRequested{}),
	EventAnswerUpdated:                   reflect.TypeOf(AnswerUpdated{}),
	EventCCIPMessageRecvFailed:           reflect.TypeOf(CCIPMessageRecvFailed{}),
	EventDTAAdded:                        reflect.TypeOf(DTAAdded{}),
	EventDTARemoved:                      reflect.TypeOf(DTARemoved{}),
	EventDTASettlementClosed:             reflect.TypeOf(DTASettlementClosed{}),
	EventDTASettlementOpened:             reflect.TypeOf(DTASettlementOpened{}),
	EventEmptyRequestType:                reflect.TypeOf(EmptyRequestType{}),
	EventInsufficientPaymentTokenBalance: reflect.TypeOf(InsufficientPaymentTokenBalance{}),
	EventSettlementFailed:                reflect.TypeOf(SettlementFailed{}),
	EventTokenWithdrawn:                  reflect.TypeOf(TokenWithdrawn{}),
	EventUnauthorizedSenderDTA:           reflect.TypeOf(UnauthorizedSenderDTA{}),
}

func parseEvent(value string) (EventName, bool) {
	ev, ok := allEvents[value]
	return ev, ok
}

func (ev EventName) String() string {
	return string(ev)
}

// TODO: currently the below structs are strictly based on the chain events but any data we add in during the cvn flow we should also add here with a note about its origin

// -----------------------------
// DTAOpenMarketplaceU events as structs
// -----------------------------

// DistributorRegistered (address distributorAddr)
type DistributorRegistered struct {
	DistributorAddr common.Address `json:"distributor_addr"`
}

//TODO: for requests of any status call function getDistributorRequest(bytes32 requestId) external view returns (DistributorRequest memory)
//  struct DistributorRequest {
//    uint256 shares;
//    uint256 amount;
//    bytes32 fundTokenId;
//    address fundAdminAddr;
//    address distributorAddr;
//    uint40 createdAt;
//    /// @notice Distributor Request Type is either Subscription or Redemption
//    DistributorRequestType requestType;
//    RequestStatus status;
//  }

// DistributorRequestCanceled (bytes32 fundTokenId, address distributorAddr, bytes32 requestId)
type DistributorRequestCanceled struct {
	RequestId       [32]byte       `json:"request_id"`
	FundTokenId     [32]byte       `json:"fund_token_id"`
	DistributorAddr common.Address `json:"distributor_addr"`
	//DTAAddr       string `json:"dta_addr"`
	//FundAdminAddr string `json:"fund_admin_addr"`
	//Amount        string `json:"amount"`
	//Shares        string `json:"shares"`
	//RequestType   uint8  `json:"request_type"`
	//Status        uint8  `json:"status"`
	//CreatedAt     string `json:"created_at"`
}

// DistributorRequestProcessed (bytes32 requestId, uint256 shares, uint8 status, bytes error)
type DistributorRequestProcessed struct {
	RequestId [32]byte `json:"request_id"`
	Shares    *big.Int `json:"shares"`
	Status    uint8    `json:"status"`
	Error     []byte   `json:"error"`
	//DTAAddr         string `json:"dta_addr"`
	//FundAdminAddr   string `json:"fund_admin_addr"`
	//FundTokenID     string `json:"fund_token_id"`
	//DistributorAddr string `json:"distributor_addr"`
	//Amount          string `json:"amount"`
	//RequestType     uint8  `json:"request_type"`
	//CreatedAt       string `json:"created_at"`
}

// DistributorRequestProcessing (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 shares, uint256 amount)
type DistributorRequestProcessing struct {
	RequestId       [32]byte       `json:"request_id"`
	FundTokenId     [32]byte       `json:"fund_token_id"`
	DistributorAddr common.Address `json:"distributor_addr"`
	Shares          *big.Int       `json:"shares"`
	Amount          *big.Int       `json:"amount"`
	//DTAAddr         string         `json:"dta_addr"`
	//FundAdminAddr   string         `json:"fund_admin_addr"`
	//RequestType     uint8          `json:"request_type"`
	//Status          uint8          `json:"status"`
	//CreatedAt       string         `json:"created_at"`
}

// FundAdminRegistered (address fundAdminAddr)
type FundAdminRegistered struct {
	FundAdminAddr common.Address `json:"fund_admin_addr"`
}

// FundTokenAllowlistUpdated (address fundAdminAddr, bytes32 fundTokenId, address distributorAddr, bool allowed)
type FundTokenAllowlistUpdated struct {
	FundAdminAddr   common.Address `json:"fund_admin_addr"`
	FundTokenId     [32]byte       `json:"fund_token_id"`
	DistributorAddr common.Address `json:"distributor_addr"`
	Allowed         bool           `json:"allowed"`
}

// FundTokenRegistered (address fundAdminAddr, bytes32 fundTokenId, address fundTokenAddr, address navAddr, uint64 tokenChainSelector)
type FundTokenRegistered struct {
	FundAdminAddr      common.Address `json:"fund_admin_addr"`
	FundTokenId        [32]byte       `json:"fund_token_id"`
	FundTokenAddr      common.Address `json:"fund_token_addr"`
	NavAddr            common.Address `json:"nav_addr"`
	TokenChainSelector uint64         `json:"token_chain_selector"`
	//TODO: these would come from an extra read to function getFundToken(address fundAdminAddr,bytes32 fundTokenId)  returns (bool enabled, FundTokenData memory);
	//DTAWalletAddr                 common.Address `json:"dta_wallet_addr"`
	//NavFeedDecimals               uint8          `json:"nav_feed_decimals"`
	//NavTTL                        uint32         `json:"nav_ttl"`
	//TimezoneOffsetSecs            int64          `json:"timezone_offset_secs"`
	//PurchaseTokenDecimals         uint8          `json:"purchase_token_decimals"`
	//PurchaseTokenRoundingDecimals uint8          `json:"purchase_token_rounding_decimals"`
	//FundTokenDecimals             uint8          `json:"fund_token_decimals"`
	//FundRoundingDecimals          uint8          `json:"fund_rounding_decimals"`
	//RequestsPerDay                uint8          `json:"requests_per_day"`
	//PaymentTokenSourceAddr        string         `json:"payment_token_source_addr"`
	//PaymentSourceChainSelector    string         `json:"payment_source_chain_selector"`
	//PaymentTokenDestAddr          string         `json:"payment_token_dest_addr"`
	//PaymentDestChainSelector      string         `json:"payment_dest_chain_selector"`
	//PaymentOffChainCurrency       uint64         `json:"payment_off_chain_currency"`
	//Enabled                       bool           `json:"enabled"`
}

// Initialized (uint64 version)
type Initialized struct {
	Version uint64 `json:"version"`
}

// InvalidDTAWallet (address fundAdminAddr, bytes32 fundTokenId, bytes32 requestId, uint64 actualChainSelector, address actualDTAAdminWalletAddr)
type InvalidDTAWallet struct {
	FundAdminAddr            common.Address `json:"fund_admin_addr"`
	FundTokenId              [32]byte       `json:"fund_token_id"`
	RequestId                [32]byte       `json:"request_id"`
	ActualChainSelector      uint64         `json:"actual_chain_selector"`
	ActualDTAAdminWalletAddr common.Address `json:"actual_dta_admin_wallet_addr"`
}

// MessageFailed (bytes32 messageId, bytes reason)
type MessageFailed struct {
	MessageId [32]byte `json:"message_id"`
	Reason    []byte   `json:"reason"`
}

// NativeFundsRecovered (address to, uint256 amount)
type NativeFundsRecovered struct {
	To     common.Address `json:"to"`
	Amount *big.Int       `json:"amount"`
}

// OwnershipTransferred (address previousOwner, address newOwner)
type OwnershipTransferred struct {
	PreviousOwner common.Address `json:"previous_owner"`
	NewOwner      common.Address `json:"new_owner"`
}

// RedemptionRequested (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 shares, uint40 createdAt)
type RedemptionRequested struct {
	FundTokenId     [32]byte       `json:"fund_token_id"`
	DistributorAddr common.Address `json:"distributor_addr"`
	RequestId       [32]byte       `json:"request_id"`
	Shares          *big.Int       `json:"shares"`
	CreatedAt       uint64         `json:"created_at"` // uint40 in Solidity -> uint64 in Go
}

// SubscriptionRequested (bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint256 amount, uint40 createdAt)
type SubscriptionRequested struct {
	FundTokenId     [32]byte       `json:"fund_token_id"`
	DistributorAddr common.Address `json:"distributor_addr"`
	RequestId       [32]byte       `json:"request_id"`
	Amount          *big.Int       `json:"amount"`
	CreatedAt       uint64         `json:"created_at"` // uint40 in Solidity -> uint64 in Go
}

// -----------------------------
// NAV events as structs
// -----------------------------

// AnswerUpdated (int256 indexed current, uint256 indexed roundId, uint256 updatedAt) note: from IDecimalAggregator.sol not DTA
type AnswerUpdated struct {
	Current   *big.Int `json:"current"`
	RoundId   *big.Int `json:"round_id"`
	UpdatedAt *big.Int `json:"updated_at"`
}

// -----------------------------
// DTAWalletU events as structs
// -----------------------------

// CCIPMessageRecvFailed (bytes32 messageId, bytes reason)
type CCIPMessageRecvFailed struct {
	MessageId [32]byte `json:"message_id"`
	Reason    []byte   `json:"reason"`
}

// DTAAdded (address dtaAddr, uint64 dtaChainSelector, bytes32 fundTokenId, address fundTokenAddr)
type DTAAdded struct {
	DtaAddr          common.Address `json:"dta_addr"`
	DtaChainSelector uint64         `json:"dta_chain_selector"`
	FundTokenId      [32]byte       `json:"fund_token_id"`
	FundTokenAddr    common.Address `json:"fund_token_addr"`
}

// DTARemoved (address dtaAddr, uint64 dtaChainSelector, bytes32 fundTokenId)
type DTARemoved struct {
	DtaAddr          common.Address `json:"dta_addr"`
	DtaChainSelector uint64         `json:"dta_chain_selector"`
	FundTokenId      [32]byte       `json:"fund_token_id"`
}

// DTASettlementClosed (address distributorAddr, uint8 requestType, bytes32 fundTokenId, uint64 dtaChainSelector, address dtaAddr, bytes32 requestId, bool success, bytes err)
type DTASettlementClosed struct {
	DistributorAddr  common.Address `json:"distributor_addr"`
	RequestType      uint8          `json:"request_type"`
	FundTokenId      [32]byte       `json:"fund_token_id"`
	DtaChainSelector uint64         `json:"dta_chain_selector"`
	DtaAddr          common.Address `json:"dta_addr"`
	RequestId        [32]byte       `json:"request_id"`
	Success          bool           `json:"success"`
	Err              []byte         `json:"err"`
}

// DTASettlementOpened (address distributorAddr, uint8 requestType, bytes32 fundTokenId, address fundAdminAddr, uint64 dtaChainSelector, address dtaAddr, bytes32 requestId, address distributorWalletAddr, uint256 shares, uint256 amount, uint8 currency)
type DTASettlementOpened struct {
	DistributorAddr       common.Address `json:"distributor_addr"`
	RequestType           uint8          `json:"request_type"`
	FundTokenId           [32]byte       `json:"fund_token_id"`
	FundAdminAddr         common.Address `json:"fund_admin_addr"`
	DtaChainSelector      uint64         `json:"dta_chain_selector"`
	DtaAddr               common.Address `json:"dta_addr"`
	RequestId             [32]byte       `json:"request_id"`
	DistributorWalletAddr common.Address `json:"distributor_wallet_addr"`
	Shares                *big.Int       `json:"shares"`
	Amount                *big.Int       `json:"amount"`
	Currency              uint8          `json:"currency"`
}

// EmptyRequestType (bytes32 messageId, bytes32 requestId)
type EmptyRequestType struct {
	MessageId [32]byte `json:"message_id"`
	RequestId [32]byte `json:"request_id"`
}

// InsufficientPaymentTokenBalance (bytes32 fundTokenId, address distributorAddr, address distributorWalletAddr, bytes32 requestId, uint256 amount)
type InsufficientPaymentTokenBalance struct {
	FundTokenId           [32]byte       `json:"fund_token_id"`
	DistributorAddr       common.Address `json:"distributor_addr"`
	DistributorWalletAddr common.Address `json:"distributor_wallet_addr"`
	RequestId             [32]byte       `json:"request_id"`
	Amount                *big.Int       `json:"amount"`
}

// SettlementFailed (bytes32 fundTokenId, address distributorAddr, address paymentTokenAddr, address distributorWalletAddr, bytes32 requestId, uint256 shares, uint256 amount, bytes errData)
type SettlementFailed struct {
	FundTokenId           [32]byte       `json:"fund_token_id"`
	DistributorAddr       common.Address `json:"distributor_addr"`
	PaymentTokenAddr      common.Address `json:"payment_token_addr"`
	DistributorWalletAddr common.Address `json:"distributor_wallet_addr"`
	RequestId             [32]byte       `json:"request_id"`
	Shares                *big.Int       `json:"shares"`
	Amount                *big.Int       `json:"amount"`
	ErrData               []byte         `json:"err_data"`
}

// TokenWithdrawn (address token, address recipient, uint256 amount)
type TokenWithdrawn struct {
	Token     common.Address `json:"token"`
	Recipient common.Address `json:"recipient"`
	Amount    *big.Int       `json:"amount"`
}

// UnauthorizedSenderDTA (address dtaAddr, uint64 dtaChainSelector, bytes32 fundTokenId, address distributorAddr, bytes32 requestId, uint8 reqType)
type UnauthorizedSenderDTA struct {
	DtaAddr          common.Address `json:"dta_addr"`
	DtaChainSelector uint64         `json:"dta_chain_selector"`
	FundTokenId      [32]byte       `json:"fund_token_id"`
	DistributorAddr  common.Address `json:"distributor_addr"`
	RequestId        [32]byte       `json:"request_id"`
	ReqType          uint8          `json:"req_type"`
}
