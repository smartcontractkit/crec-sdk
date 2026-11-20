// Package adminop packs the calldata for signature-authorized wallet
// administration: upgrades and administrative capabilities.
//
// The courier service re-packs every submitted operation and requires the
// result to equal the transaction's calldata byte for byte, so the packing
// here mirrors the courier's canonical encoding exactly: capability argument
// values in parameter order, canonically encoded (integers as decimal strings,
// addresses and byte strings as lowercase 0x hex, booleans and strings as JSON
// values), followed by the authorization's opId, deadline and signature.
package adminop

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrInvalidABIFragment is returned when the capability's abi_fragment
	// cannot be parsed as a single ABI function.
	ErrInvalidABIFragment = errors.New("invalid abi fragment")
	// ErrArgumentMismatch is returned when the arguments do not name exactly
	// the capability's own parameters, or a value is not canonically encoded.
	ErrArgumentMismatch = errors.New("arguments do not match the capability")
	// ErrInvalidAuthorization is returned when the authorization's opId or
	// signature are not well formed.
	ErrInvalidAuthorization = errors.New("invalid authorization")
	// ErrMissingTrailingInputs is returned when the capability's ABI inputs do
	// not end in (uint256 opId, uint256 deadline, bytes signature).
	ErrMissingTrailingInputs = errors.New("capability inputs must end in (uint256 opId, uint256 deadline, bytes signature)")
)

// upgradeToWithAuthorizationABI is the account contract's upgrade entry point,
// identical across wallet types. Pinned here so the SDK packs the exact
// calldata the courier validates without depending on the contracts module.
const upgradeToWithAuthorizationABI = `[{"type":"function","name":"upgradeToWithAuthorization","inputs":[{"name":"newImplementation","type":"address"},{"name":"opId","type":"uint256"},{"name":"deadline","type":"uint256"},{"name":"signature","type":"bytes"}],"outputs":[],"stateMutability":"nonpayable"}]`

// PackUpgradeToWithAuthorization packs the canonical
// upgradeToWithAuthorization(newImplementation, opId, deadline, signature)
// calldata, selector included.
func PackUpgradeToWithAuthorization(
	newImplementation common.Address,
	opID, deadline *big.Int,
	signature []byte,
) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(upgradeToWithAuthorizationABI))
	if err != nil {
		return nil, fmt.Errorf("wallet upgrade: load ABI: %w", err)
	}
	packed, err := parsed.Pack("upgradeToWithAuthorization", newImplementation, opID, deadline, signature)
	if err != nil {
		return nil, fmt.Errorf("wallet upgrade: pack upgradeToWithAuthorization: %w", err)
	}
	return packed, nil
}

// PackCapabilityCall packs the capability's call with the submission's
// arguments followed by the authorization's opId, deadline and signature,
// selector included. The abiFragment is the full ABI function object the
// wallet type version publishes; arguments must name exactly the fragment's
// parameters before the trailing authorization inputs.
func PackCapabilityCall(
	abiFragment map[string]any,
	arguments map[string]any,
	opID, deadline *big.Int,
	signature []byte,
) ([]byte, error) {
	method, err := parseCapabilityFragment(abiFragment)
	if err != nil {
		return nil, err
	}
	if deadline == nil {
		return nil, fmt.Errorf("%w: deadline is required", ErrInvalidAuthorization)
	}
	opIDBig, err := requireNonNegativeInt(opID, "opId")
	if err != nil {
		return nil, err
	}

	params := capabilityParams(method)
	if len(method.Inputs) != len(params)+3 {
		return nil, fmt.Errorf("%w, got %d inputs for %d arguments",
			ErrMissingTrailingInputs, len(method.Inputs), len(params))
	}

	values, err := capabilityArgumentValues(method, params, arguments)
	if err != nil {
		return nil, err
	}

	args := append(append([]any{}, values...), opIDBig, deadline, signature)
	packed, err := method.Inputs.Pack(args...)
	if err != nil {
		return nil, fmt.Errorf("%w: arguments do not pack: %v", ErrArgumentMismatch, err)
	}
	return append(append([]byte{}, method.ID...), packed...), nil
}

// parseCapabilityFragment unmarshals the published abi_fragment object into its
// ABI method, rejecting anything that is not exactly one function.
func parseCapabilityFragment(abiFragment map[string]any) (abi.Method, error) {
	raw, err := json.Marshal([]map[string]any{abiFragment})
	if err != nil {
		return abi.Method{}, fmt.Errorf("%w: %v", ErrInvalidABIFragment, err)
	}
	parsed, err := abi.JSON(bytes.NewReader(raw))
	if err != nil {
		return abi.Method{}, fmt.Errorf("%w: %v", ErrInvalidABIFragment, err)
	}
	if len(parsed.Methods) != 1 {
		return abi.Method{}, fmt.Errorf("%w: fragment must declare exactly one function", ErrInvalidABIFragment)
	}
	for _, m := range parsed.Methods {
		return m, nil
	}
	return abi.Method{}, fmt.Errorf("%w: fragment declares no function", ErrInvalidABIFragment)
}

