package eip712

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	chainselectors "github.com/smartcontractkit/chain-selectors"

	"github.com/smartcontractkit/crec-sdk/transact/signer"
	"github.com/smartcontractkit/crec-sdk/transact/types"
)

var (
	// ErrAuthorizationRequired is returned when the authorization input is nil.
	ErrAuthorizationRequired = errors.New("admin authorization is required")
	// ErrInvalidEIP712Type is returned when the pinned EIP-712 type string
	// cannot be parsed.
	ErrInvalidEIP712Type = errors.New("invalid EIP-712 type string")
	// ErrMissingArgument is returned when the arguments carry no value for a
	// struct field.
	ErrMissingArgument = errors.New("missing authorization argument")
	// ErrReservedArgument is returned when the arguments carry a value for the
	// authorization's reserved opId/deadline fields.
	ErrReservedArgument = errors.New("opId and deadline are reserved fields")
	// ErrInvalidArgumentValue is returned when an argument value does not match
	// the canonical encoding of its struct field type.
	ErrInvalidArgumentValue = errors.New("invalid authorization argument value")
	// ErrSignAuthorization is returned when signing the authorization fails.
	ErrSignAuthorization = errors.New("failed to sign authorization")
	// ErrCreateAuthorizationTypedData is returned when creating the EIP-712
	// typed data for the authorization fails.
	ErrCreateAuthorizationTypedData = errors.New("failed to create typed data for authorization")
	// ErrComputeAuthorizationHash is returned when computing the EIP-712 hash
	// of the authorization fails.
	ErrComputeAuthorizationHash = errors.New("failed to compute authorization hash")
)

// UpgradeAuthorizationType is the pinned EIP-712 type string the wallet
// contracts verify for upgrade authorizations. Unlike non-upgrade capabilities
// it is not published in the wallet type version catalog: it is fixed by the
// contract's AdminOperationLib and must never drift from it.
const UpgradeAuthorizationType = "UpgradeAuthorization(address newImplementation,uint256 opId,uint256 deadline)"

// Reserved struct field names every capability's pinned type ends in. The
// authorization supplies their values; capability arguments must not.
const (
	authorizationOpIDField     = "opId"
	authorizationDeadlineField = "deadline"
)

