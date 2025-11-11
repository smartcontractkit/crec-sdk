package dta

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ConcreteEvent represents any decoded concrete event payload.
type ConcreteEvent interface{}

// VerifiableEvent represents an event structure that encapsulates data about the event, its metadata, and associated blockchain transaction details.
type VerifiableEvent struct {
	CreatedAt   time.Time         `json:"created_at"`
	Event       Event             `json:"event"`
	Metadata    Metadata          `json:"metadata"`
	Parameters  map[string]string `json:"parameters"`
	Transaction Transaction       `json:"transaction"`

	// ConcreteEvent holds the decoded concrete event based on the event name and using the fields in `VerifiableEvent.Metadata.WorkflowEvent.Attributes`
	ConcreteEvent ConcreteEvent `json:"-"`
}

type Event struct {
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Service     string            `json:"service"`
	LogIndex    int               `json:"log_index"`
	Parameters  map[string]string `json:"parameters"`
	TopicHash   string            `json:"topic_hash"`
	BlockNumber int               `json:"block_number"`
}
type Metadata struct {
	ChainId       string        `json:"chainId"`
	Network       string        `json:"network"`
	WorkflowEvent WorkflowEvent `json:"workflowEvent"`
}
type WorkflowEvent struct {
	Component      string   `json:"component"`
	Attributes     Attrs    `json:"attributes"`
	ProcessLabels  []string `json:"process_labels"`
	EventTypeLabel string   `json:"event_type_label"`
}
type Transaction struct {
	Hash        string `json:"hash"`
	ChainId     string `json:"chain_id"`
	Timestamp   int    `json:"timestamp"`
	BlockNumber int    `json:"block_number"`
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
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
		}
	case EventDistributorRequestCanceled:
		concrete = &DistributorRequestCanceled{
			FundTokenId:     common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestId:       common.HexToHash(v.Event.Parameters["request_id"]),
		}
	case EventDistributorRequestProcessed:
		shares, ok := parseScientificNotationToBigInt(v.Event.Parameters["shares"])
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Event.Parameters["shares"])
		}
		status, err := parseScientificNotationToUint8(v.Event.Parameters["status"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse status: %s", name, v.Event.Parameters["status"])
		}
		concrete = &DistributorRequestProcessed{
			RequestId: common.HexToHash(v.Event.Parameters["request_id"]),
			Shares:    shares,
			Status:    status,
			Error:     []byte(v.Event.Parameters["error"]),
		}
	case EventDistributorRequestProcessing:
		shares, ok := parseScientificNotationToBigInt(v.Event.Parameters["shares"])
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Event.Parameters["shares"])
		}
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		concrete = &DistributorRequestProcessing{
			FundTokenId:     common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestId:       common.HexToHash(v.Event.Parameters["request_id"]),
			Shares:          shares,
			Amount:          amount,
		}
	case EventFundAdminRegistered:
		concrete = &FundAdminRegistered{
			FundAdminAddr: common.HexToAddress(v.Event.Parameters["fund_admin_addr"]),
		}
	case EventFundTokenAllowlistUpdated:
		allowed, err := strconv.ParseBool(v.Event.Parameters["allowed"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse allowed: %s", name, v.Event.Parameters["allowed"])
		}
		concrete = &FundTokenAllowlistUpdated{
			FundAdminAddr:   common.HexToAddress(v.Event.Parameters["fund_admin_addr"]),
			FundTokenId:     common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			Allowed:         allowed,
		}
	case EventFundTokenRegistered:
		tokenChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["token_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse token_chain_selector: %s", name, v.Event.Parameters["token_chain_selector"])
		}
		concrete = &FundTokenRegistered{
			FundAdminAddr:      common.HexToAddress(v.Event.Parameters["fund_admin_addr"]),
			FundTokenId:        common.HexToHash(v.Event.Parameters["fund_token_id"]),
			FundTokenAddr:      common.HexToAddress(v.Event.Parameters["fund_token_addr"]),
			NavAddr:            common.HexToAddress(v.Event.Parameters["nav_addr"]),
			TokenChainSelector: tokenChainSelector,
		}
	case EventInitialized:
		version, err := parseScientificNotationToUint64(v.Event.Parameters["version"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse version: %s", name, v.Event.Parameters["version"])
		}
		concrete = &Initialized{Version: version}
	case EventInvalidDTARequestSettlement:
		actualChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["actual_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse actual_chain_selector: %s", name, v.Event.Parameters["actual_chain_selector"])
		}
		concrete = &InvalidDTARequestSettlement{
			FundAdminAddr:            common.HexToAddress(v.Event.Parameters["fund_admin_addr"]),
			FundTokenId:              common.HexToHash(v.Event.Parameters["fund_token_id"]),
			RequestId:                common.HexToHash(v.Event.Parameters["request_id"]),
			ActualChainSelector:      actualChainSelector,
			ActualDTAAdminWalletAddr: common.HexToAddress(v.Event.Parameters["actual_dta_admin_wallet_addr"]),
		}
	case EventMessageFailed:
		concrete = &MessageFailed{
			MessageId: common.HexToHash(v.Event.Parameters["message_id"]),
			Reason:    []byte(v.Event.Parameters["reason"]),
		}
	case EventNativeFundsRecovered:
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		concrete = &NativeFundsRecovered{
			To:     common.HexToAddress(v.Event.Parameters["to"]),
			Amount: amount,
		}
	case EventOwnershipTransferred:
		concrete = &OwnershipTransferred{
			PreviousOwner: common.HexToAddress(v.Event.Parameters["previous_owner"]),
			NewOwner:      common.HexToAddress(v.Event.Parameters["new_owner"]),
		}
	case EventRedemptionRequested:
		shares, ok := parseScientificNotationToBigInt(v.Event.Parameters["shares"])
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Event.Parameters["shares"])
		}
		createdAt, err := parseScientificNotationToUint64(v.Event.Parameters["created_at"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse created_at: %s", name, v.Event.Parameters["created_at"])
		}
		concrete = &RedemptionRequested{
			FundTokenId:     common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestId:       common.HexToHash(v.Event.Parameters["request_id"]),
			Shares:          shares,
			CreatedAt:       createdAt,
		}
	case EventSubscriptionRequested:
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		createdAt, err := parseScientificNotationToUint64(v.Event.Parameters["created_at"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse created_at: %s", name, v.Event.Parameters["created_at"])
		}
		concrete = &SubscriptionRequested{
			FundTokenId:     common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr: common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestId:       common.HexToHash(v.Event.Parameters["request_id"]),
			Amount:          amount,
			CreatedAt:       createdAt,
		}
	case EventAnswerUpdated:
		current, ok := parseScientificNotationToBigInt(v.Event.Parameters["current"])
		if !ok {
			return fmt.Errorf("event %s unable to parse current: %s", name, v.Event.Parameters["current"])
		}
		roundId, ok := parseScientificNotationToBigInt(v.Event.Parameters["roundId"])
		if !ok {
			return fmt.Errorf("event %s unable to parse roundId: %s", name, v.Event.Parameters["roundId"])
		}
		updatedAt, ok := parseScientificNotationToBigInt(v.Event.Parameters["updatedAt"])
		if !ok {
			return fmt.Errorf("event %s unable to parse updatedAt: %s", name, v.Event.Parameters["updatedAt"])
		}
		concrete = &AnswerUpdated{Current: current, RoundId: roundId, UpdatedAt: updatedAt}
	case EventCCIPMessageRecvFailed:
		concrete = &CCIPMessageRecvFailed{
			MessageId: common.HexToHash(v.Event.Parameters["message_id"]),
			Reason:    []byte(v.Event.Parameters["reason"]),
		}
	case EventDTAAdded:
		dtaChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["dta_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Event.Parameters["dta_chain_selector"])
		}
		concrete = &DTAAdded{
			DtaAddr:          common.HexToAddress(v.Event.Parameters["dta_addr"]),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Event.Parameters["fund_token_id"]),
			FundTokenAddr:    common.HexToAddress(v.Event.Parameters["fund_token_addr"]),
		}
	case EventDTARemoved:
		dtaChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["dta_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Event.Parameters["dta_chain_selector"])
		}
		concrete = &DTARemoved{
			DtaAddr:          common.HexToAddress(v.Event.Parameters["dta_addr"]),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Event.Parameters["fund_token_id"]),
		}
	case EventDTASettlementClosed:
		requestType, err := parseScientificNotationToUint8(v.Event.Parameters["request_type"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse request_type: %s", name, v.Event.Parameters["request_type"])
		}
		dtaChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["dta_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Event.Parameters["dta_chain_selector"])
		}
		success, err := strconv.ParseBool(v.Event.Parameters["success"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse success: %s", name, v.Event.Parameters["success"])
		}
		concrete = &DTASettlementClosed{
			DistributorAddr:  common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestType:      uint8(requestType),
			FundTokenId:      common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DtaChainSelector: dtaChainSelector,
			DtaAddr:          common.HexToAddress(v.Event.Parameters["dta_addr"]),
			RequestId:        common.HexToHash(v.Event.Parameters["request_id"]),
			Success:          success,
			Err:              []byte(v.Event.Parameters["err"]),
		}
	case EventDTASettlementOpened:
		requestType, err := parseScientificNotationToUint8(v.Event.Parameters["request_type"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse request_type: %s", name, v.Event.Parameters["request_type"])
		}
		dtaChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["dta_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Event.Parameters["dta_chain_selector"])
		}
		shares, ok := parseScientificNotationToBigInt(v.Event.Parameters["shares"])
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Event.Parameters["shares"])
		}
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		currency, err := parseScientificNotationToUint8(v.Event.Parameters["currency"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse currency: %s", name, v.Event.Parameters["currency"])
		}
		concrete = &DTASettlementOpened{
			DistributorAddr:       common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestType:           requestType,
			FundTokenId:           common.HexToHash(v.Event.Parameters["fund_token_id"]),
			FundAdminAddr:         common.HexToAddress(v.Event.Parameters["fund_admin_addr"]),
			DtaChainSelector:      dtaChainSelector,
			DtaAddr:               common.HexToAddress(v.Event.Parameters["dta_addr"]),
			RequestId:             common.HexToHash(v.Event.Parameters["request_id"]),
			DistributorWalletAddr: common.HexToAddress(v.Event.Parameters["distributor_wallet_addr"]),
			Shares:                shares,
			Amount:                amount,
			Currency:              currency,
		}
	case EventEmptyRequestType:
		concrete = &EmptyRequestType{
			MessageId: common.HexToHash(v.Event.Parameters["message_id"]),
			RequestId: common.HexToHash(v.Event.Parameters["request_id"]),
		}
	case EventInsufficientPaymentTokenBalance:
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		concrete = &InsufficientPaymentTokenBalance{
			FundTokenId:           common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr:       common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			DistributorWalletAddr: common.HexToAddress(v.Event.Parameters["distributor_wallet_addr"]),
			RequestId:             common.HexToHash(v.Event.Parameters["request_id"]),
			Amount:                amount,
		}
	case EventInvalidSubscriptionCrossChainPayment:
		ccipDestTokenAmountsLength, ok := parseScientificNotationToBigInt(v.Event.Parameters["ccip_dest_token_amounts_length"])
		if !ok {
			return fmt.Errorf("event %s unable to parse ccip_dest_token_amounts_length: %s", name, v.Event.Parameters["ccip_dest_token_amounts_length"])
		}
		concrete = &InvalidSubscriptionCrossChainPayment{
			FundTokenId:                common.HexToHash(v.Event.Parameters["fund_token_id"]),
			RequestId:                  common.HexToHash(v.Event.Parameters["request_id"]),
			PaymentTokenDestAddr:       common.HexToAddress(v.Event.Parameters["payment_token_dest_addr"]),
			CCIPDestTokenAmountsLength: ccipDestTokenAmountsLength,
			CCIPPaymentTokenAddr:       common.HexToAddress(v.Event.Parameters["ccip_payment_token_addr"]),
		}
	case EventSettlementFailed:
		shares, ok := parseScientificNotationToBigInt(v.Event.Parameters["shares"])
		if !ok {
			return fmt.Errorf("event %s unable to parse shares: %s", name, v.Event.Parameters["shares"])
		}
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		concrete = &SettlementFailed{
			FundTokenId:           common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr:       common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			PaymentTokenAddr:      common.HexToAddress(v.Event.Parameters["payment_token_addr"]),
			DistributorWalletAddr: common.HexToAddress(v.Event.Parameters["distributor_wallet_addr"]),
			RequestId:             common.HexToHash(v.Event.Parameters["request_id"]),
			Shares:                shares,
			Amount:                amount,
			ErrData:               []byte(v.Event.Parameters["err_data"]),
		}
	case EventTokenWithdrawn:
		amount, ok := parseScientificNotationToBigInt(v.Event.Parameters["amount"])
		if !ok {
			return fmt.Errorf("event %s unable to parse amount: %s", name, v.Event.Parameters["amount"])
		}
		concrete = &TokenWithdrawn{
			Token:     common.HexToAddress(v.Event.Parameters["token"]),
			Recipient: common.HexToAddress(v.Event.Parameters["recipient"]),
			Amount:    amount,
		}
	case EventUnauthorizedSenderDTA:
		reqType, err := parseScientificNotationToUint8(v.Event.Parameters["req_type"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse req_type: %s", name, v.Event.Parameters["req_type"])
		}
		dtaChainSelector, err := parseScientificNotationToUint64(v.Event.Parameters["dta_chain_selector"])
		if err != nil {
			return fmt.Errorf("event %s unable to parse dta_chain_selector: %s", name, v.Event.Parameters["dta_chain_selector"])
		}
		concrete = &UnauthorizedSenderDTA{
			DtaAddr:          common.HexToAddress(v.Event.Parameters["dta_addr"]),
			DtaChainSelector: dtaChainSelector,
			FundTokenId:      common.HexToHash(v.Event.Parameters["fund_token_id"]),
			DistributorAddr:  common.HexToAddress(v.Event.Parameters["distributor_addr"]),
			RequestId:        common.HexToHash(v.Event.Parameters["request_id"]),
			ReqType:          reqType,
		}
	default:
		return fmt.Errorf("unsupported event type: %s", name)
	}

	v.ConcreteEvent = concrete
	return nil
}

// Decode parses a base64-encoded JSON string and unmarshals it into a VerifiableEvent. It returns an error if decoding or unmarshalling fails.
func Decode(ctx context.Context, verifiableEventString string) (VerifiableEvent, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(verifiableEventString)
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

// parseScientificNotationToBigInt converts scientific notation strings to big.Int
// Handles formats like "1.2e+21", "1e18", etc. that big.Int.SetString cannot parse directly
// Also handles decimal numbers like "600000000000000000000.000000"
func parseScientificNotationToBigInt(value string) (*big.Int, bool) {
	// First try direct parsing in case it's already a regular integer
	if result, ok := new(big.Int).SetString(value, 10); ok {
		return result, true
	}

	// Handle scientific notation
	lowerValue := strings.ToLower(value)
	if strings.Contains(lowerValue, "e") {
		// Split on 'e' to get mantissa and exponent
		parts := strings.Split(lowerValue, "e")
		if len(parts) != 2 {
			return nil, false
		}

		mantissaStr := parts[0]
		exponentStr := parts[1]

		// Remove optional '+' from exponent
		exponentStr = strings.TrimPrefix(exponentStr, "+")

		// Parse exponent as integer
		exponent, err := strconv.Atoi(exponentStr)
		if err != nil {
			return nil, false
		}

		// Handle negative exponents (fractional results truncated to integer)
		if exponent < 0 {
			// For negative exponents, we need to check if the result would be < 1
			// If so, truncate to 0 (integer part)
			mantissaFloat, err := strconv.ParseFloat(mantissaStr, 64)
			if err != nil {
				return nil, false
			}

			// Calculate the actual value to see if it's < 1
			actualValue := mantissaFloat * pow10(exponent)
			if actualValue < 1.0 {
				return big.NewInt(0), true
			}

			// If >= 1, we need to handle it properly
			// Convert to string without scientific notation and truncate decimal part
			decimalStr := fmt.Sprintf("%.0f", actualValue)
			if result, ok := new(big.Int).SetString(decimalStr, 10); ok {
				return result, true
			}
			return nil, false
		}

		// For positive exponents, handle manually to avoid precision loss
		var mantissaBig *big.Int

		// Check if mantissa has decimal point
		if strings.Contains(mantissaStr, ".") {
			// Split mantissa into integer and fractional parts
			decimalParts := strings.Split(mantissaStr, ".")
			if len(decimalParts) != 2 {
				return nil, false
			}

			integerPart := decimalParts[0]
			fractionalPart := decimalParts[1]

			// Combine integer and fractional parts
			combinedStr := integerPart + fractionalPart

			// Parse as big integer
			var ok bool
			mantissaBig, ok = new(big.Int).SetString(combinedStr, 10)
			if !ok {
				return nil, false
			}

			// Adjust exponent to account for the fractional digits
			exponent -= len(fractionalPart)
		} else {
			// No decimal point, parse directly
			var ok bool
			mantissaBig, ok = new(big.Int).SetString(mantissaStr, 10)
			if !ok {
				return nil, false
			}
		}

		// Multiply by 10^exponent
		if exponent > 0 {
			// Multiply by 10^exponent
			multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
			result := new(big.Int).Mul(mantissaBig, multiplier)
			return result, true
		} else if exponent < 0 {
			// Divide by 10^(-exponent) and truncate to integer
			divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exponent)), nil)
			result := new(big.Int).Div(mantissaBig, divisor)
			return result, true
		} else {
			// exponent == 0
			return mantissaBig, true
		}
	}

	// Handle decimal numbers without scientific notation (like "600000000000000000000.000000")
	if strings.Contains(value, ".") {
		// Split into integer and fractional parts
		decimalParts := strings.Split(value, ".")
		if len(decimalParts) != 2 {
			return nil, false
		}

		integerPart := decimalParts[0]
		fractionalPart := decimalParts[1]

		// Check if fractional part contains only zeros
		allZeros := true
		for _, digit := range fractionalPart {
			if digit != '0' {
				allZeros = false
				break
			}
		}

		// If fractional part is not all zeros, we cannot safely convert to big.Int
		if !allZeros {
			return nil, false
		}

		// Parse the integer part directly
		result, ok := new(big.Int).SetString(integerPart, 10)
		if !ok {
			return nil, false
		}

		// For integer conversion, we can safely truncate the all-zero decimal part
		return result, true
	}

	return nil, false
}

