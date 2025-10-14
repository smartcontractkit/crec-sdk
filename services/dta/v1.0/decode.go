package dta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

// ConcreteEvent represents any decoded concrete event payload.
type ConcreteEvent interface{}

// VerifiableEvent represents an event structure that encapsulates data about the event, its metadata, and associated blockchain transaction details.
type VerifiableEvent struct {
	CreatedAt   time.Time   `json:"createdAt"`
	Event       Event       `json:"event"`
	Metadata    Metadata    `json:"metadata"`
	Transaction Transaction `json:"transaction"`

	// ConcreteEvent holds the decoded concrete event based on the event name and using the fields in `VerifiableEvent.Metadata.WorkflowEvent.Attributes`
	ConcreteEvent ConcreteEvent `json:"-"`
}

type Event struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	RequestId string `json:"requestId"`
	TopicHash string `json:"topicHash"`
}
type Metadata struct {
	ChainId       string        `json:"chainId"`
	Network       string        `json:"network"`
	WorkflowEvent WorkflowEvent `json:"workflowEvent"`
}
type WorkflowEvent struct {
	Attributes      Attrs     `json:"attributes"`
	BusinessEventId string    `json:"business_event_id"`
	Component       string    `json:"component"`
	EventTimestamp  time.Time `json:"event_timestamp"`
	EventTypeLabel  string    `json:"event_type_label"`
	Failed          bool      `json:"failed"`
	FinalEvent      bool      `json:"final_event"`
	Id              string    `json:"id"`
	Participant     string    `json:"participant"`
	ParticipantRole string    `json:"participant_role"`
	ProcessLabels   []string  `json:"process_labels"`
	RawData         string    `json:"raw_data"`
	Title           string    `json:"title"`
}
type Transaction struct {
	Timestamp int    `json:"timestamp"`
	ChainId   string `json:"chainId"`
	Hash      string `json:"hash"`
}

type Attribute struct {
	Key        string `json:"key"`
	OnChain    bool   `json:"on_chain"`
	Value      string `json:"value"`
	Visibility string `json:"visibility"`
}

type Attrs map[string]Attribute

// Has checks if the specified key exists in the Attrs map. Returns true if the key is present; otherwise, false.
func (a Attrs) Has(key string) bool {
	_, ok := a[key]
	return ok
}

// Get retrieves the value and existence status of the specified key from the Attrs map. Returns the value and true if key exists, otherwise an empty string and false.
func (a Attrs) Get(key string) (string, bool) {
	v, ok := a[key]
	return v.Value, ok
}

// Require retrieves the value of the specified key from the Attrs map. Returns an error if the key is missing or its value is empty.
func (a Attrs) Require(key string) (string, error) {
	if v, ok := a.Get(key); ok && v != "" {
		return v, nil
	}
	return "", fmt.Errorf("missing required attribute %q", key)
}

// Default returns the value associated with the specified key if it exists and is non-empty; otherwise, it returns the provided default value.
func (a Attrs) Default(key, def string) string {
	if v, ok := a.Get(key); ok && v != "" {
		return v
	}
	return def
}