var (
	eip712TypeStringPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\((.*)\)$`)
	eip712FieldPattern      = regexp.MustCompile(`^(address|bool|string|bytes[0-9]{0,2}|u?int[0-9]{0,3})\s+([A-Za-z_][A-Za-z0-9_]*)$`)
	eip712BytesPattern      = regexp.MustCompile(`^bytes([0-9]+)?$`)
	eip712IntPattern        = regexp.MustCompile(`^(u?int)([0-9]*)$`)
)

// AdminAuthorization describes one owner signature over a capability's pinned
// EIP-712 struct. Arguments carries the capability's own argument values,
// canonically encoded exactly as they will be submitted to the API (integers
// as decimal strings, addresses and byte strings as lowercase 0x hex, booleans
// and strings as JSON values). The struct's trailing opId and deadline fields
// are supplied by OpID and Deadline.
type AdminAuthorization struct {
	// EIP712Type is the pinned type string, e.g. the catalog's eip712_type of
	// the capability being exercised.
	EIP712Type string
	// Arguments holds the capability argument values keyed by parameter name.
	Arguments map[string]any
	// OpID is the authorization's own identifier in the wallet's
	// admin-authorization namespace. It must differ from the enclosing
	// operation's wallet_operation_id.
	OpID *big.Int
	// Deadline is the Unix timestamp after which the authorization is unusable.
	Deadline *big.Int
}

// ParseEIP712Type parses a pinned EIP-712 type string such as
// "UpgradeAuthorization(address newImplementation,uint256 opId,uint256 deadline)"
// into the primary type name and its fields, preserving order.
func ParseEIP712Type(typeString string) (string, []apitypes.Type, error) {
	m := eip712TypeStringPattern.FindStringSubmatch(strings.TrimSpace(typeString))
	if m == nil {
		return "", nil, fmt.Errorf("%w: %q", ErrInvalidEIP712Type, typeString)
	}
	name, fieldsStr := m[1], m[2]

	var fields []apitypes.Type
	if strings.TrimSpace(fieldsStr) != "" {
		for _, f := range strings.Split(fieldsStr, ",") {
			fm := eip712FieldPattern.FindStringSubmatch(strings.TrimSpace(f))
			if fm == nil {
				return "", nil, fmt.Errorf("%w: field %q in %q", ErrInvalidEIP712Type, f, typeString)
			}
			fields = append(fields, apitypes.Type{Name: fm[2], Type: fm[1]})
		}
	}
	return name, fields, nil
}

// messageValue converts one canonically encoded argument into the Go value the
// EIP-712 typed-data message carries for its field type.
func messageValue(fieldType string, raw any) (any, error) {
	switch fieldType {
	case "address":
		s, ok := raw.(string)
		if !ok || !common.IsHexAddress(s) {
			return nil, fmt.Errorf("%w: address fields must be hex address strings, got %v", ErrInvalidArgumentValue, raw)
		}
		return s, nil
	case "bool":
		b, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: bool fields must be JSON booleans, got %v", ErrInvalidArgumentValue, raw)
		}
		return b, nil
	case "string":
		s, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("%w: string fields must be JSON strings, got %v", ErrInvalidArgumentValue, raw)
		}
		return s, nil
	}
	if eip712BytesPattern.MatchString(fieldType) {
		s, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s fields must be 0x-prefixed hex strings, got %v", ErrInvalidArgumentValue, fieldType, raw)
		}
		b, err := hexutil.Decode(s)
		if err != nil {
			return nil, fmt.Errorf("%w: %s field: %v", ErrInvalidArgumentValue, fieldType, err)
		}
		if fieldType == "bytes" {
			return b, nil
		}
		size, _ := strconv.Atoi(strings.TrimPrefix(fieldType, "bytes"))
		if size < 1 || size > 32 || len(b) != size {
			return nil, fmt.Errorf("%w: %s field must be exactly %d bytes", ErrInvalidArgumentValue, fieldType, size)
		}
		// go-ethereum's EIP-712 encoder reads fixed bytes from a [N]byte array.
		fixed := reflect.New(reflect.ArrayOf(size, reflect.TypeOf(byte(0)))).Elem()
		reflect.Copy(fixed, reflect.ValueOf(b))
		return fixed.Interface(), nil
	}
	if eip712IntPattern.MatchString(fieldType) {
		s, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s fields must be decimal strings, got %v", ErrInvalidArgumentValue, fieldType, raw)
		}
		n, ok := new(big.Int).SetString(s, 10)
		if !ok || n.String() != s {
			return nil, fmt.Errorf("%w: %s fields must be canonical decimal strings, got %q", ErrInvalidArgumentValue, fieldType, s)
		}
		return n.String(), nil
	}
	return nil, fmt.Errorf("%w: unsupported EIP-712 field type %q", ErrInvalidArgumentValue, fieldType)
}

// buildMessage assembles the EIP-712 message from the capability arguments and
// the authorization's reserved opId/deadline fields.
func (a *AdminAuthorization) buildMessage(fields []apitypes.Type) (apitypes.TypedDataMessage, error) {
	message := apitypes.TypedDataMessage{}
	for _, field := range fields {
		var raw any
		switch field.Name {
		case authorizationOpIDField:
			raw = a.OpID.String()
		case authorizationDeadlineField:
			raw = a.Deadline.String()
		default:
			v, ok := a.Arguments[field.Name]
			if !ok {
				return nil, fmt.Errorf("%w: %q", ErrMissingArgument, field.Name)
			}
			raw = v
		}
		value, err := messageValue(field.Type, raw)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.Name, err)
		}
		message[field.Name] = value
	}
	for name := range a.Arguments {
		if name == authorizationOpIDField || name == authorizationDeadlineField {
			return nil, fmt.Errorf("%w: %q", ErrReservedArgument, name)
		}
	}
	return message, nil
}

// typedData builds the full EIP-712 typed data for the authorization, bound to
// the wallet's own CLLSmartAccount domain.
func (a *AdminAuthorization) typedData(walletAddress common.Address, chainID *big.Int) (*apitypes.TypedData, error) {
	if a.OpID == nil || a.Deadline == nil {
		return nil, fmt.Errorf("%w: opId and deadline are required", ErrAuthorizationRequired)
	}
	primaryType, fields, err := ParseEIP712Type(a.EIP712Type)
	if err != nil {
		return nil, err
	}
	message, err := a.buildMessage(fields)
	if err != nil {
		return nil, err
	}
	if !chainID.IsInt64() {
		return nil, fmt.Errorf("%w: chain ID %s does not fit int64", ErrInvalidChainIDString, chainID.String())
	}
	domain := types.SmartAccountEIP712Domain(chainID.Int64(), walletAddress)
	return &apitypes.TypedData{
		Types: apitypes.Types{
			primaryType:   fields,
			domain.Type(): domain.Types(),
		},
		PrimaryType: primaryType,
		Domain:      domain.TypedData(),
		Message:     message,
	}, nil
}

// resolveChainID parses and validates the chain selector and resolves its EVM
// chain ID, mirroring HashOperation's chain handling.
func resolveChainID(chainSelector string) (*big.Int, error) {
	if chainSelector == "" || chainSelector == "0" {
		return nil, ErrParseChainSelector
	}
	chainSelectorUint, err := strconv.ParseUint(chainSelector, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseChainSelector, err)
	}
	chainFamily, err := chainselectors.GetSelectorFamily(chainSelectorUint)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChainFamily, err)
	}
	if chainFamily != chainselectors.FamilyEVM {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedChainFamily, chainFamily)
	}
	chainIdStr, err := chainselectors.GetChainIDFromSelector(chainSelectorUint)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetChainID, err)
	}
	chainID, ok := new(big.Int).SetString(chainIdStr, 10)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidChainIDString, chainIdStr)
	}
	return chainID, nil
}

// HashAdminAuthorization computes the EIP-712 digest of the given
// administrative authorization, bound to the wallet's own domain on the chain
// the selector names.
//   - auth: The authorization to hash.
//   - walletAddress: The wallet contract address; the EIP-712 verifying contract.
//   - chainSelector: Chain selector of the network the authorization executes on.
func (h *Handler) HashAdminAuthorization(auth *AdminAuthorization, walletAddress common.Address, chainSelector string) (common.Hash, error) {
	if auth == nil {
		return common.Hash{}, ErrAuthorizationRequired
	}
	chainID, err := resolveChainID(chainSelector)
	if err != nil {
		return common.Hash{}, err
	}
	typedData, err := auth.typedData(walletAddress, chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrCreateAuthorizationTypedData, err)
	}
	hashBytes, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrComputeAuthorizationHash, err)
	}
	return common.BytesToHash(hashBytes), nil
}

// SignAdminAuthorization signs the given administrative authorization using the
// provided signer, returning the authorization hash and the signature over it.
// Signers implementing signer.TypedDataSigner receive the full typed data;
// others sign the digest bytes.
//   - ctx: The context for the request.
//   - auth: The authorization to sign.
//   - s: The owner signer. See signer.Signer for details.
//   - walletAddress: The wallet contract address; the EIP-712 verifying contract.
//   - chainSelector: Chain selector of the network the authorization executes on.
func (h *Handler) SignAdminAuthorization(
	ctx context.Context,
	auth *AdminAuthorization,
	s signer.Signer,
	walletAddress common.Address,
	chainSelector string,
) (common.Hash, []byte, error) {
	if auth == nil {
		return common.Hash{}, nil, ErrAuthorizationRequired
	}
	if s == nil {
		return common.Hash{}, nil, ErrSignerRequired
	}

	hash, err := h.HashAdminAuthorization(auth, walletAddress, chainSelector)
	if err != nil {
		return common.Hash{}, nil, err
	}

	var sig []byte
	if typedDataSigner, ok := s.(signer.TypedDataSigner); ok {
		typedData, err := h.buildAuthorizationSignerTypedData(auth, walletAddress, chainSelector)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("%w: %w", ErrSignAuthorization, err)
		}
		sig, err = typedDataSigner.SignTypedData(ctx, typedData)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("%w: %w", ErrSignAuthorization, err)
		}
	} else {
		sig, err = s.Sign(ctx, hash.Bytes())
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("%w: %w", ErrSignAuthorization, err)
		}
	}

	h.logger.Debug("Signed Admin Authorization",
		"chain_selector", chainSelector,
		"wallet_address", walletAddress.Hex(),
		"eip712_type", auth.EIP712Type,
		"op_id", auth.OpID.String(),
		"hash", hash.Hex(),
		"signature", common.Bytes2Hex(sig))
	return hash, sig, nil
}

// buildAuthorizationSignerTypedData builds the signer.TypedData view of the
// authorization for custody providers implementing signer.TypedDataSigner.
func (h *Handler) buildAuthorizationSignerTypedData(
	auth *AdminAuthorization,
	walletAddress common.Address,
	chainSelector string,
) (*signer.TypedData, error) {
	chainID, err := resolveChainID(chainSelector)
	if err != nil {
		return nil, err
	}
	if !chainID.IsInt64() {
		return nil, fmt.Errorf("%w: chain ID %s does not fit int64", ErrInvalidChainIDString, chainID.String())
	}
	if auth.OpID == nil || auth.Deadline == nil {
		return nil, fmt.Errorf("%w: opId and deadline are required", ErrAuthorizationRequired)
	}
	primaryType, fields, err := ParseEIP712Type(auth.EIP712Type)
	if err != nil {
		return nil, err
	}
	message, err := auth.buildMessage(fields)
	if err != nil {
		return nil, err
	}

	domain := types.SmartAccountEIP712Domain(chainID.Int64(), walletAddress)
	typesFields := make([]signer.TypedDataField, len(fields))
	for i, f := range fields {
		typesFields[i] = signer.TypedDataField{Name: f.Name, Type: f.Type}
	}
	// JSON custodian view: dynamic and fixed bytes marshal as 0x hex.
	jsonMessage := make(map[string]any, len(message))
	for k, v := range message {
		if b, ok := v.([]byte); ok {
			jsonMessage[k] = hexutil.Encode(b)
			continue
		}
		rv := reflect.ValueOf(v)
		if rv.IsValid() && rv.Kind() == reflect.Array && rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(b), rv)
			jsonMessage[k] = hexutil.Encode(b)
			continue
		}
		jsonMessage[k] = v
	}
	return &signer.TypedData{
		Types: map[string][]signer.TypedDataField{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			primaryType: typesFields,
		},
		PrimaryType: primaryType,
		Domain: signer.TypedDataDomain{
			Name:              domain.Name,
			Version:           domain.Version,
			ChainID:           domain.ChainId,
			VerifyingContract: domain.VerifyingContract.Hex(),
		},
		Message: jsonMessage,
	}, nil
}