// parseScientificNotationToUint64 converts scientific notation strings to uint64
// Handles formats like "1.2e+21", "1e18", etc. that strconv.ParseUint cannot parse directly
func parseScientificNotationToUint64(value string) (uint64, error) {
	// First try direct parsing in case it's already a regular integer
	if result, err := strconv.ParseUint(value, 10, 64); err == nil {
		return result, nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, ok := parseScientificNotationToBigInt(value)
	if !ok {
		return 0, fmt.Errorf("unable to parse scientific notation: %s", value)
	}

	// Check if the result fits in uint64
	if !bigIntResult.IsUint64() {
		return 0, fmt.Errorf("value too large for uint64: %s", value)
	}

	return bigIntResult.Uint64(), nil
}

// parseScientificNotationToUint8 converts scientific notation strings to uint8
// Handles formats like "1e2", "2.5e+1", etc. that strconv.ParseUint cannot parse directly
func parseScientificNotationToUint8(value string) (uint8, error) {
	// First try direct parsing in case it's already a regular integer
	if result, err := strconv.ParseUint(value, 10, 8); err == nil {
		return uint8(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, ok := parseScientificNotationToBigInt(value)
	if !ok {
		return 0, fmt.Errorf("unable to parse scientific notation: %s", value)
	}

	// Check if the result fits in uint8 (0-255)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(255)) > 0 {
		return 0, fmt.Errorf("value out of range for uint8: %s", value)
	}

	return uint8(bigIntResult.Uint64()), nil
}

// Helper function to calculate 10^exp for small exponents
func pow10(exp int) float64 {
	if exp == 0 {
		return 1.0
	}
	if exp > 0 {
		result := 1.0
		for i := 0; i < exp; i++ {
			result *= 10.0
		}
		return result
	} else {
		result := 1.0
		for i := 0; i < -exp; i++ {
			result /= 10.0
		}
		return result
	}
}