// capabilityParams returns the method's inputs minus the trailing
// (uint256 opId, uint256 deadline, bytes signature) authorization inputs,
// when present. When absent it returns all inputs; PackCapabilityCall rejects
// the shape afterwards.
func capabilityParams(method abi.Method) abi.Arguments {
	inputs := method.Inputs
	if len(inputs) >= 3 {
		tail := inputs[len(inputs)-3:]
		if tail[0].Type.String() == "uint256" && tail[1].Type.String() == "uint256" && tail[2].Type.String() == "bytes" {
			return inputs[:len(inputs)-3]
		}
	}
	return inputs
}

func requireNonNegativeInt(n *big.Int, name string) (*big.Int, error) {
	if n == nil || n.Sign() < 0 {
		return nil, fmt.Errorf("%w: %s must be a non-negative integer", ErrInvalidAuthorization, name)
	}
	return n, nil
}

// capabilityArgumentValues decodes canonically encoded arguments into the Go
// values the capability's ABI packs, in parameter order. The arguments must
// name exactly the capability's parameters: a missing, extra or null one is a
// mismatch, since the packed call could not be the one the owner signed.
func capabilityArgumentValues(method abi.Method, params abi.Arguments, arguments map[string]any) ([]any, error) {
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}
	for name := range arguments {
		if !slices.Contains(names, name) {
			return nil, fmt.Errorf("%w: %s takes no argument %q", ErrArgumentMismatch, method.Name, name)
		}
	}
	values := make([]any, len(params))
	for i, p := range params {
		raw, ok := arguments[p.Name]
		if !ok {
			return nil, fmt.Errorf("%w: argument %q is missing", ErrArgumentMismatch, p.Name)
		}
		if raw == nil {
			return nil, fmt.Errorf("%w: argument %q is null", ErrArgumentMismatch, p.Name)
		}
		v, err := canonicalABIValue(p.Type, raw)
		if err != nil {
			return nil, fmt.Errorf("%w: argument %q: %v", ErrArgumentMismatch, p.Name, err)
		}
		values[i] = v
	}
	return values, nil
}

var (
	canonicalAddress = regexp.MustCompile(`^0x[0-9a-f]{40}$`)
	canonicalHex     = regexp.MustCompile(`^0x([0-9a-f]{2})*$`)
)

// canonicalABIValue reads one canonically encoded argument: addresses and byte
// strings as lowercase 0x hex, integers as decimal strings, booleans and
// strings as JSON values. A value in any other encoding is rejected rather
// than normalized, so what the SDK packs is exactly what the courier expects.
func canonicalABIValue(t abi.Type, raw any) (any, error) {
	switch t.T {
	case abi.BoolTy:
		b, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("must be a JSON boolean")
		}
		return b, nil
	case abi.StringTy:
		s, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("must be a JSON string")
		}
		return s, nil
	case abi.AddressTy:
		s, ok := raw.(string)
		if !ok || !canonicalAddress.MatchString(s) {
			return nil, fmt.Errorf("must be a lowercase 0x-prefixed address")
		}
		return common.HexToAddress(s), nil
	case abi.BytesTy, abi.FixedBytesTy:
		s, ok := raw.(string)
		if !ok || !canonicalHex.MatchString(s) {
			return nil, fmt.Errorf("must be lowercase 0x-prefixed hex")
		}
		b, _ := hex.DecodeString(s[2:])
		if t.T == abi.BytesTy {
			return b, nil
		}
		if len(b) != t.Size {
			return nil, fmt.Errorf("must be exactly %d bytes", t.Size)
		}
		fixed := reflect.New(t.GetType()).Elem()
		reflect.Copy(fixed, reflect.ValueOf(b))
		return fixed.Interface(), nil
	case abi.IntTy, abi.UintTy:
		return canonicalInteger(t, raw)
	}
	return nil, fmt.Errorf("type %s is not supported", t.String())
}

func canonicalInteger(t abi.Type, raw any) (any, error) {
	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("must be a decimal string")
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.String() != s {
		return nil, fmt.Errorf("must be a canonical decimal string")
	}
	limit := new(big.Int).Lsh(big.NewInt(1), uint(t.Size))
	if t.T == abi.UintTy {
		if n.Sign() < 0 || n.Cmp(limit) >= 0 {
			return nil, fmt.Errorf("out of range for %s", t.String())
		}
	} else {
		half := new(big.Int).Rsh(limit, 1)
		if n.Cmp(new(big.Int).Neg(half)) < 0 || n.Cmp(half) >= 0 {
			return nil, fmt.Errorf("out of range for %s", t.String())
		}
	}
	goType := t.GetType()
	if goType == reflect.TypeOf((*big.Int)(nil)) {
		return n, nil
	}
	// go-ethereum packs sizes up to 64 bits from their exact Go integer types.
	v := reflect.New(goType).Elem()
	if t.T == abi.UintTy {
		v.SetUint(n.Uint64())
	} else {
		v.SetInt(n.Int64())
	}
	return v.Interface(), nil
}