// UnmarshalJSON implements custom decoding to populate the concrete event
// from the attribute map. It determines the event name using the "event_type" attribute
// (or falls back to the outer Event.Name), then maps attributes into the struct fields.
func (v *VerifiableEvent) UnmarshalJSON(b []byte) error {
	// Use an alias to avoid infinite recursion
	type alias VerifiableEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return fmt.Errorf("failed to unmarshal VerifiableEvent envelope: %w", err)
	}

	// Copy envelope to receiver
	*v = VerifiableEvent(a)
	name := v.EventName()

	// Create the concrete event instance
	var concrete ConcreteEvent
	switch name {
	case EventDistributorRegistered:
		concrete = &DistributorRegistered{
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
		}
	case EventDistributorRequestCanceled:
		concrete = &DistributorRequestCanceled{
			FundTokenId:     common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestId:       common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
		}
	case EventDistributorRequestProcessed:
		shares, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["shares"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Metadata.WorkflowEvent.Attributes["shares"].Value)
		}
		status, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["status"].Value, 10, 8) // base 10, fit into uint8
		if err != nil {
			return fmt.Errorf("event %s unable to parse status: %s", name, v.Metadata.WorkflowEvent.Attributes["status"].Value)
		}
		concrete = &DistributorRequestProcessed{
			RequestId: common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Shares:    shares,
			Status:    uint8(status),
			Error:     []byte(v.Metadata.WorkflowEvent.Attributes["error"].Value),
		}
	case EventDistributorRequestProcessing:
		shares, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["shares"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Metadata.WorkflowEvent.Attributes["shares"].Value)
		}
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		concrete = &DistributorRequestProcessing{
			FundTokenId:     common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestId:       common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Shares:          shares,
			Amount:          amount,
		}
	case EventFundAdminRegistered:
		concrete = &FundAdminRegistered{
			FundAdminAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_admin_addr"].Value),
		}
	case EventFundTokenAllowlistUpdated:
		allowed, err := strconv.ParseBool(v.Metadata.WorkflowEvent.Attributes["allowed"].Value)
		if err != nil {
			return fmt.Errorf("event %s unable to parse allowed: %s", name, v.Metadata.WorkflowEvent.Attributes["allowed"].Value)
		}
		concrete = &FundTokenAllowlistUpdated{
			FundAdminAddr:   common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_admin_addr"].Value),
			FundTokenId:     common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			Allowed:         allowed,
		}
	case EventFundTokenRegistered:
		tokenChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["token_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse token_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["token_chain_selector"].Value)
		}
		concrete = &FundTokenRegistered{
			FundAdminAddr:      common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_admin_addr"].Value),
			FundTokenId:        common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			FundTokenAddr:      common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_token_addr"].Value),
			NavAddr:            common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["nav_addr"].Value),
			TokenChainSelector: tokenChainSelector,
		}
	case EventInitialized:
		version, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["version"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse version: %s", name, v.Metadata.WorkflowEvent.Attributes["version"].Value)
		}
		concrete = &Initialized{Version: version}
	case EventInvalidDTARequestSettlement:
		actualChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["actual_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse actual_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["actual_chain_selector"].Value)
		}
		concrete = &InvalidDTARequestSettlement{
			FundAdminAddr:            common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_admin_addr"].Value),
			FundTokenId:              common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			RequestId:                common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			ActualChainSelector:      actualChainSelector,
			ActualDTAAdminWalletAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["actual_dta_admin_wallet_addr"].Value),
		}
	case EventMessageFailed:
		concrete = &MessageFailed{
			MessageId: common.HexToHash(v.Metadata.WorkflowEvent.Attributes["message_id"].Value),
			Reason:    []byte(v.Metadata.WorkflowEvent.Attributes["reason"].Value),
		}
	case EventNativeFundsRecovered:
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		concrete = &NativeFundsRecovered{
			To:     common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["to"].Value),
			Amount: amount,
		}
	case EventOwnershipTransferred:
		concrete = &OwnershipTransferred{
			PreviousOwner: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["previous_owner"].Value),
			NewOwner:      common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["new_owner"].Value),
		}
	case EventRedemptionRequested:
		shares, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["shares"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Metadata.WorkflowEvent.Attributes["shares"].Value)
		}
		createdAt, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["created_at"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse created_at: %s", name, v.Metadata.WorkflowEvent.Attributes["created_at"].Value)
		}
		concrete = &RedemptionRequested{
			FundTokenId:     common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestId:       common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Shares:          shares,
			CreatedAt:       createdAt,
		}
	case EventSubscriptionRequested:
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		createdAt, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["created_at"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse created_at: %s", name, v.Metadata.WorkflowEvent.Attributes["created_at"].Value)
		}
		concrete = &SubscriptionRequested{
			FundTokenId:     common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestId:       common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Amount:          amount,
			CreatedAt:       createdAt,
		}
	case EventAnswerUpdated:
		current, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["current"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse current: %s", name, v.Metadata.WorkflowEvent.Attributes["current"].Value)
		}
		roundId, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["roundId"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse roundId: %s", name, v.Metadata.WorkflowEvent.Attributes["roundId"].Value)
		}
		updatedAt, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["updatedAt"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse updatedAt: %s", name, v.Metadata.WorkflowEvent.Attributes["updatedAt"].Value)
		}
		concrete = &AnswerUpdated{Current: current, RoundId: roundId, UpdatedAt: updatedAt}
	case EventCCIPMessageRecvFailed:
		concrete = &CCIPMessageRecvFailed{
			MessageId: common.HexToHash(v.Metadata.WorkflowEvent.Attributes["message_id"].Value),
			Reason:    []byte(v.Metadata.WorkflowEvent.Attributes["reason"].Value),
		}
	case EventDTAAdded:
		dtaChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value)
		}
		concrete = &DTAAdded{
			DtaAddr:          common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["dta_addr"].Value),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			FundTokenAddr:    common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_token_addr"].Value),
		}
	case EventDTARemoved:
		dtaChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value)
		}
		concrete = &DTARemoved{
			DtaAddr:          common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["dta_addr"].Value),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
		}
	case EventDTASettlementClosed:
		requestType, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["request_type"].Value, 10, 8)
		if err != nil {
			return fmt.Errorf("event %s unable to parse request_type: %s", name, v.Metadata.WorkflowEvent.Attributes["request_type"].Value)
		}
		dtaChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value)
		}
		success, err := strconv.ParseBool(v.Metadata.WorkflowEvent.Attributes["success"].Value)
		if err != nil {
			return fmt.Errorf("event %s unable to parse success: %s", name, v.Metadata.WorkflowEvent.Attributes["success"].Value)
		}
		concrete = &DTASettlementClosed{
			DistributorAddr:  common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestType:      uint8(requestType),
			FundTokenId:      common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DtaChainSelector: dtaChainSelector,
			DtaAddr:          common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["dta_addr"].Value),
			RequestId:        common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Success:          success,
			Err:              []byte(v.Metadata.WorkflowEvent.Attributes["err"].Value),
		}
	case EventDTASettlementOpened:
		requestType, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["request_type"].Value, 10, 8)
		if err != nil {
			return fmt.Errorf("event %s unable to parse request_type: %s", name, v.Metadata.WorkflowEvent.Attributes["request_type"].Value)
		}
		dtaChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value)
		}
		shares, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["shares"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Metadata.WorkflowEvent.Attributes["shares"].Value)
		}
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		currency, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["currency"].Value, 10, 8)
		if err != nil {
			return fmt.Errorf("event %s unable to parse currency: %s", name, v.Metadata.WorkflowEvent.Attributes["currency"].Value)
		}
		concrete = &DTASettlementOpened{
			DistributorAddr:       common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestType:           uint8(requestType),
			FundTokenId:           common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			FundAdminAddr:         common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["fund_admin_addr"].Value),
			DtaChainSelector:      dtaChainSelector,
			DtaAddr:               common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["dta_addr"].Value),
			RequestId:             common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			DistributorWalletAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_wallet_addr"].Value),
			Shares:                shares,
			Amount:                amount,
			Currency:              uint8(currency),
		}
	case EventEmptyRequestType:
		concrete = &EmptyRequestType{
			MessageId: common.HexToHash(v.Metadata.WorkflowEvent.Attributes["message_id"].Value),
			RequestId: common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
		}
	case EventInsufficientPaymentTokenBalance:
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		concrete = &InsufficientPaymentTokenBalance{
			FundTokenId:           common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr:       common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			DistributorWalletAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_wallet_addr"].Value),
			RequestId:             common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Amount:                amount,
		}
	case EventInvalidSubscriptionCrossChainPayment:
		ccipDestTokenAmountsLength, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["ccip_dest_token_amounts_length"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse ccip_dest_token_amounts_length: %s", name, v.Metadata.WorkflowEvent.Attributes["ccip_dest_token_amounts_length"].Value)
		}
		concrete = &InvalidSubscriptionCrossChainPayment{
			FundTokenId:                common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			RequestId:                  common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			PaymentTokenDestAddr:       common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["payment_token_dest_addr"].Value),
			CCIPDestTokenAmountsLength: ccipDestTokenAmountsLength,
			CCIPPaymentTokenAddr:       common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["ccip_payment_token_addr"].Value),
		}
	case EventSettlementFailed:
		shares, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["shares"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Metadata.WorkflowEvent.Attributes["shares"].Value)
		}
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		concrete = &SettlementFailed{
			FundTokenId:           common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr:       common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			PaymentTokenAddr:      common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["payment_token_addr"].Value),
			DistributorWalletAddr: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_wallet_addr"].Value),
			RequestId:             common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			Shares:                shares,
			Amount:                amount,
			ErrData:               []byte(v.Metadata.WorkflowEvent.Attributes["err_data"].Value),
		}
	case EventTokenWithdrawn:
		amount, ok := new(big.Int).SetString(v.Metadata.WorkflowEvent.Attributes["amount"].Value, 10)
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Metadata.WorkflowEvent.Attributes["amount"].Value)
		}
		concrete = &TokenWithdrawn{
			Token:     common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["token"].Value),
			Recipient: common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["recipient"].Value),
			Amount:    amount,
		}
	case EventUnauthorizedSenderDTA:
		reqType, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["req_type"].Value, 10, 8)
		if err != nil {
			return fmt.Errorf("event %s unable to parse req_type: %s", name, v.Metadata.WorkflowEvent.Attributes["req_type"].Value)
		}
		dtaChainSelector, err := strconv.ParseUint(v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value, 10, 64)
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Metadata.WorkflowEvent.Attributes["dta_chain_selector"].Value)
		}
		concrete = &UnauthorizedSenderDTA{
			DtaAddr:          common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["dta_addr"].Value),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Metadata.WorkflowEvent.Attributes["fund_token_id"].Value),
			DistributorAddr:  common.HexToAddress(v.Metadata.WorkflowEvent.Attributes["distributor_addr"].Value),
			RequestId:        common.HexToHash(v.Metadata.WorkflowEvent.Attributes["request_id"].Value),
			ReqType:          uint8(reqType),
		}
	default:
		return fmt.Errorf("unsupported event type: %s", v.Metadata.WorkflowEvent.Attributes["event_type"].Value)
	}

	v.ConcreteEvent = concrete
	return nil
}

// Decode parses a base64-encoded JSON string from the event and unmarshals it into a VerifiableEvent. It returns an error if decoding or unmarshalling fails.
func Decode(ctx context.Context, event apiClient.Event) (VerifiableEvent, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(event.VerifiableEvent)
	if err != nil {
		return VerifiableEvent{}, err
	}

	var verifiableEvent VerifiableEvent
	if err = json.Unmarshal(decodedBytes, &verifiableEvent); err != nil {
		return VerifiableEvent{}, err
	}

	return verifiableEvent, nil
}

// EventName determines and returns the event name from the workflow attributes or outer event name; defaults to EventUnknown if not resolvable.
func (v VerifiableEvent) EventName() EventName {
	var name EventName
	if attr, ok := v.Metadata.WorkflowEvent.Attributes["event_type"]; ok {
		if ev, ok := parseEvent(attr.Value); ok {
			name = ev
		}
	}
	if name == "" && v.Event.Name != "" {
		if ev, ok := parseEvent(v.Event.Name); ok {
			name = ev
		}
	}
	if name == "" {
		return EventUnknown
	}
	return name
}
